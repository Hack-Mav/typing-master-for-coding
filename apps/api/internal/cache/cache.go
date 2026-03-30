package cache

import (
	"sync"
	"time"
)

// CacheItem represents a cached item with expiration
type CacheItem struct {
	Value     interface{}
	ExpiresAt time.Time
}

// InMemoryCache provides thread-safe in-memory caching with TTL
type InMemoryCache struct {
	items   sync.Map
	maxSize int
	ttl     time.Duration
	size    int64
	mu      sync.RWMutex
}

// NewInMemoryCache creates a new in-memory cache
func NewInMemoryCache(maxSize int, ttlMinutes int) *InMemoryCache {
	cache := &InMemoryCache{
		maxSize: maxSize,
		ttl:     time.Duration(ttlMinutes) * time.Minute,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Set stores a value in the cache
func (c *InMemoryCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict items
	if c.size >= int64(c.maxSize) {
		c.evictLRU()
	}

	item := &CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
	}

	// Check if key already exists
	if _, exists := c.items.Load(key); !exists {
		c.size++
	}

	c.items.Store(key, item)
}

// Get retrieves a value from the cache
func (c *InMemoryCache) Get(key string) (interface{}, bool) {
	value, exists := c.items.Load(key)
	if !exists {
		return nil, false
	}

	item := value.(*CacheItem)

	// Check if item has expired
	if time.Now().After(item.ExpiresAt) {
		c.Delete(key)
		return nil, false
	}

	return item.Value, true
}

// Delete removes a value from the cache
func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.items.Load(key); exists {
		c.items.Delete(key)
		c.size--
	}
}

// Clear removes all items from the cache
func (c *InMemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items.Range(func(key, value interface{}) bool {
		c.items.Delete(key)
		return true
	})
	c.size = 0
}

// Size returns the current number of items in the cache
func (c *InMemoryCache) Size() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.size
}

// evictLRU removes the least recently used item (simplified LRU)
func (c *InMemoryCache) evictLRU() {
	var oldestKey interface{}
	var oldestTime time.Time

	c.items.Range(func(key, value interface{}) bool {
		item := value.(*CacheItem)
		if oldestKey == nil || item.ExpiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.ExpiresAt
		}
		return true
	})

	if oldestKey != nil {
		c.items.Delete(oldestKey)
		c.size--
	}
}

// cleanup periodically removes expired items
func (c *InMemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		var keysToDelete []interface{}

		c.items.Range(func(key, value interface{}) bool {
			item := value.(*CacheItem)
			if now.After(item.ExpiresAt) {
				keysToDelete = append(keysToDelete, key)
			}
			return true
		})

		c.mu.Lock()
		for _, key := range keysToDelete {
			c.items.Delete(key)
			c.size--
		}
		c.mu.Unlock()
	}
}
