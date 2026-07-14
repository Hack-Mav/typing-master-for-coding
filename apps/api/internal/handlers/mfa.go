package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/typing-master-for-coding-backend/internal/auth"
	"github.com/typing-master-for-coding-backend/internal/database"
	"github.com/typing-master-for-coding-backend/internal/models"
)

// MFAService handles MFA-related operations
type MFAService struct {
	db *database.DatastoreClient
}

// NewMFAService creates a new MFA service
func NewMFAService(db *database.DatastoreClient) *MFAService {
	return &MFAService{db: db}
}

// SetupMFA initiates MFA setup for a user
func (s *MFAService) SetupMFA(userID string) (*models.MFASetupResponse, error) {
	ctx := context.Background()

	// Get user
	key := database.NameKey("User", userID, nil)
	var user models.User
	err := s.db.Get(ctx, key, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if MFA is already enabled
	if user.MFAEnabled {
		return nil, fmt.Errorf("MFA is already enabled for this user")
	}

	// Generate TOTP secret
	secret, err := auth.GenerateMFASecret(user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate MFA secret: %w", err)
	}

	// Generate backup codes
	backupCodes, err := auth.GenerateBackupCodes()
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	// Hash backup codes for storage
	hashedBackupCodes := make([]string, len(backupCodes))
	for i, code := range backupCodes {
		hashedCode, err := auth.HashBackupCode(code)
		if err != nil {
			return nil, fmt.Errorf("failed to hash backup code: %w", err)
		}
		hashedBackupCodes[i] = hashedCode
	}

	// Generate QR code URL
	qrCodeURL := auth.GenerateQRCodeURL(secret, user.Email)

	// Update user with MFA setup data (temporarily, before verification)
	now := time.Now().UTC()
	user.MFASecret = secret
	user.MFABackupCodes = hashedBackupCodes
	user.MFASetupAt = &now

	_, err = s.db.Put(ctx, key, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user with MFA data: %w", err)
	}

	return &models.MFASetupResponse{
		Secret:      secret,
		QRCodeURL:   qrCodeURL,
		BackupCodes: backupCodes, // Send plain codes to user (will be invalidated after use)
	}, nil
}

// VerifyMFASetup verifies the MFA setup with a TOTP code
func (s *MFAService) VerifyMFASetup(userID string, code string) error {
	ctx := context.Background()

	// Get user
	key := database.NameKey("User", userID, nil)
	var user models.User
	err := s.db.Get(ctx, key, &user)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Check if MFA secret exists
	if user.MFASecret == "" {
		return fmt.Errorf("MFA setup not initiated")
	}

	// Validate TOTP code
	valid, err := auth.ValidateTOTPCode(user.MFASecret, code)
	if err != nil {
		return fmt.Errorf("failed to validate TOTP code: %w", err)
	}

	if !valid {
		return fmt.Errorf("invalid TOTP code")
	}

	// Enable MFA
	now := time.Now().UTC()
	user.MFAEnabled = true
	user.UpdatedAt = now

	_, err = s.db.Put(ctx, key, &user)
	if err != nil {
		return fmt.Errorf("failed to enable MFA: %w", err)
	}

	return nil
}

// ValidateMFACode validates a TOTP code for an existing user
func (s *MFAService) ValidateMFACode(userID string, code string) (bool, error) {
	ctx := context.Background()

	// Get user
	key := database.NameKey("User", userID, nil)
	var user models.User
	err := s.db.Get(ctx, key, &user)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if MFA is enabled
	if !user.MFAEnabled {
		return false, fmt.Errorf("MFA is not enabled for this user")
	}

	// Validate TOTP code
	valid, err := auth.ValidateTOTPCode(user.MFASecret, code)
	if err != nil {
		return false, fmt.Errorf("failed to validate TOTP code: %w", err)
	}

	return valid, nil
}

// ValidateBackupCode validates a backup code for MFA recovery
func (s *MFAService) ValidateBackupCode(userID string, backupCode string) (bool, error) {
	ctx := context.Background()

	// Get user
	key := database.NameKey("User", userID, nil)
	var user models.User
	err := s.db.Get(ctx, key, &user)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if MFA is enabled
	if !user.MFAEnabled {
		return false, fmt.Errorf("MFA is not enabled for this user")
	}

	// Validate backup code
	valid, updatedCodes := auth.ValidateBackupCode(backupCode, user.MFABackupCodes)
	if valid {
		// Update user with remaining backup codes
		user.MFABackupCodes = updatedCodes
		user.UpdatedAt = time.Now().UTC()

		_, err = s.db.Put(ctx, key, &user)
		if err != nil {
			return false, fmt.Errorf("failed to update backup codes: %w", err)
		}
	}

	return valid, nil
}

// DisableMFA disables MFA for a user (requires password and current TOTP)
func (s *MFAService) DisableMFA(userID string, password string, code string) error {
	ctx := context.Background()

	// Get user
	key := database.NameKey("User", userID, nil)
	var user models.User
	err := s.db.Get(ctx, key, &user)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify password
	if !auth.CheckPassword(password, user.PasswordHash) {
		return fmt.Errorf("invalid password")
	}

	// Validate current TOTP code
	valid, err := auth.ValidateTOTPCode(user.MFASecret, code)
	if err != nil {
		return fmt.Errorf("failed to validate TOTP code: %w", err)
	}

	if !valid {
		return fmt.Errorf("invalid TOTP code")
	}

	// Disable MFA
	user.MFAEnabled = false
	user.MFASecret = ""
	user.MFABackupCodes = nil
	user.MFASetupAt = nil
	user.UpdatedAt = time.Now().UTC()

	_, err = s.db.Put(ctx, key, &user)
	if err != nil {
		return fmt.Errorf("failed to disable MFA: %w", err)
	}

	return nil
}

// GetMFAStatus returns the current MFA status for a user
func (s *MFAService) GetMFAStatus(userID string) (*models.MFAStatusResponse, error) {
	ctx := context.Background()

	// Get user
	key := database.NameKey("User", userID, nil)
	var user models.User
	err := s.db.Get(ctx, key, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &models.MFAStatusResponse{
		Enabled:        user.MFAEnabled,
		SetupAt:        user.MFASetupAt,
		HasBackupCodes: len(user.MFABackupCodes) > 0,
	}, nil
}

// RegenerateBackupCodes generates new backup codes (requires current MFA verification)
func (s *MFAService) RegenerateBackupCodes(userID string, code string) ([]string, error) {
	ctx := context.Background()

	// Get user
	key := database.NameKey("User", userID, nil)
	var user models.User
	err := s.db.Get(ctx, key, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Verify current TOTP code
	valid, err := auth.ValidateTOTPCode(user.MFASecret, code)
	if err != nil {
		return nil, fmt.Errorf("failed to validate TOTP code: %w", err)
	}

	if !valid {
		return nil, fmt.Errorf("invalid TOTP code")
	}

	// Generate new backup codes
	backupCodes, err := auth.GenerateBackupCodes()
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	// Hash backup codes for storage
	hashedBackupCodes := make([]string, len(backupCodes))
	for i, code := range backupCodes {
		hashedCode, err := auth.HashBackupCode(code)
		if err != nil {
			return nil, fmt.Errorf("failed to hash backup code: %w", err)
		}
		hashedBackupCodes[i] = hashedCode
	}

	// Update user with new backup codes
	user.MFABackupCodes = hashedBackupCodes
	user.UpdatedAt = time.Now().UTC()

	_, err = s.db.Put(ctx, key, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to update backup codes: %w", err)
	}

	return backupCodes, nil
}
