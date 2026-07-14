package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/typing-master-for-coding-backend/internal/cache"
	"github.com/typing-master-for-coding-backend/internal/database"
	"github.com/typing-master-for-coding-backend/internal/rbac"
	"github.com/typing-master-for-coding-backend/internal/security"

	"github.com/gin-gonic/gin"
)

// hasAuthHeader returns true when the request includes an Authorization header.
func hasAuthHeader(c *gin.Context) bool {
	return c.GetHeader("Authorization") != ""
}

// hasAccessTokenCookie returns true when the access_token cookie is present.
func hasAccessTokenCookie(c *gin.Context) bool {
	_, err := c.Cookie("access_token")
	return err == nil
}

// isPublicAuthEndpoint returns true for routes that are inherently public and
// should not be gated by CSRF protection (e.g. login, register, anonymous).
func isPublicAuthEndpoint(c *gin.Context) bool {
	path := c.Request.URL.Path
	return strings.HasPrefix(path, "/api/v1/auth/") ||
		strings.HasPrefix(path, "/api/v1/sessions/anonymous") ||
		strings.HasPrefix(path, "/api/v1/embed/")
}

// SecurityMiddleware creates a comprehensive security middleware
func SecurityMiddleware(db *database.DatastoreClient) gin.HandlerFunc {
	// Initialize security components
	scannerConfig := security.DefaultScannerConfig()
	loggerConfig := security.DefaultLoggerConfig()

	logger := security.NewSecurityLogger(db, loggerConfig)
	scanner := security.NewVulnerabilityScanner(scannerConfig, logger)
	auditService := security.NewAuditService(db, logger)

	return func(c *gin.Context) {
		// Apply security scanning
		if threat := scanner.ScanRequest(c); threat != nil {
			logger.LogSecurityEvent(context.Background(), *threat)
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Request blocked due to security policy violation",
				"code":  "SECURITY_VIOLATION",
			})
			c.Abort()
			return
		}

		// Apply secure headers and CSRF protection
		security.SetSecureHeaders(c)

		// CSRF is only a concern for authenticated cookie-based sessions on non-public
		// routes. Public auth endpoints (login, register, anonymous, etc.) should not be
		// blocked by CSRF, even if an access_token cookie happens to be present.
		if !isPublicAuthEndpoint(c) && (hasAuthHeader(c) || hasAccessTokenCookie(c)) {
			security.CSRFProtection(c)
		}

		if c.IsAborted() {
			return
		}

		// Log audit event for authenticated requests
		if userID, exists := c.Get("user_id"); exists {
			auditEvent := security.AuditEvent{
				Type:      "api_request",
				Action:    strings.ToLower(c.Request.Method),
				Resource:  c.Request.URL.Path,
				UserID:    userID.(string),
				IP:        c.ClientIP(),
				UserAgent: c.GetHeader("User-Agent"),
				Success:   true, // Will be updated after request processing
			}

			// Store audit event in context for later completion
			c.Set("audit_event", auditEvent)
			c.Set("audit_service", auditService)
		}

		c.Next()

		// Complete audit logging after request processing
		if auditEvent, exists := c.Get("audit_event"); exists {
			if auditSvc, exists := c.Get("audit_service"); exists {
				event := auditEvent.(security.AuditEvent)
				event.Success = c.Writer.Status() < 400
				if c.Writer.Status() >= 400 {
					event.ErrorMsg = "HTTP " + http.StatusText(c.Writer.Status())
				}

				// Log asynchronously to avoid blocking response
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					auditSvc.(*security.AuditService).LogAuditEvent(ctx, event)
				}()
			}
		}
	}
}

// RateLimitMiddleware implements distributed rate limiting using a cache-backed counter.
func RateLimitMiddleware(cacheClient *cache.InMemoryCache, requestsPerMinute int) gin.HandlerFunc {
	window := time.Minute

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", clientIP)

		count := cacheClient.IncrWithTTL(key, window)
		if count > int64(requestsPerMinute) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": int(window.Seconds()),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermissionMiddleware checks if user has required permission
func RequirePermissionMiddleware(db *database.DatastoreClient, permission string) gin.HandlerFunc {
	rbacService := rbac.NewService(db)

	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		hasPermission, err := rbacService.HasPermission(c.Request.Context(), userID.(string), permission)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermissionMiddleware checks if user has any of the required permissions
func RequireAnyPermissionMiddleware(db *database.DatastoreClient, permissions []string) gin.HandlerFunc {
	rbacService := rbac.NewService(db)

	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		hasPermission, err := rbacService.HasAnyPermission(c.Request.Context(), userID.(string), permissions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SecurityAuditMiddleware logs security-relevant events
func SecurityAuditMiddleware(db *database.DatastoreClient) gin.HandlerFunc {
	logger := security.NewSecurityLogger(db, security.DefaultLoggerConfig())
	auditService := security.NewAuditService(db, logger)

	return func(c *gin.Context) {
		// Log security-relevant actions
		securityActions := []string{
			"login", "logout", "register", "password_change",
			"role_assign", "role_revoke", "permission_grant", "permission_revoke",
			"mfa_setup", "mfa_disable", "data_export", "data_delete",
		}

		path := c.Request.URL.Path
		method := c.Request.Method

		isSecurityAction := false
		for _, action := range securityActions {
			if strings.Contains(path, action) {
				isSecurityAction = true
				break
			}
		}

		if isSecurityAction {
			userID := "anonymous"
			if uid, exists := c.Get("user_id"); exists {
				userID = uid.(string)
			}

			auditEvent := security.AuditEvent{
				Type:      "security_action",
				Action:    method + " " + path,
				Resource:  "security",
				UserID:    userID,
				IP:        c.ClientIP(),
				UserAgent: c.GetHeader("User-Agent"),
				Success:   true, // Will be updated after processing
			}

			c.Set("security_audit_event", auditEvent)
			c.Set("security_audit_service", auditService)
		}

		c.Next()

		// Complete security audit logging
		if auditEvent, exists := c.Get("security_audit_event"); exists {
			if auditSvc, exists := c.Get("security_audit_service"); exists {
				event := auditEvent.(security.AuditEvent)
				event.Success = c.Writer.Status() < 400
				if c.Writer.Status() >= 400 {
					event.ErrorMsg = "HTTP " + http.StatusText(c.Writer.Status())
				}

				// Log asynchronously
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					auditSvc.(*security.AuditService).LogAuditEvent(ctx, event)
				}()
			}
		}
	}
}

// IPWhitelistMiddleware restricts access to whitelisted IPs for admin endpoints
func IPWhitelistMiddleware(allowedIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// Check if IP is in whitelist
		allowed := false
		for _, ip := range allowedIPs {
			if ip == clientIP || ip == "*" {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied from this IP address",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequestSizeMiddleware limits request body size
func RequestSizeMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":    "Request body too large",
				"max_size": maxSize,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
