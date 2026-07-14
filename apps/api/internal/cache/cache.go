package cache

import (
	"bytes"
	"container/list"
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/typing-master-for-coding-backend/internal/config"
	"github.com/typing-master-for-coding-backend/internal/models"
)

func init() {
	// Register model types used by cached values so gob can round-trip them.
	// Only value types are registered here; gob handles pointers and slices
	// automatically once the base type is known.
	gob.Register(models.Session{})
	gob.Register(models.User{})
	gob.Register(models.Language{})
	gob.Register(models.Lesson{})
	gob.Register(models.Snippet{})
	gob.Register(models.Playlist{})
	gob.Register(models.Result{})
	gob.Register(models.LessonProgress{})
	gob.Register(models.AssessmentBlueprint{})
	gob.Register(models.AssessmentCriteria{})
	gob.Register(models.AssessmentWeights{})
	gob.Register(models.AssessmentSession{})
	gob.Register(models.AssessmentSnippetResult{})
	gob.Register(models.AssessmentResult{})
	gob.Register(models.AssessmentAnalytics{})
	gob.Register(models.UserBadge{})
	gob.Register(models.Tutorial{})
	gob.Register(models.UserTutorialProgress{})
	gob.Register(models.HelpArticle{})
	gob.Register(models.FAQ{})
	gob.Register(models.Feedback{})
	gob.Register(models.SupportTicket{})
	gob.Register(models.SupportMessage{})
	gob.Register(models.SupportChannel{})
	gob.Register(models.CommunityForum{})
	gob.Register(models.ForumPost{})
	gob.Register(models.ForumReply{})
	gob.Register(models.ABTest{})
	gob.Register(models.ABTestResult{})
	gob.Register(models.UserABTestAssignment{})
	gob.Register(models.AnalyticsEvent{})
	gob.Register(models.AnalyticsConsent{})
	gob.Register(models.UserOnboardingState{})
	gob.Register(models.Tooltip{})
	gob.Register(models.SystemVersion{})
	gob.Register(models.ComplianceStandard{})
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
}

// CacheItem represents a cached item with expiration.
// The unexported key and elem fields are used internally by the in-memory
// backend to maintain O(1) LRU ordering and safe cleanup.
type CacheItem struct {
	Value     interface{}
	ExpiresAt time.Time
	key       string
	elem      *list.Element
}

// cacheBackend abstracts the underlying storage used by InMemoryCache.
type cacheBackend interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	SetWithTTL(key string, value interface{}, ttl time.Duration)
	IncrWithTTL(key string, ttl time.Duration) int64
	Delete(key string)
	Clear()
	Size() int64
	Range(fn func(key string, value interface{}) bool)
	Close() error
}

// InMemoryCache provides a caching layer that can be backed by either an
// in-memory store or Redis.
type InMemoryCache struct {
	backend cacheBackend
}

// NewInMemoryCache creates a new in-memory cache.
func NewInMemoryCache(maxSize int, ttlMinutes int) *InMemoryCache {
	return &InMemoryCache{
		backend: newMemoryCache(maxSize, ttlMinutes),
	}
}

// NewCache creates a cache, using Redis when REDIS_URL is configured, otherwise
// falling back to an in-memory cache.
func NewCache(cfg *config.Config) *InMemoryCache {
	if cfg.RedisURL == "" {
		return NewInMemoryCache(cfg.CacheMaxSize, cfg.CacheTTLMinutes)
	}

	backend, err := newRedisCache(cfg)
	if err != nil {
		log.Printf("Failed to create Redis cache, falling back to in-memory: %v", err)
		return NewInMemoryCache(cfg.CacheMaxSize, cfg.CacheTTLMinutes)
	}
	return &InMemoryCache{backend: backend}
}

// Set stores a value in the cache with the configured TTL.
func (c *InMemoryCache) Set(key string, value interface{}) {
	if c == nil || c.backend == nil {
		return
	}
	c.backend.Set(key, value)
}

// SetWithTTL stores a value with a custom TTL.
func (c *InMemoryCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	if c == nil || c.backend == nil {
		return
	}
	c.backend.SetWithTTL(key, value, ttl)
}

