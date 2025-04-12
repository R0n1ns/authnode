package grpc

import (
	"context"

	"authmicro/internal/domain/entity"
	"authmicro/internal/domain/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthServer implements the gRPC AuthService interface
type AuthServer struct {
	UnimplementedAuthServiceServer
	authService  *service.AuthService
	tokenService service.TokenService
}

// NewAuthServer creates a new gRPC auth server
func NewAuthServer(authService *service.AuthService, tokenService service.TokenService) *AuthServer {
	return &AuthServer{
		authService:  authService,
		tokenService: tokenService,
	}
}

// CreateRegistration handles user registration requests via gRPC
func (s *AuthServer) CreateRegistration(ctx context.Context, req *RegistrationRequest) (*RegistrationResponse, error) {
	// Validate and create registration session
	session, err := s.authService.CreateRegistrationSession(
		ctx,
		req.FirstName,
		req.LastName,
		req.Nickname,
		req.Email,
		req.AcceptedPrivacyPolicy,
	)

	if err != nil {
		// Handle specific errors
		switch err {
		case service.ErrInvalidInput:
			return nil, status.Error(codes.InvalidArgument, "All fields are required")
		case service.ErrNicknameExists:
			return nil, status.Error(codes.AlreadyExists, "Nickname already exists")
		case service.ErrInvalidNickname:
			return nil, status.Error(codes.InvalidArgument, "Nickname contains forbidden characters")
		case service.ErrInvalidEmail:
			return nil, status.Error(codes.InvalidArgument, "Invalid email format")
		case service.ErrPrivacyPolicyRequired:
			return nil, status.Error(codes.InvalidArgument, "Privacy policy acceptance required")
		default:
			return nil, status.Error(codes.Internal, "Internal server error")
		}
	}

	// Return success response
	return &RegistrationResponse{
		RegistrationSessionId: session.ID,
		CodeExpires:           session.CodeExpires.Unix(),
		Code:                  session.Code, // Only for debugging
	}, nil
}

// ConfirmEmail handles email confirmation during registration via gRPC
func (s *AuthServer) ConfirmEmail(ctx context.Context, req *ConfirmEmailRequest) (*EmptyResponse, error) {
	// Confirm email
	err := s.authService.ConfirmEmail(
		ctx,
		req.RegistrationSessionId,
		req.Code,
	)

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid or expired verification code")
	}

	// Return success response
	return &EmptyResponse{}, nil
}

// ResendCodeEmail handles requests to resend verification code via gRPC
func (s *AuthServer) ResendCodeEmail(ctx context.Context, req *ResendCodeEmailRequest) (*CodeResponse, error) {
	// Resend verification code
	session, err := s.authService.ResendCodeEmail(
		ctx,
		req.RegistrationSessionId,
	)

	if err != nil {
		return &CodeResponse{
			CodeExpires: 0,
			Code:        "",
		}, nil
	}

	// Return success response
	return &CodeResponse{
		CodeExpires: session.CodeExpires.Unix(),
		Code:        session.Code, // Only for debugging
	}, nil
}

// SendLoginCodeEmail handles requests to send login verification code via gRPC
func (s *AuthServer) SendLoginCodeEmail(ctx context.Context, req *SendLoginCodeEmailRequest) (*CodeResponse, error) {
	// Send login verification code
	session, err := s.authService.SendLoginCodeEmail(
		ctx,
		req.Email,
	)

	if err != nil {
		return &CodeResponse{
			CodeExpires: 0,
			Code:        "",
		}, nil
	}

	// Return success response
	return &CodeResponse{
		CodeExpires: session.CodeExpires.Unix(),
		Code:        session.Code, // Only for debugging
	}, nil
}

// ConfirmLogin handles login confirmation and token generation via gRPC
func (s *AuthServer) ConfirmLogin(ctx context.Context, req *ConfirmLoginRequest) (*TokenResponse, error) {
	// Confirm login and generate tokens
	tokenPair, err := s.authService.ConfirmLogin(
		ctx,
		req.Email,
		req.Code,
	)

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid or expired verification code")
	}

	// Return success response
	return &TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// RefreshToken handles token refresh requests via gRPC
func (s *AuthServer) RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*TokenResponse, error) {
	// Refresh token
	tokenPair, err := s.authService.RefreshToken(
		ctx,
		req.RefreshToken,
	)

	if err != nil {
		// Handle specific errors
		switch err {
		case service.ErrTokenExpired:
			return nil, status.Error(codes.Unauthenticated, "Token expired")
		default:
			return nil, status.Error(codes.Unauthenticated, "Invalid token")
		}
	}

	// Return success response
	return &TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// ValidateToken validates a token and returns user claims via gRPC
func (s *AuthServer) ValidateToken(ctx context.Context, req *ValidateTokenRequest) (*UserResponse, error) {
	// Validate token
	claims, err := s.tokenService.ValidateAccessToken(req.AccessToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Invalid or expired token")
	}

	// Return user claims
	return &UserResponse{
		UserId:   claims.UserID,
		Nickname: claims.Nickname,
		Email:    claims.Email,
		Role:     string(claims.Role),
	}, nil
}
