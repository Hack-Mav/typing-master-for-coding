# Architecture Documentation

This section contains all architecture-related documentation for the Typing Master for Coding project.

## Available Documentation

### [Technical Specification](./Typing%20Master%20for%20Coding%20%E2%80%94%20Technical%20Specification.md)
Comprehensive technical specification including:
- System overview and goals
- Functional requirements
- Architecture patterns
- Data models
- API specifications
- Security and performance considerations

## System Architecture

The Typing Master for Coding follows a clean architecture pattern:

### Frontend (React)
- **Framework**: React 19 with TypeScript
- **State Management**: Zustand
- **Styling**: Tailwind CSS
- **PWA**: Service Workers for offline support
- **Parsing**: Tree-sitter for syntax-aware validation
- **Editor**: Monaco Editor integration
- **Authentication**: Cookie-based with credentials: 'include'
- **Key Components**: AuthenticationChoice, SessionManager, AuthService, ParserManager, MetricsCalculator

### Backend (Go)
- **Framework**: Gin HTTP router
- **Architecture**: Clean architecture with layers
- **Database**: PostgreSQL (pgx + sqlx) with Redis caching
- **Authentication**: JWT-based auth using HttpOnly, Secure cookies; MFA/TOTP support
- **API**: RESTful endpoints
- **Key Services**: Auth, Scoring, Content, Analytics, Assessment, Community, RBAC
- **Security**: Rate limiting, CSRF protection, OAuth state management, account lockout

### Infrastructure
- **Containers**: Docker and Docker Compose
- **Deployment**: Production-ready configurations
- **CI/CD**: GitHub Actions
- **Monitoring**: OpenTelemetry integration (dependencies present, not yet wired)

## Key Design Principles

1. **Clean Architecture**: Clear separation of concerns
2. **Feature-Based**: Modular organization
3. **Offline-First**: PWA capabilities
4. **Syntax-Aware**: Real-time code validation
5. **Performance**: Optimized for typing speed
6. **Accessibility**: WCAG 2.2 AA compliance

## Technology Stack

- **Frontend**: React 19, TypeScript, Tailwind CSS, Monaco Editor, Zustand
- **Backend**: Go 1.25+, Gin, PostgreSQL (pgx + sqlx), Redis (go-redis)
- **Parsing**: Tree-sitter (WASM) for C++, Rust, Python, JavaScript, YAML
- **Testing**: Jest, Go testing, Playwright (E2E), fast-check (property-based)
- **Deployment**: Docker, GitHub Actions

## Data Flow

1. User interacts with React frontend
2. Frontend validates input using Tree-sitter (WASM parsers)
3. Events sent to Go backend via REST API with cookie-based authentication
4. Backend processes and stores in PostgreSQL
5. Results cached in Redis (when `REDIS_URL` is configured) or in-memory for performance
6. Session management via SessionManager with authentication choice flow
7. WebSocket support is not currently implemented

## Security Considerations

- JWT authentication with HttpOnly, Secure cookies (prevents XSS attacks on tokens)
- TOTP-based MFA with backup codes (RFC 6238 compliant)
- Required environment variables: `JWT_SECRET` and `ALLOWED_ORIGINS` (application will not start without these)
- CORS configuration with credential support
- Input validation and request size limits
- Cache-backed rate limiting (`IncrWithTTL`) with Redis when `REDIS_URL` is configured, in-memory fallback
- Double-submit CSRF protection (`csrf_token` cookie + `X-CSRF-Token` header, `SameSite=Strict`)
- Cache-backed OAuth state with TTL, bound to a session `oauth_state` cookie
- Password complexity enforcement in `Register` (8+ chars, mixed case, digit, special)
- Email verification with one-time tokens (`POST /api/v1/auth/verify-email`) and `Login`/`LoginWithMFA` requiring `EmailVerified=true`
- Account lockout and login rate limiting via cache-backed `LoginAttemptTracker` (5 attempts / 15-minute window)
- Refresh token rotation and reuse detection: `RefreshToken` validates against a stored `RefreshTokenHash`, issues a new pair, and invalidates the used token
- Session ownership verification in `UpdateSession` and `FinalizeSession` (`session.UserID` must match the authenticated `user_id`)
- HTTPS enforcement in production
- Anonymous session support with authentication choice flow
