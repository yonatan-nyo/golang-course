package admin

import (
	"net/http"
	"yonatan/labpro/models"

	"github.com/gin-gonic/gin"
)

type DashboardController struct{}

func NewDashboardController() *DashboardController {
	return &DashboardController{}
}

func (dc *DashboardController) ShowAdminDashboard(c *gin.Context) {
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

	// Get dashboard stats
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
