package router

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	"authmicro/internal/api/rest/handler"
	customMiddleware "authmicro/internal/api/rest/middleware"
	"authmicro/internal/service"
	"authmicro/pkg/logger"
)

// Router is responsible for setting up and managing the HTTP router
type Router struct {
	echo *echo.Echo
}

// NewRouter creates a new router
func NewRouter(authService *service.AuthService, tokenService *service.TokenService, logger logger.Logger) *Router {
	e := echo.New()

	// Set up middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))

	// Create handlers
	authHandler := handler.NewAuthHandler(authService, logger)

	// Create custom middleware
	authMiddleware := customMiddleware.NewAuthMiddleware(tokenService, authService, logger)

	// Create API group
	api := e.Group("/auth/v1")

	// Registration routes
	api.POST("/registration", authHandler.Register)
	api.POST("/registration/confirmEmail", authHandler.ConfirmEmail)
	api.POST("/registration/resendCodeEmail", authHandler.ResendVerificationCode)

	// Login routes
	api.POST("/login/sendCodeEmail", authHandler.SendLoginCode)
	api.POST("/login/confirmEmail", authHandler.ConfirmLogin)

	// Token routes
	api.POST("/refreshToken", authHandler.RefreshToken)

	// Protected routes example
	protected := api.Group("/protected")
	protected.Use(authMiddleware.JWT())
	protected.GET("", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "You have access to protected resource",
		})
	})

	// Admin routes example
	admin := api.Group("/admin")
	admin.Use(authMiddleware.JWT(), authMiddleware.RoleRequired(domain.DefaultRoles.Admin))
	admin.GET("", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "You have access to admin resource",
		})
	})

	// Swagger documentation
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	return &Router{
		echo: e,
	}
}

// Start starts the HTTP server
func (r *Router) Start(address string) error {
	return r.echo.Start(address)
}

// Shutdown gracefully shuts down the HTTP server
func (r *Router) Shutdown(ctx context.Context) error {
	return r.echo.Shutdown(ctx)
}
