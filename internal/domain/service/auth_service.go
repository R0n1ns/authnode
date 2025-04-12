package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"authmicro/internal/config"
	"authmicro/internal/domain/entity"
	"authmicro/internal/domain/repository"
	"github.com/google/uuid"
)

// Error definitions for auth service
var (
	ErrInvalidInput          = errors.New("invalid input")
	ErrInvalidEmail          = errors.New("invalid email format")
	ErrInvalidNickname       = errors.New("invalid nickname format")
	ErrNicknameExists        = errors.New("nickname already exists")
	ErrEmailExists           = errors.New("email already exists")
	ErrUserNotFound          = errors.New("user not found")
	ErrSessionNotFound       = errors.New("session not found")
	ErrPrivacyPolicyRequired = errors.New("privacy policy acceptance required")
	ErrInvalidCode           = errors.New("invalid or expired verification code")
	// These errors are defined in token_service.go and imported here
	// - ErrTokenExpired: token has expired
	// - ErrTokenInvalid: token is invalid
)

// AuthService handles authentication and user management
type AuthService struct {
	userRepo     repository.UserRepository
	tokenService TokenService
	emailService EmailService
	config       *config.Config
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo repository.UserRepository,
	tokenService TokenService,
	emailService EmailService,
	config *config.Config,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		tokenService: tokenService,
		emailService: emailService,
		config:       config,
	}
}

// CreateRegistrationSession starts the registration process
func (s *AuthService) CreateRegistrationSession(
	ctx context.Context,
	firstName string,
	lastName string,
	nickname string,
	email string,
	acceptedPrivacyPolicy bool,
) (*entity.RegistrationSession, error) {
	// Validate input
	if firstName == "" || lastName == "" || nickname == "" || email == "" {
		return nil, ErrInvalidInput
	}

	// Validate email format
	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}

	// Validate nickname format
	if !isValidNickname(nickname) {
		return nil, ErrInvalidNickname
	}

	// Check if nickname already exists
	existingUser, err := s.userRepo.GetUserByNickname(ctx, nickname)
	if err == nil && existingUser != nil {
		return nil, ErrNicknameExists
	}

	// Check if privacy policy is accepted
	if !acceptedPrivacyPolicy {
		return nil, ErrPrivacyPolicyRequired
	}

	// Generate verification code
	code, err := generateCode(6)
	if err != nil {
		return nil, fmt.Errorf("failed to generate verification code: %w", err)
	}

	// Create registration session
	session := &entity.RegistrationSession{
		ID:                    uuid.New().String(),
		Email:                 email,
		FirstName:             firstName,
		LastName:              lastName,
		Nickname:              nickname,
		VerificationCode:      code,
		VerificationCodeExp:   time.Now().UTC().Add(15 * time.Minute),
		AcceptedPrivacyPolicy: acceptedPrivacyPolicy,
		CreatedAt:             time.Now().UTC(),
	}
	fmt.Println(session)
	// Save session
	if err := s.userRepo.CreateRegistrationSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create registration session: %w", err)
	}

	// Send verification email
	if err := s.sendVerificationEmail(ctx, email, code); err != nil {
		return nil, fmt.Errorf("failed to send verification email: %w", err)
	}

	return session, nil
}

// ConfirmEmail verifies the email and creates the user
func (s *AuthService) ConfirmEmail(ctx context.Context, sessionID, code string) error {
	// Get registration session
	session, err := s.userRepo.GetRegistrationSessionByID(ctx, sessionID)
	if err != nil {
		return ErrSessionNotFound
	}
	fmt.Println(session.VerificationCodeExp)
	fmt.Println(time.Now().UTC())

	// Check if code is valid
	if session.VerificationCode != code || time.Now().UTC().After(session.VerificationCodeExp) {
		return ErrInvalidCode
	}

	// Create user
	user := entity.NewUser(
		uuid.New().String(),
		session.FirstName,
		session.LastName,
		session.Nickname,
		session.Email,
		session.AcceptedPrivacyPolicy,
	)

	// Set email as verified
	user.EmailVerified = true

	// Save user
	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Delete session
	if err := s.userRepo.DeleteRegistrationSession(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to delete registration session: %w", err)
	}

	return nil
}

