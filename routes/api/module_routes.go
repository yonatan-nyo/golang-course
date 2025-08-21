package api

import (
	"yonatan/labpro/controllers"
	"yonatan/labpro/middleware"

	"github.com/gin-gonic/gin"
)

func SetupModuleRoutes(api *gin.RouterGroup, moduleController *controllers.ModuleController) {
	// Individual module routes only
	modules := api.Group("/modules")
	modules.Use(middleware.AuthMiddleware())
	{
		modules.GET("/:id", moduleController.GetModule)
		modules.PUT("/:id", moduleController.UpdateModule)
		modules.DELETE("/:id", moduleController.DeleteModule)
		modules.PATCH("/:id/complete", moduleController.CompleteModule)
	}
}
