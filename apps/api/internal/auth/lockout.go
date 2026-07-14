package auth

import (
	"fmt"
	"time"
)

const (
	// maxFailedAttempts is the number of failed login attempts before a lockout.
	maxFailedAttempts = 5
	// lockoutDuration is how long an account remains locked after too many failures.
	lockoutDuration = 15 * time.Minute
	// attemptWindow is the TTL for the failed-attempt counter.
	attemptWindow = 15 * time.Minute
)

// Cache is the minimal interface required by the login tracker.
type Cache interface {
	Get(key string) (interface{}, bool)
	SetWithTTL(key string, value interface{}, ttl time.Duration)
	Delete(key string)
	IncrWithTTL(key string, ttl time.Duration) int64
}

// LoginAttemptTracker provides cache-backed account lockout for authentication.
type LoginAttemptTracker struct {
	store Cache
}

// NewLoginAttemptTracker creates a new tracker backed by the provided cache.
func NewLoginAttemptTracker(cache Cache) *LoginAttemptTracker {
	return &LoginAttemptTracker{store: cache}
}

// IsLockedOut returns true if the email is currently under lockout.
func (t *LoginAttemptTracker) IsLockedOut(email string) bool {
	if t.store == nil {
		return false
	}
	_, found := t.store.Get(lockoutKey(email))
	return found
}

// RecordFailedAttempt increments the failed attempt counter and returns the
// new count and whether the account is now locked.
func (t *LoginAttemptTracker) RecordFailedAttempt(email string) (int64, bool) {
	if t.store == nil {
		return 1, false
	}
	count := t.store.IncrWithTTL(attemptsKey(email), attemptWindow)
	locked := count >= maxFailedAttempts
	if locked {
		t.store.SetWithTTL(lockoutKey(email), true, lockoutDuration)
	}
	return count, locked
}

// RecordSuccessfulLogin clears the failed attempts and lockout keys.
func (t *LoginAttemptTracker) RecordSuccessfulLogin(email string) {
	if t.store == nil {
		return
	}
	t.store.Delete(attemptsKey(email))
	t.store.Delete(lockoutKey(email))
}

func attemptsKey(email string) string {
	return fmt.Sprintf("login_attempts:%s", email)
}

func lockoutKey(email string) string {
	return fmt.Sprintf("login_lockout:%s", email)
}
