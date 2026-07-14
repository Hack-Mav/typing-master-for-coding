package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/typing-master-for-coding-backend/internal/database"
	"github.com/typing-master-for-coding-backend/internal/models"
)

// UpdatePrivacySettings updates user privacy settings
func UpdatePrivacySettings(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Anonymous users cannot update privacy settings
		if strings.HasPrefix(userID.(string), "anon_") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anonymous users cannot update privacy settings"})
			return
		}

		var req models.PrivacySettingsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx := context.Background()
		key := database.NameKey("User", userID.(string), nil)

		var user models.User
		err := db.Get(ctx, key, &user)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// Update privacy settings
		user.PrivacyMode = req.PrivacyMode
		user.TelemetryConsent = req.TelemetryConsent
		user.DataProcessingConsent = req.DataProcessingConsent
		user.UpdatedAt = time.Now().UTC()

		// Save updated user
		_, err = db.Put(ctx, key, &user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update privacy settings"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"privacy_mode":            user.PrivacyMode,
			"telemetry_consent":       user.TelemetryConsent,
			"data_processing_consent": user.DataProcessingConsent,
		})
	}
}

// ExportUserData exports all user data (GDPR compliance)
func ExportUserData(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Anonymous users have no server-side data
		if strings.HasPrefix(userID.(string), "anon_") {
			c.JSON(http.StatusOK, gin.H{
				"message": "Anonymous users have no server-side data to export",
			})
			return
		}

		format := c.DefaultQuery("format", "json")
		ctx := context.Background()

		// Gather all user data
		userData := make(map[string]interface{})

		// Get user profile
		userKey := database.NameKey("User", userID.(string), nil)
		var user models.User
		err := db.Get(ctx, userKey, &user)
		if err == nil {
			user.ID = userID.(string)
			userData["profile"] = toUserResponse(&user)
		}

		// Get user sessions
		sessionsQuery := database.NewQuery("Session").Filter("user_id =", userID.(string))
		var sessions []*models.Session
		sessionKeys, err := db.GetAll(ctx, sessionsQuery, &sessions)
		if err == nil {
			for i, key := range sessionKeys {
				sessions[i].ID = key.Name
				if sessions[i].ID == "" {
					sessions[i].ID = key.Encode()
				}
			}
			userData["sessions"] = sessions
		} else {
			userData["sessions"] = []*models.Session{}
		}

		// Get user results
		resultsQuery := database.NewQuery("Result").Filter("user_id =", userID.(string))
		var results []*models.Result
		resultKeys, err := db.GetAll(ctx, resultsQuery, &results)
		if err == nil {
			for i, key := range resultKeys {
				results[i].SessionID = key.Name
				if results[i].SessionID == "" {
					results[i].SessionID = key.Encode()
				}
			}
			userData["results"] = results
		} else {
			userData["results"] = []*models.Result{}
		}

		// Get lesson progress
		progressQuery := database.NewQuery("LessonProgress").Filter("user_id =", userID.(string))
		var progress []*models.LessonProgress
		_, err = db.GetAll(ctx, progressQuery, &progress)
		if err == nil {
			userData["lesson_progress"] = progress
		} else {
			userData["lesson_progress"] = []*models.LessonProgress{}
		}

		// Export based on format
		if format == "csv" {
			exportDataAsCSV(c, userData)
		} else {
			c.JSON(http.StatusOK, gin.H{
				"user_id":     userID,
				"exported_at": time.Now().UTC().Format(time.RFC3339),
				"data":        userData,
			})
		}
	}
}

// DeleteUserData deletes all user data (GDPR compliance)
func DeleteUserData(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Anonymous users have no server-side data
		if strings.HasPrefix(userID.(string), "anon_") {
			c.JSON(http.StatusOK, gin.H{
				"message": "Anonymous users have no server-side data to delete",
			})
			return
		}

		var req models.DataDeletionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if !req.Confirm {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Deletion must be confirmed"})
			return
		}

		ctx := context.Background()

		// Delete user sessions
		sessionsQuery := database.NewQuery("Session").Filter("user_id =", userID.(string)).KeysOnly()
		sessionKeys, err := db.GetAll(ctx, sessionsQuery, nil)
		if err == nil && len(sessionKeys) > 0 {
			err = db.DeleteMulti(ctx, sessionKeys)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete sessions"})
				return
			}
		}

		// Delete session events
		eventsQuery := database.NewQuery("SessionEvent").Filter("user_id =", userID.(string)).KeysOnly()
		eventKeys, err := db.GetAll(ctx, eventsQuery, nil)
		if err == nil && len(eventKeys) > 0 {
			err = db.DeleteMulti(ctx, eventKeys)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete session events"})
				return
			}
		}

		// Delete results
		resultsQuery := database.NewQuery("Result").Filter("user_id =", userID.(string)).KeysOnly()
		resultKeys, err := db.GetAll(ctx, resultsQuery, nil)
		if err == nil && len(resultKeys) > 0 {
			err = db.DeleteMulti(ctx, resultKeys)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete results"})
				return
			}
		}

		// Delete lesson progress
		progressQuery := database.NewQuery("LessonProgress").Filter("user_id =", userID.(string)).KeysOnly()
		progressKeys, err := db.GetAll(ctx, progressQuery, nil)
		if err == nil && len(progressKeys) > 0 {
			err = db.DeleteMulti(ctx, progressKeys)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete lesson progress"})
				return
			}
		}

		// Delete user profile
		userKey := database.NameKey("User", userID.(string), nil)
		err = db.Delete(ctx, userKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user profile"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "All user data has been permanently deleted",
			"deleted_at": time.Now().UTC().Format(time.RFC3339),
		})
	}
}

