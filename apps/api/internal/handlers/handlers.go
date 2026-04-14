package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/typing-master-for-coding-backend/internal/cache"
	"github.com/typing-master-for-coding-backend/internal/database"
	"github.com/typing-master-for-coding-backend/internal/models"
	"github.com/typing-master-for-coding-backend/internal/scoring"
	"github.com/typing-master-for-coding-backend/internal/utils"

	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "backend/backend",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	})
}

// Authentication and user management handlers are now in auth.go
// Privacy and GDPR compliance handlers are now in privacy.go

func GetLanguages(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		query := datastore.NewQuery("Language")

		var languages []*models.Language
		keys, err := db.GetAll(ctx, query, &languages)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch languages"})
			return
		}

		for i, key := range keys {
			languages[i].ID = key.Name
		}

		c.JSON(http.StatusOK, languages)
	}
}

func GetLanguage(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		key := datastore.NameKey("Language", id, nil)

		ctx := context.Background()
		var language models.Language
		err := db.Get(ctx, key, &language)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Language not found"})
			return
		}
		language.ID = id

		c.JSON(http.StatusOK, language)
	}
}

func CreateLanguage(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var language models.Language
		if err := c.ShouldBindJSON(&language); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate required fields
		if language.ID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
			return
		}
		if language.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
			return
		}
		if language.ParserID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ParserID is required"})
			return
		}

		language.CreatedAt = utils.GetCurrentUTCTime()
		key := datastore.NameKey("Language", language.ID, nil)

		ctx := context.Background()
		_, err := db.Put(ctx, key, &language)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create language"})
			return
		}

		c.JSON(http.StatusCreated, language)
	}
}

func GetLessons(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		query := datastore.NewQuery("Lesson")

		var lessons []*models.Lesson
		keys, err := db.GetAll(ctx, query, &lessons)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch lessons"})
			return
		}

		for i, key := range keys {
			lessons[i].ID = key.Name
		}

		c.JSON(http.StatusOK, lessons)
	}
}

func GetLesson(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		key := datastore.NameKey("Lesson", id, nil)

		ctx := context.Background()
		var lesson models.Lesson
		err := db.Get(ctx, key, &lesson)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Lesson not found"})
			return
		}
		lesson.ID = id

		c.JSON(http.StatusOK, lesson)
	}
}

func GetSnippets(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		query := datastore.NewQuery("Snippet")

		var snippets []*models.Snippet
		keys, err := db.GetAll(ctx, query, &snippets)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch snippets"})
			return
		}

		for i, key := range keys {
			snippets[i].ID = key.Name
		}

		c.JSON(http.StatusOK, snippets)
	}
}

func GetSnippet(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		key := datastore.NameKey("Snippet", id, nil)

		ctx := context.Background()
		var snippet models.Snippet
		err := db.Get(ctx, key, &snippet)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Snippet not found"})
			return
		}
		snippet.ID = id

		c.JSON(http.StatusOK, snippet)
	}
}

func GetPublicLessons(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		// For now, return all lessons - in production, filter by public flag
		query := datastore.NewQuery("Lesson")

		var lessons []*models.Lesson
		keys, err := db.GetAll(ctx, query, &lessons)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch lessons"})
			return
		}

		for i, key := range keys {
			lessons[i].ID = key.Name
		}

		c.JSON(http.StatusOK, lessons)
	}
}

func GetPublicSnippets(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		// For now, return all snippets - in production, filter by public flag
		query := datastore.NewQuery("Snippet")

		var snippets []*models.Snippet
		keys, err := db.GetAll(ctx, query, &snippets)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch snippets"})
			return
		}

		for i, key := range keys {
			snippets[i].ID = key.Name
		}

		c.JSON(http.StatusOK, snippets)
	}
}

func UpdateLanguage(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		key := datastore.NameKey("Language", id, nil)

		ctx := context.Background()
		var existing models.Language
		err := db.Get(ctx, key, &existing)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Language not found"})
			return
		}

		var updates models.Language
		if err := c.ShouldBindJSON(&updates); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Update fields
		existing.Version = updates.Version
		existing.ParserID = updates.ParserID
		existing.GrammarConfig = updates.GrammarConfig
		existing.WhitespaceRules = updates.WhitespaceRules

		_, err = db.Put(ctx, key, &existing)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update language"})
			return
		}

		c.JSON(http.StatusOK, existing)
	}
}

func DeleteLanguage(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		key := datastore.NameKey("Language", id, nil)

		ctx := context.Background()
		err := db.Delete(ctx, key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete language"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Language deleted successfully"})
	}
}

