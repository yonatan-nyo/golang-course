package admin

import (
	"net/http"
	"strconv"
	"yonatan/labpro/models"
	"yonatan/labpro/services"

	"github.com/gin-gonic/gin"
)

type UserAPIController struct {
	userService *services.UserService
}

func NewUserAPIController(userService *services.UserService) *UserAPIController {
	return &UserAPIController{
		userService: userService,
	}
}

// GetUsers godoc
// @Summary      Get all users with pagination (Admin only)
// @Description  Retrieve a paginated list of all users with optional search functionality
// @Tags         admin-users
// @Produce      json
// @Security     BearerAuth
// @Param        q     query    string  false  "Search query for username"
// @Param        page  query    int     false  "Page number (default: 1)"
// @Param        limit query    int     false  "Number of items per page (default: 15, max: 50)"
// @Success      200   {object} object{status=string,message=string,data=[]object,pagination=object}
// @Failure      401   {object} object{error=string}
// @Failure      403   {object} object{error=string}
// @Failure      500   {object} object{status=string,message=string,data=object}
// @Router       /users [get]
func (uac *UserAPIController) GetUsers(c *gin.Context) {
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

	// Get query parameters
	query := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "15"))
	if limit > 50 {
		limit = 50
	}

	// Get users from service
	users, pagination, err := uac.userService.GetUsers(query, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch users",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"message":    "Users retrieved successfully",
		"data":       users,
		"pagination": pagination,
	})
}

// GetUserByID godoc
// @Summary      Get user details by ID (Admin only)
// @Description  Retrieve detailed information for a specific user by their ID
// @Tags         admin-users
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      string  true  "User ID"
// @Success      200 {object}  object{status=string,message=string,data=object}
// @Failure      401 {object}  object{error=string}
// @Failure      403 {object}  object{error=string}
// @Failure      404 {object}  object{status=string,message=string,data=object}
// @Router       /users/{id} [get]
func (uac *UserAPIController) GetUserByID(c *gin.Context) {
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

	userID := c.Param("id")
	targetUser, err := uac.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "User not found",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "User retrieved successfully",
		"data":    targetUser,
	})
}

// UpdateUserBalance godoc
// @Summary      Update user balance (Admin only)
// @Description  Increment or decrement a user's balance by a specific amount
// @Tags         admin-users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string  true   "User ID"
// @Param        request  body      object  true   "Balance increment data" example({"increment":100.50})
// @Success      200      {object}  object{status=string,message=string,data=object{id=string,username=string,balance=number}}
// @Failure      400      {object}  object{status=string,message=string,data=object}
// @Failure      401      {object}  object{error=string}
// @Failure      403      {object}  object{error=string}
// @Failure      500      {object}  object{status=string,message=string,data=object}
// @Router       /users/{id}/balance [post]
func (uac *UserAPIController) UpdateUserBalance(c *gin.Context) {
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

	userID := c.Param("id")

	var req struct {
		Increment float64 `json:"increment" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// Update user balance
	updatedUser, err := uac.userService.UpdateUserBalance(userID, req.Increment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to update user balance",
			"data":    nil,
		})
		return
	}

	result := map[string]interface{}{
		"id":       updatedUser.ID,
		"username": updatedUser.Username,
		"balance":  updatedUser.Balance,
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "User balance updated successfully",
		"data":    result,
	})
}

// UpdateUser godoc
// @Summary      Update user information (Admin only)
// @Description  Update a user's profile information including email, username, names, and optionally password
// @Tags         admin-users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string  true   "User ID"
// @Param        request  body      object  true   "User update data" example({"email":"user@example.com","username":"newusername","first_name":"John","last_name":"Doe","password":"newpassword123"})
// @Success      200      {object}  object{status=string,message=string,data=object{id=string,username=string,first_name=string,last_name=string,balance=number}}
// @Failure      400      {object}  object{status=string,message=string,data=object}
// @Failure      401      {object}  object{error=string}
// @Failure      403      {object}  object{error=string}
// @Failure      500      {object}  object{status=string,message=string,data=object}
// @Router       /users/{id} [put]
func (uac *UserAPIController) UpdateUser(c *gin.Context) {
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

	userID := c.Param("id")

	var req struct {
		Email     string `json:"email" binding:"required,email"`
		Username  string `json:"username" binding:"required"`
		FirstName string `json:"first_name" binding:"required"`
		LastName  string `json:"last_name" binding:"required"`
		Password  string `json:"password,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// Update user
	updatedUser, err := uac.userService.UpdateUser(userID, req.Email, req.Username, req.FirstName, req.LastName, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to update user: " + err.Error(),
			"data":    nil,
		})
		return
	}

	result := map[string]interface{}{
		"id":         updatedUser.ID,
		"username":   updatedUser.Username,
		"first_name": updatedUser.FirstName,
		"last_name":  updatedUser.LastName,
		"balance":    updatedUser.Balance,
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "User updated successfully",
		"data":    result,
	})
}

// DeleteUser godoc
// @Summary      Delete a user (Admin only)
// @Description  Delete a user account permanently. Admins cannot delete their own account.
// @Tags         admin-users
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      string  true  "User ID"
// @Success      204 "User deleted successfully"
// @Failure      400 {object}  object{status=string,message=string,data=object}
// @Failure      401 {object}  object{error=string}
// @Failure      403 {object}  object{error=string}
// @Failure      500 {object}  object{status=string,message=string,data=object}
// @Router       /users/{id} [delete]
func (uac *UserAPIController) DeleteUser(c *gin.Context) {
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

	userID := c.Param("id")

	// Prevent admin from deleting themselves
	if userID == userModel.ID {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Cannot delete your own account",
			"data":    nil,
		})
		return
	}

	err := uac.userService.DeleteUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to delete user",
			"data":    nil,
		})
		return
	}

	c.Status(http.StatusNoContent)
}