// GetConsentStatus retrieves user's current consent settings
func GetConsentStatus(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Anonymous users have no consent settings
		if strings.HasPrefix(userID.(string), "anon_") {
			c.JSON(http.StatusOK, gin.H{
				"is_anonymous":            true,
				"telemetry_consent":       false,
				"data_processing_consent": false,
			})
			return
		}

		ctx := context.Background()
		key := database.NameKey("User", userID.(string), nil)

		var user models.User
		err := db.Get(ctx, key, &user)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"is_anonymous":            user.IsAnonymous,
			"privacy_mode":            user.PrivacyMode,
			"telemetry_consent":       user.TelemetryConsent,
			"data_processing_consent": user.DataProcessingConsent,
		})
	}
}

// Helper function to export data as CSV
func exportDataAsCSV(c *gin.Context, userData map[string]interface{}) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=user_data_%s.csv", time.Now().Format("20060102")))

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Write profile data
	if profile, ok := userData["profile"].(models.UserResponse); ok {
		writer.Write([]string{"Profile Data"})
		writer.Write([]string{"Field", "Value"})
		writer.Write([]string{"ID", profile.ID})
		writer.Write([]string{"Handle", profile.Handle})
		writer.Write([]string{"Email", profile.Email})
		writer.Write([]string{"Locale", profile.Locale})
		writer.Write([]string{"Keyboard Layout", profile.KeyboardLayout})
		writer.Write([]string{"Privacy Mode", fmt.Sprintf("%t", profile.PrivacyMode)})
		writer.Write([]string{"Telemetry Consent", fmt.Sprintf("%t", profile.TelemetryConsent)})
		writer.Write([]string{"Data Processing Consent", fmt.Sprintf("%t", profile.DataProcessingConsent)})
		writer.Write([]string{"Created At", profile.CreatedAt})
		writer.Write([]string{"Updated At", profile.UpdatedAt})
		writer.Write([]string{})
	}

	// Write sessions data
	if sessions, ok := userData["sessions"].([]*models.Session); ok {
		writer.Write([]string{"Sessions"})
		writer.Write([]string{"ID", "Mode", "Language", "Started At", "Duration (ms)"})
		for _, session := range sessions {
			writer.Write([]string{
				session.ID,
				session.Mode,
				session.LanguageID,
				session.StartedAt.Format(time.RFC3339),
				fmt.Sprintf("%d", session.DurationMs),
			})
		}
		writer.Write([]string{})
	}

	// Write results data
	if results, ok := userData["results"].([]*models.Result); ok {
		writer.Write([]string{"Results"})
		writer.Write([]string{"Session ID", "CPM", "tWPM", "Raw Accuracy", "Token Accuracy", "Composite Score"})
		for _, result := range results {
			writer.Write([]string{
				result.SessionID,
				fmt.Sprintf("%.2f", result.CPM),
				fmt.Sprintf("%.2f", result.TWPM),
				fmt.Sprintf("%.2f", result.RawAccuracy),
				fmt.Sprintf("%.2f", result.TokenAccuracy),
				fmt.Sprintf("%d", result.CompositeScore),
			})
		}
	}
}

// AnonymizeUserData anonymizes user data for telemetry (privacy-preserving)
func AnonymizeUserData(userID string, sessionData map[string]interface{}) map[string]interface{} {
	anonymized := make(map[string]interface{})

	// Hash user ID for anonymization
	anonymized["user_hash"] = hashString(userID)

	// Only include non-identifying metrics
	if mode, ok := sessionData["mode"]; ok {
		anonymized["mode"] = mode
	}
	if language, ok := sessionData["language"]; ok {
		anonymized["language"] = language
	}
	if metrics, ok := sessionData["metrics"]; ok {
		anonymized["metrics"] = metrics
	}

	// Add timestamp but round to hour for privacy
	anonymized["timestamp_hour"] = time.Now().UTC().Truncate(time.Hour).Format(time.RFC3339)

	return anonymized
}

// Helper function to hash strings for anonymization
func hashString(s string) string {
	return fmt.Sprintf("%x", sha256Hash(s))
}

func sha256Hash(s string) []byte {
	h := sha256.Sum256([]byte(s))
	return h[:]
}
