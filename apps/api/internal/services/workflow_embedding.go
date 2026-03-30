package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/typing-master-for-coding-backend/internal/database"
	"github.com/typing-master-for-coding-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// WorkflowEmbeddingService handles embedding typing practice in developer workflows
type WorkflowEmbeddingService struct {
	db *database.DatastoreClient
}

// NewWorkflowEmbeddingService creates a new workflow embedding service
func NewWorkflowEmbeddingService(db *database.DatastoreClient) *WorkflowEmbeddingService {
	return &WorkflowEmbeddingService{
		db: db,
	}
}

// EmbedRequest represents a request to embed typing practice in a workflow
type EmbedRequest struct {
	Platform    string                 `json:"platform" binding:"required"`
	ContentType string                 `json:"content_type" binding:"required"`
	Content     string                 `json:"content" binding:"required"`
	Language    string                 `json:"language" binding:"required"`
	Mode        string                 `json:"mode" binding:"required"`
	Config      map[string]interface{} `json:"config"`
	UserID      string                 `json:"user_id,omitempty"`
}

// EmbedResponse represents the response for an embed request
type EmbedResponse struct {
	EmbedID     string                 `json:"embed_id"`
	EmbedURL    string                 `json:"embed_url"`
	SessionID   string                 `json:"session_id"`
	Platform    string                 `json:"platform"`
	ContentType string                 `json:"content_type"`
	Config      map[string]interface{} `json:"config"`
	CreatedAt   time.Time              `json:"created_at"`
	ExpiresAt   time.Time              `json:"expires_at"`
}

// WorkflowPlatform represents a supported workflow platform
type WorkflowPlatform struct {
	Name        string                 `json:"name"`
	DisplayName string                 `json:"display_name"`
	Description string                 `json:"description"`
	Features    []string               `json:"features"`
	Config      map[string]interface{} `json:"config"`
}

// CIConfig represents CI/CD pipeline configuration
type CIConfig struct {
	Platform    string                 `json:"platform"`
	Trigger     string                 `json:"trigger"`
	Steps       []CIStep               `json:"steps"`
	Environment map[string]interface{} `json:"environment"`
	Artifacts   []CIArtifact           `json:"artifacts"`
}

// CIStep represents a step in a CI/CD pipeline
type CIStep struct {
	Name        string                 `json:"name"`
	Action      string                 `json:"action"`
	Command     string                 `json:"command"`
	Timeout     int                    `json:"timeout"`
	RetryPolicy map[string]interface{} `json:"retry_policy"`
}

// CIArtifact represents an artifact from CI/CD
type CIArtifact struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Type      string    `json:"type"`
	ExpiresAt time.Time `json:"expires_at"`
	Public    bool      `json:"public"`
}

// CreateEmbed creates an embedded typing practice session for workflow integration
func (s *WorkflowEmbeddingService) CreateEmbed(ctx context.Context, req *EmbedRequest) (*EmbedResponse, error) {
	// Validate platform
	if !s.isSupportedPlatform(req.Platform) {
		return nil, fmt.Errorf("unsupported platform: %s", req.Platform)
	}

	// Validate language
	if !s.isSupportedLanguage(req.Language) {
		return nil, fmt.Errorf("unsupported language: %s", req.Language)
	}

	// Create temporary snippet for embedded session
	tempSnippet := &models.Snippet{
		LanguageID:    req.Language,
		Title:         fmt.Sprintf("Embedded Practice - %s", req.Platform),
		SourceCode:    req.Content,
		Tags:          []string{"embedded", req.Platform, "workflow"},
		Difficulty:    3,                     // Default difficulty
		EstimatedTime: len(req.Content) / 50, // Rough estimate
		Checksum:      s.calculateChecksum(req.Content),
		CreatedAt:     time.Now(),
	}

	// Save temporary snippet
	snippetKey := database.NameKey("Snippet", tempSnippet.ID)
	_, err := s.db.Client.Put(ctx, snippetKey, tempSnippet)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary snippet: %v", err)
	}
	tempSnippet.ID = snippetKey.Name

	// Create session
	session := &models.Session{
		UserID:     req.UserID,
		Mode:       req.Mode,
		LanguageID: req.Language,
		SnippetID:  tempSnippet.ID,
		StartedAt:  time.Now(),
		Settings: map[string]interface{}{
			"embedded":     true,
			"platform":     req.Platform,
			"content_type": req.ContentType,
			"config":       req.Config,
			"temporary":    true,
		},
		CreatedAt: time.Now(),
	}

	// Save session
	sessionKey := database.NameKey("Session", session.ID)
	_, err = s.db.Client.Put(ctx, sessionKey, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}
	session.ID = sessionKey.Name

	// Generate embed configuration
	embedConfig := s.generateEmbedConfig(req.Platform, req.ContentType, req.Config)

	// Create embed record
	embed := &models.Embed{
		ID:          session.ID, // Use session ID as embed ID
		Platform:    req.Platform,
		ContentType: req.ContentType,
		SessionID:   session.ID,
		Config:      embedConfig,
		UserID:      req.UserID,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour), // Embed expires after 24 hours
	}

	// Generate embed URL
	embedURL := s.generateEmbedURL(embed.ID, req.Platform)

	response := &EmbedResponse{
		EmbedID:     embed.ID,
		EmbedURL:    embedURL,
		SessionID:   session.ID,
		Platform:    req.Platform,
		ContentType: req.ContentType,
		Config:      embedConfig,
		CreatedAt:   embed.CreatedAt,
		ExpiresAt:   embed.ExpiresAt,
	}

	return response, nil
}

