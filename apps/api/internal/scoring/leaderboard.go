package scoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/typing-master-for-coding-backend/internal/database"
)

// Leaderboard entity for Datastore
type Leaderboard struct {
	ID         string `datastore:"-" json:"id"`
	UserID     string `datastore:"user_id" json:"user_id"`
	SessionID  string `datastore:"session_id" json:"session_id"`
	LanguageID string `datastore:"language_id" json:"language_id"`
	Mode       string `datastore:"mode" json:"mode"`
	Scope      string `datastore:"scope" json:"scope"`             // "global", "friends", "organization"
	TimeWindow string `datastore:"time_window" json:"time_window"` // "daily", "weekly", "monthly", "all_time"
	Username   string `datastore:"username" json:"username"`
	Handle     string `datastore:"handle" json:"handle"`

	// Rankings
	Rank     int     `datastore:"rank" json:"rank"`
	Score    int     `datastore:"score" json:"score"`
	CPM      float64 `datastore:"cpm" json:"cpm"`
	TWPM     float64 `datastore:"twpm" json:"twpm"`
	Accuracy float64 `datastore:"accuracy" json:"accuracy"`

	// Metadata
	MetricsSnapshot    map[string]interface{} `datastore:"metrics_snapshot" json:"metrics_snapshot"`
	Badge              string                 `datastore:"badge" json:"badge"`
	IsVerified         bool                   `datastore:"is_verified" json:"is_verified"`
	VerificationMethod string                 `datastore:"verification_method" json:"verification_method"`

	// Timestamps
	RecordedAt  time.Time `datastore:"recorded_at" json:"recorded_at"`
	WindowStart time.Time `datastore:"window_start" json:"window_start"`
	WindowEnd   time.Time `datastore:"window_end" json:"window_end"`
}

// LeaderboardService handles leaderboard operations and rankings
type leaderboardServiceImpl struct {
	dsClient DatastoreClient
	cache    *sync.Map
	ttl      time.Duration
}

// NewLeaderboardService creates a new leaderboard service
func NewLeaderboardService(dsClient DatastoreClient) LeaderboardService {
	service := &leaderboardServiceImpl{
		dsClient: dsClient,
		cache:    &sync.Map{},
		ttl:      time.Minute * 5, // Cache for 5 minutes
	}

	// Start cache cleanup routine
	go service.cacheCleanup()

	return service
}

// GetLeaderboard retrieves leaderboard rankings with filtering and caching
func (s *leaderboardServiceImpl) GetLeaderboard(ctx context.Context, req *LeaderboardRequest) (*LeaderboardResponse, error) {
	cacheKey := s.generateCacheKey(req)

	// Check cache first
	if cached, ok := s.cache.Load(cacheKey); ok {
		if response, ok := cached.(*LeaderboardResponse); ok && time.Since(response.GeneratedAt) < s.ttl {
			return response, nil
		}
	}

	// Query leaderboard data
	entries, err := s.queryLeaderboard(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to query leaderboard: %w", err)
	}

	// Create response
	response := &LeaderboardResponse{
		Entries:     entries,
		TotalCount:  len(entries),
		GeneratedAt: time.Now(),
		Filters:     req,
	}

	// Cache response
	s.cache.Store(cacheKey, response)

	return response, nil
}

// UpdateLeaderboard updates rankings when new scores are submitted
func (s *leaderboardServiceImpl) UpdateLeaderboard(ctx context.Context, sessionID string) error {
	// Get session metrics from scoring service
	// For now, we'll assume this is handled by the scoring service

	// Invalidate relevant cache entries
	s.invalidateCacheForSession(sessionID)

	return nil
}

