# Production Deployment Checklist

This checklist ensures a smooth deployment of the Typing Master application to production.

## Pre-Deployment

### Code Quality
- [ ] All tests passing (unit, property-based, E2E, accessibility)
- [ ] Code coverage >80%
- [ ] No linting errors
- [ ] TypeScript compilation successful
- [ ] No console errors or warnings
- [ ] Code review completed
- [ ] Security audit passed

### Performance
- [ ] Bundle size optimized (<512KB initial)
- [ ] Keystroke latency <8ms (verified)
- [ ] API response time <200ms p95
- [ ] Lighthouse score >90
- [ ] Web Vitals within targets (LCP <2.5s, FID <100ms, CLS <0.1)

### Dependencies
- [ ] All dependencies up to date
- [ ] No known security vulnerabilities
- [ ] License compliance verified
- [ ] Package-lock.json committed

## Backend Deployment

### Google Cloud Setup
- [ ] Project created: `typing-master-prod`
- [ ] Billing enabled
- [ ] APIs enabled:
  - [ ] App Engine Admin API
  - [ ] Cloud Datastore API
  - [ ] Cloud Storage API
  - [ ] Cloud CDN API
  - [ ] Cloud Monitoring API
  - [ ] Cloud Logging API
  - [ ] Secret Manager API

### Environment Configuration
- [ ] `backend/app.prod.yaml` configured
- [ ] JWT_SECRET set in Secret Manager
- [ ] Environment variables configured
- [ ] ALLOWED_ORIGINS set correctly
- [ ] Database connection tested

### Database Setup
- [ ] Datastore indexes deployed: `gcloud datastore indexes create index.yaml`
- [ ] Indexes built (check status)
- [ ] Test data loaded (if needed)
- [ ] Backup strategy configured

### Backend Deployment Steps
```bash
# 1. Test locally
cd backend
go test ./...

# 2. Build
go build -o typing-master-backend

# 3. Deploy to App Engine
gcloud app deploy app.prod.yaml --project=typing-master-prod

# 4. Deploy cron jobs
gcloud app deploy cron.yaml --project=typing-master-prod

# 5. Deploy dispatch rules
gcloud app deploy dispatch.yaml --project=typing-master-prod

# 6. Verify deployment
curl https://typing-master-prod.appspot.com/health
```

- [ ] Backend deployed successfully
- [ ] Health check passing
- [ ] Cron jobs scheduled
- [ ] Dispatch rules active
- [ ] Logs visible in Cloud Logging

## Frontend Deployment

### Build Configuration
- [ ] Production environment variables set
- [ ] API endpoints configured
- [ ] Analytics configured
- [ ] Error tracking configured (Sentry, etc.)
- [ ] Feature flags configured

### Build and Test
```bash
cd typing-master

# 1. Install dependencies
npm ci

# 2. Run tests
npm test

# 3. Build for production
npm run build

# 4. Verify build
ls -lh build/
npm run analyze
```

- [ ] Build successful
- [ ] Bundle size acceptable
- [ ] Source maps generated
- [ ] Service worker generated
- [ ] Assets optimized

### CDN Setup
```bash
# 1. Run CDN deployment script
cd backend
chmod +x deploy-cdn.sh
./deploy-cdn.sh

# 2. Wait for SSL certificate provisioning (up to 15 minutes)
gcloud compute ssl-certificates describe typing-master-ssl --global

# 3. Update DNS
# Point A records to CDN IP address
```

- [ ] Cloud Storage bucket created
- [ ] Backend bucket configured
- [ ] CDN enabled
- [ ] SSL certificate provisioned
- [ ] DNS updated
- [ ] HTTPS working
- [ ] HTTP redirects to HTTPS

### Frontend Upload
```bash
# Upload to Cloud Storage
gsutil -m cp -r build/* gs://typing-master-static/

# Set cache headers
gsutil -m setmeta -h "Cache-Control:public, max-age=31536000, immutable" \
  gs://typing-master-static/static/**

gsutil -m setmeta -h "Cache-Control:public, max-age=300, must-revalidate" \
  gs://typing-master-static/*.html
```

- [ ] Assets uploaded to Cloud Storage
- [ ] Cache headers set correctly
- [ ] CORS configured
- [ ] Public access enabled

## Monitoring and Observability

### OpenTelemetry Setup
- [ ] Telemetry endpoint configured
- [ ] Metrics collection enabled
- [ ] Traces enabled
- [ ] Sampling configured

### Cloud Monitoring
- [ ] Dashboards created
- [ ] Alert policies configured:
  - [ ] Error rate >1%
  - [ ] API latency p95 >500ms
  - [ ] Keystroke latency >10ms
  - [ ] Instance count at max
  - [ ] Database query time >200ms
  - [ ] CDN cache hit rate <80%
- [ ] Notification channels set up (email, Slack, PagerDuty)
- [ ] Uptime checks configured

### Logging
- [ ] Log aggregation configured
- [ ] Log retention policy set
- [ ] Log-based metrics created
- [ ] Log exports configured (if needed)

### Error Tracking
- [ ] Error reporting enabled
- [ ] Error notifications configured
- [ ] Source maps uploaded
- [ ] Error grouping configured

