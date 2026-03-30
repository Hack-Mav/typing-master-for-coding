package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/typing-master-for-coding-backend/internal/database"
	"github.com/typing-master-for-coding-backend/internal/models"
)

// TrackAnalyticsEvent tracks an analytics event
func TrackAnalyticsEvent(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			EventType  string                 `json:"event_type" binding:"required"`
			EventName  string                 `json:"event_name" binding:"required"`
			Properties map[string]interface{} `json:"properties"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get user ID if authenticated
		userID, exists := c.Get("user_id")
		if !exists {
			userID = ""
		}

		// Check analytics consent
		if userID != "" {
			consent, err := database.GetUserAnalyticsConsent(db, userID.(string))
			if err != nil || !consent.ConsentGiven {
				c.JSON(http.StatusForbidden, gin.H{"error": "Analytics consent not given"})
				return
			}

			// Check if event type is allowed
			if !contains(consent.AllowedEvents, req.EventType) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Event type not consented"})
				return
			}
		}

		// Extract device and browser information
		userAgent := c.GetHeader("User-Agent")
		ipAddress := c.ClientIP()
		deviceType := "unknown"
		browser := "unknown"

		// Parse user agent for device/browser info (simplified)
		if contains([]string{"Mobile", "Android", "iPhone", "iPad"}, userAgent) {
			deviceType = "mobile"
		} else if contains([]string{"Tablet", "iPad"}, userAgent) {
			deviceType = "tablet"
		} else {
			deviceType = "desktop"
		}

		if contains([]string{"Chrome"}, userAgent) {
			browser = "chrome"
		} else if contains([]string{"Firefox"}, userAgent) {
			browser = "firefox"
		} else if contains([]string{"Safari"}, userAgent) && !contains([]string{"Chrome"}, userAgent) {
			browser = "safari"
		} else if contains([]string{"Edge"}, userAgent) {
			browser = "edge"
		}

		// Create analytics event
		event := &models.AnalyticsEvent{
			ID:         generateID(),
			UserID:     toString(userID),
			SessionID:  c.GetString("session_id"),
			EventType:  req.EventType,
			EventName:  req.EventName,
			Properties: req.Properties,
			Timestamp:  time.Now(),
			UserAgent:  userAgent,
			IPAddress:  hashIPAddress(ipAddress),
			DeviceType: deviceType,
			Browser:    browser,
			Version:    "1.0.0",
		}

		if err := database.TrackAnalyticsEvent(db, event); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track event"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":  "Event tracked successfully",
			"event_id": event.ID,
		})
	}
}

// BatchTrackAnalyticsEvents tracks multiple analytics events
func BatchTrackAnalyticsEvents(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Events []struct {
				EventType  string                 `json:"event_type" binding:"required"`
				EventName  string                 `json:"event_name" binding:"required"`
				Properties map[string]interface{} `json:"properties"`
				Timestamp  time.Time              `json:"timestamp"`
			} `json:"events" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get user ID if authenticated
		userID, exists := c.Get("user_id")
		if !exists {
			userID = ""
		}

		// Check analytics consent
		if userID != "" {
			consent, err := database.GetUserAnalyticsConsent(db, userID.(string))
			if err != nil || !consent.ConsentGiven {
				c.JSON(http.StatusForbidden, gin.H{"error": "Analytics consent not given"})
				return
			}
		}

		// Process events
		var events []*models.AnalyticsEvent
		for _, eventReq := range req.Events {
			// Check consent for each event type
			if userID != "" {
				consent, _ := database.GetUserAnalyticsConsent(db, userID.(string))
				if !contains(consent.AllowedEvents, eventReq.EventType) {
					continue // Skip events not consented
				}
			}

			event := &models.AnalyticsEvent{
				ID:         generateID(),
				UserID:     toString(userID),
				SessionID:  c.GetString("session_id"),
				EventType:  eventReq.EventType,
				EventName:  eventReq.EventName,
				Properties: eventReq.Properties,
				Timestamp:  eventReq.Timestamp,
				UserAgent:  c.GetHeader("User-Agent"),
				IPAddress:  hashIPAddress(c.ClientIP()),
				Version:    "1.0.0",
			}

			events = append(events, event)
		}

		if err := database.BatchTrackAnalyticsEvents(db, events); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track events"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Events tracked successfully",
			"count":   len(events),
		})
	}
}

// GetAnalyticsConsent returns user's analytics consent status
func GetAnalyticsConsent(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		consent, err := database.GetUserAnalyticsConsent(db, userID)
		if err != nil {
			// Return default consent if none exists
			consent = &models.AnalyticsConsent{
				ID:             generateID(),
				UserID:         userID,
				ConsentGiven:   false,
				ConsentType:    "none",
				ConsentVersion: "1.0",
				DataRetention:  365,
				AllowedEvents:  []string{},
				DeclinedEvents: []string{},
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
				ExpiresAt:      time.Now().AddDate(1, 0, 0),
			}
		}

		c.JSON(http.StatusOK, consent)
	}
}

