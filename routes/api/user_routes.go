package api

import (
	"yonatan/labpro/controllers"
	"yonatan/labpro/middleware"

	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(api *gin.RouterGroup, userController *controllers.UserController) {
	// User routes (admin only)
	users := api.Group("/users")
	users.Use(middleware.AuthMiddleware())
	{
		users.GET("", userController.GetUsers)
		users.GET("/:id", userController.GetUser)
		users.POST("/:id/balance", userController.UpdateUserBalance)
		users.PUT("/:id", userController.UpdateUser)
		users.DELETE("/:id", userController.DeleteUser)
	}
}
