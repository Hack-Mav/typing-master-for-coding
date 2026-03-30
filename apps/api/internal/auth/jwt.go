package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// TokenClaims represents the JWT claims structure
type TokenClaims struct {
	UserID      string `json:"user_id"`
	Handle      string `json:"handle"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	IsAnonymous bool   `json:"is_anonymous"`
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// GenerateTokenPair creates both access and refresh tokens
func GenerateTokenPair(userID, handle, email, role string, isAnonymous bool, jwtSecret string) (*TokenPair, error) {
	// Access token expires in 15 minutes
	accessToken, err := generateToken(userID, handle, email, role, isAnonymous, jwtSecret, 15*time.Minute)
	if err != nil {
		return nil, err
	}

	// Refresh token expires in 7 days
	refreshToken, err := generateToken(userID, handle, email, role, isAnonymous, jwtSecret, 7*24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(15 * 60), // 15 minutes in seconds
		TokenType:    "Bearer",
	}, nil
}

// generateToken creates a JWT token with specified expiration
func generateToken(userID, handle, email, role string, isAnonymous bool, jwtSecret string, expiration time.Duration) (string, error) {
	now := time.Now()
	claims := TokenClaims{
		UserID:      userID,
		Handle:      handle,
		Email:       email,
		Role:        role,
		IsAnonymous: isAnonymous,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// ValidateToken validates and parses a JWT token
func ValidateToken(tokenString, jwtSecret string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrExpiredToken
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token from a valid refresh token
func RefreshAccessToken(refreshToken, jwtSecret string) (*TokenPair, error) {
	claims, err := ValidateToken(refreshToken, jwtSecret)
	if err != nil {
		return nil, err
	}

	// Generate new token pair
	return GenerateTokenPair(claims.UserID, claims.Handle, claims.Email, claims.Role, claims.IsAnonymous, jwtSecret)
}
