package rbac

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/typing-master-for-coding-backend/internal/database"
)

// Service handles role-based access control operations
type Service struct {
	db database.DatastoreInterface
}

// NewService creates a new RBAC service
func NewService(db database.DatastoreInterface) *Service {
	return &Service{db: db}
}

// InitializeDefaultRolesAndPermissions creates default roles and permissions
func (s *Service) InitializeDefaultRolesAndPermissions(ctx context.Context) error {
	// Create default permissions
	permissions := []Permission{
		{
			Name:        "create_content",
			DisplayName: "Create Content",
			Description: "Create new lessons, snippets, and other content",
			Resource:    "content",
			Action:      "create",
			Scope:       "all",
			IsActive:    true,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "update_content",
			DisplayName: "Update Content",
			Description: "Edit existing lessons, snippets, and other content",
			Resource:    "content",
			Action:      "update",
			Scope:       "all",
			IsActive:    true,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "delete_content",
			DisplayName: "Delete Content",
			Description: "Delete lessons, snippets, and other content",
			Resource:    "content",
			Action:      "delete",
			Scope:       "all",
			IsActive:    true,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "manage_users",
			DisplayName: "Manage Users",
			Description: "Create, update, and manage user accounts",
			Resource:    "users",
			Action:      "update",
			Scope:       "all",
			IsActive:    true,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "view_users",
			DisplayName: "View Users",
			Description: "View user profiles and information",
			Resource:    "users",
			Action:      "read",
			Scope:       "all",
			IsActive:    true,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "manage_sessions",
			DisplayName: "Manage Sessions",
			Description: "View and manage user typing sessions",
			Resource:    "sessions",
			Action:      "read",
			Scope:       "all",
			IsActive:    true,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "manage_tournaments",
			DisplayName: "Manage Tournaments",
			Description: "Create and manage typing tournaments",
			Resource:    "tournaments",
			Action:      "update",
			Scope:       "all",
			IsActive:    true,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "view_analytics",
			DisplayName: "View Analytics",
			Description: "Access system analytics and reports",
			Resource:    "analytics",
			Action:      "read",
			Scope:       "all",
			IsActive:    true,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
	}

	// Save permissions
	for _, perm := range permissions {
		key := database.NameKey("Permission", perm.Name, nil)
		_, err := s.db.Put(ctx, key, &perm)
		if err != nil {
			return fmt.Errorf("failed to create permission %s: %w", perm.Name, err)
		}
	}

	// Create default roles
	roles := []Role{
		{
			Name:        "user",
			DisplayName: "User",
			Description: "Basic user with standard access",
			Permissions: []string{}, // Basic users have no special permissions
			IsActive:    true,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "content_creator",
			DisplayName: "Content Creator",
			Description: "Can create and manage content",
			Permissions: []string{"create_content", "update_content"},
			IsActive:    true,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "moderator",
			DisplayName: "Moderator",
			Description: "Can moderate content and users",
			Permissions: []string{"create_content", "update_content", "delete_content", "view_users", "manage_sessions", "view_analytics"},
			IsActive:    true,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "admin",
			DisplayName: "Administrator",
			Description: "Full system access",
			Permissions: []string{"create_content", "update_content", "delete_content", "manage_users", "view_users", "manage_sessions", "manage_tournaments", "view_analytics"},
			IsActive:    true,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
	}

	// Save roles
	for _, role := range roles {
		key := database.NameKey("Role", role.Name, nil)
		_, err := s.db.Put(ctx, key, &role)
		if err != nil {
			return fmt.Errorf("failed to create role %s: %w", role.Name, err)
		}
	}

	return nil
}

// CreateRole creates a new role
func (s *Service) CreateRole(ctx context.Context, req CreateRoleRequest, createdBy string) (*Role, error) {
	// Check if role already exists
	existingKey := database.NameKey("Role", req.Name, nil)
	var existingRole Role
	err := s.db.Get(ctx, existingKey, &existingRole)
	if err == nil {
		return nil, fmt.Errorf("role with name %s already exists", req.Name)
	}

	// Create new role
	role := Role{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Permissions: req.Permissions,
		IsActive:    true,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	key := database.NameKey("Role", role.Name, nil)
	_, err = s.db.Put(ctx, key, &role)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	role.ID = key.Name
	return &role, nil
}

// GetRoles retrieves all roles
func (s *Service) GetRoles(ctx context.Context) ([]Role, error) {
	query := database.NewQuery("Role").Order("name")
	var roles []Role
	keys, err := s.db.GetAll(ctx, query, &roles)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	for i, key := range keys {
		roles[i].ID = key.Name
	}

	return roles, nil
}

// GetRole retrieves a specific role by ID
func (s *Service) GetRole(ctx context.Context, roleID string) (*Role, error) {
	key := database.NameKey("Role", roleID, nil)
	var role Role
	err := s.db.Get(ctx, key, &role)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	role.ID = key.Name
	return &role, nil
}

// UpdateRole updates an existing role
func (s *Service) UpdateRole(ctx context.Context, roleID string, req UpdateRoleRequest) error {
	key := database.NameKey("Role", roleID, nil)
	var role Role
	err := s.db.Get(ctx, key, &role)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	// Update fields if provided
	if req.DisplayName != nil {
		role.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		role.Description = *req.Description
	}
	if req.Permissions != nil {
		role.Permissions = *req.Permissions
	}
	if req.IsActive != nil {
		role.IsActive = *req.IsActive
	}

	role.UpdatedAt = time.Now().UTC()

	_, err = s.db.Put(ctx, key, &role)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	return nil
}

// DeleteRole deletes a role (soft delete by setting inactive)
func (s *Service) DeleteRole(ctx context.Context, roleID string) error {
	key := database.NameKey("Role", roleID, nil)
	var role Role
	err := s.db.Get(ctx, key, &role)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	role.IsActive = false
	role.UpdatedAt = time.Now().UTC()

	_, err = s.db.Put(ctx, key, &role)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

// AssignRole assigns a role to a user
func (s *Service) AssignRole(ctx context.Context, req AssignRoleRequest, assignedBy string) (*UserRole, error) {
	// Check if user already has this role active
	query := database.NewQuery("UserRole").
		Filter("user_id =", req.UserID).
		Filter("role_id =", req.RoleID).
		Filter("is_active =", true).
		Limit(1)

	var existingAssignments []UserRole
	_, err := s.db.GetAll(ctx, query, &existingAssignments)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing assignments: %w", err)
	}

	if len(existingAssignments) > 0 {
		return nil, fmt.Errorf("user already has this role assigned")
	}

	// Create user role assignment
	userRole := UserRole{
		UserID:     req.UserID,
		RoleID:     req.RoleID,
		AssignedBy: assignedBy,
		AssignedAt: time.Now().UTC(),
		IsActive:   true,
		ExpiresAt:  req.ExpiresAt,
		Notes:      req.Notes,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	// Generate UUID for the assignment
	userRoleID := uuid.New().String()
	key := database.NameKey("UserRole", userRoleID, nil)
	_, err = s.db.Put(ctx, key, &userRole)
	if err != nil {
		return nil, fmt.Errorf("failed to assign role: %w", err)
	}

	userRole.ID = userRoleID
	return &userRole, nil
}

// GetUserRoles retrieves all active roles for a user
func (s *Service) GetUserRoles(ctx context.Context, userID string) ([]UserRole, error) {
	now := time.Now().UTC()
	query := database.NewQuery("UserRole").
		Filter("user_id =", userID).
		Filter("is_active =", true).
		Filter("expires_at >", now)

	var userRoles []UserRole
	keys, err := s.db.GetAll(ctx, query, &userRoles)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	for i, key := range keys {
		userRoles[i].ID = key.Name
	}

	return userRoles, nil
}

// GetUserPermissions retrieves all permissions for a user based on their roles
func (s *Service) GetUserPermissions(ctx context.Context, userID string) (*UserPermissionsResponse, error) {
	// Get user's active roles
	userRoles, err := s.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	roleNames := make([]string, len(userRoles))
	roleIDs := make([]string, len(userRoles))
	for i, ur := range userRoles {
		roleNames[i] = ur.RoleID
		roleIDs[i] = ur.RoleID
	}

	// Get all permissions from user's roles
	var allPermissions []string
	seen := make(map[string]bool)

	for _, roleID := range roleIDs {
		role, err := s.GetRole(ctx, roleID)
		if err != nil {
			continue // Skip if role not found
		}

		for _, permID := range role.Permissions {
			if !seen[permID] {
				allPermissions = append(allPermissions, permID)
				seen[permID] = true
			}
		}
	}

	// Check if user has basic role (everyone is at least a user)
	hasBasicRole := false
	for _, roleName := range roleNames {
		if roleName == "user" {
			hasBasicRole = true
			break
		}
	}

	if !hasBasicRole {
		roleNames = append(roleNames, "user")
	}

	return &UserPermissionsResponse{
		UserID:      userID,
		Roles:       roleNames,
		Permissions: allPermissions,
		IsActive:    len(userRoles) > 0,
	}, nil
}

// HasPermission checks if a user has a specific permission
func (s *Service) HasPermission(ctx context.Context, userID, permission string) (bool, error) {
	// Anonymous users have no permissions
	if strings.HasPrefix(userID, "anon_") {
		return false, nil
	}

	permissions, err := s.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, perm := range permissions.Permissions {
		if perm == permission {
			return true, nil
		}
	}

	return false, nil
}

// HasAnyPermission checks if a user has any of the specified permissions
func (s *Service) HasAnyPermission(ctx context.Context, userID string, permissions []string) (bool, error) {
	for _, perm := range permissions {
		hasPerm, err := s.HasPermission(ctx, userID, perm)
		if err != nil {
			return false, err
		}
		if hasPerm {
			return true, nil
		}
	}
	return false, nil
}

// RevokeRole revokes a role from a user
func (s *Service) RevokeRole(ctx context.Context, userID, roleID string) error {
	query := database.NewQuery("UserRole").
		Filter("user_id =", userID).
		Filter("role_id =", roleID).
		Filter("is_active =", true).
		Limit(1)

	var userRoles []UserRole
	keys, err := s.db.GetAll(ctx, query, &userRoles)
	if err != nil {
		return fmt.Errorf("failed to find user role: %w", err)
	}

	if len(userRoles) == 0 {
		return fmt.Errorf("user does not have this role")
	}

	// Deactivate the role assignment
	userRole := userRoles[0]
	userRole.IsActive = false
	userRole.UpdatedAt = time.Now().UTC()

	key := keys[0]
	_, err = s.db.Put(ctx, key, &userRole)
	if err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}

	return nil
}

// GetPermissions retrieves all permissions
func (s *Service) GetPermissions(ctx context.Context) ([]Permission, error) {
	query := database.NewQuery("Permission").Order("name")
	var permissions []Permission
	keys, err := s.db.GetAll(ctx, query, &permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	for i, key := range keys {
		permissions[i].ID = key.Name
	}

	return permissions, nil
}

// CreatePermission creates a new permission
func (s *Service) CreatePermission(ctx context.Context, req CreatePermissionRequest, createdBy string) (*Permission, error) {
	// Check if permission already exists
	existingKey := database.NameKey("Permission", req.Name, nil)
	var existingPermission Permission
	err := s.db.Get(ctx, existingKey, &existingPermission)
	if err == nil {
		return nil, fmt.Errorf("permission with name %s already exists", req.Name)
	}

	// Create new permission
	permission := Permission{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Resource:    req.Resource,
		Action:      req.Action,
		Scope:       req.Scope,
		IsActive:    true,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	key := database.NameKey("Permission", permission.Name, nil)
	_, err = s.db.Put(ctx, key, &permission)
	if err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	permission.ID = key.Name
	return &permission, nil
}
