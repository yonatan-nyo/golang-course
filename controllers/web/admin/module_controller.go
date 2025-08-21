package admin

import (
	"net/http"
	"strconv"
	"yonatan/labpro/models"
	"yonatan/labpro/services"

	"github.com/gin-gonic/gin"
)

type ModuleController struct {
	moduleService *services.ModuleService
}

func NewModuleController(moduleService *services.ModuleService) *ModuleController {
	return &ModuleController{
		moduleService: moduleService,
	}
}

func (mc *ModuleController) ShowModulesPage(c *gin.Context) {
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

	// Get query parameters
	courseID := c.Query("course_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Get modules from service (pass nil for userID since admin doesn't need completion status)
	modules, pagination, err := mc.moduleService.GetModules(courseID, nil, page, limit)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "modules.html", gin.H{
			"Title": "Module Management",
			"User":  userModel,
			"Error": "Failed to fetch modules",
		})
		return
	}

	c.HTML(http.StatusOK, "modules.html", gin.H{
		"Title":      "Module Management",
		"User":       userModel,
		"Modules":    modules,
		"Pagination": pagination,
		"CourseID":   courseID,
	})
}

func (mc *ModuleController) ShowCreateModulePage(c *gin.Context) {
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

	courseID := c.Query("course_id")

	c.HTML(http.StatusOK, "module-create.html", gin.H{
		"Title":    "Create Module",
		"User":     userModel,
		"CourseID": courseID,
	})
}

func (mc *ModuleController) ShowEditModulePage(c *gin.Context) {
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

	moduleID := c.Param("id")
	module, err := mc.moduleService.GetModuleByID(moduleID, nil, "admin")
	if err != nil {
		c.HTML(http.StatusNotFound, "module-edit.html", gin.H{
			"Title": "Edit Module",
			"User":  userModel,
			"Error": "Module not found",
		})
		return
	}

	c.HTML(http.StatusOK, "module-edit.html", gin.H{
		"Title":  "Edit Module",
		"User":   userModel,
		"Module": module,
	})
}

func (mc *ModuleController) HandleCreateModule(c *gin.Context) {
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
	courseID := c.PostForm("course_id")
	title := c.PostForm("title")
	description := c.PostForm("description")
	pdfContent := c.PostForm("pdf_content")
	videoContent := c.PostForm("video_content")

	// Validate required fields
	if courseID == "" || title == "" {
		c.HTML(http.StatusBadRequest, "module-create.html", gin.H{
			"Title":    "Create Module",
			"User":     userModel,
			"CourseID": courseID,
			"Error":    "Course ID and title are required",
		})
		return
	}

	// Convert empty strings to nil pointers
	var pdfURL, videoURL *string
	if pdfContent != "" {
		pdfURL = &pdfContent
	}
	if videoContent != "" {
		videoURL = &videoContent
	}

	// Create module using service method signature
	createdModule, err := mc.moduleService.CreateModule(courseID, title, description, pdfURL, videoURL)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "module-create.html", gin.H{
			"Title":    "Create Module",
			"User":     userModel,
			"CourseID": courseID,
			"Error":    "Failed to create module: " + err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin/modules?course_id="+courseID+"&success=Module created successfully&id="+createdModule.ID)
}

func (mc *ModuleController) HandleUpdateModule(c *gin.Context) {
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

	moduleID := c.Param("id")

	// Get existing module
	existingModule, err := mc.moduleService.GetModuleByID(moduleID, nil, "admin")
	if err != nil {
		c.HTML(http.StatusNotFound, "module-edit.html", gin.H{
			"Title": "Edit Module",
			"User":  userModel,
			"Error": "Module not found",
		})
		return
	}

	// Handle form submission
	title := c.PostForm("title")
	description := c.PostForm("description")
	pdfContent := c.PostForm("pdf_content")
	videoContent := c.PostForm("video_content")

	// Validate required fields
	if title == "" {
		c.HTML(http.StatusBadRequest, "module-edit.html", gin.H{
			"Title":  "Edit Module",
			"User":   userModel,
			"Module": existingModule,
			"Error":  "Title is required",
		})
		return
	}

	// Convert empty strings to nil pointers
	var pdfURL, videoURL *string
	if pdfContent != "" {
		pdfURL = &pdfContent
	}
	if videoContent != "" {
		videoURL = &videoContent
	}

	// Update module using service method signature
	_, err = mc.moduleService.UpdateModule(moduleID, title, description, pdfURL, videoURL)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "module-edit.html", gin.H{
			"Title":  "Edit Module",
			"User":   userModel,
			"Module": existingModule,
			"Error":  "Failed to update module: " + err.Error(),
		})
		return
	}

	// Get course ID for redirect
	courseID := ""
	if courseIDValue, exists := existingModule["course_id"]; exists && courseIDValue != nil {
		courseID = courseIDValue.(string)
	}

	redirectURL := "/admin/modules?success=Module updated successfully&id=" + moduleID
	if courseID != "" {
		redirectURL += "&course_id=" + courseID
	}

	c.Redirect(http.StatusFound, redirectURL)
}

func (mc *ModuleController) HandleDeleteModule(c *gin.Context) {
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
	err := mc.moduleService.DeleteModule(moduleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete module"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Module deleted successfully"})
}
