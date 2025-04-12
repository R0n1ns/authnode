package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"authmicro/internal/domain"
	"authmicro/pkg/logger"
)

type AuthService interface {
	CreateRegistrationSession(ctx context.Context, req domain.RegistrationRequest) (*domain.RegistrationSessionResponse, []domain.FieldError, error)
	ConfirmEmail(ctx context.Context, req domain.ConfirmEmailRequest) error
	ResendVerificationCode(ctx context.Context, req domain.ResendCodeRequest) (*domain.RegistrationSessionResponse, error)
	SendLoginCode(ctx context.Context, req domain.LoginRequest) (*domain.LoginSessionResponse, error)
	ConfirmLogin(ctx context.Context, req domain.LoginConfirmRequest, userAgent, ip string) (*domain.TokenResponse, error)
	RefreshToken(ctx context.Context, req domain.RefreshTokenRequest, userAgent, ip string) (*domain.TokenResponse, error)
	HasRole(ctx context.Context, userID int64, roleName string) (bool, error)
}

type AuthHandler struct {
	authService AuthService
	logger      logger.Logger
}

func NewAuthHandler(authService AuthService, logger logger.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Create a new registration session for a user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.RegistrationRequest true "Registration request"
// @Success 200 {object} domain.RegistrationSessionResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /auth/v1/registration [post]
func (h *AuthHandler) Register(c echo.Context) error {
	var req domain.RegistrationRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Errorf("Error binding request: %v", err)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error: "Неверный формат запроса",
		})
	}

	res, fieldErrors, err := h.authService.CreateRegistrationSession(c.Request().Context(), req)
	if err != nil {
		h.logger.Errorf("Error creating registration session: %v", err)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error: "Сервер не отвечает",
		})
	}

	if fieldErrors != nil && len(fieldErrors) > 0 {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:          "Ошибка регистрации",
			DetailedErrors: fieldErrors,
		})
	}

	return c.JSON(http.StatusOK, res)
}

// ConfirmEmail handles email confirmation
// @Summary Confirm email
// @Description Confirm a user's email using a verification code
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.ConfirmEmailRequest true "Confirm email request"
// @Success 200 {object} interface{}
// @Failure 400 {object} domain.ErrorResponse
// @Router /auth/v1/registration/confirmEmail [post]
func (h *AuthHandler) ConfirmEmail(c echo.Context) error {
	var req domain.ConfirmEmailRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Errorf("Error binding request: %v", err)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error: "Неверный формат запроса",
		})
	}

	err := h.authService.ConfirmEmail(c.Request().Context(), req)
	if err != nil {
		h.logger.Errorf("Error confirming email: %v", err)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, struct{}{})
}

// ResendVerificationCode handles resending the verification code
// @Summary Resend verification code
// @Description Resend a verification code to the user's email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.ResendCodeRequest true "Resend code request"
// @Success 200 {object} domain.RegistrationSessionResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /auth/v1/registration/resendCodeEmail [post]
func (h *AuthHandler) ResendVerificationCode(c echo.Context) error {
	var req domain.ResendCodeRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Errorf("Error binding request: %v", err)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error: "Неверный формат запроса",
		})
	}

	res, err := h.authService.ResendVerificationCode(c.Request().Context(), req)
	if err != nil {
		h.logger.Errorf("Error resending verification code: %v", err)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error: "Сервер не отвечает",
		})
	}

	return c.JSON(http.StatusOK, res)
}

// SendLoginCode handles sending login code to email
// @Summary Send login code
// @Description Send a login verification code to the user's email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.LoginRequest true "Login request"
// @Success 200 {object} domain.LoginSessionResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /auth/v1/login/sendCodeEmail [post]
func (h *AuthHandler) SendLoginCode(c echo.Context) error {
	var req domain.LoginRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Errorf("Error binding request: %v", err)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error: "Неверный формат запроса",
		})
	}

	res, err := h.authService.SendLoginCode(c.Request().Context(), req)
	if err != nil {
		h.logger.Errorf("Error sending login code: %v", err)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error: "Сервер не отвечает",
		})
	}

	return c.JSON(http.StatusOK, res)
}

// ConfirmLogin handles confirming login with a code
// @Summary Confirm login
// @Description Confirm login using a verification code sent to email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.LoginConfirmRequest true "Login confirmation request"
// @Success 200 {object} domain.TokenResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /auth/v1/login/confirmEmail [post]
func (h *AuthHandler) ConfirmLogin(c echo.Context) error {
	var req domain.LoginConfirmRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Errorf("Error binding request: %v", err)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error: "Неверный формат запроса",
		})
	}

	userAgent := c.Request().UserAgent()
	ip := c.RealIP()

	res, err := h.authService.ConfirmLogin(c.Request().Context(), req, userAgent, ip)
	if err != nil {
		h.logger.Errorf("Error confirming login: %v", err)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, res)
}

// RefreshToken handles refreshing tokens
// @Summary Refresh tokens
// @Description Refresh access token using a valid refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} domain.TokenResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /auth/v1/refreshToken [post]
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req domain.RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Errorf("Error binding request: %v", err)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error: "Неверный формат запроса",
		})
	}

	userAgent := c.Request().UserAgent()
	ip := c.RealIP()

	res, err := h.authService.RefreshToken(c.Request().Context(), req, userAgent, ip)
	if err != nil {
		h.logger.Errorf("Error refreshing token: %v", err)

		// Return specific error message based on the error
		if strings.Contains(err.Error(), "token expires") {
			return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Error: "token expires",
			})
		} else if strings.Contains(err.Error(), "token invalid") {
			return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Error: "token invalid",
			})
		}

		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error: "Сервер не отвечает",
		})
	}

	return c.JSON(http.StatusOK, res)
}
