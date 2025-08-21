package user

import (
	"net/http"
	"strconv"
	"yonatan/labpro/models"
	"yonatan/labpro/services"

	"github.com/gin-gonic/gin"
)

type DashboardController struct {
	courseService *services.CourseService
	userService   *services.UserService
	moduleService *services.ModuleService
}

func NewDashboardController(courseService *services.CourseService, userService *services.UserService, moduleService *services.ModuleService) *DashboardController {
	return &DashboardController{
		courseService: courseService,
		userService:   userService,
		moduleService: moduleService,
	}
}

func (dc *DashboardController) ShowDashboard(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	userModel := user.(models.User)

	// Get user's purchased courses only
	enrolledCourses, _, err := dc.courseService.GetMyCourses(userModel.ID, "", 1, 100)
	if err != nil {
		enrolledCourses = []map[string]interface{}{} // Default to empty slice
	}

	// Calculate progress for each course using the same logic as module controller
	completedCoursesCount := 0
	for i, course := range enrolledCourses {
		courseIDInterface := course["id"]
		var courseIDStr string

		switch v := courseIDInterface.(type) {
		case float64:
			courseIDStr = strconv.FormatFloat(v, 'f', 0, 64)
		case int:
			courseIDStr = strconv.Itoa(v)
		case string:
			courseIDStr = v
		default:
			continue // Skip this course if ID format is invalid
		}

		// Get all modules for this course
		allModules, _, err := dc.moduleService.GetModules(courseIDStr, userModel.ID, 1, 100)
		if err != nil {
			continue // Skip this course if modules can't be loaded
		}

		// Calculate course progress using same logic as module controller
		courseProgress := 0
		completedModules := 0
		totalModules := len(allModules)

		if totalModules > 0 {
			for _, mod := range allModules {
				if isCompleted, ok := mod["is_completed"].(bool); ok && isCompleted {
					completedModules++
				}
			}
			courseProgress = (completedModules * 100) / totalModules
		}

		// Update the course with calculated progress
		enrolledCourses[i]["progress_percentage"] = courseProgress

		// Count completed courses (100% progress)
		if courseProgress >= 100 {
			completedCoursesCount++
		}
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"Title":           "Dashboard",
		"User":            userModel,
		"EnrolledCourses": enrolledCourses,
		"EnrolledCount":   len(enrolledCourses),
		"CompletedCount":  completedCoursesCount,
	})
}
