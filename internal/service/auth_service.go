package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"authmicro/configs"
	"authmicro/internal/domain"
	"authmicro/pkg/logger"
)

type userRepository interface {
	GetByID(context.Context, int64) (domain.User, error)
	GetByEmail(context.Context, string) (domain.User, error)
	NicknameExists(context.Context, string) (bool, error)
	EmailExists(context.Context, string) (bool, error)
}
type roleRepository interface {
	GetUserRoleNames(context.Context, int64) ([]string, error)
	HasRole(context.Context, int64, string) (bool, error)
}
type sessionRepository interface {
	CreateRegistrationSession(context.Context, domain.RegistrationSession) (string, error)
	GetRegistrationSession(context.Context, string) (domain.RegistrationSession, error)
	UpdateRegistrationCode(context.Context, string, string, time.Time, time.Time, time.Duration) error
	IncrementRegistrationAttempts(context.Context, string) error
	CompleteRegistration(context.Context, string, domain.User, string) error
	UpsertLoginSession(context.Context, domain.LoginSession, time.Duration) error
	GetLoginSession(context.Context, string) (domain.LoginSession, error)
	IncrementLoginAttempts(context.Context, string) error
	DeleteLoginSession(context.Context, string) error
}
type tokenIssuer interface {
	IssuePair(context.Context, domain.User, []string, string, string) (domain.TokenPair, error)
	ConsumeRefreshToken(context.Context, string) (domain.TokenSession, error)
	ValidateToken(string) (*domain.TokenClaims, error)
	RevokeAll(context.Context, int64) error
}
type mailer interface{ SendVerificationCode(string, string) error }

type AuthService struct {
	users    userRepository
	roles    roleRepository
	sessions sessionRepository
	tokens   tokenIssuer
	mailer   mailer
	otp      configs.OTPConfig
	logger   logger.Logger
	now      func() time.Time
}

func NewAuthService(u userRepository, r roleRepository, s sessionRepository, t tokenIssuer, m mailer, l logger.Logger, otp configs.OTPConfig) *AuthService {
	return &AuthService{users: u, roles: r, sessions: s, tokens: t, mailer: m, logger: l, otp: otp, now: time.Now}
}

func (s *AuthService) CreateRegistrationSession(ctx context.Context, req domain.RegistrationRequest) (*domain.RegistrationSessionResponse, []domain.FieldError, error) {
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Nickname = strings.TrimSpace(req.Nickname)
	req.Email = normalizeEmail(req.Email)
	fields := validateRegistration(req)
	if len(fields) > 0 {
		return nil, fields, nil
	}
	if exists, err := s.users.NicknameExists(ctx, req.Nickname); err != nil {
		return nil, nil, err
	} else if exists {
		fields = append(fields, domain.FieldError{Field: "nickname", Message: "nickname is already in use"})
	}
	if exists, err := s.users.EmailExists(ctx, req.Email); err != nil {
		return nil, nil, err
	} else if exists {
		fields = append(fields, domain.FieldError{Field: "email", Message: "email is already registered"})
	}
	if len(fields) > 0 {
		return nil, fields, nil
	}
	code, err := generateOTP()
	if err != nil {
		return nil, nil, err
	}
	now := s.now().UTC()
	expires := now.Add(s.otp.Expiration)
	id, err := s.sessions.CreateRegistrationSession(ctx, domain.RegistrationSession{
		FirstName: req.FirstName, LastName: req.LastName, Nickname: req.Nickname, Email: req.Email,
		AcceptedPrivacyPolicy: req.AcceptedPrivacyPolicy, CodeHash: s.hashOTP(code), CodeExpires: expires,
		LastSentAt: now, CreatedAt: now})
	if err != nil {
		return nil, nil, err
	}
	if err = s.mailer.SendVerificationCode(req.Email, code); err != nil {
		return nil, nil, fmt.Errorf("send verification email: %w", err)
	}
	return &domain.RegistrationSessionResponse{RegistrationSessionID: id, CodeExpires: expires.Unix()}, nil, nil
}

func (s *AuthService) ConfirmEmail(ctx context.Context, req domain.ConfirmEmailRequest) error {
	session, err := s.sessions.GetRegistrationSession(ctx, req.RegistrationSessionID)
	if err != nil {
		return domain.ErrInvalidCode
	}
	if session.Attempts >= s.otp.MaxAttempts || !session.CodeExpires.After(s.now().UTC()) {
		return domain.ErrInvalidCode
	}
	if !hmac.Equal([]byte(session.CodeHash), []byte(s.hashOTP(req.Code))) {
		_ = s.sessions.IncrementRegistrationAttempts(ctx, session.ID)
		return domain.ErrInvalidCode
	}
	user := domain.User{FirstName: session.FirstName, LastName: session.LastName, Nickname: session.Nickname,
		Email: session.Email, EmailVerified: true, AcceptedPrivacyPolicy: session.AcceptedPrivacyPolicy}
	if err = s.sessions.CompleteRegistration(ctx, session.ID, user, domain.DefaultRoles.User); err != nil {
		return fmt.Errorf("complete registration: %w", err)
	}
	return nil
}

