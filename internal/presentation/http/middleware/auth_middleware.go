package middleware

import (
	"context"
	"net/http"
	"strings"

	"authmicro/internal/domain/entity"
	"authmicro/internal/domain/service"
)

// Key type for storing values in request context
type contextKey string

// UserClaimsContextKey is the key used to store user claims in request context
const UserClaimsContextKey contextKey = "userClaims"

// AuthMiddleware handles authentication and authorization middleware
type AuthMiddleware struct {
	tokenService service.TokenService
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(tokenService service.TokenService) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
	}
}

// Authenticate verifies the JWT token and adds user claims to request context
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		// Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]

		// Validate token
		claims, err := m.tokenService.ValidateAccessToken(token)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add claims to request context
		ctx := context.WithValue(r.Context(), UserClaimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole ensures the user has the required role
func (m *AuthMiddleware) RequireRole(role entity.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get claims from context
			claims, ok := r.Context().Value(UserClaimsContextKey).(*entity.UserClaims)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check role
			if claims.Role != role {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserClaims extracts user claims from request context
func GetUserClaims(r *http.Request) (*entity.UserClaims, bool) {
	claims, ok := r.Context().Value(UserClaimsContextKey).(*entity.UserClaims)
	return claims, ok
}