// Lesson CRUD operations
func CreateLesson(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var lesson models.Lesson
		if err := c.ShouldBindJSON(&lesson); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		lesson.CreatedAt = utils.GetCurrentUTCTime()
		key := datastore.NameKey("Lesson", fmt.Sprintf("%s_%d", lesson.LanguageID, utils.GetCurrentTimestamp()), nil)

		ctx := context.Background()
		_, err := db.Put(ctx, key, &lesson)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create lesson"})
			return
		}

		lesson.ID = key.Name
		c.JSON(http.StatusCreated, lesson)
	}
}

func UpdateLesson(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		key := datastore.NameKey("Lesson", id, nil)

		ctx := context.Background()
		var existing models.Lesson
		err := db.Get(ctx, key, &existing)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Lesson not found"})
			return
		}

		var updates models.Lesson
		if err := c.ShouldBindJSON(&updates); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Update fields
		existing.Title = updates.Title
		existing.Difficulty = updates.Difficulty
		existing.Objectives = updates.Objectives
		existing.Prerequisites = updates.Prerequisites
		existing.EstimatedMinutes = updates.EstimatedMinutes
		existing.TokensCovered = updates.TokensCovered
		existing.SnippetIDs = updates.SnippetIDs
		existing.Version++

		_, err = db.Put(ctx, key, &existing)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update lesson"})
			return
		}

		c.JSON(http.StatusOK, existing)
	}
}

func DeleteLesson(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		key := datastore.NameKey("Lesson", id, nil)

		ctx := context.Background()
		err := db.Delete(ctx, key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete lesson"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Lesson deleted successfully"})
	}
}

// Snippet CRUD operations
func CreateSnippet(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var snippet models.Snippet
		if err := c.ShouldBindJSON(&snippet); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Generate checksum for validation
		snippet.Checksum = utils.GenerateChecksum(snippet.SourceCode)
		snippet.CreatedAt = utils.GetCurrentUTCTime()

		key := datastore.NameKey("Snippet", fmt.Sprintf("%s_%d", snippet.LanguageID, utils.GetCurrentTimestamp()), nil)

		ctx := context.Background()
		_, err := db.Put(ctx, key, &snippet)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create snippet"})
			return
		}

		snippet.ID = key.Name
		c.JSON(http.StatusCreated, snippet)
	}
}

func UpdateSnippet(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		key := datastore.NameKey("Snippet", id, nil)

		ctx := context.Background()
		var existing models.Snippet
		err := db.Get(ctx, key, &existing)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Snippet not found"})
			return
		}

		var updates models.Snippet
		if err := c.ShouldBindJSON(&updates); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Update fields and regenerate checksum if source code changed
		existing.Title = updates.Title
		existing.SourceCode = updates.SourceCode
		existing.Tags = updates.Tags
		existing.Difficulty = updates.Difficulty
		existing.EstimatedTime = updates.EstimatedTime
		existing.AccessibilityTags = updates.AccessibilityTags

		if existing.SourceCode != updates.SourceCode {
			existing.Checksum = utils.GenerateChecksum(updates.SourceCode)
		}

		_, err = db.Put(ctx, key, &existing)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update snippet"})
			return
		}

		c.JSON(http.StatusOK, existing)
	}
}

func DeleteSnippet(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		key := datastore.NameKey("Snippet", id, nil)

		ctx := context.Background()
		err := db.Delete(ctx, key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete snippet"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Snippet deleted successfully"})
	}
}

func CreateAnonymousTypingSession(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		var req struct {
			Mode       string                 `json:"mode" binding:"required"`
			LanguageID string                 `json:"language_id" binding:"required"`
			LessonID   string                 `json:"lesson_id"`
			SnippetID  string                 `json:"snippet_id"`
			Settings   map[string]interface{} `json:"settings"`
			DeviceID   string                 `json:"device_id"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Generate anonymous user ID if not provided
		if req.DeviceID == "" {
			req.DeviceID = fmt.Sprintf("device_%d", time.Now().UnixNano())
		}
		userID := fmt.Sprintf("anon_%s", req.DeviceID)

		// Create session
		session := models.Session{
			UserID:     userID,
			Mode:       req.Mode,
			LanguageID: req.LanguageID,
			LessonID:   req.LessonID,
			SnippetID:  req.SnippetID,
			StartedAt:  time.Now(),
			Settings:   req.Settings,
			CreatedAt:  time.Now(),
		}

		sessionID := fmt.Sprintf("session_%s_%d", userID, time.Now().UnixNano())
		key := datastore.NameKey("Session", sessionID, nil)

		_, err := db.Put(ctx, key, &session)
		if err != nil {
			log.Printf("Failed to create anonymous session: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create anonymous session"})
			return
		}

		session.ID = sessionID
		c.JSON(http.StatusCreated, session)
	}
}

func CreateSession(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		userID := c.GetString("user_id")

		var req struct {
			Mode       string                 `json:"mode" binding:"required"`
			LanguageID string                 `json:"language_id" binding:"required"`
			LessonID   string                 `json:"lesson_id"`
			SnippetID  string                 `json:"snippet_id"`
			Settings   map[string]interface{} `json:"settings"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Create session
		session := models.Session{
			UserID:     userID,
			Mode:       req.Mode,
			LanguageID: req.LanguageID,
			LessonID:   req.LessonID,
			SnippetID:  req.SnippetID,
			StartedAt:  time.Now(),
			Settings:   req.Settings,
			CreatedAt:  time.Now(),
		}

		sessionID := fmt.Sprintf("session_%s_%d", userID, time.Now().UnixNano())
		key := datastore.NameKey("Session", sessionID, nil)

		_, err := db.Put(ctx, key, &session)
		if err != nil {
			log.Printf("Failed to create session: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
			return
		}

		session.ID = sessionID
		c.JSON(http.StatusCreated, session)
	}
}

func UpdateSession(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		sessionID := c.Param("id")

		var updates struct {
			DurationMs int64                  `json:"duration_ms"`
			Settings   map[string]interface{} `json:"settings"`
		}

		if err := c.ShouldBindJSON(&updates); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get existing session
		key := datastore.NameKey("Session", sessionID, nil)
		var session models.Session
		err := db.Get(ctx, key, &session)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			return
		}

		// Update fields
		if updates.DurationMs > 0 {
			session.DurationMs = updates.DurationMs
		}
		if updates.Settings != nil {
			session.Settings = updates.Settings
		}

		// Save updated session
		_, err = db.Put(ctx, key, &session)
		if err != nil {
			log.Printf("Failed to update session: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session"})
			return
		}

		session.ID = sessionID
		c.JSON(http.StatusOK, session)
	}
}

