package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/typing-master-for-coding-backend/internal/database"
	"github.com/typing-master-for-coding-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v56/github"
	"golang.org/x/oauth2"
)

// OAuthService handles OAuth authentication for third-party integrations
type OAuthService struct {
	db           *database.DatastoreClient
	oauthConfigs map[string]*oauth2.Config
	stateStore   map[string]*OAuthState
	httpClient   *http.Client
}

// OAuthState represents an OAuth state during the authentication flow
type OAuthState struct {
	State       string                 `json:"state"`
	Provider    string                 `json:"provider"`
	UserID      string                 `json:"user_id"`
	RedirectURI string                 `json:"redirect_uri"`
	Scopes      []string               `json:"scopes"`
	CreatedAt   time.Time              `json:"created_at"`
	ExpiresAt   time.Time              `json:"expires_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// OAuthToken represents an OAuth token response
type OAuthToken struct {
	AccessToken  string                 `json:"access_token"`
	TokenType    string                 `json:"token_type"`
	RefreshToken string                 `json:"refresh_token,omitempty"`
	ExpiresIn    int                    `json:"expires_in,omitempty"`
	Scope        string                 `json:"scope,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// OAuthUserInfo represents user information from OAuth provider
type OAuthUserInfo struct {
	ID       string                 `json:"id"`
	Login    string                 `json:"login"`
	Name     string                 `json:"name"`
	Email    string                 `json:"email"`
	Avatar   string                 `json:"avatar_url"`
	Location string                 `json:"location"`
	Company  string                 `json:"company"`
	Blog     string                 `json:"blog"`
	Metadata map[string]interface{} `json:"metadata"`
}

// NewOAuthService creates a new OAuth service
func NewOAuthService(db *database.DatastoreClient) *OAuthService {
	// Initialize OAuth configurations for different providers
	oauthConfigs := make(map[string]*oauth2.Config)

	// GitHub OAuth configuration
	githubConfig := &oauth2.Config{
		ClientID:     getEnvOrDefault("GITHUB_CLIENT_ID", ""),
		ClientSecret: getEnvOrDefault("GITHUB_CLIENT_SECRET", ""),
		Scopes:       []string{"user", "repo", "admin:repo_hook"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
	}
	oauthConfigs["github"] = githubConfig

	// GitLab OAuth configuration
	gitlabConfig := &oauth2.Config{
		ClientID:     getEnvOrDefault("GITLAB_CLIENT_ID", ""),
		ClientSecret: getEnvOrDefault("GITLAB_CLIENT_SECRET", ""),
		Scopes:       []string{"api", "read_repository"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://gitlab.com/oauth/authorize",
			TokenURL: "https://gitlab.com/oauth/token",
		},
	}
	oauthConfigs["gitlab"] = gitlabConfig

	// Bitbucket OAuth configuration
	bitbucketConfig := &oauth2.Config{
		ClientID:     getEnvOrDefault("BITBUCKET_CLIENT_ID", ""),
		ClientSecret: getEnvOrDefault("BITBUCKET_CLIENT_SECRET", ""),
		Scopes:       []string{"account", "repository"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://bitbucket.org/site/oauth2/authorize",
			TokenURL: "https://bitbucket.org/site/oauth2/access_token",
		},
	}
	oauthConfigs["bitbucket"] = bitbucketConfig

	return &OAuthService{
		db:           db,
		oauthConfigs: oauthConfigs,
		stateStore:   make(map[string]*OAuthState),
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// InitiateOAuthFlow initiates the OAuth authentication flow
func (s *OAuthService) InitiateOAuthFlow(c *gin.Context, provider, userID string, redirectURI string, scopes []string) error {
	// Check if provider is supported
	config, exists := s.oauthConfigs[provider]
	if !exists {
		return fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	// Generate state parameter
	state := s.generateSecureState()

	// Store OAuth state
	oauthState := &OAuthState{
		State:       state,
		Provider:    provider,
		UserID:      userID,
		RedirectURI: redirectURI,
		Scopes:      scopes,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(10 * time.Minute), // State expires in 10 minutes
		Metadata:    make(map[string]interface{}),
	}

	s.stateStore[state] = oauthState

	// Update config with redirect URI and scopes
	config.RedirectURL = redirectURI
	if len(scopes) > 0 {
		config.Scopes = scopes
	}

	// Generate authorization URL
	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	// Redirect user to OAuth provider
	c.Redirect(http.StatusFound, authURL)
	return nil
}

// HandleOAuthCallback handles the OAuth callback from the provider
func (s *OAuthService) HandleOAuthCallback(c *gin.Context) error {
	// Get state and authorization code
	state := c.Query("state")
	code := c.Query("code")
	errorParam := c.Query("error")

	if errorParam != "" {
		return fmt.Errorf("OAuth error: %s", errorParam)
	}

	if state == "" || code == "" {
		return fmt.Errorf("missing state or authorization code")
	}

	// Retrieve OAuth state
	oauthState, exists := s.stateStore[state]
	if !exists {
		return fmt.Errorf("invalid or expired state")
	}

	// Check if state has expired
	if time.Now().After(oauthState.ExpiresAt) {
		delete(s.stateStore, state)
		return fmt.Errorf("OAuth state has expired")
	}

	// Get OAuth configuration
	config, exists := s.oauthConfigs[oauthState.Provider]
	if !exists {
		return fmt.Errorf("unsupported OAuth provider: %s", oauthState.Provider)
	}

	// Exchange authorization code for access token
	config.RedirectURL = oauthState.RedirectURI
	token, err := config.Exchange(c.Request.Context(), code)
	if err != nil {
		return fmt.Errorf("failed to exchange authorization code: %v", err)
	}

	// Get user information from provider
	userInfo, err := s.getUserInfo(c.Request.Context(), oauthState.Provider, token.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to get user info: %v", err)
	}

	// Create or update integration
	err = s.createOrUpdateIntegration(c.Request.Context(), oauthState.Provider, oauthState.UserID, token, userInfo)
	if err != nil {
		return fmt.Errorf("failed to create integration: %v", err)
	}

	// Clean up state
	delete(s.stateStore, state)

	// Redirect to success page
	c.Redirect(http.StatusFound, oauthState.RedirectURI+"?success=true&provider="+oauthState.Provider)
	return nil
}

// RefreshToken refreshes an OAuth token
func (s *OAuthService) RefreshToken(ctx context.Context, provider, userID string) (*OAuthToken, error) {
	// Get user's integration
	integration, err := s.getUserIntegration(ctx, provider, userID)
	if err != nil {
		return nil, fmt.Errorf("integration not found: %v", err)
	}

	// Get refresh token from integration config
	refreshToken, ok := integration.Config["refresh_token"].(string)
	if !ok {
		return nil, fmt.Errorf("no refresh token available")
	}

	// Get OAuth configuration
	config, exists := s.oauthConfigs[provider]
	if !exists {
		return nil, fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	// Create token source with refresh token
	tokenSource := &oauth2.TokenSource{
		Token: &oauth2.Token{
			RefreshToken: refreshToken,
		},
	}

	// Refresh the token
	newToken, err := config.TokenSource(ctx, tokenSource).Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %v", err)
	}

	// Update integration with new token
	err = s.updateIntegrationToken(ctx, integration.ID, newToken)
	if err != nil {
		return nil, fmt.Errorf("failed to update integration: %v", err)
	}

	// Return token information
	oauthToken := &OAuthToken{
		AccessToken:  newToken.AccessToken,
		TokenType:    newToken.TokenType,
		RefreshToken: newToken.RefreshToken,
		ExpiresIn:    int(newToken.Expiry.Sub(time.Now()).Seconds()),
		Scope:        strings.Join(newToken.Scope, " "),
	}

	return oauthToken, nil
}

// RevokeToken revokes an OAuth token
func (s *OAuthService) RevokeToken(ctx context.Context, provider, userID string) error {
	// Get user's integration
	integration, err := s.getUserIntegration(ctx, provider, userID)
	if err != nil {
		return fmt.Errorf("integration not found: %v", err)
	}

	// Get access token
	accessToken, ok := integration.Config["access_token"].(string)
	if !ok {
		return fmt.Errorf("no access token available")
	}

	// Revoke token based on provider
	switch provider {
	case "github":
		err = s.revokeGitHubToken(ctx, accessToken)
	case "gitlab":
		err = s.revokeGitLabToken(ctx, accessToken)
	case "bitbucket":
		err = s.revokeBitbucketToken(ctx, accessToken)
	default:
		return fmt.Errorf("token revocation not supported for provider: %s", provider)
	}

	if err != nil {
		return fmt.Errorf("failed to revoke token: %v", err)
	}

	// Delete integration
	return s.deleteIntegration(ctx, integration.ID)
}

// ValidateToken validates if an OAuth token is still valid
func (s *OAuthService) ValidateToken(ctx context.Context, provider, accessToken string) (bool, error) {
	switch provider {
	case "github":
		return s.validateGitHubToken(ctx, accessToken)
	case "gitlab":
		return s.validateGitLabToken(ctx, accessToken)
	case "bitbucket":
		return s.validateBitbucketToken(ctx, accessToken)
	default:
		return false, fmt.Errorf("token validation not supported for provider: %s", provider)
	}
}

// GetOAuthURL returns the OAuth authorization URL for a provider
func (s *OAuthService) GetOAuthURL(provider, userID, redirectURI string, scopes []string) (string, error) {
	config, exists := s.oauthConfigs[provider]
	if !exists {
		return "", fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	// Generate state
	state := s.generateSecureState()

	// Store OAuth state
	oauthState := &OAuthState{
		State:       state,
		Provider:    provider,
		UserID:      userID,
		RedirectURI: redirectURI,
		Scopes:      scopes,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(10 * time.Minute),
		Metadata:    make(map[string]interface{}),
	}

	s.stateStore[state] = oauthState

	// Update config
	config.RedirectURL = redirectURI
	if len(scopes) > 0 {
		config.Scopes = scopes
	}

	// Generate authorization URL
	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	return authURL, nil
}

// Helper functions

func (s *OAuthService) generateSecureState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (s *OAuthService) getUserInfo(ctx context.Context, provider, accessToken string) (*OAuthUserInfo, error) {
	switch provider {
	case "github":
		return s.getGitHubUserInfo(ctx, accessToken)
	case "gitlab":
		return s.getGitLabUserInfo(ctx, accessToken)
	case "bitbucket":
		return s.getBitbucketUserInfo(ctx, accessToken)
	default:
		return nil, fmt.Errorf("user info retrieval not supported for provider: %s", provider)
	}
}

func (s *OAuthService) getGitHubUserInfo(ctx context.Context, accessToken string) (*OAuthUserInfo, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	userInfo := &OAuthUserInfo{
		ID:       fmt.Sprintf("%d", user.GetID()),
		Login:    user.GetLogin(),
		Name:     user.GetName(),
		Email:    user.GetEmail(),
		Avatar:   user.GetAvatarURL(),
		Location: user.GetLocation(),
		Company:  user.GetCompany(),
		Blog:     user.GetBlog(),
		Metadata: map[string]interface{}{
			"repos_count": user.GetPublicRepos(),
			"followers":   user.GetFollowers(),
			"following":   user.GetFollowing(),
			"created_at":  user.GetCreatedAt(),
		},
	}

	return userInfo, nil
}

func (s *OAuthService) getGitLabUserInfo(ctx context.Context, accessToken string) (*OAuthUserInfo, error) {
	// Implementation for GitLab user info
	// This would use GitLab API to get user information
	return nil, fmt.Errorf("GitLab user info retrieval not implemented")
}

func (s *OAuthService) getBitbucketUserInfo(ctx context.Context, accessToken string) (*OAuthUserInfo, error) {
	// Implementation for Bitbucket user info
	// This would use Bitbucket API to get user information
	return nil, fmt.Errorf("Bitbucket user info retrieval not implemented")
}

func (s *OAuthService) createOrUpdateIntegration(ctx context.Context, provider, userID string, token *oauth2.Token, userInfo *OAuthUserInfo) error {
	// Check if integration already exists
	q := s.db.Client.NewQuery("Integration").
		FilterField("UserID", "=", userID).
		FilterField("Provider", "=", provider)

	var integrations []*models.Integration
	keys, err := s.db.Client.GetAll(ctx, q, &integrations)
	if err != nil {
		return err
	}

	// Prepare integration config
	config := map[string]interface{}{
		"access_token": token.AccessToken,
		"token_type":   token.TokenType,
		"expires_at":   token.Expiry,
		"scopes":       strings.Join(token.Scope, ","),
		"user_info":    userInfo,
		"last_updated": time.Now(),
	}

	if token.RefreshToken != "" {
		config["refresh_token"] = token.RefreshToken
	}

	metadata := map[string]interface{}{
		"user_login": userInfo.Login,
		"user_name":  userInfo.Name,
		"user_email": userInfo.Email,
		"avatar":     userInfo.Avatar,
		"created_at": time.Now(),
	}

	for k, v := range userInfo.Metadata {
		metadata[k] = v
	}

	if len(integrations) > 0 {
		// Update existing integration
		integration := integrations[0]
		integration.Config = config
		integration.Metadata = metadata
		integration.Status = "active"
		integration.UpdatedAt = time.Now()

		_, err = s.db.Client.Put(ctx, integration)
		return err
	} else {
		// Create new integration
		integration := &models.Integration{
			UserID:    userID,
			Provider:  provider,
			Status:    "active",
			Config:    config,
			Metadata:  metadata,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, err = s.db.Client.Put(ctx, integration)
		return err
	}
}

func (s *OAuthService) getUserIntegration(ctx context.Context, provider, userID string) (*models.Integration, error) {
	q := s.db.Client.NewQuery("Integration").
		FilterField("UserID", "=", userID).
		FilterField("Provider", "=", provider).
		FilterField("Status", "=", "active")

	var integrations []*models.Integration
	keys, err := s.db.Client.GetAll(ctx, q, &integrations)
	if err != nil {
		return nil, err
	}

	if len(integrations) == 0 {
		return nil, fmt.Errorf("integration not found")
	}

	integration := integrations[0]
	integration.ID = keys[0].Name
	return integration, nil
}

func (s *OAuthService) updateIntegrationToken(ctx context.Context, integrationID string, token *oauth2.Token) error {
	key := s.db.Client.NameKey("Integration", integrationID)
	var integration models.Integration

	err := s.db.Client.Get(ctx, key, &integration)
	if err != nil {
		return err
	}

	// Update token in config
	integration.Config["access_token"] = token.AccessToken
	integration.Config["token_type"] = token.TokenType
	integration.Config["expires_at"] = token.Expiry
	integration.Config["last_updated"] = time.Now()

	if token.RefreshToken != "" {
		integration.Config["refresh_token"] = token.RefreshToken
	}

	integration.UpdatedAt = time.Now()

	_, err = s.db.Client.Put(ctx, &integration)
	return err
}

func (s *OAuthService) deleteIntegration(ctx context.Context, integrationID string) error {
	key := s.db.Client.NameKey("Integration", integrationID)
	return s.db.Client.Delete(ctx, key)
}

func (s *OAuthService) revokeGitHubToken(ctx context.Context, accessToken string) error {
	// GitHub doesn't provide a direct token revocation endpoint
	// Tokens can be revoked from user settings or by using the client ID/secret
	// For now, we'll just delete the integration
	return nil
}

func (s *OAuthService) revokeGitLabToken(ctx context.Context, accessToken string) error {
	// Implementation for GitLab token revocation
	return nil
}

func (s *OAuthService) revokeBitbucketToken(ctx context.Context, accessToken string) error {
	// Implementation for Bitbucket token revocation
	return nil
}

func (s *OAuthService) validateGitHubToken(ctx context.Context, accessToken string) (bool, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	_, _, err := client.Users.Get(ctx, "")
	return err == nil, nil
}

func (s *OAuthService) validateGitLabToken(ctx context.Context, accessToken string) (bool, error) {
	// Implementation for GitLab token validation
	return false, fmt.Errorf("GitLab token validation not implemented")
}

func (s *OAuthService) validateBitbucketToken(ctx context.Context, accessToken string) (bool, error) {
	// Implementation for Bitbucket token validation
	return false, fmt.Errorf("Bitbucket token validation not implemented")
}

func getEnvOrDefault(key, defaultValue string) string {
	// In a real implementation, this would use os.Getenv
	// For now, return default values
	return defaultValue
}
