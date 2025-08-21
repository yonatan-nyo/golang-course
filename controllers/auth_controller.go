package controllers

import (
	"net/http"
	"yonatan/labpro/models"
	"yonatan/labpro/services"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService *services.AuthService
}

func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

func (ac *AuthController) Register(c *gin.Context) {
	var req struct {
		FirstName       string `json:"first_name" binding:"required"`
		LastName        string `json:"last_name" binding:"required"`
		Username        string `json:"username" binding:"required"`
		Email           string `json:"email" binding:"required,email"`
		Password        string `json:"password" binding:"required"`
		ConfirmPassword string `json:"confirm_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid input: " + err.Error(),
			"data":    nil,
		})
		return
	}

	if req.Password != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Passwords do not match",
			"data":    nil,
		})
		return
	}

	user, err := ac.authService.Register(req.FirstName, req.LastName, req.Username, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "User registered successfully",
		"data": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
		},
	})
}

func (ac *AuthController) Login(c *gin.Context) {
	var req struct {
		Identifier string `json:"identifier" binding:"required"`
		Password   string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid input: " + err.Error(),
			"data":    nil,
		})
		return
	}

	token, user, err := ac.authService.Login(req.Identifier, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Login successful",
		"data": gin.H{
			"username": user.Username,
			"token":    token,
		},
	})
}

func (ac *AuthController) GetSelf(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "User not authenticated",
			"data":    nil,
		})
		return
	}

	userModel := user.(models.User)
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "User details retrieved successfully",
		"data": gin.H{
			"id":         userModel.ID,
			"username":   userModel.Username,
			"email":      userModel.Email,
			"first_name": userModel.FirstName,
			"last_name":  userModel.LastName,
			"balance":    userModel.Balance,
		},
	})
}

// Web Authentication Methods

func (ac *AuthController) ShowLoginPage(c *gin.Context) {
	// Check if user is already authenticated
	if ac.isUserAuthenticated(c) {
		ac.redirectAuthenticatedUser(c)
		return
	}

	c.HTML(http.StatusOK, "login.html", gin.H{
		"Title": "Login",
	})
}

func (ac *AuthController) ShowRegisterPage(c *gin.Context) {
	// Check if user is already authenticated
	if ac.isUserAuthenticated(c) {
		ac.redirectAuthenticatedUser(c)
		return
	}

	c.HTML(http.StatusOK, "register.html", gin.H{
		"Title": "Register",
	})
}

// Helper function to check if user is authenticated
func (ac *AuthController) isUserAuthenticated(c *gin.Context) bool {
	token, err := c.Cookie("token")
	return err == nil && token != ""
}

// Helper function to redirect authenticated users to appropriate dashboard
func (ac *AuthController) redirectAuthenticatedUser(c *gin.Context) {
	// We'll use a simple redirect to the root path which has the logic to redirect based on user role
	c.Redirect(http.StatusFound, "/")
}

func (ac *AuthController) HandleLogin(c *gin.Context) {
	identifier := c.PostForm("identifier")
	password := c.PostForm("password")

	if identifier == "" || password == "" {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"Title": "Login",
			"Error": "Please provide both identifier and password",
		})
		return
	}

	token, user, err := ac.authService.Login(identifier, password)
	if err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"Title": "Login",
			"Error": err.Error(),
		})
		return
	}

	// Set token as a cookie
	c.SetCookie("token", token, 3600*24*7, "/", "", false, true) // 7 days

	// Determine redirect URL based on user role
	redirectURL := "/dashboard"
	userRole := "user"
	if user.IsAdmin {
		redirectURL = "/admin/dashboard"
		userRole = "admin"
	}

	// Create a simple HTML page that sets localStorage and redirects
	redirectHTML := `<!DOCTYPE html>
<html>
<head>
    <title>Redirecting...</title>
</head>
<body>
    <div style="display: flex; justify-content: center; align-items: center; height: 100vh; font-family: Arial, sans-serif;">
        <div style="text-align: center;">
            <div style="display: inline-block; width: 40px; height: 40px; border: 4px solid #f3f3f3; border-top: 4px solid #3498db; border-radius: 50%; animation: spin 1s linear infinite;"></div>
            <p>Redirecting to dashboard...</p>
        </div>
    </div>
    <style>
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
    </style>
    <script>
        localStorage.setItem('isLoggedIn', 'true');
        localStorage.setItem('userRole', '` + userRole + `');
        setTimeout(function() {
            window.location.href = '` + redirectURL + `';
        }, 500);
    </script>
</body>
</html>`

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(redirectHTML))
}

func (ac *AuthController) HandleRegister(c *gin.Context) {
	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")

	// Store form data for re-display on error
	formData := gin.H{
		"FirstName": firstName,
		"LastName":  lastName,
		"Username":  username,
		"Email":     email,
	}

	if firstName == "" || lastName == "" || username == "" || email == "" || password == "" || confirmPassword == "" {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"Title":    "Register",
			"Error":    "All fields are required",
			"FormData": formData,
		})
		return
	}

	if password != confirmPassword {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"Title":    "Register",
			"Error":    "Passwords do not match",
			"FormData": formData,
		})
		return
	}

	_, err := ac.authService.Register(firstName, lastName, username, email, password)
	if err != nil {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"Title":    "Register",
			"Error":    err.Error(),
			"FormData": formData,
		})
		return
	}

	c.HTML(http.StatusOK, "login.html", gin.H{
		"Title":   "Login",
		"Success": "Account created successfully! Please login.",
	})
}

func (ac *AuthController) ShowDashboard(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	userModel := user.(models.User)

	// Redirect admins to admin dashboard
	if userModel.IsAdmin {
		c.Redirect(http.StatusFound, "/admin/dashboard")
		return
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"Title": "Dashboard",
		"User":  userModel,
	})
}

func (ac *AuthController) HandleLogout(c *gin.Context) {
	// Clear the token cookie
	c.SetCookie("token", "", -1, "/", "", false, true)

	// Create a simple HTML page that clears localStorage and redirects
	logoutHTML := `<!DOCTYPE html>
<html>
<head>
    <title>Logging out...</title>
</head>
<body>
    <div style="display: flex; justify-content: center; align-items: center; height: 100vh; font-family: Arial, sans-serif;">
        <div style="text-align: center;">
            <div style="display: inline-block; width: 40px; height: 40px; border: 4px solid #f3f3f3; border-top: 4px solid #3498db; border-radius: 50%; animation: spin 1s linear infinite;"></div>
            <p>Logging out...</p>
        </div>
    </div>
    <style>
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
    </style>
    <script>
        localStorage.removeItem('isLoggedIn');
        localStorage.removeItem('userRole');
        setTimeout(function() {
            window.location.href = '/auth/login';
        }, 500);
    </script>
</body>
</html>`

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(logoutHTML))
}

// Admin Dashboard Methods

func (ac *AuthController) ShowAdminDashboard(c *gin.Context) {
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

	// Get dashboard stats (you can implement these methods in services)
	stats := gin.H{
		"TotalUsers":     100,      // Replace with actual count from database
		"TotalCourses":   25,       // Replace with actual count from database
		"TotalRevenue":   15000.50, // Replace with actual calculation
		"ActiveStudents": 85,       // Replace with actual count from database
	}

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"Title": "Admin Dashboard",
		"User":  userModel,
		"Stats": stats,
	})
}
