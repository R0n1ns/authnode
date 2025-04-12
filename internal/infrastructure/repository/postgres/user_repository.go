package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"authmicro/internal/domain/entity"
	"authmicro/internal/domain/repository"
)

// UserRepository implements the user repository interface using PostgreSQL
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &UserRepository{
		db: db,
	}
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
	query := `
		SELECT id, first_name, last_name, nickname, email, email_verified, role, 
		       created_at, updated_at, last_login_at, password_reset, 
		       accepted_terms_at, accepted_privacy_at
		FROM users
		WHERE id = $1
	`

	var user entity.User
	var roleStr string
	var lastLoginAt, acceptedTermsAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Nickname,
		&user.Email,
		&user.EmailVerified,
		&roleStr,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
		&user.PasswordReset,
		&acceptedTermsAt,
		&user.AcceptedPrivacyAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	user.Role = entity.Role(roleStr)

	if lastLoginAt.Valid {
		user.LastLoginAt = lastLoginAt.Time
	}

	if acceptedTermsAt.Valid {
		user.AcceptedTermsAt = acceptedTermsAt.Time
	}

	user.ActiveRefreshTokens = make(map[string]time.Time)

	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, first_name, last_name, nickname, email, email_verified, role, 
		       created_at, updated_at, last_login_at, password_reset, 
		       accepted_terms_at, accepted_privacy_at
		FROM users
		WHERE email = $1
	`

	var user entity.User
	var roleStr string
	var lastLoginAt, acceptedTermsAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Nickname,
		&user.Email,
		&user.EmailVerified,
		&roleStr,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
		&user.PasswordReset,
		&acceptedTermsAt,
		&user.AcceptedPrivacyAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	user.Role = entity.Role(roleStr)

	if lastLoginAt.Valid {
		user.LastLoginAt = lastLoginAt.Time
	}

	if acceptedTermsAt.Valid {
		user.AcceptedTermsAt = acceptedTermsAt.Time
	}

	user.ActiveRefreshTokens = make(map[string]time.Time)

	return &user, nil
}

// GetUserByNickname retrieves a user by nickname
func (r *UserRepository) GetUserByNickname(ctx context.Context, nickname string) (*entity.User, error) {
	query := `
		SELECT id, first_name, last_name, nickname, email, email_verified, role, 
		       created_at, updated_at, last_login_at, password_reset, 
		       accepted_terms_at, accepted_privacy_at
		FROM users
		WHERE nickname = $1
	`

	var user entity.User
	var roleStr string
	var lastLoginAt, acceptedTermsAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, nickname).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Nickname,
		&user.Email,
		&user.EmailVerified,
		&roleStr,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
		&user.PasswordReset,
		&acceptedTermsAt,
		&user.AcceptedPrivacyAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user by nickname: %w", err)
	}

	user.Role = entity.Role(roleStr)

	if lastLoginAt.Valid {
		user.LastLoginAt = lastLoginAt.Time
	}

	if acceptedTermsAt.Valid {
		user.AcceptedTermsAt = acceptedTermsAt.Time
	}

	user.ActiveRefreshTokens = make(map[string]time.Time)

	return &user, nil
}

// CreateUser creates a new user
func (r *UserRepository) CreateUser(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (
			id, first_name, last_name, nickname, email, email_verified, role, 
			created_at, updated_at, password_reset, accepted_privacy_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.FirstName,
		user.LastName,
		user.Nickname,
		user.Email,
		user.EmailVerified,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
		user.PasswordReset,
		user.AcceptedPrivacyAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// UpdateUser updates an existing user
func (r *UserRepository) UpdateUser(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET first_name = $1, last_name = $2, nickname = $3, email = $4, 
		    email_verified = $5, role = $6, updated_at = $7, last_login_at = $8, 
		    password_reset = $9, accepted_terms_at = $10, accepted_privacy_at = $11
		WHERE id = $12
	`

	var lastLoginAt, acceptedTermsAt sql.NullTime

	if !user.LastLoginAt.IsZero() {
		lastLoginAt = sql.NullTime{Time: user.LastLoginAt, Valid: true}
	}

	if !user.AcceptedTermsAt.IsZero() {
		acceptedTermsAt = sql.NullTime{Time: user.AcceptedTermsAt, Valid: true}
	}

	_, err := r.db.ExecContext(ctx, query,
		user.FirstName,
		user.LastName,
		user.Nickname,
		user.Email,
		user.EmailVerified,
		user.Role,
		user.UpdatedAt,
		lastLoginAt,
		user.PasswordReset,
		acceptedTermsAt,
		user.AcceptedPrivacyAt,
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// DeleteUser deletes a user
func (r *UserRepository) DeleteUser(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// GetRegistrationSessionByID retrieves a registration session by ID
func (r *UserRepository) GetRegistrationSessionByID(ctx context.Context, id string) (*entity.RegistrationSession, error) {
	query := `
		SELECT id, email, first_name, last_name, nickname, verification_code, 
		       verification_code_exp, accepted_privacy_policy, created_at
		FROM registration_sessions
		WHERE id = $1
	`

	var session entity.RegistrationSession

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID,
		&session.Email,
		&session.FirstName,
		&session.LastName,
		&session.Nickname,
		&session.VerificationCode,
		&session.VerificationCodeExp,
		&session.AcceptedPrivacyPolicy,
		&session.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("registration session not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get registration session: %w", err)
	}

	return &session, nil
}

// CreateRegistrationSession creates a new registration session
func (r *UserRepository) CreateRegistrationSession(ctx context.Context, session *entity.RegistrationSession) error {
	query := `
		INSERT INTO registration_sessions (
			id, email, first_name, last_name, nickname, verification_code, 
			verification_code_exp, accepted_privacy_policy, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.Email,
		session.FirstName,
		session.LastName,
		session.Nickname,
		session.VerificationCode,
		session.VerificationCodeExp,
		session.AcceptedPrivacyPolicy,
		session.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create registration session: %w", err)
	}

	return nil
}

// DeleteRegistrationSession deletes a registration session
func (r *UserRepository) DeleteRegistrationSession(ctx context.Context, id string) error {
	query := `DELETE FROM registration_sessions WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete registration session: %w", err)
	}

	return nil
}

