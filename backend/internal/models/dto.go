package models

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Handle              string `json:"handle" binding:"required,min=3,max=30"`
	Email               string `json:"email" binding:"required,email"`
	Password            string `json:"password" binding:"required,min=8"`
	Locale              string `json:"locale"`
	KeyboardLayout      string `json:"keyboard_layout"`
	TelemetryConsent    bool   `json:"telemetry_consent"`
	DataProcessingConsent bool `json:"data_processing_consent"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	Handle              *string `json:"handle,omitempty"`
	Locale              *string `json:"locale,omitempty"`
	KeyboardLayout      *string `json:"keyboard_layout,omitempty"`
	PrivacyMode         *bool   `json:"privacy_mode,omitempty"`
	TelemetryConsent    *bool   `json:"telemetry_consent,omitempty"`
	DataProcessingConsent *bool `json:"data_processing_consent,omitempty"`
	Settings            map[string]interface{} `json:"settings,omitempty"`
}

// PrivacySettingsRequest represents privacy settings update
type PrivacySettingsRequest struct {
	PrivacyMode         bool `json:"privacy_mode"`
	TelemetryConsent    bool `json:"telemetry_consent"`
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
	ID                  string                 `json:"id"`
	Handle              string                 `json:"handle"`
	Email               string                 `json:"email"`
	IsAnonymous         bool                   `json:"is_anonymous"`
	Locale              string                 `json:"locale"`
	KeyboardLayout      string                 `json:"keyboard_layout"`
	PrivacyMode         bool                   `json:"privacy_mode"`
	TelemetryConsent    bool                   `json:"telemetry_consent"`
	DataProcessingConsent bool                 `json:"data_processing_consent"`
	Settings            map[string]interface{} `json:"settings"`
	CreatedAt           string                 `json:"created_at"`
	UpdatedAt           string                 `json:"updated_at"`
}

// AnonymousSessionRequest represents an anonymous user session creation
type AnonymousSessionRequest struct {
	DeviceID       string `json:"device_id" binding:"required"`
	KeyboardLayout string `json:"keyboard_layout"`
	Locale         string `json:"locale"`
}
