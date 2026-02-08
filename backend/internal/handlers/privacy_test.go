package handlers

import (
	"net/http"
	"testing"

	"github.com/typing-master-for-coding-backend/internal/middleware"
	"github.com/typing-master-for-coding-backend/internal/models"
	"github.com/typing-master-for-coding-backend/internal/testutil"

	"github.com/stretchr/testify/assert"
)

// TestUpdatePrivacySettings tests privacy settings update endpoint
func TestUpdatePrivacySettings(t *testing.T) {
	router, mockDB, _ := testutil.SetupTestRouter()

	router.PUT("/api/v1/privacy/settings", middleware.AuthMiddleware("test-secret-key-for-testing-only"), UpdatePrivacySettings(mockDB))

	// Requirement 6: Privacy controls and GDPR compliance
	t.Run("Successful Privacy Settings Update", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")

		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		reqBody := models.PrivacySettingsRequest{
			PrivacyMode:           true,
			TelemetryConsent:      false,
			DataProcessingConsent: false,
		}

		w := testutil.MakeRequest(router, "PUT", "/api/v1/privacy/settings", reqBody, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		assert.Equal(t, true, response["privacy_mode"])
		assert.Equal(t, false, response["telemetry_consent"])
		assert.Equal(t, false, response["data_processing_consent"])
	})

	t.Run("Anonymous User Cannot Update Privacy Settings", func(t *testing.T) {
		token, _ := testutil.GenerateTestToken("anon_device123", "Anonymous", "", true)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		reqBody := models.PrivacySettingsRequest{
			PrivacyMode: true,
		}

		w := testutil.MakeRequest(router, "PUT", "/api/v1/privacy/settings", reqBody, headers)

		testutil.AssertErrorResponse(t, w, http.StatusForbidden, "Anonymous users cannot update privacy settings")
	})

	t.Run("Opt Out of All Tracking", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")

		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		reqBody := models.PrivacySettingsRequest{
			PrivacyMode:           true,
			TelemetryConsent:      false,
			DataProcessingConsent: false,
		}

		w := testutil.MakeRequest(router, "PUT", "/api/v1/privacy/settings", reqBody, headers)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestGetConsentStatus tests consent status retrieval
func TestGetConsentStatus(t *testing.T) {
	router, mockDB, _ := testutil.SetupTestRouter()

	router.GET("/api/v1/privacy/consent", middleware.AuthMiddleware("test-secret-key-for-testing-only"), GetConsentStatus(mockDB))

	t.Run("Get Consent Status for Registered User", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")

		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		w := testutil.MakeRequest(router, "GET", "/api/v1/privacy/consent", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		assert.Contains(t, response, "telemetry_consent")
		assert.Contains(t, response, "data_processing_consent")
		assert.Contains(t, response, "privacy_mode")
	})

	t.Run("Get Consent Status for Anonymous User", func(t *testing.T) {
		token, _ := testutil.GenerateTestToken("anon_device123", "Anonymous", "", true)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		w := testutil.MakeRequest(router, "GET", "/api/v1/privacy/consent", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		assert.Equal(t, true, response["is_anonymous"])
		assert.Equal(t, false, response["telemetry_consent"])
		assert.Equal(t, false, response["data_processing_consent"])
	})
}

// TestExportUserData tests GDPR data export functionality
func TestExportUserData(t *testing.T) {
	router, mockDB, _ := testutil.SetupTestRouter()

	router.GET("/api/v1/privacy/export", middleware.AuthMiddleware("test-secret-key-for-testing-only"), ExportUserData(mockDB))

	// Requirement 6: GDPR-friendly data export
	t.Run("Export User Data as JSON", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")

		// Create some test data
		testutil.CreateTestSession(mockDB, "session1", user.ID, "drill", "javascript")

		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		w := testutil.MakeRequest(router, "GET", "/api/v1/privacy/export?format=json", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		assert.Contains(t, response, "user_id")
		assert.Contains(t, response, "exported_at")
		assert.Contains(t, response, "data")

		data := response["data"].(map[string]interface{})
		assert.Contains(t, data, "profile")
		assert.Contains(t, data, "sessions")
	})

	t.Run("Export User Data as CSV", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")

		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		w := testutil.MakeRequest(router, "GET", "/api/v1/privacy/export?format=csv", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "text/csv")
		assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
	})

	t.Run("Anonymous User Export Returns Empty", func(t *testing.T) {
		token, _ := testutil.GenerateTestToken("anon_device123", "Anonymous", "", true)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		w := testutil.MakeRequest(router, "GET", "/api/v1/privacy/export", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		assert.Contains(t, response, "message")
		assert.Contains(t, response["message"], "Anonymous users have no server-side data")
	})

	t.Run("Export Includes All User Data Types", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")

		// Create comprehensive test data
		testutil.CreateTestSession(mockDB, "session1", user.ID, "drill", "javascript")
		testutil.CreateTestSession(mockDB, "session2", user.ID, "zen", "python")

		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		w := testutil.MakeRequest(router, "GET", "/api/v1/privacy/export", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		data := response["data"].(map[string]interface{})

		// Verify all data types are included
		profile := data["profile"].(map[string]interface{})
		assert.Equal(t, "testuser", profile["handle"])
		assert.Equal(t, "test@example.com", profile["email"])

		sessions := data["sessions"].([]interface{})
		assert.GreaterOrEqual(t, len(sessions), 2)
	})
}

// TestDeleteUserData tests GDPR data deletion functionality
func TestDeleteUserData(t *testing.T) {
	router, mockDB, _ := testutil.SetupTestRouter()

	router.POST("/api/v1/privacy/delete", middleware.AuthMiddleware("test-secret-key-for-testing-only"), DeleteUserData(mockDB))

	// Requirement 6: Right to deletion (GDPR compliance)
	t.Run("Successful User Data Deletion", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")

		// Create test data
		testutil.CreateTestSession(mockDB, "session1", user.ID, "drill", "javascript")

		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		reqBody := models.DataDeletionRequest{
			Confirm: true,
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/privacy/delete", reqBody, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		assert.Contains(t, response, "message")
		assert.Contains(t, response["message"], "permanently deleted")
		assert.Contains(t, response, "deleted_at")
	})

	t.Run("Deletion Requires Confirmation", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")

		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		reqBody := models.DataDeletionRequest{
			Confirm: false,
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/privacy/delete", reqBody, headers)

		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "Deletion must be confirmed")
	})

	t.Run("Anonymous User Deletion Returns Success", func(t *testing.T) {
		token, _ := testutil.GenerateTestToken("anon_device123", "Anonymous", "", true)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		reqBody := models.DataDeletionRequest{
			Confirm: true,
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/privacy/delete", reqBody, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		assert.Contains(t, response["message"], "Anonymous users have no server-side data")
	})

	t.Run("Missing Confirmation Field", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")

		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		reqBody := map[string]interface{}{
			// Missing confirm field
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/privacy/delete", reqBody, headers)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestAnonymizeUserData tests data anonymization helper
func TestAnonymizeUserData(t *testing.T) {
	t.Run("Anonymize Session Data", func(t *testing.T) {
		userID := "user123"
		sessionData := map[string]interface{}{
			"mode":     "drill",
			"language": "javascript",
			"metrics": map[string]interface{}{
				"wpm":      85.5,
				"accuracy": 95.2,
			},
			"user_email": "test@example.com", // Should be removed
		}

		anonymized := AnonymizeUserData(userID, sessionData)

		// Should contain hashed user ID
		assert.Contains(t, anonymized, "user_hash")
		assert.NotEqual(t, userID, anonymized["user_hash"])

		// Should contain non-identifying data
		assert.Equal(t, "drill", anonymized["mode"])
		assert.Equal(t, "javascript", anonymized["language"])
		assert.Contains(t, anonymized, "metrics")

		// Should not contain identifying information
		assert.NotContains(t, anonymized, "user_email")

		// Should have timestamp rounded to hour
		assert.Contains(t, anonymized, "timestamp_hour")
	})

	t.Run("Hash Consistency", func(t *testing.T) {
		userID := "user123"

		hash1 := hashString(userID)
		hash2 := hashString(userID)

		// Same input should produce same hash
		assert.Equal(t, hash1, hash2)

		// Different input should produce different hash
		hash3 := hashString("user456")
		assert.NotEqual(t, hash1, hash3)
	})
}

// TestPrivacyModeIsolation tests that privacy mode properly isolates data
func TestPrivacyModeIsolation(t *testing.T) {
	router, mockDB, _ := testutil.SetupTestRouter()

	router.PUT("/api/v1/privacy/settings", middleware.AuthMiddleware("test-secret-key-for-testing-only"), UpdatePrivacySettings(mockDB))
	router.GET("/api/v1/privacy/consent", middleware.AuthMiddleware("test-secret-key-for-testing-only"), GetConsentStatus(mockDB))

	t.Run("Privacy Mode Prevents Telemetry", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "privacyuser", "privacy@example.com")

		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		// Enable privacy mode
		reqBody := models.PrivacySettingsRequest{
			PrivacyMode:           true,
			TelemetryConsent:      false,
			DataProcessingConsent: false,
		}

		w := testutil.MakeRequest(router, "PUT", "/api/v1/privacy/settings", reqBody, headers)
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify settings were applied
		w = testutil.MakeRequest(router, "GET", "/api/v1/privacy/consent", nil, headers)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		assert.Equal(t, true, response["privacy_mode"])
		assert.Equal(t, false, response["telemetry_consent"])
	})
}
