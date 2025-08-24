package admin

import (
	webAdminCourse "yonatan/labpro/controllers/web/admin"
	webAdminDashboard "yonatan/labpro/controllers/web/admin"
	webAdminModule "yonatan/labpro/controllers/web/admin"
	webAdminUser "yonatan/labpro/controllers/web/admin"
	"yonatan/labpro/middleware"

	"github.com/gin-gonic/gin"
)

func SetupAdminRoutes(webRoutes *gin.RouterGroup,
	adminDashboardController *webAdminDashboard.DashboardController,
	adminCourseController *webAdminCourse.CourseController,
	adminUserController *webAdminUser.UserController,
	adminModuleController *webAdminModule.ModuleController) {

	// Admin routes (admin authentication required)
	adminRoutes := webRoutes.Group("/admin")
	adminRoutes.Use(middleware.WebAdminMiddleware())
	{
		adminRoutes.GET("", adminDashboardController.ShowAdminDashboard)
		adminRoutes.GET("/dashboard", adminDashboardController.ShowAdminDashboard)

		// Course management routes
		adminRoutes.GET("/courses", adminCourseController.ShowCoursesPage)
		adminRoutes.GET("/courses/create", adminCourseController.ShowCreateCoursePage)
		adminRoutes.POST("/courses/create", adminCourseController.HandleCreateCourse)
		adminRoutes.GET("/courses/:id/edit", adminCourseController.ShowEditCoursePage)
		adminRoutes.POST("/courses/:id/edit", adminCourseController.HandleUpdateCourse)
		adminRoutes.DELETE("/courses/:id", adminCourseController.HandleDeleteCourse)

		// Course modules management
		adminRoutes.GET("/courses/:id/modules", adminModuleController.ShowCourseModulesPage)
		adminRoutes.GET("/courses/:id/modules/create", adminModuleController.ShowCreateModulePageForCourse)

		// User management routes
		adminRoutes.GET("/users", adminUserController.ShowUsersPage)
		adminRoutes.GET("/users/create", adminUserController.ShowCreateUserPage)
		adminRoutes.POST("/users/create", adminUserController.CreateUser)
		adminRoutes.GET("/users/:id", adminUserController.ShowUserDetails)
		adminRoutes.GET("/users/:id/edit", adminUserController.ShowEditUserPage)
		adminRoutes.POST("/users/:id/edit", adminUserController.HandleUpdateUser)
		adminRoutes.POST("/users/:id/balance", adminUserController.HandleUpdateBalance)
		adminRoutes.DELETE("/users/:id", adminUserController.HandleDeleteUser)

		// Module management routes
		adminRoutes.GET("/modules", adminModuleController.ShowModulesPage)
		adminRoutes.GET("/modules/create", adminModuleController.ShowCreateModulePage)
		adminRoutes.POST("/modules/create", adminModuleController.HandleCreateModule)
		adminRoutes.GET("/modules/:id/edit", adminModuleController.ShowEditModulePage)
		adminRoutes.POST("/modules/:id/edit", adminModuleController.HandleUpdateModule)
		adminRoutes.DELETE("/modules/:id", adminModuleController.HandleDeleteModule)
	}
}
