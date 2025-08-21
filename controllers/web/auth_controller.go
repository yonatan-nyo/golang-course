package web

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
		"FirstName":       firstName,
		"LastName":        lastName,
		"Username":        username,
		"Email":           email,
		"Password":        password,
		"ConfirmPassword": confirmPassword,
	}

	// Validate required fields
	if firstName == "" || lastName == "" || username == "" || email == "" || password == "" {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"Title": "Register",
			"Error": "All fields are required",
			"Form":  formData,
		})
		return
	}

	// Validate password confirmation
	if password != confirmPassword {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"Title": "Register",
			"Error": "Passwords do not match",
			"Form":  formData,
		})
		return
	}

	// Register user
	user, err := ac.authService.Register(firstName, lastName, username, email, password)
	if err != nil {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"Title": "Register",
			"Error": err.Error(),
			"Form":  formData,
		})
		return
	}

	// Auto login after registration
	token, _, err := ac.authService.Login(username, password)
	if err != nil {
		// Registration succeeded but login failed, redirect to login page
		c.HTML(http.StatusOK, "login.html", gin.H{
			"Title":   "Login",
			"Success": "Registration successful! Please login with your credentials.",
		})
		return
	}

	// Set token as a cookie
	c.SetCookie("token", token, 3600*24*7, "/", "", false, true)

	// Determine redirect URL based on user role
	redirectURL := "/dashboard"
	userRole := "user"
	if user.IsAdmin {
		redirectURL = "/admin/dashboard"
		userRole = "admin"
	}

	// Create redirect HTML
	redirectHTML := `<!DOCTYPE html>
<html>
<head>
    <title>Registration Successful</title>
</head>
<body>
    <div style="display: flex; justify-content: center; align-items: center; height: 100vh; font-family: Arial, sans-serif;">
        <div style="text-align: center;">
            <div style="display: inline-block; width: 40px; height: 40px; border: 4px solid #f3f3f3; border-top: 4px solid #3498db; border-radius: 50%; animation: spin 1s linear infinite;"></div>
            <p>Registration successful! Redirecting to dashboard...</p>
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
        }, 1000);
    </script>
</body>
</html>`

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(redirectHTML))
}

func (ac *AuthController) HandleLogout(c *gin.Context) {
	// Clear the token cookie
	c.SetCookie("token", "", -1, "/", "", false, true)

	// Create logout HTML that clears localStorage and redirects
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

// Helper function to check if user is authenticated
func (ac *AuthController) isUserAuthenticated(c *gin.Context) bool {
	token, err := c.Cookie("token")
	hasToken := err == nil && token != ""
	return hasToken
}

// Helper function to redirect authenticated users to appropriate dashboard
func (ac *AuthController) redirectAuthenticatedUser(c *gin.Context) {
	// We'll use a simple redirect to the root path which has the logic to redirect based on user role
	c.Redirect(http.StatusFound, "/")
}
