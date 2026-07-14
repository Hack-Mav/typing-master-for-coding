# Development Documentation

This section contains all development-related documentation for the Typing Master for Coding project.

## Available Documentation

### Implementation Guides
- [Implementation Complete](./IMPLEMENTATION_COMPLETE.md) - Overall implementation status and summary
- [Codebase Critique](./CODEBASE_CRITIQUE.md) - Critical review of current codebase state and recent fixes
- [Restructure Plan](./RESTRUCTURE_PLAN.md) - Information about the recent project restructuring

### Testing & Quality
- [Testing Guide](./TESTING_GUIDE.md) - Comprehensive testing strategies and guidelines
- [Traceability Matrix](./TRACEABILITY_MATRIX.md) - Requirements to implementation traceability

### Setup & Integration
- [Integration Setup Guide](./INTEGRATION_SETUP_GUIDE.md) - Setting up development integrations

## Development Setup

For quick development setup, use the Makefile:

```bash
# Install all dependencies
make install

# Start development environment
make dev

# Run tests
make test

# Build applications
make build
```

## Code Organization

The project follows a feature-based architecture:

### Frontend (`apps/web/src/`)
- `components/` - Reusable UI components (AccessibilitySettings, MonacoTypingInterface, practice modes)
- `services/` - API and business logic services (AuthService, SessionManager, ParserManager, MetricsCalculator)
- `types/` - TypeScript type definitions
- `utils/` - Utility functions and helpers
- `workers/` - Web Workers for background processing
- `routes/` - Route configuration and lazy loading

### Backend (`apps/api/internal/`)
- `handlers/` - HTTP request handlers (auth, content, scoring, analytics, assessment, community)
- `services/` - Business logic layer
- `models/` - Data models and entities
- `database/` - Database operations (PostgreSQL integration)
- `middleware/` - HTTP middleware (auth, rate limiting, CSRF, telemetry)
- `security/` - Security utilities (CSRF, vulnerability scanning)
- `scoring/` - Metrics calculation and anti-cheat detection
- `auth/` - Authentication and authorization logic
- `cache/` - Caching layer (Redis with in-memory fallback)

## Contributing

1. Follow the existing code structure
2. Write tests for new features
3. Update documentation as needed
4. Use the provided linting and formatting tools

## Development Tools

- **Frontend**: React 19, TypeScript, Tailwind CSS, Monaco Editor, Zustand
- **Backend**: Go 1.25+, Gin framework
- **Testing**: Jest, Go testing, Playwright (E2E), fast-check (property-based)
- **Linting**: ESLint, Prettier, go vet
- **Containers**: Docker, Docker Compose
- **Databases**: PostgreSQL (pgx + sqlx), Redis (go-redis) with in-memory fallback

## Current Status

### Frontend Tests
- **Test Suites**: 14 suites with 145 tests
- **Passing**: 13 suites, 131 tests
- **Failing**: 1 suite (`CppRustParserIntegration`), 14 tests
- **Issue**: Tree-sitter WASM loading in Jest/jsdom environment (requires `--experimental-vm-modules`)

### Backend Tests
- **Status**: All backend tests pass (`go test ./...`)
- **Coverage**: Auth, handlers, scoring, security packages tested

### Recent Fixes
- Authentication now uses HttpOnly, Secure cookies
- MFA/TOTP implemented with RFC 6238 compliance
- Scoring helper functions fully implemented
- PostgreSQL/Redis persistence integrated
- Docker/CI paths corrected to use `apps/api` and `apps/web`
- Security controls: rate limiting, CSRF, OAuth state, password complexity, email verification, account lockout, refresh token rotation, session ownership verification
