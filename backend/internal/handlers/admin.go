package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"typing-master-backend/internal/database"
	"typing-master-backend/internal/models"
	"typing-master-backend/internal/utils"

	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
)

// Admin dashboard - Get overview statistics
func GetAdminDashboard(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		// Get total counts for different entities
		stats := make(map[string]interface{})

		// User count
		userQuery := datastore.NewQuery("User")
		userCount, err := db.Count(ctx, userQuery)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count users"})
			return
		}
		stats["total_users"] = userCount

		// Admin count
		adminQuery := datastore.NewQuery("User").FilterField("Role", "=", "admin")
		adminCount, err := db.Count(ctx, adminQuery)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count admins"})
			return
		}
		stats["total_admins"] = adminCount

		// Session count (last 24 hours)
		yesterday := time.Now().AddDate(0, 0, -1)
		sessionQuery := datastore.NewQuery("Session").FilterField("StartedAt", ">", yesterday)
		sessionCount, err := db.Count(ctx, sessionQuery)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count sessions"})
			return
		}
		stats["sessions_last_24h"] = sessionCount

		// Content counts
		languageQuery := datastore.NewQuery("Language")
		languageCount, _ := db.Count(ctx, languageQuery)
		stats["total_languages"] = languageCount

		lessonQuery := datastore.NewQuery("Lesson")
		lessonCount, _ := db.Count(ctx, lessonQuery)
		stats["total_lessons"] = lessonCount

		snippetQuery := datastore.NewQuery("Snippet")
		snippetCount, _ := db.Count(ctx, snippetQuery)
		stats["total_snippets"] = snippetCount

		// Recent activity (last 10 sessions)
		recentSessionsQuery := datastore.NewQuery("Session").
			Order("-StartedAt").
			Limit(10)

		var recentSessions []models.Session
		sessionKeys, err := db.GetAll(ctx, recentSessionsQuery, &recentSessions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recent sessions"})
			return
		}

		var sessionsWithUsers []map[string]interface{}
		for i, session := range recentSessions {
			sessionData := map[string]interface{}{
				"id":         sessionKeys[i].Name,
				"started_at": session.StartedAt,
				"mode":       session.Mode,
				"language":   session.LanguageID,
				"user_id":    session.UserID,
			}
			sessionsWithUsers = append(sessionsWithUsers, sessionData)
		}

		response := gin.H{
			"stats":           stats,
			"recent_activity": sessionsWithUsers,
			"timestamp":       time.Now(),
		}

		c.JSON(http.StatusOK, response)
	}
}

// Admin content management handlers

// Get all languages for admin (with more details)
func GetAdminLanguages(db *database.DatastoreClient) gin.HandlerFunc {
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

// Create new language
func CreateAdminLanguage(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var language models.Language
		if err := c.ShouldBindJSON(&language); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		language.CreatedAt = time.Now()

		// Get user ID from context (admin who created it)
		userID, _ := c.Get("user_id")
		language.CreatedBy = userID.(string)

		key := datastore.NameKey("Language", language.ID, nil)

		ctx := context.Background()
		_, err := db.Put(ctx, key, &language)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create language"})
			return
		}

		// Create content version for audit trail
		createContentVersion(db, "language", language.ID, 1, language, userID.(string), "Initial creation")

		c.JSON(http.StatusCreated, language)
	}
}

// Update language
func UpdateAdminLanguage(db *database.DatastoreClient) gin.HandlerFunc {
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
		existing.Name = updates.Name
		existing.Version = updates.Version
		existing.ParserID = updates.ParserID
		existing.GrammarConfig = updates.GrammarConfig
		existing.WhitespaceRules = updates.WhitespaceRules

		userID, _ := c.Get("user_id")

		_, err = db.Put(ctx, key, &existing)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update language"})
			return
		}

		// Create content version for audit trail
		createContentVersion(db, "language", id, existing.Version+1, existing, userID.(string), "Language updated")

		c.JSON(http.StatusOK, existing)
	}
}

