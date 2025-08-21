package user

import (
	"net/http"
	"strconv"
	"yonatan/labpro/models"
	"yonatan/labpro/services"

	"github.com/gin-gonic/gin"
)

type CourseController struct {
	courseService *services.CourseService
	moduleService *services.ModuleService
}

func NewCourseController(courseService *services.CourseService, moduleService *services.ModuleService) *CourseController {
	return &CourseController{
		courseService: courseService,
		moduleService: moduleService,
	}
}

func (cc *CourseController) ShowCoursesPage(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	userModel := user.(models.User)

	// Get query parameters for pagination and search
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))
	query := c.Query("q")

	// Get available courses
	courses, pagination, err := cc.courseService.GetCourses(query, page, limit, userModel.ID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "user-courses.html", gin.H{
			"Title": "Available Courses",
			"User":  userModel,
			"Error": "Failed to fetch courses",
		})
		return
	}

	c.HTML(http.StatusOK, "user-courses.html", gin.H{
		"Title":      "Available Courses",
		"User":       userModel,
		"Courses":    courses,
		"Pagination": pagination,
		"Query":      query,
	})
}

func (cc *CourseController) ShowCourseDetail(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	userModel := user.(models.User)
	courseID := c.Param("id")

	// Get course details
	course, err := cc.courseService.GetCourseByID(courseID, userModel.ID)
	if err != nil {
		c.HTML(http.StatusNotFound, "course-detail.html", gin.H{
			"Title": "Course Detail",
			"User":  userModel,
			"Error": "Course not found",
		})
		return
	}

	// Get course modules
	modules, _, err := cc.moduleService.GetModules(courseID, userModel.ID, 1, 100)
	if err != nil {
		modules = []map[string]interface{}{} // Default to empty slice
	}

	c.HTML(http.StatusOK, "course-detail.html", gin.H{
		"Title":   "Course Detail",
		"User":    userModel,
		"Course":  course,
		"Modules": modules,
	})
}

func (cc *CourseController) ShowMyCourses(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	userModel := user.(models.User)

	// Get query parameters for pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))

	// Get user's enrolled courses
	enrolledCourses, pagination, err := cc.courseService.GetCourses("", page, limit, userModel.ID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "my-courses.html", gin.H{
			"Title": "My Courses",
			"User":  userModel,
			"Error": "Failed to fetch enrolled courses",
		})
		return
	}

	c.HTML(http.StatusOK, "my-courses.html", gin.H{
		"Title":      "My Courses",
		"User":       userModel,
		"Courses":    enrolledCourses,
		"Pagination": pagination,
	})
}

func (cc *CourseController) HandlePurchaseCourse(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userModel := user.(models.User)
	courseID := c.Param("id")

	// Purchase course
	result, err := cc.courseService.BuyCourse(courseID, userModel.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully purchased course",
		"data":    result,
	})
}