func RecordEvents(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		sessionID := c.Param("id")

		var events []models.SessionEvent
		if err := c.ShouldBindJSON(&events); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verify session exists (check cache first)
		var session models.Session
		if cachedSession, found := cache.Get(sessionID); found {
			session = cachedSession.(models.Session)
		} else {
			sessionKey := datastore.NameKey("Session", sessionID, nil)
			err := db.Get(ctx, sessionKey, &session)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
				return
			}
			cache.Set(sessionID, session) // Cache the session for future requests
		}

		// Store events in batch
		keys := make([]*datastore.Key, len(events))
		entities := make([]interface{}, len(events))

		baseTimestamp := time.Now().UnixNano()
		for i, event := range events {
			event.SessionID = sessionID
			event.CreatedAt = time.Now()
			eventID := fmt.Sprintf("%s_event_%d", sessionID, baseTimestamp+int64(i))
			keys[i] = datastore.NameKey("SessionEvent", eventID, nil)
			entities[i] = &event
		}

		var putMultiErr error
		_, putMultiErr = db.PutMulti(ctx, keys, entities)
		if putMultiErr != nil {
			log.Printf("Failed to record events: %v", putMultiErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record events"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Events recorded successfully", "count": len(events)})
	}
}

func FinalizeSession(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		sessionID := c.Param("id")

		// Get session
		sessionKey := datastore.NameKey("Session", sessionID, nil)
		var session models.Session
		err := db.Get(ctx, sessionKey, &session)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			return
		}

		// Mark session as ended
		endTime := time.Now()
		session.EndedAt = &endTime
		if session.DurationMs == 0 {
			session.DurationMs = endTime.Sub(session.StartedAt).Milliseconds()
		}

		// Save updated session
		_, err = db.Put(ctx, sessionKey, &session)
		if err != nil {
			log.Printf("Failed to finalize session: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finalize session"})
			return
		}

		// Process scoring if service is initialized
		if scoringService != nil {
			go func() {
				// Process asynchronously to avoid blocking response
				metrics, err := scoringService.ProcessSession(context.Background(), sessionID)
				if err != nil {
					log.Printf("Failed to process session metrics: %v", err)
					return
				}

				// Update leaderboards
				err = scoringService.UpdateLeaderboard(context.Background(), sessionID)
				if err != nil {
					log.Printf("Failed to update leaderboards: %v", err)
				}

				log.Printf("Session %s processed: Score=%d, CPM=%.2f, TWPM=%.2f",
					sessionID, metrics.CompositeScore, metrics.CPM, metrics.TWPM)
			}()
		}

		session.ID = sessionID
		c.JSON(http.StatusOK, gin.H{
			"message": "Session finalized successfully",
			"session": session,
		})
	}
}

