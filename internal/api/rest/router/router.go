package router

import (
	"context"
	"net/http"
	"time"

	"authmicro/internal/api/rest/handler"
	custom "authmicro/internal/api/rest/middleware"
	"authmicro/internal/domain"
	"authmicro/internal/service"
	"authmicro/pkg/logger"
	"github.com/labstack/echo/v4"
	em "github.com/labstack/echo/v4/middleware"
)

type databasePinger interface{ PingContext(context.Context) error }
type Router struct{ echo *echo.Echo }

func NewRouter(auth *service.AuthService, tokens *service.TokenService, log logger.Logger, db databasePinger) *Router {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(em.RequestID(), em.Recover(), em.Secure(), em.BodyLimit("1M"))
	e.Use(em.TimeoutWithConfig(em.TimeoutConfig{Timeout: 15 * time.Second}))
	e.Use(em.CORSWithConfig(em.CORSConfig{AllowOrigins: []string{"http://localhost:3000"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization}}))
	h := handler.NewAuthHandler(auth, log)
	guard := custom.NewAuthMiddleware(tokens, auth, log)
	api := e.Group("/api/v1/auth")
	api.POST("/registrations", h.Register)
	api.POST("/registrations/confirm", h.ConfirmEmail)
	api.POST("/registrations/resend", h.ResendVerificationCode)
	api.POST("/login/code", h.SendLoginCode)
	api.POST("/login/confirm", h.ConfirmLogin)
	api.POST("/tokens/refresh", h.RefreshToken)
	api.POST("/logout", h.Logout)
	protected := api.Group("")
	protected.Use(guard.JWT())
	protected.POST("/logout-all", h.LogoutAll)
	protected.GET("/me", func(c echo.Context) error { return c.JSON(http.StatusOK, c.Get("user")) })
	admin := protected.Group("/admin")
	admin.Use(guard.RoleRequired(domain.DefaultRoles.Admin))
	admin.GET("/check", func(c echo.Context) error { return c.NoContent(http.StatusNoContent) })
	e.GET("/live", func(c echo.Context) error { return c.JSON(http.StatusOK, map[string]string{"status": "ok"}) })
	e.GET("/ready", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "not ready"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ready"})
	})
	e.File("/openapi.yaml", "api/openapi.yaml")
	return &Router{echo: e}
}
func (r *Router) Start(address string) error         { return r.echo.Start(address) }
func (r *Router) Shutdown(ctx context.Context) error { return r.echo.Shutdown(ctx) }
func (r *Router) Handler() http.Handler              { return r.echo }
