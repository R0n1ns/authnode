package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"authmicro/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type SessionRepository struct{ db *sqlx.DB }

func NewSessionRepository(db *sqlx.DB) *SessionRepository { return &SessionRepository{db: db} }

func (r *SessionRepository) CreateRegistrationSession(ctx context.Context, s domain.RegistrationSession) (string, error) {
	s.ID = uuid.NewString()
	q := `INSERT INTO registration_sessions
 (id,first_name,last_name,nickname,email,accepted_privacy_policy,code_hash,code_expires,attempts,last_sent_at,created_at)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,0,$9,$10)`
	_, err := r.db.ExecContext(ctx, q, s.ID, s.FirstName, s.LastName, s.Nickname, s.Email,
		s.AcceptedPrivacyPolicy, s.CodeHash, s.CodeExpires, s.LastSentAt, s.CreatedAt)
	return s.ID, err
}

func (r *SessionRepository) GetRegistrationSession(ctx context.Context, id string) (domain.RegistrationSession, error) {
	var s domain.RegistrationSession
	q := `SELECT id,first_name,last_name,nickname,email,accepted_privacy_policy,code_hash,
 code_expires,attempts,last_sent_at,created_at FROM registration_sessions WHERE id=$1`
	if err := r.db.GetContext(ctx, &s, q, id); err != nil {
		return s, domain.ErrNotFound
	}
	return s, nil
}

func (r *SessionRepository) UpdateRegistrationCode(ctx context.Context, id, hash string, expires, sentAt time.Time, cooldown time.Duration) error {
	q := `UPDATE registration_sessions SET code_hash=$1,code_expires=$2,attempts=0,last_sent_at=$3
 WHERE id=$4 AND last_sent_at <= $5`
	result, err := r.db.ExecContext(ctx, q, hash, expires, sentAt, id, sentAt.Add(-cooldown))
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return domain.ErrRateLimited
	}
	return nil
}
func (r *SessionRepository) IncrementRegistrationAttempts(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE registration_sessions SET attempts=attempts+1 WHERE id=$1`, id)
	return err
}

func (r *SessionRepository) CompleteRegistration(ctx context.Context, sessionID string, u domain.User, role string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	now := time.Now().UTC()
	var userID, roleID int64
	err = tx.QueryRowContext(ctx, `INSERT INTO users(first_name,last_name,nickname,email,email_verified,
 accepted_privacy_policy,created_at,updated_at) VALUES($1,$2,$3,$4,true,$5,$6,$6) RETURNING id`,
		u.FirstName, u.LastName, u.Nickname, u.Email, u.AcceptedPrivacyPolicy, now).Scan(&userID)
	if err != nil {
		return err
	}
	if err = tx.QueryRowContext(ctx, `SELECT id FROM roles WHERE name=$1`, role).Scan(&roleID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO user_roles(user_id,role_id,created_at) VALUES($1,$2,$3)`, userID, roleID, now); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM registration_sessions WHERE id=$1`, sessionID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *SessionRepository) UpsertLoginSession(ctx context.Context, s domain.LoginSession, cooldown time.Duration) error {
	s.ID = uuid.NewString()
	q := `INSERT INTO login_sessions(id,email,code_hash,code_expires,attempts,last_sent_at,created_at)
 VALUES($1,$2,$3,$4,0,$5,$6)
 ON CONFLICT(email) DO UPDATE SET id=excluded.id,code_hash=excluded.code_hash,
 code_expires=excluded.code_expires,attempts=0,last_sent_at=excluded.last_sent_at,created_at=excluded.created_at
 WHERE login_sessions.last_sent_at <= $7 RETURNING id`
	var id string
	err := r.db.QueryRowContext(ctx, q, s.ID, s.Email, s.CodeHash, s.CodeExpires, s.LastSentAt, s.CreatedAt,
		s.LastSentAt.Add(-cooldown)).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrRateLimited
	}
	return err
}
func (r *SessionRepository) GetLoginSession(ctx context.Context, email string) (domain.LoginSession, error) {
	var s domain.LoginSession
	q := `SELECT id,email,code_hash,code_expires,attempts,last_sent_at,created_at FROM login_sessions WHERE email=$1`
	if err := r.db.GetContext(ctx, &s, q, email); err != nil {
		return s, domain.ErrInvalidCode
	}
	return s, nil
}
func (r *SessionRepository) IncrementLoginAttempts(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE login_sessions SET attempts=attempts+1 WHERE id=$1`, id)
	return err
}
func (r *SessionRepository) DeleteLoginSession(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM login_sessions WHERE id=$1`, id)
	return err
}

func (r *SessionRepository) CreateTokenSession(ctx context.Context, s domain.TokenSession) error {
	q := `INSERT INTO token_sessions(id,user_id,refresh_token_hash,user_agent,ip,expires_at,created_at)
 VALUES($1,$2,$3,$4,$5,$6,$7)`
	_, err := r.db.ExecContext(ctx, q, s.ID, s.UserID, s.RefreshTokenHash, s.UserAgent, s.IP, s.ExpiresAt, s.CreatedAt)
	return err
}
func (r *SessionRepository) ConsumeTokenSession(ctx context.Context, hash string) (domain.TokenSession, error) {
	var s domain.TokenSession
	q := `DELETE FROM token_sessions WHERE refresh_token_hash=$1 AND expires_at>NOW()
 RETURNING id,user_id,refresh_token_hash,user_agent,ip,expires_at,created_at`
	if err := r.db.GetContext(ctx, &s, q, hash); err != nil {
		return s, domain.ErrInvalidToken
	}
	return s, nil
}
func (r *SessionRepository) DeleteUserTokenSessions(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM token_sessions WHERE user_id=$1`, userID)
	return err
}