func GetResults(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		userID := c.GetString("user_id")

		// Query parameters
		limit := 50
		if limitStr := c.Query("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
				limit = l
			}
		}

		languageID := c.Query("language_id")
		mode := c.Query("mode")

		// Get user's sessions first to filter results
		sessionQuery := datastore.NewQuery("Session").FilterField("UserID", "=", userID)
		if languageID != "" {
			sessionQuery = sessionQuery.FilterField("LanguageID", "=", languageID)
		}
		if mode != "" {
			sessionQuery = sessionQuery.FilterField("Mode", "=", mode)
		}
		sessionQuery = sessionQuery.Order("-CreatedAt").Limit(limit)

		var sessions []*models.Session
		sessionKeys, err := db.GetAll(ctx, sessionQuery, &sessions)
		if err != nil {
			log.Printf("Failed to fetch sessions: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch results"})
			return
		}

		// Get results for these sessions
		results := make([]models.Result, 0)
		for i := range sessions {
			sessionID := sessionKeys[i].Name
			resultKey := datastore.NameKey("Result", sessionID, nil)
			var result models.Result
			err := db.Get(ctx, resultKey, &result)
			if err == nil {
				result.SessionID = sessionID
				results = append(results, result)
			}
		}

		c.JSON(http.StatusOK, results)
	}
}

func GetResult(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		sessionID := c.Param("session_id")

		// Get result
		key := datastore.NameKey("Result", sessionID, nil)
		var result models.Result
		err := db.Get(ctx, key, &result)
		if err != nil {
			if err == datastore.ErrNoSuchEntity {
				c.JSON(http.StatusNotFound, gin.H{"error": "Result not found"})
				return
			}
			log.Printf("Failed to fetch result: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch result"})
			return
		}

		result.SessionID = sessionID

		// Also get session details
		sessionKey := datastore.NameKey("Session", sessionID, nil)
		var session models.Session
		err = db.Get(ctx, sessionKey, &session)
		if err == nil {
			session.ID = sessionID
		}

		c.JSON(http.StatusOK, gin.H{
			"result":  result,
			"session": session,
		})
	}
}

func GetLeaderboards(cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use the leaderboard service if initialized
		if leaderboardService == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Leaderboard service not initialized"})
			return
		}

		languageID := c.Query("language_id")
		if languageID == "" {
			languageID = "python"
		}

		mode := c.Query("mode")
		if mode == "" {
			mode = "timed"
		}

		scope := c.Query("scope")
		if scope == "" {
			scope = "global"
		}

		timeWindow := c.Query("time_window")
		if timeWindow == "" {
			timeWindow = "weekly"
		}

		limitStr := c.Query("limit")
		limit := 50
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
				limit = l
			}
		}

		req := &scoring.LeaderboardRequest{
			LanguageID: languageID,
			Mode:       mode,
			Scope:      scope,
			TimeWindow: timeWindow,
			Limit:      limit,
		}

		ctx := context.Background()
		response, err := leaderboardService.GetLeaderboard(ctx, req)
		if err != nil {
			log.Printf("Failed to get leaderboard: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve leaderboard"})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

func CreatePlaylist(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var playlist models.Playlist
		if err := c.ShouldBindJSON(&playlist); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		playlist.CreatedAt = time.Now()
		playlist.UpdatedAt = time.Now()
		key := datastore.NameKey("Playlist", fmt.Sprintf("%s_%d", playlist.UserID, time.Now().Unix()), nil)

		ctx := context.Background()
		_, err := db.Put(ctx, key, &playlist)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create playlist"})
			return
		}

		playlist.ID = key.Name
		c.JSON(http.StatusCreated, playlist)
	}
}

func GetPlaylists(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		query := datastore.NewQuery("Playlist")

		var playlists []*models.Playlist
		keys, err := db.GetAll(ctx, query, &playlists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch playlists"})
			return
		}

		for i, key := range keys {
			playlists[i].ID = key.Name
		}

		c.JSON(http.StatusOK, playlists)
	}
}

func GetPlaylist(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		key := datastore.NameKey("Playlist", id, nil)

		ctx := context.Background()
		var playlist models.Playlist
		err := db.Get(ctx, key, &playlist)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
			return
		}
		playlist.ID = id

		c.JSON(http.StatusOK, playlist)
	}
}

func UpdatePlaylist(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		key := datastore.NameKey("Playlist", id, nil)

		ctx := context.Background()
		var existing models.Playlist
		err := db.Get(ctx, key, &existing)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
			return
		}

		var updates models.Playlist
		if err := c.ShouldBindJSON(&updates); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Update fields
		existing.Title = updates.Title
		existing.Description = updates.Description
		existing.LessonIDs = updates.LessonIDs
		existing.SnippetIDs = updates.SnippetIDs
		existing.Tags = updates.Tags
		existing.IsPublic = updates.IsPublic
		existing.Difficulty = updates.Difficulty
		existing.Version++
		existing.UpdatedAt = time.Now()

		_, err = db.Put(ctx, key, &existing)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update playlist"})
			return
		}

		c.JSON(http.StatusOK, existing)
	}
}

func DeletePlaylist(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		key := datastore.NameKey("Playlist", id, nil)

		ctx := context.Background()
		err := db.Delete(ctx, key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete playlist"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Playlist deleted successfully"})
	}
}

// Content versioning and validation handlers

