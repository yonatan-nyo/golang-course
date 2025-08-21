package router

import (
	"yonatan/labpro/config"
	apiAuth "yonatan/labpro/controllers/api"
	apiAdminCourse "yonatan/labpro/controllers/api/admin"
	apiAdminModule "yonatan/labpro/controllers/api/admin"
	apiAdminUser "yonatan/labpro/controllers/api/admin"
	apiUserCourse "yonatan/labpro/controllers/api/user"
	apiUserModule "yonatan/labpro/controllers/api/user"
	webAuthController "yonatan/labpro/controllers/web"
	webAdminCourse "yonatan/labpro/controllers/web/admin"
	webAdminDashboard "yonatan/labpro/controllers/web/admin"
	webAdminModule "yonatan/labpro/controllers/web/admin"
	webAdminUser "yonatan/labpro/controllers/web/admin"
	webUserCourse "yonatan/labpro/controllers/web/user"
	webUserDashboard "yonatan/labpro/controllers/web/user"
	webUserModule "yonatan/labpro/controllers/web/user"
	"yonatan/labpro/database"
	"yonatan/labpro/routes/api"
	"yonatan/labpro/routes/web"
	"yonatan/labpro/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
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
	courseService := services.NewCourseService(db, cfg)
	moduleService := services.NewModuleService(db, cfg)
	userService := services.NewUserService(db)

	// Initialize controllers
	webAuthCtrl := webAuthController.NewAuthController(authService)
	webAdminDashboardCtrl := webAdminDashboard.NewDashboardController()
	webAdminCourseCtrl := webAdminCourse.NewCourseController(courseService)
	webAdminUserCtrl := webAdminUser.NewUserController(userService)
	webAdminModuleCtrl := webAdminModule.NewModuleController(moduleService)
	webUserDashboardCtrl := webUserDashboard.NewDashboardController(courseService, userService)
	webUserCourseCtrl := webUserCourse.NewCourseController(courseService, moduleService)
	webUserModuleCtrl := webUserModule.NewModuleController(moduleService)

	apiAuthCtrl := apiAuth.NewAuthAPIController(authService)
	apiAdminCourseCtrl := apiAdminCourse.NewCourseAPIController(courseService)
	apiAdminModuleCtrl := apiAdminModule.NewModuleAPIController(moduleService)
	apiAdminUserCtrl := apiAdminUser.NewUserAPIController(userService)
	apiUserCourseCtrl := apiUserCourse.NewCourseAPIController(courseService)
	apiUserModuleCtrl := apiUserModule.NewModuleAPIController(moduleService)

	// Setup web routes (HTML pages)
	web.SetupWebRoutes(r, webAuthCtrl, webAdminDashboardCtrl, webAdminCourseCtrl, webAdminUserCtrl, webAdminModuleCtrl, webUserDashboardCtrl, webUserCourseCtrl, webUserModuleCtrl)

	// Setup API routes
	apiGroup := r.Group("/api")
	{
		api.SetupAPIRoutes(apiGroup, apiAuthCtrl, apiAdminCourseCtrl, apiAdminModuleCtrl, apiAdminUserCtrl, apiUserCourseCtrl, apiUserModuleCtrl)
	}

	return r
}
