package api

import (
	apiAdminModule "yonatan/labpro/controllers/api/admin"
	apiUserModule "yonatan/labpro/controllers/api/user"
	"yonatan/labpro/middleware"

	"github.com/gin-gonic/gin"
)

func SetupModuleRoutes(api *gin.RouterGroup,
	adminModuleController *apiAdminModule.ModuleAPIController,
	userModuleController *apiUserModule.ModuleAPIController) {

	// User module routes
	modules := api.Group("/modules")
	modules.Use(middleware.AuthMiddleware())
	{
		// GET /api/modules/:id
		modules.GET("/:id", userModuleController.GetModuleByID)
		// PATCH /api/modules/:id/complete
		modules.PATCH("/:id/complete", userModuleController.CompleteModule)
	}

	// Course modules routes (both admin and user)
	courseModules := api.Group("/courses/:courseId/modules")
	courseModules.Use(middleware.AuthMiddleware())
	{
		// GET /api/courses/:courseId/modules (all authenticated users)
		courseModules.GET("", userModuleController.GetCourseModules)
	}

	// Admin module routes
	adminModules := api.Group("/modules")
	adminModules.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		// PUT /api/modules/:id (admin only)
		adminModules.PUT("/:id", adminModuleController.UpdateModule)
		// DELETE /api/modules/:id (admin only)
		adminModules.DELETE("/:id", adminModuleController.DeleteModule)
	}

	// Admin course module routes
	adminCourseModules := api.Group("/courses/:courseId/modules")
	adminCourseModules.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		// POST /api/courses/:courseId/modules (admin only)
		adminCourseModules.POST("", adminModuleController.CreateModule)
		// PATCH /api/courses/:courseId/modules/reorder (admin only)
		adminCourseModules.PATCH("/reorder", adminModuleController.ReorderModules)
	}
}
