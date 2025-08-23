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

// GetCourses godoc
// @Summary      Get all available courses
// @Description  Get a paginated list of available courses with optional search
// @Tags         courses
// @Produce      json
// @Security     BearerAuth
// @Param        q      query     string  false  "Search query"
// @Param        page   query     int     false  "Page number (default: 1)"
// @Param        limit  query     int     false  "Items per page (default: 15, max: 50)"
// @Success      200    {object}  object{status=string,message=string,data=array,pagination=object}
// @Failure      401    {object}  object{error=string}
// @Failure      500    {object}  object{status=string,message=string,data=object}
// @Router       /courses [get]
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

// GetCourseByID godoc
// @Summary      Get course by ID
// @Description  Get detailed information about a specific course
// @Tags         courses
// @Produce      json
// @Security     BearerAuth
// @Param        courseId  path      string  true  "Course ID"
// @Success      200       {object}  object{status=string,message=string,data=object}
// @Failure      401       {object}  object{error=string}
// @Failure      404       {object}  object{status=string,message=string,data=object}
// @Router       /courses/{courseId} [get]
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

// GetMyCourses godoc
// @Summary      Get user's enrolled courses
// @Description  Get a paginated list of courses that the user has purchased/enrolled in
// @Tags         courses
// @Produce      json
// @Security     BearerAuth
// @Param        q      query     string  false  "Search query"
// @Param        page   query     int     false  "Page number (default: 1)"
// @Param        limit  query     int     false  "Items per page (default: 15, max: 50)"
// @Success      200    {object}  object{status=string,message=string,data=array,pagination=object}
// @Failure      401    {object}  object{error=string}
// @Failure      500    {object}  object{status=string,message=string,data=object}
// @Router       /courses/my-courses [get]
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

// PurchaseCourse godoc
// @Summary      Purchase a course
// @Description  Purchase/enroll in a specific course
// @Tags         courses
// @Produce      json
// @Security     BearerAuth
// @Param        courseId  path      string  true  "Course ID"
// @Success      200       {object}  object{status=string,message=string,data=object}
// @Failure      400       {object}  object{status=string,message=string,data=object}
// @Failure      401       {object}  object{error=string}
// @Router       /courses/{courseId}/buy [post]
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
