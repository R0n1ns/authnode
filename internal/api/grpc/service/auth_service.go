package service

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "authmicro/internal/api/grpc/proto"
	"authmicro/internal/domain"
	"authmicro/pkg/logger"
)

type authService interface {
	CreateRegistrationSession(ctx context.Context, req domain.RegistrationRequest) (*domain.RegistrationSessionResponse, []domain.FieldError, error)
	ConfirmEmail(ctx context.Context, req domain.ConfirmEmailRequest) error
	ResendVerificationCode(ctx context.Context, req domain.ResendCodeRequest) (*domain.RegistrationSessionResponse, error)
	SendLoginCode(ctx context.Context, req domain.LoginRequest) (*domain.RegistrationSessionResponse, error)
	ConfirmLogin(ctx context.Context, req domain.LoginConfirmRequest, userAgent, ip string) (*domain.TokenResponse, error)
	RefreshToken(ctx context.Context, req domain.RefreshTokenRequest, userAgent, ip string) (*domain.TokenResponse, error)
	HasRole(ctx context.Context, userID int64, roleName string) (bool, error)
}

type tokenService interface {
	ValidateToken(token string) (*domain.TokenClaims, error)
}

type AuthGRPCService struct {
	pb.UnimplementedAuthServiceServer
	authService  authService
	tokenService tokenService
	logger       logger.Logger
}

func NewAuthGRPCService(authService authService, tokenService tokenService, logger logger.Logger) *AuthGRPCService {
	return &AuthGRPCService{
		authService:  authService,
		tokenService: tokenService,
		logger:       logger,
	}
}

// CreateRegistrationSession creates a new registration session
func (s *AuthGRPCService) CreateRegistrationSession(ctx context.Context, req *pb.RegistrationRequest) (*pb.RegistrationSessionResponse, error) {
	domainReq := domain.RegistrationRequest{
		FirstName:             req.FirstName,
		LastName:              req.LastName,
		Nickname:              req.Nickname,
		Email:                 req.Email,
		AcceptedPrivacyPolicy: req.AcceptedPrivacyPolicy,
	}

	res, fieldErrors, err := s.authService.CreateRegistrationSession(ctx, domainReq)
	if err != nil {
		s.logger.Errorf("Error creating registration session: %v", err)
		return nil, status.Errorf(codes.Internal, "Сервер не отвечает")
	}

	if fieldErrors != nil && len(fieldErrors) > 0 {
		var pbFieldErrors []*pb.FieldError
		for _, fe := range fieldErrors {
			pbFieldErrors = append(pbFieldErrors, &pb.FieldError{
				Field:   fe.Field,
				Message: fe.Message,
			})
		}

		return nil, status.Errorf(codes.InvalidArgument, "Ошибка регистрации")
	}

	return &pb.RegistrationSessionResponse{
		RegistrationSessionId: res.RegistrationSessionID,
		CodeExpires:           res.CodeExpires,
		Code:                  res.Code,
	}, nil
}

// ConfirmEmail confirms a user's email
func (s *AuthGRPCService) ConfirmEmail(ctx context.Context, req *pb.ConfirmEmailRequest) (*pb.EmptyResponse, error) {
	domainReq := domain.ConfirmEmailRequest{
		RegistrationSessionID: req.RegistrationSessionId,
		Code:                  req.Code,
	}

	err := s.authService.ConfirmEmail(ctx, domainReq)
	if err != nil {
		s.logger.Errorf("Error confirming email: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &pb.EmptyResponse{}, nil
}

// ResendVerificationCode resends a verification code
func (s *AuthGRPCService) ResendVerificationCode(ctx context.Context, req *pb.ResendCodeRequest) (*pb.RegistrationSessionResponse, error) {
	domainReq := domain.ResendCodeRequest{
		RegistrationSessionID: req.RegistrationSessionId,
	}

	res, err := s.authService.ResendVerificationCode(ctx, domainReq)
	if err != nil {
		s.logger.Errorf("Error resending verification code: %v", err)
		return nil, status.Errorf(codes.Internal, "Сервер не отвечает")
	}

	return &pb.RegistrationSessionResponse{
		RegistrationSessionId: res.RegistrationSessionID,
		CodeExpires:           res.CodeExpires,
		Code:                  res.Code,
	}, nil
}

// SendLoginCode sends a login code
func (s *AuthGRPCService) SendLoginCode(ctx context.Context, req *pb.LoginRequest) (*pb.RegistrationSessionResponse, error) {
	domainReq := domain.LoginRequest{
		Email: req.Email,
	}

	res, err := s.authService.SendLoginCode(ctx, domainReq)
	if err != nil {
		s.logger.Errorf("Error sending login code: %v", err)
		return nil, status.Errorf(codes.Internal, "Сервер не отвечает")
	}

	return &pb.RegistrationSessionResponse{
		CodeExpires: res.CodeExpires,
		Code:        res.Code,
	}, nil
}

// ConfirmLogin confirms a login
func (s *AuthGRPCService) ConfirmLogin(ctx context.Context, req *pb.LoginConfirmRequest) (*pb.TokenResponse, error) {
	domainReq := domain.LoginConfirmRequest{
		Email: req.Email,
		Code:  req.Code,
	}

	res, err := s.authService.ConfirmLogin(ctx, domainReq, req.UserAgent, req.Ip)
	if err != nil {
		s.logger.Errorf("Error confirming login: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &pb.TokenResponse{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	}, nil
}

// RefreshToken refreshes an access token
func (s *AuthGRPCService) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.TokenResponse, error) {
	domainReq := domain.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	}

	res, err := s.authService.RefreshToken(ctx, domainReq, req.UserAgent, req.Ip)
	if err != nil {
		s.logger.Errorf("Error refreshing token: %v", err)

		// Return specific error message based on the error
		if err.Error() == "token expires" {
			return nil, status.Errorf(codes.InvalidArgument, "token expires")
		} else if err.Error() == "token invalid" {
			return nil, status.Errorf(codes.InvalidArgument, "token invalid")
		}

		return nil, status.Errorf(codes.Internal, "Сервер не отвечает")
	}

	return &pb.TokenResponse{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	}, nil
}

// ValidateToken validates a token
func (s *AuthGRPCService) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	claims, err := s.tokenService.ValidateToken(req.Token)
	if err != nil {
		s.logger.Errorf("Error validating token: %v", err)
		return &pb.ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	return &pb.ValidateTokenResponse{
		Valid:    true,
		UserId:   claims.UserID,
		Email:    claims.Email,
		Nickname: claims.Nickname,
		Roles:    claims.Roles,
	}, nil
}

// HasRole checks if a user has a specific role
func (s *AuthGRPCService) HasRole(ctx context.Context, req *pb.HasRoleRequest) (*pb.HasRoleResponse, error) {
	hasRole, err := s.authService.HasRole(ctx, req.UserId, req.RoleName)
	if err != nil {
		s.logger.Errorf("Error checking role: %v", err)
		return nil, status.Errorf(codes.Internal, "Сервер не отвечает")
	}

	return &pb.HasRoleResponse{
		HasRole: hasRole,
	}, nil
}