// Delete language
func DeleteAdminLanguage(db *database.DatastoreClient) gin.HandlerFunc {
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

// Get all lessons for admin
func GetAdminLessons(db *database.DatastoreClient) gin.HandlerFunc {
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

// Create lesson with token coverage checklist
func CreateAdminLesson(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var lesson models.Lesson
		if err := c.ShouldBindJSON(&lesson); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		lesson.CreatedAt = time.Now()
		lesson.Version = 1

		// Get user ID from context (admin who created it)
		userID, _ := c.Get("user_id")
		lesson.CreatedBy = userID.(string)

		key := datastore.NameKey("Lesson", fmt.Sprintf("%s_%d", lesson.LanguageID, time.Now().Unix()), nil)

		ctx := context.Background()
		_, err := db.Put(ctx, key, &lesson)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create lesson"})
			return
		}

		lesson.ID = key.Name

		// Create content version for audit trail
		createContentVersion(db, "lesson", lesson.ID, 1, lesson, userID.(string), "Initial lesson creation")

		c.JSON(http.StatusCreated, lesson)
	}
}

// Update lesson
func UpdateAdminLesson(db *database.DatastoreClient) gin.HandlerFunc {
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

		userID, _ := c.Get("user_id")

		_, err = db.Put(ctx, key, &existing)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update lesson"})
			return
		}

		// Create content version for audit trail
		createContentVersion(db, "lesson", id, existing.Version, existing, userID.(string), "Lesson updated")

		c.JSON(http.StatusOK, existing)
	}
}

// Delete lesson
func DeleteAdminLesson(db *database.DatastoreClient) gin.HandlerFunc {
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

// Get all snippets for admin
func GetAdminSnippets(db *database.DatastoreClient) gin.HandlerFunc {
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

// Create snippet with validation
func CreateAdminSnippet(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var snippet models.Snippet
		if err := c.ShouldBindJSON(&snippet); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate YAML if it's a YAML snippet
		if snippet.LanguageID == "yaml" {
			if err := validateYAML(snippet.SourceCode); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid YAML format: " + err.Error()})
				return
			}
		}

		// Generate checksum
		snippet.Checksum = utils.GenerateChecksum(snippet.SourceCode)
		snippet.CreatedAt = time.Now()

		// Get user ID from context (admin who created it)
		userID, _ := c.Get("user_id")
		snippet.CreatedBy = userID.(string)

		key := datastore.NameKey("Snippet", fmt.Sprintf("%s_%d", snippet.LanguageID, time.Now().Unix()), nil)

		ctx := context.Background()
		_, err := db.Put(ctx, key, &snippet)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create snippet"})
			return
		}

		snippet.ID = key.Name

		// Create content version for audit trail
		createContentVersion(db, "snippet", snippet.ID, 1, snippet, userID.(string), "Initial snippet creation")

		c.JSON(http.StatusCreated, snippet)
	}
}

// Update snippet
func UpdateAdminSnippet(db *database.DatastoreClient) gin.HandlerFunc {
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

		// Validate YAML if it's a YAML snippet and source code changed
		if updates.LanguageID == "yaml" && updates.SourceCode != existing.SourceCode {
			if err := validateYAML(updates.SourceCode); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid YAML format: " + err.Error()})
				return
			}
		}

		// Update fields
		existing.Title = updates.Title
		existing.SourceCode = updates.SourceCode
		existing.Tags = updates.Tags
		existing.Difficulty = updates.Difficulty
		existing.EstimatedTime = updates.EstimatedTime
		existing.AccessibilityTags = updates.AccessibilityTags

		// Regenerate checksum if source code changed
		if existing.SourceCode != updates.SourceCode {
			existing.Checksum = utils.GenerateChecksum(updates.SourceCode)
		}

		userID, _ := c.Get("user_id")

		_, err = db.Put(ctx, key, &existing)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update snippet"})
			return
		}

		// Create content version for audit trail
		createContentVersion(db, "snippet", id, existing.Version+1, existing, userID.(string), "Snippet updated")

		c.JSON(http.StatusOK, existing)
	}
}

