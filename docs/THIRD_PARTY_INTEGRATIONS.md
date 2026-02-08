# Third-Party Integrations Documentation

This document provides comprehensive documentation for the third-party integrations implemented in Typing Master for Coding as per task 15.1.

## Overview

The third-party integrations enable seamless embedding of typing practice sessions into developer workflows and tools, allowing users to practice typing real code in their preferred development environments.

## Supported Integrations

### 1. VS Code Extension

#### Features
- **Embedded Practice Sessions**: Practice typing directly in VS Code with real-time metrics
- **GitHub Integration**: Share snippets to GitHub repositories and create pull requests
- **Language Support**: Auto-detect programming languages and provide syntax-aware practice
- **Leaderboard Integration**: View and contribute to global leaderboards
- **Configurable Settings**: Customize practice modes, duration, and display options

#### Installation
1. Install from VS Code Marketplace or sideload the extension
2. Configure API URL in VS Code settings: `typing-master.apiUrl`
3. Set GitHub token for sharing: `typing-master.githubToken`

#### Commands
- `typing-master.startPractice`: Start a new practice session
- `typing-master.practiceCurrentFile`: Practice typing the current file
- `typing-master.practiceSelection`: Practice typing selected text
- `typing-master.showLeaderboard`: Display global leaderboards
- `typing-master.shareToGitHub`: Share current snippet to GitHub
- `typing-master.syncWithGitHub`: Sync with GitHub repositories

#### Configuration Options
```json
{
  "typing-master.apiUrl": "https://api.typing-master-for-coding.com",
  "typing-master.practiceMode": "timed",
  "typing-master.duration": 3,
  "typing-master.showRealTimeMetrics": true,
  "typing-master.autoSaveResults": true,
  "typing-master.githubToken": ""
}
```

#### API Endpoints Used
- `POST /api/v1/integrations/embedded/session` - Create embedded practice session
- `POST /api/v1/integrations/github/share` - Share snippet to GitHub
- `GET /api/v1/integrations/github/repos` - Get GitHub repositories
- `GET /api/v1/public/leaderboard` - Get leaderboard data

### 2. GitHub Integration

#### Features
- **Snippet Sharing**: Share typing practice snippets to GitHub repositories
- **Pull Request Creation**: Automatically create PRs for shared snippets
- **Repository Sync**: Sync with repositories containing typing practice content
- **OAuth Authentication**: Secure GitHub token management
- **Content Analysis**: Automatically detect typing practice content in repositories

#### Authentication
Uses OAuth2 flow with GitHub personal access tokens. Tokens are stored securely and used for API calls.

#### API Endpoints
- `POST /api/v1/integrations/github/share` - Share snippet to GitHub
- `GET /api/v1/integrations/github/repos` - Get user repositories
- `POST /api/v1/integrations/oauth/github/flow` - Initiate OAuth flow
- `GET /api/v1/integrations/oauth/github/callback` - Handle OAuth callback
- `POST /api/v1/integrations/oauth/github/refresh` - Refresh OAuth token
- `DELETE /api/v1/integrations/oauth/github` - Revoke OAuth token

#### Sharing Process
1. User selects code snippet in VS Code
2. Extension prompts for repository details (owner, repo name, file path)
3. Optional: Create pull request with custom title and description
4. API creates file/branch and optionally PR
5. Returns file URL or PR URL for immediate access

#### Repository Detection
Automatically detects repositories with typing practice content using:
- File name patterns: `typing-practice`, `typing-snippet`, `code-typing`
- Directory patterns: `snippets/`, `examples/`, `practice/`, `typing/`
- Content analysis for common practice patterns

### 3. Workflow Embedding

#### Supported Platforms

##### GitHub Actions
- **Triggers**: push, pull_request, schedule
- **Features**: CI/CD integration, PR checks, automated testing
- **Configuration**: `.github/workflows/typing-practice.yml`