// GetUserRank retrieves a specific user's rank in a leaderboard
func (s *leaderboardServiceImpl) GetUserRank(ctx context.Context, userID, languageID, mode, scope, timeWindow string) (*UserRank, error) {
	req := &LeaderboardRequest{
		LanguageID: languageID,
		Mode:       mode,
		Scope:      scope,
		TimeWindow: timeWindow,
		Limit:      1000, // Get enough entries to find user's rank
	}

	response, err := s.GetLeaderboard(ctx, req)
	if err != nil {
		return nil, err
	}

	// Find user's rank
	for i, entry := range response.Entries {
		if entry.UserID == userID {
			return &UserRank{
				Rank:       i + 1,
				Score:      entry.Score,
				TotalUsers: response.TotalCount,
				Percentile: float64(response.TotalCount-i) / float64(response.TotalCount) * 100,
			}, nil
		}
	}

	return &UserRank{
		Rank:       0, // Not found in top rankings
		TotalUsers: response.TotalCount,
		Percentile: 0,
	}, nil
}

// GetLeaderboardTrends retrieves trending scores and improvements
func (s *leaderboardServiceImpl) GetLeaderboardTrends(ctx context.Context, req *LeaderboardTrendsRequest) (*LeaderboardTrendsResponse, error) {
	// TODO: Implement trends analysis
	return &LeaderboardTrendsResponse{
		Trends: []LeaderboardTrend{},
	}, nil
}

// LeaderboardRequest represents a leaderboard query request
type LeaderboardRequest struct {
	LanguageID string `json:"language_id"`
	Mode       string `json:"mode"`
	Scope      string `json:"scope"`       // "global", "friends", "organization"
	TimeWindow string `json:"time_window"` // "daily", "weekly", "monthly", "all_time"
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
	UserID     string `json:"user_id,omitempty"` // For friends/organization scope
}

// LeaderboardResponse represents a leaderboard query response
type LeaderboardResponse struct {
	Entries     []LeaderboardEntry  `json:"entries"`
	TotalCount  int                 `json:"total_count"`
	GeneratedAt time.Time           `json:"generated_at"`
	Filters     *LeaderboardRequest `json:"filters"`
}

// LeaderboardEntry represents a single leaderboard entry
type LeaderboardEntry struct {
	UserID             string    `json:"user_id"`
	Username           string    `json:"username"`
	Handle             string    `json:"handle"`
	Score              int       `json:"score"`
	Rank               int       `json:"rank"`
	CPM                float64   `json:"cpm"`
	TWPM               float64   `json:"twpm"`
	Accuracy           float64   `json:"accuracy"`
	Badge              string    `json:"badge"`
	IsVerified         bool      `json:"is_verified"`
	VerificationMethod string    `json:"verification_method"`
	SessionCount       int       `json:"session_count"`
	LastActive         time.Time `json:"last_active"`
}

// UserRank represents a user's ranking information
type UserRank struct {
	Rank       int     `json:"rank"`
	Score      int     `json:"score"`
	TotalUsers int     `json:"total_users"`
	Percentile float64 `json:"percentile"`
}

// LeaderboardTrendsRequest represents a trends query request
type LeaderboardTrendsRequest struct {
	LanguageID string `json:"language_id"`
	Mode       string `json:"mode"`
	Scope      string `json:"scope"`
	TimeWindow string `json:"time_window"`
	Period     string `json:"period"` // "7d", "30d", "90d"
	Metric     string `json:"metric"` // "score", "cpm", "twpm", "accuracy"
}

// LeaderboardTrendsResponse represents a trends query response
type LeaderboardTrendsResponse struct {
	Trends      []LeaderboardTrend `json:"trends"`
	GeneratedAt time.Time          `json:"generated_at"`
}

// LeaderboardTrend represents trending data for leaderboards
type LeaderboardTrend struct {
	Date             time.Time `json:"date"`
	Average          float64   `json:"average"`
	Median           float64   `json:"median"`
	TopScore         int       `json:"top_score"`
	ParticipantCount int       `json:"participant_count"`
}

// TournamentLeaderboard represents tournament-specific leaderboard
type TournamentLeaderboard struct {
	TournamentID string             `json:"tournament_id"`
	Title        string             `json:"title"`
	Entries      []LeaderboardEntry `json:"entries"`
	Status       string             `json:"status"`
	EndTime      time.Time          `json:"end_time"`
}

