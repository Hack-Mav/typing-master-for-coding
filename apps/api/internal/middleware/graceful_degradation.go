package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ServiceStatus represents the health status of dependent services
type ServiceStatus struct {
	Database    bool
	Cache       bool
	ExternalAPI bool
	Storage     bool
	Auth        bool
}

// GracefulDegradationMiddleware provides fallback behavior when services fail
func GracefulDegradationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check service health
		serviceStatus := checkServiceHealth(c)

		// Store service status in context
		c.Set("service_status", serviceStatus)

		// Set appropriate headers based on service availability
		if !serviceStatus.Database {
			c.Header("X-Database-Status", "degraded")
		}

		if !serviceStatus.Cache {
			c.Header("X-Cache-Status", "degraded")
		}

		if !serviceStatus.ExternalAPI {
			c.Header("X-External-API-Status", "degraded")
		}

		// Apply service-specific fallbacks
		if !serviceStatus.Database {
			handleDatabaseFailure(c)
		}

		if !serviceStatus.Cache {
			handleCacheFailure(c)
		}

		if !serviceStatus.ExternalAPI {
			handleExternalAPIFailure(c)
		}

		c.Next()
	}
}

// checkServiceHealth performs health checks on dependent services
func checkServiceHealth(c *gin.Context) ServiceStatus {
	status := ServiceStatus{
		Database:    true,
		Cache:       true,
		ExternalAPI: true,
		Storage:     true,
		Auth:        true,
	}

	// Check database connectivity with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	// Try to access database from context if available
	if _, exists := c.Get("db"); exists {
		// Simple health check - in a real implementation, this would ping the database
		select {
		case <-ctx.Done():
			status.Database = false
		default:
			// Database is responsive
		}
	}

	// Check cache availability
	if _, exists := c.Get("cache"); exists {
		select {
		case <-ctx.Done():
			status.Cache = false
		default:
			// Cache is responsive
		}
	}

	// Check external API availability (simulated)
	select {
	case <-ctx.Done():
		status.ExternalAPI = false
	default:
		// External APIs are responsive
	}

	return status
}

// handleDatabaseFailure provides fallback behavior when database fails
func handleDatabaseFailure(c *gin.Context) {
	path := c.Request.URL.Path
	method := c.Request.Method

	// For read-only operations, try to serve from cache or static data
	if method == "GET" {
		switch {
		case path == "/api/v1/health":
			// Health check can still respond
			c.Set("database_fallback", "static")
		case strings.Contains(path, "/api/v1/languages"):
			// Serve static language data
			c.Set("database_fallback", "static_languages")
		case strings.Contains(path, "/api/v1/lessons"):
			// Serve cached lesson data if available
			c.Set("database_fallback", "cached_lessons")
		default:
			// General read operations get degraded response
			c.Set("database_fallback", "degraded_read")
		}
	} else {
		// Write operations are rejected gracefully
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":       "Database temporarily unavailable",
			"message":     "Write operations are currently disabled. Please try again later.",
			"retry_after": 30,
		})
		c.Abort()
		return
	}
}

// handleCacheFailure provides fallback behavior when cache fails
func handleCacheFailure(c *gin.Context) {
	// Set header indicating cache is degraded
	c.Header("X-Cache-Behavior", "bypass")

	// Continue without caching - responses will be slower but functional
	c.Set("cache_fallback", "direct")
}

// handleExternalAPIFailure provides fallback behavior when external APIs fail
func handleExternalAPIFailure(c *gin.Context) {
	path := c.Request.URL.Path

	// Check if this is an external API dependent endpoint
	if strings.Contains(path, "/integrations/") || strings.Contains(path, "/oauth/") {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":       "External services temporarily unavailable",
			"message":     "Third-party integrations are currently experiencing issues. Please try again later.",
			"retry_after": 60,
		})
		c.Abort()
		return
	}

	// For other endpoints, continue with degraded functionality
	c.Set("external_api_fallback", "limited")
}

// CircuitBreakerMiddleware implements circuit breaker pattern for service failures
func CircuitBreakerMiddleware(serviceName string, failureThreshold int, timeout time.Duration) gin.HandlerFunc {
	failureCount := 0
	lastFailureTime := time.Time{}

	return func(c *gin.Context) {
		now := time.Now()

		// Reset circuit breaker if timeout has passed
		if failureCount >= failureThreshold && now.Sub(lastFailureTime) > timeout {
			failureCount = 0
		}

		// Check if circuit is open
		if failureCount >= failureThreshold {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":       fmt.Sprintf("%s service temporarily unavailable", serviceName),
				"message":     "Service is experiencing high error rates. Please try again later.",
				"retry_after": int(timeout.Seconds()),
			})
			c.Abort()
			return
		}

		// Execute request and track failures
		c.Next()

		// Check if request failed
		if c.Writer.Status() >= 500 {
			failureCount++
			lastFailureTime = now
		}
	}
}

// RetryMiddleware implements retry logic for transient failures
func RetryMiddleware(maxRetries int, retryDelay time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only retry for idempotent requests
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" && c.Request.Method != "OPTIONS" {
			c.Next()
			return
		}

		for i := 0; i <= maxRetries; i++ {
			// Create a copy of the response writer to capture status
			writer := &responseWriter{ResponseWriter: c.Writer, status: 200}
			c.Writer = writer

			c.Next()

			// If successful or client error, don't retry
			if writer.status < 500 || writer.status == 0 {
				break
			}

			// If this is the last attempt, don't delay
			if i < maxRetries {
				time.Sleep(retryDelay)
			}
		}
	}
}

// responseWriter is a wrapper to capture response status
type responseWriter struct {
	gin.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// FallbackDataMiddleware provides static fallback data for critical endpoints
func FallbackDataMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if database fallback is needed
		if fallback, exists := c.Get("database_fallback"); exists {
			switch fallback {
			case "static_languages":
				serveStaticLanguages(c)
			case "static":
				serveStaticHealth(c)
			}
		}
	}
}

// serveStaticLanguages provides fallback language data
func serveStaticLanguages(c *gin.Context) {
	staticLanguages := []gin.H{
		{"id": "1", "name": "JavaScript", "extension": "js", "category": "web"},
		{"id": "2", "name": "Python", "extension": "py", "category": "backend"},
		{"id": "3", "name": "TypeScript", "extension": "ts", "category": "web"},
		{"id": "4", "name": "Go", "extension": "go", "category": "backend"},
		{"id": "5", "name": "Rust", "extension": "rs", "category": "systems"},
	}

	c.JSON(http.StatusOK, gin.H{
		"languages": staticLanguages,
		"source":    "fallback",
		"message":   "Serving from static data due to database unavailability",
	})
}

// serveStaticHealth provides fallback health check
func serveStaticHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "degraded",
		"database":  "unavailable",
		"cache":     "available",
		"timestamp": time.Now().UTC(),
		"message":   "Service running in degraded mode",
	})
}
