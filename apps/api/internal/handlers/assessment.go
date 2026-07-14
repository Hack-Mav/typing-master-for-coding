package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/typing-master-for-coding-backend/internal/cache"
	"github.com/typing-master-for-coding-backend/internal/database"
	"github.com/typing-master-for-coding-backend/internal/models"
)

// Assessment Blueprint handlers

func GetAssessmentBlueprints(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		language := c.Query("language")

		cacheKey := fmt.Sprintf("assessment_blueprints_%s", language)
		if cached, found := cache.Get(cacheKey); found {
			c.JSON(http.StatusOK, cached)
			return
		}

		query := database.NewQuery("AssessmentBlueprint")
		if language != "" {
			query = query.Filter("language =", language)
		}
		query = query.Order("difficulty")

		var blueprints []models.AssessmentBlueprint
		keys, err := db.GetAll(ctx, query, &blueprints)
		if err != nil {
			log.Printf("Error fetching assessment blueprints: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assessment blueprints"})
			return
		}

		// Set IDs from keys
		for i, key := range keys {
			blueprints[i].ID = key.Name
			if blueprints[i].ID == "" && key.ID != 0 {
				blueprints[i].ID = strconv.FormatInt(key.ID, 10)
			}
		}

		// Cache the results
		cache.Set(cacheKey, blueprints)
		c.JSON(http.StatusOK, blueprints)
	}
}

func GetAssessmentBlueprint(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		blueprintID := c.Param("id")

		cacheKey := fmt.Sprintf("assessment_blueprint_%s", blueprintID)
		if cached, found := cache.Get(cacheKey); found {
			c.JSON(http.StatusOK, cached)
			return
		}

		key := database.NameKey("AssessmentBlueprint", blueprintID, nil)
		var blueprint models.AssessmentBlueprint

		err := db.Get(ctx, key, &blueprint)
		if err != nil {
			if err == database.ErrNoSuchEntity {
				c.JSON(http.StatusNotFound, gin.H{"error": "Assessment blueprint not found"})
				return
			}
			log.Printf("Error fetching assessment blueprint: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assessment blueprint"})
			return
		}

		blueprint.ID = blueprintID

		// Cache the result
		cache.Set(cacheKey, blueprint)
		c.JSON(http.StatusOK, blueprint)
	}
}

func CreateAssessmentBlueprint(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		var blueprint models.AssessmentBlueprint
		if err := c.ShouldBindJSON(&blueprint); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Validate required fields
		if blueprint.Name == "" || blueprint.Language == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name and language are required"})
			return
		}

		// Generate ID if not provided
		if blueprint.ID == "" {
			blueprint.ID = fmt.Sprintf("%s-%s-%d", blueprint.Language, "assessment", time.Now().Unix())
		}

		blueprint.CreatedAt = time.Now()
		blueprint.Version = 1

		key := database.NameKey("AssessmentBlueprint", blueprint.ID, nil)
		_, err := db.Put(ctx, key, &blueprint)
		if err != nil {
			log.Printf("Error creating assessment blueprint: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create assessment blueprint"})
			return
		}

		// Invalidate cache
		cache.Delete(fmt.Sprintf("assessment_blueprints_%s", blueprint.Language))

		c.JSON(http.StatusCreated, blueprint)
	}
}

// Assessment Session handlers

