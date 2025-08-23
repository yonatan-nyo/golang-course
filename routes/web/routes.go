package web

import (
	"os"
	"path/filepath"
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

// getProjectRoot finds the project root by looking for go.mod file
func getProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Walk up the directory tree to find go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached filesystem root
		}
		dir = parent
	}

	return ""
}

// getTemplatePattern returns the absolute path pattern for templates
func getTemplatePattern() string {
	if projectRoot := getProjectRoot(); projectRoot != "" {
		return filepath.Join(projectRoot, "templates", "**", "*")
	}
	return "templates/**/*"
}

func SetupWebRoutes(r *gin.Engine,
	authController *webAuth.AuthController,
	adminDashboardController *webAdminDashboard.DashboardController,
	adminCourseController *webAdminCourse.CourseController,
	adminUserController *webAdminUser.UserController,
	adminModuleController *webAdminModule.ModuleController,
	userDashboardController *webUserDashboard.DashboardController,
	userCourseController *webUserCourse.CourseController,
	userModuleController *webUserModule.ModuleController) {
	// Load HTML templates with absolute path
	r.LoadHTMLGlob(getTemplatePattern())

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