##### GitLab CI/CD
- **Triggers**: push, merge_request, schedule
- **Features**: CI/CD integration, MR checks, pipeline integration
- **Configuration**: `.gitlab-ci.yml`

##### Jenkins
- **Triggers**: SCM, timer, upstream
- **Features**: Pipeline as code, multibranch pipelines, shared libraries
- **Configuration**: `Jenkinsfile`

##### Azure Pipelines
- **Triggers**: Continuous integration, pull request, scheduled
- **Features**: YAML pipelines, classic pipelines, multi-stage pipelines
- **Configuration**: `azure-pipelines.yml`

#### API Endpoints
- `POST /api/v1/embed` - Create embedded session
- `GET /api/v1/embed/:embed_id` - Get embedded session
- `POST /api/v1/embed/ci/:platform` - Generate CI configuration
- `GET /api/v1/embed/platforms` - Get supported platforms
- `POST /api/v1/embed/webhook/:platform` - Handle platform webhooks

#### CI/CD Integration Process
1. User selects platform and configuration
2. API generates platform-specific CI configuration
3. Configuration includes:
   - Setup steps for Typing Master CLI
   - Practice execution with specified parameters
   - Result upload and reporting
   - Environment variables and secrets
4. User integrates configuration into their CI/CD pipeline

#### Embedded Session Features
- **Temporary Sessions**: 24-hour expiration for security
- **Platform-Specific UI**: Optimized interfaces for different platforms
- **Real-time Metrics**: Live typing statistics and progress
- **Result Reporting**: Standardized result formats (JUnit, JSON)
- **Webhook Support**: Real-time status updates to platforms

## API Reference

### Integration Management

#### Create Integration
```http
POST /api/v1/integrations
Authorization: Bearer <token>
Content-Type: application/json

{
  "provider": "github",
  "config": {
    "access_token": "github_token_here"
  }
}
```

#### Get Integrations
```http
GET /api/v1/integrations
Authorization: Bearer <token>
```

#### Delete Integration
```http
DELETE /api/v1/integrations/:id
Authorization: Bearer <token>
```

### GitHub Integration

#### Share Snippet to GitHub
```http
POST /api/v1/integrations/github/share
Authorization: Bearer <token>
Content-Type: application/json

{
  "access_token": "github_token",
  "repo_owner": "username",
  "repo_name": "repository",
  "file_path": "snippets/example.js",
  "content": "function test() { return 'hello'; }",
  "commit_message": "Add typing practice snippet",
  "create_pr": true,
  "pr_title": "Add JavaScript typing practice",
  "pr_description": "This PR adds a new typing practice snippet",
  "branch": "feature/typing-practice"
}
```

#### Get GitHub Repositories
```http
GET /api/v1/integrations/github/repos
Authorization: Bearer <github_token>
```

### Workflow Embedding

#### Create Embedded Session
```http
POST /api/v1/embed
Content-Type: application/json

{
  "user_id": "optional_user_id",
  "platform": "vscode",
  "content_type": "snippet",
  "content": "function test() { return 'hello'; }",
  "language": "javascript",
  "mode": "timed",
  "config": {
    "duration": 3,
    "show_metrics": true
  }
}
```

#### Get Embedded Session
```http
GET /api/v1/embed/:embed_id
```

#### Generate CI Configuration
```http
POST /api/v1/embed/ci/:platform
Content-Type: application/json

{
  "language": "javascript",
  "mode": "timed",
  "duration": 3,
  "upload_results": true
}
```

## Error Handling

### Standard Error Responses
```json
{
  "error": "Error message description",
  "code": "ERROR_CODE",
  "details": {}
}
```

### Common Error Codes
- `INVALID_REQUEST`: Malformed request body or parameters
- `UNAUTHORIZED`: Missing or invalid authentication
- `FORBIDDEN`: User lacks permission for the requested resource
- `NOT_FOUND`: Requested resource does not exist
- `RATE_LIMITED`: Too many requests, try again later
- `INTEGRATION_ERROR`: Third-party service integration failed
- `VALIDATION_ERROR`: Request validation failed

