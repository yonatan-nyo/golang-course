package web

import (
	"yonatan/labpro/controllers"
	"yonatan/labpro/middleware"
	"yonatan/labpro/models"

	"github.com/gin-gonic/gin"
)

func SetupWebRoutes(r *gin.Engine, authController *controllers.AuthController) {
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

		webRoutes.GET("/dashboard", middleware.WebAuthMiddleware(), authController.ShowDashboard)

		// Admin routes (admin authentication required)
		adminRoutes := webRoutes.Group("/admin")
		adminRoutes.Use(middleware.WebAdminMiddleware())
		{
			adminRoutes.GET("/", authController.ShowAdminDashboard)
			adminRoutes.GET("/dashboard", authController.ShowAdminDashboard)
			// Add more admin routes here as needed
		}

		// Add more protected routes here as needed
	}
}
