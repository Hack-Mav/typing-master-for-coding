package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"typing-master-backend/internal/cache"
	"typing-master-backend/internal/database"
	"typing-master-backend/internal/models"

	"github.com/gin-gonic/gin"
	"cloud.google.com/go/datastore"
)

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "typing-master-backend",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	})
}

// Helper function to generate checksum for code snippets
func generateChecksum(code string) string {
	hash := sha256.Sum256([]byte(code))
	return fmt.Sprintf("%x", hash)
}

// Placeholder handlers - will be implemented in subsequent tasks
func Register(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Register endpoint not implemented yet"})
	}
}

func Login(db *database.DatastoreClient, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Login endpoint not implemented yet"})
	}
}

func RefreshToken(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "RefreshToken endpoint not implemented yet"})
	}
}

func GetProfile(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "GetProfile endpoint not implemented yet"})
	}
}

func UpdateProfile(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "UpdateProfile endpoint not implemented yet"})
	}
}

func GetLanguages(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		query := datastore.NewQuery("Language")

		var languages []models.Language
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

		language.CreatedAt = time.Now()
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

		var lessons []models.Lesson
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

		var snippets []models.Snippet
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

		var lessons []models.Lesson
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

		var snippets []models.Snippet
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

		lesson.CreatedAt = time.Now()
		key := datastore.NameKey("Lesson", fmt.Sprintf("%s_%d", lesson.LanguageID, time.Now().Unix()), nil)

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
		snippet.Checksum = generateChecksum(snippet.SourceCode)
		snippet.CreatedAt = time.Now()

		key := datastore.NameKey("Snippet", fmt.Sprintf("%s_%d", snippet.LanguageID, time.Now().Unix()), nil)

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
			existing.Checksum = generateChecksum(updates.SourceCode)
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

func CreateSession(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "CreateSession endpoint not implemented yet"})
	}
}

func UpdateSession(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "UpdateSession endpoint not implemented yet"})
	}
}

func RecordEvents(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "RecordEvents endpoint not implemented yet"})
	}
}

func FinalizeSession(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "FinalizeSession endpoint not implemented yet"})
	}
}

func GetResults(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "GetResults endpoint not implemented yet"})
	}
}

func GetResult(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "GetResult endpoint not implemented yet"})
	}
}