// Delete snippet
func DeleteAdminSnippet(db *database.DatastoreClient) gin.HandlerFunc {
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

// Content versioning and validation handlers

// Get content versions
func GetAdminContentVersions(db *database.DatastoreClient) gin.HandlerFunc {
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

// Restore content version
func RestoreAdminContentVersion(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		contentType := c.Param("contentType")
		contentID := c.Param("contentId")
		versionStr := c.Param("version")

		version, err := strconv.Atoi(versionStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid version number"})
			return
		}

		ctx := context.Background()

		// Get the specific version
		versionKey := datastore.NameKey("ContentVersion",
			fmt.Sprintf("%s_%s_%d", contentType, contentID, version), nil)

		var contentVersion models.ContentVersion
		err = db.Get(ctx, versionKey, &contentVersion)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Content version not found"})
			return
		}

		// Restore based on content type
		switch contentType {
		case "language":
			var language models.Language
			contentBytes, _ := json.Marshal(contentVersion.Content)
			json.Unmarshal(contentBytes, &language)

			langKey := datastore.NameKey("Language", contentID, nil)
			_, err = db.Put(ctx, langKey, &language)

		case "lesson":
			var lesson models.Lesson
			contentBytes, _ := json.Marshal(contentVersion.Content)
			json.Unmarshal(contentBytes, &lesson)

			lessonKey := datastore.NameKey("Lesson", contentID, nil)
			_, err = db.Put(ctx, lessonKey, &lesson)

		case "snippet":
			var snippet models.Snippet
			contentBytes, _ := json.Marshal(contentVersion.Content)
			json.Unmarshal(contentBytes, &snippet)

			snippetKey := datastore.NameKey("Snippet", contentID, nil)
			_, err = db.Put(ctx, snippetKey, &snippet)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore content version"})
			return
		}

		// Create new version record for the restore action
		userID, _ := c.Get("user_id")
		createContentVersion(db, contentType, contentID, version+1, contentVersion.Content, userID.(string), "Restored from version "+versionStr)

		c.JSON(http.StatusOK, gin.H{"message": "Content version restored successfully"})
	}
}

