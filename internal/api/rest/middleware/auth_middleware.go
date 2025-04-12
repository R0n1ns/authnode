package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"authmicro/internal/domain"
	"authmicro/pkg/logger"
)

type TokenService interface {
	ValidateToken(token string) (*domain.TokenClaims, error)
}

type AuthService interface {
	HasRole(ctx context.Context, userID int64, roleName string) (bool, error)
}

type AuthMiddleware struct {
	tokenService TokenService
	authService  AuthService
	logger       logger.Logger
}

func NewAuthMiddleware(tokenService TokenService, authService AuthService, logger logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
		authService:  authService,
		logger:       logger,
	}
}

// JWT middleware to validate JWT token
func (m *AuthMiddleware) JWT() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get token from Authorization header
			auth := c.Request().Header.Get("Authorization")
			if auth == "" {
				return c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
					Error: "Unauthorized",
				})
			}

			// Check if the authorization header has the correct format
			parts := strings.Split(auth, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
					Error: "Invalid authorization header format",
				})
			}

			// Extract token
			token := parts[1]

			// Validate token
			claims, err := m.tokenService.ValidateToken(token)
			if err != nil {
				m.logger.Errorf("Error validating token: %v", err)
				return c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
					Error: "Invalid or expired token",
				})
			}

			// Set user in context
			c.Set("user", claims)
			c.Set("userID", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("nickname", claims.Nickname)
			c.Set("roles", claims.Roles)

			return next(c)
		}
	}
}

// RoleRequired middleware to check if the user has the required role
func (m *AuthMiddleware) RoleRequired(role string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if user is authenticated
			claims, ok := c.Get("user").(*domain.TokenClaims)
			if !ok {
				return c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
					Error: "Unauthorized",
				})
			}

			// Check if user has the required role
			for _, r := range claims.Roles {
				if r == role {
					return next(c)
				}
			}

			// If not in token, double-check with database (token might be outdated)
			hasRole, err := m.authService.HasRole(c.Request().Context(), claims.UserID, role)
			if err != nil {
				m.logger.Errorf("Error checking role: %v", err)
				return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
					Error: "Internal server error",
				})
			}

			if hasRole {
				return next(c)
			}

			return c.JSON(http.StatusForbidden, domain.ErrorResponse{
				Error: "Insufficient permissions",
			})
		}
	}
}
