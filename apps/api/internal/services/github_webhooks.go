package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/typing-master-for-coding-backend/internal/database"

	"github.com/gin-gonic/gin"
)

// GitHubWebhookService handles GitHub webhook events
type GitHubWebhookService struct {
	db        *database.DatastoreClient
	secretKey string
}

// NewGitHubWebhookService creates a new GitHub webhook service
func NewGitHubWebhookService(db *database.DatastoreClient, secretKey string) *GitHubWebhookService {
	return &GitHubWebhookService{
		db:        db,
		secretKey: secretKey,
	}
}

// GitHubWebhookEvent represents a GitHub webhook event
type GitHubWebhookEvent struct {
	Type      string          `json:"type"`
	Action    string          `json:"action"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
	Processed bool            `json:"processed"`
}

// GitHubPushEvent represents a GitHub push event
type GitHubPushEvent struct {
	Ref        string           `json:"ref"`
	Before     string           `json:"before"`
	After      string           `json:"after"`
	Repository GitHubRepository `json:"repository"`
	Pusher     GitHubPusher     `json:"pusher"`
	Commits    []GitHubCommit   `json:"commits"`
}

// GitHubPullRequestEvent represents a GitHub pull request event
type GitHubPullRequestEvent struct {
	Action      string            `json:"action"`
	PullRequest GitHubPullRequest `json:"pull_request"`
	Repository  GitHubRepository  `json:"repository"`
	Sender      GitHubUser        `json:"sender"`
}

// GitHubRepository represents a GitHub repository
type GitHubRepository struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	FullName      string     `json:"full_name"`
	Owner         GitHubUser `json:"owner"`
	Private       bool       `json:"private"`
	Description   string     `json:"description"`
	Language      string     `json:"language"`
	DefaultBranch string     `json:"default_branch"`
}

// GitHubUser represents a GitHub user
type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// GitHubPusher represents a GitHub pusher
type GitHubPusher struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GitHubCommit represents a GitHub commit
type GitHubCommit struct {
	ID       string       `json:"id"`
	Message  string       `json:"message"`
	Author   GitHubAuthor `json:"author"`
	Added    []string     `json:"added"`
	Removed  []string     `json:"removed"`
	Modified []string     `json:"modified"`
	URL      string       `json:"url"`
}

// GitHubAuthor represents a GitHub commit author
type GitHubAuthor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GitHubPullRequest represents a GitHub pull request
type GitHubPullRequest struct {
	ID        int64          `json:"id"`
	Number    int            `json:"number"`
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	State     string         `json:"state"`
	User      GitHubUser     `json:"user"`
	Head      GitHubPRBranch `json:"head"`
	Base      GitHubPRBranch `json:"base"`
	URL       string         `json:"html_url"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// GitHubPRBranch represents a pull request branch
type GitHubPRBranch struct {
	Label string `json:"label"`
	Ref   string `json:"ref"`
	SHA   string `json:"sha"`
}

// HandleWebhook handles incoming GitHub webhook events
func (s *GitHubWebhookService) HandleWebhook(c *gin.Context) {
	// Get event type
	eventType := c.GetHeader("X-GitHub-Event")
	if eventType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing X-GitHub-Event header"})
		return
	}

	// Get signature
	signature := c.GetHeader("X-Hub-Signature-256")
	if signature == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing X-Hub-Signature-256 header"})
		return
	}

	// Read body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read request body"})
		return
	}

	// Verify signature
	if !s.verifySignature(body, signature) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	// Parse event action
	action := c.GetHeader("X-GitHub-Event")
	if strings.Contains(eventType, ":") {
		parts := strings.Split(eventType, ":")
		if len(parts) == 2 {
			eventType = parts[0]
			action = parts[1]
		}
	}

	// Store webhook event
	webhookEvent := &GitHubWebhookEvent{
		Type:      eventType,
		Action:    action,
		Payload:   body,
		Timestamp: time.Now(),
		Processed: false,
	}

	// Save webhook event to database
	err = s.storeWebhookEvent(c.Request.Context(), webhookEvent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store webhook event"})
		return
	}

	// Process the event based on type
	go s.processWebhookEvent(c.Request.Context(), webhookEvent)

	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// verifySignature verifies the GitHub webhook signature