// Create content version for audit trail
func CreateContentVersion(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var version models.ContentVersion
		if err := c.ShouldBindJSON(&version); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Generate checksum for the content
		contentJSON, _ := json.Marshal(version.Content)
		version.Checksum = utils.GenerateChecksum(string(contentJSON))
		version.CreatedAt = time.Now()

		key := datastore.NameKey("ContentVersion", fmt.Sprintf("%s_%s_%d", version.ContentType, version.ContentID, version.Version), nil)

		ctx := context.Background()
		_, err := db.Put(ctx, key, &version)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create content version"})
			return
		}

		version.ID = key.Name
		c.JSON(http.StatusCreated, version)
	}
}

// Get content versions for a specific content item
func GetContentVersions(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		contentType := c.Param("contentType")
		contentID := c.Param("contentId")

		ctx := context.Background()
		query := datastore.NewQuery("ContentVersion").
			FilterField("ContentType", "=", contentType).
			FilterField("ContentID", "=", contentID).
			Order("-Version")

		var versions []*models.ContentVersion
		keys, err := db.GetAll(ctx, query, &versions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch content versions"})
			return
		}

		for i, key := range keys {
			versions[i].ID = key.Name
		}

		c.JSON(http.StatusOK, versions)
	}
}

// Validate content checksum
func ValidateContentChecksum(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var validation models.ContentValidation
		if err := c.ShouldBindJSON(&validation); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validation.ValidatedAt = time.Now()

		// For now, mark as valid - in production, implement actual validation logic
		validation.ValidationStatus = "valid"
		validation.ValidationErrors = []string{}

		key := datastore.NameKey("ContentValidation", fmt.Sprintf("%s_%s_%d", validation.ContentType, validation.ContentID, time.Now().Unix()), nil)

		ctx := context.Background()
		_, err := db.Put(ctx, key, &validation)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save validation result"})
			return
		}

		validation.ID = key.Name
		c.JSON(http.StatusOK, validation)
	}
}

// Import snippet with code normalization
func ImportSnippet(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var importRequest struct {
			LanguageID string   `json:"language_id" binding:"required"`
			Title      string   `json:"title" binding:"required"`
			SourceCode string   `json:"source_code" binding:"required"`
			Tags       []string `json:"tags"`
			Difficulty int      `json:"difficulty"`
		}

		if err := c.ShouldBindJSON(&importRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Normalize the source code (basic normalization)
		normalizedCode := utils.NormalizeCode(importRequest.SourceCode)

		// Auto-assess difficulty if not provided
		difficulty := importRequest.Difficulty
		if difficulty == 0 {
			difficulty = utils.AssessCodeDifficulty(normalizedCode)
		}

		// Auto-generate tags if not provided
		tags := importRequest.Tags
		if len(tags) == 0 {
			tags = utils.GenerateTagsFromCode(normalizedCode)
		}

		snippet := models.Snippet{
			LanguageID:        importRequest.LanguageID,
			Title:             importRequest.Title,
			SourceCode:        normalizedCode,
			Tags:              tags,
			Difficulty:        difficulty,
			EstimatedTime:     utils.EstimateCodingTime(normalizedCode),
			Checksum:          utils.GenerateChecksum(normalizedCode),
			AccessibilityTags: utils.GenerateAccessibilityTags(normalizedCode),
			CreatedAt:         time.Now(),
		}

		key := datastore.NameKey("Snippet", fmt.Sprintf("%s_%d", snippet.LanguageID, time.Now().Unix()), nil)

		ctx := context.Background()
		_, err := db.Put(ctx, key, &snippet)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to import snippet"})
			return
		}

		snippet.ID = key.Name
		c.JSON(http.StatusCreated, snippet)
	}
}

// Helper functions for content processing
// (These functions have been moved to internal/utils/content.go)

// Lesson progression system handlers

// Get lesson progression for a user
func GetLessonProgress(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		lessonID := c.Param("id")

		ctx := context.Background()
		query := datastore.NewQuery("LessonProgress").
			FilterField("UserID", "=", userID).
			FilterField("LessonID", "=", lessonID)

		var progress []*models.LessonProgress
		keys, err := db.GetAll(ctx, query, &progress)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch lesson progress"})
			return
		}

		if len(progress) == 0 {
			c.JSON(http.StatusOK, gin.H{"message": "No progress found for this lesson"})
			return
		}

		progress[0].ID = keys[0].Name
		c.JSON(http.StatusOK, progress[0])
	}
}

