package middleware

import (
	"net/http"
	"strings"

	"github.com/auth-service/internal/domain/entity"
	"github.com/auth-service/internal/domain/service"
	"github.com/gin-gonic/gin"
)

// JWTMiddleware creates middleware for JWT authentication
func JWTMiddleware(tokenService service.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Extract token from header
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || strings.ToLower(headerParts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header format must be 'Bearer {token}'",
			})
			c.Abort()
			return
		}

		// Validate token
		claims, err := tokenService.ValidateAccessToken(headerParts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Set claims in context
		c.Set("claims", claims)
		c.Next()
	}
}

// RoleMiddleware creates middleware for role-based authorization
func RoleMiddleware(roles ...entity.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get claims from context
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*entity.UserClaims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
			c.Abort()
			return
		}

		// Check if user has required role
		hasRole := false
		for _, role := range roles {
			if userClaims.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Forbidden: insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
