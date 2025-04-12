package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// TokenType defines the type of JWT token
type TokenType string

const (
	// AccessToken is used for API access
	AccessToken TokenType = "access"
	// RefreshToken is used to get new access tokens
	RefreshToken TokenType = "refresh"
)

// CustomClaims contains the claims for the JWT token
type CustomClaims struct {
	jwt.RegisteredClaims
	UserID   uint64    `json:"user_id"`
	Email    string    `json:"email"`
	Nickname string    `json:"nickname"`
	Role     string    `json:"role"`
	Type     TokenType `json:"type"`
}

// GenerateToken generates a new JWT token
func GenerateToken(
	userID uint64,
	email string,
	nickname string,
	role string,
	tokenType TokenType,
	secretKey string,
	duration time.Duration,
) (string, error) {
	// Create the claims
	claims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		UserID:   userID,
		Email:    email,
		Nickname: nickname,
		Role:     role,
		Type:     tokenType,
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// VerifyToken validates a JWT token
func VerifyToken(tokenString string, secretKey string) (*CustomClaims, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Extract and verify claims
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		// Check if token is expired
		if time.Now().UTC().After(claims.ExpiresAt.Time) {
			return nil, errors.New("token has expired")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
