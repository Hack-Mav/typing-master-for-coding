package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/typing-master-for-coding-backend/internal/database"
	"github.com/typing-master-for-coding-backend/internal/models"
)

// GetTutorials returns all available tutorials
func GetTutorials(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		category := c.Query("category")
		difficulty := c.Query("difficulty")

		var tutorials []models.Tutorial
		var err error

		if category != "" {
			tutorials, err = database.GetTutorialsByCategory(db, category)
		} else if difficulty != "" {
			tutorials, err = database.GetTutorialsByDifficulty(db, difficulty)
		} else {
			tutorials, err = database.GetAllTutorials(db)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tutorials"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"tutorials": tutorials,
			"total":     len(tutorials),
		})
	}
}

// GetTutorial returns a specific tutorial by ID
func GetTutorial(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		tutorialID := c.Param("id")

		tutorial, err := database.GetTutorialByID(db, tutorialID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tutorial not found"})
			return
		}

		c.JSON(http.StatusOK, tutorial)
	}
}

// StartTutorial starts a tutorial for a user
func StartTutorial(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		tutorialID := c.Param("id")

		// Check if tutorial exists
		tutorial, err := database.GetTutorialByID(db, tutorialID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tutorial not found"})
			return
		}

		// Get or create user progress
		progress, err := database.GetUserTutorialProgress(db, userID, tutorialID)
		if err != nil {
			// Create new progress
			progress = &models.UserTutorialProgress{
				ID:             generateID(),
				UserID:         userID,
				TutorialID:     tutorialID,
				CurrentStep:    0,
				CompletedSteps: []string{},
				Status:         "in_progress",
				StartedAt:      time.Now(),
				LastAccessedAt: time.Now(),
			}

			if err := database.CreateUserTutorialProgress(db, progress); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start tutorial"})
				return
			}
		} else {
			// Update existing progress
			progress.Status = "in_progress"
			progress.LastAccessedAt = time.Now()
			if progress.StartedAt.IsZero() {
				progress.StartedAt = time.Now()
			}

			if err := database.UpdateUserTutorialProgress(db, progress); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tutorial progress"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"tutorial": tutorial,
			"progress": progress,
			"message":  "Tutorial started successfully",
		})
	}
}