// Initialize or update lesson progress
func UpdateLessonProgress(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		lessonID := c.Param("id")

		var progressUpdate struct {
			CurrentStage    string   `json:"current_stage"`
			ProgressPercent float64  `json:"progress_percent"`
			TokensCovered   []string `json:"tokens_covered"`
			Score           int      `json:"score"`
		}

		if err := c.ShouldBindJSON(&progressUpdate); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx := context.Background()

		// Check if progress already exists
		query := datastore.NewQuery("LessonProgress").
			FilterField("UserID", "=", userID).
			FilterField("LessonID", "=", lessonID)

		var existingProgress []*models.LessonProgress
		keys, err := db.GetAll(ctx, query, &existingProgress)

		var progress *models.LessonProgress
		var key *datastore.Key

		if len(existingProgress) > 0 {
			// Update existing progress
			progress = existingProgress[0]
			key = keys[0]

			// Update progress fields
			if progressUpdate.CurrentStage != "" {
				progress.CurrentStage = progressUpdate.CurrentStage
			}
			if progressUpdate.ProgressPercent > 0 {
				progress.ProgressPercent = progressUpdate.ProgressPercent
			}
			if len(progressUpdate.TokensCovered) > 0 {
				progress.TokensCovered = progressUpdate.TokensCovered
			}
			if progressUpdate.Score > progress.BestScore {
				progress.BestScore = progressUpdate.Score
			}
			progress.AttemptsCount++
			progress.LastAttemptAt = time.Now()
			progress.UpdatedAt = time.Now()

			// Check if lesson is completed
			if progressUpdate.ProgressPercent >= 100 {
				progress.IsCompleted = true
				if !contains(progress.CompletedStages, progressUpdate.CurrentStage) {
					progress.CompletedStages = append(progress.CompletedStages, progressUpdate.CurrentStage)
				}
			}
		} else {
			// Create new progress
			newProgress := &models.LessonProgress{
				UserID:          userID,
				LessonID:        lessonID,
				CurrentStage:    progressUpdate.CurrentStage,
				ProgressPercent: progressUpdate.ProgressPercent,
				TokensCovered:   progressUpdate.TokensCovered,
				IsUnlocked:      true,
				BestScore:       progressUpdate.Score,
				AttemptsCount:   1,
				LastAttemptAt:   time.Now(),
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}
			progress = newProgress
			key = datastore.NameKey("LessonProgress", fmt.Sprintf("%s_%s_%d", userID, lessonID, time.Now().Unix()), nil)
		}

		_, err = db.Put(ctx, key, progress)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update lesson progress"})
			return
		}

		progress.ID = key.Name
		c.JSON(http.StatusOK, progress)
	}
}

// Check lesson prerequisites and unlock status
func CheckLessonPrerequisites(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		lessonID := c.Param("id")

		ctx := context.Background()

		// Get lesson details
		lessonKey := datastore.NameKey("Lesson", lessonID, nil)
		var lesson models.Lesson
		err := db.Get(ctx, lessonKey, &lesson)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Lesson not found"})
			return
		}

		// Check user progress for prerequisites
		var unlocked = true
		var missingPrereqs []string

		for _, prereqID := range lesson.Prerequisites {
			query := datastore.NewQuery("LessonProgress").
				FilterField("UserID", "=", userID).
				FilterField("LessonID", "=", prereqID)

			var prereqProgress []*models.LessonProgress
			_, err := db.GetAll(ctx, query, &prereqProgress)
			if err != nil || len(prereqProgress) == 0 || !prereqProgress[0].IsCompleted {
				unlocked = false
				missingPrereqs = append(missingPrereqs, prereqID)
			}
		}

		response := gin.H{
			"lesson_id":       lessonID,
			"is_unlocked":     unlocked,
			"missing_prereqs": missingPrereqs,
			"lesson":          lesson,
		}

		c.JSON(http.StatusOK, response)
	}
}

// Get lesson progression flow for a specific lesson
func GetLessonProgressionFlow(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		lessonID := c.Param("id")

		ctx := context.Background()

		// Get lesson details
		lessonKey := datastore.NameKey("Lesson", lessonID, nil)
		var lesson models.Lesson
		err := db.Get(ctx, lessonKey, &lesson)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Lesson not found"})
			return
		}

		// Define the standard progression flow
		flow := []gin.H{
			{
				"stage":             "intro",
				"description":       "Introduction to concepts",
				"estimated_minutes": lesson.EstimatedMinutes / 5,
				"tokens_to_cover":   lesson.TokensCovered[:min(2, len(lesson.TokensCovered))],
			},
			{
				"stage":             "core",
				"description":       "Core concepts and syntax",
				"estimated_minutes": lesson.EstimatedMinutes / 2,
				"tokens_to_cover":   lesson.TokensCovered[:min(4, len(lesson.TokensCovered))],
			},
			{
				"stage":             "idioms",
				"description":       "Language idioms and patterns",
				"estimated_minutes": lesson.EstimatedMinutes / 3,
				"tokens_to_cover":   lesson.TokensCovered,
			},
			{
				"stage":             "advanced",
				"description":       "Advanced concepts and edge cases",
				"estimated_minutes": lesson.EstimatedMinutes / 4,
				"tokens_to_cover":   lesson.TokensCovered,
			},
			{
				"stage":             "review",
				"description":       "Review and practice",
				"estimated_minutes": lesson.EstimatedMinutes / 5,
				"tokens_to_cover":   lesson.TokensCovered,
			},
		}

		response := gin.H{
			"lesson_id": lessonID,
			"lesson":    lesson,
			"flow":      flow,
		}

		c.JSON(http.StatusOK, response)
	}
}

