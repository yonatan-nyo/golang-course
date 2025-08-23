package middleware

import (
	"net/http"
	"strings"
	"yonatan/labpro/config"
	"yonatan/labpro/database"
	"yonatan/labpro/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Authorization header required",
				"data":    nil,
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Bearer token required",
				"data":    nil,
			})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Invalid token",
				"data":    nil,
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Invalid token claims",
				"data":    nil,
			})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Invalid user ID in token",
				"data":    nil,
			})
			c.Abort()
			return
		}

		var user models.User
		if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "User not found",
				"data":    nil,
			})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Set("user_id", user.ID)
		if user.IsAdmin {
			c.Set("user_role", "admin")
		} else {
			c.Set("user_role", "user")
		}
		c.Next()
	}
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "User not authenticated",
				"data":    nil,
			})
			c.Abort()
			return
		}

		userModel, ok := user.(models.User)
		if !ok || !userModel.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"status":  "error",
				"message": "Admin access required",
				"data":    nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
