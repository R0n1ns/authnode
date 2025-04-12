package service

import (
	"errors"
	"fmt"
	"time"

	"authmicro/internal/config"
	"authmicro/internal/domain/entity"
	"github.com/golang-jwt/jwt/v5"
)

// Error definitions for token service (shared with auth_service.go)
var (
	ErrTokenExpired = errors.New("token has expired")
	ErrTokenInvalid = errors.New("token is invalid")
)

// TokenService defines the interface for token operations
type TokenService interface {
	GenerateAccessToken(user *entity.User) (string, error)
	GenerateRefreshToken(user *entity.User) (string, error)
	ValidateAccessToken(tokenString string) (*entity.UserClaims, error)
	ValidateRefreshToken(tokenString string) (*entity.UserClaims, error)
}

// tokenService implements the TokenService interface
type tokenService struct {
	config config.JWTConfig
}

// NewTokenService creates a new token service
func NewTokenService(config config.JWTConfig) TokenService {
	return &tokenService{
		config: config,
	}
}

// GenerateAccessToken generates a new access token for a user
func (s *tokenService) GenerateAccessToken(user *entity.User) (string, error) {
	// Create the Claims
	claims := jwt.MapClaims{
		"userId":    user.ID,
		"email":     user.Email,
		"role":      user.Role,
		"tokenType": "access",
		"exp":       time.Now().UTC().Add(time.Duration(s.config.AccessTokenExpiryMinutes) * time.Minute).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret key
	tokenString, err := token.SignedString([]byte(s.config.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken generates a new refresh token for a user
func (s *tokenService) GenerateRefreshToken(user *entity.User) (string, error) {
	// Create the Claims
	claims := jwt.MapClaims{
		"userId":    user.ID,
		"email":     user.Email,
		"role":      user.Role,
		"tokenType": "refresh",
		"exp":       time.Now().UTC().Add(time.Duration(s.config.RefreshTokenExpiryDays) * 24 * time.Hour).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret key
	tokenString, err := token.SignedString([]byte(s.config.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateAccessToken validates an access token and returns the claims
func (s *tokenService) ValidateAccessToken(tokenString string) (*entity.UserClaims, error) {
	return s.validateToken(tokenString, "access")
}

// ValidateRefreshToken validates a refresh token and returns the claims
func (s *tokenService) ValidateRefreshToken(tokenString string) (*entity.UserClaims, error) {
	return s.validateToken(tokenString, "refresh")
}

// validateToken validates a token and returns the claims
func (s *tokenService) validateToken(tokenString, tokenType string) (*entity.UserClaims, error) {
	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Return secret key
		return []byte(s.config.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Check if token is valid
	if !token.Valid {
		return nil, ErrTokenInvalid
	}

	// Get claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Check token type
	if claims["tokenType"] != tokenType {
		return nil, fmt.Errorf("invalid token type: expected %s", tokenType)
	}

	// Extract claims
	userID, ok := claims["userId"].(string)
	if !ok {
		return nil, errors.New("invalid userId claim")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return nil, errors.New("invalid email claim")
	}

	roleStr, ok := claims["role"].(string)
	if !ok {
		return nil, errors.New("invalid role claim")
	}

	// Create user claims
	userClaims := &entity.UserClaims{
		UserID:    userID,
		Email:     email,
		Role:      entity.Role(roleStr),
		TokenType: tokenType,
	}

	return userClaims, nil
}
