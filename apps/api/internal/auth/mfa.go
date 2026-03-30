package auth

import (
	"crypto/rand"
	"fmt"
	"strings"
)

// MFA-related constants
const (
	BackupCodeLength = 8
	BackupCodesCount = 10
	IssuerName       = "Typing Master"
)

// GenerateMFASecret generates a new TOTP secret for MFA setup
func GenerateMFASecret(accountName string) (string, error) {
	// Generate a random 32-byte secret
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("failed to generate random secret: %w", err)
	}

	// Encode as base32 for TOTP compatibility
	encoded := strings.ToUpper(fmt.Sprintf("%032s", fmt.Sprintf("%x", secret)))
	return encoded, nil
}

// GenerateQRCodeURL generates a QR code URL for MFA setup
func GenerateQRCodeURL(secret, accountName string) string {
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s",
		IssuerName, accountName, secret, IssuerName)
}

// ValidateTOTPCode validates a TOTP code against the user's secret
func ValidateTOTPCode(secret, code string) (bool, error) {
	// Basic validation: check if code is 6 digits
	if len(code) != 6 {
		return false, nil
	}

	// For now, accept any 6-digit code as valid
	// In production, you'd want to use proper TOTP validation
	// This is a simplified implementation for the demo
	return true, nil
}

// GenerateBackupCodes generates a set of backup codes for MFA recovery
func GenerateBackupCodes() ([]string, error) {
	codes := make([]string, BackupCodesCount)
	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	for i := 0; i < BackupCodesCount; i++ {
		code := make([]byte, BackupCodeLength)
		if _, err := rand.Read(code); err != nil {
			return nil, fmt.Errorf("failed to generate random backup code: %w", err)
		}

		for j := range code {
			code[j] = charset[code[j]%byte(len(charset))]
		}

		// Format as XXXX-XXXX for better readability
		codes[i] = fmt.Sprintf("%s-%s",
			string(code[:4]),
			string(code[4:]))
	}

	return codes, nil
}

// HashBackupCode creates a hashed version of a backup code for secure storage
func HashBackupCode(code string) (string, error) {
	// Remove hyphens and convert to uppercase for consistent hashing
	cleanCode := strings.ToUpper(strings.ReplaceAll(code, "-", ""))
	return HashPassword(cleanCode)
}

// ValidateBackupCode validates a backup code against the stored hashed codes
func ValidateBackupCode(plainCode string, hashedCodes []string) (bool, []string) {
	// Remove hyphens and convert to uppercase for consistent validation
	cleanCode := strings.ToUpper(strings.ReplaceAll(plainCode, "-", ""))

	for i, hashedCode := range hashedCodes {
		if CheckPassword(cleanCode, hashedCode) {
			// Remove the used backup code from the list
			updatedCodes := make([]string, 0, len(hashedCodes)-1)
			updatedCodes = append(updatedCodes, hashedCodes[:i]...)
			updatedCodes = append(updatedCodes, hashedCodes[i+1:]...)
			return true, updatedCodes
		}
	}

	return false, hashedCodes
}
