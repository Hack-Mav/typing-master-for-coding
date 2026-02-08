package database

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/iterator"
)

// QueryOptimizer provides optimized query patterns for Datastore
type QueryOptimizer struct {
	client *datastore.Client
}

// NewQueryOptimizer creates a new query optimizer
func NewQueryOptimizer(client *datastore.Client) *QueryOptimizer {
	return &QueryOptimizer{
		client: client,
	}
}

// QueryOptions holds common query options
type QueryOptions struct {
	Limit         int
	Offset        int
	StartAfter    interface{}
	UseKeysOnly   bool
	UseProjection []string
}

// GetUserSessionsPaginated retrieves user sessions with cursor-based pagination
func (qo *QueryOptimizer) GetUserSessionsPaginated(
	ctx context.Context,
	userID string,
	limit int,
	cursor string,
) ([]*Session, string, error) {
	query := datastore.NewQuery("Session").
		Filter("user_id =", userID).
		Order("-created_at").
		Limit(limit)

	// Use cursor for pagination (more efficient than offset)
	if cursor != "" {
		decodedCursor, err := datastore.DecodeCursor(cursor)
		if err != nil {
			return nil, "", fmt.Errorf("invalid cursor: %w", err)
		}
		query = query.Start(decodedCursor)
	}

	var sessions []*Session
	it := qo.client.Run(ctx, query)

	for {
		var session Session
		key, err := it.Next(&session)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, "", err
		}
		session.ID = key.Name
		sessions = append(sessions, &session)
	}

	// Get cursor for next page
	nextCursor, err := it.Cursor()
	if err != nil {
		return sessions, "", nil
	}

	return sessions, nextCursor.String(), nil
}

// GetLeaderboardOptimized retrieves leaderboard with optimized query
func (qo *QueryOptimizer) GetLeaderboardOptimized(
	ctx context.Context,
	languageID string,
	mode string,
	timeWindow string,
	limit int,
) ([]*Result, error) {
	// Use composite index for efficient filtering and sorting
	query := datastore.NewQuery("Result").
		Filter("language_id =", languageID).
		Filter("mode =", mode).
		Order("-composite_score").
		Order("-created_at").
		Limit(limit)

	// Add time window filter if specified
	if timeWindow != "" {
		cutoff := getTimeWindowCutoff(timeWindow)
		query = query.Filter("created_at >=", cutoff)
	}

	var results []*Result
	_, err := qo.client.GetAll(ctx, query, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetSessionEventsOptimized retrieves session events efficiently
func (qo *QueryOptimizer) GetSessionEventsOptimized(
	ctx context.Context,
	sessionID string,
	errorsOnly bool,
) ([]*SessionEvent, error) {
	query := datastore.NewQuery("SessionEvent").
		Filter("session_id =", sessionID).
		Order("timestamp_ms")

	if errorsOnly {
		query = query.Filter("error_flag =", true)
	}

	var events []*SessionEvent
	_, err := qo.client.GetAll(ctx, query, &events)
	if err != nil {
		return nil, err
	}

	return events, nil
}

// BatchGetByKeys retrieves multiple entities by keys efficiently
func (qo *QueryOptimizer) BatchGetByKeys(
	ctx context.Context,
	keys []*datastore.Key,
	dst interface{},
) error {
	// Datastore batch get is limited to 1000 entities
	const batchSize = 1000

	if len(keys) <= batchSize {
		return qo.client.GetMulti(ctx, keys, dst)
	}

	// Split into batches
	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}

		batch := keys[i:end]
		if err := qo.client.GetMulti(ctx, batch, dst); err != nil {
			return err
		}
	}

	return nil
}

// CountEntities counts entities matching a query (uses keys-only for efficiency)
func (qo *QueryOptimizer) CountEntities(
	ctx context.Context,
	kind string,
	filters map[string]interface{},
) (int, error) {
	query := datastore.NewQuery(kind).KeysOnly()

	for field, value := range filters {
		query = query.Filter(field, value)
	}

	keys, err := qo.client.GetAll(ctx, query, nil)
	if err != nil {
		return 0, err
	}

	return len(keys), nil
}