func GetLeaderboards(cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "GetLeaderboards endpoint not implemented yet"})
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

		var playlists []models.Playlist
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
		version.Checksum = generateChecksum(string(contentJSON))
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

		var versions []models.ContentVersion
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
			LanguageID  string `json:"language_id" binding:"required"`
			Title       string `json:"title" binding:"required"`
			SourceCode  string `json:"source_code" binding:"required"`
			Tags        []string `json:"tags"`
			Difficulty  int    `json:"difficulty"`
		}

		if err := c.ShouldBindJSON(&importRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Normalize the source code (basic normalization)
		normalizedCode := normalizeCode(importRequest.SourceCode)

		// Auto-assess difficulty if not provided
		difficulty := importRequest.Difficulty
		if difficulty == 0 {
			difficulty = assessCodeDifficulty(normalizedCode)
		}

		// Auto-generate tags if not provided
		tags := importRequest.Tags
		if len(tags) == 0 {
			tags = generateTagsFromCode(normalizedCode)
		}

		snippet := models.Snippet{
			LanguageID:        importRequest.LanguageID,
			Title:             importRequest.Title,
			SourceCode:        normalizedCode,
			Tags:              tags,
			Difficulty:        difficulty,
			EstimatedTime:     estimateCodingTime(normalizedCode),
			Checksum:          generateChecksum(normalizedCode),
			AccessibilityTags: generateAccessibilityTags(normalizedCode),
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

func normalizeCode(code string) string {
	// Basic normalization: trim whitespace, normalize line endings
	lines := strings.Split(strings.TrimSpace(code), "\n")
	var normalized []string
	for _, line := range lines {
		normalized = append(normalized, strings.TrimSpace(line))
	}
	return strings.Join(normalized, "\n")
}

func assessCodeDifficulty(code string) int {
	lines := strings.Split(code, "\n")
	lineCount := len(lines)

	// Simple heuristic based on line count and complexity indicators
	if lineCount <= 10 {
		return 1 // Easy
	} else if lineCount <= 30 {
		return 2 // Medium
	} else if lineCount <= 60 {
		return 3 // Hard
	} else {
		return 4 // Expert
	}
}

func generateTagsFromCode(code string) []string {
	var tags []string

	// Simple tag generation based on code patterns
	if strings.Contains(code, "function") || strings.Contains(code, "def ") {
		tags = append(tags, "function")
	}
	if strings.Contains(code, "class") {
		tags = append(tags, "class")
	}
	if strings.Contains(code, "if ") || strings.Contains(code, "if(") {
		tags = append(tags, "conditional")
	}
	if strings.Contains(code, "for ") || strings.Contains(code, "for(") {
		tags = append(tags, "loop")
	}
	if strings.Contains(code, "import") || strings.Contains(code, "from ") {
		tags = append(tags, "import")
	}

	// Default tag if no patterns found
	if len(tags) == 0 {
		tags = append(tags, "general")
	}

	return tags
}

func estimateCodingTime(code string) int {
	lines := strings.Split(code, "\n")
	lineCount := len(lines)

	// Simple estimation: ~1 minute per 10 lines, minimum 1 minute
	estimatedMinutes := lineCount / 10
	if estimatedMinutes < 1 {
		estimatedMinutes = 1
	}

	return estimatedMinutes
}

func generateAccessibilityTags(code string) map[string]interface{} {
	accessibility := make(map[string]interface{})

	// Basic accessibility assessment
	lines := strings.Split(code, "\n")
	lineCount := len(lines)

	accessibility["line_count"] = lineCount
	accessibility["estimated_time"] = estimateCodingTime(code)
	accessibility["complexity_score"] = assessCodeDifficulty(code)

	// Check for potential accessibility issues
	if lineCount > 50 {
		accessibility["long_content"] = true
	}

	return accessibility
}

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

		var progress []models.LessonProgress
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
			CurrentStage    string  `json:"current_stage"`
			ProgressPercent float64 `json:"progress_percent"`
			TokensCovered   []string `json:"tokens_covered"`
			Score           int     `json:"score"`
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

		var existingProgress []models.LessonProgress
		keys, err := db.GetAll(ctx, query, &existingProgress)

		var progress models.LessonProgress
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
			progress = models.LessonProgress{
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
			key = datastore.NameKey("LessonProgress", fmt.Sprintf("%s_%s_%d", userID, lessonID, time.Now().Unix()), nil)
		}

		_, err = db.Put(ctx, key, &progress)
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

			var prereqProgress []models.LessonProgress
			_, err := db.GetAll(ctx, query, &prereqProgress)
			if err != nil || len(prereqProgress) == 0 || !prereqProgress[0].IsCompleted {
				unlocked = false
				missingPrereqs = append(missingPrereqs, prereqID)
			}
		}

		response := gin.H{
			"lesson_id":     lessonID,
			"is_unlocked":   unlocked,
			"missing_prereqs": missingPrereqs,
			"lesson":        lesson,
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
				"stage":       "intro",
				"description": "Introduction to concepts",
				"estimated_minutes": lesson.EstimatedMinutes / 5,
				"tokens_to_cover": lesson.TokensCovered[:min(2, len(lesson.TokensCovered))],
			},
			{
				"stage":       "core",
				"description": "Core concepts and syntax",
				"estimated_minutes": lesson.EstimatedMinutes / 2,
				"tokens_to_cover": lesson.TokensCovered[:min(4, len(lesson.TokensCovered))],
			},
			{
				"stage":       "idioms",
				"description": "Language idioms and patterns",
				"estimated_minutes": lesson.EstimatedMinutes / 3,
				"tokens_to_cover": lesson.TokensCovered,
			},
			{
				"stage":       "advanced",
				"description": "Advanced concepts and edge cases",
				"estimated_minutes": lesson.EstimatedMinutes / 4,
				"tokens_to_cover": lesson.TokensCovered,
			},
			{
				"stage":       "review",
				"description": "Review and practice",
				"estimated_minutes": lesson.EstimatedMinutes / 5,
				"tokens_to_cover": lesson.TokensCovered,
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
		var lessons []models.Lesson
		lessonKeys, err := db.GetAll(ctx, lessonQuery, &lessons)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch lessons"})
			return
		}

		// Get user's progress for all lessons
		progressQuery := datastore.NewQuery("LessonProgress").
			FilterField("UserID", "=", userID)

		var progressList []models.LessonProgress
		_, err = db.GetAll(ctx, progressQuery, &progressList)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch progress"})
			return
		}

		// Create progress map for quick lookup
		progressMap := make(map[string]models.LessonProgress)
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
			"total_lessons":      totalLessons,
			"completed_lessons":  completedLessons,
			"overall_progress":   overallProgress,
			"average_score":      averageScore,
			"completion_rate":    float64(completedLessons) / float64(totalLessons) * 100,
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}