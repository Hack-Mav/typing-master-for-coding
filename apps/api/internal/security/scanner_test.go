package security

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestVulnerabilityScanner_XSSDetection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultScannerConfig()
	logger := &SecurityLogger{} // Mock logger for testing
	scanner := NewVulnerabilityScanner(config, logger)

	tests := []struct {
		name         string
		queryParams  map[string]string
		expectThreat bool
		expectedType string
	}{
		{
			name: "Clean query parameters",
			queryParams: map[string]string{
				"search": "hello world",
				"page":   "1",
			},
			expectThreat: false,
		},
		{
			name: "XSS in query parameter",
			queryParams: map[string]string{
				"search": "<script>alert('xss')</script>",
			},
			expectThreat: true,
			expectedType: "XSS",
		},
		{
			name: "JavaScript URL",
			queryParams: map[string]string{
				"redirect": "javascript:alert('xss')",
			},
			expectThreat: true,
			expectedType: "XSS",
		},
		{
			name: "Event handler injection",
			queryParams: map[string]string{
				"input": "onload=alert('xss')",
			},
			expectThreat: true,
			expectedType: "XSS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with query parameters
			req := httptest.NewRequest("GET", "/test", nil)
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

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

func TestVulnerabilityScanner_SQLInjectionDetection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultScannerConfig()
	logger := &SecurityLogger{} // Mock logger for testing
	scanner := NewVulnerabilityScanner(config, logger)

	tests := []struct {
		name         string
		queryParams  map[string]string
		expectThreat bool
		expectedType string
	}{
		{
			name: "Clean query parameters",
			queryParams: map[string]string{
				"id":   "123",
				"name": "john",
			},
			expectThreat: false,
		},
		{
			name: "Union select injection",
			queryParams: map[string]string{
				"id": "1 UNION SELECT * FROM users",
			},
			expectThreat: true,
			expectedType: "SQL_INJECTION",
		},
		{
			name: "Drop table injection",
			queryParams: map[string]string{
				"query": "'; DROP TABLE users; --",
			},
			expectThreat: true,
			expectedType: "SQL_INJECTION",
		},
		{
			name: "Always true condition",
			queryParams: map[string]string{
				"filter": "' OR '1'='1",
			},
			expectThreat: true,
			expectedType: "SQL_INJECTION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with query parameters
			req := httptest.NewRequest("GET", "/test", nil)
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

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

func TestVulnerabilityScanner_FormDataScanning(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := DefaultScannerConfig()
	logger := &SecurityLogger{} // Mock logger for testing
	scanner := NewVulnerabilityScanner(config, logger)

	tests := []struct {
		name         string
		formData     map[string]string
		expectThreat bool
		expectedType string
	}{
		{
			name: "Clean form data",
			formData: map[string]string{
				"username": "john",
				"email":    "john@example.com",
			},
			expectThreat: false,
		},
		{
			name: "XSS in form data",
			formData: map[string]string{
				"comment": "<script>alert('xss')</script>",
			},
			expectThreat: true,
			expectedType: "XSS",
		},
		{
			name: "SQL injection in form data",
			formData: map[string]string{
				"search": "'; DROP TABLE users; --",
			},
			expectThreat: true,
			expectedType: "SQL_INJECTION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create form data
			formData := url.Values{}
			for key, value := range tt.formData {
				formData.Set(key, value)
			}

			// Create POST request with form data
			req := httptest.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
			name:           "XSS attack",
			url:            "/test?search=<script>alert('xss')</script>",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "SQL injection",
			url:            "/test?id=%27+OR+%271%27%3D%271",
			expectedStatus: http.StatusForbidden,
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

	// Create router with CSRF middleware
	router := gin.New()
	router.Use(CSRFProtectionMiddleware())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	tests := []struct {
		name           string
		method         string
		headers        map[string]string
		expectedStatus int
	}{
		{
			name:           "GET request (should pass)",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST without origin/referer",
			method:         "POST",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:   "POST with origin header",
			method: "POST",
			headers: map[string]string{
				"Origin": "https://example.com",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "POST with referer header",
			method: "POST",
			headers: map[string]string{
				"Referer": "https://example.com/page",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)

			// Add headers
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "Expected status code to match")
		})
	}
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
