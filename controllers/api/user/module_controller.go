package user

import (
	"net/http"
	"strconv"
	"yonatan/labpro/models"
	"yonatan/labpro/services"

	"github.com/gin-gonic/gin"
)

type ModuleAPIController struct {
	moduleService *services.ModuleService
}

func NewModuleAPIController(moduleService *services.ModuleService) *ModuleAPIController {
	return &ModuleAPIController{
		moduleService: moduleService,
	}
}

func (mac *ModuleAPIController) GetCourseModules(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userModel := user.(models.User)
	courseID := c.Param("courseId")

	// Get query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "15"))
	if limit > 50 {
		limit = 50
	}

	// Get course modules
	modules, pagination, err := mac.moduleService.GetModules(courseID, userModel.ID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch modules",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"message":    "Modules retrieved successfully",
		"data":       modules,
		"pagination": pagination,
	})
}

func (mac *ModuleAPIController) GetModuleByID(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userModel := user.(models.User)
	moduleID := c.Param("id")

	// Determine user role
	userRole := "user"
	if userModel.IsAdmin {
		userRole = "admin"
	}

	// Get module details
	module, err := mac.moduleService.GetModuleByID(moduleID, userModel.ID, userRole)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "Module not found or access denied",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Module retrieved successfully",
		"data":    module,
	})
}

func (mac *ModuleAPIController) CompleteModule(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userModel := user.(models.User)
	moduleID := c.Param("id")

	// Mark module as completed
	result, err := mac.moduleService.CompleteModule(moduleID, userModel.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Module completed successfully",
		"data":    result,
	})
}
