package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/typing-master-for-coding-backend/internal/auth"
	"github.com/typing-master-for-coding-backend/internal/cache"
	"github.com/typing-master-for-coding-backend/internal/database"
	"github.com/typing-master-for-coding-backend/internal/models"
	"github.com/typing-master-for-coding-backend/internal/rbac"
)

// setAuthCookies sets HttpOnly cookies for access and refresh tokens
func setAuthCookies(c *gin.Context, tokens *auth.TokenPair, isSecure bool) {
	// Access token cookie (15 minutes)
	c.SetSameSite(http.SameSiteStrictMode)
	accessCookie := http.Cookie{
		Name:     "access_token",
		Value:    tokens.AccessToken,
		Path:     "/",
		MaxAge:   15 * 60, // 15 minutes
		Secure:   isSecure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(c.Writer, &accessCookie)

	// Refresh token cookie (7 days)
	refreshCookie := http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 7 days
		Secure:   isSecure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(c.Writer, &refreshCookie)
}

// clearAuthCookies clears the auth cookies
func clearAuthCookies(c *gin.Context) {
	c.SetSameSite(http.SameSiteStrictMode)

	accessCookie := http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(c.Writer, &accessCookie)

	refreshCookie := http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(c.Writer, &refreshCookie)
}

// Register handles user registration
func Register(db *database.DatastoreClient, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Enforce password complexity
		if err := auth.ValidatePassword(req.Password); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx := context.Background()

		// Check if email already exists
		query := database.NewQuery("User").Filter("email =", req.Email).Limit(1)
		var existingUsers []models.User
		_, err := db.GetAll(ctx, query, &existingUsers)
		if err == nil {
			for _, existingUser := range existingUsers {
				if existingUser.Email == req.Email {
					c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
					return
				}
			}
		}

		// Check if handle already exists
		query = database.NewQuery("User").Filter("handle =", req.Handle).Limit(1)
		_, err = db.GetAll(ctx, query, &existingUsers)
		if err == nil {
			for _, existingUser := range existingUsers {
				if existingUser.Handle == req.Handle {
					c.JSON(http.StatusConflict, gin.H{"error": "Handle already taken"})
					return
				}
			}
		}

		// Hash password
		passwordHash, err := auth.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
			return
		}

		// Create user
		userID := uuid.New().String()
		now := time.Now().UTC()
		user := models.User{
			ID:                     userID,
			Handle:                 req.Handle,
			Email:                  req.Email,
			Role:                   "user", // Default role for new users
			PasswordHash:           passwordHash,
			IsAnonymous:            false,
			Locale:                 req.Locale,
			KeyboardLayout:         req.KeyboardLayout,
			PrivacyMode:            false,
			TelemetryConsent:       req.TelemetryConsent,
			DataProcessingConsent:  req.DataProcessingConsent,
			Settings:               make(map[string]interface{}),
			CreatedAt:              now,
			UpdatedAt:              now,
			LastLoginAt:            &now,
			EmailVerified:          false,
			EmailVerificationToken: uuid.New().String(),
		}

		// Save to Datastore
		key := database.NameKey("User", userID, nil)
		_, err = db.Put(ctx, key, &user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		// Automatically assign the basic "user" role via RBAC system
		rbacService := rbac.NewService(db)
		_, err = rbacService.AssignRole(ctx, rbac.AssignRoleRequest{
			UserID: userID,
			RoleID: "user",
		}, "system")
		if err != nil {
			// Log error but don't fail user creation
			fmt.Printf("Failed to assign user role: %v\n", err)
		}

		// Generate JWT tokens and store the refresh token hash for rotation
		tokens, err := auth.GenerateTokenPair(userID, user.Handle, user.Email, user.Role, false, jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
			return
		}
		user.RefreshTokenHash = auth.HashRefreshToken(tokens.RefreshToken)
		_, _ = db.Put(ctx, key, &user)

		// Set HttpOnly cookies for security
		c.SetCookie("access_token", tokens.AccessToken, int(15*60), "/", "", true, true)
		c.SetCookie("refresh_token", tokens.RefreshToken, int(7*24*60*60), "/", "", true, true)

		c.JSON(http.StatusCreated, gin.H{
			"user":   toUserResponse(&user),
			"tokens": tokens,
		})
	}
}

