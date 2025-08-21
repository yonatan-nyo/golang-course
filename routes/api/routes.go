package api

import (
	apiAuth "yonatan/labpro/controllers/api"
	apiAdminCourse "yonatan/labpro/controllers/api/admin"
	apiAdminModule "yonatan/labpro/controllers/api/admin"
	apiAdminUser "yonatan/labpro/controllers/api/admin"
	apiUserCourse "yonatan/labpro/controllers/api/user"
	apiUserModule "yonatan/labpro/controllers/api/user"

	"github.com/gin-gonic/gin"
)

// SetupAPIRoutes sets up all API routes
func SetupAPIRoutes(api *gin.RouterGroup,
	authController *apiAuth.AuthAPIController,
	adminCourseController *apiAdminCourse.CourseAPIController,
	adminModuleController *apiAdminModule.ModuleAPIController,
	adminUserController *apiAdminUser.UserAPIController,
	userCourseController *apiUserCourse.CourseAPIController,
	userModuleController *apiUserModule.ModuleAPIController) {
	// Setup all API route groups
	SetupAuthRoutes(api, authController)
	SetupCourseRoutes(api, adminCourseController, userCourseController)
	SetupModuleRoutes(api, adminModuleController, userModuleController)
	SetupUserRoutes(api, adminUserController)
}
