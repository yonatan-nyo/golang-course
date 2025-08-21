package services

import (
	"errors"
	"strings"
	"yonatan/labpro/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (us *UserService) GetUsers(query string, page, limit int) ([]map[string]interface{}, map[string]interface{}, error) {
	var users []models.User
	var total int64

	db := us.db.Model(&models.User{})

	// Apply search filter
	if query != "" {
		searchTerm := "%" + strings.ToLower(query) + "%"
		db = db.Where("LOWER(username) LIKE ? OR LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR LOWER(email) LIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm)
	}

	// Count total
	db.Count(&total)

	// Apply pagination
	offset := (page - 1) * limit
	if err := db.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, nil, err
	}

	// Convert to response format (exclude password)
	result := make([]map[string]interface{}, len(users))
	for i, user := range users {
		result[i] = map[string]interface{}{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"balance":    user.Balance,
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

func (us *UserService) GetUserByID(id string) (map[string]interface{}, error) {
	var user models.User
	if err := us.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}

	// Count courses purchased by this user
	var coursesPurchased int64
	us.db.Model(&models.UserCourse{}).Where("user_id = ?", id).Count(&coursesPurchased)

	result := map[string]interface{}{
		"id":                user.ID,
		"username":          user.Username,
		"email":             user.Email,
		"first_name":        user.FirstName,
		"last_name":         user.LastName,
		"balance":           user.Balance,
		"courses_purchased": coursesPurchased,
	}

	return result, nil
}

func (us *UserService) UpdateUserBalance(id string, increment float64) (*models.User, error) {
	var user models.User
	if err := us.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}

	user.Balance += increment
	if user.Balance < 0 {
		user.Balance = 0
	}

	if err := us.db.Save(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (us *UserService) UpdateUser(id, email, username, firstName, lastName, password string) (*models.User, error) {
	var user models.User
	if err := us.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}

	// Check if username or email already exists (excluding current user)
	var existingUser models.User
	if err := us.db.Where("(username = ? OR email = ?) AND id != ?", username, email, id).First(&existingUser).Error; err == nil {
		return nil, errors.New("username or email already exists")
	}

	user.Email = email
	user.Username = username
	user.FirstName = firstName
	user.LastName = lastName

	// Update password if provided
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.Password = string(hashedPassword)
	}

	if err := us.db.Save(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (us *UserService) DeleteUser(id string) error {
	// Check if user exists
	var user models.User
	if err := us.db.First(&user, "id = ?", id).Error; err != nil {
		return errors.New("user not found")
	}

	// Don't allow deleting admin user
	if user.Username == "admin" {
		return errors.New("cannot delete admin user")
	}

	// Start transaction
	tx := us.db.Begin()

	// Delete user's course purchases
	if err := tx.Where("user_id = ?", id).Delete(&models.UserCourse{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete user's module progress
	if err := tx.Where("user_id = ?", id).Delete(&models.UserModuleProgress{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete user
	if err := tx.Delete(&user).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
