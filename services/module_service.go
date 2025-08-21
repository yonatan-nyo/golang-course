package services

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
	"yonatan/labpro/models"

	"gorm.io/gorm"
)

type ModuleService struct {
	db *gorm.DB
}

func NewModuleService(db *gorm.DB) *ModuleService {
	return &ModuleService{db: db}
}

func (ms *ModuleService) CreateModule(courseID, title, description string, pdfURL, videoURL *string) (*models.Module, error) {
	// Get the next order number for this course
	var maxOrder int
	ms.db.Model(&models.Module{}).Where("course_id = ?", courseID).Select("COALESCE(MAX(order), 0)").Scan(&maxOrder)

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

	db := ms.db.Model(&models.Module{}).Where("course_id = ?", courseID)

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

		result[i] = map[string]interface{}{
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
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	pagination := map[string]interface{}{
		"current_page": page,
		"total_pages":  totalPages,
		"total_items":  total,
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
		if err := tx.Model(&models.Module{}).Where("id = ? AND course_id = ?", item.ID, courseID).Update("order", item.Order).Error; err != nil {
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
	// Create uploads directory if it doesn't exist
	uploadDir := "./uploads/pdfs"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
	filepath := filepath.Join(uploadDir, filename)

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy file content
	if _, err = io.Copy(dst, src); err != nil {
		return "", err
	}

	// Return relative URL
	return fmt.Sprintf("/uploads/pdfs/%s", filename), nil
}

func (ms *ModuleService) SaveVideo(file *multipart.FileHeader) (string, error) {
	// Create uploads directory if it doesn't exist
	uploadDir := "./uploads/videos"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
	filepath := filepath.Join(uploadDir, filename)

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy file content
	if _, err = io.Copy(dst, src); err != nil {
		return "", err
	}

	// Return relative URL
	return fmt.Sprintf("/uploads/videos/%s", filename), nil
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
