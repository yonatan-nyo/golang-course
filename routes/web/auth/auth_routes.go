package auth

import (
	webAuth "yonatan/labpro/controllers/web"

	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(webRoutes *gin.RouterGroup, authController *webAuth.AuthController) {
	// Public routes (no authentication required)
	authRoutes := webRoutes.Group("/auth")
	{
		authRoutes.GET("/login", authController.ShowLoginPage)
		authRoutes.POST("/login", authController.HandleLogin)
		authRoutes.GET("/register", authController.ShowRegisterPage)
		authRoutes.POST("/register", authController.HandleRegister)
		authRoutes.POST("/logout", authController.HandleLogout)
	}
}
