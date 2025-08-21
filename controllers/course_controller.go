package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"yonatan/labpro/models"
	"yonatan/labpro/services"

	"github.com/gin-gonic/gin"
)

type CourseController struct {
	courseService *services.CourseService
}

func NewCourseController(courseService *services.CourseService) *CourseController {
	return &CourseController{courseService: courseService}
}

func (cc *CourseController) CreateCourse(c *gin.Context) {
	// Check if user is admin
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "Access denied. Admin only.",
			"data":    nil,
		})
		return
	}

	var course models.Course
	course.Title = c.PostForm("title")
	course.Description = c.PostForm("description")
	course.Instructor = c.PostForm("instructor")

	// Parse topics from form
	topicsStr := c.PostForm("topics")
	if topicsStr != "" {
		course.Topics = strings.Split(topicsStr, ",")
		// Trim spaces
		for i, topic := range course.Topics {
			course.Topics[i] = strings.TrimSpace(topic)
		}
	}

	priceStr := c.PostForm("price")
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid price format",
			"data":    nil,
		})
		return
	}
	course.Price = price

	// Handle file upload for thumbnail
	file, err := c.FormFile("thumbnail_image")
	if err == nil {
		// Save file and get URL (implement file storage logic)
		thumbnailURL, err := cc.courseService.SaveThumbnail(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to save thumbnail",
				"data":    nil,
			})
			return
		}
		course.ThumbnailImage = &thumbnailURL
	}

	// Validate required fields
	if course.Title == "" || course.Description == "" || course.Instructor == "" || len(course.Topics) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Title, description, instructor, and topics are required",
			"data":    nil,
		})
		return
	}

	createdCourse, err := cc.courseService.CreateCourse(&course)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to create course: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Course created successfully",
		"data":    createdCourse,
	})
}

func (cc *CourseController) GetCourses(c *gin.Context) {
	// Get query parameters
	q := c.Query("q")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "15")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 15
	}

	// Get user ID from context to check purchased status
	userID, _ := c.Get("user_id")

	courses, pagination, err := cc.courseService.GetCourses(q, page, limit, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get courses: " + err.Error(),
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

func (cc *CourseController) GetCourse(c *gin.Context) {
	id := c.Param("id")

	userID, _ := c.Get("user_id")

	course, err := cc.courseService.GetCourseByID(id, userID)
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

func (cc *CourseController) UpdateCourse(c *gin.Context) {
	// Check if user is admin
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "Access denied. Admin only.",
			"data":    nil,
		})
		return
	}

	id := c.Param("id")

	var course models.Course
	course.ID = id
	course.Title = c.PostForm("title")
	course.Description = c.PostForm("description")
	course.Instructor = c.PostForm("instructor")

	// Parse topics from form
	topicsStr := c.PostForm("topics")
	if topicsStr != "" {
		course.Topics = strings.Split(topicsStr, ",")
		// Trim spaces
		for i, topic := range course.Topics {
			course.Topics[i] = strings.TrimSpace(topic)
		}
	}

	priceStr := c.PostForm("price")
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid price format",
			"data":    nil,
		})
		return
	}
	course.Price = price

	// Handle file upload for thumbnail
	file, err := c.FormFile("thumbnail_image")
	if err == nil {
		// Save file and get URL (implement file storage logic)
		thumbnailURL, err := cc.courseService.SaveThumbnail(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to save thumbnail",
				"data":    nil,
			})
			return
		}
		course.ThumbnailImage = &thumbnailURL
	}

	// Validate required fields
	if course.Title == "" || course.Description == "" || course.Instructor == "" || len(course.Topics) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Title, description, instructor, and topics are required",
			"data":    nil,
		})
		return
	}

	updatedCourse, err := cc.courseService.UpdateCourse(&course)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to update course: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Course updated successfully",
		"data":    updatedCourse,
	})
}

func (cc *CourseController) DeleteCourse(c *gin.Context) {
	// Check if user is admin
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "Access denied. Admin only.",
			"data":    nil,
		})
		return
	}

	id := c.Param("id")

	err := cc.courseService.DeleteCourse(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to delete course: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.Status(http.StatusNoContent)
}

func (cc *CourseController) BuyCourse(c *gin.Context) {
	id := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized",
			"data":    nil,
		})
		return
	}

	result, err := cc.courseService.BuyCourse(id, userID.(string))
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

func (cc *CourseController) GetMyCourses(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized",
			"data":    nil,
		})
		return
	}

	// Get query parameters
	q := c.Query("q")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "15")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 15
	}

	courses, pagination, err := cc.courseService.GetMyCourses(userID.(string), q, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get my courses: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"message":    "My courses retrieved successfully",
		"data":       courses,
		"pagination": pagination,
	})
}