func (c *InMemoryCache) IncrWithTTL(key string, ttl time.Duration) int64 {
	if c == nil || c.backend == nil {
		return 0
	}
	return c.backend.IncrWithTTL(key, ttl)
}

// Get retrieves a value from the cache.
func (c *InMemoryCache) Get(key string) (interface{}, bool) {
	if c == nil || c.backend == nil {
		return nil, false
	}
	return c.backend.Get(key)
}

// Delete removes a value from the cache.
func (c *InMemoryCache) Delete(key string) {
	if c == nil || c.backend == nil {
		return
	}
	c.backend.Delete(key)
}

// Clear removes all items from the cache.
func (c *InMemoryCache) Clear() {
	if c == nil || c.backend == nil {
		return
	}
	c.backend.Clear()
}

// Size returns the current number of items in the cache.
func (c *InMemoryCache) Size() int64 {
	if c == nil || c.backend == nil {
		return 0
	}
	return c.backend.Size()
}

// Range iterates over the cache.
func (c *InMemoryCache) Range(fn func(key string, value interface{}) bool) {
	if c == nil || c.backend == nil {
		return
	}
	c.backend.Range(fn)
}

// Close closes the cache backend.
func (c *InMemoryCache) Close() error {
	if c == nil || c.backend == nil {
		return nil
	}
	return c.backend.Close()
}

// ---- memory cache backend ----

type memoryCache struct {
	items   map[string]*CacheItem
	maxSize int
	ttl     time.Duration
	order   *list.List
	mu      sync.RWMutex
}

func newMemoryCache(maxSize int, ttlMinutes int) *memoryCache {
	m := &memoryCache{
		items:   make(map[string]*CacheItem),
		maxSize: maxSize,
		ttl:     time.Duration(ttlMinutes) * time.Minute,
		order:   list.New(),
	}
	go m.cleanup()
	return m
}

func (m *memoryCache) Set(key string, value interface{}) {
	m.SetWithTTL(key, value, m.ttl)
}

func (m *memoryCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.maxSize > 0 && len(m.items) >= m.maxSize {
		m.evictLRU()
	}

	if item, exists := m.items[key]; exists {
		item.Value = value
		item.ExpiresAt = time.Now().UTC().Add(ttl)
		m.order.MoveToFront(item.elem)
		return
	}

	item := &CacheItem{Value: value, ExpiresAt: time.Now().UTC().Add(ttl), key: key}
	item.elem = m.order.PushFront(item)
	m.items[key] = item
}

func (m *memoryCache) Get(key string) (interface{}, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.items[key]
	if !exists {
		return nil, false
	}

	if time.Now().UTC().After(item.ExpiresAt) {
		m.removeItem(item)
		delete(m.items, key)
		return nil, false
	}

	m.order.MoveToFront(item.elem)
	return item.Value, true
}

func (m *memoryCache) IncrWithTTL(key string, ttl time.Duration) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	if item, exists := m.items[key]; exists && !now.After(item.ExpiresAt) {
		if v, ok := item.Value.(int64); ok {
			item.Value = v + 1
			item.ExpiresAt = now.Add(ttl)
			m.order.MoveToFront(item.elem)
			return v + 1
		}
	}

	if m.maxSize > 0 && len(m.items) >= m.maxSize {
		m.evictLRU()
	}

	item := &CacheItem{Value: int64(1), ExpiresAt: now.Add(ttl), key: key}
	item.elem = m.order.PushFront(item)
	m.items[key] = item
	return 1
}

func (m *memoryCache) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if item, exists := m.items[key]; exists {
		m.removeItem(item)
		delete(m.items, key)
	}
}

func (m *memoryCache) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items = make(map[string]*CacheItem)
	m.order = list.New()
}

func (m *memoryCache) Size() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int64(len(m.items))
}

func (m *memoryCache) Range(fn func(key string, value interface{}) bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	now := time.Now().UTC()
	for k, item := range m.items {
		if now.After(item.ExpiresAt) {
			continue
		}
		if !fn(k, item.Value) {
			break
		}
	}
}

