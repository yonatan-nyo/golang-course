package api

import (
	"yonatan/labpro/config"
	apiAuth "yonatan/labpro/controllers/api"
	"yonatan/labpro/middleware"

	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(api *gin.RouterGroup, authController *apiAuth.AuthAPIController, cfg *config.Config) {
	auth := api.Group("/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
		auth.POST("/logout", authController.Logout)
		auth.GET("/self", middleware.AuthMiddleware(cfg), authController.GetProfile)
	}
}
