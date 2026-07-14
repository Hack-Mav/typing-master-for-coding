# Third-Party Integrations Setup Guide

This guide provides step-by-step instructions for setting up and configuring the third-party integrations for Typing Master for Coding.

## Quick Setup Overview

1. **VS Code Extension**: Install and configure the VS Code extension
2. **GitHub Integration**: Set up GitHub token and repository access
3. **Workflow Embedding**: Configure CI/CD pipeline integration
4. **Testing**: Verify all integrations are working correctly

## 1. VS Code Extension Setup

### Installation

#### Option A: From VS Code Marketplace (Recommended)
1. Open VS Code
2. Go to Extensions (Ctrl+Shift+X)
3. Search for "Typing Master for Coding"
4. Click Install

#### Option B: Sideloading (Development)
1. Clone the repository
2. Navigate to `vscode-extension/` directory
3. Run `npm install`
4. Run `npm run compile`
5. Press F5 in VS Code to open a new Extension Development Host

### Configuration

1. Open VS Code Settings (Ctrl+,)
2. Search for "typing-master"
3. Configure the following settings:

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

### Verification

1. Open any code file
2. Right-click and select "Practice Current File"
3. Verify the typing practice session opens correctly
4. Check that real-time metrics are displayed

## 2. GitHub Integration Setup

### Creating GitHub Personal Access Token

1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Configure token permissions:
   - **Repo**: Full control of private repositories
   - **Workflow**: Update GitHub Action workflows (if creating PRs)
4. Set expiration (recommended: 90 days)
5. Generate token and copy it immediately

### Configuring GitHub Token in VS Code

1. Open VS Code Command Palette (Ctrl+Shift+P)
2. Type "Share to GitHub" and select the command
3. When prompted, enter your GitHub token
4. Choose whether to save the token to VS Code settings

### Testing GitHub Integration

1. Select some code in VS Code
2. Right-click and select "Share to GitHub"
3. Enter repository details:
   - Repository owner: your GitHub username
   - Repository name: target repository
   - File path: where to save the snippet
4. Choose whether to create a pull request
5. Verify the file/PR is created successfully

### Alternative: OAuth Flow

For production environments or team setups:

1. Create a GitHub OAuth App:
   - Go to GitHub Settings → Developer settings → OAuth Apps
   - Create new OAuth App
   - Set callback URL: `https://api.typing-master-for-coding.com/api/v1/integrations/oauth/github/callback`

2. Configure environment variables:
   ```bash
   GITHUB_CLIENT_ID=your_client_id
   GITHUB_CLIENT_SECRET=your_client_secret
   ```

3. Use OAuth endpoints for authentication instead of personal access tokens

## 3. Workflow Embedding Setup

### GitHub Actions Integration

1. Create workflow file:
   ```bash
   mkdir -p .github/workflows
   touch .github/workflows/typing-practice.yml
   ```

2. Generate configuration using API:
   ```bash
   curl -X POST https://api.typing-master-for-coding.com/api/v1/embed/ci/github-actions \
     -H "Content-Type: application/json" \
     -d '{
       "language": "javascript",
       "mode": "timed",
       "duration": 3
     }'
   ```

3. Add generated configuration to workflow file:
   ```yaml
   name: Typing Practice Check
   on: [push, pull_request]
   jobs:
     typing-practice:
       runs-on: ubuntu-latest
       steps:
         - name: Setup Typing Master
           run: curl -sSL https://install.typing-master.com | bash
         - name: Run Typing Practice
           run: typing-master practice --mode timed --duration 3 --language javascript
         - name: Upload Results
           run: typing-master upload --format junit --file results.xml
   ```

4. Add secrets to repository:
   - Go to repository Settings → Secrets and variables → Actions
   - Add `TYPING_MASTER_API_KEY` secret

### GitLab CI/CD Integration

