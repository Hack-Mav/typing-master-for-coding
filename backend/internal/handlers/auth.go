package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"typing-master-backend/internal/auth"
	"typing-master-backend/internal/database"
	"typing-master-backend/internal/models"

	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Register handles user registration
func Register(db *database.DatastoreClient) gin.HandlerFunc {
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
			ID:                  userID,
			Handle:              req.Handle,
			Email:               req.Email,
			PasswordHash:        passwordHash,
			IsAnonymous:         false,
			Locale:              req.Locale,
			KeyboardLayout:      req.KeyboardLayout,
			PrivacyMode:         false,
			TelemetryConsent:    req.TelemetryConsent,
			DataProcessingConsent: req.DataProcessingConsent,
			Settings:            make(map[string]interface{}),
			CreatedAt:           now,
			UpdatedAt:           now,
			LastLoginAt:         &now,
		}

		// Save to Datastore
		key := datastore.NameKey("User", userID, nil)
		_, err = db.Put(ctx, key, &user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		// Generate JWT tokens
		tokens, err := auth.GenerateTokenPair(userID, user.Handle, user.Email, false, c.GetString("jwt_secret"))
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

		// Update last login time
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
		tokens, err := auth.GenerateTokenPair(user.ID, user.Handle, user.Email, false, jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user":   toUserResponse(&user),
			"tokens": tokens,
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
		tokens, err := auth.GenerateTokenPair(anonymousID, "Anonymous", "", true, jwtSecret)
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
	return models.UserResponse{
		ID:                  user.ID,
		Handle:              user.Handle,
		Email:               user.Email,
		IsAnonymous:         user.IsAnonymous,
		Locale:              user.Locale,
		KeyboardLayout:      user.KeyboardLayout,
		PrivacyMode:         user.PrivacyMode,
		TelemetryConsent:    user.TelemetryConsent,
		DataProcessingConsent: user.DataProcessingConsent,
		Settings:            user.Settings,
		CreatedAt:           user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           user.UpdatedAt.Format(time.RFC3339),
	}
}
