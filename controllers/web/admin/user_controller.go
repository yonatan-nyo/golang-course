package admin

import (
	"net/http"
	"strconv"
	"yonatan/labpro/models"
	"yonatan/labpro/services"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService *services.UserService
}

func NewUserController(userService *services.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

func (uc *UserController) ShowUsersPage(c *gin.Context) {
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

	// Get query parameters for pagination and search
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	query := c.Query("q")

	// Get users from service
	users, pagination, err := uc.userService.GetUsers(query, page, limit)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "users.html", gin.H{
			"Title": "User Management",
			"User":  userModel,
			"Error": "Failed to fetch users",
		})
		return
	}

	c.HTML(http.StatusOK, "users.html", gin.H{
		"Title":      "User Management",
		"User":       userModel,
		"Users":      users,
		"Pagination": pagination,
		"Query":      query,
	})
}

func (uc *UserController) ShowEditUserPage(c *gin.Context) {
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

	userID := c.Param("id")
	targetUser, err := uc.userService.GetUserByID(userID)
	if err != nil {
		c.HTML(http.StatusNotFound, "user-edit.html", gin.H{
			"Title": "Edit User",
			"User":  userModel,
			"Error": "User not found",
		})
		return
	}

	c.HTML(http.StatusOK, "user-edit.html", gin.H{
		"Title":      "Edit User",
		"User":       userModel,
		"TargetUser": targetUser,
	})
}

func (uc *UserController) HandleUpdateUser(c *gin.Context) {
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

	userID := c.Param("id")

	// Get existing user
	targetUser, err := uc.userService.GetUserByID(userID)
	if err != nil {
		c.HTML(http.StatusNotFound, "user-edit.html", gin.H{
			"Title": "Edit User",
			"User":  userModel,
			"Error": "User not found",
		})
		return
	}

	// Handle form submission
	username := c.PostForm("username")
	email := c.PostForm("email")
	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")
	password := c.PostForm("password")
	isAdminStr := c.PostForm("is_admin")

	// Validate required fields
	if username == "" || email == "" || firstName == "" || lastName == "" {
		c.HTML(http.StatusBadRequest, "user-edit.html", gin.H{
			"Title":      "Edit User",
			"User":       userModel,
			"TargetUser": targetUser,
			"Error":      "Username, email, first name, and last name are required",
		})
		return
	}

	isAdmin := isAdminStr == "on" || isAdminStr == "true"

	// Update user (note: password is optional, empty string means no change)
	_, err = uc.userService.UpdateUser(userID, email, username, firstName, lastName, password)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "user-edit.html", gin.H{
			"Title":      "Edit User",
			"User":       userModel,
			"TargetUser": targetUser,
			"Error":      "Failed to update user: " + err.Error(),
		})
		return
	}

	// Handle admin status change separately if needed (you might need to add this to the service)
	_ = isAdmin // For now, we're not updating admin status as it's not in the UpdateUser method

	c.Redirect(http.StatusFound, "/admin/users?success=User updated successfully&id="+userID)
}

func (uc *UserController) HandleDeleteUser(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete your own account"})
		return
	}

	err := uc.userService.DeleteUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (uc *UserController) ShowUserDetails(c *gin.Context) {
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

	userID := c.Param("id")
	targetUser, err := uc.userService.GetUserByID(userID)
	if err != nil {
		c.HTML(http.StatusNotFound, "user-details.html", gin.H{
			"Title": "User Details",
			"User":  userModel,
			"Error": "User not found",
		})
		return
	}

	// Get user's enrolled courses (for now, just return empty slice)
	// TODO: Add GetUserEnrolledCourses method to UserService if needed
	enrolledCourses := []models.Course{}

	c.HTML(http.StatusOK, "user-details.html", gin.H{
		"Title":           "User Details",
		"User":            userModel,
		"TargetUser":      targetUser,
		"EnrolledCourses": enrolledCourses,
	})
}
