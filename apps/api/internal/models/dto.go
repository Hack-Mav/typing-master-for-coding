package models

import (
	"time"
)

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Handle                string `json:"handle" binding:"required,min=3,max=30"`
	Email                 string `json:"email" binding:"required,email"`
	Password              string `json:"password" binding:"required,min=8"`
	Locale                string `json:"locale"`
	KeyboardLayout        string `json:"keyboard_layout"`
	TelemetryConsent      bool   `json:"telemetry_consent"`
	DataProcessingConsent bool   `json:"data_processing_consent"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	Handle                *string                `json:"handle,omitempty"`
	Locale                *string                `json:"locale,omitempty"`
	KeyboardLayout        *string                `json:"keyboard_layout,omitempty"`
	PrivacyMode           *bool                  `json:"privacy_mode,omitempty"`
	TelemetryConsent      *bool                  `json:"telemetry_consent,omitempty"`
	DataProcessingConsent *bool                  `json:"data_processing_consent,omitempty"`
	Settings              map[string]interface{} `json:"settings,omitempty"`
}

// PrivacySettingsRequest represents privacy settings update
type PrivacySettingsRequest struct {
	PrivacyMode           bool `json:"privacy_mode"`
	TelemetryConsent      bool `json:"telemetry_consent"`
	DataProcessingConsent bool `json:"data_processing_consent"`
}

// DataExportRequest represents a GDPR data export request
type DataExportRequest struct {
	Format string `json:"format" binding:"required,oneof=json csv"`
}

// DataDeletionRequest represents a GDPR data deletion request
type DataDeletionRequest struct {
	Confirm bool `json:"confirm" binding:"required"`
}

// UserResponse represents a sanitized user response
type UserResponse struct {
	ID                    string                 `json:"id"`
	Handle                string                 `json:"handle"`
	Email                 string                 `json:"email"`
	IsAnonymous           bool                   `json:"is_anonymous"`
	Locale                string                 `json:"locale"`
	KeyboardLayout        string                 `json:"keyboard_layout"`
	PrivacyMode           bool                   `json:"privacy_mode"`
	TelemetryConsent      bool                   `json:"telemetry_consent"`
	DataProcessingConsent bool                   `json:"data_processing_consent"`
	Settings              map[string]interface{} `json:"settings"`
	CreatedAt             string                 `json:"created_at"`
	UpdatedAt             string                 `json:"updated_at"`

	// MFA fields
	MFAEnabled bool   `json:"mfa_enabled"`
	MFASetupAt string `json:"mfa_setup_at,omitempty"`
}

// AnonymousSessionRequest represents an anonymous user session creation
type AnonymousSessionRequest struct {
	DeviceID       string `json:"device_id" binding:"required"`
	KeyboardLayout string `json:"keyboard_layout"`
	Locale         string `json:"locale"`
}

// MFA Setup and Management DTOs

// MFASetupRequest represents a request to set up MFA
type MFASetupRequest struct {
	// No additional fields needed for initial setup
}

// MFASetupResponse represents the response for MFA setup
type MFASetupResponse struct {
	Secret      string   `json:"secret"`       // Base32 encoded secret for manual entry
	QRCodeURL   string   `json:"qr_code_url"`  // URL for QR code generation
	BackupCodes []string `json:"backup_codes"` // One-time use backup codes
}

// MFAVerifyRequest represents a request to verify MFA setup
type MFAVerifyRequest struct {
	Code string `json:"code" binding:"required,len=6"` // 6-digit TOTP code
}

// MFAVerifyResponse represents the response for MFA verification
type MFAVerifyResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// MFAStatusResponse represents the current MFA status for a user
type MFAStatusResponse struct {
	Enabled        bool       `json:"enabled"`
	SetupAt        *time.Time `json:"setup_at,omitempty"`
	HasBackupCodes bool       `json:"has_backup_codes"`
}

// MFADisableRequest represents a request to disable MFA
type MFADisableRequest struct {
	Password string `json:"password" binding:"required"`   // Verify password before disabling
	Code     string `json:"code" binding:"required,len=6"` // Current TOTP code for verification
}

// MFARecoveryRequest represents a request to recover account using backup codes
type MFARecoveryRequest struct {
	BackupCode string `json:"backup_code" binding:"required,len=8"` // 8-character backup code
}

// MFARecoveryResponse represents the response for MFA recovery
type MFARecoveryResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// LoginWithMFARequest represents a login request with MFA verification
type LoginWithMFARequest struct {
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required"`
	Code           string `json:"code" binding:"required,len=6"` // 6-digit TOTP code
	RememberDevice bool   `json:"remember_device"`               // Optional: remember device for 30 days
}

// RBAC DTOs for Role-Based Access Control

// CreateRoleRequest represents a request to create a new role
type CreateRoleRequest struct {
	Name        string   `json:"name" binding:"required,min=2,max=50"`
	DisplayName string   `json:"display_name" binding:"required,max=100"`
	Description string   `json:"description" binding:"max=500"`
	Permissions []string `json:"permissions"`
}

// UpdateRoleRequest represents a request to update an existing role
type UpdateRoleRequest struct {
	DisplayName string   `json:"display_name" binding:"max=100"`
	Description string   `json:"description" binding:"max=500"`
	Permissions []string `json:"permissions"`
	IsActive    *bool    `json:"is_active"`
}

// RoleResponse represents a role response
type RoleResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	IsActive    bool     `json:"is_active"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// CreatePermissionRequest represents a request to create a new permission
type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	DisplayName string `json:"display_name" binding:"required,max=100"`
	Description string `json:"description" binding:"max=500"`
	Resource    string `json:"resource" binding:"required"`
	Action      string `json:"action" binding:"required,oneof=create read update delete"`
	Scope       string `json:"scope" binding:"required,oneof=own all team"`
}

// UpdatePermissionRequest represents a request to update an existing permission
type UpdatePermissionRequest struct {
	DisplayName string `json:"display_name" binding:"max=100"`
	Description string `json:"description" binding:"max=500"`
	Resource    string `json:"resource"`
	Action      string `json:"action" binding:"oneof=create read update delete"`
	Scope       string `json:"scope" binding:"oneof=own all team"`
	IsActive    *bool  `json:"is_active"`
}

// PermissionResponse represents a permission response
type PermissionResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Scope       string `json:"scope"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// AssignRoleRequest represents a request to assign a role to a user
type AssignRoleRequest struct {
	UserID    string     `json:"user_id" binding:"required"`
	RoleID    string     `json:"role_id" binding:"required"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Notes     string     `json:"notes" binding:"max=500"`
}

// UserRoleResponse represents a user role assignment response
type UserRoleResponse struct {
	ID              string `json:"id"`
	UserID          string `json:"user_id"`
	RoleID          string `json:"role_id"`
	RoleName        string `json:"role_name"`
	RoleDisplayName string `json:"role_display_name"`
	AssignedBy      string `json:"assigned_by"`
	AssignedAt      string `json:"assigned_at"`
	IsActive        bool   `json:"is_active"`
	ExpiresAt       string `json:"expires_at,omitempty"`
	Notes           string `json:"notes"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// UserPermissionsResponse represents the permissions for a user
type UserPermissionsResponse struct {
	UserID      string   `json:"user_id"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	IsActive    bool     `json:"is_active"`
}
