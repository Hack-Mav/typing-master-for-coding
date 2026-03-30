# Development Documentation

This section contains all development-related documentation for the Typing Master for Coding project.

## Available Documentation

### Implementation Guides
- [Implementation Complete](./IMPLEMENTATION_COMPLETE.md) - Overall implementation status and summary
- [Implementation Task 6](./IMPLEMENTATION_TASK_6.md) - Specific task implementation details
- [Task 10 Implementation Summary](./TASK_10_IMPLEMENTATION_SUMMARY.md) - Task 10 completion summary

### Testing & Quality
- [Testing Guide](./TESTING_GUIDE.md) - Comprehensive testing strategies and guidelines
- [Traceability Matrix](./TRACEABILITY_MATRIX.md) - Requirements to implementation traceability

### Setup & Integration
- [Integration Setup Guide](./INTEGRATION_SETUP_GUIDE.md) - Setting up development integrations
- [Restructure Plan](./RESTRUCTURE_PLAN.md) - Information about the recent project restructuring

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

- `apps/web/src/features/` - Frontend feature modules
- `apps/api/internal/` - Backend internal modules
- `packages/` - Shared utilities and types

## Contributing

1. Follow the existing code structure
2. Write tests for new features
3. Update documentation as needed
4. Use the provided linting and formatting tools

## Development Tools

- **Frontend**: React, TypeScript, Tailwind CSS
- **Backend**: Go, Gin framework
- **Testing**: Jest, Go testing
- **Linting**: ESLint, go vet
- **Containers**: Docker, Docker Compose
