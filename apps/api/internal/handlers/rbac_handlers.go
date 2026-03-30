package handlers

import (
	"context"
	"net/http"

	"github.com/typing-master-for-coding-backend/internal/database"
	"github.com/typing-master-for-coding-backend/internal/models"
	"github.com/typing-master-for-coding-backend/internal/rbac"

	"github.com/gin-gonic/gin"
)

// InitializeRBAC initializes default roles and permissions
func InitializeRBAC(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		rbacService := rbac.NewService(db)

		err := rbacService.InitializeDefaultRolesAndPermissions(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize RBAC: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "RBAC initialized successfully"})
	}
}

// CreateRole creates a new role
func CreateRole(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get user ID from context (set by auth middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		ctx := context.Background()
		rbacService := rbac.NewService(db)

		role, err := rbacService.CreateRole(ctx, rbac.CreateRoleRequest{
			Name:        req.Name,
			Description: req.Description,
			Permissions: req.Permissions,
		}, userID.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, role)
	}
}

// GetRoles retrieves all roles
func GetRoles(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		rbacService := rbac.NewService(db)

		roles, err := rbacService.GetRoles(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get roles"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"roles": roles})
	}
}

// GetRole retrieves a specific role
func GetRole(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID := c.Param("id")
		if roleID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Role ID is required"})
			return
		}

		ctx := context.Background()
		rbacService := rbac.NewService(db)

		role, err := rbacService.GetRole(ctx, roleID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		}

		c.JSON(http.StatusOK, role)
	}
}

// UpdateRole updates an existing role
func UpdateRole(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID := c.Param("id")
		if roleID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Role ID is required"})
			return
		}

		var req models.UpdateRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx := context.Background()
		rbacService := rbac.NewService(db)

		err := rbacService.UpdateRole(ctx, roleID, rbac.UpdateRoleRequest{
			Description: &req.Description,
			Permissions: &req.Permissions,
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Role updated successfully"})
	}
}

// DeleteRole deletes a role (soft delete)
func DeleteRole(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID := c.Param("id")
		if roleID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Role ID is required"})
			return
		}

		ctx := context.Background()
		rbacService := rbac.NewService(db)

		err := rbacService.DeleteRole(ctx, roleID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
	}
}

// CreatePermission creates a new permission
func CreatePermission(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreatePermissionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get user ID from context (set by auth middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		ctx := context.Background()
		rbacService := rbac.NewService(db)

		permission, err := rbacService.CreatePermission(ctx, rbac.CreatePermissionRequest{
			Name:        req.Name,
			Description: req.Description,
			Resource:    req.Resource,
			Action:      req.Action,
		}, userID.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, permission)
	}
}

// GetPermissions retrieves all permissions
func GetPermissions(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		rbacService := rbac.NewService(db)

		permissions, err := rbacService.GetPermissions(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get permissions"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"permissions": permissions})
	}
}

// AssignRole assigns a role to a user
func AssignRole(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.AssignRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get user ID from context (set by auth middleware)
		assignedBy, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		ctx := context.Background()
		rbacService := rbac.NewService(db)

		userRole, err := rbacService.AssignRole(ctx, rbac.AssignRoleRequest{
			UserID: req.UserID,
			RoleID: req.RoleID,
		}, assignedBy.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, userRole)
	}
}

// GetUserRoles retrieves all roles for a user
func GetUserRoles(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
			return
		}

		ctx := context.Background()
		rbacService := rbac.NewService(db)

		userRoles, err := rbacService.GetUserRoles(ctx, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"user_roles": userRoles})
	}
}

// GetUserPermissions retrieves permissions for a user
func GetUserPermissions(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
			return
		}

		ctx := context.Background()
		rbacService := rbac.NewService(db)

		permissions, err := rbacService.GetUserPermissions(ctx, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user permissions"})
			return
		}

		c.JSON(http.StatusOK, permissions)
	}
}

// RevokeRole revokes a role from a user
func RevokeRole(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		roleID := c.Param("role_id")

		if userID == "" || roleID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID and Role ID are required"})
			return
		}

		ctx := context.Background()
		rbacService := rbac.NewService(db)

		err := rbacService.RevokeRole(ctx, userID, roleID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Role revoked successfully"})
	}
}

// CheckPermission checks if current user has a specific permission
func CheckPermission(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		permission := c.Param("permission")
		if permission == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Permission is required"})
			return
		}

		// Get user ID from context (set by auth middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		ctx := context.Background()
		rbacService := rbac.NewService(db)

		hasPermission, err := rbacService.HasPermission(ctx, userID.(string), permission)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permission"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id":    userID,
			"permission": permission,
			"has_access": hasPermission,
		})
	}
}