1. Create `.gitlab-ci.yml` file:
   ```yaml
   typing-practice:
     stage: test
     image: ubuntu:latest
     script:
       - curl -sSL https://install.typing-master.com | bash
       - typing-master practice --mode timed --duration 3 --language python
     artifacts:
       reports:
         junit: results.xml
   ```

2. Add environment variables:
   - Go to Project Settings → CI/CD → Variables
   - Add `TYPING_MASTER_API_KEY` variable

### Jenkins Integration

1. Create `Jenkinsfile`:
   ```groovy
   pipeline {
       agent any
       stages {
           stage('Typing Practice') {
               steps {
                   sh 'curl -sSL https://install.typing-master.com | bash'
                   sh 'typing-master practice --mode accuracy --language java'
               }
           }
       }
       post {
           always {
               junit 'target/typing-results.xml'
           }
       }
   }
   ```

2. Configure credentials:
   - Add `TYPING_MASTER_API_KEY` as a secret text credential

## 4. Testing and Verification

### VS Code Extension Tests

1. **Basic Functionality**:
   ```bash
   # Open a test file
   echo "function test() { return 'hello world'; }" > test.js
   
   # Test practice commands
   # Right-click → Practice Current File
   # Right-click → Practice Selection
   # Command Palette → Show Leaderboard
   ```

2. **GitHub Integration**:
   ```bash
   # Select code and share to GitHub
   # Verify file/PR creation
   # Check repository for shared content
   ```

### Backend API Tests

1. **Integration Endpoints**:
   ```bash
   # Test creating integration
   curl -X POST http://localhost:8080/api/v1/integrations \
     -H "Authorization: Bearer your_token" \
     -H "Content-Type: application/json" \
     -d '{"provider": "github", "config": {"access_token": "test"}}'
   
   # Test GitHub sharing
   curl -X POST http://localhost:8080/api/v1/integrations/github/share \
     -H "Authorization: Bearer your_token" \
     -H "Content-Type: application/json" \
     -d '{"access_token": "github_token", "repo_owner": "user", "repo_name": "repo", "file_path": "test.js", "content": "console.log('test');"}'
   ```

2. **Embedded Sessions**:
   ```bash
   # Create embedded session
   curl -X POST http://localhost:8080/api/v1/embed \
     -H "Content-Type: application/json" \
     -d '{"platform": "vscode", "content": "console.log('test');", "language": "javascript"}'
   
   # Test session access
   curl http://localhost:8080/api/v1/embed/session_id
   ```

### CI/CD Pipeline Tests

1. **GitHub Actions**:
   - Push to repository with workflow file
   - Check Actions tab for workflow run
   - Verify typing practice execution
   - Check for uploaded results

2. **GitLab CI**:
   - Push to repository with `.gitlab-ci.yml`
   - Monitor pipeline execution
   - Verify test results

3. **Jenkins**:
   - Trigger Jenkins pipeline
   - Monitor build logs
   - Check test results

## 5. Troubleshooting Common Issues

### VS Code Extension Issues

**Extension not loading:**
- Check VS Code version compatibility (requires VS Code 1.74+)
- Verify extension compilation: `npm run compile`
- Check VS Code developer console for errors

**API connection issues:**
- Verify `typing-master.apiUrl` configuration
- Check network connectivity to backend
- Verify backend is running and accessible

**GitHub sharing failures:**
- Verify GitHub token permissions
- Check token expiration
- Ensure repository access permissions

### Backend Issues

**Integration creation failures:**
- Check authentication headers
- Verify request body format
- Check database connection

**GitHub API errors:**
- Verify GitHub token validity
- Check GitHub API rate limits
- Ensure proper OAuth configuration

### CI/CD Issues

**Workflow execution failures:**
- Check environment variables/secrets
- Verify Typing Master CLI installation
- Check repository permissions

**Result upload failures:**
- Verify API key configuration
- Check network connectivity
- Verify file permissions

## 6. Advanced Configuration

### Custom Platform Integration

For custom workflow platforms:

