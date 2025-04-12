package database

import (
	"database/sql"
	"fmt"
	"os"

	"authmicro/internal/config"
	_ "github.com/lib/pq"
)

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(config config.DatabaseConfig) (*sql.DB, error) {
	var connStr string

	// Use DATABASE_URL if available (recommended for Replit)
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL != "" {
		connStr = databaseURL
	} else {
		// Fallback to manual configuration
		connStr = fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			config.Host,
			config.Port,
			config.User,
			config.Password,
			config.DBName,
		)
	}

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Check connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// InitSchema initializes the database schema
func InitSchema(db *sql.DB) error {
	// Create schema
	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// schema defines the database schema
const schema = `
-- Users table
CREATE TABLE IF NOT EXISTS users (
    id                 VARCHAR(36) PRIMARY KEY,
    first_name         VARCHAR(100) NOT NULL,
    last_name          VARCHAR(100) NOT NULL,
    nickname           VARCHAR(100) UNIQUE NOT NULL,
    email              VARCHAR(255) UNIQUE NOT NULL,
    email_verified     BOOLEAN NOT NULL DEFAULT FALSE,
    role               VARCHAR(20) NOT NULL DEFAULT 'user',
    created_at         TIMESTAMP NOT NULL,
    updated_at         TIMESTAMP NOT NULL,
    last_login_at      TIMESTAMP,
    password_reset     BOOLEAN NOT NULL DEFAULT FALSE,
    accepted_terms_at  TIMESTAMP,
    accepted_privacy_at TIMESTAMP NOT NULL
);

-- Index on email
CREATE INDEX IF NOT EXISTS users_email_idx ON users(email);

-- Index on nickname
CREATE INDEX IF NOT EXISTS users_nickname_idx ON users(nickname);

-- Registration sessions table
CREATE TABLE IF NOT EXISTS registration_sessions (
    id                    VARCHAR(36) PRIMARY KEY,
    email                 VARCHAR(255) NOT NULL,
    first_name            VARCHAR(100) NOT NULL,
    last_name             VARCHAR(100) NOT NULL,
    nickname              VARCHAR(100) NOT NULL,
    verification_code     VARCHAR(6) NOT NULL,
    verification_code_exp TIMESTAMP NOT NULL,
    accepted_privacy_policy BOOLEAN NOT NULL,
    created_at            TIMESTAMP NOT NULL
);

-- Index on email
CREATE INDEX IF NOT EXISTS registration_sessions_email_idx ON registration_sessions(email);

-- Login sessions table
CREATE TABLE IF NOT EXISTS login_sessions (
    id            VARCHAR(36) PRIMARY KEY,
    email         VARCHAR(255) NOT NULL,
    login_code    VARCHAR(6) NOT NULL,
    login_code_exp TIMESTAMP NOT NULL,
    created_at    TIMESTAMP NOT NULL
);

-- Index on email
CREATE INDEX IF NOT EXISTS login_sessions_email_idx ON login_sessions(email);

-- Refresh tokens table
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         VARCHAR(36) PRIMARY KEY,
    user_id    VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL
);

-- Index on user_id
CREATE INDEX IF NOT EXISTS refresh_tokens_user_id_idx ON refresh_tokens(user_id);
`
