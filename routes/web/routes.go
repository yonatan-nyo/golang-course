package web

import (
	webAuth "yonatan/labpro/controllers/web"
	webAdminCourse "yonatan/labpro/controllers/web/admin"
	webAdminDashboard "yonatan/labpro/controllers/web/admin"
	webAdminModule "yonatan/labpro/controllers/web/admin"
	webAdminUser "yonatan/labpro/controllers/web/admin"
	webUserCourse "yonatan/labpro/controllers/web/user"
	webUserDashboard "yonatan/labpro/controllers/web/user"
	webUserModule "yonatan/labpro/controllers/web/user"
	"yonatan/labpro/middleware"
	"yonatan/labpro/models"

	"github.com/gin-gonic/gin"
)

func SetupWebRoutes(r *gin.Engine,
	authController *webAuth.AuthController,
	adminDashboardController *webAdminDashboard.DashboardController,
	adminCourseController *webAdminCourse.CourseController,
	adminUserController *webAdminUser.UserController,
	adminModuleController *webAdminModule.ModuleController,
	userDashboardController *webUserDashboard.DashboardController,
	userCourseController *webUserCourse.CourseController,
	userModuleController *webUserModule.ModuleController) {
	// Load HTML templates
	r.LoadHTMLGlob("templates/**/*")

	// Web routes (serve HTML pages)
	webRoutes := r.Group("/")
	{
		// Public routes (no authentication required)
		authRoutes := webRoutes.Group("/auth")
		{
			authRoutes.GET("/login", authController.ShowLoginPage)
			authRoutes.POST("/login", authController.HandleLogin)
			authRoutes.GET("/register", authController.ShowRegisterPage)
			authRoutes.POST("/register", authController.HandleRegister)
			authRoutes.POST("/logout", authController.HandleLogout)
		}

		// Root route - redirect to dashboard if authenticated, login if not
		webRoutes.Use(middleware.OptionalWebAuthMiddleware())
		webRoutes.GET("/", func(c *gin.Context) {

			if user, exists := c.Get("user"); exists {
				userModel := user.(models.User)

				if userModel.IsAdmin {
					c.Redirect(302, "/admin/dashboard")
				} else {
					c.Redirect(302, "/dashboard")
				}
			} else {
				c.Redirect(302, "/auth/login")
			}
		})

		webRoutes.GET("/dashboard", middleware.WebAuthMiddleware(), userDashboardController.ShowDashboard)

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

			// User management routes
			adminRoutes.GET("/users", adminUserController.ShowUsersPage)
			adminRoutes.GET("/users/:id/edit", adminUserController.ShowEditUserPage)
			adminRoutes.POST("/users/:id/edit", adminUserController.HandleUpdateUser)
			adminRoutes.DELETE("/users/:id", adminUserController.HandleDeleteUser)

			// Module management routes
			adminRoutes.GET("/modules", adminModuleController.ShowModulesPage)
			adminRoutes.GET("/modules/create", adminModuleController.ShowCreateModulePage)
			adminRoutes.POST("/modules/create", adminModuleController.HandleCreateModule)
			adminRoutes.GET("/modules/:id/edit", adminModuleController.ShowEditModulePage)
			adminRoutes.POST("/modules/:id/edit", adminModuleController.HandleUpdateModule)
			adminRoutes.DELETE("/modules/:id", adminModuleController.HandleDeleteModule)
		}

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

		// Add more protected routes here as needed
	}
}
