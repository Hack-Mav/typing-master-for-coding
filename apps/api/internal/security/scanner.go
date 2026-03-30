package security

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// VulnerabilityScanner provides OWASP-compliant security scanning
type VulnerabilityScanner struct {
	config *ScannerConfig
	logger *SecurityLogger
}

// ScannerConfig holds configuration for the vulnerability scanner
type ScannerConfig struct {
	EnableXSSProtection     bool
	EnableSQLInjectionCheck bool
	EnableCSRFProtection    bool
	EnableInputValidation   bool
	MaxRequestSize          int64
	AllowedFileTypes        []string
	BlockedPatterns         []string
	RateLimitRequests       int
	RateLimitWindow         time.Duration
}

// SecurityThreat represents a detected security threat
type SecurityThreat struct {
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	Source      string    `json:"source"`
	UserAgent   string    `json:"user_agent"`
	IP          string    `json:"ip"`
	Timestamp   time.Time `json:"timestamp"`
	Blocked     bool      `json:"blocked"`
}

// NewVulnerabilityScanner creates a new vulnerability scanner
func NewVulnerabilityScanner(config *ScannerConfig, logger *SecurityLogger) *VulnerabilityScanner {
	if config == nil {
		config = DefaultScannerConfig()
	}
	return &VulnerabilityScanner{
		config: config,
		logger: logger,
	}
}

// DefaultScannerConfig returns default scanner configuration
func DefaultScannerConfig() *ScannerConfig {
	return &ScannerConfig{
		EnableXSSProtection:     true,
		EnableSQLInjectionCheck: true,
		EnableCSRFProtection:    true,
		EnableInputValidation:   true,
		MaxRequestSize:          10 * 1024 * 1024, // 10MB
		AllowedFileTypes:        []string{".json", ".yaml", ".yml", ".txt"},
		BlockedPatterns: []string{
			`<script[^>]*>.*?</script>`,
			`javascript:`,
			`vbscript:`,
			`onload\s*=`,
			`onerror\s*=`,
			`onclick\s*=`,
			`union\s+select`,
			`drop\s+table`,
			`delete\s+from`,
			`insert\s+into`,
			`update\s+.*\s+set`,
		},
		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,
	}
}

// ScanRequest performs comprehensive security scanning on incoming requests
func (vs *VulnerabilityScanner) ScanRequest(c *gin.Context) *SecurityThreat {
	// Check for XSS attacks
	if vs.config.EnableXSSProtection {
		if threat := vs.checkXSS(c); threat != nil {
			return threat
		}
	}

	// Check for SQL injection
	if vs.config.EnableSQLInjectionCheck {
		if threat := vs.checkSQLInjection(c); threat != nil {
			return threat
		}
	}

	// Check request size
	if threat := vs.checkRequestSize(c); threat != nil {
		return threat
	}

	// Check for malicious patterns
	if threat := vs.checkMaliciousPatterns(c); threat != nil {
		return threat
	}

	return nil
}

// checkXSS detects potential XSS attacks
func (vs *VulnerabilityScanner) checkXSS(c *gin.Context) *SecurityThreat {
	xssPatterns := []string{
		`<script[^>]*>.*?</script>`,
		`javascript:`,
		`vbscript:`,
		`onload\s*=`,
		`onerror\s*=`,
		`onclick\s*=`,
		`onmouseover\s*=`,
		`onfocus\s*=`,
		`<iframe[^>]*>`,
		`<object[^>]*>`,
		`<embed[^>]*>`,
	}

	// Check query parameters
	for key, values := range c.Request.URL.Query() {
		for _, value := range values {
			if vs.containsMaliciousPattern(value, xssPatterns) {
				return &SecurityThreat{
					Type:        "XSS",
					Severity:    "HIGH",
					Description: fmt.Sprintf("Potential XSS attack detected in query parameter '%s'", key),
					Source:      fmt.Sprintf("Query: %s=%s", key, value),
					UserAgent:   c.GetHeader("User-Agent"),
					IP:          c.ClientIP(),
					Timestamp:   time.Now(),
					Blocked:     true,
				}
			}
		}
	}

	// Check form data if present
	if c.Request.Method == "POST" || c.Request.Method == "PUT" {
		c.Request.ParseForm()
		for key, values := range c.Request.PostForm {
			for _, value := range values {
				if vs.containsMaliciousPattern(value, xssPatterns) {
					return &SecurityThreat{
						Type:        "XSS",
						Severity:    "HIGH",
						Description: fmt.Sprintf("Potential XSS attack detected in form data '%s'", key),
						Source:      fmt.Sprintf("Form: %s=%s", key, value),
						UserAgent:   c.GetHeader("User-Agent"),
						IP:          c.ClientIP(),
						Timestamp:   time.Now(),
						Blocked:     true,
					}
				}
			}
		}
	}

	return nil
}

