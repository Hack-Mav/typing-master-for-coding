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
- Go 1.21+
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
   
   # Start databases
   docker-compose up -d postgres redis
   
   # Install frontend dependencies
   cd typing-master
   npm install
   
   # Install backend dependencies
   cd ../backend
   go mod download
   ```

3. **Run the application**
   ```bash
   # Terminal 1: Start backend
   cd backend
   go run main.go
   
   # Terminal 2: Start frontend
   cd typing-master
   npm start
   ```

4. **Access the application**
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - Health Check: http://localhost:8080/health

### Using Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down
```

## Project Structure

```
typing-master-for-coding/
├── typing-master/          # React frontend
│   ├── src/
│   ├── public/
│   └── package.json
├── backend/                # Go backend
│   ├── internal/
│   ├── main.go
│   └── go.mod
├── database/               # Database initialization
│   └── init/
├── nginx/                  # Nginx configuration
├── .github/workflows/      # CI/CD pipelines
├── docker-compose.yml      # Development environment
├── docker-compose.prod.yml # Production environment
└── README.md
```

## Development

### Frontend Development

```bash
cd typing-master

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
cd backend

# Run tests
go test ./...

# Format code
go fmt ./...

# Run with hot reload (install air first: go install github.com/cosmtrek/air@latest)
air

# Build for production
go build -o main .
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
# Frontend tests
cd typing-master
npm test -- --coverage

# Backend tests
cd backend
go test -v -race -coverprofile=coverage.out ./...
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

## Environment Variables

### Backend (.env)
```bash
ENVIRONMENT=development
DATABASE_URL=postgres://user:password@localhost:5432/typing_master?sslmode=disable
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-secret-key
PORT=8080
FRONTEND_URL=http://localhost:3000
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