// Login handles user authentication
func Login(db *database.DatastoreClient, cacheClient *cache.InMemoryCache, jwtSecret string) gin.HandlerFunc {
	tracker := auth.NewLoginAttemptTracker(cacheClient)

	return func(c *gin.Context) {
		var req models.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Account lockout / rate limiting
		if tracker.IsLockedOut(req.Email) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Account locked due to too many failed attempts. Try again later."})
			return
		}

		ctx := context.Background()

		// Find user by email
		query := database.NewQuery("User").Filter("email =", req.Email).Limit(1)
		var users []models.User
		keys, err := db.GetAll(ctx, query, &users)
		if err != nil || len(users) == 0 {
			tracker.RecordFailedAttempt(req.Email)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		// Find the user with the matching email
		var user models.User
		var userKey *database.Key
		found := false
		for i, u := range users {
			if u.Email == req.Email {
				user = u
				userKey = keys[i]
				found = true
				break
			}
		}

		if !found {
			tracker.RecordFailedAttempt(req.Email)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		user.ID = userKey.Name

		// Check password
		if !auth.CheckPassword(req.Password, user.PasswordHash) {
			tracker.RecordFailedAttempt(req.Email)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		// Require email verification before completing login
		if !user.EmailVerified {
			c.JSON(http.StatusForbidden, gin.H{"error": "Email not verified", "code": "EMAIL_VERIFICATION_REQUIRED"})
			return
		}

		// Check if MFA is enabled for this user
		if user.MFAEnabled {
			// MFA is required - don't complete login yet
			c.JSON(http.StatusOK, gin.H{
				"requires_mfa": true,
				"user_id":      user.ID,
				"message":      "MFA verification required",
			})
			return
		}

		// MFA not enabled - complete login normally
		tracker.RecordSuccessfulLogin(req.Email)

		now := time.Now().UTC()
		user.LastLoginAt = &now
		user.UpdatedAt = now

		key := database.NameKey("User", user.ID, nil)

		// Generate JWT tokens and rotate refresh token
		tokens, err := auth.GenerateTokenPair(user.ID, user.Handle, user.Email, user.Role, false, jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
			return
		}
		user.RefreshTokenHash = auth.HashRefreshToken(tokens.RefreshToken)

		_, err = db.Put(ctx, key, &user)
		if err != nil {
			// Log error but don't fail login
			fmt.Printf("Failed to update user after login: %v\n", err)
		}

		// Set HttpOnly cookies for security
		c.SetCookie("access_token", tokens.AccessToken, int(15*60), "/", "", true, true)
		c.SetCookie("refresh_token", tokens.RefreshToken, int(7*24*60*60), "/", "", true, true)

		c.JSON(http.StatusOK, gin.H{
			"requires_mfa": false,
			"user":         toUserResponse(&user),
			"tokens":       tokens,
		})
	}
}

// RefreshToken handles token refresh with rotation and reuse detection
func RefreshToken(db *database.DatastoreClient, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate refresh token
		claims, err := auth.ValidateToken(req.RefreshToken, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
			return
		}

		// Load the user and verify the supplied token matches the stored refresh token hash
		ctx := context.Background()
		key := database.NameKey("User", claims.UserID, nil)
		var user models.User
		err = db.Get(ctx, key, &user)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
			return
		}

		user.ID = claims.UserID
		if !auth.CheckRefreshToken(req.RefreshToken, user.RefreshTokenHash) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
			return
		}

		// Generate new token pair and rotate the refresh token hash
		tokens, err := auth.GenerateTokenPair(user.ID, user.Handle, user.Email, user.Role, claims.IsAnonymous, jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
			return
		}
		user.RefreshTokenHash = auth.HashRefreshToken(tokens.RefreshToken)
		user.UpdatedAt = time.Now().UTC()
		_, _ = db.Put(ctx, key, &user)

		// Set HttpOnly cookies for security
		c.SetCookie("access_token", tokens.AccessToken, int(15*60), "/", "", true, true)
		c.SetCookie("refresh_token", tokens.RefreshToken, int(7*24*60*60), "/", "", true, true)

		c.JSON(http.StatusOK, tokens)
	}
}

// CreateAnonymousSession creates an anonymous user session
func CreateAnonymousSession(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.AnonymousSessionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Generate anonymous user ID based on device ID
		anonymousID := fmt.Sprintf("anon_%s", req.DeviceID)

		// Generate JWT tokens for anonymous user with longer expiration (24 hours)
		tokens, err := auth.GenerateTokenPair(anonymousID, "Anonymous", "", "user", true, jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
			return
		}

		// Set HttpOnly cookies for security
		c.SetCookie("access_token", tokens.AccessToken, int(15*60), "/", "", true, true)
		c.SetCookie("refresh_token", tokens.RefreshToken, int(7*24*60*60), "/", "", true, true)

		// Create anonymous user response
		anonymousUser := gin.H{
			"id":                      anonymousID,
			"handle":                  "Anonymous",
			"email":                   "",
			"is_anonymous":            true,
			"keyboard_layout":         req.KeyboardLayout,
			"locale":                  req.Locale,
			"privacy_mode":            true, // Anonymous users get privacy mode by default
			"telemetry_consent":       false,
			"data_processing_consent": false,
			"settings":                make(map[string]interface{}),
			"created_at":              time.Now().UTC().Format(time.RFC3339),
			"updated_at":              time.Now().UTC().Format(time.RFC3339),
			"role":                    "user",
		}

		c.JSON(http.StatusOK, gin.H{
			"user":         anonymousUser,
			"tokens":       tokens,
			"requires_mfa": false, // Anonymous users don't use MFA
		})
	}
}

