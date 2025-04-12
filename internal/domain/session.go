package domain

import (
	"time"
)

// RegistrationSession represents a session for user registration process
type RegistrationSession struct {
	ID        string    `db:"id"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Nickname  string    `db:"nickname"`
	Email     string    `db:"email"`
	AcceptedPrivacyPolicy bool `db:"accepted_privacy_policy"`
	Code      string    `db:"code"`
	CodeExpires time.Time `db:"code_expires"`
	CreatedAt time.Time `db:"created_at"`
}

// LoginSession represents a session for user login process
type LoginSession struct {
	ID         string    `db:"id"`
	Email      string    `db:"email"`
	Code       string    `db:"code"`
	CodeExpires time.Time `db:"code_expires"`
	CreatedAt  time.Time `db:"created_at"`
}

// TokenSession represents an active refresh token session
type TokenSession struct {
	ID           string    `db:"id"`
	UserID       int64     `db:"user_id"`
	RefreshToken string    `db:"refresh_token"`
	UserAgent    string    `db:"user_agent"`
	IP           string    `db:"ip"`
	ExpiresAt    time.Time `db:"expires_at"`
	CreatedAt    time.Time `db:"created_at"`
}
