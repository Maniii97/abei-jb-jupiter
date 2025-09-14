package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware configures and returns CORS middleware with appropriate settings for the API
func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowAllOrigins: true,

		// Allow all HTTP methods
		AllowMethods: []string{"*"},

		// Allow all headers
		AllowHeaders: []string{"*"},

		ExposeHeaders: []string{
			"Content-Length",
			"X-Rate-Limit-Limit",
			"X-Rate-Limit-Remaining",
			"X-Rate-Limit-Reset",
		},

		AllowCredentials: true,

		MaxAge: 12 * time.Hour,
	})
}