func (m *memoryCache) Close() error { return nil }

func (m *memoryCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		m.mu.Lock()
		now := time.Now().UTC()
		for k, item := range m.items {
			if now.After(item.ExpiresAt) {
				m.removeItem(item)
				delete(m.items, k)
			}
		}
		m.mu.Unlock()
	}
}

func (m *memoryCache) evictLRU() {
	if m.order.Len() == 0 {
		return
	}
	oldest := m.order.Back()
	if oldest == nil {
		return
	}
	item := oldest.Value.(*CacheItem)
	m.order.Remove(oldest)
	delete(m.items, item.key)
}

func (m *memoryCache) removeItem(item *CacheItem) {
	if item != nil && item.elem != nil {
		m.order.Remove(item.elem)
		item.elem = nil
	}
}

// ---- redis cache backend ----

type redisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func newRedisCache(cfg *config.Config) (*redisCache, error) {
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		opts = &redis.Options{
			Addr: cfg.RedisURL,
		}
	}

	opts.PoolSize = cfg.RedisPoolSize
	opts.MinIdleConns = cfg.RedisMinIdleConns

	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return &redisCache{
		client: client,
		ttl:    time.Duration(cfg.CacheTTLMinutes) * time.Minute,
	}, nil
}

func (r *redisCache) Set(key string, value interface{}) {
	r.SetWithTTL(key, value, r.ttl)
}

func (r *redisCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	data, err := encodeGob(value)
	if err != nil {
		log.Printf("Failed to encode cache value for %s: %v", key, err)
		return
	}
	if err := r.client.Set(context.Background(), key, data, ttl).Err(); err != nil {
		log.Printf("Failed to write cache value for %s: %v", key, err)
	}
}

func (r *redisCache) IncrWithTTL(key string, ttl time.Duration) int64 {
	ctx := context.Background()
	const script = `
local current = redis.call('incr', KEYS[1])
if current == 1 then
    redis.call('expire', KEYS[1], ARGV[1])
end
return current
`
	ttlSec := strconv.FormatInt(int64(ttl.Seconds()), 10)
	n, err := r.client.Eval(ctx, script, []string{key}, ttlSec).Int64()
	if err != nil {
		log.Printf("Failed to incr cache key %s: %v", key, err)
		return 0
	}
	return n
}

func (r *redisCache) Get(key string) (interface{}, bool) {
	data, err := r.client.Get(context.Background(), key).Bytes()
	if err == redis.Nil {
		return nil, false
	}
	if err != nil {
		log.Printf("Failed to read cache value for %s: %v", key, err)
		return nil, false
	}

	value, err := decodeGob(data)
	if err != nil {
		log.Printf("Failed to decode cache value for %s: %v", key, err)
		return nil, false
	}
	return value, true
}

func (r *redisCache) Delete(key string) {
	if err := r.client.Del(context.Background(), key).Err(); err != nil {
		log.Printf("Failed to delete cache key %s: %v", key, err)
	}
}

func (r *redisCache) Clear() {
	if err := r.client.FlushDB(context.Background()).Err(); err != nil {
		log.Printf("Failed to flush Redis cache: %v", err)
	}
}

func (r *redisCache) Size() int64 {
	size, err := r.client.DBSize(context.Background()).Result()
	if err != nil {
		log.Printf("Failed to get Redis cache size: %v", err)
		return 0
	}
	return size
}

func (r *redisCache) Range(fn func(key string, value interface{}) bool) {
	ctx := context.Background()
	iter := r.client.Scan(ctx, 0, "", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		value, ok := r.Get(key)
		if !ok {
			continue
		}
		if !fn(key, value) {
			break
		}
	}
	if err := iter.Err(); err != nil {
		log.Printf("Failed to scan Redis cache: %v", err)
	}
}

func (r *redisCache) Close() error {
	return r.client.Close()
}

func encodeGob(value interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(&value); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decodeGob(data []byte) (interface{}, error) {
	var value interface{}
	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}