// UpdateTutorialProgress updates a user's progress in a tutorial
func UpdateTutorialProgress(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		tutorialID := c.Param("id")

		var req struct {
			StepID    string `json:"step_id" binding:"required"`
			Completed bool   `json:"completed"`
			TimeSpent int    `json:"time_spent"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		progress, err := database.GetUserTutorialProgress(db, userID, tutorialID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tutorial progress not found"})
			return
		}

		// Update progress
		if req.Completed {
			progress.CompletedSteps = append(progress.CompletedSteps, req.StepID)
			progress.CurrentStep++
		}

		progress.TimeSpent += req.TimeSpent
		progress.LastAccessedAt = time.Now()

		// Check if tutorial is completed
		tutorial, err := database.GetTutorialByID(db, tutorialID)
		if err == nil && len(progress.CompletedSteps) >= len(tutorial.Steps) {
			progress.Status = "completed"
			now := time.Now()
			progress.CompletedAt = &now
		}

		if err := database.UpdateUserTutorialProgress(db, progress); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update progress"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"progress": progress,
			"message":  "Progress updated successfully",
		})
	}
}

// GetUserTutorialProgress returns a user's progress for all tutorials
func GetUserTutorialProgress(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		progress, err := database.GetUserAllTutorialProgress(db, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tutorial progress"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"progress": progress,
			"total":    len(progress),
		})
	}
}

// GetTooltips returns available tooltips for the current context
func GetTooltips(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		pageContext := c.Query("page_context")
		userID := c.GetString("user_id")

		// Get user's onboarding state to check viewed tooltips
		userState, err := database.GetUserOnboardingState(db, userID)
		if err != nil {
			userState = &models.UserOnboardingState{
				ViewedTooltips: []string{},
			}
		}

		// Get tooltips for context
		tooltips, err := database.GetTooltipsByContext(db, pageContext)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tooltips"})
			return
		}

		// Filter out already viewed tooltips (unless they're persistent)
		var availableTooltips []models.Tooltip
		for _, tooltip := range tooltips {
			if tooltip.IsPersistent || !contains(userState.ViewedTooltips, tooltip.ID) {
				availableTooltips = append(availableTooltips, tooltip)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"tooltips": availableTooltips,
			"total":    len(availableTooltips),
		})
	}
}

// MarkTooltipViewed marks a tooltip as viewed by a user
func MarkTooltipViewed(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		tooltipID := c.Param("id")

		userState, err := database.GetUserOnboardingState(db, userID)
		if err != nil {
			// Create new onboarding state
			userState = &models.UserOnboardingState{
				ID:                generateID(),
				UserID:            userID,
				ViewedTooltips:    []string{tooltipID},
				OverallProgress:   0.0,
				LastActivityAt:    time.Now(),
				OnboardingVersion: "1.0",
				Preferences: models.OnboardingPreferences{
					ShowTooltips:       true,
					AutoStartTutorials: true,
					InteractiveMode:    true,
					CompactView:        false,
					Language:           "en",
				},
				CreatedAt: time.Now(),
			}
		} else {
			// Add to viewed tooltips if not already present
			if !contains(userState.ViewedTooltips, tooltipID) {
				userState.ViewedTooltips = append(userState.ViewedTooltips, tooltipID)
			}
			userState.LastActivityAt = time.Now()
		}

		if err := database.UpdateUserOnboardingState(db, userState); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark tooltip as viewed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Tooltip marked as viewed"})
	}
}

// SearchHelpArticles searches help articles
func SearchHelpArticles(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		category := c.Query("category")
		limitStr := c.DefaultQuery("limit", "10")

		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			limit = 10
		}

		articles, err := database.SearchHelpArticles(db, query, category, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search help articles"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"articles": articles,
			"total":    len(articles),
			"query":    query,
		})
	}
}

// GetHelpArticle returns a specific help article
func GetHelpArticle(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		articleID := c.Param("id")
		userID := c.GetString("user_id")

		article, err := database.GetHelpArticleByID(db, articleID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Help article not found"})
			return
		}

		// Increment view count
		article.ViewCount++
		database.UpdateHelpArticle(db, article)

		// Mark as read in user's onboarding state
		userState, err := database.GetUserOnboardingState(db, userID)
		if err == nil {
			if !contains(userState.ReadArticles, articleID) {
				userState.ReadArticles = append(userState.ReadArticles, articleID)
				database.UpdateUserOnboardingState(db, userState)
			}
		}

		c.JSON(http.StatusOK, article)
	}
}

// GetFAQs returns frequently asked questions
func GetFAQs(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		category := c.Query("category")

		faqs, err := database.GetFAQsByCategory(db, category)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch FAQs"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"faqs":  faqs,
			"total": len(faqs),
		})
	}
}

// RateHelpArticle rates a help article as helpful or not
func RateHelpArticle(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		articleID := c.Param("id")

		var req struct {
			Helpful bool `json:"helpful" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		article, err := database.GetHelpArticleByID(db, articleID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Help article not found"})
			return
		}

		if req.Helpful {
			article.HelpfulCount++
		} else {
			article.NotHelpfulCount++
		}

		if err := database.UpdateHelpArticle(db, article); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to rate article"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Article rated successfully"})
	}
}

// GetOnboardingState returns the user's onboarding state
func GetOnboardingState(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		state, err := database.GetUserOnboardingState(db, userID)
		if err != nil {
			// Return default state if none exists
			state = &models.UserOnboardingState{
				OverallProgress:   0.0,
				LastActivityAt:    time.Now(),
				OnboardingVersion: "1.0",
				Preferences: models.OnboardingPreferences{
					ShowTooltips:       true,
					AutoStartTutorials: true,
					InteractiveMode:    true,
					CompactView:        false,
					Language:           "en",
				},
			}
		}

		c.JSON(http.StatusOK, state)
	}
}

// UpdateOnboardingPreferences updates user's onboarding preferences
func UpdateOnboardingPreferences(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		var req models.OnboardingPreferences
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userState, err := database.GetUserOnboardingState(db, userID)
		if err != nil {
			// Create new onboarding state
			userState = &models.UserOnboardingState{
				ID:                generateID(),
				UserID:            userID,
				OverallProgress:   0.0,
				LastActivityAt:    time.Now(),
				OnboardingVersion: "1.0",
				Preferences:       req,
				CreatedAt:         time.Now(),
			}
		} else {
			userState.Preferences = req
			userState.LastActivityAt = time.Now()
		}

		if err := database.UpdateUserOnboardingState(db, userState); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update preferences"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"preferences": userState.Preferences,
			"message":     "Preferences updated successfully",
		})
	}
}

// Helper functions
func generateID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}
