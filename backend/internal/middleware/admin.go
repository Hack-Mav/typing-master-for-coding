package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminAuthMiddleware checks if user has admin privileges
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user info from JWT claims (set by auth middleware)
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
			c.Abort()
			return
		}

		// Check if user is admin or moderator
		userRole := role.(string)
		if userRole != "admin" && userRole != "moderator" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ModeratorAuthMiddleware checks if user has moderator or admin privileges
func ModeratorAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user info from JWT claims (set by auth middleware)
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
			c.Abort()
			return
		}

		// Check if user is moderator or admin
		userRole := role.(string)
		if userRole != "admin" && userRole != "moderator" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Moderator privileges required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
