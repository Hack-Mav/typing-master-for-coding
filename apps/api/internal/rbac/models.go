package rbac

import (
	"time"
)

// Role entity for Datastore - defines roles in the system
type Role struct {
	ID          string    `datastore:"-" json:"id"`
	Name        string    `datastore:"name" json:"name"`                 // "admin", "moderator", "content_creator", "user"
	DisplayName string    `datastore:"display_name" json:"display_name"` // "Administrator", "Moderator", etc.
	Description string    `datastore:"description" json:"description"`
	Permissions []string  `datastore:"permissions" json:"permissions"` // List of permission IDs
	IsActive    bool      `datastore:"is_active" json:"is_active"`
	CreatedBy   string    `datastore:"created_by" json:"created_by"`
	CreatedAt   time.Time `datastore:"created_at" json:"created_at"`
	UpdatedAt   time.Time `datastore:"updated_at" json:"updated_at"`
}

// Permission entity for Datastore - defines granular permissions
type Permission struct {
	ID          string    `datastore:"-" json:"id"`
	Name        string    `datastore:"name" json:"name"` // "create_content", "manage_users", etc.
	DisplayName string    `datastore:"display_name" json:"display_name"`
	Description string    `datastore:"description" json:"description"`
	Resource    string    `datastore:"resource" json:"resource"` // "content", "users", "sessions", etc.
	Action      string    `datastore:"action" json:"action"`     // "create", "read", "update", "delete"
	Scope       string    `datastore:"scope" json:"scope"`       // "own", "all", "team"
	IsActive    bool      `datastore:"is_active" json:"is_active"`
	CreatedBy   string    `datastore:"created_by" json:"created_by"`
	CreatedAt   time.Time `datastore:"created_at" json:"created_at"`
	UpdatedAt   time.Time `datastore:"updated_at" json:"updated_at"`
}

// UserRole entity for Datastore - assigns roles to users
type UserRole struct {
	ID         string     `datastore:"-" json:"id"`
	UserID     string     `datastore:"user_id" json:"user_id"`
	RoleID     string     `datastore:"role_id" json:"role_id"`
	AssignedBy string     `datastore:"assigned_by" json:"assigned_by"`
	AssignedAt time.Time  `datastore:"assigned_at" json:"assigned_at"`
	IsActive   bool       `datastore:"is_active" json:"is_active"`
	ExpiresAt  *time.Time `datastore:"expires_at" json:"expires_at,omitempty"`
	Notes      string     `datastore:"notes" json:"notes"`
	CreatedAt  time.Time  `datastore:"created_at" json:"created_at"`
	UpdatedAt  time.Time  `datastore:"updated_at" json:"updated_at"`
}

// Request/Response DTOs for RBAC operations

// CreateRoleRequest represents a request to create a new role
type CreateRoleRequest struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// UpdateRoleRequest represents a request to update an existing role
type UpdateRoleRequest struct {
	DisplayName *string   `json:"display_name,omitempty"`
	Description *string   `json:"description,omitempty"`
	Permissions *[]string `json:"permissions,omitempty"`
	IsActive    *bool     `json:"is_active,omitempty"`
}

// RoleResponse represents a role response
type RoleResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	Permissions []string  `json:"permissions"`
	IsActive    bool      `json:"is_active"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreatePermissionRequest represents a request to create a new permission
type CreatePermissionRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Scope       string `json:"scope"`
}

// UpdatePermissionRequest represents a request to update an existing permission
type UpdatePermissionRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Description *string `json:"description,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// PermissionResponse represents a permission response
type PermissionResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Scope       string    `json:"scope"`
	IsActive    bool      `json:"is_active"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AssignRoleRequest represents a request to assign a role to a user
type AssignRoleRequest struct {
	UserID    string     `json:"user_id"`
	RoleID    string     `json:"role_id"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Notes     string     `json:"notes,omitempty"`
}

// UserRoleResponse represents a user role assignment response
type UserRoleResponse struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	RoleID     string     `json:"role_id"`
	AssignedBy string     `json:"assigned_by"`
	AssignedAt time.Time  `json:"assigned_at"`
	IsActive   bool       `json:"is_active"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	Notes      string     `json:"notes"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// UserPermissionsResponse represents a user's permissions response
type UserPermissionsResponse struct {
	UserID      string   `json:"user_id"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	IsActive    bool     `json:"is_active"`
}
