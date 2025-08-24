package services

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
	"yonatan/labpro/config"
	"yonatan/labpro/models"

	"gorm.io/gorm"
)

type ModuleService struct {
	db     *gorm.DB
	config *config.Config
}

func NewModuleService(db *gorm.DB, cfg *config.Config) *ModuleService {
	return &ModuleService{db: db, config: cfg}
}

func (ms *ModuleService) CreateModule(courseID, title, description string, pdfURL, videoURL *string) (*models.Module, error) {
	// Get the next order number for this course
	var maxOrder int
	ms.db.Model(&models.Module{}).Where("course_id = ?", courseID).Select("COALESCE(MAX(\"order\"), 0)").Scan(&maxOrder)

	module := models.Module{
		CourseID:     courseID,
		Title:        title,
		Description:  description,
		Order:        maxOrder + 1,
		PDFContent:   pdfURL,
		VideoContent: videoURL,
	}

	if err := ms.db.Create(&module).Error; err != nil {
		return nil, err
	}

	return &module, nil
}

func (ms *ModuleService) GetModules(courseID string, userID interface{}, page, limit int) ([]map[string]interface{}, map[string]interface{}, error) {
	var modules []models.Module
	var total int64

	db := ms.db.Model(&models.Module{}).Preload("Course").Where("course_id = ?", courseID)

	// Count total
	db.Count(&total)

	// Apply pagination
	offset := (page - 1) * limit
	if err := db.Order("\"order\" ASC").Offset(offset).Limit(limit).Find(&modules).Error; err != nil {
		return nil, nil, err
	}

	// Convert to response format
	result := make([]map[string]interface{}, len(modules))
	for i, module := range modules {
		isCompleted := false
		if userID != nil {
			// Check if user completed this module
			var progress models.UserModuleProgress
			err := ms.db.Where("user_id = ? AND module_id = ?", userID, module.ID).First(&progress).Error
			isCompleted = (err == nil && progress.IsCompleted)
		}

		courseInfo := map[string]interface{}{
			"id":         module.Course.ID,
			"title":      module.Course.Title,
			"instructor": module.Course.Instructor,
		}

		result[i] = map[string]interface{}{
			"id":            module.ID,
			"course_id":     module.CourseID,
			"course":        courseInfo,
			"title":         module.Title,
			"description":   module.Description,
			"order":         module.Order,
			"pdf_content":   module.PDFContent,
			"video_content": module.VideoContent,
			"is_completed":  isCompleted,
			"created_at":    module.CreatedAt.Format("Jan 2, 2006"),
			"updated_at":    module.UpdatedAt,
		}
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	prevPage := page - 1
	if prevPage < 1 {
		prevPage = 1
	}

	nextPage := page + 1
	if nextPage > totalPages {
		nextPage = totalPages
	}

	pagination := map[string]interface{}{
		"current_page": page,
		"total_pages":  totalPages,
		"total_items":  total,
		"prev_page":    prevPage,
		"next_page":    nextPage,
	}

	return result, pagination, nil
}

func (ms *ModuleService) GetModuleByID(id string, userID interface{}, userRole string) (map[string]interface{}, error) {
	var module models.Module
	if err := ms.db.First(&module, "id = ?", id).Error; err != nil {
		return nil, err
	}

	// Check if user has access to this module (purchased course or admin)
	if userRole != "admin" && userID != nil {
		hasAccess, err := ms.CheckCourseAccess(userID.(string), module.CourseID)
		if err != nil || !hasAccess {
			return nil, errors.New("access denied")
		}
	}

	isCompleted := false
	if userID != nil && userRole != "admin" {
		// Check if user completed this module
		var progress models.UserModuleProgress
		err := ms.db.Where("user_id = ? AND module_id = ?", userID, module.ID).First(&progress).Error
		isCompleted = (err == nil && progress.IsCompleted)
	}

	result := map[string]interface{}{
		"id":            module.ID,
		"course_id":     module.CourseID,
		"title":         module.Title,
		"description":   module.Description,
		"order":         module.Order,
		"pdf_content":   module.PDFContent,
		"video_content": module.VideoContent,
		"is_completed":  isCompleted,
		"created_at":    module.CreatedAt,
		"updated_at":    module.UpdatedAt,
	}

	return result, nil
}

func (ms *ModuleService) UpdateModule(id, title, description string, pdfURL, videoURL *string) (*models.Module, error) {
	var module models.Module
	if err := ms.db.First(&module, "id = ?", id).Error; err != nil {
		return nil, err
	}

	module.Title = title
	module.Description = description

	if pdfURL != nil {
		module.PDFContent = pdfURL
	}
	if videoURL != nil {
		module.VideoContent = videoURL
	}

	if err := ms.db.Save(&module).Error; err != nil {
		return nil, err
	}

	return &module, nil
}

func (ms *ModuleService) DeleteModule(id string) error {
	// Delete module progress records
	if err := ms.db.Where("module_id = ?", id).Delete(&models.UserModuleProgress{}).Error; err != nil {
		return err
	}

	// Delete the module
	if err := ms.db.Delete(&models.Module{}, "id = ?", id).Error; err != nil {
		return err
	}

	return nil
}

func (ms *ModuleService) ReorderModules(courseID string, moduleOrder []struct {
	ID    string `json:"id" binding:"required"`
	Order int    `json:"order" binding:"required"`
}) (map[string]interface{}, error) {
	// Start transaction
	tx := ms.db.Begin()

	for _, item := range moduleOrder {
		if err := tx.Model(&models.Module{}).Where("id = ? AND course_id = ?", item.ID, courseID).Update("\"order\"", item.Order).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	tx.Commit()

	result := map[string]interface{}{
		"module_order": moduleOrder,
	}

	return result, nil
}

func (ms *ModuleService) CompleteModule(moduleID, userID string) (map[string]interface{}, error) {
	// Check if module exists and user has access
	var module models.Module
	if err := ms.db.First(&module, "id = ?", moduleID).Error; err != nil {
		return nil, errors.New("module not found")
	}

	// Check if user purchased the course
	hasAccess, err := ms.CheckCourseAccess(userID, module.CourseID)
	if err != nil || !hasAccess {
		return nil, errors.New("access denied. Course not purchased")
	}

	// Create or update user module progress
	var progress models.UserModuleProgress
	err = ms.db.Where("user_id = ? AND module_id = ?", userID, moduleID).First(&progress).Error
	if err == gorm.ErrRecordNotFound {
		// Create new progress record
		progress = models.UserModuleProgress{
			UserID:      userID,
			ModuleID:    moduleID,
			IsCompleted: true,
		}
		if err := ms.db.Create(&progress).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		// Update existing record
		progress.IsCompleted = true
		if err := ms.db.Save(&progress).Error; err != nil {
			return nil, err
		}
	}

	// Calculate course progress
	var totalModules, completedModules int64
	ms.db.Model(&models.Module{}).Where("course_id = ?", module.CourseID).Count(&totalModules)
	ms.db.Model(&models.UserModuleProgress{}).
		Joins("JOIN modules ON user_module_progress.module_id = modules.id").
		Where("user_module_progress.user_id = ? AND modules.course_id = ? AND user_module_progress.is_completed = ?",
			userID, module.CourseID, true).Count(&completedModules)

	percentage := float64(0)
	if totalModules > 0 {
		percentage = float64(completedModules) / float64(totalModules) * 100
	}

	result := map[string]interface{}{
		"module_id":    moduleID,
		"is_completed": true,
		"course_progress": map[string]interface{}{
			"total_modules":     totalModules,
			"completed_modules": completedModules,
			"percentage":        percentage,
		},
		"certificate_url": nil,
	}

	// If 100% complete, generate certificate
	if percentage >= 100 {
		certificateURL, err := ms.generateCertificate(userID, module.CourseID)
		if err == nil {
			result["certificate_url"] = certificateURL
		}
	}

	return result, nil
}

func (ms *ModuleService) CheckCourseAccess(userID, courseID string) (bool, error) {
	var userCourse models.UserCourse
	err := ms.db.Where("user_id = ? AND course_id = ?", userID, courseID).First(&userCourse).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (ms *ModuleService) SavePDF(file *multipart.FileHeader) (string, error) {
	log.Printf("SavePDF: Starting to save PDF file: %s (size: %d bytes)", file.Filename, file.Size)

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	log.Printf("SavePDF: File content type: %s", contentType)
	if contentType != "application/pdf" {
		log.Printf("SavePDF: Invalid file type: %s", contentType)
		return "", errors.New("invalid file type: only PDF files are allowed")
	}

	// Validate file size (10MB limit)
	if file.Size > 10*1024*1024 {
		log.Printf("SavePDF: File too large: %d bytes", file.Size)
		return "", errors.New("file size too large: maximum 10MB allowed for PDF files")
	}

	// Create uploads directory if it doesn't exist
	uploadDir := "./uploads/pdfs"
	log.Printf("SavePDF: Creating upload directory: %s", uploadDir)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("SavePDF: Failed to create upload directory: %v", err)
		return "", fmt.Errorf("failed to create upload directory: %v", err)
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
	filePath := filepath.Join(uploadDir, filename)
	log.Printf("SavePDF: Generated file path: %s", filePath)

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		log.Printf("SavePDF: Failed to open uploaded file: %v", err)
		return "", fmt.Errorf("failed to open uploaded file: %v", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("SavePDF: Failed to create destination file: %v", err)
		return "", fmt.Errorf("failed to create destination file: %v", err)
	}
	defer dst.Close()

	// Copy file content
	bytesWritten, err := io.Copy(dst, src)
	if err != nil {
		// Clean up created file on error
		os.Remove(filePath)
		log.Printf("SavePDF: Failed to copy file content: %v", err)
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	log.Printf("SavePDF: Successfully wrote %d bytes to %s", bytesWritten, filePath)

	// Return complete URL including base URL from config
	resultURL := fmt.Sprintf("%s/uploads/pdfs/%s", ms.config.BaseURL, filename)
	log.Printf("SavePDF: Generated URL: %s", resultURL)
	return resultURL, nil
}

func (ms *ModuleService) SaveVideo(file *multipart.FileHeader) (string, error) {
	log.Printf("SaveVideo: Starting to save video file: %s (size: %d bytes)", file.Filename, file.Size)

	// Validate file type (basic check)
	contentType := file.Header.Get("Content-Type")
	log.Printf("SaveVideo: File content type: %s", contentType)
	validVideoTypes := []string{
		"video/mp4",
		"video/avi",
		"video/mov",
		"video/quicktime",
		"video/x-msvideo",
		"video/webm",
		"video/ogg",
	}

	isValidType := false
	for _, validType := range validVideoTypes {
		if contentType == validType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		log.Printf("SaveVideo: Invalid file type: %s", contentType)
		return "", errors.New("invalid file type: only video files (MP4, AVI, MOV, WebM, OGG) are allowed")
	}

	// Validate file size (100MB limit)
	if file.Size > 100*1024*1024 {
		log.Printf("SaveVideo: File too large: %d bytes", file.Size)
		return "", errors.New("file size too large: maximum 100MB allowed for video files")
	}

	// Create uploads directory if it doesn't exist
	uploadDir := "./uploads/videos"
	log.Printf("SaveVideo: Creating upload directory: %s", uploadDir)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("SaveVideo: Failed to create upload directory: %v", err)
		return "", fmt.Errorf("failed to create upload directory: %v", err)
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
	filePath := filepath.Join(uploadDir, filename)
	log.Printf("SaveVideo: Generated file path: %s", filePath)

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		log.Printf("SaveVideo: Failed to open uploaded file: %v", err)
		return "", fmt.Errorf("failed to open uploaded file: %v", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("SaveVideo: Failed to create destination file: %v", err)
		return "", fmt.Errorf("failed to create destination file: %v", err)
	}
	defer dst.Close()

	// Copy file content
	bytesWritten, err := io.Copy(dst, src)
	if err != nil {
		// Clean up created file on error
		os.Remove(filePath)
		log.Printf("SaveVideo: Failed to copy file content: %v", err)
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	log.Printf("SaveVideo: Successfully wrote %d bytes to %s", bytesWritten, filePath)

	// Return complete URL including base URL from config
	resultURL := fmt.Sprintf("%s/uploads/videos/%s", ms.config.BaseURL, filename)
	log.Printf("SaveVideo: Generated URL: %s", resultURL)
	return resultURL, nil
}

func (ms *ModuleService) generateCertificate(userID, courseID string) (string, error) {
	// Get user and course info
	var user models.User
	var course models.Course

	if err := ms.db.First(&user, "id = ?", userID).Error; err != nil {
		return "", err
	}

	if err := ms.db.First(&course, "id = ?", courseID).Error; err != nil {
		return "", err
	}

	// Create certificates directory if it doesn't exist
	uploadDir := "./uploads/certificates"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	// Generate certificate content (simple text format for now)
	certificateContent := fmt.Sprintf(
		"CERTIFICATE OF COMPLETION\n\n"+
			"This is to certify that\n\n"+
			"%s %s (%s)\n\n"+
			"has successfully completed the course\n\n"+
			"%s\n\n"+
			"Instructor: %s\n\n"+
			"Date of Completion: %s\n",
		user.FirstName, user.LastName, user.Username,
		course.Title,
		course.Instructor,
		time.Now().Format("January 2, 2006"),
	)

	// Save certificate
	filename := fmt.Sprintf("certificate_%s_%s_%d.txt", userID, courseID, time.Now().Unix())
	filepath := filepath.Join(uploadDir, filename)

	if err := os.WriteFile(filepath, []byte(certificateContent), 0644); err != nil {
		return "", err
	}

	// Return relative URL
	return fmt.Sprintf("/uploads/certificates/%s", filename), nil
}
