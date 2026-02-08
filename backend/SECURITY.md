# Security Implementation Guide

This document describes the comprehensive security enhancements implemented in the Typing Master for Coding backend.

## Overview

The security implementation includes:
- OWASP-compliant vulnerability scanning
- Real-time security event logging and alerting
- Comprehensive audit trail
- Dependency vulnerability scanning
- Protection against common web vulnerabilities (XSS, CSRF, SQL Injection)
- Role-based access control (RBAC)
- Multi-factor authentication (MFA)

## Security Components

### 1. Vulnerability Scanner (`security/scanner.go`)

Provides real-time protection against:
- **XSS Attacks**: Detects script injection attempts in query parameters and form data
- **SQL Injection**: Identifies malicious SQL patterns in user input
- **CSRF Attacks**: Validates origin and referer headers for state-changing requests
- **Oversized Requests**: Enforces request size limits
- **Malicious Patterns**: Custom pattern matching for known attack vectors

**Configuration:**
```go
config := &ScannerConfig{
    EnableXSSProtection:     true,
    EnableSQLInjectionCheck: true,
    EnableCSRFProtection:    true,
    MaxRequestSize:          10 * 1024 * 1024, // 10MB
    RateLimitRequests:       100,
    RateLimitWindow:         time.Minute,
}
```

### 2. Security Logger (`security/logger.go`)

Comprehensive security event logging with:
- **Real-time Event Logging**: All security events logged to database and files
- **Automated Alerting**: Threshold-based alert generation
- **Metrics Collection**: Security metrics for monitoring and reporting
- **Event Correlation**: Groups related security events

**Event Types:**
- `XSS`: Cross-site scripting attempts
- `SQL_INJECTION`: SQL injection attempts
- `OVERSIZED_REQUEST`: Request size violations
- `MALICIOUS_PATTERN`: Custom pattern matches
- `RATE_LIMIT_EXCEEDED`: Rate limiting violations

### 3. Audit Service (`security/audit.go`)

Maintains comprehensive audit trails:
- **User Actions**: All user actions logged with context
- **System Events**: System-level security events
- **Compliance Reports**: Automated compliance report generation
- **Violation Detection**: Identifies compliance violations

**Audit Event Types:**
- `api_request`: API endpoint access
- `security_action`: Security-related actions (login, MFA, etc.)
- `privilege_change`: Role and permission changes
- `data_access`: Sensitive data access

### 4. Dependency Scanner (`security/dependency.go`)

Automated vulnerability scanning for dependencies:
- **Go Modules**: Scans `go.mod` dependencies using `govulncheck` or OSV API
- **NPM Packages**: Scans `package.json` dependencies using `npm audit`
- **Periodic Scans**: Configurable automated scanning
- **Vulnerability Database**: Integration with OSV (Open Source Vulnerabilities) database

**Supported Tools:**
- `govulncheck`: Official Go vulnerability scanner
- `npm audit`: NPM's built-in security auditing
- OSV API: Fallback vulnerability database

### 5. Security Middleware (`middleware/security.go`)

Comprehensive middleware stack:
- **Request Scanning**: Real-time vulnerability scanning
- **Rate Limiting**: IP-based rate limiting
- **Secure Headers**: OWASP-recommended security headers
- **CSRF Protection**: Cross-site request forgery protection
- **Audit Logging**: Automatic audit trail generation
- **Permission Checking**: RBAC enforcement

## Security Headers

The following security headers are automatically applied:

```
X-XSS-Protection: 1; mode=block
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
Strict-Transport-Security: max-age=31536000; includeSubDomains
Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline'
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: geolocation=(), microphone=(), camera=()
```

## API Endpoints

### Security Management (Admin Only)

- `GET /api/v1/security/events` - Retrieve security events
- `GET /api/v1/security/alerts` - Retrieve security alerts
- `PUT /api/v1/security/alerts/:id/resolve` - Resolve security alert
- `GET /api/v1/security/metrics` - Get security metrics
- `GET /api/v1/security/status` - Get overall security status

### Audit Management (Admin Only)

- `GET /api/v1/audit/events` - Retrieve audit events
- `POST /api/v1/audit/compliance-report` - Generate compliance report

### Dependency Scanning (Admin Only)

- `POST /api/v1/security/scan-dependencies` - Trigger dependency scan

## Configuration

Security settings can be configured via environment variables:

### Scanner Configuration
```bash
SECURITY_ENABLE_XSS_PROTECTION=true
SECURITY_ENABLE_SQL_INJECTION_CHECK=true
SECURITY_ENABLE_CSRF_PROTECTION=true
SECURITY_MAX_REQUEST_SIZE=10485760
SECURITY_RATE_LIMIT_REQUESTS=100
SECURITY_RATE_LIMIT_WINDOW=1m
```