// GetProfile retrieves the authenticated user's profile
func GetProfile(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Check if anonymous user
		if strings.HasPrefix(userID.(string), "anon_") {
			c.JSON(http.StatusOK, gin.H{
				"id":           userID,
				"handle":       "Anonymous",
				"is_anonymous": true,
			})
			return
		}

		ctx := context.Background()
		key := database.NameKey("User", userID.(string), nil)

		var user models.User
		err := db.Get(ctx, key, &user)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		user.ID = userID.(string)
		c.JSON(http.StatusOK, toUserResponse(&user))
	}
}

// UpdateProfile updates the authenticated user's profile
func UpdateProfile(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Anonymous users cannot update profile
		if strings.HasPrefix(userID.(string), "anon_") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anonymous users cannot update profile"})
			return
		}

		var req models.UpdateProfileRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx := context.Background()
		key := database.NameKey("User", userID.(string), nil)

		var user models.User
		err := db.Get(ctx, key, &user)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// Update fields if provided
		if req.Handle != nil {
			user.Handle = *req.Handle
		}
		if req.Locale != nil {
			user.Locale = *req.Locale
		}
		if req.KeyboardLayout != nil {
			user.KeyboardLayout = *req.KeyboardLayout
		}
		if req.PrivacyMode != nil {
			user.PrivacyMode = *req.PrivacyMode
		}
		if req.TelemetryConsent != nil {
			user.TelemetryConsent = *req.TelemetryConsent
		}
		if req.DataProcessingConsent != nil {
			user.DataProcessingConsent = *req.DataProcessingConsent
		}
		if req.Settings != nil {
			user.Settings = req.Settings
		}

		user.UpdatedAt = time.Now().UTC()

		// Save updated user
		_, err = db.Put(ctx, key, &user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}

		user.ID = userID.(string)
		c.JSON(http.StatusOK, toUserResponse(&user))
	}
}

// Helper function to convert User to UserResponse
func toUserResponse(user *models.User) models.UserResponse {
	var mfaSetupAt string
	if user.MFASetupAt != nil {
		mfaSetupAt = user.MFASetupAt.Format(time.RFC3339)
	}

	return models.UserResponse{
		ID:                    user.ID,
		Handle:                user.Handle,
		Email:                 user.Email,
		IsAnonymous:           user.IsAnonymous,
		Locale:                user.Locale,
		KeyboardLayout:        user.KeyboardLayout,
		PrivacyMode:           user.PrivacyMode,
		TelemetryConsent:      user.TelemetryConsent,
		DataProcessingConsent: user.DataProcessingConsent,
		Settings:              user.Settings,
		CreatedAt:             user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:             user.UpdatedAt.Format(time.RFC3339),

		// MFA fields
		MFAEnabled: user.MFAEnabled,
		MFASetupAt: mfaSetupAt,

		// Security fields
		EmailVerified: user.EmailVerified,
	}
}

// MFA Handler Functions

// SetupMFA handles MFA setup initiation
func SetupMFA(db *database.DatastoreClient) gin.HandlerFunc {
	mfaService := NewMFAService(db)

	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Anonymous users cannot set up MFA
		if strings.HasPrefix(userID.(string), "anon_") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anonymous users cannot set up MFA"})
			return
		}

		var req models.MFASetupRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		setupResponse, err := mfaService.SetupMFA(userID.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, setupResponse)
	}
}

// VerifyMFASetup handles MFA setup verification
func VerifyMFASetup(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Anonymous users cannot verify MFA
		if strings.HasPrefix(userID.(string), "anon_") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anonymous users cannot set up MFA"})
			return
		}

		var req models.MFAVerifyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		mfaService := NewMFAService(db)
		err := mfaService.VerifyMFASetup(userID.(string), req.Code)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, models.MFAVerifyResponse{
			Success: true,
			Message: "MFA setup verified successfully",
		})
	}
}

// GetMFAStatus returns the current MFA status for the authenticated user
func GetMFAStatus(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Anonymous users don't have MFA
		if strings.HasPrefix(userID.(string), "anon_") {
			c.JSON(http.StatusOK, models.MFAStatusResponse{
				Enabled:        false,
				HasBackupCodes: false,
			})
			return
		}

		mfaService := NewMFAService(db)
		status, err := mfaService.GetMFAStatus(userID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, status)
	}
}

