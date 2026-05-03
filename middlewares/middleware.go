package middlewares

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/milua25/e-commerce-backend/tokens"
)

func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation for adding a product to the cart goes here
		clientToken := c.GetHeader("Authorization")
		if clientToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization token is required"})
			return
		}
		// Remove "Bearer " prefix

		tokenString := strings.TrimPrefix(clientToken, "Bearer ")
		if tokenString == clientToken {
			log.Println("JWT token not found in Authorization header")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Validate the token and extract user information
		claims, err := tokens.ValidateToken(tokenString, "your_secret_key")
		if err != nil {
			log.Println("Invalid JWT token:", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// You can set the user information in the context for further use
		c.Set("email", claims.Email)
		c.Set("firstName", claims.FirstName)
		c.Set("lastName", claims.LastName)
		c.Set("uid", claims.Uid)
		c.Next()
	}
}
