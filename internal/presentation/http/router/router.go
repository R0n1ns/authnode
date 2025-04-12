package router

import (
	"net/http"

	"authmicro/internal/presentation/http/handler"
	"authmicro/internal/presentation/http/middleware"
	"github.com/go-chi/chi/v5"
)

// SetupRoutes sets up all the routes for the application
func SetupRoutes(r chi.Router, authHandler *handler.AuthHandler, authMiddleware *middleware.AuthMiddleware) {
	// Public routes
	r.Route("/api/v1/auth", func(r chi.Router) {
		// Registration
		r.Post("/register", authHandler.StartRegistration)
		r.Post("/verify-email", authHandler.VerifyEmail)
		r.Post("/resend-code", authHandler.ResendVerificationCode)

		// Login
		r.Post("/login", authHandler.Login)
		r.Post("/verify-login", authHandler.VerifyLogin)
		r.Post("/refresh-token", authHandler.RefreshToken)

		// Protected routes example
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)

			r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
				// Get user info from context
				claims, ok := middleware.GetUserClaims(r)
				if !ok {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				// Return user info
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"id":"` + claims.UserID + `","email":"` + claims.Email + `"}`))
			})
		})
	})

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}
