package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"authmicro/configs"
	"authmicro/internal/domain"
)

type TokenService struct {
	config      configs.JWTConfig
	sessionRepo sessionRepository
}

func NewTokenService(config configs.JWTConfig, sessionRepo sessionRepository) *TokenService {
	return &TokenService{
		config:      config,
		sessionRepo: sessionRepo,
	}
}

// GenerateTokenPair generates a new access and refresh token pair
func (s *TokenService) GenerateTokenPair(ctx context.Context, user domain.User, roles []string) (domain.TokenPair, error) {
	// Generate access token
	accessToken, err := s.generateAccessToken(user, roles)
	if err != nil {
		return domain.TokenPair{}, err
	}

	// Generate refresh token
	refreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return domain.TokenPair{}, err
	}

	return domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *TokenService) ValidateToken(tokenString string) (*domain.TokenClaims, error) {
	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.config.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	// Check if token is valid
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Extract user data from claims
	userID, ok := claims["userId"].(float64)
	if !ok {
		return nil, errors.New("invalid userId claim")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return nil, errors.New("invalid email claim")
	}

	nickname, ok := claims["nickname"].(string)
	if !ok {
		return nil, errors.New("invalid nickname claim")
	}

	// Extract roles from claims
	rolesInterface, ok := claims["roles"].([]interface{})
	if !ok {
		return nil, errors.New("invalid roles claim")
	}

	roles := make([]string, len(rolesInterface))
	for i, role := range rolesInterface {
		roles[i], ok = role.(string)
		if !ok {
			return nil, errors.New("invalid role in roles claim")
		}
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, errors.New("invalid exp claim")
	}

	iat, ok := claims["iat"].(float64)
	if !ok {
		return nil, errors.New("invalid iat claim")
	}

	// Return token claims
	return &domain.TokenClaims{
		UserID:    int64(userID),
		Email:     email,
		Nickname:  nickname,
		Roles:     roles,
		ExpiresAt: int64(exp),
		IssuedAt:  int64(iat),
	}, nil
}

// StoreRefreshToken stores a refresh token in the database
func (s *TokenService) StoreRefreshToken(ctx context.Context, userID int64, refreshToken, userAgent, ip string) error {
	session := domain.TokenSession{
		ID:           uuid.New().String(),
		UserID:       userID,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		IP:           ip,
		ExpiresAt:    time.Now().UTC().Add(s.config.RefreshTokenExpiration),
		CreatedAt:    time.Now().UTC(),
	}

	return s.sessionRepo.CreateTokenSession(ctx, session)
}

// RevokeRefreshToken revokes a refresh token
func (s *TokenService) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	// Get token session
	session, err := s.sessionRepo.GetTokenSession(ctx, refreshToken)
	if err != nil {
		return err
	}

	// Delete token session
	return s.sessionRepo.DeleteTokenSession(ctx, session.ID)
}

// RevokeAllUserTokens revokes all tokens for a user
func (s *TokenService) RevokeAllUserTokens(ctx context.Context, userID int64) error {
	return s.sessionRepo.DeleteUserTokenSessions(ctx, userID)
}

// generateAccessToken generates a new access token
func (s *TokenService) generateAccessToken(user domain.User, roles []string) (string, error) {
	// Set token expiration time
	expirationTime := time.Now().UTC().Add(s.config.AccessTokenExpiration)

	// Create claims
	claims := jwt.MapClaims{
		"userId":   user.ID,
		"email":    user.Email,
		"nickname": user.Nickname,
		"roles":    roles,
		"exp":      expirationTime.Unix(),
		"iat":      time.Now().UTC().Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(s.config.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// generateRefreshToken generates a new refresh token
func (s *TokenService) generateRefreshToken(userID int64) (string, error) {
	// Set token expiration time
	expirationTime := time.Now().UTC().Add(s.config.RefreshTokenExpiration)

	// Create claims
	claims := jwt.MapClaims{
		"userId": userID,
		"exp":    expirationTime.Unix(),
		"iat":    time.Now().UTC().Unix(),
		"jti":    uuid.New().String(), // JWT ID for the token
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(s.config.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
