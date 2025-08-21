package controllers

import (
	"net/http"
	"strconv"
	"yonatan/labpro/services"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService *services.UserService
}

func NewUserController(userService *services.UserService) *UserController {
	return &UserController{userService: userService}
}

func (uc *UserController) GetUsers(c *gin.Context) {
	// Check if user is admin
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "Access denied. Admin only.",
			"data":    nil,
		})
		return
	}

	// Get query parameters
	q := c.Query("q")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "15")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 15
	}

	users, pagination, err := uc.userService.GetUsers(q, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get users: " + err.Error(),
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

func (uc *UserController) GetUser(c *gin.Context) {
	// Check if user is admin
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "Access denied. Admin only.",
			"data":    nil,
		})
		return
	}

	id := c.Param("id")

	user, err := uc.userService.GetUserByID(id)
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
		"data":    user,
	})
}

func (uc *UserController) UpdateUserBalance(c *gin.Context) {
	// Check if user is admin
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "Access denied. Admin only.",
			"data":    nil,
		})
		return
	}

	id := c.Param("id")

	var req struct {
		Increment float64 `json:"increment" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid input: " + err.Error(),
			"data":    nil,
		})
		return
	}

	user, err := uc.userService.UpdateUserBalance(id, req.Increment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to update user balance: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "User balance updated successfully",
		"data": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"balance":  user.Balance,
		},
	})
}

func (uc *UserController) UpdateUser(c *gin.Context) {
	// Check if user is admin
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "Access denied. Admin only.",
			"data":    nil,
		})
		return
	}

	id := c.Param("id")

	// Check if trying to update admin user
	if id == "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "Cannot update admin user",
			"data":    nil,
		})
		return
	}

	var req struct {
		Email     string `json:"email" binding:"required,email"`
		Username  string `json:"username" binding:"required"`
		FirstName string `json:"first_name" binding:"required"`
		LastName  string `json:"last_name" binding:"required"`
		Password  string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid input: " + err.Error(),
			"data":    nil,
		})
		return
	}

	user, err := uc.userService.UpdateUser(id, req.Email, req.Username, req.FirstName, req.LastName, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to update user: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "User updated successfully",
		"data": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"balance":    user.Balance,
		},
	})
}

func (uc *UserController) DeleteUser(c *gin.Context) {
	// Check if user is admin
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "Access denied. Admin only.",
			"data":    nil,
		})
		return
	}

	id := c.Param("id")
	currentUserID, _ := c.Get("user_id")

	// Check if trying to delete admin user or self
	if id == "admin" || id == currentUserID {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "Cannot delete admin user or yourself",
			"data":    nil,
		})
		return
	}

	err := uc.userService.DeleteUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to delete user: " + err.Error(),
			"data":    nil,
		})
		return
	}

	c.Status(http.StatusNoContent)
}
