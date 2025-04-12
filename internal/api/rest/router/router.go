// Package router provides routing functionality for the REST API
// @title Auth Service API
// @version 1.0
// @description Authentication and authorization service API
// @host localhost:8000
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token
package router

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "authmicro/docs" // Import docs for swagger
	"authmicro/internal/api/rest/handler"
	custommiddleware "authmicro/internal/api/rest/middleware"
	"authmicro/internal/service"
	"authmicro/pkg/logger"
)

// Router is an interface for the REST API router
type Router interface {
	Start(addr string) error
	Shutdown(ctx context.Context) error
}

// EchoRouter implements the Router interface using Echo framework
type EchoRouter struct {
	e      *echo.Echo
	logger logger.Logger
}

// NewRouter creates a new instance of the Router
func NewRouter(authService *service.AuthService, tokenService *service.TokenService, logger logger.Logger) *EchoRouter {
	e := echo.New()

	// Add middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, logger)

	// Initialize middleware
	authMiddleware := custommiddleware.NewAuthMiddleware(tokenService, authService, logger)

	// Public routes (no auth required)
	v1 := e.Group("/auth/v1")

	// Health check
	// @Summary Health check
	// @Description Get health status of the service
	// @Tags health
	// @Produce json
	// @Success 200 {object} map[string]string
	// @Router /auth/v1/health [get]
	v1.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// Swagger documentation
	e.GET("/swagger/*", echoSwagger.EchoWrapHandler(echoSwagger.URL("/swagger/doc.json")))

	// Registration endpoints
	registration := v1.Group("/registration")
	registration.POST("", authHandler.Register)
	registration.POST("/confirmEmail", authHandler.ConfirmEmail)
	registration.POST("/resendCodeEmail", authHandler.ResendVerificationCode)

	// Login endpoints
	login := v1.Group("/login")
	login.POST("/sendCodeEmail", authHandler.SendLoginCode)
	login.POST("/confirmEmail", authHandler.ConfirmLogin)

	// Token refresh
	v1.POST("/refreshToken", authHandler.RefreshToken)

	// Protected routes (auth required)
	// This would be where we add endpoints that require authentication
	protected := e.Group("/api/v1")
	protected.Use(authMiddleware.JWT())

	// Admin routes (admin role required)
	admin := protected.Group("/admin")
	admin.Use(authMiddleware.RoleRequired("admin"))

	return &EchoRouter{
		e:      e,
		logger: logger,
	}
}

// Start starts the HTTP server
func (r *EchoRouter) Start(addr string) error {
	return r.e.Start(addr)
}

// Shutdown gracefully shuts down the HTTP server
func (r *EchoRouter) Shutdown(ctx context.Context) error {
	return r.e.Shutdown(ctx)
}
