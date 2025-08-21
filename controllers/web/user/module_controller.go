package user

import (
	"net/http"
	"strconv"
	"yonatan/labpro/models"
	"yonatan/labpro/services"

	"github.com/gin-gonic/gin"
)

type ModuleController struct {
	moduleService *services.ModuleService
	courseService *services.CourseService
}

func NewModuleController(moduleService *services.ModuleService, courseService *services.CourseService) *ModuleController {
	return &ModuleController{
		moduleService: moduleService,
		courseService: courseService,
	}
}

func (mc *ModuleController) ShowModuleDetail(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	userModel := user.(models.User)
	moduleIDStr := c.Param("id")

	// Determine user role
	userRole := "user"
	if userModel.IsAdmin {
		userRole = "admin"
	}

	// Get module details
	module, err := mc.moduleService.GetModuleByID(moduleIDStr, userModel.ID, userRole)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Module not found"})
		return
	}

	// Extract course ID from module data
	courseIDInterface, exists := module["course_id"]
	if !exists {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Invalid module data"})
		return
	}

	var courseIDStr string
	switch v := courseIDInterface.(type) {
	case float64:
		courseIDStr = strconv.FormatFloat(v, 'f', 0, 64)
	case int:
		courseIDStr = strconv.Itoa(v)
	case string:
		courseIDStr = v
	default:
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Invalid course ID format"})
		return
	}

	// Get course details
	course, err := mc.courseService.GetCourseByID(courseIDStr, userModel.ID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Course not found"})
		return
	}

	// Check access via module service
	hasAccess, err := mc.moduleService.CheckCourseAccess(userModel.ID, courseIDStr)
	if err != nil || (!hasAccess && !userModel.IsAdmin) {
		c.HTML(http.StatusForbidden, "error.html", gin.H{"error": "You don't have access to this course"})
		return
	}

	// Get all modules in this course for navigation
	allModules, _, err := mc.moduleService.GetModules(courseIDStr, userModel.ID, 1, 100)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load course modules"})
		return
	}

	// Calculate course progress
	courseProgress := 0
	completedModules := 0
	totalModules := len(allModules)

	if !userModel.IsAdmin && totalModules > 0 {
		for _, mod := range allModules {
			if isCompleted, ok := mod["is_completed"].(bool); ok && isCompleted {
				completedModules++
			}
		}
		courseProgress = (completedModules * 100) / totalModules
	}

	c.HTML(http.StatusOK, "module-detail.html", gin.H{
		"Module":           module,
		"Course":           course,
		"AllModules":       allModules,
		"CourseProgress":   courseProgress,
		"CompletedModules": completedModules,
		"TotalModules":     totalModules,
		"User":             userModel,
	})
}

func (mc *ModuleController) HandleCompleteModule(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userModel := user.(models.User)
	moduleIDStr := c.Param("id")

	// Mark module as completed
	result, err := mc.moduleService.CompleteModule(moduleIDStr, userModel.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Module completed successfully",
		"data":    result,
	})
}
