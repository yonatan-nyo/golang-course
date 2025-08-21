package user

import (
	"net/http"
	"yonatan/labpro/models"
	"yonatan/labpro/services"

	"github.com/gin-gonic/gin"
)

type ModuleController struct {
	moduleService *services.ModuleService
}

func NewModuleController(moduleService *services.ModuleService) *ModuleController {
	return &ModuleController{
		moduleService: moduleService,
	}
}

func (mc *ModuleController) ShowModuleDetail(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	userModel := user.(models.User)
	moduleID := c.Param("id")

	// Get module details
	module, err := mc.moduleService.GetModuleByID(moduleID, userModel.ID, "user")
	if err != nil {
		c.HTML(http.StatusNotFound, "module-detail.html", gin.H{
			"Title": "Module Detail",
			"User":  userModel,
			"Error": "Module not found or access denied",
		})
		return
	}

	c.HTML(http.StatusOK, "module-detail.html", gin.H{
		"Title":  "Module Detail",
		"User":   userModel,
		"Module": module,
	})
}

func (mc *ModuleController) HandleCompleteModule(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userModel := user.(models.User)
	moduleID := c.Param("id")

	// Mark module as completed
	result, err := mc.moduleService.CompleteModule(moduleID, userModel.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Module completed successfully",
		"data":    result,
	})
}