// GetEmbed retrieves an embedded session configuration
func (s *WorkflowEmbeddingService) GetEmbed(ctx context.Context, embedID string) (*EmbedResponse, error) {
	// Get session
	sessionKey := database.NameKey("Session", embedID)
	var session models.Session
	if err := s.db.Client.Get(ctx, sessionKey, &session); err != nil {
		return nil, fmt.Errorf("session not found: %v", err)
	}

	// Check if embed has expired
	if session.EndedAt != nil && session.EndedAt.Before(time.Now()) {
		return nil, fmt.Errorf("embed has expired")
	}

	// Get snippet
	snippetKey := database.NameKey("Snippet", session.SnippetID)
	var snippet models.Snippet
	if err := s.db.Client.Get(ctx, snippetKey, &snippet); err != nil {
		return nil, fmt.Errorf("snippet not found: %v", err)
	}

	// Extract embed configuration from session settings
	embedConfig, _ := session.Settings["config"].(map[string]interface{})
	platform, _ := session.Settings["platform"].(string)
	contentType, _ := session.Settings["content_type"].(string)

	// Generate embed URL
	embedURL := s.generateEmbedURL(embedID, platform)

	response := &EmbedResponse{
		EmbedID:     embedID,
		EmbedURL:    embedURL,
		SessionID:   session.ID,
		Platform:    platform,
		ContentType: contentType,
		Config:      embedConfig,
		CreatedAt:   session.CreatedAt,
		ExpiresAt:   session.CreatedAt.Add(24 * time.Hour),
	}

	return response, nil
}

// GenerateCIConfig generates CI/CD pipeline configuration for typing practice integration
func (s *WorkflowEmbeddingService) GenerateCIConfig(ctx context.Context, platform string, config map[string]interface{}) (*CIConfig, error) {
	switch platform {
	case "github-actions":
		return s.generateGitHubActionsConfig(config)
	case "gitlab-ci":
		return s.generateGitLabCIConfig(config)
	case "jenkins":
		return s.generateJenkinsConfig(config)
	case "azure-pipelines":
		return s.generateAzurePipelinesConfig(config)
	default:
		return nil, fmt.Errorf("unsupported CI platform: %s", platform)
	}
}

// GetSupportedPlatforms returns list of supported workflow platforms
func (s *WorkflowEmbeddingService) GetSupportedPlatforms() []WorkflowPlatform {
	return []WorkflowPlatform{
		{
			Name:        "github-actions",
			DisplayName: "GitHub Actions",
			Description: "Integrate typing practice into GitHub Actions workflows",
			Features:    []string{"CI/CD integration", "Pull request checks", "Automated testing"},
			Config: map[string]interface{}{
				"supported_triggers": []string{"push", "pull_request", "schedule"},
				"yaml_template":      ".github/workflows/typing-practice.yml",
			},
		},
		{
			Name:        "gitlab-ci",
			DisplayName: "GitLab CI/CD",
			Description: "Integrate typing practice into GitLab CI/CD pipelines",
			Features:    []string{"CI/CD integration", "Merge request checks", "Pipeline integration"},
			Config: map[string]interface{}{
				"supported_triggers": []string{"push", "merge_request", "schedule"},
				"yaml_template":      ".gitlab-ci.yml",
			},
		},
		{
			Name:        "jenkins",
			DisplayName: "Jenkins",
			Description: "Integrate typing practice into Jenkins pipelines",
			Features:    []string{"Pipeline as code", "Multibranch pipelines", "Shared libraries"},
			Config: map[string]interface{}{
				"supported_triggers": []string{"scm", "timer", "upstream"},
				"file_template":      "Jenkinsfile",
			},
		},
		{
			Name:        "azure-pipelines",
			DisplayName: "Azure Pipelines",
			Description: "Integrate typing practice into Azure DevOps pipelines",
			Features:    []string{"YAML pipelines", "Classic pipelines", "Multi-stage pipelines"},
			Config: map[string]interface{}{
				"supported_triggers": []string{"continuous integration", "pull request", "scheduled"},
				"yaml_template":      "azure-pipelines.yml",
			},
		},
	}
}

