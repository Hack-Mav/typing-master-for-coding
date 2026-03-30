# Project Restructure Plan

## Current Issues
1. Scattered documentation files in root directory
2. Backend deployment configs mixed with source code
3. Frontend lacks proper feature-based organization
4. No shared resources between frontend/backend
5. Deployment and infrastructure files spread across multiple locations

## Proposed New Structure

```
typing-master-for-coding/
├── README.md                           # Main project README
├── LICENSE                             # License file
├── .gitignore                          # Git ignore rules
├── docker-compose.yml                  # Development environment
├── docker-compose.prod.yml             # Production environment
├── Makefile                            # Common commands and scripts
│
├── docs/                               # All documentation
│   ├── README.md                       # Documentation index
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
│   ├── web/                            # React frontend (formerly typing-master)
│   │   ├── public/
│   │   ├── src/
│   │   │   ├── components/             # Reusable UI components
│   │   │   │   ├── common/              # Generic components
│   │   │   │   ├── forms/               # Form components
│   │   │   │   └── layout/              # Layout components
│   │   │   ├── features/                # Feature-based modules
│   │   │   │   ├── auth/                # Authentication
│   │   │   │   ├── typing/              # Core typing functionality
│   │   │   │   ├── lessons/             # Lesson management
│   │   │   │   ├── progress/            # Progress tracking
│   │   │   │   ├── leaderboards/        # Leaderboard functionality
│   │   │   │   └── settings/            # User settings
│   │   │   ├── hooks/                   # Custom React hooks
│   │   │   ├── services/                # API and external services
│   │   │   ├── store/                   # State management (Zustand)
│   │   │   ├── utils/                   # Frontend utilities
│   │   │   ├── types/                   # Frontend-specific types
│   │   │   ├── styles/                  # Global styles and Tailwind
│   │   │   └── workers/                 # Web Workers
│   │   ├── package.json
│   │   ├── tsconfig.json
│   │   ├── tailwind.config.js
│   │   └── craco.config.js
│   │
│   └── api/                             # Go backend (formerly backend)
│       ├── cmd/                         # Application entry points
│       │   └── server/
│       │       └── main.go
│       ├── internal/                    # Private application code
│       │   ├── api/                     # API handlers and routes
│       │   ├── auth/                    # Authentication logic
│       │   ├── cache/                   # Caching layer
│       │   ├── config/                  # Configuration management
│       │   ├── database/                # Database operations
│       │   ├── errors/                  # Error handling
│       │   ├── middleware/              # HTTP middleware
│       │   ├── models/                  # Data models
│       │   ├── repository/              # Data access layer
│       │   ├── scoring/                 # Scoring algorithms
│       │   ├── security/                # Security utilities
│       │   ├── services/                # Business logic
│       │   ├── telemetry/               # Monitoring and logging
│       │   └── utils/                   # Backend utilities
│       ├── pkg/                         # Public library code
│       ├── migrations/                  # Database migrations
│       ├── tests/                       # Integration and e2e tests
│       ├── go.mod
│       ├── go.sum
│       └── Dockerfile
│
├── infrastructure/                      # Infrastructure and deployment
│   ├── docker/                          # Docker configurations
│   │   ├── nginx/
│   │   └── postgres/
│   ├── kubernetes/                      # K8s manifests (if needed)
│   ├── terraform/                       # Infrastructure as code
│   └── ci-cd/                          # CI/CD pipeline configurations
│       ├── .github/
│       └── scripts/
│
├── tools/                               # Development tools and scripts
│   ├── scripts/                         # Utility scripts
│   ├── generators/                      # Code generators
│   └── linters/                         # Linting configurations
│
└── tests/                              # Cross-application tests
    ├── e2e/                            # End-to-end tests
    ├── integration/                    # Integration tests
    └── load/                           # Load testing
```

## Benefits of New Structure

1. **Clear separation of concerns**: Each directory has a specific purpose
2. **Feature-based organization**: Frontend features are self-contained
3. **Shared resources**: Common types and utilities are centralized
4. **Infrastructure separation**: Deployment configs are isolated
5. **Scalable**: Easy to add new applications or features
6. **Maintainable**: Clear navigation and understanding of codebase

## Migration Steps

1. Create new directory structure
2. Move and rename existing files
3. Update import paths and configurations
4. Update documentation
5. Test all functionality
6. Update CI/CD pipelines
