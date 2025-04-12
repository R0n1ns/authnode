package repository

import (
	"context"

	"authmicro/internal/domain/entity"
)

// UserRepository defines the interface for user repository operations
type UserRepository interface {
	// User operations
	GetUserByID(ctx context.Context, id string) (*entity.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	GetUserByNickname(ctx context.Context, nickname string) (*entity.User, error)
	CreateUser(ctx context.Context, user *entity.User) error
	UpdateUser(ctx context.Context, user *entity.User) error
	DeleteUser(ctx context.Context, id string) error

	// Registration session operations
	GetRegistrationSessionByID(ctx context.Context, id string) (*entity.RegistrationSession, error)
	CreateRegistrationSession(ctx context.Context, session *entity.RegistrationSession) error
	DeleteRegistrationSession(ctx context.Context, id string) error

	// Login session operations
	GetLoginSessionByEmail(ctx context.Context, email string) (*entity.LoginSession, error)
	CreateLoginSession(ctx context.Context, session *entity.LoginSession) error
	DeleteLoginSession(ctx context.Context, email string) error

	// Refresh token operations
	SaveRefreshToken(ctx context.Context, userID, tokenID, tokenHash string, expiresAt, createdAt int64) error
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (string, string, error)
	DeleteRefreshToken(ctx context.Context, tokenID string) error
	DeleteAllRefreshTokens(ctx context.Context, userID string) error
}