1. **Create Platform Handler**:
   ```go
   // apps/api/internal/services/custom_platform.go
   func (s *CustomPlatformService) GenerateConfig(config map[string]interface{}) (*CIConfig, error) {
       // Custom configuration generation
   }
   ```

2. **Add Platform Support**:
   ```go
   // apps/api/internal/services/workflow_embedding.go
   func (s *WorkflowEmbeddingService) GetSupportedPlatforms() []WorkflowPlatform {
       return []WorkflowPlatform{
           // ... existing platforms
           {
               Name:        "custom-platform",
               DisplayName: "Custom Platform",
               Description: "Custom workflow integration",
           },
       }
   }
   ```

3. **Add API Routes**:
   ```go
   // apps/api/internal/api/router.go
   embed.POST("/ci/custom-platform", handlers.GenerateCIConfig(db))
   ```

### Environment Variables

```bash
# Backend Configuration
PORT=8080
JWT_SECRET=<replace-with-a-strong-secret>
ALLOWED_ORIGINS=https://app.typing-master-for-coding.com,http://localhost:3000
DATABASE_URL=postgres://user:password@localhost:5432/typing_master?sslmode=disable
REDIS_URL=redis://localhost:6379

# Note: JWT_SECRET and ALLOWED_ORIGINS are required at startup. The application will not start without them.

# GitHub Integration
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret

# Other Integrations
GITLAB_CLIENT_ID=your_gitlab_client_id
GITLAB_CLIENT_SECRET=your_gitlab_client_secret
```

### Monitoring and Logging

Enable detailed logging for troubleshooting:

```bash
# Backend
LOG_LEVEL=debug
ENABLE_METRICS=true

# VS Code Extension
TYPING_MASTER_DEBUG=true
```

## 7. Security Best Practices

1. **Token Management**:
   - Use short-lived tokens when possible
   - Rotate tokens regularly
   - Store tokens securely (never in code)
   - Refresh tokens are single-use: a successful `POST /api/v1/auth/refresh` invalidates the previously stored refresh token

2. **Access Control**:
   - Use principle of least privilege
   - Regularly review integration permissions
   - Revoke unused integrations
   - OAuth state is bound to a session cookie; ensure cookies are sent and received by the same browser session
   - `POST /api/v1/auth/login` and `POST /api/v1/auth/login/mfa` enforce account lockout after 5 failed attempts within a 15-minute window
   - `PUT /api/v1/sessions/:id` and `POST /api/v1/sessions/:id/finalize` verify the caller owns the session

3. **Network Security**:
   - Use HTTPS for all API calls
   - Validate SSL certificates
   - Rate limiting is configured via `RateLimitMiddleware` and uses Redis when `REDIS_URL` is set (in-memory fallback otherwise)
   - State-changing API requests must include the `X-CSRF-Token` header matching the `csrf_token` cookie

4. **User Authentication**:
   - New accounts must be created with a password meeting complexity rules (8+ characters, uppercase, lowercase, digit, and special character)
   - New registrations receive an `email_verified` flag of `false`; call `POST /api/v1/auth/verify-email` with the verification token to enable login

5. **Data Protection**:
   - Encrypt sensitive data at rest
   - Use secure communication channels
   - Regular security audits

## 8. Support and Resources

### Documentation
- [Full Integration Documentation](./THIRD_PARTY_INTEGRATIONS.md)
- [API Reference](./API_REFERENCE.md)
- [Development Guide](./DEVELOPMENT.md)

### Community
- [GitHub Issues](https://github.com/typing-master/typing-master-for-coding/issues)
- [Discord Community](https://discord.gg/typing-master)
- [Stack Overflow Tag](https://stackoverflow.com/questions/tagged/typing-master)

### Professional Support
- [Enterprise Support](mailto:enterprise@typing-master-for-coding.com)
- [Consulting Services](mailto:consulting@typing-master-for-coding.com)
- [Training Programs](mailto:training@typing-master-for-coding.com)

---

For additional help or questions, please refer to the main documentation or create an issue on GitHub.
