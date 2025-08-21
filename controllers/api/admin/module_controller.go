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
