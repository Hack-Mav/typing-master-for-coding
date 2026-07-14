package security

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
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
	MaxRequestSize int64
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
		MaxRequestSize: 10 * 1024 * 1024, // 10MB
	}
}

// ScanRequest performs security scanning on incoming requests
func (vs *VulnerabilityScanner) ScanRequest(c *gin.Context) *SecurityThreat {
	return vs.checkRequestSize(c)
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

const csrfCookieName = "csrf_token"
const csrfHeaderName = "X-CSRF-Token"

func generateCSRFToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// CSRFProtection validates the CSRF token for state-changing requests.
// It sets a double-submit cookie on safe requests and verifies the token
// matches a header or form value for unsafe requests.
func CSRFProtection(c *gin.Context) {
	if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
		if _, err := c.Cookie(csrfCookieName); err != nil {
			token := generateCSRFToken()
			isSecure := c.Request.TLS != nil || strings.EqualFold(c.Request.Header.Get("X-Forwarded-Proto"), "https")
			http.SetCookie(c.Writer, &http.Cookie{
				Name:     csrfCookieName,
				Value:    token,
				Path:     "/",
				Secure:   isSecure,
				HttpOnly: false,
				SameSite: http.SameSiteStrictMode,
				MaxAge:   86400,
			})
		}
		return
	}

	token, err := c.Cookie(csrfCookieName)
	if err != nil || token == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "CSRF protection: missing CSRF token cookie",
			"code":  "CSRF_VIOLATION",
		})
		c.Abort()
		return
	}

	submitted := c.GetHeader(csrfHeaderName)
	if submitted == "" {
		submitted = c.PostForm("_csrf_token")
	}
	if subtle.ConstantTimeCompare([]byte(submitted), []byte(token)) != 1 {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "CSRF protection: invalid CSRF token",
			"code":  "CSRF_VIOLATION",
		})
		c.Abort()
		return
	}
}

// CSRFProtectionMiddleware returns a Gin middleware that applies CSRFProtection.
func CSRFProtectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		CSRFProtection(c)
		c.Next()
	}
}

// SetSecureHeaders sets OWASP security headers.
func SetSecureHeaders(c *gin.Context) {
	c.Header("X-XSS-Protection", "1; mode=block")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Frame-Options", "DENY")
	c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
	c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
	c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
}

// SecureHeadersMiddleware adds security headers to the response.
func SecureHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		SetSecureHeaders(c)
		c.Next()
	}
}