## Testing in Production

### Smoke Tests
```bash
# Backend health
curl https://typing-master.app/health

# Frontend loading
curl -I https://typing-master.app

# API endpoints
curl https://typing-master.app/api/languages

# CDN caching
curl -I https://typing-master.app/static/js/main.js
```

- [ ] Health endpoint responding
- [ ] Frontend loads correctly
- [ ] API endpoints working
- [ ] CDN serving assets
- [ ] Cache headers correct
- [ ] HTTPS enforced

### E2E Tests on Production
```bash
cd e2e
BASE_URL=https://typing-master.app npm test
```

- [ ] E2E tests passing on production
- [ ] All critical user flows working
- [ ] Performance acceptable

### Load Testing
```bash
cd load-tests
BASE_URL=https://typing-master.app ./run-load-tests.sh
```

- [ ] Load tests passing
- [ ] Performance targets met
- [ ] Auto-scaling working
- [ ] No errors under load

## Security

### SSL/TLS
- [ ] SSL certificate valid
- [ ] HTTPS enforced
- [ ] TLS 1.2+ only
- [ ] HSTS header set
- [ ] Certificate auto-renewal configured

### Headers
- [ ] Content-Security-Policy set
- [ ] X-Content-Type-Options: nosniff
- [ ] X-Frame-Options: DENY
- [ ] X-XSS-Protection: 1; mode=block
- [ ] Referrer-Policy set

### Authentication
- [ ] JWT secret secure
- [ ] Token expiration configured
- [ ] Refresh token rotation enabled
- [ ] Rate limiting enabled
- [ ] CORS configured correctly

### Data Protection
- [ ] Data encryption at rest
- [ ] Data encryption in transit
- [ ] PII handling compliant
- [ ] GDPR compliance verified
- [ ] Data retention policy implemented

## Compliance

### Privacy
- [ ] Privacy policy updated
- [ ] Cookie consent implemented
- [ ] Data export functionality working
- [ ] Data deletion functionality working
- [ ] Anonymization working

### Accessibility
- [ ] WCAG 2.2 AA compliance verified
- [ ] Accessibility tests passing
- [ ] Screen reader tested
- [ ] Keyboard navigation working
- [ ] Color contrast verified

### Legal
- [ ] Terms of service updated
- [ ] License information correct
- [ ] Third-party attributions included
- [ ] DMCA policy in place (if applicable)

## Documentation

- [ ] API documentation updated
- [ ] User documentation updated
- [ ] Admin documentation updated
- [ ] Deployment guide updated
- [ ] Troubleshooting guide updated
- [ ] Changelog updated
- [ ] Release notes prepared

## Communication

### Internal
- [ ] Team notified of deployment
- [ ] Deployment window communicated
- [ ] Rollback plan documented
- [ ] On-call schedule set

### External
- [ ] Status page updated
- [ ] Users notified (if needed)
- [ ] Social media announcement (if applicable)
- [ ] Blog post published (if applicable)

## Post-Deployment

### Verification (First 15 minutes)
- [ ] Monitor error rates
- [ ] Check response times
- [ ] Verify auto-scaling
- [ ] Check logs for errors
- [ ] Test critical user flows
- [ ] Verify metrics collection

### Monitoring (First Hour)
- [ ] Error rate normal
- [ ] Performance metrics normal
- [ ] No unusual patterns in logs
- [ ] User feedback positive
- [ ] No support tickets

### Monitoring (First 24 Hours)
- [ ] Daily active users normal
- [ ] Conversion rates normal
- [ ] No performance degradation
- [ ] No security incidents
- [ ] Database performance normal

### Performance Review (First Week)
- [ ] Review all metrics
- [ ] Analyze user feedback
- [ ] Check for issues
- [ ] Plan optimizations
- [ ] Document lessons learned

## Rollback Plan

### Triggers for Rollback
- Error rate >5%
- Critical functionality broken
- Security vulnerability discovered
- Performance degradation >50%
- Data corruption detected

### Rollback Steps
```bash
# Backend rollback
gcloud app versions list
gcloud app services set-traffic default --splits=PREVIOUS_VERSION=1

# Frontend rollback
gsutil -m cp -r gs://typing-master-static-backup/* gs://typing-master-static/

# Database rollback (if needed)
# Restore from backup
```

- [ ] Rollback procedure tested
- [ ] Backup verified
- [ ] Team trained on rollback
- [ ] Rollback time <15 minutes

## Sign-off

### Technical Lead
- [ ] Code quality approved
- [ ] Performance targets met
- [ ] Security review passed
- [ ] Tests passing

### Product Manager
- [ ] Features complete
- [ ] User acceptance testing passed
- [ ] Documentation complete
- [ ] Release notes approved

### DevOps
- [ ] Infrastructure ready
- [ ] Monitoring configured
- [ ] Backups verified
- [ ] Rollback tested

### Final Approval
- [ ] All checklist items completed
- [ ] Deployment approved
- [ ] Go-live authorized

---

**Deployment Date**: _______________
**Deployed By**: _______________
**Version**: _______________
**Git Commit**: _______________

## Notes

_Add any deployment-specific notes here_
