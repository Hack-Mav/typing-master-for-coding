package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"typing-master-backend/internal/auth"
	"typing-master-backend/internal/cache"
	"typing-master-backend/internal/config"
	"typing-master-backend/internal/database"
	"typing-master-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// SetupTestRouter creates a test router with mock dependencies
func SetupTestRouter() (*gin.Engine, *database.DatastoreClient, *cache.InMemoryCache) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	mockDatastore := database.NewMockDatastore()
	mockDB := &database.DatastoreClient{
		Client:    nil,
		ProjectID: "test-project",
		IsMock:    true,
		Mock:      mockDatastore,
	}
	mockCache := cache.NewInMemoryCache(100, 5) // 5 minutes TTL
	
	return router, mockDB, mockCache
}

// CreateTestConfig creates a test configuration
func CreateTestConfig() *config.Config {
	return &config.Config{
		JWTSecret:      "test-secret-key-for-testing-only",
		AllowedOrigins: []string{"http://localhost:3000"},
		Port:           "8080",
	}
}

// MakeRequest makes an HTTP request to a test router
func MakeRequest(router *gin.Engine, method, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}
	
	req, _ := http.NewRequest(method, path, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	return w
}

// GenerateTestToken generates a JWT token for testing
func GenerateTestToken(userID, handle, email string, isAnonymous bool) (string, error) {
	tokens, err := auth.GenerateTokenPair(userID, handle, email, isAnonymous, "test-secret-key-for-testing-only")
	if err != nil {
		return "", err
	}
	return tokens.AccessToken, nil
}

// AssertJSONResponse asserts that the response matches expected JSON
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedBody interface{}) {
	assert.Equal(t, expectedStatus, w.Code)
	
	if expectedBody != nil {
		var actualBody interface{}
		err := json.Unmarshal(w.Body.Bytes(), &actualBody)
		assert.NoError(t, err)
		
		expectedJSON, _ := json.Marshal(expectedBody)
		var expectedMap interface{}
		json.Unmarshal(expectedJSON, &expectedMap)
		
		assert.Equal(t, expectedMap, actualBody)
	}
}

// CreateTestUser creates a test user in the mock database
func CreateTestUser(mockDB *database.DatastoreClient, userID, handle, email string) *models.User {
	passwordHash, _ := auth.HashPassword("password123")
	now := time.Now().UTC()
	
	user := &models.User{
		ID:                  userID,
		Handle:              handle,
		Email:               email,
		PasswordHash:        passwordHash,
		IsAnonymous:         false,
		Locale:              "en-US",
		KeyboardLayout:      "QWERTY",
		PrivacyMode:         false,
		TelemetryConsent:    true,
		DataProcessingConsent: true,
		Settings:            make(map[string]interface{}),
		CreatedAt:           now,
		UpdatedAt:           now,
		LastLoginAt:         &now,
	}
	
	mockDB.Put(context.Background(), nil, user)
	return user
}

// CreateTestLanguage creates a test language in the mock database
func CreateTestLanguage(mockDB *database.DatastoreClient, id, name string) *models.Language {
	language := &models.Language{
		ID:       id,
		Name:     name,
		Version:  1,
		ParserID: "tree-sitter-" + id,
		GrammarConfig: map[string]interface{}{
			"enabled": true,
		},
		WhitespaceRules: map[string]interface{}{
			"indent": "spaces",
		},
		CreatedBy: "admin",
		CreatedAt: time.Now().UTC(),
	}
	
	mockDB.Put(context.Background(), nil, language)
	return language
}

// CreateTestLesson creates a test lesson in the mock database
func CreateTestLesson(mockDB *database.DatastoreClient, id, languageID, title string) *models.Lesson {
	lesson := &models.Lesson{
		ID:               id,
		LanguageID:       languageID,
		Title:            title,
		Difficulty:       1,
		Objectives:       []string{"Learn basics"},
		Prerequisites:    []string{},
		EstimatedMinutes: 30,
		TokensCovered:    []string{"function", "variable"},
		SnippetIDs:       []string{},
		Version:          1,
		CreatedBy:        "admin",
		CreatedAt:        time.Now().UTC(),
	}
	
	mockDB.Put(context.Background(), nil, lesson)
	return lesson
}

// CreateTestSnippet creates a test snippet in the mock database
func CreateTestSnippet(mockDB *database.DatastoreClient, id, languageID, title, code string) *models.Snippet {
	snippet := &models.Snippet{
		ID:           id,
		LanguageID:   languageID,
		Title:        title,
		SourceCode:   code,
		Tags:         []string{"test"},
		Difficulty:   1,
		EstimatedTime: 5,
		Checksum:     "abc123",
		AccessibilityTags: map[string]interface{}{
			"line_count": 10,
		},
		CreatedBy: "admin",
		Version:   1,
		CreatedAt: time.Now().UTC(),
	}
	
	mockDB.Put(context.Background(), nil, snippet)
	return snippet
}

// CreateTestSession creates a test session in the mock database
func CreateTestSession(mockDB *database.DatastoreClient, id, userID, mode, languageID string) *models.Session {
	session := &models.Session{
		ID:         id,
		UserID:     userID,
		Mode:       mode,
		LanguageID: languageID,
		StartedAt:  time.Now().UTC(),
		Settings:   make(map[string]interface{}),
		CreatedAt:  time.Now().UTC(),
	}
	
	mockDB.Put(context.Background(), nil, session)
	return session
}

// AssertErrorResponse asserts that the response contains an error
func AssertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedErrorContains string) {
	assert.Equal(t, expectedStatus, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	errorMsg, exists := response["error"]
	assert.True(t, exists, "Response should contain an error field")
	
	if expectedErrorContains != "" {
		assert.Contains(t, errorMsg, expectedErrorContains)
	}
}

// ParseJSON parses JSON bytes into a target interface
func ParseJSON(data []byte, target interface{}) error {
	return json.Unmarshal(data, target)
}
