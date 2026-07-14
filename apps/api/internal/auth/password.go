package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost is the default bcrypt cost
	DefaultCost = 12

	// MinPasswordLength is the minimum accepted password length
	MinPasswordLength = 8
)

var (
	uppercase    = regexp.MustCompile(`[A-Z]`)
	lowercase    = regexp.MustCompile(`[a-z]`)
	digit        = regexp.MustCompile(`[0-9]`)
	special      = regexp.MustCompile(`[^A-Za-z0-9]`)
	errPassword  = errors.New("password must be at least 8 characters and contain at least one uppercase letter, one lowercase letter, one digit, and one special character")
)

// ValidatePassword checks password complexity requirements.
func ValidatePassword(password string) error {
	if len(password) < MinPasswordLength {
		return errPassword
	}
	if !uppercase.MatchString(password) {
		return errPassword
	}
	if !lowercase.MatchString(password) {
		return errPassword
	}
	if !digit.MatchString(password) {
		return errPassword
	}
	if !special.MatchString(password) {
		return errPassword
	}
	return nil
}

// HashPassword generates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword compares a password with its hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// HashRefreshToken returns a SHA-256 hash of a refresh token string.
func HashRefreshToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// CheckRefreshToken compares a refresh token with its hash.
func CheckRefreshToken(token, hash string) bool {
	return HashRefreshToken(token) == hash
}
