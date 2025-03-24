package entity

import (
	"time"
)

// Role represents user role for authorization
type Role string

const (
	// RoleUser represents a regular user
	RoleUser Role = "user"
	// RoleAdmin represents an administrator
	RoleAdmin Role = "admin"
)

// User represents a user entity
type User struct {
	ID                  string    `json:"id"`
	FirstName           string    `json:"firstName"`
	LastName            string    `json:"lastName"`
	Nickname            string    `json:"nickname"`
	Email               string    `json:"email"`
	EmailVerified       bool      `json:"emailVerified"`
	Role                Role      `json:"role"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
	LastLoginAt         time.Time `json:"lastLoginAt,omitempty"`
	PasswordReset       bool      `json:"passwordReset"`
	AcceptedTermsAt     time.Time `json:"acceptedTermsAt"`
	AcceptedPrivacyAt   time.Time `json:"acceptedPrivacyAt"`
	ActiveRefreshTokens map[string]time.Time
}

// NewUser creates a new user
func NewUser(
	id string,
	firstName string,
	lastName string,
	nickname string,
	email string,
	acceptedPrivacy bool,
) *User {
	now := time.Now()
	user := &User{
		ID:                  id,
		FirstName:           firstName,
		LastName:            lastName,
		Nickname:            nickname,
		Email:               email,
		EmailVerified:       false,
		Role:                RoleUser,
		CreatedAt:           now,
		UpdatedAt:           now,
		PasswordReset:       false,
		AcceptedPrivacyAt:   now,
		ActiveRefreshTokens: make(map[string]time.Time),
	}

	if acceptedPrivacy {
		user.AcceptedPrivacyAt = now
	}

	return user
}

// UserClaims represents the JWT token claims for a user
type UserClaims struct {
	UserID    string `json:"userId"`
	Email     string `json:"email"`
	Role      Role   `json:"role"`
	TokenType string `json:"tokenType"`
}

// TokenPair represents an access and refresh token pair
type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// RegistrationSession represents a registration session
type RegistrationSession struct {
	ID                    string    `json:"id"`
	Email                 string    `json:"email"`
	FirstName             string    `json:"firstName"`
	LastName              string    `json:"lastName"`
	Nickname              string    `json:"nickname"`
	VerificationCode      string    `json:"verificationCode"`
	VerificationCodeExp   time.Time `json:"verificationCodeExp"`
	AcceptedPrivacyPolicy bool      `json:"acceptedPrivacyPolicy"`
	CreatedAt             time.Time `json:"createdAt"`
}

// LoginSession represents a login session
type LoginSession struct {
	Email        string    `json:"email"`
	LoginCode    string    `json:"loginCode"`
	LoginCodeExp time.Time `json:"loginCodeExp"`
	CreatedAt    time.Time `json:"createdAt"`
}
