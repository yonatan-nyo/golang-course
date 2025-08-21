package router

import (
	"yonatan/labpro/controllers"
	"yonatan/labpro/database"
	"yonatan/labpro/routes/api"
	"yonatan/labpro/routes/web"
	"yonatan/labpro/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"*"}
	r.Use(cors.New(config))

	// Serve static files
	r.Static("/static", "./static")
	r.Static("/uploads", "./uploads")

	// Get database connection
	db := database.GetDB()

	// Initialize services
	authService := services.NewAuthService()
	courseService := services.NewCourseService(db)
	moduleService := services.NewModuleService(db)
	userService := services.NewUserService(db)

	// Initialize controllers
	authController := controllers.NewAuthController(authService)
	courseController := controllers.NewCourseController(courseService)
	moduleController := controllers.NewModuleController(moduleService)
	userController := controllers.NewUserController(userService)

	// Setup web routes (HTML pages)
	web.SetupWebRoutes(r, authController)

	// Setup API routes
	apiGroup := r.Group("/api")
	{
		api.SetupAPIRoutes(apiGroup, authController, courseController, moduleController, userController)
	}

	return r
}
