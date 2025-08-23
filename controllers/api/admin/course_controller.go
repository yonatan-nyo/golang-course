package admin

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

// CreateCourse godoc
// @Summary      Create a new course (Admin only)
// @Description  Create a new course with title, description, instructor, price, topics and thumbnail
// @Tags         admin-courses
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        title        formData  string   true   "Course title"
// @Param        description  formData  string   false  "Course description"
// @Param        instructor   formData  string   true   "Course instructor"
// @Param        price        formData  string   true   "Course price"
// @Param        topics       formData  []string false  "Course topics array"
// @Param        thumbnail    formData  file     false  "Course thumbnail image"
// @Success      201          {object}  object{status=string,message=string,data=object}
// @Failure      400          {object}  object{status=string,message=string,data=object}
// @Failure      401          {object}  object{error=string}
// @Failure      403          {object}  object{error=string}
// @Failure      500          {object}  object{status=string,message=string,data=object}
// @Router       /courses [post]
func (cac *CourseAPIController) CreateCourse(c *gin.Context) {
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

	// Handle form data for multipart/form-data
	title := c.PostForm("title")
	description := c.PostForm("description")
	instructor := c.PostForm("instructor")
	priceStr := c.PostForm("price")
	topics := c.PostFormArray("topics")

	if title == "" || instructor == "" || priceStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Title, instructor, and price are required",
			"data":    nil,
		})
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid price format",
			"data":    nil,
		})
		return
	}

	// Handle thumbnail upload
	thumbnailURL := ""
	if file, header, err := c.Request.FormFile("thumbnail_image"); err == nil && header != nil {
		defer file.Close()
		thumbnailURL, err = cac.courseService.SaveThumbnail(header)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to save thumbnail: " + err.Error(),
				"data":    nil,
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

	createdCourse, err := cac.courseService.CreateCourse(course)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to create course",
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

// UpdateCourse godoc
// @Summary      Update a course (Admin only)
// @Description  Update an existing course with new information
// @Tags         admin-courses
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        courseId     path      string   true   "Course ID"
// @Param        title        formData  string   false  "Course title"
// @Param        description  formData  string   false  "Course description"
// @Param        instructor   formData  string   false  "Course instructor"
// @Param        price        formData  string   false  "Course price"
// @Param        topics       formData  []string false  "Course topics array"
// @Param        thumbnail    formData  file     false  "Course thumbnail image"
// @Success      200          {object}  object{status=string,message=string,data=object}
// @Failure      400          {object}  object{status=string,message=string,data=object}
// @Failure      401          {object}  object{error=string}
// @Failure      403          {object}  object{error=string}
// @Failure      404          {object}  object{status=string,message=string,data=object}
// @Failure      500          {object}  object{status=string,message=string,data=object}
// @Router       /courses/{courseId} [put]
func (cac *CourseAPIController) UpdateCourse(c *gin.Context) {
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

	courseID := c.Param("courseId")

	// Get existing course to preserve thumbnail if no new one is provided
	existingCourse, err := cac.courseService.GetCourseByID(courseID, userModel.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "Course not found",
			"data":    nil,
		})
		return
	}

	// Handle form data for multipart/form-data
	title := c.PostForm("title")
	description := c.PostForm("description")
	instructor := c.PostForm("instructor")
	priceStr := c.PostForm("price")
	topics := c.PostFormArray("topics")

	if title == "" || instructor == "" || priceStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Title, instructor, and price are required",
			"data":    nil,
		})
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid price format",
			"data":    nil,
		})
		return
	}

	// Create course object for update, preserving existing thumbnail
	existingThumbnail := ""
	if thumbnail, ok := existingCourse["thumbnail_image"].(string); ok {
		existingThumbnail = thumbnail
	}

	course := &models.Course{
		ID:          courseID,
		Title:       title,
		Description: description,
		Instructor:  instructor,
		Price:       price,
		Topics:      topics,
		Thumbnail:   existingThumbnail, // Preserve existing thumbnail
	}

	// Handle thumbnail upload if provided - this will override the preserved thumbnail
	if file, header, err := c.Request.FormFile("thumbnail_image"); err == nil && header != nil {
		defer file.Close()
		thumbnailURL, err := cac.courseService.SaveThumbnail(header)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to save thumbnail: " + err.Error(),
				"data":    nil,
			})
			return
		}
		course.Thumbnail = thumbnailURL
	}

	updatedCourse, err := cac.courseService.UpdateCourse(course)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to update course",
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

// DeleteCourse godoc
// @Summary      Delete a course (Admin only)
// @Description  Delete an existing course and all its associated data
// @Tags         admin-courses
// @Produce      json
// @Security     BearerAuth
// @Param        courseId  path      string  true  "Course ID"
// @Success      200       {object}  object{status=string,message=string,data=object}
// @Failure      401       {object}  object{error=string}
// @Failure      403       {object}  object{error=string}
// @Failure      500       {object}  object{status=string,message=string,data=object}
// @Router       /courses/{courseId} [delete]
func (cac *CourseAPIController) DeleteCourse(c *gin.Context) {
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

	courseID := c.Param("courseId")
	err := cac.courseService.DeleteCourse(courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to delete course",
			"data":    nil,
		})
		return
	}

	c.Status(http.StatusNoContent)
}
