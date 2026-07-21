package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"authmicro/configs"
	"authmicro/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type tokenSessionRepository interface {
	CreateTokenSession(context.Context, domain.TokenSession) error
	ConsumeTokenSession(context.Context, string) (domain.TokenSession, error)
	DeleteUserTokenSessions(context.Context, int64) error
}

type TokenService struct {
	config configs.JWTConfig
	repo   tokenSessionRepository
	now    func() time.Time
}

type accessClaims struct {
	Email     string   `json:"email"`
	Nickname  string   `json:"nickname"`
	Roles     []string `json:"roles"`
	TokenType string   `json:"token_type"`
	jwt.RegisteredClaims
}

func NewTokenService(c configs.JWTConfig, r tokenSessionRepository) *TokenService {
	return &TokenService{config: c, repo: r, now: time.Now}
}

func (s *TokenService) IssuePair(ctx context.Context, u domain.User, roles []string, ua, ip string) (domain.TokenPair, error) {
	access, err := s.accessToken(u, roles)
	if err != nil {
		return domain.TokenPair{}, err
	}
	refresh, err := randomToken()
	if err != nil {
		return domain.TokenPair{}, err
	}
	now := s.now().UTC()
	session := domain.TokenSession{ID: uuid.NewString(), UserID: u.ID, RefreshTokenHash: hashToken(refresh),
		UserAgent: ua, IP: ip, CreatedAt: now, ExpiresAt: now.Add(s.config.RefreshTokenExpiration)}
	if err = s.repo.CreateTokenSession(ctx, session); err != nil {
		return domain.TokenPair{}, fmt.Errorf("store refresh session: %w", err)
	}
	return domain.TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

func (s *TokenService) ConsumeRefreshToken(ctx context.Context, raw string) (domain.TokenSession, error) {
	if raw == "" {
		return domain.TokenSession{}, domain.ErrInvalidToken
	}
	session, err := s.repo.ConsumeTokenSession(ctx, hashToken(raw))
	if err != nil || !session.ExpiresAt.After(s.now().UTC()) {
		return domain.TokenSession{}, domain.ErrInvalidToken
	}
	return session, nil
}

func (s *TokenService) RevokeAll(ctx context.Context, userID int64) error {
	return s.repo.DeleteUserTokenSessions(ctx, userID)
}

func (s *TokenService) ValidateToken(raw string) (*domain.TokenClaims, error) {
	c := new(accessClaims)
	token, err := jwt.ParseWithClaims(raw, c, func(*jwt.Token) (any, error) { return []byte(s.config.Secret), nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithIssuer(s.config.Issuer),
		jwt.WithAudience(s.config.Audience), jwt.WithExpirationRequired())
	if err != nil || !token.Valid || c.TokenType != "access" {
		return nil, domain.ErrInvalidToken
	}
	var userID int64
	if _, err = fmt.Sscan(c.Subject, &userID); err != nil || userID < 1 {
		return nil, domain.ErrInvalidToken
	}
	return &domain.TokenClaims{UserID: userID, Email: c.Email, Nickname: c.Nickname, Roles: c.Roles,
		ExpiresAt: c.ExpiresAt.Unix(), IssuedAt: c.IssuedAt.Unix(), TokenID: c.ID}, nil
}

func (s *TokenService) accessToken(u domain.User, roles []string) (string, error) {
	now := s.now().UTC()
	c := accessClaims{Email: u.Email, Nickname: u.Nickname, Roles: roles, TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{Issuer: s.config.Issuer, Subject: fmt.Sprint(u.ID),
			Audience: jwt.ClaimStrings{s.config.Audience}, ID: uuid.NewString(), IssuedAt: jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(s.config.AccessTokenExpiration))}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString([]byte(s.config.Secret))
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