### Logger Configuration
```bash
SECURITY_ENABLE_FILE_LOGGING=true
SECURITY_ENABLE_DATABASE_LOGGING=true
SECURITY_ENABLE_ALERTING=true
SECURITY_ALERT_THRESHOLD=10
SECURITY_ALERT_WINDOW=1h
SECURITY_LOG_LEVEL=INFO
```

### Dependency Scanner Configuration
```bash
SECURITY_ENABLE_GO_MOD_SCAN=true
SECURITY_ENABLE_NPM_SCAN=true
SECURITY_SCAN_INTERVAL=24h
SECURITY_VULNERABILITY_DB_URL=https://osv.dev/v1/query
SECURITY_ALERT_ON_HIGH_SEVERITY=true
```

### General Security Configuration
```bash
SECURITY_ENABLE_SECURITY_HEADERS=true
SECURITY_ENABLE_RATE_LIMITING=true
SECURITY_ENABLE_IP_WHITELIST=false
SECURITY_ALLOWED_ADMIN_IPS=127.0.0.1,::1
SECURITY_SESSION_TIMEOUT=24h
SECURITY_PASSWORD_MIN_LENGTH=8
SECURITY_ACCOUNT_LOCKOUT_THRESHOLD=5
SECURITY_ACCOUNT_LOCKOUT_DURATION=30m
```

## Compliance Features

### OWASP Top 10 Protection

1. **Injection**: SQL injection detection and prevention
2. **Broken Authentication**: MFA, secure session management
3. **Sensitive Data Exposure**: Encryption, secure headers
4. **XML External Entities (XXE)**: Input validation
5. **Broken Access Control**: RBAC implementation
6. **Security Misconfiguration**: Secure defaults, configuration validation
7. **Cross-Site Scripting (XSS)**: XSS detection and prevention
8. **Insecure Deserialization**: Input validation and sanitization
9. **Using Components with Known Vulnerabilities**: Dependency scanning
10. **Insufficient Logging & Monitoring**: Comprehensive audit logging

### Compliance Reports

Generate compliance reports for:
- Security event analysis
- Failed login attempts
- Privilege escalation detection
- Data access patterns
- Violation identification
- Recommendations generation

## Monitoring and Alerting

### Security Metrics

- Total security events
- Blocked events ratio
- Events by type and severity
- Top attacking IPs
- Alert counts and trends

### Alert Conditions

Alerts are automatically generated for:
- Multiple failed login attempts from same IP
- High-severity vulnerability detection
- Unusual privilege escalation patterns
- Excessive security events
- System security score degradation

### Security Score

The system calculates an overall security score (0-100) based on:
- Recent security events
- Active alerts by severity
- Block rate effectiveness
- Compliance violations

## Best Practices

### For Administrators

1. **Regular Monitoring**: Check security dashboard daily
2. **Alert Response**: Investigate and resolve alerts promptly
3. **Dependency Updates**: Keep dependencies updated
4. **Access Review**: Regularly review user roles and permissions
5. **Compliance Reports**: Generate monthly compliance reports

### For Developers

1. **Input Validation**: Always validate and sanitize user input
2. **Error Handling**: Don't expose sensitive information in errors
3. **Logging**: Log security-relevant actions appropriately
4. **Testing**: Include security testing in development workflow
5. **Dependencies**: Keep dependencies updated and scan regularly

## Incident Response

### Security Event Response

1. **Detection**: Automated detection via security scanner
2. **Logging**: Event logged with full context
3. **Alerting**: Automatic alert generation if thresholds exceeded
4. **Investigation**: Admin review of events and context
5. **Response**: Block IPs, update rules, or escalate as needed
6. **Resolution**: Mark alerts as resolved with notes

### Vulnerability Response

1. **Detection**: Automated dependency scanning
2. **Assessment**: Review vulnerability severity and impact
3. **Prioritization**: Address critical and high-severity first
4. **Remediation**: Update dependencies or apply patches
5. **Verification**: Re-scan to confirm resolution
6. **Documentation**: Update security documentation

## Integration

The security system integrates with:
- **Authentication**: JWT token validation and MFA
- **Authorization**: RBAC permission checking
- **Logging**: Structured security and audit logging
- **Monitoring**: Metrics collection and alerting
- **CI/CD**: Automated dependency scanning

## Troubleshooting

### Common Issues

1. **High False Positives**: Adjust scanner sensitivity in configuration
2. **Performance Impact**: Tune rate limiting and scanning thresholds
3. **Alert Fatigue**: Increase alert thresholds or improve filtering
4. **Dependency Scan Failures**: Ensure required tools are installed

### Debug Mode

Enable debug logging:
```bash
SECURITY_LOG_LEVEL=DEBUG
```

### Health Checks

Monitor security system health via:
- Security metrics endpoint
- System status endpoint
- Application logs
- Database audit tables

## Future Enhancements

Planned security improvements:
- Machine learning-based anomaly detection
- Integration with external SIEM systems
- Advanced behavioral analysis
- Automated incident response
- Enhanced compliance reporting
- Real-time threat intelligence feeds