// UpdateAnalyticsConsent updates user's analytics consent
func UpdateAnalyticsConsent(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		var req struct {
			ConsentGiven   bool     `json:"consent_given" binding:"required"`
			ConsentType    string   `json:"consent_type" binding:"required"`
			DataRetention  int      `json:"data_retention"`
			AllowedEvents  []string `json:"allowed_events"`
			DeclinedEvents []string `json:"declined_events"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		consent := &models.AnalyticsConsent{
			ID:             generateID(),
			UserID:         userID,
			ConsentGiven:   req.ConsentGiven,
			ConsentType:    req.ConsentType,
			ConsentVersion: "1.0",
			DataRetention:  req.DataRetention,
			AllowedEvents:  req.AllowedEvents,
			DeclinedEvents: req.DeclinedEvents,
			IPAddressHash:  hashIPAddress(c.ClientIP()),
			UserAgentHash:  hashString(c.GetHeader("User-Agent")),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			ExpiresAt:      time.Now().AddDate(1, 0, 0),
		}

		if err := database.UpdateAnalyticsConsent(db, consent); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update consent"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"consent": consent,
			"message": "Analytics consent updated successfully",
		})
	}
}

// GetAnalyticsEvents returns analytics events (for admins)
func GetAnalyticsEvents(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		eventType := c.Query("event_type")
		eventName := c.Query("event_name")
		startDate := c.Query("start_date")
		endDate := c.Query("end_date")
		limit := c.DefaultQuery("limit", "100")

		events, err := database.GetAnalyticsEvents(db, eventType, eventName, startDate, endDate, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch analytics events"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"events": events,
			"total":  len(events),
		})
	}
}

// GetAnalyticsStats returns analytics statistics (for admins)
func GetAnalyticsStats(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := database.GetAnalyticsStats(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch analytics stats"})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}

// GetABTestsAnalytics returns available A/B tests for analytics
func GetABTestsAnalytics(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Analytics A/B tests not implemented"})
	}
}

// GetABTest returns a specific A/B test
func GetABTest(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		testID := c.Param("id")

		test, err := database.GetABTestByID(db, testID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "A/B test not found"})
			return
		}

		c.JSON(http.StatusOK, test)
	}
}

// GetUserABTestAssignment returns user's A/B test assignments
func GetUserABTestAssignment(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		testID := c.Param("test_id")

		assignment, err := database.GetUserABTestAssignment(db, userID, testID)
		if err != nil {
			// Create new assignment if none exists
			test, err := database.GetABTestByID(db, testID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "A/B test not found"})
				return
			}

			// Check if user is in target audience (simplified for basic ABTest model)
			if test.UserPercentage < 1.0 {
				// Simple hash-based inclusion for demo purposes
				hash := 0
				for _, char := range userID {
					hash += int(char)
				}
				if float64(hash%100) > test.UserPercentage*100 {
					c.JSON(http.StatusForbidden, gin.H{"error": "User not included in test"})
					return
				}
			}

			// Assign variant based on variants (simplified)
			variantID := "variant_1" // Default to first variant
			if len(test.Variants) > 1 {
				hash := 0
				for _, char := range userID {
					hash += int(char)
				}
				variantIndex := hash % len(test.Variants)
				variant := test.Variants[variantIndex]
				if id, exists := variant["id"].(string); exists {
					variantID = id
				}
			}

			assignment = &models.UserABTestAssignment{
				ID:         generateID(),
				UserID:     userID,
				ABTestID:   testID,
				VariantID:  variantID,
				AssignedAt: time.Now(),
				IsExcluded: false,
			}

			database.CreateUserABTestAssignment(db, assignment)
		}

		c.JSON(http.StatusOK, assignment)
	}
}

// CreateABTestAnalytics creates a new A/B test (for analytics)
func CreateABTestAnalytics(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.ABTest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		req.ID = generateID()
		req.Status = "draft"
		req.CreatedAt = time.Now()
		req.UpdatedAt = time.Now()

		if err := database.CreateABTest(db, &req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create A/B test"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"test":    req,
			"message": "A/B test created successfully",
		})
	}
}

// UpdateABTestAnalytics updates an A/B test (for analytics)
func UpdateABTestAnalytics(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		testID := c.Param("id")

		var req struct {
			Name           string                   `json:"name"`
			Description    string                   `json:"description"`
			Status         string                   `json:"status"`
			Variants       []map[string]interface{} `json:"variants"`
			UserPercentage float64                  `json:"user_percentage"`
			StartDate      time.Time                `json:"start_date"`
			EndDate        time.Time                `json:"end_date"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		test, err := database.GetABTestByID(db, testID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "A/B test not found"})
			return
		}

		// Update fields
		if req.Name != "" {
			test.Name = req.Name
		}
		if req.Description != "" {
			test.Description = req.Description
		}
		if req.Status != "" {
			test.Status = req.Status
		}
		if len(req.Variants) > 0 {
			test.Variants = req.Variants
		}
		if req.UserPercentage > 0 {
			test.UserPercentage = req.UserPercentage
		}
		if !req.StartDate.IsZero() {
			test.StartDate = req.StartDate
		}
		if !req.EndDate.IsZero() {
			test.EndDate = req.EndDate
		}
		test.UpdatedAt = time.Now()

		if err := database.UpdateABTest(db, test); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update A/B test"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"test":    test,
			"message": "A/B test updated successfully",
		})
	}
}

// GetABTestResultsAnalytics returns A/B test results (for analytics)
func GetABTestResultsAnalytics(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		testID := c.Param("id")

		results, err := database.GetABTestResults(db, testID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch A/B test results"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"results": results,
			"total":   len(results),
		})
	}
}

// Helper functions
func toString(value interface{}) string {
	if value == nil {
		return ""
	}
	return value.(string)
}

func hashIPAddress(ip string) string {
	hash := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(hash[:])
}

func isUserInTargetAudience(userID string, audience models.TargetAudience) bool {
	// Simplified audience filtering - in a real implementation,
	// this would check user properties against filters
	return true
}

func assignVariant(splits []models.TrafficSplit) string {
	// Simple variant assignment based on weight
	// In a real implementation, this would use consistent hashing
	// to ensure the same user always gets the same variant
	totalWeight := 0.0
	for _, split := range splits {
		totalWeight += split.Weight
	}

	// For simplicity, return the first variant
	return splits[0].VariantID
}
