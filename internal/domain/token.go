package domain

import (
	"time"
)

// TokenClaims represents the claims in a JWT token
type TokenClaims struct {
	UserID    int64    `json:"userId"`
	Email     string   `json:"email"`
	Nickname  string   `json:"nickname"`
	Roles     []string `json:"roles"`
	ExpiresAt int64    `json:"exp"`
	IssuedAt  int64    `json:"iat"`
}

// TokenPair represents a pair of access and refresh tokens
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// RefreshSession stores information about a refresh token session
type RefreshSession struct {
	ID           string
	UserID       int64
	RefreshToken string
	ExpiresAt    time.Time
	CreatedAt    time.Time
}
