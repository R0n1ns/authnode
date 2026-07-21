package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"authmicro/configs"
	"authmicro/internal/domain"
)

type memoryTokens struct {
	sessions map[string]domain.TokenSession
}

func newMemoryTokens() *memoryTokens {
	return &memoryTokens{sessions: map[string]domain.TokenSession{}}
}
func (m *memoryTokens) CreateTokenSession(_ context.Context, s domain.TokenSession) error {
	m.sessions[s.RefreshTokenHash] = s
	return nil
}
func (m *memoryTokens) ConsumeTokenSession(_ context.Context, hash string) (domain.TokenSession, error) {
	session, ok := m.sessions[hash]
	if !ok {
		return session, domain.ErrInvalidToken
	}
	delete(m.sessions, hash)
	return session, nil
}
func (m *memoryTokens) DeleteUserTokenSessions(_ context.Context, userID int64) error {
	for hash, session := range m.sessions {
		if session.UserID == userID {
			delete(m.sessions, hash)
		}
	}
	return nil
}
func testJWTConfig() configs.JWTConfig {
	return configs.JWTConfig{
		Secret: "01234567890123456789012345678901", Issuer: "authnode", Audience: "authnode-api",
		AccessTokenExpiration: time.Minute, RefreshTokenExpiration: time.Hour,
	}
}

func TestTokenServiceIssueValidateAndConsume(t *testing.T) {
	repo := newMemoryTokens()
	svc := NewTokenService(testJWTConfig(), repo)
	user := domain.User{ID: 42, Email: "dev@example.com", Nickname: "dev"}
	pair, err := svc.IssuePair(context.Background(), user, []string{"user"}, "agent", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := svc.ValidateToken(pair.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != 42 || claims.Email != user.Email || len(claims.Roles) != 1 {
		t.Fatalf("unexpected claims: %+v", claims)
	}
	session, err := svc.ConsumeRefreshToken(context.Background(), pair.RefreshToken)
	if err != nil || session.UserID != 42 {
		t.Fatalf("unexpected consumed session: %+v, %v", session, err)
	}
	if _, err = svc.ConsumeRefreshToken(context.Background(), pair.RefreshToken); !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("refresh replay must fail, got %v", err)
	}
}

func TestTokenServiceRejectsTamperedAndExpiredAccessTokens(t *testing.T) {
	user := domain.User{ID: 1, Email: "dev@example.com", Nickname: "dev"}
	svc := NewTokenService(testJWTConfig(), newMemoryTokens())
	pair, err := svc.IssuePair(context.Background(), user, nil, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = svc.ValidateToken(pair.AccessToken + "x"); !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("tampered token accepted: %v", err)
	}
	cfg := testJWTConfig()
	cfg.AccessTokenExpiration = -time.Minute
	expiredSvc := NewTokenService(cfg, newMemoryTokens())
	expiredPair, err := expiredSvc.IssuePair(context.Background(), user, nil, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = expiredSvc.ValidateToken(expiredPair.AccessToken); !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("expired token accepted: %v", err)
	}
}
