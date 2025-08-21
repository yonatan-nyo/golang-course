package services

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
	"yonatan/labpro/models"

	"gorm.io/gorm"
)

type CourseService struct {
	db *gorm.DB
}

func NewCourseService(db *gorm.DB) *CourseService {
	return &CourseService{db: db}
}

func (cs *CourseService) CreateCourse(course *models.Course) (*models.Course, error) {
	if err := cs.db.Create(course).Error; err != nil {
		return nil, err
	}
	return course, nil
}

func (cs *CourseService) GetCourses(query string, page, limit int, userID interface{}) ([]map[string]interface{}, map[string]interface{}, error) {
	var courses []models.Course
	var total int64

	db := cs.db.Model(&models.Course{})

	// Apply search filter
	if query != "" {
		searchTerm := "%" + strings.ToLower(query) + "%"
		db = db.Where("LOWER(title) LIKE ? OR LOWER(instructor) LIKE ? OR EXISTS (SELECT 1 FROM unnest(topics) AS topic WHERE LOWER(topic) LIKE ?)",
			searchTerm, searchTerm, searchTerm)
	}

	// Count total
	db.Count(&total)

	// Apply pagination
	offset := (page - 1) * limit
	if err := db.Offset(offset).Limit(limit).Preload("Modules").Find(&courses).Error; err != nil {
		return nil, nil, err
	}

	// Convert to response format
	result := make([]map[string]interface{}, len(courses))
	for i, course := range courses {
		isPurchased := false
		if userID != nil {
			// Check if user purchased this course
			var userCourse models.UserCourse
			err := cs.db.Where("user_id = ? AND course_id = ?", userID, course.ID).First(&userCourse).Error
			isPurchased = (err == nil)
		}

		result[i] = map[string]interface{}{
			"id":              course.ID,
			"title":           course.Title,
			"instructor":      course.Instructor,
			"description":     course.Description,
			"topics":          course.Topics,
			"price":           course.Price,
			"thumbnail_image": course.ThumbnailImage,
			"total_modules":   len(course.Modules),
			"created_at":      course.CreatedAt,
			"updated_at":      course.UpdatedAt,
			"is_purchased":    isPurchased,
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

func (cs *CourseService) GetCourseByID(id string, userID interface{}) (map[string]interface{}, error) {
	var course models.Course
	if err := cs.db.Preload("Modules").First(&course, "id = ?", id).Error; err != nil {
		return nil, err
	}

	isPurchased := false
	if userID != nil {
		// Check if user purchased this course
		var userCourse models.UserCourse
		err := cs.db.Where("user_id = ? AND course_id = ?", userID, course.ID).First(&userCourse).Error
		isPurchased = (err == nil)
	}

	result := map[string]interface{}{
		"id":              course.ID,
		"title":           course.Title,
		"description":     course.Description,
		"instructor":      course.Instructor,
		"topics":          course.Topics,
		"price":           course.Price,
		"thumbnail_image": course.ThumbnailImage,
		"total_modules":   len(course.Modules),
		"created_at":      course.CreatedAt,
		"updated_at":      course.UpdatedAt,
		"is_purchased":    isPurchased,
	}

	return result, nil
}

func (cs *CourseService) UpdateCourse(course *models.Course) (*models.Course, error) {
	if err := cs.db.Save(course).Error; err != nil {
		return nil, err
	}
	return course, nil
}

func (cs *CourseService) DeleteCourse(id string) error {
	// First delete all modules associated with this course
	if err := cs.db.Where("course_id = ?", id).Delete(&models.Module{}).Error; err != nil {
		return err
	}

	// Delete user course relationships
	if err := cs.db.Where("course_id = ?", id).Delete(&models.UserCourse{}).Error; err != nil {
		return err
	}

	// Delete the course
	if err := cs.db.Delete(&models.Course{}, "id = ?", id).Error; err != nil {
		return err
	}

	return nil
}

func (cs *CourseService) BuyCourse(courseID, userID string) (map[string]interface{}, error) {
	// Check if course exists
	var course models.Course
	if err := cs.db.First(&course, "id = ?", courseID).Error; err != nil {
		return nil, errors.New("course not found")
	}

	// Check if user already purchased this course
	var existingUserCourse models.UserCourse
	if err := cs.db.Where("user_id = ? AND course_id = ?", userID, courseID).First(&existingUserCourse).Error; err == nil {
		return nil, errors.New("course already purchased")
	}

	// Get user and check balance
	var user models.User
	if err := cs.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, errors.New("user not found")
	}

	if user.Balance < course.Price {
		return nil, errors.New("insufficient balance")
	}

	// Start transaction
	tx := cs.db.Begin()

	// Deduct balance
	user.Balance -= course.Price
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create user course relationship
	userCourse := models.UserCourse{
		UserID:   userID,
		CourseID: courseID,
	}
	if err := tx.Create(&userCourse).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	result := map[string]interface{}{
		"course_id":      courseID,
		"user_balance":   user.Balance,
		"transaction_id": userCourse.ID,
	}

	return result, nil
}

func (cs *CourseService) GetMyCourses(userID, query string, page, limit int) ([]map[string]interface{}, map[string]interface{}, error) {
	var userCourses []models.UserCourse
	var total int64

	db := cs.db.Model(&models.UserCourse{}).Where("user_id = ?", userID)

	// Apply search filter
	if query != "" {
		searchTerm := "%" + strings.ToLower(query) + "%"
		db = db.Joins("JOIN courses ON user_courses.course_id = courses.id").
			Where("LOWER(courses.title) LIKE ? OR LOWER(courses.instructor) LIKE ? OR EXISTS (SELECT 1 FROM unnest(courses.topics) AS topic WHERE LOWER(topic) LIKE ?)",
				searchTerm, searchTerm, searchTerm)
	}

	// Count total
	db.Count(&total)

	// Apply pagination
	offset := (page - 1) * limit
	if err := db.Offset(offset).Limit(limit).Preload("Course").Find(&userCourses).Error; err != nil {
		return nil, nil, err
	}

	// Convert to response format
	result := make([]map[string]interface{}, len(userCourses))
	for i, userCourse := range userCourses {
		// Calculate progress percentage
		var totalModules int64
		var completedModules int64

		cs.db.Model(&models.Module{}).Where("course_id = ?", userCourse.CourseID).Count(&totalModules)
		cs.db.Model(&models.UserModuleProgress{}).
			Joins("JOIN modules ON user_module_progress.module_id = modules.id").
			Where("user_module_progress.user_id = ? AND modules.course_id = ? AND user_module_progress.is_completed = ?",
				userID, userCourse.CourseID, true).Count(&completedModules)

		progressPercentage := float64(0)
		if totalModules > 0 {
			progressPercentage = float64(completedModules) / float64(totalModules) * 100
		}

		result[i] = map[string]interface{}{
			"id":                  userCourse.Course.ID,
			"title":               userCourse.Course.Title,
			"instructor":          userCourse.Course.Instructor,
			"topics":              userCourse.Course.Topics,
			"thumbnail_image":     userCourse.Course.ThumbnailImage,
			"progress_percentage": progressPercentage,
			"purchased_at":        userCourse.PurchasedAt,
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

func (cs *CourseService) SaveThumbnail(file *multipart.FileHeader) (string, error) {
	// Create uploads directory if it doesn't exist
	uploadDir := "./uploads/thumbnails"
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
	return fmt.Sprintf("/uploads/thumbnails/%s", filename), nil
}
