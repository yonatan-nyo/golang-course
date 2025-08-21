package controllers

import (
	"net/http"
	"strconv"
	"yonatan/labpro/services"

	"github.com/gin-gonic/gin"
)

type ModuleController struct {
	moduleService *services.ModuleService
}

func NewModuleController(moduleService *services.ModuleService) *ModuleController {
	return &ModuleController{moduleService: moduleService}
}

func (mc *ModuleController) CreateModule(c *gin.Context) {
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

	courseID := c.Param("courseId")
	title := c.PostForm("title")
	description := c.PostForm("description")

	if title == "" || description == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Title and description are required",
			"data":    nil,
		})
		return
	}

	// Handle file uploads
	var pdfURL, videoURL *string

	if pdfFile, err := c.FormFile("pdf_content"); err == nil {
		url, err := mc.moduleService.SavePDF(pdfFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to save PDF: " + err.Error(),
				"data":    nil,
			})
			return
		}
		pdfURL = &url
	}

	if videoFile, err := c.FormFile("video_content"); err == nil {
		url, err := mc.moduleService.SaveVideo(videoFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to save video: " + err.Error(),
				"data":    nil,
			})
			return
		}
		videoURL = &url
	}

	module, err := mc.moduleService.CreateModule(courseID, title, description, pdfURL, videoURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to create module: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Module created successfully",
		"data":    module,
	})
}

func (mc *ModuleController) GetModules(c *gin.Context) {
	courseID := c.Param("courseId")
	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")

	// Get query parameters
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

	// Check if user has access to this course (purchased or admin)
	if userRole != "admin" {
		hasAccess, err := mc.moduleService.CheckCourseAccess(userID.(string), courseID)
		if err != nil || !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{
				"status":  "error",
				"message": "Access denied. Course not purchased.",
				"data":    nil,
			})
			return
		}
	}

	modules, pagination, err := mc.moduleService.GetModules(courseID, userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get modules: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"message":    "Modules retrieved successfully",
		"data":       modules,
		"pagination": pagination,
	})
}

func (mc *ModuleController) GetModule(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")

	module, err := mc.moduleService.GetModuleByID(id, userID, userRole.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "Module not found or access denied",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Module retrieved successfully",
		"data":    module,
	})
}

func (mc *ModuleController) UpdateModule(c *gin.Context) {
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
	title := c.PostForm("title")
	description := c.PostForm("description")

	if title == "" || description == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Title and description are required",
			"data":    nil,
		})
		return
	}

	// Handle file uploads
	var pdfURL, videoURL *string

	if pdfFile, err := c.FormFile("pdf_content"); err == nil {
		url, err := mc.moduleService.SavePDF(pdfFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to save PDF: " + err.Error(),
				"data":    nil,
			})
			return
		}
		pdfURL = &url
	}

	if videoFile, err := c.FormFile("video_content"); err == nil {
		url, err := mc.moduleService.SaveVideo(videoFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to save video: " + err.Error(),
				"data":    nil,
			})
			return
		}
		videoURL = &url
	}

	module, err := mc.moduleService.UpdateModule(id, title, description, pdfURL, videoURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to update module: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Module updated successfully",
		"data":    module,
	})
}

func (mc *ModuleController) DeleteModule(c *gin.Context) {
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

	err := mc.moduleService.DeleteModule(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to delete module: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.Status(http.StatusNoContent)
}

func (mc *ModuleController) ReorderModules(c *gin.Context) {
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
			"message": "Invalid input: " + err.Error(),
			"data":    nil,
		})
		return
	}

	result, err := mc.moduleService.ReorderModules(courseID, req.ModuleOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to reorder modules: " + err.Error(),
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

func (mc *ModuleController) CompleteModule(c *gin.Context) {
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

	result, err := mc.moduleService.CompleteModule(id, userID.(string))
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
		"message": "Module completed successfully",
		"data":    result,
	})
}
