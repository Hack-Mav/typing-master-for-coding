package security

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestVulnerabilityScanner_RequestSizeLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultScannerConfig()
	config.MaxRequestSize = 100 // Set small limit for testing
	logger := &SecurityLogger{} // Mock logger for testing
	scanner := NewVulnerabilityScanner(config, logger)

	tests := []struct {
		name          string
		contentLength int64
		expectThreat  bool
		expectedType  string
	}{
		{
			name:          "Small request",
			contentLength: 50,
			expectThreat:  false,
		},
		{
			name:          "Large request",
			contentLength: 200,
			expectThreat:  true,
			expectedType:  "OVERSIZED_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with specified content length
			req := httptest.NewRequest("POST", "/test", strings.NewReader("test body"))
			req.ContentLength = tt.contentLength

			// Create Gin context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Scan request
			threat := scanner.ScanRequest(c)

			if tt.expectThreat {
				assert.NotNil(t, threat, "Expected threat to be detected")
				assert.Equal(t, tt.expectedType, threat.Type, "Expected threat type to match")
				assert.True(t, threat.Blocked, "Expected threat to be blocked")
			} else {
				assert.Nil(t, threat, "Expected no threat to be detected")
			}
		})
	}
}

func TestSecurityScanMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultScannerConfig()
	logger := NewSecurityLogger(nil, nil) // Mock logger for testing
	scanner := NewVulnerabilityScanner(config, logger)

	// Create router with security middleware
	router := gin.New()
	router.Use(scanner.SecurityScanMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	tests := []struct {
		name           string
		url            string
		expectedStatus int
	}{
		{
			name:           "Clean request",
			url:            "/test?search=hello",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "XSS attack (no longer blocked by regex)",
			url:            "/test?search=<script>alert('xss')</script>",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "SQL injection (no longer blocked by regex)",
			url:            "/test?id=%27+OR+%271%27%3D%271",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Health check (should be allowed)",
			url:            "/health",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "Expected status code to match")
		})
	}
}

func TestCSRFProtectionMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CSRFProtectionMiddleware())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// GET request should set CSRF cookie
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var token string
	for _, c := range w.Result().Cookies() {
		if c.Name == csrfCookieName {
			token = c.Value
			break
		}
	}
	assert.NotEmpty(t, token, "Expected CSRF cookie to be set")

	// POST without CSRF cookie should fail
	req = httptest.NewRequest("POST", "/test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)

	// POST with cookie but no token header should fail
	req = httptest.NewRequest("POST", "/test", nil)
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: token})
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)

	// POST with cookie and matching token header should pass
	req = httptest.NewRequest("POST", "/test", nil)
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: token})
	req.Header.Set(csrfHeaderName, token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// POST with cookie and mismatched token header should fail
	req = httptest.NewRequest("POST", "/test", nil)
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: token})
	req.Header.Set(csrfHeaderName, "invalid-token")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestSecureHeadersMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create router with secure headers middleware
	router := gin.New()
	router.Use(SecureHeadersMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Check that security headers are set
	expectedHeaders := map[string]string{
		"X-XSS-Protection":          "1; mode=block",
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
	}

	for header, expectedValue := range expectedHeaders {
		actualValue := w.Header().Get(header)
		assert.Equal(t, expectedValue, actualValue, "Expected header %s to have value %s", header, expectedValue)
	}

	// Check CSP header exists
	csp := w.Header().Get("Content-Security-Policy")
	assert.Contains(t, csp, "default-src 'self'", "Expected CSP header to contain default-src 'self'")
}