// GetRecentResults retrieves recent results with projection (memory efficient)
func (qo *QueryOptimizer) GetRecentResults(
	ctx context.Context,
	userID string,
	limit int,
	fields []string,
) ([]map[string]interface{}, error) {
	query := datastore.NewQuery("Result").
		Filter("user_id =", userID).
		Order("-created_at").
		Limit(limit)

	// Use projection to fetch only required fields
	for _, field := range fields {
		query = query.Project(field)
	}

	var results []map[string]interface{}
	it := qo.client.Run(ctx, query)

	for {
		var result map[string]interface{}
		_, err := it.Next(&result)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// DeleteOldEntities deletes entities older than specified duration
func (qo *QueryOptimizer) DeleteOldEntities(
	ctx context.Context,
	kind string,
	olderThan time.Duration,
) (int, error) {
	cutoff := time.Now().Add(-olderThan)

	query := datastore.NewQuery(kind).
		Filter("created_at <", cutoff).
		KeysOnly().
		Limit(500) // Delete in batches

	var deleted int
	for {
		keys, err := qo.client.GetAll(ctx, query, nil)
		if err != nil {
			return deleted, err
		}

		if len(keys) == 0 {
			break
		}

		if err := qo.client.DeleteMulti(ctx, keys); err != nil {
			return deleted, err
		}

		deleted += len(keys)

		// If we got fewer than limit, we're done
		if len(keys) < 500 {
			break
		}
	}

	return deleted, nil
}

// AggregateResults performs aggregation on results
func (qo *QueryOptimizer) AggregateResults(
	ctx context.Context,
	userID string,
	startDate time.Time,
	endDate time.Time,
) (map[string]float64, error) {
	query := datastore.NewQuery("Result").
		Filter("user_id =", userID).
		Filter("created_at >=", startDate).
		Filter("created_at <", endDate)

	var results []*Result
	_, err := qo.client.GetAll(ctx, query, &results)
	if err != nil {
		return nil, err
	}

	// Calculate aggregates
	aggregates := map[string]float64{
		"total_sessions": float64(len(results)),
		"avg_cpm":        0,
		"avg_accuracy":   0,
		"avg_score":      0,
		"max_score":      0,
	}

	if len(results) == 0 {
		return aggregates, nil
	}

	var totalCPM, totalAccuracy, totalScore float64
	var maxScore float64

	for _, result := range results {
		totalCPM += result.CPM
		totalAccuracy += result.RawAccuracy
		totalScore += float64(result.CompositeScore)

		if float64(result.CompositeScore) > maxScore {
			maxScore = float64(result.CompositeScore)
		}
	}

	count := float64(len(results))
	aggregates["avg_cpm"] = totalCPM / count
	aggregates["avg_accuracy"] = totalAccuracy / count
	aggregates["avg_score"] = totalScore / count
	aggregates["max_score"] = maxScore

	return aggregates, nil
}

// Helper function to get time window cutoff
func getTimeWindowCutoff(window string) time.Time {
	now := time.Now()
	switch window {
	case "daily":
		return now.AddDate(0, 0, -1)
	case "weekly":
		return now.AddDate(0, 0, -7)
	case "monthly":
		return now.AddDate(0, -1, 0)
	case "yearly":
		return now.AddDate(-1, 0, 0)
	default:
		return now.AddDate(0, 0, -7) // Default to weekly
	}
}

// ArchiveOldSessionData orchestrates the deletion of old session-related data
func (qo *QueryOptimizer) ArchiveOldSessionData(ctx context.Context) error {
	// Define retention policies
	sessionRetention := 90 * 24 * time.Hour      // 90 days for sessions
	sessionEventRetention := 90 * 24 * time.Hour // 90 days for session events
	resultRetention := 365 * 24 * time.Hour      // 1 year for results

	// Delete old sessions
	deletedSessions, err := qo.DeleteOldEntities(ctx, "Session", sessionRetention)
	if err != nil {
		return fmt.Errorf("failed to delete old sessions: %w", err)
	}
	fmt.Printf("Archived %d old sessions.\n", deletedSessions)

	// Delete old session events
	deletedEvents, err := qo.DeleteOldEntities(ctx, "SessionEvent", sessionEventRetention)
	if err != nil {
		return fmt.Errorf("failed to delete old session events: %w", err)
	}
	fmt.Printf("Archived %d old session events.\n", deletedEvents)

	// Delete old results
	deletedResults, err := qo.DeleteOldEntities(ctx, "Result", resultRetention)
	if err != nil {
		return fmt.Errorf("failed to delete old results: %w", err)
	}
	fmt.Printf("Archived %d old results.\n", deletedResults)

	return nil
}

// Session represents a typing session
type Session struct {
	ID         string                 `datastore:"-"`
	UserID     string                 `datastore:"user_id"`
	Mode       string                 `datastore:"mode"`
	LanguageID string                 `datastore:"language_id"`
	CreatedAt  time.Time              `datastore:"created_at"`
	Settings   map[string]interface{} `datastore:"settings"`
}

// Result represents session results
type Result struct {
	SessionID      string                 `datastore:"-"`
	CPM            float64                `datastore:"cpm"`
	RawAccuracy    float64                `datastore:"raw_accuracy"`
	CompositeScore int                    `datastore:"composite_score"`
	CreatedAt      time.Time              `datastore:"created_at"`
	LanguageID     string                 `datastore:"language_id"`
	Mode           string                 `datastore:"mode"`
	Breakdown      map[string]interface{} `datastore:"breakdown"`
}

// SessionEvent represents a keystroke event
type SessionEvent struct {
	ID          string    `datastore:"-"`
	SessionID   string    `datastore:"session_id"`
	TimestampMs int64     `datastore:"timestamp_ms"`
	ErrorFlag   bool      `datastore:"error_flag"`
	CreatedAt   time.Time `datastore:"created_at"`
}
