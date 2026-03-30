package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/typing-master-for-coding-backend/internal/database"
	"github.com/typing-master-for-coding-backend/internal/rbac"

	"github.com/gin-gonic/gin"
)

// RequirePermission middleware checks if user has a specific permission
func RequirePermission(db *database.DatastoreClient, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Anonymous users cannot have permissions
		if strings.HasPrefix(userID.(string), "anon_") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anonymous users do not have permission: " + permission})
			c.Abort()
			return
		}

		// Check if user has the required permission
		rbacService := rbac.NewService(db)
		ctx := context.Background()
		hasPermission, err := rbacService.HasPermission(ctx, userID.(string), permission)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied: " + permission})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission middleware checks if user has any of the specified permissions
func RequireAnyPermission(db *database.DatastoreClient, permissions []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Anonymous users cannot have permissions
		if strings.HasPrefix(userID.(string), "anon_") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anonymous users do not have any of the required permissions"})
			c.Abort()
			return
		}

		// Check if user has any of the required permissions
		rbacService := rbac.NewService(db)
		ctx := context.Background()
		hasPermission, err := rbacService.HasAnyPermission(ctx, userID.(string), permissions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied: requires one of " + strings.Join(permissions, ", ")})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole middleware checks if user has a specific role
func RequireRole(db *database.DatastoreClient, roleName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Anonymous users cannot have roles
		if strings.HasPrefix(userID.(string), "anon_") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anonymous users do not have role: " + roleName})
			c.Abort()
			return
		}

		// Check if user has the required role
		rbacService := rbac.NewService(db)
		ctx := context.Background()
		permissions, err := rbacService.GetUserPermissions(ctx, userID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user roles"})
			c.Abort()
			return
		}

		hasRole := false
		for _, role := range permissions.Roles {
			if role == roleName {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Role required: " + roleName})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ContentCreatorAuthMiddleware checks if user has content creation permissions
func ContentCreatorAuthMiddleware(db *database.DatastoreClient) gin.HandlerFunc {
	return RequireAnyPermission(db, []string{"create_content", "update_content", "delete_content"})
}

// UserManagementAuthMiddleware checks if user can manage other users
func UserManagementAuthMiddleware(db *database.DatastoreClient) gin.HandlerFunc {
	return RequirePermission(db, "manage_users")
}

// AnalyticsAuthMiddleware checks if user can access analytics
func AnalyticsAuthMiddleware(db *database.DatastoreClient) gin.HandlerFunc {
	return RequirePermission(db, "view_analytics")
}

// TournamentManagementAuthMiddleware checks if user can manage tournaments
func TournamentManagementAuthMiddleware(db *database.DatastoreClient) gin.HandlerFunc {
	return RequirePermission(db, "manage_tournaments")
}

// AdminAuthMiddleware checks if user has admin privileges (legacy function for backward compatibility)
func AdminAuthMiddleware(db *database.DatastoreClient) gin.HandlerFunc {
	return RequireRole(db, "admin")
}
