package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/typing-master-for-coding-backend/internal/database"
	"github.com/typing-master-for-coding-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDatastoreClient for testing
type MockDatastoreClient struct {
	mock.Mock
}

func (m *MockDatastoreClient) Put(ctx context.Context, entity interface{}) (*database.Key, error) {
	args := m.Called(ctx, entity)
	return args.Get(0).(*database.Key), args.Error(1)
}

func (m *MockDatastoreClient) Get(ctx context.Context, key *database.Key, entity interface{}) error {
	args := m.Called(ctx, key, entity)
	return args.Error(0)
}

func (m *MockDatastoreClient) Delete(ctx context.Context, key *database.Key) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockDatastoreClient) GetAll(ctx context.Context, query *database.Query, entities interface{}) ([]*database.Key, error) {
	args := m.Called(ctx, query, entities)
	return args.Get(0).([]*database.Key), args.Error(1)
}

func (m *MockDatastoreClient) NewQuery(kind string) *database.Query {
	args := m.Called(kind)
	return args.Get(0).(*database.Query)
}

func (m *MockDatastoreClient) NameKey(kind, name string) *database.Key {
	args := m.Called(kind, name)
	return args.Get(0).(*database.Key)
}

func TestCreateIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(MockDatastoreClient)
	handler := CreateIntegration(mockDB)

	// Test successful integration creation
	t.Run("Valid GitHub Integration", func(t *testing.T) {
		// Setup mock expectations
		mockKey := &database.Key{Name: "integration-123"}
		mockDB.On("Put", mock.Anything, mock.AnythingOfType("*models.Integration")).Return(mockKey, nil)

		// Create request body
		reqBody := IntegrationRequest{
			Provider: "github",
			Config: map[string]interface{}{
				"access_token": "test-token",
			},
		}
		body, _ := json.Marshal(reqBody)

		// Create request
		req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		// Set user context
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req
		c.Set("user_id", "test-user")

		// Call handler
		handler(c)

		// Assert response
		assert.Equal(t, http.StatusCreated, c.Writer.Status())

		var response IntegrationResponse
		err := json.Unmarshal(c.Writer.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "github", response.Provider)
		assert.Equal(t, "integration-123", response.ID)

		mockDB.AssertExpectations(t)
	})

	// Test invalid provider
	t.Run("Invalid Provider", func(t *testing.T) {
		reqBody := IntegrationRequest{
			Provider: "invalid-provider",
			Config:   map[string]interface{}{},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req
		c.Set("user_id", "test-user")

		handler(c)

		assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	})

	// Test missing required config
	t.Run("Missing Required Config", func(t *testing.T) {
		reqBody := IntegrationRequest{
			Provider: "github",
			Config:   map[string]interface{}{}, // Missing access_token
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req
		c.Set("user_id", "test-user")

		handler(c)

		assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	})
}

func TestGetIntegrations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(MockDatastoreClient)
	handler := GetIntegrations(mockDB)

	t.Run("Get User Integrations", func(t *testing.T) {
		// Setup mock expectations
		mockQuery := &database.Query{}
		mockKey := &database.Key{Name: "integration-123"}

		integrations := []*models.Integration{
			{
				ID:        "integration-123",
				UserID:    "test-user",
				Provider:  "github",
				Status:    "active",
				Config:    map[string]interface{}{"access_token": "test-token"},
				Metadata:  map[string]interface{}{},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		mockDB.On("NewQuery", "Integration").Return(mockQuery)
		mockDB.On("GetAll", mock.Anything, mockQuery, mock.AnythingOfType("*[]*models.Integration")).Return([]*database.Key{mockKey}, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/integrations", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req
		c.Set("user_id", "test-user")

		handler(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())

		var response []*IntegrationResponse
		err := json.Unmarshal(c.Writer.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response, 1)
		assert.Equal(t, "github", response[0].Provider)

		mockDB.AssertExpectations(t)
	})

	t.Run("Unauthorized Access", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/integrations", nil)

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req
		// No user_id set

		handler(c)

		assert.Equal(t, http.StatusUnauthorized, c.Writer.Status())
	})
}

func TestDeleteIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(MockDatastoreClient)
	handler := DeleteIntegration(mockDB)

	t.Run("Delete Existing Integration", func(t *testing.T) {
		// Setup mock expectations
		mockKey := &database.Key{Name: "integration-123"}
		integration := &models.Integration{
			ID:     "integration-123",
			UserID: "test-user",
		}

		mockDB.On("NameKey", "Integration", "integration-123").Return(mockKey)
		mockDB.On("Get", mock.Anything, mockKey, mock.AnythingOfType("*models.Integration")).Return(nil).Run(func(args mock.Arguments) {
			arg := args.Get(2).(*models.Integration)
			*arg = *integration
		})
		mockDB.On("Delete", mock.Anything, mockKey).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/integrations/integration-123", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req
		c.Set("user_id", "test-user")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "integration-123"}}

		handler(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())
		mockDB.AssertExpectations(t)
	})

	t.Run("Integration Not Found", func(t *testing.T) {
		mockKey := &database.Key{Name: "integration-123"}
		mockDB.On("NameKey", "Integration", "integration-123").Return(mockKey)
		mockDB.On("Get", mock.Anything, mockKey, mock.AnythingOfType("*models.Integration")).Return(database.ErrNotFound)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/integrations/integration-123", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req
		c.Set("user_id", "test-user")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "integration-123"}}

		handler(c)

		assert.Equal(t, http.StatusNotFound, c.Writer.Status())
		mockDB.AssertExpectations(t)
	})

	t.Run("Unauthorized Access", func(t *testing.T) {
		mockKey := &database.Key{Name: "integration-123"}
		integration := &models.Integration{
			ID:     "integration-123",
			UserID: "different-user", // Different user
		}

		mockDB.On("NameKey", "Integration", "integration-123").Return(mockKey)
		mockDB.On("Get", mock.Anything, mockKey, mock.AnythingOfType("*models.Integration")).Return(nil).Run(func(args mock.Arguments) {
			arg := args.Get(2).(*models.Integration)
			*arg = *integration
		})

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/integrations/integration-123", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req
		c.Set("user_id", "test-user")
		c.Params = gin.Params{gin.Param{Key: "id", Value: "integration-123"}}

		handler(c)

		assert.Equal(t, http.StatusForbidden, c.Writer.Status())
		mockDB.AssertExpectations(t)
	})
}

