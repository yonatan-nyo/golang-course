package api

import (
	"yonatan/labpro/config"
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
	userModuleController *apiUserModule.ModuleAPIController,
	cfg *config.Config) {
	// Setup all API route groups
	SetupAuthRoutes(api, authController, cfg)
	SetupCourseRoutes(api, adminCourseController, userCourseController, cfg)
	SetupModuleRoutes(api, adminModuleController, userModuleController, cfg)
	SetupUserRoutes(api, adminUserController, cfg)
}
