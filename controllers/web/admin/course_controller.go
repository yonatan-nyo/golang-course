package admin

import (
	"net/http"
	"strconv"
	"yonatan/labpro/models"
	"yonatan/labpro/services"

	"github.com/gin-gonic/gin"
)

type CourseController struct {
	courseService *services.CourseService
}

func NewCourseController(courseService *services.CourseService) *CourseController {
	return &CourseController{
		courseService: courseService,
	}
}

func (cc *CourseController) ShowCoursesPage(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	userModel := user.(models.User)
	if !userModel.IsAdmin {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}

	// Get query parameters for pagination and search
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	query := c.Query("q")

	// Get courses from service
	courses, pagination, err := cc.courseService.GetCourses(query, page, limit, userModel.ID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "courses.html", gin.H{
			"Title": "Courses Management",
			"User":  userModel,
			"Error": "Failed to fetch courses",
		})
		return
	}

	c.HTML(http.StatusOK, "courses.html", gin.H{
		"Title":      "Courses Management",
		"User":       userModel,
		"Courses":    courses,
		"Pagination": pagination,
		"Query":      query,
	})
}

func (cc *CourseController) ShowCreateCoursePage(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	userModel := user.(models.User)
	if !userModel.IsAdmin {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}

	c.HTML(http.StatusOK, "course-create.html", gin.H{
		"Title": "Create Course",
		"User":  userModel,
	})
}

func (cc *CourseController) ShowEditCoursePage(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	userModel := user.(models.User)
	if !userModel.IsAdmin {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}

	courseID := c.Param("id")
	course, err := cc.courseService.GetCourseByID(courseID, userModel.ID)
	if err != nil {
		c.HTML(http.StatusNotFound, "course-edit.html", gin.H{
			"Title": "Edit Course",
			"User":  userModel,
			"Error": "Course not found",
		})
		return
	}

	c.HTML(http.StatusOK, "course-edit.html", gin.H{
		"Title":  "Edit Course",
		"User":   userModel,
		"Course": course,
	})
}

func (cc *CourseController) HandleCreateCourse(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	userModel := user.(models.User)
	if !userModel.IsAdmin {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}

	// Handle form submission
	title := c.PostForm("title")
	description := c.PostForm("description")
	instructor := c.PostForm("instructor")
	priceStr := c.PostForm("price")
	topics := c.PostFormArray("topics")

	// Validate required fields
	if title == "" || instructor == "" || priceStr == "" {
		c.HTML(http.StatusBadRequest, "course-create.html", gin.H{
			"Title": "Create Course",
			"User":  userModel,
			"Error": "Title, instructor, and price are required",
		})
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "course-create.html", gin.H{
			"Title": "Create Course",
			"User":  userModel,
			"Error": "Invalid price format",
		})
		return
	}

	// Handle file upload
	thumbnailURL := ""
	file, header, err := c.Request.FormFile("thumbnail")
	if err == nil && header != nil {
		defer file.Close()
		thumbnailURL, err = cc.courseService.SaveThumbnail(header)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "course-create.html", gin.H{
				"Title": "Create Course",
				"User":  userModel,
				"Error": "Failed to save thumbnail: " + err.Error(),
			})
			return
		}
	}

	// Create course
	course := &models.Course{
		Title:       title,
		Description: description,
		Instructor:  instructor,
		Price:       price,
		Thumbnail:   thumbnailURL,
		Topics:      topics,
	}

	createdCourse, err := cc.courseService.CreateCourse(course)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "course-create.html", gin.H{
			"Title": "Create Course",
			"User":  userModel,
			"Error": "Failed to create course: " + err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin/courses?success=Course created successfully&id="+createdCourse.ID)
}

func (cc *CourseController) HandleUpdateCourse(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	userModel := user.(models.User)
	if !userModel.IsAdmin {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}

	courseID := c.Param("id")

	// Get existing course
	existingCourse, err := cc.courseService.GetCourseByID(courseID, userModel.ID)
	if err != nil {
		c.HTML(http.StatusNotFound, "course-edit.html", gin.H{
			"Title": "Edit Course",
			"User":  userModel,
			"Error": "Course not found",
		})
		return
	}

	// Handle form submission
	title := c.PostForm("title")
	description := c.PostForm("description")
	instructor := c.PostForm("instructor")
	priceStr := c.PostForm("price")
	topics := c.PostFormArray("topics")

	// Validate required fields
	if title == "" || instructor == "" || priceStr == "" {
		c.HTML(http.StatusBadRequest, "course-edit.html", gin.H{
			"Title":  "Edit Course",
			"User":   userModel,
			"Course": existingCourse,
			"Error":  "Title, instructor, and price are required",
		})
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "course-edit.html", gin.H{
			"Title":  "Edit Course",
			"User":   userModel,
			"Course": existingCourse,
			"Error":  "Invalid price format",
		})
		return
	}

	// Handle file upload
	thumbnailURL := ""
	if thumbnailValue, exists := existingCourse["thumbnail_image"]; exists && thumbnailValue != nil {
		thumbnailURL = thumbnailValue.(string)
	}

	file, header, err := c.Request.FormFile("thumbnail")
	if err == nil && header != nil {
		defer file.Close()
		thumbnailURL, err = cc.courseService.SaveThumbnail(header)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "course-edit.html", gin.H{
				"Title":  "Edit Course",
				"User":   userModel,
				"Course": existingCourse,
				"Error":  "Failed to save thumbnail: " + err.Error(),
			})
			return
		}
	}

	// Update course
	course := &models.Course{
		ID:          courseID,
		Title:       title,
		Description: description,
		Instructor:  instructor,
		Price:       price,
		Thumbnail:   thumbnailURL,
		Topics:      topics,
	}

	_, err = cc.courseService.UpdateCourse(course)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "course-edit.html", gin.H{
			"Title":  "Edit Course",
			"User":   userModel,
			"Course": existingCourse,
			"Error":  "Failed to update course: " + err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin/courses?success=Course updated successfully&id="+courseID)
}

func (cc *CourseController) HandleDeleteCourse(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userModel := user.(models.User)
	if !userModel.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	courseID := c.Param("id")
	err := cc.courseService.DeleteCourse(courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete course"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Course deleted successfully"})
}
