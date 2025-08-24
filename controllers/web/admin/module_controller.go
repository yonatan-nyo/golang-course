package admin

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"yonatan/labpro/models"
	"yonatan/labpro/services"

	"github.com/gin-gonic/gin"
)

type ModuleController struct {
	moduleService *services.ModuleService
	courseService *services.CourseService
}

func NewModuleController(moduleService *services.ModuleService, courseService *services.CourseService) *ModuleController {
	return &ModuleController{
		moduleService: moduleService,
		courseService: courseService,
	}
}

func (mc *ModuleController) ShowCourseModulesPage(c *gin.Context) {
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

	// Get course ID from URL parameter
	courseID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Get modules from service (pass nil for userID since admin doesn't need completion status)
	modules, pagination, err := mc.moduleService.GetModules(courseID, nil, page, limit)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "modules.html", gin.H{
			"Title": "Course Module Management",
			"User":  userModel,
			"Error": "Failed to fetch modules",
		})
		return
	}

	c.HTML(http.StatusOK, "modules.html", gin.H{
		"Title":      "Course Module Management",
		"User":       userModel,
		"Modules":    modules,
		"Pagination": pagination,
		"CourseID":   courseID,
	})
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

	// Get all courses for dropdown
	courses, _, err := mc.courseService.GetCourses("", 1, 1000, userModel.ID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "module-create.html", gin.H{
			"Title":    "Create Module",
			"User":     userModel,
			"CourseID": courseID,
			"Error":    "Failed to fetch courses",
		})
		return
	}

	c.HTML(http.StatusOK, "module-create.html", gin.H{
		"Title":    "Create Module",
		"User":     userModel,
		"CourseID": courseID,
		"Courses":  courses,
	})
}