// GetLoginSessionByEmail retrieves a login session by email
func (r *UserRepository) GetLoginSessionByEmail(ctx context.Context, email string) (*entity.LoginSession, error) {
	query := `
		SELECT email, login_code, login_code_exp, created_at
		FROM login_sessions
		WHERE email = $1
	`

	var session entity.LoginSession

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&session.Email,
		&session.LoginCode,
		&session.LoginCodeExp,
		&session.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("login session not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get login session: %w", err)
	}

	return &session, nil
}

// CreateLoginSession creates a new login session
func (r *UserRepository) CreateLoginSession(ctx context.Context, session *entity.LoginSession) error {
	// Delete any existing login session for this email
	_, err := r.db.ExecContext(ctx, "DELETE FROM login_sessions WHERE email = $1", session.Email)
	if err != nil {
		return fmt.Errorf("failed to delete existing login session: %w", err)
	}

	query := `
		INSERT INTO login_sessions (
			id, email, login_code, login_code_exp, created_at
		) VALUES (
			uuid_generate_v4(), $1, $2, $3, $4
		)
	`

	_, err = r.db.ExecContext(ctx, query,
		session.Email,
		session.LoginCode,
		session.LoginCodeExp,
		session.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create login session: %w", err)
	}

	return nil
}

// DeleteLoginSession deletes a login session
func (r *UserRepository) DeleteLoginSession(ctx context.Context, email string) error {
	query := `DELETE FROM login_sessions WHERE email = $1`

	_, err := r.db.ExecContext(ctx, query, email)
	if err != nil {
		return fmt.Errorf("failed to delete login session: %w", err)
	}

	return nil
}

// SaveRefreshToken saves a refresh token
func (r *UserRepository) SaveRefreshToken(ctx context.Context, userID, tokenID, tokenHash string, expiresAt, createdAt int64) error {
	query := `
		INSERT INTO refresh_tokens (
			id, user_id, token_hash, expires_at, created_at
		) VALUES (
			$1, $2, $3, to_timestamp($4), to_timestamp($5)
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		tokenID,
		userID,
		tokenHash,
		expiresAt,
		createdAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save refresh token: %w", err)
	}

	return nil
}

// GetRefreshTokenByHash retrieves a refresh token by hash
func (r *UserRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (string, string, error) {
	query := `
		SELECT id, user_id
		FROM refresh_tokens
		WHERE token_hash = $1 AND expires_at > NOW()
	`

	var tokenID, userID string

	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(&tokenID, &userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", fmt.Errorf("refresh token not found or expired: %w", err)
		}
		return "", "", fmt.Errorf("failed to get refresh token: %w", err)
	}

	return tokenID, userID, nil
}

// DeleteRefreshToken deletes a refresh token
func (r *UserRepository) DeleteRefreshToken(ctx context.Context, tokenID string) error {
	query := `DELETE FROM refresh_tokens WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}

// DeleteAllRefreshTokens deletes all refresh tokens for a user
func (r *UserRepository) DeleteAllRefreshTokens(ctx context.Context, userID string) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete all refresh tokens: %w", err)
	}

	return nil
}
