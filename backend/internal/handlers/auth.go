package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/typing-master-for-coding-backend/internal/auth"
	"github.com/typing-master-for-coding-backend/internal/database"
	"github.com/typing-master-for-coding-backend/internal/models"
	"github.com/typing-master-for-coding-backend/internal/rbac"

	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Register handles user registration
func Register(db *database.DatastoreClient, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx := context.Background()

		// Check if email already exists
		query := datastore.NewQuery("User").Filter("email =", req.Email).Limit(1)
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
		query = datastore.NewQuery("User").Filter("handle =", req.Handle).Limit(1)
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
			ID:                    userID,
			Handle:                req.Handle,
			Email:                 req.Email,
			Role:                  "user", // Default role for new users
			PasswordHash:          passwordHash,
			IsAnonymous:           false,
			Locale:                req.Locale,
			KeyboardLayout:        req.KeyboardLayout,
			PrivacyMode:           false,
			TelemetryConsent:      req.TelemetryConsent,
			DataProcessingConsent: req.DataProcessingConsent,
			Settings:              make(map[string]interface{}),
			CreatedAt:             now,
			UpdatedAt:             now,
			LastLoginAt:           &now,
		}

		// Save to Datastore
		key := datastore.NameKey("User", userID, nil)
		_, err = db.Put(ctx, key, &user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		// Automatically assign the basic "user" role via RBAC system
		rbacService := rbac.NewService(db)
		_, err = rbacService.AssignRole(ctx, models.AssignRoleRequest{
			UserID: userID,
			RoleID: "user",
		}, "system")
		if err != nil {
			// Log error but don't fail user creation
			fmt.Printf("Failed to assign user role: %v\n", err)
		}

		// Generate JWT tokens
		tokens, err := auth.GenerateTokenPair(userID, user.Handle, user.Email, user.Role, false, jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"user":   toUserResponse(&user),
			"tokens": tokens,
		})
	}
}

// Login handles user authentication
func Login(db *database.DatastoreClient, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx := context.Background()

		// Find user by email
		query := datastore.NewQuery("User").Filter("email =", req.Email).Limit(1)
		var users []models.User
		keys, err := db.GetAll(ctx, query, &users)
		if err != nil || len(users) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		// Find the user with the matching email
		var user models.User
		var userKey *datastore.Key
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		user.ID = userKey.Name

		// Check password
		if !auth.CheckPassword(req.Password, user.PasswordHash) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
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
		now := time.Now().UTC()
		user.LastLoginAt = &now
		user.UpdatedAt = now

		key := datastore.NameKey("User", user.ID, nil)
		_, err = db.Put(ctx, key, &user)
		if err != nil {
			// Log error but don't fail login
			fmt.Printf("Failed to update last login time: %v\n", err)
		}

		// Generate JWT tokens
		tokens, err := auth.GenerateTokenPair(user.ID, user.Handle, user.Email, user.Role, false, jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"requires_mfa": false,
			"user":         toUserResponse(&user),
			"tokens":       tokens,
		})
	}
}

// RefreshToken handles token refresh
func RefreshToken(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate and refresh token
		tokens, err := auth.RefreshAccessToken(req.RefreshToken, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
			return
		}

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

		// Generate JWT tokens for anonymous user
		tokens, err := auth.GenerateTokenPair(anonymousID, "Anonymous", "", "user", true, jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user": gin.H{
				"id":              anonymousID,
				"handle":          "Anonymous",
				"is_anonymous":    true,
				"keyboard_layout": req.KeyboardLayout,
				"locale":          req.Locale,
			},
			"tokens": tokens,
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
		key := datastore.NameKey("User", userID.(string), nil)

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
		key := datastore.NameKey("User", userID.(string), nil)

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
func LoginWithMFA(db *database.DatastoreClient, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.LoginWithMFARequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx := context.Background()

		// Find user by email
		query := datastore.NewQuery("User").Filter("email =", req.Email).Limit(1)
		var users []models.User
		keys, err := db.GetAll(ctx, query, &users)
		if err != nil || len(users) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		user := users[0]
		userKey := keys[0]
		user.ID = userKey.Name

		// Verify password again for security
		if !auth.CheckPassword(req.Password, user.PasswordHash) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
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
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid MFA code"})
				return
			}
		}

		// MFA verification successful - complete login
		now := time.Now().UTC()
		user.LastLoginAt = &now
		user.UpdatedAt = now

		key := datastore.NameKey("User", user.ID, nil)
		_, err = db.Put(ctx, key, &user)
		if err != nil {
			// Log error but don't fail login
			fmt.Printf("Failed to update last login time: %v\n", err)
		}

		// Generate JWT tokens
		tokens, err := auth.GenerateTokenPair(user.ID, user.Handle, user.Email, user.Role, false, jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"requires_mfa": false,
			"user":         toUserResponse(&user),
			"tokens":       tokens,
		})
	}
}