// HandleEmbedWebhook handles webhooks from workflow platforms
func (s *WorkflowEmbeddingService) HandleEmbedWebhook(c *gin.Context) {
	platform := c.Param("platform")
	if !s.isSupportedPlatform(platform) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported platform"})
		return
	}

	// Parse webhook payload based on platform
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// Process webhook asynchronously
	go s.processWorkflowWebhook(c.Request.Context(), platform, payload)

	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// Helper functions

func (s *WorkflowEmbeddingService) isSupportedPlatform(platform string) bool {
	supportedPlatforms := []string{
		"github-actions", "gitlab-ci", "jenkins", "azure-pipelines",
		"vscode", "jetbrains", "vim", "emacs", "sublime",
	}

	for _, p := range supportedPlatforms {
		if p == platform {
			return true
		}
	}
	return false
}

func (s *WorkflowEmbeddingService) isSupportedLanguage(language string) bool {
	supportedLanguages := []string{
		"javascript", "typescript", "python", "java", "cpp", "c", "rust",
		"go", "ruby", "php", "csharp", "swift", "kotlin", "scala", "yaml",
	}

	for _, lang := range supportedLanguages {
		if lang == language {
			return true
		}
	}
	return false
}

func (s *WorkflowEmbeddingService) calculateChecksum(content string) string {
	// Simple checksum implementation - in production, use proper hashing
	return fmt.Sprintf("%x", len(content)*17)
}

func (s *WorkflowEmbeddingService) generateEmbedConfig(platform, contentType string, userConfig map[string]interface{}) map[string]interface{} {
	baseConfig := map[string]interface{}{
		"platform":     platform,
		"content_type": contentType,
		"theme":        "auto",
		"show_metrics": true,
		"allow_pause":  true,
		"auto_submit":  false,
		"fullscreen":   false,
	}

	// Merge user config
	for k, v := range userConfig {
		baseConfig[k] = v
	}

	// Add platform-specific defaults
	switch platform {
	case "github-actions":
		baseConfig["show_progress"] = true
		baseConfig["exit_on_complete"] = true
	case "vscode":
		baseConfig["integrated_ui"] = true
		baseConfig["editor_theme"] = "vscode"
	case "jetbrains":
		baseConfig["integrated_ui"] = true
		baseConfig["editor_theme"] = "jetbrains"
	}

	return baseConfig
}

func (s *WorkflowEmbeddingService) generateEmbedURL(embedID, platform string) string {
	baseURL := "https://api.typing-master-for-coding.com"
	return fmt.Sprintf("%s/embed/%s/%s", baseURL, platform, embedID)
}

func (s *WorkflowEmbeddingService) generateGitHubActionsConfig(config map[string]interface{}) (*CIConfig, error) {
	return &CIConfig{
		Platform: "github-actions",
		Trigger:  "push, pull_request",
		Steps: []CIStep{
			{
				Name:    "Setup Typing Master",
				Action:  "setup",
				Command: "curl -sSL https://install.typing-master.com | bash",
				Timeout: 300,
			},
			{
				Name:    "Run Typing Practice",
				Action:  "practice",
				Command: "typing-master practice --mode timed --duration 3 --language javascript",
				Timeout: 600,
			},
			{
				Name:    "Upload Results",
				Action:  "upload",
				Command: "typing-master upload --format junit --file results.xml",
				Timeout: 120,
			},
		},
		Environment: map[string]interface{}{
			"TYPING_MASTER_API_KEY": "${{ secrets.TYPING_MASTER_API_KEY }}",
			"TYPING_MASTER_USER":    "${{ github.actor }}",
		},
		Artifacts: []CIArtifact{
			{
				Name:      "typing-results",
				Path:      "results.xml",
				Type:      "junit",
				ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
				Public:    false,
			},
		},
	}, nil
}

