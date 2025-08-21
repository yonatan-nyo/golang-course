package api

import (
	"net/http"
	"yonatan/labpro/models"
	"yonatan/labpro/services"

	"github.com/gin-gonic/gin"
)

type AuthAPIController struct {
	authService *services.AuthService
}

func NewAuthAPIController(authService *services.AuthService) *AuthAPIController {
	return &AuthAPIController{
		authService: authService,
	}
}

func (aac *AuthAPIController) Login(c *gin.Context) {
	var req struct {
		Identifier string `json:"identifier" binding:"required"`
		Password   string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	token, user, err := aac.authService.Login(req.Identifier, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	result := map[string]interface{}{
		"username": user.Username,
		"token":    token,
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Login successful",
		"data":    result,
	})
}

func (aac *AuthAPIController) Register(c *gin.Context) {
	var req struct {
		Username        string `json:"username" binding:"required"`
		Email           string `json:"email" binding:"required,email"`
		FirstName       string `json:"first_name" binding:"required"`
		LastName        string `json:"last_name" binding:"required"`
		Password        string `json:"password" binding:"required,min=8"`
		ConfirmPassword string `json:"confirm_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	if req.Password != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Password and confirm password do not match",
			"data":    nil,
		})
		return
	}

	user, err := aac.authService.Register(req.FirstName, req.LastName, req.Username, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	result := map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Registration successful",
		"data":    result,
	})
}

func (aac *AuthAPIController) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Logout successful",
		"data":    nil,
	})
}

func (aac *AuthAPIController) GetProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized",
			"data":    nil,
		})
		return
	}

	userModel := user.(models.User)
	result := map[string]interface{}{
		"id":         userModel.ID,
		"username":   userModel.Username,
		"email":      userModel.Email,
		"first_name": userModel.FirstName,
		"last_name":  userModel.LastName,
		"balance":    userModel.Balance,
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Profile retrieved successfully",
		"data":    result,
	})
}