func (s *GitHubWebhookService) verifySignature(body []byte, signature string) bool {
	if s.secretKey == "" {
		return true // Skip verification if no secret is configured
	}

	mac := hmac.New(sha256.New, []byte(s.secretKey))
	mac.Write(body)
	expectedSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// storeWebhookEvent stores a webhook event in the database
func (s *GitHubWebhookService) storeWebhookEvent(ctx context.Context, event *GitHubWebhookEvent) error {
	// Note: This would require creating a WebhookEvent model in the database
	// For now, we'll just log the event
	fmt.Printf("Received GitHub webhook event: %s:%s at %s\n", event.Type, event.Action, event.Timestamp)
	return nil
}

// processWebhookEvent processes a webhook event asynchronously
func (s *GitHubWebhookService) processWebhookEvent(ctx context.Context, event *GitHubWebhookEvent) {
	switch event.Type {
	case "push":
		s.handlePushEvent(ctx, event)
	case "pull_request":
		s.handlePullRequestEvent(ctx, event)
	case "repository":
		s.handleRepositoryEvent(ctx, event)
	default:
		fmt.Printf("Unhandled GitHub webhook event: %s\n", event.Type)
	}

	// Mark event as processed
	event.Processed = true
	_ = s.storeWebhookEvent(ctx, event)
}

// handlePushEvent handles push events
func (s *GitHubWebhookService) handlePushEvent(ctx context.Context, event *GitHubWebhookEvent) {
	var pushEvent GitHubPushEvent
	if err := json.Unmarshal(event.Payload, &pushEvent); err != nil {
		fmt.Printf("Failed to parse push event: %v\n", err)
		return
	}

	// Check if any files were added/modified that might be typing practice snippets
	for _, commit := range pushEvent.Commits {
		allFiles := append(append(commit.Added, commit.Removed...), commit.Modified...)
		for _, file := range allFiles {
			if s.isTypingPracticeFile(file) {
				s.processTypingPracticeFile(ctx, pushEvent.Repository, file, commit)
			}
		}
	}
}

// handlePullRequestEvent handles pull request events
func (s *GitHubWebhookService) handlePullRequestEvent(ctx context.Context, event *GitHubWebhookEvent) {
	var prEvent GitHubPullRequestEvent
	if err := json.Unmarshal(event.Payload, &prEvent); err != nil {
		fmt.Printf("Failed to parse pull request event: %v\n", err)
		return
	}

	// Check if PR contains typing practice snippets
	if s.isTypingPracticePR(&prEvent.PullRequest) {
		s.processTypingPracticePR(ctx, &prEvent)
	}
}

// handleRepositoryEvent handles repository events
func (s *GitHubWebhookService) handleRepositoryEvent(ctx context.Context, event *GitHubWebhookEvent) {
	// Handle repository creation, deletion, etc.
	fmt.Printf("Repository event: %s\n", event.Action)
}

// isTypingPracticeFile checks if a file is likely a typing practice snippet
func (s *GitHubWebhookService) isTypingPracticeFile(filename string) bool {
	// Check file extensions
	practiceExtensions := []string{
		".js", ".ts", ".py", ".java", ".cpp", ".c", ".rs", ".go", ".rb", ".php",
	}

	for _, ext := range practiceExtensions {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}

	// Check file patterns
	practicePatterns := []string{
		"practice", "snippet", "typing", "exercise", "drill",
	}

	lowerFilename := strings.ToLower(filename)
	for _, pattern := range practicePatterns {
		if strings.Contains(lowerFilename, pattern) {
			return true
		}
	}

	return false
}

// isTypingPracticePR checks if a pull request is related to typing practice
func (s *GitHubWebhookService) isTypingPracticePR(pr *GitHubPullRequest) bool {
	text := strings.ToLower(pr.Title + " " + pr.Body)
	practiceKeywords := []string{
		"typing", "practice", "snippet", "exercise", "drill",
		"typing master", "typing-practice",
	}

	for _, keyword := range practiceKeywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}

	return false
}

// processTypingPracticeFile processes a file that might be a typing practice snippet
func (s *GitHubWebhookService) processTypingPracticeFile(ctx context.Context, repo GitHubRepository, filename string, commit GitHubCommit) {
	// This would typically:
	// 1. Fetch the file content from GitHub
	// 2. Analyze if it's suitable for typing practice
	// 3. Create a snippet if appropriate
	// 4. Notify relevant users

	fmt.Printf("Processing typing practice file: %s in %s\n", filename, repo.FullName)

	// Create a notification or update for users who might be interested
	// This is where you could integrate with your notification system
}

// processTypingPracticePR processes a pull request related to typing practice
func (s *GitHubWebhookService) processTypingPracticePR(ctx context.Context, prEvent *GitHubPullRequestEvent) {
	// This would typically:
	// 1. Analyze the PR for typing practice content
	// 2. Extract snippets from the PR
	// 3. Create snippets in the system
	// 4. Notify users about new practice content

	fmt.Printf("Processing typing practice PR: #%d in %s\n",
		prEvent.PullRequest.Number, prEvent.Repository.FullName)
}

// GetWebhookConfig returns webhook configuration for setting up GitHub webhooks
func (s *GitHubWebhookService) GetWebhookConfig() map[string]interface{} {
	return map[string]interface{}{
		"url":          "https://api.typing-master-for-coding.com/webhooks/github",
		"content_type": "json",
		"secret":       s.secretKey,
		"events": []string{
			"push",
			"pull_request",
			"repository",
		},
		"active": true,
	}
}

// SetupWebhook sets up a webhook for a GitHub repository
func (s *GitHubWebhookService) SetupWebhook(ctx context.Context, accessToken, owner, repo string) error {
	// This would use the GitHub API to create a webhook
	// Implementation would require GitHub client setup similar to GitHubService
	fmt.Printf("Setting up webhook for %s/%s\n", owner, repo)
	return nil
}

// RemoveWebhook removes a webhook from a GitHub repository
func (s *GitHubWebhookService) RemoveWebhook(ctx context.Context, accessToken, owner, repo string, webhookID int64) error {
	// This would use the GitHub API to remove a webhook
	fmt.Printf("Removing webhook %d from %s/%s\n", webhookID, owner, repo)
	return nil
}
