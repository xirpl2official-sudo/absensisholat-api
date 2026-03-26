package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// SwaggerAuthMiddleware provides basic authentication for Swagger documentation
func SwaggerAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Swagger password from environment variable
		swaggerPassword := os.Getenv("SWAGGER_PASSWORD")

		// If no password is set, allow access (for development)
		if swaggerPassword == "" {
			c.Next()
			return
		}

		// Get password from query parameter or header
		password := c.Query("password")
		if password == "" {
			password = c.GetHeader("X-Swagger-Password")
		}

		// Validate password
		if password != swaggerPassword {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or missing Swagger password",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
