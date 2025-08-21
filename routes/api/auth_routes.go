package api

import (
	"yonatan/labpro/controllers"
	"yonatan/labpro/middleware"

	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(api *gin.RouterGroup, authController *controllers.AuthController) {
	auth := api.Group("/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
		auth.GET("/self", middleware.AuthMiddleware(), authController.GetSelf)
	}
}
