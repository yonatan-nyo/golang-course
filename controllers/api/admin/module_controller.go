package admin

import (
	"net/http"
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

// CreateModule godoc
// @Summary      Create a new module (Admin only)
// @Description  Create a new module for a specific course with title, description, PDF and video files
// @Tags         admin-modules
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        courseId     path      string  true   "Course ID"
// @Param        title        formData  string  true   "Module title"
// @Param        description  formData  string  false  "Module description"
// @Param        pdf_file     formData  file    false  "PDF file"
// @Param        video_file   formData  file    false  "Video file"
// @Success      201          {object}  object{status=string,message=string,data=object}
// @Failure      400          {object}  object{status=string,message=string,data=object}
// @Failure      401          {object}  object{error=string}
// @Failure      403          {object}  object{error=string}
// @Failure      500          {object}  object{status=string,message=string,data=object}
// @Router       /modules/{courseId} [post]
func (mac *ModuleAPIController) CreateModule(c *gin.Context) {
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

	// Handle form data for multipart/form-data
	title := c.PostForm("title")
	description := c.PostForm("description")

	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Title is required",
			"data":    nil,
		})
		return
	}

	// Handle file uploads
	var pdfURL, videoURL *string

	// Handle PDF file
	if file, header, err := c.Request.FormFile("pdf_content"); err == nil && header != nil {
		defer file.Close()
		pdfContent, err := mac.moduleService.SavePDF(header)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to save PDF: " + err.Error(),
				"data":    nil,
			})
			return
		}
		pdfURL = &pdfContent
	}

	// Handle Video file
	if file, header, err := c.Request.FormFile("video_content"); err == nil && header != nil {
		defer file.Close()
		videoContent, err := mac.moduleService.SaveVideo(header)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to save video: " + err.Error(),
				"data":    nil,
			})
			return
		}
		videoURL = &videoContent
	}

	// Create module
	createdModule, err := mac.moduleService.CreateModule(courseID, title, description, pdfURL, videoURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to create module",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Module created successfully",
		"data":    createdModule,
	})
}

// UpdateModule godoc
// @Summary      Update a module (Admin only)
// @Description  Update an existing module with new information
// @Tags         admin-modules
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id           path      string  true   "Module ID"
// @Param        title        formData  string  false  "Module title"
// @Param        description  formData  string  false  "Module description"
// @Param        pdf_file     formData  file    false  "PDF file"
// @Param        video_file   formData  file    false  "Video file"
// @Success      200          {object}  object{status=string,message=string,data=object}
// @Failure      400          {object}  object{status=string,message=string,data=object}
// @Failure      401          {object}  object{error=string}
// @Failure      403          {object}  object{error=string}
// @Failure      404          {object}  object{status=string,message=string,data=object}
// @Failure      500          {object}  object{status=string,message=string,data=object}
// @Router       /modules/update/{id} [put]
func (mac *ModuleAPIController) UpdateModule(c *gin.Context) {
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

	moduleID := c.Param("id")

	// Get existing module to preserve files if no new ones are provided
	existingModule, err := mac.moduleService.GetModuleByID(moduleID, userModel.ID, "admin")
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "Module not found",
			"data":    nil,
		})
		return
	}

	// Handle form data for multipart/form-data
	title := c.PostForm("title")
	description := c.PostForm("description")

	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Title is required",
			"data":    nil,
		})
		return
	}

	// Handle file uploads, preserving existing files if no new ones are provided
	var pdfURL, videoURL *string

	// Preserve existing PDF if available
	if existingPDF, ok := existingModule["pdf_content"].(string); ok && existingPDF != "" {
		pdfURL = &existingPDF
	}

	// Preserve existing video if available
	if existingVideo, ok := existingModule["video_content"].(string); ok && existingVideo != "" {
		videoURL = &existingVideo
	}

	// Handle PDF file upload (this will override the preserved PDF)
	if file, header, err := c.Request.FormFile("pdf_content"); err == nil && header != nil {
		defer file.Close()
		pdfContent, err := mac.moduleService.SavePDF(header)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to save PDF: " + err.Error(),
				"data":    nil,
			})
			return
		}
		pdfURL = &pdfContent
	}

	// Handle Video file upload (this will override the preserved video)
	if file, header, err := c.Request.FormFile("video_content"); err == nil && header != nil {
		defer file.Close()
		videoContent, err := mac.moduleService.SaveVideo(header)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to save video: " + err.Error(),
				"data":    nil,
			})
			return
		}
		videoURL = &videoContent
	}

	// Update module
	updatedModule, err := mac.moduleService.UpdateModule(moduleID, title, description, pdfURL, videoURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to update module",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Module updated successfully",
		"data":    updatedModule,
	})
}

// DeleteModule godoc
// @Summary      Delete a module (Admin only)
// @Description  Delete an existing module and all its associated data
// @Tags         admin-modules
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      string  true  "Module ID"
// @Success      200 {object}  object{status=string,message=string,data=object}
// @Failure      401 {object}  object{error=string}
// @Failure      403 {object}  object{error=string}
// @Failure      500 {object}  object{status=string,message=string,data=object}
// @Router       /modules/{id} [delete]
func (mac *ModuleAPIController) DeleteModule(c *gin.Context) {
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

	moduleID := c.Param("id")
	err := mac.moduleService.DeleteModule(moduleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to delete module",
			"data":    nil,
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// ReorderModules godoc
// @Summary      Reorder modules within a course (Admin only)
// @Description  Update the order of modules within a specific course
// @Tags         admin-modules
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        courseId  path      string  true   "Course ID"
// @Param        request   body      object  true   "Module order data" example({"module_orders":[{"module_id":"123","order":1},{"module_id":"456","order":2}]})
// @Success      200       {object}  object{status=string,message=string,data=object}
// @Failure      400       {object}  object{error=string}
// @Failure      401       {object}  object{error=string}
// @Failure      403       {object}  object{error=string}
// @Failure      500       {object}  object{status=string,message=string,data=object}
// @Router       /courses/{courseId}/modules/reorder [put]
func (mac *ModuleAPIController) ReorderModules(c *gin.Context) {
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

	var req struct {
		ModuleOrder []struct {
			ID    string `json:"id" binding:"required"`
			Order int    `json:"order" binding:"required"`
		} `json:"module_order" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// Reorder modules
	result, err := mac.moduleService.ReorderModules(courseID, req.ModuleOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to reorder modules",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Modules reordered successfully",
		"data":    result,
	})
}