func (s *AuthService) ResendVerificationCode(ctx context.Context, req domain.ResendCodeRequest) (*domain.RegistrationSessionResponse, error) {
	session, err := s.sessions.GetRegistrationSession(ctx, req.RegistrationSessionID)
	now := s.now().UTC()
	expires := now.Add(s.otp.Expiration)
	if err != nil {
		return &domain.RegistrationSessionResponse{RegistrationSessionID: req.RegistrationSessionID, CodeExpires: expires.Unix()}, nil
	}
	code, err := generateOTP()
	if err != nil {
		return nil, err
	}
	if err = s.sessions.UpdateRegistrationCode(ctx, session.ID, s.hashOTP(code), expires, now, s.otp.ResendDelay); err != nil {
		return nil, err
	}
	if err = s.mailer.SendVerificationCode(session.Email, code); err != nil {
		return nil, fmt.Errorf("send verification email: %w", err)
	}
	return &domain.RegistrationSessionResponse{RegistrationSessionID: session.ID, CodeExpires: expires.Unix()}, nil
}

func (s *AuthService) SendLoginCode(ctx context.Context, req domain.LoginRequest) (*domain.LoginSessionResponse, error) {
	email := normalizeEmail(req.Email)
	now := s.now().UTC()
	expires := now.Add(s.otp.Expiration)
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, domain.ErrInvalidInput
	}
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return &domain.LoginSessionResponse{CodeExpires: expires.Unix()}, nil
	}
	code, err := generateOTP()
	if err != nil {
		return nil, err
	}
	err = s.sessions.UpsertLoginSession(ctx, domain.LoginSession{Email: user.Email, CodeHash: s.hashOTP(code),
		CodeExpires: expires, LastSentAt: now, CreatedAt: now}, s.otp.ResendDelay)
	if err != nil {
		return nil, err
	}
	if err = s.mailer.SendVerificationCode(user.Email, code); err != nil {
		return nil, fmt.Errorf("send login email: %w", err)
	}
	return &domain.LoginSessionResponse{CodeExpires: expires.Unix()}, nil
}

func (s *AuthService) ConfirmLogin(ctx context.Context, req domain.LoginConfirmRequest, ua, ip string) (*domain.TokenResponse, error) {
	session, err := s.sessions.GetLoginSession(ctx, normalizeEmail(req.Email))
	if err != nil || session.Attempts >= s.otp.MaxAttempts || !session.CodeExpires.After(s.now().UTC()) {
		return nil, domain.ErrInvalidCode
	}
	if !hmac.Equal([]byte(session.CodeHash), []byte(s.hashOTP(req.Code))) {
		_ = s.sessions.IncrementLoginAttempts(ctx, session.ID)
		return nil, domain.ErrInvalidCode
	}
	user, err := s.users.GetByEmail(ctx, session.Email)
	if err != nil {
		return nil, domain.ErrInvalidCode
	}
	roles, err := s.roles.GetUserRoleNames(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	pair, err := s.tokens.IssuePair(ctx, user, roles, ua, ip)
	if err != nil {
		return nil, err
	}
	if err = s.sessions.DeleteLoginSession(ctx, session.ID); err != nil {
		return nil, err
	}
	return &domain.TokenResponse{AccessToken: pair.AccessToken, RefreshToken: pair.RefreshToken}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req domain.RefreshTokenRequest, ua, ip string) (*domain.TokenResponse, error) {
	old, err := s.tokens.ConsumeRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}
	if old.UserAgent != "" && ua != "" && old.UserAgent != ua {
		return nil, domain.ErrInvalidToken
	}
	user, err := s.users.GetByID(ctx, old.UserID)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}
	roles, err := s.roles.GetUserRoleNames(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	pair, err := s.tokens.IssuePair(ctx, user, roles, ua, ip)
	if err != nil {
		return nil, err
	}
	return &domain.TokenResponse{AccessToken: pair.AccessToken, RefreshToken: pair.RefreshToken}, nil
}
func (s *AuthService) Logout(ctx context.Context, refresh string) error {
	_, err := s.tokens.ConsumeRefreshToken(ctx, refresh)
	return err
}
func (s *AuthService) LogoutAll(ctx context.Context, userID int64) error {
	return s.tokens.RevokeAll(ctx, userID)
}
func (s *AuthService) GetUserByID(ctx context.Context, id int64) (domain.User, error) {
	return s.users.GetByID(ctx, id)
}
func (s *AuthService) GetUserRoles(ctx context.Context, id int64) ([]string, error) {
	return s.roles.GetUserRoleNames(ctx, id)
}
func (s *AuthService) HasRole(ctx context.Context, id int64, role string) (bool, error) {
	return s.roles.HasRole(ctx, id, role)
}

func validateRegistration(r domain.RegistrationRequest) []domain.FieldError {
	var out []domain.FieldError
	if r.FirstName == "" {
		out = append(out, domain.FieldError{Field: "firstName", Message: "field is required"})
	}
	if r.LastName == "" {
		out = append(out, domain.FieldError{Field: "lastName", Message: "field is required"})
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`).MatchString(r.Nickname) {
		out = append(out, domain.FieldError{Field: "nickname", Message: "use 3-32 letters, digits or underscores"})
	}
	if _, err := mail.ParseAddress(r.Email); err != nil {
		out = append(out, domain.FieldError{Field: "email", Message: "invalid email"})
	}
	if !r.AcceptedPrivacyPolicy {
		out = append(out, domain.FieldError{Field: "acceptedPrivacyPolicy", Message: "must be accepted"})
	}
	return out
}
func normalizeEmail(v string) string { return strings.ToLower(strings.TrimSpace(v)) }
func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
func (s *AuthService) hashOTP(code string) string {
	h := hmac.New(sha256.New, []byte(s.otp.Pepper))
	_, _ = h.Write([]byte(code))
	return hex.EncodeToString(h.Sum(nil))
}

var _ = errors.Is
