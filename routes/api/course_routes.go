package api

import (
	"yonatan/labpro/controllers"
	"yonatan/labpro/middleware"

	"github.com/gin-gonic/gin"
)

func SetupCourseRoutes(api *gin.RouterGroup, courseController *controllers.CourseController) {
	courses := api.Group("/courses")
	courses.Use(middleware.AuthMiddleware())
	{
		// List and create courses
		courses.GET("", courseController.GetCourses)
		courses.POST("", courseController.CreateCourse)
		courses.GET("/my-courses", courseController.GetMyCourses)

		// Course-specific operations with ID parameter
		courses.GET("/:id", courseController.GetCourse)
		courses.POST("/:id/buy", courseController.BuyCourse)
		courses.PUT("/:id", courseController.UpdateCourse)
		courses.DELETE("/:id", courseController.DeleteCourse)
	}
}
