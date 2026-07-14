package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

// MFA-related constants
const (
	BackupCodeLength = 8
	BackupCodesCount = 10
	IssuerName       = "Typing Master"
)

// GenerateMFASecret generates a new TOTP secret for MFA setup
func GenerateMFASecret(accountName string) (string, error) {
	// Generate a random 20-byte secret (160 bits) per RFC 4226/6238
	secret := make([]byte, 20)
	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("failed to generate random secret: %w", err)
	}

	// Encode as base32 (no padding) for TOTP compatibility
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret)
	return encoded, nil
}

// GenerateQRCodeURL generates a QR code URL for MFA setup
func GenerateQRCodeURL(secret, accountName string) string {
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s",
		IssuerName, accountName, secret, IssuerName)
}

// ValidateTOTPCode validates a TOTP code against the user's secret
// It checks the current time step plus one step before and after to allow
// for minor clock skew, per RFC 6238 recommendations.
func ValidateTOTPCode(secret, code string) (bool, error) {
	if len(code) != 6 || !isAllDigits(code) {
		return false, nil
	}

	now := time.Now()
	for _, offset := range []int{-1, 0, 1} {
		t := now.Add(time.Duration(offset) * totpPeriod)
		expected, err := generateTOTP(secret, t)
		if err != nil {
			return false, err
		}
		if expected == code {
			return true, nil
		}
	}

	return false, nil
}

const totpPeriod = 30 * time.Second

func isAllDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// generateTOTP computes a 6-digit TOTP code for the given secret and time.
func generateTOTP(secret string, t time.Time) (string, error) {
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(secret))
	if err != nil {
		return "", fmt.Errorf("invalid TOTP secret: %w", err)
	}

	counter := uint64(t.Unix()) / uint64(totpPeriod.Seconds())
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	sum := mac.Sum(nil)

	offset := sum[len(sum)-1] & 0x0F
	code := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7FFFFFFF
	code %= 1000000

	return fmt.Sprintf("%06d", code), nil
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