// ResendCodeEmail resends the verification code
func (s *AuthService) ResendCodeEmail(ctx context.Context, sessionID string) (*entity.RegistrationSession, error) {
	// Get registration session
	session, err := s.userRepo.GetRegistrationSessionByID(ctx, sessionID)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	// Generate new verification code
	code, err := generateCode(6)
	if err != nil {
		return nil, fmt.Errorf("failed to generate verification code: %w", err)
	}

	// Update session
	session.VerificationCode = code
	session.VerificationCodeExp = time.Now().UTC().Add(15 * time.Minute)

	// Update session (we need to delete and recreate since we don't have an update method)
	if err := s.userRepo.DeleteRegistrationSession(ctx, sessionID); err != nil {
		return nil, fmt.Errorf("failed to update registration session: %w", err)
	}

	if err := s.userRepo.CreateRegistrationSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to update registration session: %w", err)
	}

	// Send verification email
	if err := s.sendVerificationEmail(ctx, session.Email, code); err != nil {
		return nil, fmt.Errorf("failed to send verification email: %w", err)
	}

	return session, nil
}

// SendLoginCodeEmail starts the login process
func (s *AuthService) SendLoginCodeEmail(ctx context.Context, email string) (*entity.LoginSession, error) {
	// Validate email format
	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}

	// Check if user exists
	_, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		// Do not reveal if the user exists or not
		// Instead, return a fake session
		return &entity.LoginSession{
			Email:        email,
			LoginCode:    "",
			LoginCodeExp: time.Now().UTC().Add(15 * time.Minute),
			CreatedAt:    time.Now().UTC(),
		}, nil
	}

	// Generate login code
	code, err := generateCode(6)
	if err != nil {
		return nil, fmt.Errorf("failed to generate login code: %w", err)
	}

	// Create login session
	session := &entity.LoginSession{
		Email:        email,
		LoginCode:    code,
		LoginCodeExp: time.Now().UTC().Add(15 * time.Minute),
		CreatedAt:    time.Now().UTC(),
	}

	// Save session
	if err := s.userRepo.CreateLoginSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create login session: %w", err)
	}

	// Send login email
	subject := "Your Login Code"
	body := fmt.Sprintf("Your login code is: %s", code)
	if err := s.emailService.SendEmail(ctx, email, subject, body); err != nil {
		return nil, fmt.Errorf("failed to send login email: %w", err)
	}

	return session, nil
}

// ConfirmLogin verifies the login code and returns tokens
func (s *AuthService) ConfirmLogin(ctx context.Context, email, code string) (*entity.TokenPair, error) {
	// Get login session
	session, err := s.userRepo.GetLoginSessionByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCode
	}

	// Check if code is valid
	if session.LoginCode != code || time.Now().UTC().After(session.LoginCodeExp) {
		return nil, ErrInvalidCode
	}

	// Get user
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Update last login time
	user.LastLoginAt = time.Now().UTC()
	user.UpdatedAt = time.Now().UTC()
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Generate tokens
	accessToken, err := s.tokenService.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.tokenService.GenerateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Delete login session
	if err := s.userRepo.DeleteLoginSession(ctx, email); err != nil {
		return nil, fmt.Errorf("failed to delete login session: %w", err)
	}

	// Return tokens
	return &entity.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken refreshes the access token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*entity.TokenPair, error) {
	// Validate refresh token
	claims, err := s.tokenService.ValidateRefreshToken(refreshToken)
	if err != nil {
		if errors.Is(err, ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	// Get user
	user, err := s.userRepo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Generate new tokens
	newAccessToken, err := s.tokenService.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.tokenService.GenerateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Return tokens
	return &entity.TokenPair{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// sendVerificationEmail sends a verification email
func (s *AuthService) sendVerificationEmail(ctx context.Context, email, code string) error {
	subject := "Verify Your Email"
	body := fmt.Sprintf("Your verification code is: %s", code)
	return s.emailService.SendEmail(ctx, email, subject, body)
}

// Helper functions

// isValidEmail validates email format
func isValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(email)
}

// isValidNickname validates nickname format
func isValidNickname(nickname string) bool {
	// Min length 3, max length 30, alphanumeric and underscores only
	pattern := `^[a-zA-Z0-9_]{3,30}$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(nickname)
}

// generateCode generates a random code of specified length
func generateCode(length int) (string, error) {
	const charset = "0123456789"
	code := strings.Builder{}
	code.Grow(length)

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		code.WriteByte(charset[n.Int64()])
	}

	return code.String(), nil
}
