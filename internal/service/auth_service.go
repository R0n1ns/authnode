package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"authmicro/internal/domain"
	"authmicro/pkg/logger"
)

type userRepository interface {
	Create(ctx context.Context, user domain.User) (int64, error)
	GetByID(ctx context.Context, id int64) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	GetByNickname(ctx context.Context, nickname string) (domain.User, error)
	UpdateEmailVerificationStatus(ctx context.Context, userID int64, verified bool) error
	NicknameExists(ctx context.Context, nickname string) (bool, error)
	EmailExists(ctx context.Context, email string) (bool, error)
}

type roleRepository interface {
	CreateRole(ctx context.Context, name string) (int64, error)
	GetRoleByName(ctx context.Context, name string) (domain.Role, error)
	GetRoleByID(ctx context.Context, id int64) (domain.Role, error)
	AssignRoleToUser(ctx context.Context, userID, roleID int64) error
	GetUserRoles(ctx context.Context, userID int64) ([]domain.Role, error)
	GetUserRoleNames(ctx context.Context, userID int64) ([]string, error)
	RemoveRoleFromUser(ctx context.Context, userID, roleID int64) error
	HasRole(ctx context.Context, userID int64, roleName string) (bool, error)
}

type sessionRepository interface {
	CreateRegistrationSession(ctx context.Context, session domain.RegistrationSession) (string, error)
	GetRegistrationSession(ctx context.Context, id string) (domain.RegistrationSession, error)
	UpdateRegistrationSessionCode(ctx context.Context, id, code string, expires time.Time) error
	DeleteRegistrationSession(ctx context.Context, id string) error
	CreateLoginSession(ctx context.Context, session domain.LoginSession) (string, error)
	GetLoginSessionByEmailAndCode(ctx context.Context, email, code string) (domain.LoginSession, error)
	DeleteLoginSession(ctx context.Context, id string) error
	DeleteExpiredLoginSessions(ctx context.Context) error
	CreateTokenSession(ctx context.Context, session domain.TokenSession) error
	GetTokenSession(ctx context.Context, refreshToken string) (domain.TokenSession, error)
	DeleteTokenSession(ctx context.Context, id string) error
	DeleteExpiredTokenSessions(ctx context.Context) error
	DeleteUserTokenSessions(ctx context.Context, userID int64) error
}

type tokenService interface {
	GenerateTokenPair(ctx context.Context, user domain.User, roles []string) (domain.TokenPair, error)
	ValidateToken(token string) (*domain.TokenClaims, error)
	StoreRefreshToken(ctx context.Context, userID int64, refreshToken, userAgent, ip string) error
	RevokeRefreshToken(ctx context.Context, refreshToken string) error
}

type emailService interface {
	SendVerificationCode(to, code string) error
}

type AuthService struct {
	userRepo    userRepository
	roleRepo    roleRepository
	sessionRepo sessionRepository
	tokenSvc    tokenService
	emailSvc    emailService
	logger      logger.Logger
}

func NewAuthService(
	userRepo userRepository,
	roleRepo roleRepository,
	sessionRepo sessionRepository,
	tokenSvc tokenService,
	emailSvc emailService,
	logger logger.Logger,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		sessionRepo: sessionRepo,
		tokenSvc:    tokenSvc,
		emailSvc:    emailSvc,
		logger:      logger,
	}
}

