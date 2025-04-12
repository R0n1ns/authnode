package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"authmicro/internal/domain"
)

type SessionRepository struct {
	db *sqlx.DB
}

func NewSessionRepository(db *sqlx.DB) *SessionRepository {
	return &SessionRepository{
		db: db,
	}
}

// CreateRegistrationSession creates a new registration session
func (r *SessionRepository) CreateRegistrationSession(ctx context.Context, session domain.RegistrationSession) (string, error) {
	query := `
                INSERT INTO registration_sessions (id, first_name, last_name, nickname, email, accepted_privacy_policy, code, code_expires, created_at)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
                RETURNING id`

	id := uuid.New().String()
	now := time.Now().UTC()

	err := r.db.QueryRowContext(
		ctx,
		query,
		id,
		session.FirstName,
		session.LastName,
		session.Nickname,
		session.Email,
		session.AcceptedPrivacyPolicy,
		session.Code,
		session.CodeExpires,
		now,
	).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}

// GetRegistrationSession retrieves a registration session by ID
func (r *SessionRepository) GetRegistrationSession(ctx context.Context, id string) (domain.RegistrationSession, error) {
	query := `
                SELECT id, first_name, last_name, nickname, email, accepted_privacy_policy, code, code_expires, created_at
                FROM registration_sessions
                WHERE id = $1`

	var session domain.RegistrationSession
	err := r.db.GetContext(ctx, &session, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.RegistrationSession{}, errors.New("registration session not found")
		}
		return domain.RegistrationSession{}, err
	}

	return session, nil
}

// UpdateRegistrationSessionCode updates the verification code for a registration session
func (r *SessionRepository) UpdateRegistrationSessionCode(ctx context.Context, id, code string, expires time.Time) error {
	query := `
                UPDATE registration_sessions
                SET code = $1, code_expires = $2
                WHERE id = $3`

	_, err := r.db.ExecContext(ctx, query, code, expires, id)
	return err
}

// DeleteRegistrationSession deletes a registration session
func (r *SessionRepository) DeleteRegistrationSession(ctx context.Context, id string) error {
	query := `DELETE FROM registration_sessions WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// CreateLoginSession creates a new login session
func (r *SessionRepository) CreateLoginSession(ctx context.Context, session domain.LoginSession) (string, error) {
	query := `
                INSERT INTO login_sessions (id, email, code, code_expires, created_at)
                VALUES ($1, $2, $3, $4, $5)
                RETURNING id`

	id := uuid.New().String()
	now := time.Now().UTC()

	err := r.db.QueryRowContext(
		ctx,
		query,
		id,
		session.Email,
		session.Code,
		session.CodeExpires,
		now,
	).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}

// GetLoginSessionByEmailAndCode retrieves a login session by email and code
func (r *SessionRepository) GetLoginSessionByEmailAndCode(ctx context.Context, email, code string) (domain.LoginSession, error) {
	query := `
                SELECT id, email, code, code_expires, created_at
                FROM login_sessions
                WHERE email = $1 AND code = $2 AND code_expires > NOW()`

	var session domain.LoginSession
	err := r.db.GetContext(ctx, &session, query, email, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.LoginSession{}, errors.New("invalid or expired code")
		}
		return domain.LoginSession{}, err
	}

	return session, nil
}

// DeleteLoginSession deletes a login session
func (r *SessionRepository) DeleteLoginSession(ctx context.Context, id string) error {
	query := `DELETE FROM login_sessions WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// DeleteExpiredLoginSessions deletes all expired login sessions
func (r *SessionRepository) DeleteExpiredLoginSessions(ctx context.Context) error {
	query := `DELETE FROM login_sessions WHERE code_expires < NOW()`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

// CreateTokenSession creates a new token session
func (r *SessionRepository) CreateTokenSession(ctx context.Context, session domain.TokenSession) error {
	query := `
                INSERT INTO token_sessions (id, user_id, refresh_token, user_agent, ip, expires_at, created_at)
                VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(
		ctx,
		query,
		session.ID,
		session.UserID,
		session.RefreshToken,
		session.UserAgent,
		session.IP,
		session.ExpiresAt,
		session.CreatedAt,
	)
	return err
}

// GetTokenSession retrieves a token session by refresh token
func (r *SessionRepository) GetTokenSession(ctx context.Context, refreshToken string) (domain.TokenSession, error) {
	query := `
                SELECT id, user_id, refresh_token, user_agent, ip, expires_at, created_at
                FROM token_sessions
                WHERE refresh_token = $1 AND expires_at > NOW()`

	var session domain.TokenSession
	err := r.db.GetContext(ctx, &session, query, refreshToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.TokenSession{}, errors.New("token expires")
		}
		return domain.TokenSession{}, err
	}

	return session, nil
}

// DeleteTokenSession deletes a token session
func (r *SessionRepository) DeleteTokenSession(ctx context.Context, id string) error {
	query := `DELETE FROM token_sessions WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// DeleteExpiredTokenSessions deletes all expired token sessions
func (r *SessionRepository) DeleteExpiredTokenSessions(ctx context.Context) error {
	query := `DELETE FROM token_sessions WHERE expires_at < NOW()`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

// DeleteUserTokenSessions deletes all token sessions for a user
func (r *SessionRepository) DeleteUserTokenSessions(ctx context.Context, userID int64) error {
	query := `DELETE FROM token_sessions WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