func (s *WorkflowEmbeddingService) generateGitLabCIConfig(config map[string]interface{}) (*CIConfig, error) {
	return &CIConfig{
		Platform: "gitlab-ci",
		Trigger:  "push, merge_request",
		Steps: []CIStep{
			{
				Name:    "Setup Typing Master",
				Action:  "setup",
				Command: "curl -sSL https://install.typing-master.com | bash",
				Timeout: 300,
			},
			{
				Name:    "Run Typing Practice",
				Action:  "practice",
				Command: "typing-master practice --mode timed --duration 3 --language python",
				Timeout: 600,
			},
		},
		Environment: map[string]interface{}{
			"TYPING_MASTER_API_KEY": "$TYPING_MASTER_API_KEY",
			"TYPING_MASTER_USER":    "$GITLAB_USER_LOGIN",
		},
		Artifacts: []CIArtifact{
			{
				Name:      "typing-results",
				Path:      "results.xml",
				Type:      "junit",
				ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
				Public:    false,
			},
		},
	}, nil
}

func (s *WorkflowEmbeddingService) generateJenkinsConfig(config map[string]interface{}) (*CIConfig, error) {
	return &CIConfig{
		Platform: "jenkins",
		Trigger:  "scm, timer",
		Steps: []CIStep{
			{
				Name:    "Setup Typing Master",
				Action:  "setup",
				Command: "curl -sSL https://install.typing-master.com | bash",
				Timeout: 300,
			},
			{
				Name:    "Run Typing Practice",
				Action:  "practice",
				Command: "typing-master practice --mode accuracy --language java",
				Timeout: 600,
			},
		},
		Environment: map[string]interface{}{
			"TYPING_MASTER_API_KEY": "${TYPING_MASTER_API_KEY}",
			"TYPING_MASTER_USER":    "${BUILD_USER}",
		},
		Artifacts: []CIArtifact{
			{
				Name:      "typing-results",
				Path:      "target/typing-results.xml",
				Type:      "junit",
				ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
				Public:    false,
			},
		},
	}, nil
}

func (s *WorkflowEmbeddingService) generateAzurePipelinesConfig(config map[string]interface{}) (*CIConfig, error) {
	return &CIConfig{
		Platform: "azure-pipelines",
		Trigger:  "continuous integration, pull request",
		Steps: []CIStep{
			{
				Name:    "Setup Typing Master",
				Action:  "setup",
				Command: "curl -sSL https://install.typing-master.com | bash",
				Timeout: 300,
			},
			{
				Name:    "Run Typing Practice",
				Action:  "practice",
				Command: "typing-master practice --mode timed --duration 5 --language typescript",
				Timeout: 600,
			},
		},
		Environment: map[string]interface{}{
			"TYPING_MASTER_API_KEY": "$(TYPING_MASTER_API_KEY)",
			"TYPING_MASTER_USER":    "$(Build.RequestedFor)",
		},
		Artifacts: []CIArtifact{
			{
				Name:      "typing-results",
				Path:      "$(Build.ArtifactStagingDirectory)/results.xml",
				Type:      "junit",
				ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
				Public:    false,
			},
		},
	}, nil
}

func (s *WorkflowEmbeddingService) processWorkflowWebhook(ctx context.Context, platform string, payload map[string]interface{}) {
	// Process webhook based on platform
	switch platform {
	case "github-actions":
		s.processGitHubActionsWebhook(ctx, payload)
	case "gitlab-ci":
		s.processGitLabCIWebhook(ctx, payload)
	case "jenkins":
		s.processJenkinsWebhook(ctx, payload)
	case "azure-pipelines":
		s.processAzurePipelinesWebhook(ctx, payload)
	}
}

func (s *WorkflowEmbeddingService) processGitHubActionsWebhook(ctx context.Context, payload map[string]interface{}) {
	// Process GitHub Actions webhook
	fmt.Printf("Processing GitHub Actions webhook: %v\n", payload)
}

func (s *WorkflowEmbeddingService) processGitLabCIWebhook(ctx context.Context, payload map[string]interface{}) {
	// Process GitLab CI webhook
	fmt.Printf("Processing GitLab CI webhook: %v\n", payload)
}

func (s *WorkflowEmbeddingService) processJenkinsWebhook(ctx context.Context, payload map[string]interface{}) {
	// Process Jenkins webhook
	fmt.Printf("Processing Jenkins webhook: %v\n", payload)
}

func (s *WorkflowEmbeddingService) processAzurePipelinesWebhook(ctx context.Context, payload map[string]interface{}) {
	// Process Azure Pipelines webhook
	fmt.Printf("Processing Azure Pipelines webhook: %v\n", payload)
}