// Get user's overall lesson progression summary
func GetUserProgressionSummary(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		ctx := context.Background()

		// Get all lessons
		lessonQuery := datastore.NewQuery("Lesson")
		var lessons []*models.Lesson
		lessonKeys, err := db.GetAll(ctx, lessonQuery, &lessons)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch lessons"})
			return
		}

		// Get user's progress for all lessons
		progressQuery := datastore.NewQuery("LessonProgress").
			FilterField("UserID", "=", userID)

		var progressList []*models.LessonProgress
		_, err = db.GetAll(ctx, progressQuery, &progressList)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch progress"})
			return
		}

		// Create progress map for quick lookup
		progressMap := make(map[string]*models.LessonProgress)
		for _, p := range progressList {
			progressMap[p.LessonID] = p
		}

		// Calculate summary statistics
		totalLessons := len(lessons)
		completedLessons := 0
		totalProgress := 0.0
		averageScore := 0.0
		scoreCount := 0

		for i, lesson := range lessons {
			lesson.ID = lessonKeys[i].Name
			if progress, exists := progressMap[lesson.ID]; exists {
				if progress.IsCompleted {
					completedLessons++
				}
				totalProgress += progress.ProgressPercent
				if progress.BestScore > 0 {
					averageScore += float64(progress.BestScore)
					scoreCount++
				}
			}
		}

		overallProgress := 0.0
		if totalLessons > 0 {
			overallProgress = totalProgress / float64(totalLessons)
		}

		if scoreCount > 0 {
			averageScore = averageScore / float64(scoreCount)
		}

		summary := gin.H{
			"total_lessons":     totalLessons,
			"completed_lessons": completedLessons,
			"overall_progress":  overallProgress,
			"average_score":     averageScore,
			"completion_rate":   float64(completedLessons) / float64(totalLessons) * 100,
		}

		c.JSON(http.StatusOK, summary)
	}
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Scoring and leaderboard handlers

// Global scoring service instance (would be injected via dependency injection)
var scoringService *scoring.Service
var leaderboardService scoring.LeaderboardService
var antiCheatService *scoring.AntiCheatService
var tournamentService *scoring.TournamentService

// InitializeScoringServices initializes the scoring services (called during app startup)
func InitializeScoringServices(dsClient *datastore.Client) {
	scoringService = scoring.NewService(scoring.NewDatastoreClient(dsClient))
	leaderboardService = scoring.NewLeaderboardService(scoring.NewDatastoreClient(dsClient))
	antiCheatService = scoring.NewAntiCheatService(scoring.NewDatastoreClient(dsClient))
	tournamentService = scoring.NewTournamentService(scoring.NewDatastoreClient(dsClient), *antiCheatService, leaderboardService)
}

// ProcessSessionMetrics processes and stores session metrics
func ProcessSessionMetrics(c *gin.Context) {
	var req scoring.SessionMetrics
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	ctx := context.Background()

	// Process session metrics
	metrics, err := scoringService.ProcessSession(ctx, req.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process session metrics"})
		return
	}

	// Update leaderboards
	err = scoringService.UpdateLeaderboard(ctx, req.SessionID)
	if err != nil {
		log.Printf("Failed to update leaderboards: %v", err)
		// Don't fail the request for leaderboard update issues
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Session metrics processed successfully",
		"metrics": metrics,
	})
}

