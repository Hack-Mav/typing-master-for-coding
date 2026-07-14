package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/typing-master-for-coding-backend/internal/database"
	"github.com/typing-master-for-coding-backend/internal/models"

	"github.com/google/go-github/v56/github"
	"golang.org/x/oauth2"
)

// GitHubService handles GitHub API interactions
type GitHubService struct {
	db         *database.DatastoreClient
	httpClient *http.Client
}

// NewGitHubService creates a new GitHub service instance
func NewGitHubService(db *database.DatastoreClient) *GitHubService {
	return &GitHubService{
		db:         db,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
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

// GitHubShareResponse represents the response from a GitHub share operation
type GitHubShareResponse struct {
	Success     bool                   `json:"success"`
	CommitSHA   string                 `json:"commit_sha,omitempty"`
	FileURL     string                 `json:"file_url,omitempty"`
	PullRequest *GitHubPullRequestInfo `json:"pull_request,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// GitHubPullRequestInfo contains information about a created pull request
type GitHubPullRequestInfo struct {
	Number int    `json:"number"`
	URL    string `json:"url"`
	State  string `json:"state"`
}

// GitHubRepoInfo contains repository information
type GitHubRepoInfo struct {
	Name          string    `json:"name"`
	FullName      string    `json:"full_name"`
	Description   string    `json:"description"`
	Language      string    `json:"language"`
	Stars         int       `json:"stargazers_count"`
	Forks         int       `json:"forks_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Private       bool      `json:"private"`
	DefaultBranch string    `json:"default_branch"`
}

// GitHubUserInfo contains user information
type GitHubUserInfo struct {
	Login     string    `json:"login"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Bio       string    `json:"bio"`
	Location  string    `json:"location"`
	Company   string    `json:"company"`
	Website   string    `json:"blog"`
	Repos     int       `json:"public_repos"`
	Followers int       `json:"followers"`
	Following int       `json:"following"`
	CreatedAt time.Time `json:"created_at"`
}

// ShareSnippetToGitHub shares a snippet to a GitHub repository
func (s *GitHubService) ShareSnippetToGitHub(ctx context.Context, req *GitHubShareRequest) (*GitHubShareResponse, error) {
	// Create GitHub client
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: req.AccessToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Determine branch
	branch := req.Branch
	if branch == "" {
		// Get default branch
		repo, _, err := client.Repositories.Get(ctx, req.RepoOwner, req.RepoName)
		if err != nil {
			return &GitHubShareResponse{
				Success: false,
				Error:   fmt.Sprintf("failed to get repository: %v", err),
			}, nil
		}
		branch = repo.GetDefaultBranch()
	}

	// Create file in repository
	fileContent := &github.RepositoryContentFileOptions{
		Message: github.String(req.CommitMessage),
		Content: []byte(req.Content),
		Branch:  github.String(branch),
	}

	fileResponse, _, err := client.Repositories.CreateFile(ctx, req.RepoOwner, req.RepoName, req.FilePath, fileContent)
	if err != nil {
		return &GitHubShareResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to create file: %v", err),
		}, nil
	}

	response := &GitHubShareResponse{
		Success:   true,
		CommitSHA: *fileResponse.Commit.SHA,
		FileURL:   fmt.Sprintf("https://github.com/%s/%s/blob/%s/%s", req.RepoOwner, req.RepoName, branch, req.FilePath),
	}

	// Create pull request if requested
	if req.CreatePR {
		prTitle := req.PRTitle
		if prTitle == "" {
			prTitle = fmt.Sprintf("Add typing practice snippet: %s", extractFilenameFromPath(req.FilePath))
		}

		prDescription := req.PRDescription
		if prDescription == "" {
			prDescription = fmt.Sprintf("Added typing practice snippet from Typing Master for Coding\n\nFile: %s\n\nThis snippet was shared from the Typing Master for Coding platform.", req.FilePath)
		}

		// Create a new branch for the PR
		branchName := fmt.Sprintf("typing-snippet-%d", time.Now().Unix())
		ref := &github.Reference{
			Ref: github.String(fmt.Sprintf("refs/heads/%s", branchName)),
			Object: &github.GitObject{
				SHA: github.String(*fileResponse.Commit.SHA),
			},
		}

		_, _, err = client.Git.CreateRef(ctx, req.RepoOwner, req.RepoName, ref)
		if err != nil {
			return &GitHubShareResponse{
				Success: false,
				Error:   fmt.Sprintf("failed to create branch: %v", err),
			}, nil
		}

		// Create pull request
		pr := &github.NewPullRequest{
			Title: github.String(prTitle),
			Head:  github.String(branchName),
			Base:  github.String(branch),
			Body:  github.String(prDescription),
		}

		prResp, _, err := client.PullRequests.Create(ctx, req.RepoOwner, req.RepoName, pr)
		if err != nil {
			return &GitHubShareResponse{
				Success: false,
				Error:   fmt.Sprintf("failed to create pull request: %v", err),
			}, nil
		}

		response.PullRequest = &GitHubPullRequestInfo{
			Number: prResp.GetNumber(),
			URL:    prResp.GetHTMLURL(),
			State:  prResp.GetState(),
		}
	}

	return response, nil
}

// GetUserRepositories retrieves user's GitHub repositories
func (s *GitHubService) GetUserRepositories(ctx context.Context, accessToken string) ([]GitHubRepoInfo, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	opt := &github.RepositoryListOptions{ListOptions: github.ListOptions{PerPage: 100}}
	var allRepos []GitHubRepoInfo

	for {
		repos, resp, err := client.Repositories.List(ctx, "", opt)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories: %v", err)
		}

		for _, repo := range repos {
			repoInfo := GitHubRepoInfo{
				Name:          repo.GetName(),
				FullName:      repo.GetFullName(),
				Description:   repo.GetDescription(),
				Language:      repo.GetLanguage(),
				Stars:         repo.GetStargazersCount(),
				Forks:         repo.GetForksCount(),
				CreatedAt:     repo.GetCreatedAt().Time,
				UpdatedAt:     repo.GetUpdatedAt().Time,
				Private:       repo.GetPrivate(),
				DefaultBranch: repo.GetDefaultBranch(),
			}
			allRepos = append(allRepos, repoInfo)
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

// GetUserInfo retrieves GitHub user information
func (s *GitHubService) GetUserInfo(ctx context.Context, accessToken string) (*GitHubUserInfo, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}

	userInfo := &GitHubUserInfo{
		Login:     user.GetLogin(),
		Name:      user.GetName(),
		Email:     user.GetEmail(),
		Bio:       user.GetBio(),
		Location:  user.GetLocation(),
		Company:   user.GetCompany(),
		Website:   user.GetBlog(),
		Repos:     user.GetPublicRepos(),
		Followers: user.GetFollowers(),
		Following: user.GetFollowing(),
		CreatedAt: user.GetCreatedAt().Time,
	}

	return userInfo, nil
}

// ValidateRepositoryAccess checks if the user has access to the specified repository
func (s *GitHubService) ValidateRepositoryAccess(ctx context.Context, accessToken, owner, repo string) (bool, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	_, _, err := client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return false, fmt.Errorf("repository access denied or not found: %v", err)
	}

	return true, nil
}

// CreateGitHubIntegration creates or updates a GitHub integration for a user
func (s *GitHubService) CreateGitHubIntegration(ctx context.Context, userID, accessToken string) error {
	// Get user info to validate the token
	userInfo, err := s.GetUserInfo(ctx, accessToken)
	if err != nil {
		return fmt.Errorf("failed to validate GitHub token: %v", err)
	}

	// Check if integration already exists
	q := database.NewQuery("Integration").
		Filter("UserID =", userID).
		Filter("Provider =", "github")

	var integrations []*models.Integration
	keys, err := s.db.GetAll(ctx, q, &integrations)
	if err != nil {
		return fmt.Errorf("failed to query existing integrations: %v", err)
	}

	// Prepare integration config
	config := map[string]interface{}{
		"access_token": accessToken,
		"user_login":   userInfo.Login,
		"user_name":    userInfo.Name,
		"created_at":   time.Now(),
	}

	metadata := map[string]interface{}{
		"repos_count": userInfo.Repos,
		"followers":   userInfo.Followers,
		"following":   userInfo.Following,
	}

	if len(integrations) > 0 {
		// Update existing integration
		integration := integrations[0]
		integration.Config = config
		integration.Metadata = metadata
		integration.Status = "active"
		integration.UpdatedAt = time.Now()

		_, err = s.db.Put(ctx, keys[0], integration)
		if err != nil {
			return fmt.Errorf("failed to update integration: %v", err)
		}
	} else {
		// Create new integration
		integration := &models.Integration{
			UserID:    userID,
			Provider:  "github",
			Status:    "active",
			Config:    config,
			Metadata:  metadata,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, err = s.db.Put(ctx, keys[0], integration)
		if err != nil {
			return fmt.Errorf("failed to create integration: %v", err)
		}
	}

	return nil
}

// GetGitHubIntegration retrieves a user's GitHub integration
func (s *GitHubService) GetGitHubIntegration(ctx context.Context, userID string) (*models.Integration, error) {
	q := database.NewQuery("Integration").
		Filter("UserID =", userID).
		Filter("Provider =", "github").
		Filter("Status =", "active")

	var integrations []*models.Integration
	keys, err := s.db.GetAll(ctx, q, &integrations)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub integration: %v", err)
	}

	if len(integrations) == 0 {
		return nil, fmt.Errorf("GitHub integration not found")
	}

	integration := integrations[0]
	integration.ID = keys[0].Name
	return integration, nil
}

// DeleteGitHubIntegration removes a user's GitHub integration
func (s *GitHubService) DeleteGitHubIntegration(ctx context.Context, userID string) error {
	q := database.NewQuery("Integration").
		Filter("UserID =", userID).
		Filter("Provider =", "github")

	var integrations []*models.Integration
	keys, err := s.db.GetAll(ctx, q, &integrations)
	if err != nil {
		return fmt.Errorf("failed to find GitHub integration: %v", err)
	}

	if len(integrations) == 0 {
		return fmt.Errorf("GitHub integration not found")
	}

	// Delete the integration
	for _, key := range keys {
		err = s.db.Delete(ctx, key)
		if err != nil {
			return fmt.Errorf("failed to delete integration: %v", err)
		}
	}

	return nil
}

// Helper functions

func extractFilenameFromPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return path
}

// IsGitHubTokenValid checks if a GitHub token is valid
func (s *GitHubService) IsGitHubTokenValid(ctx context.Context, accessToken string) bool {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	_, _, err := client.Users.Get(ctx, "")
	return err == nil
}

// countTypingSnippets counts typing practice snippets in a repository
func (s *GitHubService) countTypingSnippets(ctx context.Context, client *github.Client, repo *github.Repository) (int, error) {
	// Search for common typing practice file patterns
	searchTerms := []string{
		"typing-practice",
		"typing-snippet",
		"code-typing",
		"practice-snippet",
		"typing-master",
	}

	totalCount := 0
	for _, term := range searchTerms {
		query := fmt.Sprintf("%s repo:%s", term, repo.GetFullName())
		result, _, err := client.Search.Code(ctx, query, &github.SearchOptions{
			TextMatch: true,
		})
		if err != nil {
			continue
		}
		totalCount += result.GetTotal()
	}

	// Also check for common snippet directories
	snippetDirs := []string{"snippets/", "examples/", "practice/", "typing/"}
	for _, dir := range snippetDirs {
		query := fmt.Sprintf("filename:%s repo:%s", dir, repo.GetFullName())
		result, _, err := client.Search.Code(ctx, query, &github.SearchOptions{
			TextMatch: true,
		})
		if err != nil {
			continue
		}
		totalCount += result.GetTotal()
	}

	return totalCount, nil
}
