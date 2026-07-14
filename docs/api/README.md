# API Documentation

This section contains all API-related documentation for the Typing Master for Coding project.

## Available Documentation

### [Admin CMS Implementation Summary](./ADMIN_CMS_IMPLEMENTATION_SUMMARY.md)
Complete summary of the Content Management System implementation for administrators.

### [Admin CMS Quick Start](./ADMIN_CMS_QUICK_START.md)
Quick start guide for administrators using the CMS features.

### [User Management Summary](./USER_MANAGEMENT_SUMMARY.md)
Overview of user management features and implementation details.

## API Endpoints

For detailed API endpoint documentation, please refer to the main [Technical Specification](../architecture/Typing%20Master%20for%20Coding%20%E2%80%94%20Technical%20Specification.md).

## Authentication

The API uses JWT-based authentication with HttpOnly, Secure cookies for enhanced security. Tokens are automatically managed by the browser and are not accessible to JavaScript, preventing XSS attacks. Details can be found in the [Admin CMS Quick Start](./ADMIN_CMS_QUICK_START.md) guide.

### Authentication Flow
- **Anonymous Sessions**: Users can start sessions without authentication via `POST /api/v1/sessions` with anonymous mode
- **Authentication Choice**: Frontend presents choice to continue anonymously or sign in/register
- **Cookie-based Auth**: All authenticated requests use HttpOnly, Secure cookies with `credentials: 'include'`
- **MFA/TOTP Support**: Two-factor authentication with RFC 6238 compliant TOTP and backup codes

### Security Requirements
- `JWT_SECRET` environment variable is required and must be set before the application starts
- `ALLOWED_ORIGINS` environment variable is required for CORS configuration
- All authentication endpoints set HttpOnly, Secure cookies with appropriate expiration times
- Frontend must use `credentials: 'include'` in fetch requests to send cookies
- State-changing requests (POST, PUT, DELETE, PATCH) must include the `X-CSRF-Token` header matching the `csrf_token` cookie
- Rate limiting is enforced per IP via `RateLimitMiddleware` and uses Redis when `REDIS_URL` is set, with an in-memory fallback
- OAuth state is bound to a session `oauth_state` cookie and validated from the cache
- `POST /api/v1/auth/register` enforces password complexity (8+ chars, mixed case, digit, and special character)
- `POST /api/v1/auth/verify-email` verifies email using a one-time token; `POST /api/v1/auth/login` and `POST /api/v1/auth/login/mfa` require `EmailVerified=true`
- `POST /api/v1/auth/login` and `POST /api/v1/auth/login/mfa` lock the account after 5 failed attempts in a 15-minute window
- `POST /api/v1/auth/refresh` rotates the refresh token and invalidates the previously used one
- `PUT /api/v1/sessions/:id` and `POST /api/v1/sessions/:id/finalize` require the caller to own the session

## Key API Services

### Backend Services (Go)
- **Auth Service**: Authentication, MFA/TOTP, password management, email verification
- **Scoring Service**: Metrics calculation, anti-cheat detection, composite scoring
- **Content Service**: Lessons, snippets, playlists, assessments management
- **Analytics Service**: Performance tracking, insights, progress analysis
- **Assessment Service**: Standardized tests, benchmarking, badge awarding
- **Community Service**: Leaderboards, social features, user profiles
- **RBAC Service**: Role-based access control, permissions management

### Frontend Services (TypeScript)
- **AuthService**: Cookie-based authentication, token management
- **SessionManager**: Session lifecycle, authentication choice flow
- **ParserManager**: Tree-sitter WASM integration for syntax validation
- **MetricsCalculator**: Real-time metrics calculation and analysis
- **ContentService**: Lesson/snippet fetching and caching
- **LeaderboardService**: Leaderboard data and rankings
- **AdminService**: CMS operations for administrators

## Data Models

Information about data models and schemas is available in the [Technical Specification](../architecture/Typing%20Master%20for%20Coding%20%E2%80%94%20Technical%20Specification.md).
