package user

import (
	"net/http"
	"strconv"
	"yonatan/labpro/models"
	"yonatan/labpro/services"

	"github.com/gin-gonic/gin"
)

type CourseAPIController struct {
	courseService *services.CourseService
}

func NewCourseAPIController(courseService *services.CourseService) *CourseAPIController {
	return &CourseAPIController{
		courseService: courseService,
	}
}

func (cac *CourseAPIController) GetCourses(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userModel := user.(models.User)

	// Get query parameters
	query := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "15"))
	if limit > 50 {
		limit = 50
	}

	// Get available courses
	courses, pagination, err := cac.courseService.GetCourses(query, page, limit, userModel.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch courses",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"message":    "Courses retrieved successfully",
		"data":       courses,
		"pagination": pagination,
	})
}

func (cac *CourseAPIController) GetCourseByID(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userModel := user.(models.User)
	courseID := c.Param("courseId")

	// Get course details
	course, err := cac.courseService.GetCourseByID(courseID, userModel.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "Course not found",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Course retrieved successfully",
		"data":    course,
	})
}

func (cac *CourseAPIController) GetMyCourses(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userModel := user.(models.User)

	// Get query parameters
	query := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "15"))
	if limit > 50 {
		limit = 50
	}

	// Get user's enrolled courses
	enrolledCourses, pagination, err := cac.courseService.GetMyCourses(userModel.ID, query, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch enrolled courses",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"message":    "My courses retrieved successfully",
		"data":       enrolledCourses,
		"pagination": pagination,
	})
}

func (cac *CourseAPIController) PurchaseCourse(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userModel := user.(models.User)
	courseID := c.Param("courseId")

	// Purchase course
	result, err := cac.courseService.BuyCourse(courseID, userModel.ID)
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
		"message": "Course purchased successfully",
		"data":    result,
	})
}
