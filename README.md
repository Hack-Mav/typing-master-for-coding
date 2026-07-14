# Typing Master for Coding

A desktop-first, web-enabled application designed to help developers practice typing real code with syntax-aware guidance across multiple programming languages.

## Features

- **Multi-language Support**: Complete syntax tutorials for C++, Rust, Python, JavaScript, and YAML
- **Syntax-aware Guidance**: Real-time validation using Tree-sitter grammars
- **Multiple Practice Modes**: Tutorials, Timed Drills, Accuracy Mode, Zen Mode, Custom Snippets, and Assessments
- **Comprehensive Metrics**: CPM, Token WPM, accuracy metrics, and detailed performance analysis
- **Offline-first PWA**: Works offline with automatic sync when online
- **Accessibility Compliant**: WCAG 2.2 AA compliance with multiple keyboard layouts

## Quick Start

### Prerequisites

- Node.js 18+
- Go 1.25+
- Docker and Docker Compose
- Git

### Development Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd typing-master-for-coding
   ```

2. **Start the development environment**
   ```bash
   # Copy environment variables
   cp .env.example .env
   
   # Install all dependencies
   make install
   
   # Start databases
   make docker-up
   ```

3. **Run the application**
   ```bash
   # Start all services
   make dev
   
   # Or start individually:
   # Terminal 1: Start backend
   make dev-api
   
   # Terminal 2: Start frontend  
   make dev-web
   ```

4. **Access the application**
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - Health Check: http://localhost:8080/health

### Using Docker Compose

```bash
# Start all services
make docker-up

# View logs
docker-compose logs -f

# Stop all services
make docker-down
```

## Project Structure

```
typing-master-for-coding/
├── README.md                           # Main project README
├── LICENSE                             # License file
├── Makefile                            # Common commands and scripts
├── docker-compose.yml                  # Development environment
├── docker-compose.prod.yml             # Production environment
│
├── docs/                               # All documentation
│   ├── api/                            # API documentation
│   ├── deployment/                     # Deployment guides
│   ├── development/                    # Development guides
│   └── architecture/                   # Architecture docs
│
├── packages/                           # Shared resources
│   ├── types/                          # Shared TypeScript types
│   ├── utils/                          # Shared utilities
│   └── schemas/                        # Shared validation schemas
│
├── apps/                               # Main applications
│   ├── web/                            # React frontend
│   │   ├── src/
│   │   │   ├── components/             # Reusable UI components (practice modes, MonacoTypingInterface)
│   │   │   ├── services/               # API and business logic services (AuthService, SessionManager, ParserManager)
│   │   │   ├── types/                  # TypeScript type definitions
│   │   │   ├── utils/                  # Utility functions and helpers
│   │   │   ├── workers/                # Web Workers for background processing
│   │   │   └── routes/                 # Route configuration and lazy loading
│   │   └── package.json
│   │
│   └── api/                            # Go backend
│       ├── cmd/                        # Application entry point
│       ├── internal/                   # Private application code
│       │   ├── handlers/               # HTTP request handlers (auth, content, scoring, analytics)
│       │   ├── services/               # Business logic layer
│       │   ├── models/                 # Data models and entities
│       │   ├── database/               # Database operations (PostgreSQL integration)
│       │   ├── middleware/             # HTTP middleware (auth, rate limiting, CSRF)
│       │   ├── security/               # Security utilities (CSRF, vulnerability scanning)
│       │   ├── scoring/                # Metrics calculation and anti-cheat detection
│       │   ├── auth/                   # Authentication and authorization logic
│       │   └── cache/                  # Caching layer (Redis with in-memory fallback)
│       └── go.mod
│
├── infrastructure/                     # Infrastructure and deployment
│   ├── docker/                         # Docker configurations
│   ├── kubernetes/                     # K8s manifests (if needed)
│   ├── terraform/                      # Infrastructure as code
│   └── ci-cd/                         # CI/CD pipeline configurations
│
├── tools/                              # Development tools and scripts
├── tests/                              # Cross-application tests
│   ├── e2e/                            # End-to-end tests
│   ├── integration/                    # Integration tests
│   └── load/                           # Load testing
│       └── performance/                # Performance testing
```

## Development

### Frontend Development

```bash
cd apps/web

# Run tests
npm test

# Run linting
npm run lint

# Format code
npm run format

# Type checking
npm run type-check

# Build for production
npm run build
```

### Backend Development

```bash
cd apps/api

# Run tests
go test ./...

# Format code
go fmt ./...

# Run with hot reload (install air first: go install github.com/cosmtrek/air@latest)
air

# Build for production
go build -o bin/server cmd/server/main.go
```

### Using Make Commands

```bash
# Install all dependencies
make install

# Start all development services
make dev

# Run all tests
make test

# Run linting for all applications
make lint

# Format all code
make format

# Build all applications
make build

# Clean build artifacts
make clean
```

### Database Management

```bash
# Connect to PostgreSQL
docker-compose exec postgres psql -U typing_master_user -d typing_master

# Connect to Redis
docker-compose exec redis redis-cli

# Reset database
docker-compose down -v
docker-compose up -d postgres redis
```

## Testing

### Running Tests

```bash
# Using Make commands (recommended)
make test              # Run all tests
make test-web          # Run frontend tests only
make test-api          # Run backend tests only
make test-e2e          # Run end-to-end tests

# Or run individually:

# Frontend tests
cd apps/web
npm test -- --coverage

# Backend tests
cd apps/api
go test -v -race -coverprofile=coverage.out ./...

# End-to-end tests
cd tests
npm test
```

### CI/CD

The project uses GitHub Actions for continuous integration and deployment:

- **CI Pipeline**: Runs on every push and PR
  - Frontend: ESLint, Prettier, tests, build
  - Backend: go vet, go fmt, tests
  - Security: Trivy vulnerability scanning

- **Deployment**: Automatic deployment on release
  - Builds Docker images
  - Pushes to GitHub Container Registry
  - Deploys to production environment

## Current Status

- Frontend `apps/web`: 13 of 14 suites passing, 131 of 145 tests passing. The remaining failure is `CppRustParserIntegration.test.ts` (WASM dynamic import issue in Jest).
- Backend `apps/api`: `go test ./...` passes.
- Authentication uses HttpOnly, Secure cookies; MFA/TOTP (RFC 6238 compliant) and backup codes are implemented.
- Scoring helper functions are implemented; Docker/CI paths use `apps/api` and `apps/web`.
- Cache-backed rate limiting, double-submit CSRF protection, and cache-backed OAuth state are implemented.
- Password complexity validation, email verification, account lockout, refresh token rotation/reuse detection, and session ownership verification are implemented.
- Anonymous session support with authentication choice flow implemented.
- PostgreSQL (pgx + sqlx) and Redis (go-redis) integration complete with in-memory fallback.

## Environment Variables

### Backend (.env)
```bash
ENVIRONMENT=development
DATABASE_URL=postgres://user:password@localhost:5432/typing_master?sslmode=disable
REDIS_URL=redis://localhost:6379
JWT_SECRET=<replace-with-a-strong-secret>  # required - application will not start without this
ALLOWED_ORIGINS=http://localhost:3000  # required - application will not start without this
PORT=8080
```

### Frontend
```bash
REACT_APP_API_URL=http://localhost:8080/api/v1
REACT_APP_ENVIRONMENT=development
```

## Production Deployment

1. **Set up production environment variables**
2. **Configure SSL certificates in nginx/ssl/**
3. **Deploy using Docker Compose**
   ```bash
   docker-compose -f docker-compose.prod.yml up -d
   ```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linting
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.