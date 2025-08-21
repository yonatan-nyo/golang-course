package services

import (
	"errors"
	"time"
	"yonatan/labpro/config"
	"yonatan/labpro/database"
	"yonatan/labpro/models"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (as *AuthService) Register(firstName, lastName, username, email, password string) (*models.User, error) {
	// Check if username or email already exists
	var existingUser models.User
	if err := database.DB.Where("username = ? OR email = ?", username, email).First(&existingUser).Error; err == nil {
		if existingUser.Username == username {
			return nil, errors.New("username already exists")
		}
		return nil, errors.New("email already exists")
	}

	// Create new user
	user := models.User{
		FirstName: firstName,
		LastName:  lastName,
		Username:  username,
		Email:     email,
		Balance:   0,
		IsAdmin:   false,
	}

	if err := user.SetPassword(password); err != nil {
		return nil, errors.New("failed to hash password")
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return nil, errors.New("failed to create user")
	}

	return &user, nil
}

func (as *AuthService) Login(identifier, password string) (string, *models.User, error) {
	var user models.User

	// Find user by username or email
	if err := database.DB.Where("username = ? OR email = ?", identifier, identifier).First(&user).Error; err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	// Check password
	if !user.CheckPassword(password) {
		return "", nil, errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := as.generateToken(user.ID)
	if err != nil {
		return "", nil, errors.New("failed to generate token")
	}

	return token, &user, nil
}

func (as *AuthService) generateToken(userID string) (string, error) {
	cfg := config.Load()

	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 1).Unix(), // 1 hour expiration
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}
