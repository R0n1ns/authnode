package domain

import "time"

type RegistrationSession struct {
	ID                    string    `db:"id"`
	FirstName             string    `db:"first_name"`
	LastName              string    `db:"last_name"`
	Nickname              string    `db:"nickname"`
	Email                 string    `db:"email"`
	AcceptedPrivacyPolicy bool      `db:"accepted_privacy_policy"`
	CodeHash              string    `db:"code_hash"`
	CodeExpires           time.Time `db:"code_expires"`
	Attempts              int       `db:"attempts"`
	LastSentAt            time.Time `db:"last_sent_at"`
	CreatedAt             time.Time `db:"created_at"`
}

type LoginSession struct {
	ID          string    `db:"id"`
	Email       string    `db:"email"`
	CodeHash    string    `db:"code_hash"`
	CodeExpires time.Time `db:"code_expires"`
	Attempts    int       `db:"attempts"`
	LastSentAt  time.Time `db:"last_sent_at"`
	CreatedAt   time.Time `db:"created_at"`
}

type TokenSession struct {
	ID               string    `db:"id"`
	UserID           int64     `db:"user_id"`
	RefreshTokenHash string    `db:"refresh_token_hash"`
	UserAgent        string    `db:"user_agent"`
	IP               string    `db:"ip"`
	ExpiresAt        time.Time `db:"expires_at"`
	CreatedAt        time.Time `db:"created_at"`
}
