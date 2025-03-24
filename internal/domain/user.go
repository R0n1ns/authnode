package domain

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID                    int64     `json:"id" db:"id"`
	FirstName             string    `json:"firstName" db:"first_name"`
	LastName              string    `json:"lastName" db:"last_name"`
	Nickname              string    `json:"nickname" db:"nickname"`
	Email                 string    `json:"email" db:"email"`
	EmailVerified         bool      `json:"emailVerified" db:"email_verified"`
	AcceptedPrivacyPolicy bool      `json:"acceptedPrivacyPolicy" db:"accepted_privacy_policy"`
	CreatedAt             time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt             time.Time `json:"updatedAt" db:"updated_at"`
}

// RegistrationRequest represents the data needed to register a new user
type RegistrationRequest struct {
	FirstName             string `json:"firstName" validate:"required"`
	LastName              string `json:"lastName" validate:"required"`
	Nickname              string `json:"nickname" validate:"required,alphanum"`
	Email                 string `json:"email" validate:"required,email"`
	AcceptedPrivacyPolicy bool   `json:"acceptedPrivacyPolicy" validate:"required,eq=true"`
}

// RegistrationSessionResponse represents the response for a registration session creation
type RegistrationSessionResponse struct {
	RegistrationSessionID string `json:"registrationSessionId"`
	CodeExpires           int64  `json:"codeExpires"`
	Code                  string `json:"code,omitempty"` // Only for debugging
}

// ConfirmEmailRequest represents the data needed to confirm an email
type ConfirmEmailRequest struct {
	RegistrationSessionID string `json:"registrationSessionId" validate:"required,uuid"`
	Code                  string `json:"code" validate:"required,len=4,numeric"`
}

// ResendCodeRequest represents the data needed to resend a confirmation code
type ResendCodeRequest struct {
	RegistrationSessionID string `json:"registrationSessionId" validate:"required,uuid"`
}

// LoginRequest represents the data needed to send a login code
type LoginRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// LoginSessionResponse represents the response after sending a login code
type LoginSessionResponse struct {
	CodeExpires int64  `json:"codeExpires"`
	Code        string `json:"code,omitempty"` // Only for debugging
}

// LoginConfirmRequest represents the data needed to confirm a login
type LoginConfirmRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=4,numeric"`
}

// TokenResponse represents the token pair response
type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// RefreshTokenRequest represents the data needed to refresh a token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error          string       `json:"error"`
	DetailedErrors []FieldError `json:"detailedErrors,omitempty"`
}

// FieldError represents a field-specific error
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
