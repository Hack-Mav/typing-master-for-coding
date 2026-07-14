package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateMFASecret(t *testing.T) {
	secret, err := GenerateMFASecret("test@example.com")
	require.NoError(t, err)
	assert.NotEmpty(t, secret)
	assert.Equal(t, 32, len(secret), "base32 encoded 20-byte secret should be 32 chars")
	assert.Equal(t, strings.ToUpper(secret), secret, "base32 secret should be uppercase")
	assert.True(t, isBase32(secret), "secret should be valid base32")
}

func TestValidateTOTPCode(t *testing.T) {
	secret, err := GenerateMFASecret("test@example.com")
	require.NoError(t, err)

	now := time.Now()
	validCode, err := generateTOTP(secret, now)
	require.NoError(t, err)

	// Valid code at current time
	ok, err := ValidateTOTPCode(secret, validCode)
	require.NoError(t, err)
	assert.True(t, ok)

	// Invalid code
	ok, err = ValidateTOTPCode(secret, "000000")
	require.NoError(t, err)
	assert.False(t, ok)

	// Wrong length
	ok, err = ValidateTOTPCode(secret, "12345")
	require.NoError(t, err)
	assert.False(t, ok)

	// Non-digit code
	ok, err = ValidateTOTPCode(secret, "12345a")
	require.NoError(t, err)
	assert.False(t, ok)

	// Code within one step before/after should still be valid
	for _, offset := range []int{-1, 1} {
		otherTime := now.Add(time.Duration(offset) * totpPeriod)
		otherCode, err := generateTOTP(secret, otherTime)
		require.NoError(t, err)
		ok, err := ValidateTOTPCode(secret, otherCode)
		require.NoError(t, err)
		assert.True(t, ok, "code %d steps away from current should be valid", offset)
	}

	// Code outside the skew window should be invalid
	farTime := now.Add(2 * totpPeriod)
	farCode, err := generateTOTP(secret, farTime)
	require.NoError(t, err)
	ok, err = ValidateTOTPCode(secret, farCode)
	require.NoError(t, err)
	assert.False(t, ok)
}

func isBase32(s string) bool {
	for _, r := range s {
		if !((r >= 'A' && r <= 'Z') || (r >= '2' && r <= '7')) {
			return false
		}
	}
	return true
}
