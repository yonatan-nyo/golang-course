package user

import (
	"net/http"
	"yonatan/labpro/models"
	"yonatan/labpro/services"

	"github.com/gin-gonic/gin"
)

type DashboardController struct {
	courseService *services.CourseService
	userService   *services.UserService
}

func NewDashboardController(courseService *services.CourseService, userService *services.UserService) *DashboardController {
	return &DashboardController{
		courseService: courseService,
		userService:   userService,
	}
}

func (dc *DashboardController) ShowDashboard(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	userModel := user.(models.User)

	// Get user's enrolled courses
	enrolledCourses, _, err := dc.courseService.GetCourses("", 1, 10, userModel.ID)
	if err != nil {
		enrolledCourses = []map[string]interface{}{} // Default to empty slice
	}

	// Get featured/available courses (courses not enrolled by user)
	availableCourses, _, err := dc.courseService.GetCourses("", 1, 6, "")
	if err != nil {
		availableCourses = []map[string]interface{}{} // Default to empty slice
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"Title":            "Dashboard",
		"User":             userModel,
		"EnrolledCourses":  enrolledCourses,
		"AvailableCourses": availableCourses,
	})
}