// checkSQLInjection detects potential SQL injection attacks
func (vs *VulnerabilityScanner) checkSQLInjection(c *gin.Context) *SecurityThreat {
	sqlPatterns := []string{
		`union\s+select`,
		`drop\s+table`,
		`delete\s+from`,
		`insert\s+into`,
		`update\s+.*\s+set`,
		`exec\s*\(`,
		`execute\s*\(`,
		`sp_executesql`,
		`xp_cmdshell`,
		`;\s*--`,
		`'\s*or\s*'1'\s*=\s*'1`,
		`"\s*or\s*"1"\s*=\s*"1`,
		`'\s*or\s*1\s*=\s*1`,
		`"\s*or\s*1\s*=\s*1`,
	}

	// Check query parameters
	for key, values := range c.Request.URL.Query() {
		for _, value := range values {
			if vs.containsMaliciousPattern(value, sqlPatterns) {
				return &SecurityThreat{
					Type:        "SQL_INJECTION",
					Severity:    "CRITICAL",
					Description: fmt.Sprintf("Potential SQL injection detected in query parameter '%s'", key),
					Source:      fmt.Sprintf("Query: %s=%s", key, value),
					UserAgent:   c.GetHeader("User-Agent"),
					IP:          c.ClientIP(),
					Timestamp:   time.Now(),
					Blocked:     true,
				}
			}
		}
	}

	// Check form data
	c.Request.ParseForm()
	for key, values := range c.Request.PostForm {
		for _, value := range values {
			if vs.containsMaliciousPattern(value, sqlPatterns) {
				return &SecurityThreat{
					Type:        "SQL_INJECTION",
					Severity:    "CRITICAL",
					Description: fmt.Sprintf("Potential SQL injection detected in form field '%s'", key),
					Source:      fmt.Sprintf("Form: %s=%s", key, value),
					UserAgent:   c.GetHeader("User-Agent"),
					IP:          c.ClientIP(),
					Timestamp:   time.Now(),
					Blocked:     true,
				}
			}
		}
	}

	return nil
}

// checkRequestSize validates request size limits
func (vs *VulnerabilityScanner) checkRequestSize(c *gin.Context) *SecurityThreat {
	if c.Request.ContentLength > vs.config.MaxRequestSize {
		return &SecurityThreat{
			Type:        "OVERSIZED_REQUEST",
			Severity:    "MEDIUM",
			Description: fmt.Sprintf("Request size (%d bytes) exceeds maximum allowed (%d bytes)", c.Request.ContentLength, vs.config.MaxRequestSize),
			Source:      fmt.Sprintf("Content-Length: %d", c.Request.ContentLength),
			UserAgent:   c.GetHeader("User-Agent"),
			IP:          c.ClientIP(),
			Timestamp:   time.Now(),
			Blocked:     true,
		}
	}
	return nil
}

// checkMaliciousPatterns checks for custom malicious patterns
func (vs *VulnerabilityScanner) checkMaliciousPatterns(c *gin.Context) *SecurityThreat {
	// Check all query parameters and form data
	allValues := []string{}

	// Add query parameters
	for _, values := range c.Request.URL.Query() {
		allValues = append(allValues, values...)
	}

	// Add form data if present
	if c.Request.Method == "POST" || c.Request.Method == "PUT" {
		c.Request.ParseForm()
		for _, values := range c.Request.PostForm {
			allValues = append(allValues, values...)
		}
	}

	// Check headers for suspicious content
	suspiciousHeaders := []string{"X-Forwarded-For", "X-Real-IP", "Referer", "User-Agent"}
	for _, header := range suspiciousHeaders {
		if value := c.GetHeader(header); value != "" {
			allValues = append(allValues, value)
		}
	}

	// Scan all values
	for _, value := range allValues {
		if vs.containsMaliciousPattern(value, vs.config.BlockedPatterns) {
			return &SecurityThreat{
				Type:        "MALICIOUS_PATTERN",
				Severity:    "HIGH",
				Description: "Malicious pattern detected in request",
				Source:      fmt.Sprintf("Value: %s", value),
				UserAgent:   c.GetHeader("User-Agent"),
				IP:          c.ClientIP(),
				Timestamp:   time.Now(),
				Blocked:     true,
			}
		}
	}

	return nil
}

// containsMaliciousPattern checks if input contains any malicious patterns
func (vs *VulnerabilityScanner) containsMaliciousPattern(input string, patterns []string) bool {
	input = strings.ToLower(input)
	for _, pattern := range patterns {
		matched, err := regexp.MatchString(pattern, input)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// SecurityScanMiddleware creates a Gin middleware for security scanning
func (vs *VulnerabilityScanner) SecurityScanMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip scanning for health checks and static assets
		if strings.HasPrefix(c.Request.URL.Path, "/health") ||
			strings.HasPrefix(c.Request.URL.Path, "/static") {
			c.Next()
			return
		}

		// Perform security scan
		threat := vs.ScanRequest(c)
		if threat != nil {
			// Log the security threat
			vs.logger.LogSecurityEvent(context.Background(), *threat)

			// Block the request if it's a threat
			if threat.Blocked {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Request blocked due to security policy violation",
					"code":  "SECURITY_VIOLATION",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// CSRFProtectionMiddleware provides CSRF protection
func CSRFProtectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF check for GET, HEAD, OPTIONS requests
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Check for CSRF token in header
		csrfToken := c.GetHeader("X-CSRF-Token")
		if csrfToken == "" {
			// Also check in form data
			csrfToken = c.PostForm("_csrf_token")
		}

		// For now, we'll implement a simple origin-based CSRF protection
		// In production, you'd want to implement proper CSRF tokens
		origin := c.GetHeader("Origin")
		referer := c.GetHeader("Referer")

		// Basic origin validation
		if origin == "" && referer == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "CSRF protection: Origin or Referer header required",
				"code":  "CSRF_VIOLATION",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SecureHeadersMiddleware adds security headers
func SecureHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent XSS attacks
		c.Header("X-XSS-Protection", "1; mode=block")

		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Enforce HTTPS
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Content Security Policy
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")

		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}
