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
	"yonatan/labpro/routes/web/admin"
	"yonatan/labpro/routes/web/auth"
	"yonatan/labpro/routes/web/user"

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
		// Setup auth routes
		auth.SetupAuthRoutes(webRoutes, authController)

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

		// Setup admin routes
		admin.SetupAdminRoutes(webRoutes, adminDashboardController, adminCourseController, adminUserController, adminModuleController)

		// Setup user routes
		user.SetupUserRoutes(webRoutes, userDashboardController, userCourseController, userModuleController)
	}
}
