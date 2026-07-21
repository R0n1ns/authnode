package domain

import "time"

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

type RegistrationRequest struct {
	FirstName             string `json:"firstName"`
	LastName              string `json:"lastName"`
	Nickname              string `json:"nickname"`
	Email                 string `json:"email"`
	AcceptedPrivacyPolicy bool   `json:"acceptedPrivacyPolicy"`
}
type RegistrationSessionResponse struct {
	RegistrationSessionID string `json:"registrationSessionId"`
	CodeExpires           int64  `json:"codeExpires"`
}
type ConfirmEmailRequest struct {
	RegistrationSessionID string `json:"registrationSessionId"`
	Code                  string `json:"code"`
}
type ResendCodeRequest struct {
	RegistrationSessionID string `json:"registrationSessionId"`
}
type LoginRequest struct {
	Email string `json:"email"`
}
type LoginSessionResponse struct {
	CodeExpires int64 `json:"codeExpires"`
}
type LoginConfirmRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}
type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}
type ErrorResponse struct {
	Error          string       `json:"error"`
	DetailedErrors []FieldError `json:"detailedErrors,omitempty"`
}
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