func (mc *ModuleController) ShowCreateModulePageForCourse(c *gin.Context) {
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

	// Get the specific course for display
	course, err := mc.courseService.GetCourseByID(courseID, userModel.ID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "module-create.html", gin.H{
			"Title": "Create Module",
			"User":  userModel,
			"Error": "Failed to fetch course details",
		})
		return
	}

	c.HTML(http.StatusOK, "module-create.html", gin.H{
		"Title":  "Create Module",
		"User":   userModel,
		"Course": course,
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
	log.Println("=== HandleCreateModule: Starting request processing ===")

	user, exists := c.Get("user")
	if !exists {
		log.Println("HandleCreateModule: User not found in context, redirecting to login")
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	userModel := user.(models.User)
	if !userModel.IsAdmin {
		log.Printf("HandleCreateModule: User %s is not admin, redirecting to dashboard", userModel.Email)
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}

	log.Printf("HandleCreateModule: Admin user %s processing request", userModel.Email)

	// Log request details
	log.Printf("HandleCreateModule: Request method: %s", c.Request.Method)
	log.Printf("HandleCreateModule: Content-Type: %s", c.Request.Header.Get("Content-Type"))
	log.Printf("HandleCreateModule: Content-Length: %d", c.Request.ContentLength)

	// Parse multipart form first
	err := c.Request.ParseMultipartForm(100 << 20) // 100MB max memory
	if err != nil {
		log.Printf("HandleCreateModule: Failed to parse multipart form: %v", err)
		c.HTML(http.StatusBadRequest, "module-create.html", gin.H{
			"Title": "Create Module",
			"User":  userModel,
			"Error": "Failed to parse form data: " + err.Error(),
		})
		return
	}
	log.Printf("HandleCreateModule: Multipart form parsed successfully")

	// Handle form submission
	courseID := c.PostForm("course_id")
	title := c.PostForm("title")
	description := c.PostForm("description")

	log.Printf("HandleCreateModule: Form data - courseID: %s, title: %s, description length: %d", courseID, title, len(description))

	// Validate required fields
	if courseID == "" || title == "" {
		log.Println("HandleCreateModule: Validation failed - missing required fields")
		c.HTML(http.StatusBadRequest, "module-create.html", gin.H{
			"Title":    "Create Module",
			"User":     userModel,
			"CourseID": courseID,
			"Error":    "Course ID and title are required",
		})
		return
	}

	// Handle file uploads
	var pdfURL, videoURL *string

	// Log multipart form info
	if c.Request.MultipartForm != nil {
		log.Printf("HandleCreateModule: MultipartForm exists with %d files", len(c.Request.MultipartForm.File))
		for field, files := range c.Request.MultipartForm.File {
			log.Printf("HandleCreateModule: Field %s has %d files", field, len(files))
			for i, file := range files {
				log.Printf("HandleCreateModule: File %d - Name: %s, Size: %d, Header: %v", i, file.Filename, file.Size, file.Header)
			}
		}
	} else {
		log.Println("HandleCreateModule: No MultipartForm found")
	}

	// Handle PDF file upload
	log.Println("HandleCreateModule: Attempting to get PDF file")
	pdfFile, pdfHeader, pdfErr := c.Request.FormFile("pdf_file")
	if pdfErr != nil {
		log.Printf("HandleCreateModule: PDF file error: %v", pdfErr)
	} else {
		log.Printf("HandleCreateModule: PDF file found - Name: %s, Size: %d", pdfHeader.Filename, pdfHeader.Size)
	}

	if pdfErr == nil && pdfHeader != nil && pdfHeader.Size > 0 {
		defer pdfFile.Close()
		log.Printf("HandleCreateModule: Processing PDF file: %s (size: %d bytes)", pdfHeader.Filename, pdfHeader.Size)

		// Validate PDF file size (10MB limit)
		if pdfHeader.Size > 10*1024*1024 {
			log.Printf("HandleCreateModule: PDF file too large: %d bytes", pdfHeader.Size)
			c.HTML(http.StatusBadRequest, "module-create.html", gin.H{
				"Title":    "Create Module",
				"User":     userModel,
				"CourseID": courseID,
				"Error":    "PDF file size must be less than 10MB",
			})
			return
		}

		log.Println("HandleCreateModule: Calling SavePDF service method")
		pdfContent, err := mc.moduleService.SavePDF(pdfHeader)
		if err != nil {
			log.Printf("HandleCreateModule: Failed to save PDF: %v", err)
			c.HTML(http.StatusInternalServerError, "module-create.html", gin.H{
				"Title":    "Create Module",
				"User":     userModel,
				"CourseID": courseID,
				"Error":    "Failed to save PDF: " + err.Error(),
			})
			return
		}
		log.Printf("HandleCreateModule: PDF saved successfully at: %s", pdfContent)
		pdfURL = &pdfContent
	} else {
		log.Println("HandleCreateModule: No valid PDF file provided")
	}

	// Handle Video file upload
	log.Println("HandleCreateModule: Attempting to get video file")
	videoFile, videoHeader, videoErr := c.Request.FormFile("video_file")
	if videoErr != nil {
		log.Printf("HandleCreateModule: Video file error: %v", videoErr)
	} else {
		log.Printf("HandleCreateModule: Video file found - Name: %s, Size: %d", videoHeader.Filename, videoHeader.Size)
	}

	if videoErr == nil && videoHeader != nil && videoHeader.Size > 0 {
		defer videoFile.Close()
		log.Printf("HandleCreateModule: Processing video file: %s (size: %d bytes)", videoHeader.Filename, videoHeader.Size)

		// Validate video file size (100MB limit)
		if videoHeader.Size > 100*1024*1024 {
			log.Printf("HandleCreateModule: Video file too large: %d bytes", videoHeader.Size)
			c.HTML(http.StatusBadRequest, "module-create.html", gin.H{
				"Title":    "Create Module",
				"User":     userModel,
				"CourseID": courseID,
				"Error":    "Video file size must be less than 100MB",
			})
			return
		}

		log.Println("HandleCreateModule: Calling SaveVideo service method")
		videoContent, err := mc.moduleService.SaveVideo(videoHeader)
		if err != nil {
			log.Printf("HandleCreateModule: Failed to save video: %v", err)
			c.HTML(http.StatusInternalServerError, "module-create.html", gin.H{
				"Title":    "Create Module",
				"User":     userModel,
				"CourseID": courseID,
				"Error":    "Failed to save video: " + err.Error(),
			})
			return
		}
		log.Printf("HandleCreateModule: Video saved successfully at: %s", videoContent)
		videoURL = &videoContent
	} else {
		log.Println("HandleCreateModule: No valid video file provided")
	}

	// Log final URLs
	if pdfURL != nil {
		log.Printf("HandleCreateModule: Final PDF URL: %s", *pdfURL)
	} else {
		log.Println("HandleCreateModule: No PDF URL set")
	}
	if videoURL != nil {
		log.Printf("HandleCreateModule: Final Video URL: %s", *videoURL)
	} else {
		log.Println("HandleCreateModule: No Video URL set")
	}

	// Create module using service method signature
	log.Printf("HandleCreateModule: Creating module with courseID: %s, title: %s", courseID, title)
	createdModule, err := mc.moduleService.CreateModule(courseID, title, description, pdfURL, videoURL)
	if err != nil {
		log.Printf("HandleCreateModule: Failed to create module: %v", err)
		c.HTML(http.StatusInternalServerError, "module-create.html", gin.H{
			"Title":    "Create Module",
			"User":     userModel,
			"CourseID": courseID,
			"Error":    "Failed to create module: " + err.Error(),
		})
		return
	}

	log.Printf("HandleCreateModule: Module created successfully with ID: %s", createdModule.ID)
	redirectURL := fmt.Sprintf("/admin/modules?course_id=%s&success=Module created successfully&id=%s", courseID, createdModule.ID)
	log.Printf("HandleCreateModule: Redirecting to: %s", redirectURL)
	c.Redirect(http.StatusFound, redirectURL)
}

func (mc *ModuleController) HandleUpdateModule(c *gin.Context) {
	log.Println("=== HandleUpdateModule: Starting request processing ===")

	user, exists := c.Get("user")
	if !exists {
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	userModel := user.(models.User)
	if !userModel.IsAdmin {
		log.Printf("HandleUpdateModule: User %s is not admin, redirecting to dashboard", userModel.Email)
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}

	moduleID := c.Param("id")
	log.Printf("HandleUpdateModule: Updating module with ID: %s", moduleID)

	// Get existing module
	existingModule, err := mc.moduleService.GetModuleByID(moduleID, nil, "admin")
	if err != nil {
		log.Printf("HandleUpdateModule: Failed to get existing module: %v", err)
		c.HTML(http.StatusNotFound, "module-edit.html", gin.H{
			"Title": "Edit Module",
			"User":  userModel,
			"Error": "Module not found",
		})
		return
	}

	log.Printf("HandleUpdateModule: Found existing module: %s", existingModule["title"])

	// Parse multipart form first
	err = c.Request.ParseMultipartForm(100 << 20) // 100MB max memory
	if err != nil {
		log.Printf("HandleUpdateModule: Failed to parse multipart form: %v", err)
		c.HTML(http.StatusBadRequest, "module-edit.html", gin.H{
			"Title":  "Edit Module",
			"User":   userModel,
			"Module": existingModule,
			"Error":  "Failed to parse form data: " + err.Error(),
		})
		return
	}
	log.Printf("HandleUpdateModule: Multipart form parsed successfully")

	// Handle form submission
	title := c.PostForm("title")
	description := c.PostForm("description")

	log.Printf("HandleUpdateModule: Form data - title: %s, description length: %d", title, len(description))

	// Validate required fields
	if title == "" {
		log.Println("HandleUpdateModule: Title validation failed")
		c.HTML(http.StatusBadRequest, "module-edit.html", gin.H{
			"Title":  "Edit Module",
			"User":   userModel,
			"Module": existingModule,
			"Error":  "Title is required",
		})
		return
	}

	// Handle file uploads - start with existing URLs
	var pdfURL, videoURL *string

	// Get existing URLs
	if existingPDF, ok := existingModule["pdf_content"]; ok && existingPDF != nil {
		if pdfStr, ok := existingPDF.(string); ok && pdfStr != "" {
			pdfURL = &pdfStr
		}
	}
	if existingVideo, ok := existingModule["video_content"]; ok && existingVideo != nil {
		if videoStr, ok := existingVideo.(string); ok && videoStr != "" {
			videoURL = &videoStr
		}
	}

	// Handle PDF file upload (if new file is provided)
	if c.Request.MultipartForm != nil {
		log.Printf("HandleUpdateModule: MultipartForm exists with %d files", len(c.Request.MultipartForm.File))
		for field, files := range c.Request.MultipartForm.File {
			log.Printf("HandleUpdateModule: Field %s has %d files", field, len(files))
		}
	}

	pdfFile, pdfHeader, pdfErr := c.Request.FormFile("pdf_file")
	if pdfErr == nil && pdfHeader != nil && pdfHeader.Size > 0 {
		defer pdfFile.Close()
		log.Printf("HandleUpdateModule: Processing new PDF file: %s (size: %d bytes)", pdfHeader.Filename, pdfHeader.Size)

		// Validate PDF file size (10MB limit)
		if pdfHeader.Size > 10*1024*1024 {
			log.Printf("HandleUpdateModule: PDF file too large: %d bytes", pdfHeader.Size)
			c.HTML(http.StatusBadRequest, "module-edit.html", gin.H{
				"Title":  "Edit Module",
				"User":   userModel,
				"Module": existingModule,
				"Error":  "PDF file size must be less than 10MB",
			})
			return
		}

		log.Println("HandleUpdateModule: Calling SavePDF service method")
		pdfContent, err := mc.moduleService.SavePDF(pdfHeader)
		if err != nil {
			log.Printf("HandleUpdateModule: Failed to save PDF: %v", err)
			c.HTML(http.StatusInternalServerError, "module-edit.html", gin.H{
				"Title":  "Edit Module",
				"User":   userModel,
				"Module": existingModule,
				"Error":  "Failed to save PDF: " + err.Error(),
			})
			return
		}
		log.Printf("HandleUpdateModule: PDF saved successfully at: %s", pdfContent)
		pdfURL = &pdfContent
	} else {
		log.Printf("HandleUpdateModule: No new PDF file provided (error: %v)", pdfErr)
	}

	// Handle Video file upload (if new file is provided)
	videoFile, videoHeader, videoErr := c.Request.FormFile("video_file")
	if videoErr == nil && videoHeader != nil && videoHeader.Size > 0 {
		defer videoFile.Close()
		log.Printf("HandleUpdateModule: Processing new video file: %s (size: %d bytes)", videoHeader.Filename, videoHeader.Size)

		// Validate video file size (100MB limit)
		if videoHeader.Size > 100*1024*1024 {
			log.Printf("HandleUpdateModule: Video file too large: %d bytes", videoHeader.Size)
			c.HTML(http.StatusBadRequest, "module-edit.html", gin.H{
				"Title":  "Edit Module",
				"User":   userModel,
				"Module": existingModule,
				"Error":  "Video file size must be less than 100MB",
			})
			return
		}

		log.Println("HandleUpdateModule: Calling SaveVideo service method")
		videoContent, err := mc.moduleService.SaveVideo(videoHeader)
		if err != nil {
			log.Printf("HandleUpdateModule: Failed to save video: %v", err)
			c.HTML(http.StatusInternalServerError, "module-edit.html", gin.H{
				"Title":  "Edit Module",
				"User":   userModel,
				"Module": existingModule,
				"Error":  "Failed to save video: " + err.Error(),
			})
			return
		}
		log.Printf("HandleUpdateModule: Video saved successfully at: %s", videoContent)
		videoURL = &videoContent
	} else {
		log.Printf("HandleUpdateModule: No new video file provided (error: %v)", videoErr)
	}

	// Log final URLs
	if pdfURL != nil {
		log.Printf("HandleUpdateModule: Final PDF URL: %s", *pdfURL)
	} else {
		log.Println("HandleUpdateModule: No PDF URL set")
	}
	if videoURL != nil {
		log.Printf("HandleUpdateModule: Final Video URL: %s", *videoURL)
	} else {
		log.Println("HandleUpdateModule: No Video URL set")
	}

	// Update module using service method signature
	log.Printf("HandleUpdateModule: Updating module with ID: %s, title: %s", moduleID, title)
	_, err = mc.moduleService.UpdateModule(moduleID, title, description, pdfURL, videoURL)
	if err != nil {
		log.Printf("HandleUpdateModule: Failed to update module: %v", err)
		c.HTML(http.StatusInternalServerError, "module-edit.html", gin.H{
			"Title":  "Edit Module",
			"User":   userModel,
			"Module": existingModule,
			"Error":  "Failed to update module: " + err.Error(),
		})
		return
	}

	log.Printf("HandleUpdateModule: Module updated successfully")

	// Get course ID for redirect
	courseID := ""
	if courseIDValue, exists := existingModule["course_id"]; exists && courseIDValue != nil {
		courseID = courseIDValue.(string)
	}

	redirectURL := "/admin/modules?success=Module updated successfully&id=" + moduleID
	if courseID != "" {
		redirectURL += "&course_id=" + courseID
	}

	log.Printf("HandleUpdateModule: Redirecting to: %s", redirectURL)
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
