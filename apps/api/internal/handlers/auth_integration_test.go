package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/typing-master-for-coding-backend/internal/auth"
	"github.com/typing-master-for-coding-backend/internal/middleware"
	"github.com/typing-master-for-coding-backend/internal/models"
)

func TestAnonymousSessionCreation(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	jwtSecret := "test-secret"

	router := gin.New()
	router.POST("/auth/anonymous", CreateAnonymousSession(jwtSecret))

	// Test request
	reqBody := models.AnonymousSessionRequest{
		DeviceID:       "test-device-123",
		KeyboardLayout: "qwerty",
		Locale:         "en-US",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/auth/anonymous", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	// Check user structure
	user, ok := response["user"].(map[string]interface{})
	if !ok {
		t.Error("User not found in response")
	}

	if user["id"] != "anon_test-device-123" {
		t.Errorf("Expected user ID 'anon_test-device-123', got %v", user["id"])
	}

	if user["is_anonymous"] != true {
		t.Error("User should be marked as anonymous")
	}

	// Check tokens
	tokens, ok := response["tokens"].(map[string]interface{})
	if !ok {
		t.Error("Tokens not found in response")
	}

	if tokens["token_type"] != "Bearer" {
		t.Errorf("Expected token type 'Bearer', got %v", tokens["token_type"])
	}
}

func TestJWTTokenValidation(t *testing.T) {
	jwtSecret := "test-secret"

	// Generate tokens for anonymous user
	tokens, err := auth.GenerateTokenPair("anon_test-device", "Anonymous", "", "user", true, jwtSecret)
	if err != nil {
		t.Errorf("Failed to generate tokens: %v", err)
	}

	// Validate the access token
	claims, err := auth.ValidateToken(tokens.AccessToken, jwtSecret)
	if err != nil {
		t.Errorf("Failed to validate token: %v", err)
	}

	if claims.UserID != "anon_test-device" {
		t.Errorf("Expected user ID 'anon_test-device', got %s", claims.UserID)
	}

	if claims.IsAnonymous != true {
		t.Error("User should be marked as anonymous")
	}

	if claims.Role != "user" {
		t.Errorf("Expected role 'user', got %s", claims.Role)
	}
}

func TestAuthMiddlewareWithAnonymousUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSecret := "test-secret"

	// Generate tokens for anonymous user
	tokens, err := auth.GenerateTokenPair("anon_test-device", "Anonymous", "", "user", true, jwtSecret)
	if err != nil {
		t.Errorf("Failed to generate tokens: %v", err)
	}

	// Setup router with auth middleware
	router := gin.New()
	router.Use(middleware.AuthMiddleware(jwtSecret))
	router.GET("/protected", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found"})
			return
		}

		isAnonymous, exists := c.Get("is_anonymous")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Anonymous flag not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id":      userID,
			"is_anonymous": isAnonymous,
		})
	})

	// Test request with anonymous token
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if response["user_id"] != "anon_test-device" {
		t.Errorf("Expected user ID 'anon_test-device', got %v", response["user_id"])
	}

	if response["is_anonymous"] != true {
		t.Error("User should be marked as anonymous")
	}
}

func TestNormalUserAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSecret := "test-secret"

	// Generate tokens for normal user
	tokens, err := auth.GenerateTokenPair("user-123", "testuser", "test@example.com", "user", false, jwtSecret)
	if err != nil {
		t.Errorf("Failed to generate tokens: %v", err)
	}

	// Setup router with auth middleware
	router := gin.New()
	router.Use(middleware.AuthMiddleware(jwtSecret))
	router.GET("/protected", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found"})
			return
		}

		isAnonymous, exists := c.Get("is_anonymous")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Anonymous flag not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id":      userID,
			"is_anonymous": isAnonymous,
		})
	})

	// Test request with normal user token
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if response["user_id"] != "user-123" {
		t.Errorf("Expected user ID 'user-123', got %v", response["user_id"])
	}

	if response["is_anonymous"] != false {
		t.Error("User should not be marked as anonymous")
	}
}