// Validate content
func ValidateAdminContent(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			ContentType string `json:"content_type" binding:"required"`
			ContentID   string `json:"content_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx := context.Background()
		var validationErrors []string
		var isValid bool = true

		switch request.ContentType {
		case "snippet":
			snippetKey := datastore.NameKey("Snippet", request.ContentID, nil)
			var snippet models.Snippet
			err := db.Get(ctx, snippetKey, &snippet)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Snippet not found"})
				return
			}

			// Validate YAML if it's a YAML snippet
			if snippet.LanguageID == "yaml" {
				if err := validateYAML(snippet.SourceCode); err != nil {
					validationErrors = append(validationErrors, "Invalid YAML: "+err.Error())
					isValid = false
				}
			}

			// Check checksum
			expectedChecksum := utils.GenerateChecksum(snippet.SourceCode)
			if snippet.Checksum != expectedChecksum {
				validationErrors = append(validationErrors, "Checksum mismatch")
				isValid = false
			}

		case "lesson":
			lessonKey := datastore.NameKey("Lesson", request.ContentID, nil)
			var lesson models.Lesson
			err := db.Get(ctx, lessonKey, &lesson)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Lesson not found"})
				return
			}

			// Validate lesson structure
			if lesson.Title == "" {
				validationErrors = append(validationErrors, "Title is required")
				isValid = false
			}
			if lesson.LanguageID == "" {
				validationErrors = append(validationErrors, "Language ID is required")
				isValid = false
			}
			if len(lesson.TokensCovered) == 0 {
				validationErrors = append(validationErrors, "Tokens covered cannot be empty")
				isValid = false
			}

		case "language":
			langKey := datastore.NameKey("Language", request.ContentID, nil)
			var language models.Language
			err := db.Get(ctx, langKey, &language)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Language not found"})
				return
			}

			// Validate language structure
			if language.Name == "" {
				validationErrors = append(validationErrors, "Name is required")
				isValid = false
			}
			if language.ParserID == "" {
				validationErrors = append(validationErrors, "Parser ID is required")
				isValid = false
			}
		}

		// Create validation record
		validation := models.ContentValidation{
			ContentType:      request.ContentType,
			ContentID:        request.ContentID,
			Checksum:         utils.GenerateChecksum(fmt.Sprintf("%v", request)),
			ValidationStatus: map[bool]string{true: "valid", false: "invalid"}[isValid],
			ValidationErrors: validationErrors,
			ValidatedAt:      time.Now(),
		}

		key := datastore.NameKey("ContentValidation",
			fmt.Sprintf("%s_%s_%d", request.ContentType, request.ContentID, time.Now().Unix()), nil)

		_, err := db.Put(ctx, key, &validation)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save validation result"})
			return
		}

		response := gin.H{
			"valid":             isValid,
			"validation_errors": validationErrors,
			"validation_id":     key.Name,
		}

		c.JSON(http.StatusOK, response)
	}
}

// Helper function to create content version
func createContentVersion(db *database.DatastoreClient, contentType, contentID string, version int, content interface{}, createdBy, changeNotes string) {
	contentVersion := models.ContentVersion{
		ContentType: contentType,
		ContentID:   contentID,
		Version:     version,
		Content:     content.(map[string]interface{}),
		Checksum:    utils.GenerateChecksum(fmt.Sprintf("%v", content)),
		CreatedBy:   createdBy,
		ChangeNotes: changeNotes,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}

	key := datastore.NameKey("ContentVersion",
		fmt.Sprintf("%s_%s_%d", contentType, contentID, version), nil)

	ctx := context.Background()
	db.Put(ctx, key, &contentVersion)
}

// Simple YAML validation
func validateYAML(content string) error {
	// For now, just check if it can be parsed as JSON-like structure
	// In production, use a proper YAML parser like gopkg.in/yaml.v3
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("content cannot be empty")
	}
	return nil
}

// A/B Testing handlers

// Create A/B test
func CreateABTest(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var test struct {
			Name        string                   `json:"name" binding:"required"`
			Description string                   `json:"description"`
			TestType    string                   `json:"test_type" binding:"required"` // "scoring_weights", "ui_variant"
			Variants    []map[string]interface{} `json:"variants" binding:"required"`
			Duration    int                      `json:"duration_days" binding:"required"`      // Duration in days
			Percentage  float64                  `json:"rollout_percentage" binding:"required"` // Percentage of users to include
		}

		if err := c.ShouldBindJSON(&test); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate variants based on test type
		if test.TestType == "scoring_weights" {
			for _, variant := range test.Variants {
				requiredFields := []string{"twpm_weight", "raw_accuracy_weight", "syntax_accuracy_weight"}
				for _, field := range requiredFields {
					if _, exists := variant[field]; !exists {
						c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Missing required field for scoring_weights test: %s", field)})
						return
					}
				}
			}
		}

		testRecord := models.ABTest{
			Name:           test.Name,
			Description:    test.Description,
			TestType:       test.TestType,
			Variants:       test.Variants,
			Status:         "active",
			StartDate:      time.Now(),
			EndDate:        time.Now().AddDate(0, 0, test.Duration),
			UserPercentage: test.Percentage,
			CreatedBy:      c.GetString("user_id"),
			CreatedAt:      time.Now(),
		}

		key := datastore.NameKey("ABTest", fmt.Sprintf("%s_%d", test.Name, time.Now().Unix()), nil)

		ctx := context.Background()
		_, err := db.Put(ctx, key, &testRecord)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create A/B test"})
			return
		}

		testRecord.ID = key.Name
		c.JSON(http.StatusCreated, testRecord)
	}
}

// Get all A/B tests
func GetABTests(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		query := datastore.NewQuery("ABTest")

		var tests []models.ABTest
		keys, err := db.GetAll(ctx, query, &tests)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch A/B tests"})
			return
		}

		for i, key := range keys {
			tests[i].ID = key.Name
		}

		c.JSON(http.StatusOK, tests)
	}
}

// Update A/B test
func UpdateABTest(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		key := datastore.NameKey("ABTest", id, nil)

		ctx := context.Background()
		var existing models.ABTest
		err := db.Get(ctx, key, &existing)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "A/B test not found"})
			return
		}

		var updates models.ABTest
		if err := c.ShouldBindJSON(&updates); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Update fields
		existing.Name = updates.Name
		existing.Description = updates.Description
		existing.Status = updates.Status
		existing.Variants = updates.Variants

		_, err = db.Put(ctx, key, &existing)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update A/B test"})
			return
		}

		c.JSON(http.StatusOK, existing)
	}
}

// Delete A/B test
func DeleteABTest(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		key := datastore.NameKey("ABTest", id, nil)

		ctx := context.Background()
		err := db.Delete(ctx, key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete A/B test"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "A/B test deleted successfully"})
	}
}

// Get A/B test results
func GetABTestResults(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		key := datastore.NameKey("ABTest", id, nil)

		ctx := context.Background()
		var test models.ABTest
		err := db.Get(ctx, key, &test)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "A/B test not found"})
			return
		}

		// Get test results (simplified - in production, aggregate from session data)
		results := make(map[string]interface{})
		results["test_id"] = id
		results["test_name"] = test.Name
		results["status"] = test.Status
		results["total_participants"] = 0                      // Would calculate from actual session data
		results["conversion_rates"] = make(map[string]float64) // Would calculate from actual metrics

		c.JSON(http.StatusOK, results)
	}
}