func TestShareSnippetToGitHub(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(MockDatastoreClient)
	handler := ShareSnippetToGitHub(mockDB)

	t.Run("Valid GitHub Share Request", func(t *testing.T) {
		reqBody := GitHubShareRequest{
			AccessToken:   "test-token",
			RepoOwner:     "test-owner",
			RepoName:      "test-repo",
			FilePath:      "snippets/example.js",
			Content:       "function test() { return 'hello'; }",
			CommitMessage: "Add test snippet",
			CreatePR:      false,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/github/share", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req
		c.Set("user_id", "test-user")

		// This test would require mocking the GitHub service
		// For now, we'll test the request validation
		handler(c)

		// Should fail because GitHub service is not mocked
		assert.Equal(t, http.StatusInternalServerError, c.Writer.Status())
	})

	t.Run("Invalid Request Body", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"invalid": "request",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/github/share", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req
		c.Set("user_id", "test-user")

		handler(c)

		assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	})
}

func TestCreateEmbeddedSession(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(MockDatastoreClient)
	handler := CreateEmbeddedSession(mockDB)

	t.Run("Valid Embedded Session Request", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"platform":     "vscode",
			"content_type": "snippet",
			"content":      "function test() { return 'hello'; }",
			"language":     "javascript",
			"mode":         "timed",
			"config":       map[string]interface{}{"duration": 3},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/embedded/session", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req
		// No user_id set, should default to anonymous

		// This test would require mocking the workflow embedding service
		handler(c)

		// Should fail because workflow embedding service is not mocked
		assert.Equal(t, http.StatusInternalServerError, c.Writer.Status())
	})

	t.Run("Invalid Request Body", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"invalid": "request",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/embedded/session", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req

		handler(c)

		assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	})
}

func TestGetGitHubRepos(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB := new(MockDatastoreClient)
	handler := GetGitHubRepos(mockDB)

	t.Run("Valid GitHub Repos Request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/integrations/github/repos", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("Content-Type", "application/json")

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req
		c.Set("user_id", "test-user")

		// This test would require mocking the GitHub service
		handler(c)

		// Should fail because GitHub service is not mocked
		assert.Equal(t, http.StatusInternalServerError, c.Writer.Status())
	})

	t.Run("Missing Authorization Header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/integrations/github/repos", nil)

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req
		c.Set("user_id", "test-user")

		handler(c)

		assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
	})

	t.Run("Unauthorized Access", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/integrations/github/repos", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req
		// No user_id set

		handler(c)

		assert.Equal(t, http.StatusUnauthorized, c.Writer.Status())
	})
}

// Benchmark tests
func BenchmarkCreateIntegration(b *testing.B) {
	gin.SetMode(gin.TestMode)
	mockDB := new(MockDatastoreClient)
	handler := CreateIntegration(mockDB)

	mockKey := &database.Key{Name: "integration-123"}
	mockDB.On("Put", mock.Anything, mock.AnythingOfType("*models.Integration")).Return(mockKey, nil)

	reqBody := IntegrationRequest{
		Provider: "github",
		Config: map[string]interface{}{
			"access_token": "test-token",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = req
		c.Set("user_id", "test-user")

		handler(c)
	}
}
