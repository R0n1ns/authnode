package handler

import (
	"context"
	"errors"
	"net/http"

	"authmicro/internal/domain"
	"authmicro/pkg/logger"
	"github.com/labstack/echo/v4"
)

type AuthService interface {
	CreateRegistrationSession(context.Context, domain.RegistrationRequest) (*domain.RegistrationSessionResponse, []domain.FieldError, error)
	ConfirmEmail(context.Context, domain.ConfirmEmailRequest) error
	ResendVerificationCode(context.Context, domain.ResendCodeRequest) (*domain.RegistrationSessionResponse, error)
	SendLoginCode(context.Context, domain.LoginRequest) (*domain.LoginSessionResponse, error)
	ConfirmLogin(context.Context, domain.LoginConfirmRequest, string, string) (*domain.TokenResponse, error)
	RefreshToken(context.Context, domain.RefreshTokenRequest, string, string) (*domain.TokenResponse, error)
	Logout(context.Context, string) error
	LogoutAll(context.Context, int64) error
}
type AuthHandler struct {
	service AuthService
	logger  logger.Logger
}

func NewAuthHandler(s AuthService, l logger.Logger) *AuthHandler {
	return &AuthHandler{service: s, logger: l}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req domain.RegistrationRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c)
	}
	res, fields, err := h.service.CreateRegistrationSession(c.Request().Context(), req)
	if err != nil {
		return h.writeError(c, err)
	}
	if len(fields) > 0 {
		return c.JSON(http.StatusUnprocessableEntity, domain.ErrorResponse{Error: "validation failed", DetailedErrors: fields})
	}
	return c.JSON(http.StatusCreated, res)
}
func (h *AuthHandler) ConfirmEmail(c echo.Context) error {
	var req domain.ConfirmEmailRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c)
	}
	if err := h.service.ConfirmEmail(c.Request().Context(), req); err != nil {
		return h.writeError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
func (h *AuthHandler) ResendVerificationCode(c echo.Context) error {
	var req domain.ResendCodeRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c)
	}
	res, err := h.service.ResendVerificationCode(c.Request().Context(), req)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, res)
}
func (h *AuthHandler) SendLoginCode(c echo.Context) error {
	var req domain.LoginRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c)
	}
	res, err := h.service.SendLoginCode(c.Request().Context(), req)
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusAccepted, res)
}
func (h *AuthHandler) ConfirmLogin(c echo.Context) error {
	var req domain.LoginConfirmRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c)
	}
	res, err := h.service.ConfirmLogin(c.Request().Context(), req, c.Request().UserAgent(), c.RealIP())
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, res)
}
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req domain.RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c)
	}
	res, err := h.service.RefreshToken(c.Request().Context(), req, c.Request().UserAgent(), c.RealIP())
	if err != nil {
		return h.writeError(c, err)
	}
	return c.JSON(http.StatusOK, res)
}
func (h *AuthHandler) Logout(c echo.Context) error {
	var req domain.RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c)
	}
	if err := h.service.Logout(c.Request().Context(), req.RefreshToken); err != nil {
		return h.writeError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
func (h *AuthHandler) LogoutAll(c echo.Context) error {
	claims, ok := c.Get("user").(*domain.TokenClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "unauthorized"})
	}
	if err := h.service.LogoutAll(c.Request().Context(), claims.UserID); err != nil {
		return h.writeError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
func (h *AuthHandler) writeError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, domain.ErrRateLimited):
		return c.JSON(http.StatusTooManyRequests, domain.ErrorResponse{Error: domain.ErrRateLimited.Error()})
	case errors.Is(err, domain.ErrInvalidCode):
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: domain.ErrInvalidCode.Error()})
	case errors.Is(err, domain.ErrInvalidInput):
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: domain.ErrInvalidInput.Error()})
	case errors.Is(err, domain.ErrInvalidToken):
		return c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Error: domain.ErrInvalidToken.Error()})
	default:
		h.logger.Errorf("request failed: %v", err)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "internal server error"})
	}
}
func badRequest(c echo.Context) error {
	return c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: "invalid request body"})
}