// DisableMFA handles MFA disable request
func DisableMFA(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Anonymous users don't have MFA to disable
		if strings.HasPrefix(userID.(string), "anon_") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anonymous users cannot disable MFA"})
			return
		}

		var req models.MFADisableRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		mfaService := NewMFAService(db)
		err := mfaService.DisableMFA(userID.(string), req.Password, req.Code)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "MFA disabled successfully"})
	}
}

// RegenerateMFABackupCodes generates new backup codes for the user
func RegenerateMFABackupCodes(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Anonymous users cannot regenerate backup codes
		if strings.HasPrefix(userID.(string), "anon_") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anonymous users cannot regenerate backup codes"})
			return
		}

		var req models.MFAVerifyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		mfaService := NewMFAService(db)
		backupCodes, err := mfaService.RegenerateBackupCodes(userID.(string), req.Code)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"backup_codes": backupCodes,
			"message":      "Backup codes regenerated successfully",
		})
	}
}

// LoginWithMFA completes login with MFA verification
func LoginWithMFA(db *database.DatastoreClient, cacheClient *cache.InMemoryCache, jwtSecret string) gin.HandlerFunc {
	tracker := auth.NewLoginAttemptTracker(cacheClient)

	return func(c *gin.Context) {
		var req models.LoginWithMFARequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Account lockout / rate limiting
		if tracker.IsLockedOut(req.Email) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Account locked due to too many failed attempts. Try again later."})
			return
		}

		ctx := context.Background()

		// Find user by email
		query := database.NewQuery("User").Filter("email =", req.Email).Limit(1)
		var users []models.User
		keys, err := db.GetAll(ctx, query, &users)
		if err != nil || len(users) == 0 {
			tracker.RecordFailedAttempt(req.Email)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		user := users[0]
		userKey := keys[0]
		user.ID = userKey.Name

		// Verify password again for security
		if !auth.CheckPassword(req.Password, user.PasswordHash) {
			tracker.RecordFailedAttempt(req.Email)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		// Require email verification before completing login
		if !user.EmailVerified {
			c.JSON(http.StatusForbidden, gin.H{"error": "Email not verified", "code": "EMAIL_VERIFICATION_REQUIRED"})
			return
		}

		// Check if MFA is enabled
		if !user.MFAEnabled {
			c.JSON(http.StatusBadRequest, gin.H{"error": "MFA is not enabled for this user"})
			return
		}

		// Validate TOTP code or backup code
		mfaService := NewMFAService(db)
		valid, err := mfaService.ValidateMFACode(user.ID, req.Code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if !valid {
			// Try backup codes if TOTP failed
			valid, err = mfaService.ValidateBackupCode(user.ID, req.Code)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if !valid {
				tracker.RecordFailedAttempt(req.Email)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid MFA code"})
				return
			}
		}

		tracker.RecordSuccessfulLogin(req.Email)

		// MFA verification successful - complete login
		now := time.Now().UTC()
		user.LastLoginAt = &now
		user.UpdatedAt = now

		key := database.NameKey("User", user.ID, nil)

		// Generate JWT tokens and rotate refresh token
		tokens, err := auth.GenerateTokenPair(user.ID, user.Handle, user.Email, user.Role, false, jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
			return
		}
		user.RefreshTokenHash = auth.HashRefreshToken(tokens.RefreshToken)

		_, err = db.Put(ctx, key, &user)
		if err != nil {
			// Log error but don't fail login
			fmt.Printf("Failed to update user after login: %v\n", err)
		}

		// Set HttpOnly cookies for security
		c.SetCookie("access_token", tokens.AccessToken, int(15*60), "/", "", true, true)
		c.SetCookie("refresh_token", tokens.RefreshToken, int(7*24*60*60), "/", "", true, true)

		c.JSON(http.StatusOK, gin.H{
			"requires_mfa": false,
			"user":         toUserResponse(&user),
			"tokens":       tokens,
		})
	}
}

// Logout handles user logout by clearing HttpOnly cookies
func Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Clear HttpOnly cookies by setting them with empty values and past expiration
		c.SetCookie("access_token", "", -1, "/", "", true, true)
		c.SetCookie("refresh_token", "", -1, "/", "", true, true)

		c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
	}
}

// VerifyEmail verifies a user's email address using a one-time token.
func VerifyEmail(db *database.DatastoreClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email" binding:"required,email"`
			Token string `json:"token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx := context.Background()
		query := database.NewQuery("User").Filter("email =", req.Email).Limit(1)
		var users []models.User
		keys, err := db.GetAll(ctx, query, &users)
		if err != nil || len(users) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification request"})
			return
		}

		user := users[0]
		userKey := keys[0]
		user.ID = userKey.Name

		if user.EmailVerificationToken != req.Token {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification token"})
			return
		}

		user.EmailVerified = true
		user.EmailVerificationToken = ""
		user.UpdatedAt = time.Now().UTC()

		_, err = db.Put(ctx, userKey, &user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify email"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
	}
}
