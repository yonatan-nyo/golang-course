package middleware

import (
	"net/http"
	"yonatan/labpro/config"
	"yonatan/labpro/database"
	"yonatan/labpro/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// WebAuthMiddleware checks for authentication via cookies for web routes
func WebAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Load()

		// Get token from cookie
		token, err := c.Cookie("token")
		if err != nil || token == "" {
			c.Redirect(http.StatusFound, "/auth/login")
			c.Abort()
			return
		}

		// Parse and validate token
		claims := &jwt.MapClaims{}
		parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !parsedToken.Valid {
			c.SetCookie("token", "", -1, "/", "", false, true) // Clear invalid token
			c.Redirect(http.StatusFound, "/auth/login")
			c.Abort()
			return
		}

		// Extract user ID from claims
		userID, ok := (*claims)["user_id"].(string)
		if !ok {
			c.SetCookie("token", "", -1, "/", "", false, true)
			c.Redirect(http.StatusFound, "/auth/login")
			c.Abort()
			return
		}

		// Get user from database
		db := database.GetDB()
		var user models.User
		if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
			c.SetCookie("token", "", -1, "/", "", false, true)
			c.Redirect(http.StatusFound, "/auth/login")
			c.Abort()
			return
		}

		// Set user in context
		c.Set("user", user)
		c.Next()
	}
}

// OptionalWebAuthMiddleware checks for authentication but doesn't redirect if not authenticated
func OptionalWebAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		cfg := config.Load()

		// Get token from cookie
		token, err := c.Cookie("token")
		if err != nil || token == "" {
			c.Next()
			return
		}

		// Parse and validate token
		claims := &jwt.MapClaims{}
		parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !parsedToken.Valid {
			c.SetCookie("token", "", -1, "/", "", false, true)
			c.Next()
			return
		}

		// Extract user ID from claims
		userID, ok := (*claims)["user_id"].(string)
		if !ok {
			c.SetCookie("token", "", -1, "/", "", false, true)
			c.Next()
			return
		}

		// Get user from database
		db := database.GetDB()
		var user models.User
		if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
			c.SetCookie("token", "", -1, "/", "", false, true)
			c.Next()
			return
		}

		// Set user in context
		c.Set("user", user)
		c.Next()
	}
}
