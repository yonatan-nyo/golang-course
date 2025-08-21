package api

import (
	apiAdminUser "yonatan/labpro/controllers/api/admin"
	"yonatan/labpro/middleware"

	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(api *gin.RouterGroup,
	adminUserController *apiAdminUser.UserAPIController) {

	// All user routes are admin-only according to the contract
	users := api.Group("/users")
	users.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		// GET /api/users
		users.GET("", adminUserController.GetUsers)
		// GET /api/users/:id
		users.GET("/:id", adminUserController.GetUserByID)
		// POST /api/users/:id/balance
		users.POST("/:id/balance", adminUserController.UpdateUserBalance)
		// PUT /api/users/:id
		users.PUT("/:id", adminUserController.UpdateUser)
		// DELETE /api/users/:id
		users.DELETE("/:id", adminUserController.DeleteUser)
	}
}