### GitHub-Specific Errors
- `INVALID_TOKEN`: GitHub token is invalid or expired
- `INSUFFICIENT_PERMISSIONS`: Token lacks required permissions
- `REPOSITORY_NOT_FOUND`: Repository does not exist or is inaccessible
- `FILE_CONFLICT`: File already exists in target location
- `PR_CREATION_FAILED`: Pull request creation failed

## Security Considerations

### Token Management
- GitHub tokens are stored encrypted in the database
- Tokens are never exposed in API responses
- Token refresh and revocation are supported
- Tokens have configurable expiration times

### Authentication
- JWT tokens for API authentication
- OAuth2 flow for third-party integrations
- Secure token storage and transmission
- Rate limiting to prevent abuse

### Data Privacy
- Temporary embedded sessions expire after 24 hours
- User data is anonymized where possible
- Compliance with GDPR and privacy regulations
- Audit logging for integration activities

## Development Guide

### Setting Up Development Environment

1. **Backend Setup**
   ```bash
   cd backend
   go mod download
   cp .env.example .env
   # Configure environment variables
   go run main.go
   ```

2. **VS Code Extension Setup**
   ```bash
   cd vscode-extension
   npm install
   npm run compile
   # Debug in VS Code using F5
   ```

3. **Testing**
   ```bash
   # Backend tests
   cd backend
   go test ./internal/handlers/...
   
   # Extension tests
   cd vscode-extension
   npm test
   ```

### Adding New Integrations

1. **Create Service**: Implement service in `backend/internal/services/`
2. **Add Handlers**: Create handlers in `backend/internal/handlers/`
3. **Update Router**: Add routes in `backend/internal/api/router.go`
4. **Add Tests**: Create comprehensive test coverage
5. **Update Documentation**: Document new integration features

### Best Practices

- Use dependency injection for services
- Implement proper error handling and logging
- Add comprehensive test coverage
- Follow Go and TypeScript conventions
- Document all public APIs
- Use environment variables for configuration
- Implement proper authentication and authorization
- Add monitoring and metrics collection

## Troubleshooting

### Common Issues

#### VS Code Extension Not Loading
- Check that the extension is properly compiled
- Verify API URL configuration
- Check network connectivity to backend
- Review VS Code developer console for errors

#### GitHub Integration Failing
- Verify GitHub token has required permissions
- Check token expiration and refresh
- Ensure repository access permissions
- Review GitHub API rate limits

#### Embedded Sessions Not Working
- Verify embed URL generation
- Check session expiration times
- Ensure proper CORS configuration
- Review platform-specific configurations

### Debugging Tools

- **Backend Logs**: Check application logs for detailed error information
- **VS Code Developer Console**: Use for extension debugging
- **Network Tab**: Monitor API requests and responses
- **GitHub API Explorer**: Test GitHub API calls directly

## Future Enhancements

### Planned Features
- **Additional IDE Support**: JetBrains, Vim, Emacs integrations
- **Advanced CI/CD**: More platform support and custom workflows
- **Team Features**: Organization-level integrations and sharing
- **Analytics**: Detailed integration usage statistics
- **Custom Platforms**: Support for custom workflow platforms

### Roadmap
1. **Q1 2024**: Additional IDE integrations (JetBrains, Vim)
2. **Q2 2024**: Enhanced team collaboration features
3. **Q3 2024**: Advanced analytics and reporting
4. **Q4 2024**: Custom platform framework and marketplace

## Support

For integration-related issues:
- **Documentation**: This guide and API reference
- **GitHub Issues**: Report bugs and feature requests
- **Community Forum**: Get help from other users
- **Developer Support**: Contact development team for technical assistance

## License

This integration framework is licensed under the MIT License. See LICENSE file for details.
