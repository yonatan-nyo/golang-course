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

// Login godoc
// @Summary      User login
// @Description  Authenticate user with identifier (username/email) and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        login  body      object{identifier=string,password=string}  true  "Login credentials"
// @Success      200    {object}  object{status=string,message=string,data=object{username=string,token=string}}
// @Failure      400    {object}  object{status=string,message=string,data=object}
// @Failure      401    {object}  object{status=string,message=string,data=object}
// @Router       /auth/login [post]
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

// Register godoc
// @Summary      User registration
// @Description  Register a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        register  body      object{username=string,email=string,first_name=string,last_name=string,password=string,confirm_password=string}  true  "Registration data"
// @Success      201       {object}  object{status=string,message=string,data=object{username=string,token=string}}
// @Failure      400       {object}  object{status=string,message=string,data=object}
// @Failure      409       {object}  object{status=string,message=string,data=object}
// @Router       /auth/register [post]
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

// Logout godoc
// @Summary      User logout
// @Description  Logout current user
// @Tags         auth
// @Produce      json
// @Success      200  {object}  object{status=string,message=string,data=object}
// @Router       /auth/logout [post]
func (aac *AuthAPIController) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Logout successful",
		"data":    nil,
	})
}

// GetProfile godoc
// @Summary      Get current user profile
// @Description  Get the profile of the currently authenticated user
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  object{status=string,message=string,data=object{id=string,username=string,email=string,first_name=string,last_name=string,role=string}}
// @Failure      401  {object}  object{status=string,message=string,data=object}
// @Router       /auth/self [get]
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