// queryLeaderboard performs the actual leaderboard query
func (s *leaderboardServiceImpl) queryLeaderboard(ctx context.Context, req *LeaderboardRequest) ([]LeaderboardEntry, error) {
	var entries []LeaderboardEntry

	// Calculate time window
	startTime, endTime := s.calculateTimeWindow(req.TimeWindow)

	// Build query based on scope
	query := database.NewQuery("Leaderboard").
		FilterField("LanguageID", "=", req.LanguageID).
		FilterField("Mode", "=", req.Mode).
		FilterField("Scope", "=", req.Scope).
		FilterField("TimeWindow", "=", req.TimeWindow).
		FilterField("WindowStart", ">=", startTime).
		FilterField("WindowEnd", "<=", endTime).
		Order("-Score").
		Limit(req.Limit).
		Offset(req.Offset)

	// Execute query
	var leaderboardEntities []Leaderboard
	_, err := s.dsClient.GetAll(ctx, query, &leaderboardEntities)
	if err != nil {
		return nil, fmt.Errorf("failed to query leaderboard: %w", err)
	}

	// Convert to response format
	for i, entity := range leaderboardEntities {
		entry := LeaderboardEntry{
			UserID:             entity.UserID,
			Username:           entity.Username,
			Handle:             entity.Handle,
			Score:              entity.Score,
			Rank:               i + 1 + req.Offset,
			CPM:                entity.CPM,
			TWPM:               entity.TWPM,
			Accuracy:           entity.Accuracy,
			Badge:              entity.Badge,
			IsVerified:         entity.IsVerified,
			VerificationMethod: entity.VerificationMethod,
			SessionCount:       1, // TODO: Calculate from session data
			LastActive:         entity.RecordedAt,
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// calculateTimeWindow calculates start and end times for time window filtering
func (s *leaderboardServiceImpl) calculateTimeWindow(timeWindow string) (time.Time, time.Time) {
	now := time.Now()
	var startTime time.Time

	switch timeWindow {
	case "daily":
		startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "weekly":
		// Start of week (Monday)
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday is 0, but we want Monday as start
		}
		startTime = now.AddDate(0, 0, -(weekday - 1))
		startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, startTime.Location())
	case "monthly":
		startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	case "all_time":
		startTime = time.Date(2020, 1, 1, 0, 0, 0, 0, now.Location()) // Arbitrary early date
	default:
		startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	}

	return startTime, now
}

// generateCacheKey generates a cache key for leaderboard requests
func (s *leaderboardServiceImpl) generateCacheKey(req *LeaderboardRequest) string {
	return fmt.Sprintf("lb:%s:%s:%s:%s:%d:%d",
		req.LanguageID, req.Mode, req.Scope, req.TimeWindow, req.Limit, req.Offset)
}

// invalidateCacheForSession invalidates cache entries related to a session
func (s *leaderboardServiceImpl) invalidateCacheForSession(sessionID string) {
	// TODO: Implement cache invalidation logic
	// This would iterate through cache keys and remove entries containing the session ID
}

// cacheCleanup periodically cleans up expired cache entries
func (s *leaderboardServiceImpl) cacheCleanup() {
	ticker := time.NewTicker(time.Minute * 10)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanupExpiredCache()
	}
}

// cleanupExpiredCache removes expired entries from cache
func (s *leaderboardServiceImpl) cleanupExpiredCache() {
	now := time.Now()
	s.cache.Range(func(key, value interface{}) bool {
		if response, ok := value.(*LeaderboardResponse); ok {
			if now.Sub(response.GeneratedAt) > s.ttl {
				s.cache.Delete(key)
			}
		}
		return true
	})
}

// Helper methods for user data lookup
func (s *leaderboardServiceImpl) getUsername(userID string) string {
	// TODO: Implement user lookup from User entity
	return fmt.Sprintf("user_%s", userID[:8])
}

func (s *leaderboardServiceImpl) getUserHandle(userID string) string {
	// TODO: Implement user handle lookup
	return s.getUsername(userID)
}
