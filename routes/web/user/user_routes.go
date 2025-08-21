package user

import (
	webUserCourse "yonatan/labpro/controllers/web/user"
	webUserDashboard "yonatan/labpro/controllers/web/user"
	webUserModule "yonatan/labpro/controllers/web/user"
	"yonatan/labpro/middleware"

	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(webRoutes *gin.RouterGroup,
	userDashboardController *webUserDashboard.DashboardController,
	userCourseController *webUserCourse.CourseController,
	userModuleController *webUserModule.ModuleController) {

	// Dashboard route
	webRoutes.GET("/dashboard", middleware.WebAuthMiddleware(), userDashboardController.ShowDashboard)

	// User routes (user authentication required)
	userRoutes := webRoutes.Group("/")
	userRoutes.Use(middleware.WebAuthMiddleware())
	{
		// Course browsing
		userRoutes.GET("/courses", userCourseController.ShowCoursesPage)
		userRoutes.GET("/courses/:id", userCourseController.ShowCourseDetail)
		userRoutes.POST("/courses/:id/purchase", userCourseController.HandlePurchaseCourse)
		userRoutes.GET("/my-courses", userCourseController.ShowMyCourses)

		// Module viewing
		userRoutes.GET("/modules/:id", userModuleController.ShowModuleDetail)
		userRoutes.POST("/modules/:id/complete", userModuleController.HandleCompleteModule)
	}
}
