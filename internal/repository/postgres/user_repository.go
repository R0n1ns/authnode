package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"authmicro/internal/domain"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) (int64, error) {
	query := `
                INSERT INTO users (first_name, last_name, nickname, email, email_verified, accepted_privacy_policy, created_at, updated_at) 
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
                RETURNING id`

	var id int64
	now := time.Now().UTC()

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.FirstName,
		user.LastName,
		user.Nickname,
		user.Email,
		user.EmailVerified,
		user.AcceptedPrivacyPolicy,
		now,
		now,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (domain.User, error) {
	query := `
                SELECT id, first_name, last_name, nickname, email, email_verified, accepted_privacy_policy, created_at, updated_at 
                FROM users 
                WHERE id = $1`

	var user domain.User
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.User{}, errors.New("user not found")
		}
		return domain.User{}, err
	}

	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	query := `
                SELECT id, first_name, last_name, nickname, email, email_verified, accepted_privacy_policy, created_at, updated_at 
                FROM users 
                WHERE email = $1`

	var user domain.User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.User{}, errors.New("user not found")
		}
		return domain.User{}, err
	}

	return user, nil
}

func (r *UserRepository) GetByNickname(ctx context.Context, nickname string) (domain.User, error) {
	query := `
                SELECT id, first_name, last_name, nickname, email, email_verified, accepted_privacy_policy, created_at, updated_at 
                FROM users 
                WHERE nickname = $1`

	var user domain.User
	err := r.db.GetContext(ctx, &user, query, nickname)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.User{}, errors.New("user not found")
		}
		return domain.User{}, err
	}

	return user, nil
}

func (r *UserRepository) UpdateEmailVerificationStatus(ctx context.Context, userID int64, verified bool) error {
	query := `
                UPDATE users 
                SET email_verified = $1, updated_at = $2 
                WHERE id = $3`

	_, err := r.db.ExecContext(ctx, query, verified, time.Now().UTC(), userID)
	return err
}

func (r *UserRepository) NicknameExists(ctx context.Context, nickname string) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM users WHERE nickname = $1)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, nickname)
	return exists, err
}

func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, email)
	return exists, err
}
