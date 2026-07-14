# Deployment Documentation

This section contains all deployment-related documentation for the Typing Master for Coding project.

## Available Documentation

### [Deployment Checklist](./DEPLOYMENT_CHECKLIST.md)
Comprehensive checklist for deploying the application to production environments.

### [Third Party Integrations](./THIRD_PARTY_INTEGRATIONS.md)
Guide for integrating with third-party services and platforms.

## Quick Deployment

For a quick deployment setup, use the provided Makefile commands:

```bash
# Start development environment
make docker-up

# Deploy to production
docker-compose -f docker-compose.prod.yml up -d
```

## Environment Configuration

### Required Environment Variables
- `JWT_SECRET` - Required for JWT token signing (application will not start without this)
- `ALLOWED_ORIGINS` - Required for CORS configuration (application will not start without this)
- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string (optional, in-memory fallback available)

### Optional Environment Variables
- `ENVIRONMENT` - Development or production mode
- `PORT` - Backend server port (default: 8080)

See the main [README.md](../../README.md#environment-variables) for complete environment variable documentation.

## Infrastructure

Infrastructure configurations are located in the `infrastructure/` directory:
- `infrastructure/docker/` - Docker configurations (nginx, postgres)
- `infrastructure/ci-cd/` - CI/CD pipeline configurations (GitHub Actions)
- `infrastructure/kubernetes/` - Kubernetes manifests (currently empty)
- `infrastructure/terraform/` - Infrastructure as code (currently empty)

## Docker Configuration

### Development
- `docker-compose.yml` - Development environment with PostgreSQL and Redis
- Uses `apps/api` and `apps/web` build contexts
- Environment variables with `${VAR:?}` syntax for required values

### Production
- `docker-compose.prod.yml` - Production environment
- Includes nginx reverse proxy
- SSL certificate configuration in `infrastructure/docker/nginx/ssl/`

## CI/CD Pipeline

GitHub Actions workflow is configured in `.github/workflows/ci.yml`:
- Runs on every push and pull request
- Frontend: ESLint, Prettier, tests, build
- Backend: go vet, go fmt, tests
- Security: Trivy vulnerability scanning
- Uses correct paths: `apps/web` and `apps/api`

## Production Considerations

- Ensure all environment variables are properly configured
- Set up SSL certificates in nginx/ssl/
- Configure monitoring and logging
- Review security settings
- Test deployment in staging environment first
- Ensure PostgreSQL and Redis are properly configured
- Verify JWT_SECRET and ALLOWED_ORIGINS are set in production
