package api

import (
	apiAdminCourse "yonatan/labpro/controllers/api/admin"
	apiUserCourse "yonatan/labpro/controllers/api/user"
	"yonatan/labpro/middleware"

	"github.com/gin-gonic/gin"
)

func SetupCourseRoutes(api *gin.RouterGroup,
	adminCourseController *apiAdminCourse.CourseAPIController,
	userCourseController *apiUserCourse.CourseAPIController) {

	// User course routes
	courses := api.Group("/courses")
	courses.Use(middleware.AuthMiddleware())
	{
		// GET /api/courses
		courses.GET("", userCourseController.GetCourses)
		// GET /api/courses/:courseId
		courses.GET("/:courseId", userCourseController.GetCourseByID)
		// POST /api/courses/:courseId/buy
		courses.POST("/:courseId/buy", userCourseController.PurchaseCourse)
		// GET /api/courses/my-courses
		courses.GET("/my-courses", userCourseController.GetMyCourses)
	}

	// Admin course routes
	adminCourses := api.Group("/courses")
	adminCourses.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		// POST /api/courses (admin only)
		adminCourses.POST("", adminCourseController.CreateCourse)
		// PUT /api/courses/:courseId (admin only)
		adminCourses.PUT("/:courseId", adminCourseController.UpdateCourse)
		// DELETE /api/courses/:courseId (admin only)
		adminCourses.DELETE("/:courseId", adminCourseController.DeleteCourse)
	}
}