// CreateRegistrationSession creates a new registration session
func (s *AuthService) CreateRegistrationSession(ctx context.Context, req domain.RegistrationRequest) (*domain.RegistrationSessionResponse, []domain.FieldError, error) {
	var fieldErrors []domain.FieldError

	// Validate firstName
	if req.FirstName == "" {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:   "firstName",
			Message: "Поле пустое",
		})
	}

	// Validate lastName
	if req.LastName == "" {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:   "lastName",
			Message: "Поле пустое",
		})
	}

	// Validate nickname
	if req.Nickname == "" {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:   "nickname",
			Message: "Поле пустое",
		})
	} else {
		// Check if nickname contains only allowed characters
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9]+$`, req.Nickname)
		if !matched {
			fieldErrors = append(fieldErrors, domain.FieldError{
				Field:   "nickname",
				Message: "В nickname используются запрещённые символы",
			})
		} else {
			// Check if nickname is unique
			exists, err := s.userRepo.NicknameExists(ctx, req.Nickname)
			if err != nil {
				s.logger.Errorf("Error checking nickname existence: %v", err)
				return nil, nil, err
			}
			if exists {
				fieldErrors = append(fieldErrors, domain.FieldError{
					Field:   "nickname",
					Message: "Такой nickname уже существует",
				})
			}
		}
	}

	// Validate email
	if req.Email == "" {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:   "email",
			Message: "Поле пустое",
		})
	} else {
		// Check if email is valid
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, req.Email)
		if !matched {
			fieldErrors = append(fieldErrors, domain.FieldError{
				Field:   "email",
				Message: "Введенная строка не является электронной почтой",
			})
		}
	}

	// Validate acceptedPrivacyPolicy
	if !req.AcceptedPrivacyPolicy {
		fieldErrors = append(fieldErrors, domain.FieldError{
			Field:   "acceptedPrivacyPolicy",
			Message: "Не принято пользовательское соглашение",
		})
	}

	// Return errors if any
	if len(fieldErrors) > 0 {
		return nil, fieldErrors, nil
	}

	// Generate verification code
	code := generateCode()
	codeExpires := time.Now().UTC().Add(15 * time.Minute)

	// Create registration session
	session := domain.RegistrationSession{
		FirstName:             req.FirstName,
		LastName:              req.LastName,
		Nickname:              req.Nickname,
		Email:                 req.Email,
		AcceptedPrivacyPolicy: req.AcceptedPrivacyPolicy,
		Code:                  code,
		CodeExpires:           codeExpires,
		CreatedAt:             time.Now().UTC(),
	}

	sessionID, err := s.sessionRepo.CreateRegistrationSession(ctx, session)
	if err != nil {
		s.logger.Errorf("Error creating registration session: %v", err)
		return nil, nil, err
	}

	// Send verification code
	err = s.emailSvc.SendVerificationCode(req.Email, code)
	if err != nil {
		s.logger.Errorf("Error sending verification code: %v", err)
		// We don't want to fail the registration process if email sending fails
		// Just log the error and continue
	}

	return &domain.RegistrationSessionResponse{
		RegistrationSessionID: sessionID,
		CodeExpires:           codeExpires.Unix(),
		Code:                  code, // Only for debugging
	}, nil, nil
}

// ConfirmEmail confirms a user's email during registration
func (s *AuthService) ConfirmEmail(ctx context.Context, req domain.ConfirmEmailRequest) error {
	// Get registration session
	session, err := s.sessionRepo.GetRegistrationSession(ctx, req.RegistrationSessionID)
	if err != nil {
		// Don't expose that the session doesn't exist
		return errors.New("Неверный или истекший код подтверждения. Пожалуйста, запросите новый код и попробуйте снова")
	}

	// Check if code is valid and not expired
	if session.Code != req.Code || time.Now().UTC().After(session.CodeExpires) {
		fmt.Println(session)
		return errors.New("Неверный или истекший код подтверждения. Пожалуйста, запросите новый код и попробуйте снова")
	}

	// Create new user
	user := domain.User{
		FirstName:             session.FirstName,
		LastName:              session.LastName,
		Nickname:              session.Nickname,
		Email:                 session.Email,
		EmailVerified:         true,
		AcceptedPrivacyPolicy: session.AcceptedPrivacyPolicy,
	}

	userID, err := s.userRepo.Create(ctx, user)
	if err != nil {
		s.logger.Errorf("Error creating user: %v", err)
		return err
	}

	// Assign default user role
	role, err := s.roleRepo.GetRoleByName(ctx, domain.DefaultRoles.User)
	if err != nil {
		s.logger.Errorf("Error getting default user role: %v", err)
		return err
	}

	err = s.roleRepo.AssignRoleToUser(ctx, userID, role.ID)
	if err != nil {
		s.logger.Errorf("Error assigning role to user: %v", err)
		return err
	}

	// Delete registration session
	err = s.sessionRepo.DeleteRegistrationSession(ctx, req.RegistrationSessionID)
	if err != nil {
		s.logger.Errorf("Error deleting registration session: %v", err)
		// Don't fail if we can't delete the session
	}

	return nil
}

// ResendVerificationCode resends the verification code for a registration session
func (s *AuthService) ResendVerificationCode(ctx context.Context, req domain.ResendCodeRequest) (*domain.RegistrationSessionResponse, error) {
	// Get registration session
	session, err := s.sessionRepo.GetRegistrationSession(ctx, req.RegistrationSessionID)
	if err != nil {
		// Don't expose that the session doesn't exist
		return &domain.RegistrationSessionResponse{
			RegistrationSessionID: req.RegistrationSessionID,
			CodeExpires:           time.Now().UTC().Add(15 * time.Minute).Unix(),
			Code:                  generateCode(), // Only for debugging
		}, nil
	}

	// Generate new verification code
	code := generateCode()
	codeExpires := time.Now().UTC().Add(15 * time.Minute)

	// Update registration session
	err = s.sessionRepo.UpdateRegistrationSessionCode(ctx, req.RegistrationSessionID, code, codeExpires)
	if err != nil {
		s.logger.Errorf("Error updating registration session code: %v", err)
		return nil, err
	}

	// Send verification code
	err = s.emailSvc.SendVerificationCode(session.Email, code)
	if err != nil {
		s.logger.Errorf("Error sending verification code: %v", err)
		// We don't want to fail the registration process if email sending fails
		// Just log the error and continue
	}

	return &domain.RegistrationSessionResponse{
		RegistrationSessionID: req.RegistrationSessionID,
		CodeExpires:           codeExpires.Unix(),
		Code:                  code, // Only for debugging
	}, nil
}

// SendLoginCode sends a login code to a user's email
func (s *AuthService) SendLoginCode(ctx context.Context, req domain.LoginRequest) (*domain.LoginSessionResponse, error) {
	// Generate verification code
	code := generateCode()
	codeExpires := time.Now().UTC().Add(15 * time.Minute)

	// Create login session
	session := domain.LoginSession{
		Email:       req.Email,
		Code:        code,
		CodeExpires: codeExpires,
		CreatedAt:   time.Now().UTC(),
	}

	_, err := s.sessionRepo.CreateLoginSession(ctx, session)
	if err != nil {
		s.logger.Errorf("Error creating login session: %v", err)
		return nil, err
	}

	// Check if the email exists
	_, err = s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil {
		// If email exists, send verification code
		err = s.emailSvc.SendVerificationCode(req.Email, code)
		if err != nil {
			s.logger.Errorf("Error sending verification code: %v", err)
			// We don't want to fail the login process if email sending fails
			// Just log the error and continue
		}
	} else {
		// If email doesn't exist, we still return a success response for privacy reasons
		// But we don't actually send an email
		s.logger.Infof("Login attempt with non-existent email: %s", req.Email)
	}

	return &domain.LoginSessionResponse{
		CodeExpires: codeExpires.Unix(),
		Code:        code, // Only for debugging
	}, nil
}

// ConfirmLogin confirms a login attempt with a verification code
func (s *AuthService) ConfirmLogin(ctx context.Context, req domain.LoginConfirmRequest, userAgent, ip string) (*domain.TokenResponse, error) {
	// Get login session
	session, err := s.sessionRepo.GetLoginSessionByEmailAndCode(ctx, req.Email, req.Code)
	if err != nil {

		return nil, errors.New("неверный или истекший код подтверждения. Пожалуйста, запросите новый код и попробуйте снова")
	}

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("неверный или истекший код подтверждения. Пожалуйста, запросите новый код и попробуйте снова")
	}

	// Get user roles
	roles, err := s.roleRepo.GetUserRoleNames(ctx, user.ID)
	if err != nil {
		s.logger.Errorf("Error getting user roles: %v", err)
		return nil, err
	}

	// Generate token pair
	tokenPair, err := s.tokenSvc.GenerateTokenPair(ctx, user, roles)
	if err != nil {
		s.logger.Errorf("Error generating token pair: %v", err)
		return nil, err
	}

	// Store refresh token
	err = s.tokenSvc.StoreRefreshToken(ctx, user.ID, tokenPair.RefreshToken, userAgent, ip)
	if err != nil {
		s.logger.Errorf("Error storing refresh token: %v", err)
		return nil, err
	}

	// Delete login session
	err = s.sessionRepo.DeleteLoginSession(ctx, session.ID)
	if err != nil {
		s.logger.Errorf("Error deleting login session: %v", err)
		// Don't fail if we can't delete the session
	}

	return &domain.TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, req domain.RefreshTokenRequest, userAgent, ip string) (*domain.TokenResponse, error) {
	// Get token session
	tokenSession, err := s.sessionRepo.GetTokenSession(ctx, req.RefreshToken)
	if err != nil {
		if err.Error() == "token expires" {
			return nil, errors.New("token expires")
		}
		s.logger.Errorf("Error getting token session: %v", err)
		return nil, errors.New("token invalid")
	}

	// Get user by ID
	user, err := s.userRepo.GetByID(ctx, tokenSession.UserID)
	if err != nil {
		s.logger.Errorf("Error getting user by ID: %v", err)
		return nil, errors.New("token invalid")
	}

	// Get user roles
	roles, err := s.roleRepo.GetUserRoleNames(ctx, user.ID)
	if err != nil {
		s.logger.Errorf("Error getting user roles: %v", err)
		return nil, err
	}

	// Revoke old refresh token
	err = s.tokenSvc.RevokeRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		s.logger.Errorf("Error revoking refresh token: %v", err)
		// Don't fail if we can't revoke the token
	}

	// Generate new token pair
	tokenPair, err := s.tokenSvc.GenerateTokenPair(ctx, user, roles)
	if err != nil {
		s.logger.Errorf("Error generating token pair: %v", err)
		return nil, err
	}

	// Store new refresh token
	err = s.tokenSvc.StoreRefreshToken(ctx, user.ID, tokenPair.RefreshToken, userAgent, ip)
	if err != nil {
		s.logger.Errorf("Error storing refresh token: %v", err)
		return nil, err
	}

	return &domain.TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(ctx context.Context, id int64) (domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// GetUserRoles retrieves all roles for a user
func (s *AuthService) GetUserRoles(ctx context.Context, userID int64) ([]string, error) {
	return s.roleRepo.GetUserRoleNames(ctx, userID)
}

// HasRole checks if a user has a specific role
func (s *AuthService) HasRole(ctx context.Context, userID int64, roleName string) (bool, error) {
	return s.roleRepo.HasRole(ctx, userID, roleName)
}

// generateCode generates a random 4-digit verification code
func generateCode() string {
	// Seed the random number generator
	rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

	// Generate a random number between 1000 and 9999
	code := rand.Intn(9000) + 1000

	return fmt.Sprintf("%d", code)
}
