# Documentation

Welcome to the Typing Master for Coding documentation hub.

## Documentation Structure

### [📡 API Documentation](./api/)
- **Admin CMS Implementation Summary** - Complete CMS implementation guide
- **Admin CMS Quick Start** - Quick start for administrators
- **User Management Summary** - User management features overview
- **API Overview** - Authentication flow, security requirements, and key services

### [🚀 Deployment Guides](./deployment/)
- **Deployment Checklist** - Production deployment checklist
- **Third Party Integrations** - External service integrations
- **Deployment Overview** - Environment configuration, Docker setup, CI/CD pipeline

### [💻 Development Guides](./development/)
- **Implementation Complete** - Overall implementation status and summary
- **Codebase Critique** - Critical review of current codebase state and recent fixes
- **Restructure Plan** - Information about the recent project restructuring
- **Testing Guide** - Comprehensive testing strategies
- **Traceability Matrix** - Requirements traceability
- **Integration Setup** - Development environment setup
- **Development Overview** - Code organization, tools, and current status

### [🏗️ Architecture Documentation](./architecture/)
- **Technical Specification** - Complete technical requirements and design
- **Architecture Overview** - System architecture, technology stack, data flow, and security considerations

## Quick Links

- [Project README](../README.md) - Main project documentation
- [Technical Specification](./architecture/Typing%20Master%20for%20Coding%20%E2%80%94%20Technical%20Specification.md) - Detailed technical requirements

## Current Status

- **Frontend (`apps/web`)**: 13 of 14 test suites passing, 131 of 145 tests passing. The remaining failing suite is `CppRustParserIntegration.test.ts` (blocked by `web-tree-sitter` WASM dynamic import in Jest). `SessionManager`, `TimedDrillMode`, `ZenMode`, `App`, `MetricsCalculator`, and `ParserManager` property tests have been fixed and pass.
- **Backend (`apps/api`)**: `go test ./...` passes for all packages with tests.
- **Recent changes**: Authentication now uses HttpOnly, Secure cookies; MFA/TOTP (RFC 6238 compliant) is implemented; scoring helper functions are implemented; Docker/CI paths use `apps/api` and `apps/web`; `JWT_SECRET` and `ALLOWED_ORIGINS` are required at startup; cache-backed rate limiting, double-submit CSRF protection, cache-backed OAuth state, password complexity validation, email verification, account lockout, refresh token rotation/reuse detection, and session ownership verification are implemented; PostgreSQL (pgx + sqlx) and Redis (go-redis) integration complete with in-memory fallback; anonymous session support with authentication choice flow implemented.

See [Testing Guide](./development/TESTING_GUIDE.md) and [Codebase Critique](./development/CODEBASE_CRITIQUE.md) for details.

## Getting Started

1. Read the main [README.md](../README.md) for project overview and quick start
2. Check [Development Guides](./development/) for detailed development setup
3. Review [API Documentation](./api/) for backend integration
4. Consult [Deployment Guides](./deployment/) for production setup
5. Review [Architecture Documentation](./architecture/) for system design and technical specifications

## Contributing to Documentation

Documentation is an important part of this project. When contributing:

1. Keep documentation up-to-date with code changes
2. Use clear, concise language
3. Include code examples where helpful
4. Follow the existing structure and formatting
5. Test any instructions or examples provided