// GetLeaderboard retrieves leaderboard rankings
func GetLeaderboard(c *gin.Context) {
	languageID := c.Query("language_id")
	if languageID == "" {
		languageID = "python" // Default to Python
	}

	mode := c.Query("mode")
	if mode == "" {
		mode = "timed" // Default to timed mode
	}

	scope := c.Query("scope")
	if scope == "" {
		scope = "global" // Default to global
	}

	timeWindow := c.Query("time_window")
	if timeWindow == "" {
		timeWindow = "weekly" // Default to weekly
	}

	limitStr := c.Query("limit")
	limit := 50 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	req := &scoring.LeaderboardRequest{
		LanguageID: languageID,
		Mode:       mode,
		Scope:      scope,
		TimeWindow: timeWindow,
		Limit:      limit,
	}

	ctx := context.Background()
	response, err := leaderboardService.GetLeaderboard(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve leaderboard"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetUserRank retrieves a user's rank in the leaderboard
func GetUserRank(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	languageID := c.Query("language_id")
	if languageID == "" {
		languageID = "python"
	}

	mode := c.Query("mode")
	if mode == "" {
		mode = "timed"
	}

	scope := c.Query("scope")
	if scope == "" {
		scope = "global"
	}

	timeWindow := c.Query("time_window")
	if timeWindow == "" {
		timeWindow = "weekly"
	}

	ctx := context.Background()
	rank, err := leaderboardService.GetUserRank(ctx, userID, languageID, mode, scope, timeWindow)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user rank"})
		return
	}

	c.JSON(http.StatusOK, rank)
}

// GetLeaderboardTrends retrieves leaderboard trends and analytics
func GetLeaderboardTrends(c *gin.Context) {
	languageID := c.Query("language_id")
	if languageID == "" {
		languageID = "python"
	}

	mode := c.Query("mode")
	if mode == "" {
		mode = "timed"
	}

	scope := c.Query("scope")
	if scope == "" {
		scope = "global"
	}

	timeWindow := c.Query("time_window")
	if timeWindow == "" {
		timeWindow = "weekly"
	}

	period := c.Query("period")
	if period == "" {
		period = "30d"
	}

	metric := c.Query("metric")
	if metric == "" {
		metric = "score"
	}

	req := &scoring.LeaderboardTrendsRequest{
		LanguageID: languageID,
		Mode:       mode,
		Scope:      scope,
		TimeWindow: timeWindow,
		Period:     period,
		Metric:     metric,
	}

	ctx := context.Background()
	response, err := leaderboardService.GetLeaderboardTrends(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve leaderboard trends"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// AnalyzeAntiCheat performs anti-cheat analysis on a session
func AnalyzeAntiCheat(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
		return
	}

	ctx := context.Background()
	report, err := antiCheatService.AnalyzeSession(ctx, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze session for anti-cheat"})
		return
	}

	c.JSON(http.StatusOK, report)
}

// Tournament handlers

// CreateTournament creates a new tournament
func CreateTournament(c *gin.Context) {
	var req scoring.CreateTournamentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Get user ID from JWT token (would be implemented in auth middleware)
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	req.CreatedBy = userID

	ctx := context.Background()
	tournament, err := tournamentService.CreateTournament(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tournament"})
		return
	}

	c.JSON(http.StatusCreated, tournament)
}

// GetTournament retrieves tournament information
func GetTournament(c *gin.Context) {
	tournamentID := c.Param("tournament_id")
	if tournamentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tournament ID is required"})
		return
	}

	ctx := context.Background()
	tournament, err := tournamentService.GetTournament(ctx, tournamentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
		return
	}

	c.JSON(http.StatusOK, tournament)
}

// RegisterForTournament registers a user for a tournament
func RegisterForTournament(c *gin.Context) {
	tournamentID := c.Param("tournament_id")
	if tournamentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tournament ID is required"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	ctx := context.Background()
	participant, err := tournamentService.RegisterParticipant(ctx, tournamentID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, participant)
}

// GetTournamentLeaderboard retrieves tournament leaderboard
func GetTournamentLeaderboard(c *gin.Context) {
	tournamentID := c.Param("tournament_id")
	if tournamentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tournament ID is required"})
		return
	}

	ctx := context.Background()
	leaderboard, err := tournamentService.GetTournamentLeaderboard(ctx, tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tournament leaderboard"})
		return
	}

	c.JSON(http.StatusOK, leaderboard)
}

// SubmitTournamentResult submits a tournament result
func SubmitTournamentResult(c *gin.Context) {
	tournamentID := c.Param("tournament_id")
	if tournamentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tournament ID is required"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req struct {
		SessionID string `json:"session_id" binding:"required"`
		Score     int    `json:"score" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	ctx := context.Background()
	err := tournamentService.SubmitTournamentResult(ctx, tournamentID, userID, req.SessionID, req.Score)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tournament result submitted successfully"})
}

// GetUserTournaments retrieves tournaments for a user
func GetUserTournaments(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	status := c.Query("status") // Optional status filter

	ctx := context.Background()
	tournaments, err := tournamentService.GetUserTournaments(ctx, userID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user tournaments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tournaments": tournaments,
		"count":       len(tournaments),
	})
}

// Tournament control endpoints (admin only)

// StartTournament starts a tournament
func StartTournament(c *gin.Context) {
	tournamentID := c.Param("tournament_id")
	if tournamentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tournament ID is required"})
		return
	}

	// TODO: Add admin authorization check
	ctx := context.Background()
	err := tournamentService.StartTournament(ctx, tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tournament started successfully"})
}

// EndTournament ends a tournament
func EndTournament(c *gin.Context) {
	tournamentID := c.Param("tournament_id")
	if tournamentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tournament ID is required"})
		return
	}

	// TODO: Add admin authorization check
	ctx := context.Background()
	err := tournamentService.EndTournament(ctx, tournamentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tournament ended successfully"})
}
