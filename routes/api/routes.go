package api

import (
	"yonatan/labpro/controllers"

	"github.com/gin-gonic/gin"
)

// SetupAPIRoutes sets up all API routes
func SetupAPIRoutes(api *gin.RouterGroup, authController *controllers.AuthController, courseController *controllers.CourseController, moduleController *controllers.ModuleController, userController *controllers.UserController) {
	// Setup all API route groups
	SetupAuthRoutes(api, authController)
	SetupCourseRoutes(api, courseController)
	SetupModuleRoutes(api, moduleController)
	SetupUserRoutes(api, userController)
}