func CreateAssessmentSession(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		var request struct {
			BlueprintID string `json:"blueprintId" binding:"required"`
			UserID      string `json:"userId"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Verify blueprint exists
		blueprintKey := database.NameKey("AssessmentBlueprint", request.BlueprintID, nil)
		var blueprint models.AssessmentBlueprint
		err := db.Get(ctx, blueprintKey, &blueprint)
		if err != nil {
			if err == database.ErrNoSuchEntity {
				c.JSON(http.StatusNotFound, gin.H{"error": "Assessment blueprint not found"})
				return
			}
			log.Printf("Error fetching blueprint: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify blueprint"})
			return
		}

		// Create assessment session
		session := models.AssessmentSession{
			ID:                  fmt.Sprintf("session-%d", time.Now().UnixNano()),
			BlueprintID:         request.BlueprintID,
			UserID:              request.UserID,
			StartedAt:           time.Now(),
			Status:              "not_started",
			CurrentSnippetIndex: 0,
			SnippetResults:      []models.AssessmentSnippetResult{},
			Metadata: models.AssessmentMetadata{
				Language:          blueprint.Language,
				Difficulty:        blueprint.Difficulty,
				TotalSnippets:     len(blueprint.SnippetIDs),
				EstimatedDuration: blueprint.EstimatedDuration,
			},
		}

		key := database.NameKey("AssessmentSession", session.ID, nil)
		_, err = db.Put(ctx, key, &session)
		if err != nil {
			log.Printf("Error creating assessment session: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create assessment session"})
			return
		}

		c.JSON(http.StatusCreated, session)
	}
}

func GetAssessmentSession(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		sessionID := c.Param("id")

		key := database.NameKey("AssessmentSession", sessionID, nil)
		var session models.AssessmentSession

		err := db.Get(ctx, key, &session)
		if err != nil {
			if err == database.ErrNoSuchEntity {
				c.JSON(http.StatusNotFound, gin.H{"error": "Assessment session not found"})
				return
			}
			log.Printf("Error fetching assessment session: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assessment session"})
			return
		}

		session.ID = sessionID
		c.JSON(http.StatusOK, session)
	}
}

func RecordSnippetResult(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		sessionID := c.Param("id")

		var result models.AssessmentSnippetResult
		if err := c.ShouldBindJSON(&result); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Get current session
		sessionKey := database.NameKey("AssessmentSession", sessionID, nil)
		var session models.AssessmentSession
		err := db.Get(ctx, sessionKey, &session)
		if err != nil {
			if err == database.ErrNoSuchEntity {
				c.JSON(http.StatusNotFound, gin.H{"error": "Assessment session not found"})
				return
			}
			log.Printf("Error fetching session: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch session"})
			return
		}

		// Add result to session
		session.SnippetResults = append(session.SnippetResults, result)
		session.CurrentSnippetIndex++
		session.Status = "in_progress"

		// Update session
		_, err = db.Put(ctx, sessionKey, &session)
		if err != nil {
			log.Printf("Error updating session: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Snippet result recorded"})
	}
}

func FinalizeAssessment(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		sessionID := c.Param("id")

		var request struct {
			TimeExpired bool `json:"timeExpired"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Get session
		sessionKey := database.NameKey("AssessmentSession", sessionID, nil)
		var session models.AssessmentSession
		err := db.Get(ctx, sessionKey, &session)
		if err != nil {
			if err == database.ErrNoSuchEntity {
				c.JSON(http.StatusNotFound, gin.H{"error": "Assessment session not found"})
				return
			}
			log.Printf("Error fetching session: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch session"})
			return
		}

		// Get blueprint for scoring criteria
		blueprintKey := database.NameKey("AssessmentBlueprint", session.BlueprintID, nil)
		var blueprint models.AssessmentBlueprint
		err = db.Get(ctx, blueprintKey, &blueprint)
		if err != nil {
			log.Printf("Error fetching blueprint: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch blueprint"})
			return
		}

		// Calculate final result
		result := calculateAssessmentResult(session, blueprint, request.TimeExpired)

		// Update session status
		session.Status = "completed"
		session.CompletedAt = &[]time.Time{time.Now()}[0]
		session.OverallResult = &result

		// Save updated session
		_, err = db.Put(ctx, sessionKey, &session)
		if err != nil {
			log.Printf("Error finalizing session: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finalize session"})
			return
		}

		// Store result separately for analytics
		resultKey := database.NameKey("AssessmentResult", sessionID, nil)
		_, err = db.Put(ctx, resultKey, &result)
		if err != nil {
			log.Printf("Error storing assessment result: %v", err)
			// Don't fail the request, just log the error
		}

		c.JSON(http.StatusOK, result)
	}
}

func GetAssessmentSnippet(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		blueprintID := c.Param("id")
		snippetID := c.Param("snippetId")

		cacheKey := fmt.Sprintf("assessment_snippet_%s_%s", blueprintID, snippetID)
		if cached, found := cache.Get(cacheKey); found {
			c.JSON(http.StatusOK, cached)
			return
		}

		// Get snippet
		snippetKey := database.NameKey("Snippet", snippetID, nil)
		var snippet models.Snippet
		err := db.Get(ctx, snippetKey, &snippet)
		if err != nil {
			if err == database.ErrNoSuchEntity {
				c.JSON(http.StatusNotFound, gin.H{"error": "Assessment snippet not found"})
				return
			}
			log.Printf("Error fetching snippet: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch snippet"})
			return
		}

		snippet.ID = snippetID

		// Cache the result
		cache.Set(cacheKey, snippet)
		c.JSON(http.StatusOK, snippet)
	}
}

// Assessment Analytics handlers

func GetAssessmentAnalytics(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		language := c.Query("language")

		cacheKey := fmt.Sprintf("assessment_analytics_%s", language)
		if cached, found := cache.Get(cacheKey); found {
			c.JSON(http.StatusOK, cached)
			return
		}

		// For MVP, return mock analytics
		// TODO: Implement proper analytics querying

		// For MVP, return mock analytics
		analytics := models.AssessmentAnalytics{
			TotalAttempts:   1250,
			PassRate:        0.72,
			AverageScore:    745,
			AverageDuration: 18.5,
			CommonFailurePoints: []models.FailurePoint{
				{
					SnippetID:           "javascript-snippet-2",
					Position:            45,
					ErrorType:           "syntax_error",
					Frequency:           0.35,
					AverageRecoveryTime: 2500,
				},
			},
			DifficultyDistribution: map[int]float64{
				1: 0.05,
				2: 0.15,
				3: 0.40,
				4: 0.30,
				5: 0.10,
			},
			LanguagePerformance: map[string]models.LanguagePerformance{
				"javascript": {
					AverageScore:  745,
					PassRate:      0.72,
					CommonErrors:  []string{"missing_semicolon", "bracket_mismatch"},
					StrengthAreas: []string{"function_syntax", "variable_declaration"},
				},
				"python": {
					AverageScore:  720,
					PassRate:      0.68,
					CommonErrors:  []string{"indentation_error", "colon_missing"},
					StrengthAreas: []string{"list_comprehension", "function_definition"},
				},
			},
		}

		// Cache the result
		cache.Set(cacheKey, analytics)
		c.JSON(http.StatusOK, analytics)
	}
}

