package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/typing-master-for-coding-backend/internal/database"
	"github.com/typing-master-for-coding-backend/internal/rbac"
	"github.com/typing-master-for-coding-backend/internal/security"

	"github.com/gin-gonic/gin"
)

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
		scanMiddleware := scanner.SecurityScanMiddleware()
		scanMiddleware(c)

		// If request was blocked by scanner, don't continue
		if c.IsAborted() {
			return
		}

		// Apply secure headers
		secureHeadersMiddleware := security.SecureHeadersMiddleware()
		secureHeadersMiddleware(c)

		// Apply CSRF protection for state-changing requests
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" && c.Request.Method != "OPTIONS" {
			csrfMiddleware := security.CSRFProtectionMiddleware()
			csrfMiddleware(c)

			if c.IsAborted() {
				return
			}
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

// RateLimitMiddleware implements rate limiting
func RateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
	// Simple in-memory rate limiter (in production, use Redis or similar)
	clients := make(map[string][]time.Time)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		// Clean old entries
		if requests, exists := clients[clientIP]; exists {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if now.Sub(reqTime) < time.Minute {
					validRequests = append(validRequests, reqTime)
				}
			}
			clients[clientIP] = validRequests
		}

		// Check rate limit
		if len(clients[clientIP]) >= requestsPerMinute {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": 60,
			})
			c.Abort()
			return
		}

		// Add current request
		clients[clientIP] = append(clients[clientIP], now)

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
