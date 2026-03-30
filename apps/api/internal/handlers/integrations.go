package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/typing-master-for-coding-backend/internal/database"
	"github.com/typing-master-for-coding-backend/internal/models"
	"github.com/typing-master-for-coding-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// IntegrationRequest represents a request to create an integration
type IntegrationRequest struct {
	Provider string                 `json:"provider" binding:"required"`
	Config   map[string]interface{} `json:"config" binding:"required"`
}

// IntegrationResponse represents a response for integration operations
type IntegrationResponse struct {
	ID        string                 `json:"id"`
	Provider  string                 `json:"provider"`
	Status    string                 `json:"status"`
	Config    map[string]interface{} `json:"config"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// GitHubShareRequest represents a request to share content to GitHub
type GitHubShareRequest struct {
	AccessToken   string `json:"access_token" binding:"required"`
	RepoOwner     string `json:"repo_owner" binding:"required"`
	RepoName      string `json:"repo_name" binding:"required"`
	FilePath      string `json:"file_path" binding:"required"`
	Content       string `json:"content" binding:"required"`
	CommitMessage string `json:"commit_message" binding:"required"`
	CreatePR      bool   `json:"create_pr"`
	PRTitle       string `json:"pr_title,omitempty"`
	PRDescription string `json:"pr_description,omitempty"`
	Branch        string `json:"branch,omitempty"`
}

// CreateIntegration creates a new third-party integration
func CreateIntegration(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req IntegrationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		// Validate provider
		if !isValidProvider(req.Provider) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider"})
			return
		}

		// Validate config based on provider
		if err := validateIntegrationConfig(req.Provider, req.Config); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Create integration
		integration := &models.Integration{
			UserID:    userID,
			Provider:  req.Provider,
			Status:    "active",
			Config:    req.Config,
			Metadata:  make(map[string]interface{}),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		key := datastore.NameKey("Integration", userID+time.Now().String(), nil)

		// Save integration
		key, err := db.Client.Put(c.Request.Context(), key, integration)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create integration"})
			return
		}

		integration.ID = key.Name

		response := &IntegrationResponse{
			ID:        integration.ID,
			Provider:  integration.Provider,
			Status:    integration.Status,
			Config:    integration.Config,
			Metadata:  integration.Metadata,
			CreatedAt: integration.CreatedAt,
			UpdatedAt: integration.UpdatedAt,
		}

		c.JSON(http.StatusCreated, response)
	}
}

// GetIntegrations retrieves all integrations for the authenticated user
func GetIntegrations(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		// Query integrations
		q := db.NewQuery("Integration").
			FilterField("UserID", "=", userID).
			Order("-CreatedAt")

		var integrations []*models.Integration
		keys, err := db.Client.GetAll(c.Request.Context(), q, &integrations)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve integrations"})
			return
		}

		// Prepare response
		var responses []*IntegrationResponse
		for i, integration := range integrations {
			integration.ID = keys[i].Name
			response := &IntegrationResponse{
				ID:        integration.ID,
				Provider:  integration.Provider,
				Status:    integration.Status,
				Config:    integration.Config,
				Metadata:  integration.Metadata,
				CreatedAt: integration.CreatedAt,
				UpdatedAt: integration.UpdatedAt,
			}
			responses = append(responses, response)
		}

		c.JSON(http.StatusOK, responses)
	}
}

// DeleteIntegration deletes a third-party integration
func DeleteIntegration(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		integrationID := c.Param("id")
		if integrationID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "integration ID required"})
			return
		}

		// Get integration
		key := db.NameKey("Integration", integrationID, nil)
		var integration models.Integration
		if err := db.Client.Get(c.Request.Context(), key, &integration); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
			return
		}

		// Check ownership
		if integration.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}

		// Delete integration
		if err := db.Client.Delete(c.Request.Context(), key); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete integration"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "integration deleted successfully"})
	}
}

// ShareSnippetToGitHub shares a snippet to a GitHub repository
func ShareSnippetToGitHub(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req GitHubShareRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		// Create GitHub service
		githubService := services.NewGitHubService(db)

		// Prepare share request
		shareReq := &services.GitHubShareRequest{
			AccessToken:   req.AccessToken,
			RepoOwner:     req.RepoOwner,
			RepoName:      req.RepoName,
			FilePath:      req.FilePath,
			Content:       req.Content,
			CommitMessage: req.CommitMessage,
			CreatePR:      req.CreatePR,
			PRTitle:       req.PRTitle,
			PRDescription: req.PRDescription,
			Branch:        req.Branch,
		}

		// Share to GitHub
		response, err := githubService.ShareSnippetToGitHub(c.Request.Context(), shareReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if !response.Success {
			c.JSON(http.StatusBadRequest, gin.H{"error": response.Error})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

// CreateEmbeddedSession creates an embedded typing practice session
func CreateEmbeddedSession(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req services.EmbedRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID := c.GetString("user_id")
		if userID == "" {
			userID = "anonymous"
		}

		req.UserID = userID

		// Create embed service
		embedService := services.NewWorkflowEmbeddingService(db)

		// Create embed
		embed, err := embedService.CreateEmbed(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, embed)
	}
}

// GetEmbedSession retrieves an embedded session configuration
func GetEmbedSession(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("session_id")
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "session ID required"})
			return
		}

		// Create embed service
		embedService := services.NewWorkflowEmbeddingService(db)

		// Get embed
		embed, err := embedService.GetEmbed(c.Request.Context(), sessionID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, embed)
	}
}

// CreateEmbed creates an embedded typing practice session for workflow integration
func CreateEmbed(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req services.EmbedRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get user ID (optional for public embeds)
		userID := c.GetString("user_id")
		if userID == "" {
			userID = "anonymous"
		}

		req.UserID = userID

		// Create embed service
		embedService := services.NewWorkflowEmbeddingService(db)

		// Create embed
		embed, err := embedService.CreateEmbed(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, embed)
	}
}

// GetEmbed retrieves an embedded session configuration
func GetEmbed(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		embedID := c.Param("embed_id")
		if embedID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "embed ID required"})
			return
		}

		// Create embed service
		embedService := services.NewWorkflowEmbeddingService(db)

		// Get embed
		embed, err := embedService.GetEmbed(c.Request.Context(), embedID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, embed)
	}
}

// GenerateCIConfig generates CI/CD pipeline configuration for typing practice integration
func GenerateCIConfig(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		platform := c.Param("platform")
		if platform == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "platform required"})
			return
		}

		var config map[string]interface{}
		if err := c.ShouldBindJSON(&config); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Create embed service
		embedService := services.NewWorkflowEmbeddingService(db)

		// Generate CI config
		ciConfig, err := embedService.GenerateCIConfig(c.Request.Context(), platform, config)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, ciConfig)
	}
}

// GetSupportedPlatforms returns list of supported workflow platforms
func GetSupportedPlatforms(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create embed service
		embedService := services.NewWorkflowEmbeddingService(db)

		// Get supported platforms
		platforms := embedService.GetSupportedPlatforms()

		c.JSON(http.StatusOK, platforms)
	}
}

// HandleEmbedWebhook handles webhooks from workflow platforms
func HandleEmbedWebhook() gin.HandlerFunc {
	return func(c *gin.Context) {
		platform := c.Param("platform")
		if platform == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "platform required"})
			return
		}

		// Create embed service
		embedService := services.NewWorkflowEmbeddingService(nil)

		// Handle webhook
		embedService.HandleEmbedWebhook(c)
	}
}

// InitiateOAuthFlow initiates OAuth authentication for a provider
func InitiateOAuthFlow(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		provider := c.Param("provider")
		if provider == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "provider required"})
			return
		}

		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		redirectURI := c.Query("redirect_uri")
		if redirectURI == "" {
			redirectURI = "http://localhost:3000/oauth/callback"
		}

		scopes := strings.Split(c.DefaultQuery("scopes", "user,repo"), ",")

		// Create OAuth service
		oauthService := services.NewOAuthService(db)

		// Initiate OAuth flow
		err := oauthService.InitiateOAuthFlow(c, provider, userID, redirectURI, scopes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
}

// HandleOAuthCallback handles OAuth callback from providers
func HandleOAuthCallback(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create OAuth service
		oauthService := services.NewOAuthService(db)

		// Handle OAuth callback
		err := oauthService.HandleOAuthCallback(c)
		if err != nil {
			// Redirect to error page
			redirectURI := c.Query("redirect_uri")
			if redirectURI == "" {
				redirectURI = "http://localhost:3000/oauth/callback"
			}
			c.Redirect(http.StatusFound, redirectURI+"?error="+url.QueryEscape(err.Error()))
			return
		}
	}
}

// RefreshOAuthToken refreshes an OAuth token
func RefreshOAuthToken(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		provider := c.Param("provider")
		if provider == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "provider required"})
			return
		}

		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		// Create OAuth service
		oauthService := services.NewOAuthService(db)

		// Refresh token
		token, err := oauthService.RefreshToken(c.Request.Context(), provider, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, token)
	}
}

// RevokeOAuthToken revokes an OAuth token
func RevokeOAuthToken(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		provider := c.Param("provider")
		if provider == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "provider required"})
			return
		}

		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		// Create OAuth service
		oauthService := services.NewOAuthService(db)

		// Revoke token
		err := oauthService.RevokeToken(c.Request.Context(), provider, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "token revoked successfully"})
	}
}

// GetOAuthURL returns OAuth authorization URL
func GetOAuthURL(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		provider := c.Param("provider")
		if provider == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "provider required"})
			return
		}

		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		redirectURI := c.Query("redirect_uri")
		if redirectURI == "" {
			redirectURI = "http://localhost:3000/oauth/callback"
		}

		scopes := strings.Split(c.DefaultQuery("scopes", "user,repo"), ",")

		// Create OAuth service
		oauthService := services.NewOAuthService(db)

		// Get OAuth URL
		authURL, err := oauthService.GetOAuthURL(provider, userID, redirectURI, scopes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"auth_url": authURL})
	}
}

// GetGitHubRepos retrieves user's GitHub repositories with typing practice snippets
func GetGitHubRepos(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		accessToken := c.GetHeader("Authorization")
		if accessToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "GitHub access token required"})
			return
		}

		// Remove "Bearer " prefix if present
		if strings.HasPrefix(accessToken, "Bearer ") {
			accessToken = accessToken[7:]
		}

		// Create GitHub service
		githubService := services.NewGitHubService(db)

		// Get user repositories
		repos, err := githubService.GetUserRepositories(c.Request.Context(), accessToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, repos)
	}
}

// Helper functions

func isValidProvider(provider string) bool {
	validProviders := []string{"github", "gitlab", "bitbucket", "vscode", "jetbrains"}
	for _, p := range validProviders {
		if p == provider {
			return true
		}
	}
	return false
}

func validateIntegrationConfig(provider string, config map[string]interface{}) error {
	switch provider {
	case "github":
		if _, ok := config["access_token"]; !ok {
			return fmt.Errorf("access_token required for GitHub integration")
		}
	case "gitlab":
		if _, ok := config["access_token"]; !ok {
			return fmt.Errorf("access_token required for GitLab integration")
		}
	case "bitbucket":
		if _, ok := config["access_token"]; !ok {
			return fmt.Errorf("access_token required for Bitbucket integration")
		}
	case "vscode":
		if _, ok := config["extension_id"]; !ok {
			return fmt.Errorf("extension_id required for VS Code integration")
		}
	case "jetbrains":
		if _, ok := config["plugin_id"]; !ok {
			return fmt.Errorf("plugin_id required for JetBrains integration")
		}
	}
	return nil
}

func calculateChecksum(content string) string {
	// Simple checksum implementation - in production, use proper hashing
	return fmt.Sprintf("%x", len(content)*17)
}