// Badge system handlers

func GetUserBadges(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		userID := c.Query("userId")

		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "userId parameter is required"})
			return
		}

		cacheKey := fmt.Sprintf("user_badges_%s", userID)
		if cached, found := cache.Get(cacheKey); found {
			c.JSON(http.StatusOK, cached)
			return
		}

		query := database.NewQuery("UserBadge").Filter("user_id =", userID)
		var userBadges []models.UserBadge
		keys, err := db.GetAll(ctx, query, &userBadges)
		if err != nil {
			log.Printf("Error fetching user badges: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user badges"})
			return
		}

		// Get badge details
		badges := make([]models.AssessmentBadge, 0, len(userBadges))
		for i, userBadge := range userBadges {
			badgeKey := database.NameKey("Badge", userBadge.BadgeID, nil)
			var badge models.AssessmentBadge
			err := db.Get(ctx, badgeKey, &badge)
			if err != nil {
				log.Printf("Error fetching badge %s: %v", userBadge.BadgeID, err)
				continue
			}
			badge.ID = keys[i].Name
			badge.EarnedAt = &userBadge.EarnedAt
			badges = append(badges, badge)
		}

		// Cache the result
		cache.Set(cacheKey, badges)
		c.JSON(http.StatusOK, badges)
	}
}

// Assessment Scheduling handlers

func ScheduleAssessment(db *database.DatastoreClient, cache *cache.InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		var schedule models.AssessmentSchedule
		if err := c.ShouldBindJSON(&schedule); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Validate required fields
		if schedule.UserID == "" || schedule.AssessmentID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "UserID and AssessmentID are required"})
			return
		}

		schedule.RescheduledCount = 0
		schedule.Completed = false
		schedule.ReminderSent = false

		scheduleID := fmt.Sprintf("%s-%s-%d", schedule.UserID, schedule.AssessmentID, time.Now().Unix())
		key := database.NameKey("AssessmentSchedule", scheduleID, nil)

		_, err := db.Put(ctx, key, &schedule)
		if err != nil {
			log.Printf("Error scheduling assessment: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to schedule assessment"})
			return
		}

		c.JSON(http.StatusCreated, schedule)
	}
}

// Helper function to calculate assessment result
func calculateAssessmentResult(session models.AssessmentSession, blueprint models.AssessmentBlueprint, timeExpired bool) models.AssessmentResult {
	// This is a simplified calculation - in a real implementation,
	// this would use the advanced scoring algorithms from the StructuralAnalyzer

	totalScore := 0.0
	totalSnippets := len(session.SnippetResults)
	passedSnippets := 0

	for _, result := range session.SnippetResults {
		if result.Passed {
			passedSnippets++
		}
		// Add metrics-based scoring here
		totalScore += float64(result.Metrics.RawAccuracy * 100)
	}

	if totalSnippets > 0 {
		totalScore = totalScore / float64(totalSnippets)
	}

	// Apply time penalty if expired
	if timeExpired {
		totalScore *= 0.8
	}

	// Determine grade
	grade := "F"
	switch {
	case totalScore >= 95:
		grade = "A+"
	case totalScore >= 90:
		grade = "A"
	case totalScore >= 85:
		grade = "A-"
	case totalScore >= 80:
		grade = "B+"
	case totalScore >= 75:
		grade = "B"
	case totalScore >= 70:
		grade = "B-"
	case totalScore >= 65:
		grade = "C+"
	case totalScore >= 60:
		grade = "C"
	case totalScore >= 55:
		grade = "C-"
	case totalScore >= 50:
		grade = "D"
	}

	passed := totalScore >= float64(blueprint.PassingCriteria.MinimumAccuracy*100)

	return models.AssessmentResult{
		OverallScore: int(totalScore),
		Passed:       passed,
		Grade:        grade,
		Breakdown: models.AssessmentScoreBreakdown{
			SpeedScore:         75,
			AccuracyScore:      totalScore,
			StructuralScore:    80,
			SyntaxScore:        85,
			ConsistencyScore:   70,
			ErrorRecoveryScore: 75,
			TotalPenalties: func() float64 {
				if timeExpired {
					return 20
				} else {
					return 10
				}
			}(),
			BonusPoints: func() float64 {
				if passed {
					return 10
				} else {
					return 0
				}
			}(),
		},
		Recommendations: []string{
			"Focus on improving typing consistency",
			"Practice more complex syntax patterns",
			"Work on error correction efficiency",
		},
		CertificateEligible: totalScore >= 80,
		RetakeAllowed:       !passed,
		NextAssessmentSuggestion: func() *string {
			if totalScore >= 80 {
				suggestion := "advanced-assessment"
				return &suggestion
			}
			return nil
		}(),
	}
}
