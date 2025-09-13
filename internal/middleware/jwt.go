package middleware

import (
	"api/internal/services"
	"api/pkg/errors"
	"api/pkg/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type JWTMiddleware struct {
	jwtService services.JWTServiceInterface
}

func NewJWTMiddleware(jwtService services.JWTServiceInterface) *JWTMiddleware {
	return &JWTMiddleware{jwtService: jwtService}
}

// AuthRequired middleware validates JWT token
func (m *JWTMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := m.extractTokenFromHeader(c)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "missing or invalid authorization header")
			c.Abort()
			return
		}

		claims, err := m.jwtService.GetClaimsFromToken(token)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}

		// Set user information in context
		if userID, ok := claims["user_id"].(float64); ok {
			c.Set("user_id", uint(userID))
		}
		if isAdmin, ok := claims["is_admin"].(bool); ok {
			c.Set("is_admin", isAdmin)
		}

		c.Next()
	}
}

// AdminRequired middleware ensures user is an admin
func (m *JWTMiddleware) AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			response.Error(c, http.StatusForbidden, "admin access required")
			c.Abort()
			return
		}
		c.Next()
	}
}

// extractTokenFromHeader extracts JWT token from Authorization header
func (m *JWTMiddleware) extractTokenFromHeader(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.NewUnauthorizedError("authorization header required", nil)
	}

	// Check for Bearer prefix
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", errors.NewUnauthorizedError("invalid authorization header format", nil)
	}

	token := strings.TrimPrefix(authHeader, bearerPrefix)
	if token == "" {
		return "", errors.NewUnauthorizedError("token missing", nil)
	}

	return token, nil
}
