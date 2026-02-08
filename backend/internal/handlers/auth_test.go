package handlers

import (
	"net/http"
	"testing"

	"github.com/typing-master-for-coding-backend/internal/middleware"
	"github.com/typing-master-for-coding-backend/internal/models"
	"github.com/typing-master-for-coding-backend/internal/testutil"

	"github.com/stretchr/testify/assert"
)

// TestRegister tests user registration endpoint
func TestRegister(t *testing.T) {
	router, mockDB, _ := testutil.SetupTestRouter()

	router.POST("/api/v1/auth/register", Register(mockDB, "test-secret-key-for-testing-only"))

	t.Run("Successful Registration", func(t *testing.T) {
		mockDB.Clear()

		reqBody := models.RegisterRequest{
			Handle:                "testuser",
			Email:                 "test@example.com",
			Password:              "password123",
			Locale:                "en-US",
			KeyboardLayout:        "QWERTY",
			TelemetryConsent:      true,
			DataProcessingConsent: true,
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/auth/register", reqBody, nil)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := testutil.ParseJSON(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Contains(t, response, "user")
		assert.Contains(t, response, "tokens")

		user := response["user"].(map[string]interface{})
		assert.Equal(t, "testuser", user["handle"])
		assert.Equal(t, "test@example.com", user["email"])

		tokens := response["tokens"].(map[string]interface{})
		assert.Contains(t, tokens, "access_token")
		assert.Contains(t, tokens, "refresh_token")
	})

	t.Run("Duplicate Email", func(t *testing.T) {
		mockDB.Clear()
		testutil.CreateTestUser(mockDB, "user1", "existing", "test@example.com")

		reqBody := models.RegisterRequest{
			Handle:   "newuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/auth/register", reqBody, nil)

		testutil.AssertErrorResponse(t, w, http.StatusConflict, "Email already registered")
	})

	t.Run("Duplicate Handle", func(t *testing.T) {
		mockDB.Clear()
		testutil.CreateTestUser(mockDB, "user1", "testuser", "existing@example.com")

		reqBody := models.RegisterRequest{
			Handle:   "testuser",
			Email:    "new@example.com",
			Password: "password123",
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/auth/register", reqBody, nil)

		testutil.AssertErrorResponse(t, w, http.StatusConflict, "Handle already taken")
	})

	t.Run("Invalid Email Format", func(t *testing.T) {
		reqBody := models.RegisterRequest{
			Handle:   "testuser",
			Email:    "invalid-email",
			Password: "password123",
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/auth/register", reqBody, nil)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Password Too Short", func(t *testing.T) {
		reqBody := models.RegisterRequest{
			Handle:   "testuser",
			Email:    "test@example.com",
			Password: "short",
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/auth/register", reqBody, nil)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Missing Required Fields", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"email": "test@example.com",
			// Missing handle and password
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/auth/register", reqBody, nil)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Requirement 6: Privacy - Anonymous mode support
	t.Run("Privacy Consent Tracking", func(t *testing.T) {
		mockDB.Clear()

		reqBody := models.RegisterRequest{
			Handle:                "privacyuser",
			Email:                 "privacy@example.com",
			Password:              "password123",
			TelemetryConsent:      false,
			DataProcessingConsent: false,
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/auth/register", reqBody, nil)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		user := response["user"].(map[string]interface{})
		assert.Equal(t, false, user["telemetry_consent"])
		assert.Equal(t, false, user["data_processing_consent"])
	})
}

// TestLogin tests user login endpoint
func TestLogin(t *testing.T) {
	router, mockDB, _ := testutil.SetupTestRouter()

	router.POST("/api/v1/auth/login", Login(mockDB, "test-secret-key-for-testing-only"))

	t.Run("Successful Login", func(t *testing.T) {
		mockDB.Clear()
		testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")

		reqBody := models.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/auth/login", reqBody, nil)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		assert.Contains(t, response, "user")
		assert.Contains(t, response, "tokens")

		tokens := response["tokens"].(map[string]interface{})
		assert.Contains(t, tokens, "access_token")
		assert.Contains(t, tokens, "refresh_token")
	})

	t.Run("Invalid Email", func(t *testing.T) {
		mockDB.Clear()

		reqBody := models.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/auth/login", reqBody, nil)

		testutil.AssertErrorResponse(t, w, http.StatusUnauthorized, "Invalid email or password")
	})

	t.Run("Invalid Password", func(t *testing.T) {
		mockDB.Clear()
		testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")

		reqBody := models.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/auth/login", reqBody, nil)

		testutil.AssertErrorResponse(t, w, http.StatusUnauthorized, "Invalid email or password")
	})

	t.Run("Missing Credentials", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"email": "test@example.com",
			// Missing password
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/auth/login", reqBody, nil)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestRefreshToken tests token refresh endpoint
func TestRefreshToken(t *testing.T) {
	router, _, _ := testutil.SetupTestRouter()

	router.POST("/api/v1/auth/refresh", RefreshToken("test-secret-key-for-testing-only"))

	t.Run("Successful Token Refresh", func(t *testing.T) {
		// Generate a valid refresh token
		token, err := testutil.GenerateTestToken("user1", "testuser", "test@example.com", false)
		assert.NoError(t, err)

		reqBody := map[string]interface{}{
			"refresh_token": token,
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/auth/refresh", reqBody, nil)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		assert.Contains(t, response, "access_token")
	})

	t.Run("Invalid Refresh Token", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"refresh_token": "invalid.token.here",
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/auth/refresh", reqBody, nil)

		testutil.AssertErrorResponse(t, w, http.StatusUnauthorized, "Invalid or expired")
	})

	t.Run("Missing Refresh Token", func(t *testing.T) {
		reqBody := map[string]interface{}{}

		w := testutil.MakeRequest(router, "POST", "/api/v1/auth/refresh", reqBody, nil)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestCreateAnonymousSession tests anonymous session creation
func TestCreateAnonymousSession(t *testing.T) {
	router, _, _ := testutil.SetupTestRouter()

	router.POST("/api/v1/auth/anonymous", CreateAnonymousSession("test-secret-key-for-testing-only"))

	// Requirement 6: Anonymous mode with device-local storage
	t.Run("Successful Anonymous Session Creation", func(t *testing.T) {
		reqBody := models.AnonymousSessionRequest{
			DeviceID:       "device-123-456",
			KeyboardLayout: "QWERTY",
			Locale:         "en-US",
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/auth/anonymous", reqBody, nil)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		assert.Contains(t, response, "user")
		assert.Contains(t, response, "tokens")

		user := response["user"].(map[string]interface{})
		assert.Equal(t, true, user["is_anonymous"])
		assert.Equal(t, "Anonymous", user["handle"])
		assert.Contains(t, user["id"], "anon_")

		tokens := response["tokens"].(map[string]interface{})
		assert.Contains(t, tokens, "access_token")
	})

	t.Run("Missing Device ID", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"keyboard_layout": "QWERTY",
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/auth/anonymous", reqBody, nil)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Anonymous Session with Custom Keyboard Layout", func(t *testing.T) {
		reqBody := models.AnonymousSessionRequest{
			DeviceID:       "device-789",
			KeyboardLayout: "Dvorak",
			Locale:         "fr-FR",
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/auth/anonymous", reqBody, nil)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		user := response["user"].(map[string]interface{})
		assert.Equal(t, "Dvorak", user["keyboard_layout"])
		assert.Equal(t, "fr-FR", user["locale"])
	})
}

// TestGetProfile tests user profile retrieval
func TestGetProfile(t *testing.T) {
	router, mockDB, _ := testutil.SetupTestRouter()

	router.GET("/api/v1/profile", middleware.AuthMiddleware("test-secret-key-for-testing-only"), GetProfile(mockDB))

	t.Run("Successful Profile Retrieval", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")

		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		w := testutil.MakeRequest(router, "GET", "/api/v1/profile", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		assert.Equal(t, "testuser", response["handle"])
		assert.Equal(t, "test@example.com", response["email"])
		assert.NotContains(t, response, "password_hash")
	})

	t.Run("Anonymous User Profile", func(t *testing.T) {
		token, _ := testutil.GenerateTestToken("anon_device123", "Anonymous", "", true)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		w := testutil.MakeRequest(router, "GET", "/api/v1/profile", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		assert.Equal(t, true, response["is_anonymous"])
		assert.Equal(t, "Anonymous", response["handle"])
	})

	t.Run("Unauthorized Access", func(t *testing.T) {
		w := testutil.MakeRequest(router, "GET", "/api/v1/profile", nil, nil)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Bearer invalid.token.here",
		}

		w := testutil.MakeRequest(router, "GET", "/api/v1/profile", nil, headers)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestUpdateProfile tests user profile update
func TestUpdateProfile(t *testing.T) {
	router, mockDB, _ := testutil.SetupTestRouter()

	router.PUT("/api/v1/profile", middleware.AuthMiddleware("test-secret-key-for-testing-only"), UpdateProfile(mockDB))

	// Requirement 5: Customization - keyboard layouts and settings
	t.Run("Successful Profile Update", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")

		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		newHandle := "updateduser"
		newLocale := "fr-FR"
		newLayout := "Dvorak"

		reqBody := models.UpdateProfileRequest{
			Handle:         &newHandle,
			Locale:         &newLocale,
			KeyboardLayout: &newLayout,
		}

		w := testutil.MakeRequest(router, "PUT", "/api/v1/profile", reqBody, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		assert.Equal(t, "updateduser", response["handle"])
		assert.Equal(t, "fr-FR", response["locale"])
		assert.Equal(t, "Dvorak", response["keyboard_layout"])
	})

	t.Run("Anonymous User Cannot Update Profile", func(t *testing.T) {
		token, _ := testutil.GenerateTestToken("anon_device123", "Anonymous", "", true)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		newHandle := "newhandle"
		reqBody := models.UpdateProfileRequest{
			Handle: &newHandle,
		}

		w := testutil.MakeRequest(router, "PUT", "/api/v1/profile", reqBody, headers)

		testutil.AssertErrorResponse(t, w, http.StatusForbidden, "Anonymous users cannot update profile")
	})

	t.Run("Partial Update", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")

		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		newLocale := "es-ES"
		reqBody := models.UpdateProfileRequest{
			Locale: &newLocale,
		}

		w := testutil.MakeRequest(router, "PUT", "/api/v1/profile", reqBody, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		assert.Equal(t, "es-ES", response["locale"])
		assert.Equal(t, "testuser", response["handle"]) // Should remain unchanged
	})

	t.Run("Update Privacy Settings", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")

		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		privacyMode := true
		telemetry := false

		reqBody := models.UpdateProfileRequest{
			PrivacyMode:      &privacyMode,
			TelemetryConsent: &telemetry,
		}

		w := testutil.MakeRequest(router, "PUT", "/api/v1/profile", reqBody, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		assert.Equal(t, true, response["privacy_mode"])
		assert.Equal(t, false, response["telemetry_consent"])
	})
}
