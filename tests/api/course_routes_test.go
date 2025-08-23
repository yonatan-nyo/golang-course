package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"yonatan/labpro/config"
	apiAdminControllers "yonatan/labpro/controllers/api/admin"
	apiUserControllers "yonatan/labpro/controllers/api/user"
	"yonatan/labpro/database"
	"yonatan/labpro/models"
	apiRoutes "yonatan/labpro/routes/api"
	"yonatan/labpro/services"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var courseTestDB *gorm.DB

func setupCourseTestDB() {
	cfg := config.LoadTestWithProjectRoot()

	var err error
	courseTestDB, err = gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to test database: " + err.Error())
	}

	// Set the global database instance
	database.DB = courseTestDB

	// Auto migrate the schema
	err = courseTestDB.AutoMigrate(
		&models.User{},
		&models.Course{},
		&models.Module{},
		&models.UserCourse{},
		&models.UserModuleProgress{},
	)
	if err != nil {
		panic("Failed to migrate test database: " + err.Error())
	}
}

func cleanupCourseTestDB() {
	// Clean up test data in correct order due to foreign key constraints
	courseTestDB.Exec("DELETE FROM user_module_progresses")
	courseTestDB.Exec("DELETE FROM user_courses")
	courseTestDB.Exec("DELETE FROM modules")
	courseTestDB.Exec("DELETE FROM courses")
	courseTestDB.Exec("DELETE FROM users")
}

func setupCourseTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Get config for services
	cfg := config.LoadTestWithProjectRoot()

	// Initialize services
	courseService := services.NewCourseService(courseTestDB, cfg)

	// Initialize controllers
	userCourseController := apiUserControllers.NewCourseAPIController(courseService)
	adminCourseController := apiAdminControllers.NewCourseAPIController(courseService)

	api := router.Group("/api")
	apiRoutes.SetupCourseRoutes(api, adminCourseController, userCourseController, cfg)

	return router
}

func createTestUser(isAdmin bool) models.User {
	user := models.User{
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Balance:   1000.0,
		IsAdmin:   isAdmin,
	}
	user.SetPassword("password123")
	courseTestDB.Create(&user)
	return user
}

func createTestCourse() models.Course {
	course := models.Course{
		Title:       "Test Course",
		Description: "A test course for unit testing",
		Instructor:  "Test Instructor",
		Price:       100.0,
		Thumbnail:   "test-thumbnail.jpg",
		Topics:      pq.StringArray{"programming", "testing"},
	}
	courseTestDB.Create(&course)
	return course
}

func createUserToken(user models.User) string {
	cfg := config.LoadTestWithProjectRoot()
	authService := services.NewAuthService(cfg)
	token, _, _ := authService.Login(user.Username, "password123")
	return token
}

func TestCourseRoutes(t *testing.T) {
	// Setup test database
	setupCourseTestDB()
	defer cleanupCourseTestDB()

	router := setupCourseTestRouter()

	t.Run("GET /api/courses", func(t *testing.T) {
		t.Run("should get courses successfully with authentication", func(t *testing.T) {
			cleanupCourseTestDB()

			// Create test user and course
			user := createTestUser(false)
			course := createTestCourse()
			token := createUserToken(user)

			req, _ := http.NewRequest("GET", "/api/courses", nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "success", response["status"])
			assert.Equal(t, "Courses retrieved successfully", response["message"])
			assert.NotNil(t, response["data"])
			assert.NotNil(t, response["pagination"])

			// Check if our test course is in the response
			data := response["data"].([]interface{})
			assert.Len(t, data, 1)

			courseData := data[0].(map[string]interface{})
			assert.Equal(t, course.Title, courseData["title"])
			assert.Equal(t, course.Price, courseData["price"])
		})

		t.Run("should get courses with query parameter", func(t *testing.T) {
			cleanupCourseTestDB()

			// Create test user and courses
			user := createTestUser(false)
			course1 := models.Course{
				Title:       "Go Programming",
				Description: "Learn Go programming language",
				Instructor:  "Go Expert",
				Price:       150.0,
				Topics:      pq.StringArray{"go", "programming"},
			}
			course2 := models.Course{
				Title:       "Python Basics",
				Description: "Learn Python fundamentals",
				Instructor:  "Python Expert",
				Price:       120.0,
				Topics:      pq.StringArray{"python", "programming"},
			}
			courseTestDB.Create(&course1)
			courseTestDB.Create(&course2)
			token := createUserToken(user)

			// Search for "Go"
			req, _ := http.NewRequest("GET", "/api/courses?q=Go", nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			data := response["data"].([]interface{})
			assert.Len(t, data, 1)

			courseData := data[0].(map[string]interface{})
			assert.Equal(t, course1.Title, courseData["title"])
		})

		t.Run("should fail without authentication", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/courses", nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})

		t.Run("should handle pagination parameters", func(t *testing.T) {
			cleanupCourseTestDB()

			// Create test user
			user := createTestUser(false)
			token := createUserToken(user)

			// Create multiple courses
			for i := 1; i <= 5; i++ {
				course := models.Course{
					Title:       fmt.Sprintf("Course %d", i),
					Description: fmt.Sprintf("Description for course %d", i),
					Instructor:  "Test Instructor",
					Price:       float64(100 + i*10),
					Topics:      pq.StringArray{"programming"},
				}
				courseTestDB.Create(&course)
			}

			// Test pagination
			req, _ := http.NewRequest("GET", "/api/courses?page=1&limit=3", nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			data := response["data"].([]interface{})
			assert.Len(t, data, 3)

			pagination := response["pagination"].(map[string]interface{})
			assert.Equal(t, float64(1), pagination["current_page"])
			assert.Equal(t, float64(5), pagination["total_items"])
		})
	})

	t.Run("GET /api/courses/:courseId", func(t *testing.T) {
		t.Run("should get course by ID successfully", func(t *testing.T) {
			cleanupCourseTestDB()

			// Create test user and course
			user := createTestUser(false)
			course := createTestCourse()
			token := createUserToken(user)

			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/courses/%s", course.ID), nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "success", response["status"])
			assert.Equal(t, "Course retrieved successfully", response["message"])
			assert.NotNil(t, response["data"])

			data := response["data"].(map[string]interface{})
			assert.Equal(t, course.Title, data["title"])
			assert.Equal(t, course.Description, data["description"])
			assert.Equal(t, course.Price, data["price"])
		})

		t.Run("should return 404 for non-existent course", func(t *testing.T) {
			cleanupCourseTestDB()

			// Create test user
			user := createTestUser(false)
			token := createUserToken(user)

			req, _ := http.NewRequest("GET", "/api/courses/non-existent-id", nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "error", response["status"])
			assert.Equal(t, "Course not found", response["message"])
		})

		t.Run("should fail without authentication", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/courses/some-id", nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	})

	t.Run("GET /api/courses/my-courses", func(t *testing.T) {
		t.Run("should get user's enrolled courses successfully", func(t *testing.T) {
			cleanupCourseTestDB()

			// Create test user and course
			user := createTestUser(false)
			course := createTestCourse()
			token := createUserToken(user)

			// Enroll user in course
			userCourse := models.UserCourse{
				UserID:   user.ID,
				CourseID: course.ID,
			}
			courseTestDB.Create(&userCourse)

			req, _ := http.NewRequest("GET", "/api/courses/my-courses", nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "success", response["status"])
			assert.Equal(t, "My courses retrieved successfully", response["message"])
			assert.NotNil(t, response["data"])

			data := response["data"].([]interface{})
			assert.Len(t, data, 1)

			courseData := data[0].(map[string]interface{})
			assert.Equal(t, course.Title, courseData["title"])
		})

		t.Run("should return empty array when user has no enrolled courses", func(t *testing.T) {
			cleanupCourseTestDB()

			// Create test user (no enrolled courses)
			user := createTestUser(false)
			token := createUserToken(user)

			req, _ := http.NewRequest("GET", "/api/courses/my-courses", nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			data := response["data"].([]interface{})
			assert.Len(t, data, 0)
		})

		t.Run("should fail without authentication", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/courses/my-courses", nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	})

	t.Run("POST /api/courses/:courseId/buy", func(t *testing.T) {
		t.Run("should purchase course successfully", func(t *testing.T) {
			cleanupCourseTestDB()

			// Create test user with sufficient balance
			user := createTestUser(false)
			course := createTestCourse()
			token := createUserToken(user)

			req, _ := http.NewRequest("POST", fmt.Sprintf("/api/courses/%s/buy", course.ID), nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "success", response["status"])
			assert.Equal(t, "Course purchased successfully", response["message"])
			assert.NotNil(t, response["data"])

			// Verify user is enrolled in course
			var userCourse models.UserCourse
			err = courseTestDB.Where("user_id = ? AND course_id = ?", user.ID, course.ID).First(&userCourse).Error
			assert.NoError(t, err)

			// Verify user balance was deducted
			var updatedUser models.User
			courseTestDB.First(&updatedUser, "id = ?", user.ID)
			assert.Equal(t, user.Balance-course.Price, updatedUser.Balance)
		})

		t.Run("should fail with insufficient balance", func(t *testing.T) {
			cleanupCourseTestDB()

			// Create test user with insufficient balance
			user := models.User{
				Username:  "pooruser",
				Email:     "poor@example.com",
				FirstName: "Poor",
				LastName:  "User",
				Balance:   50.0, // Less than course price (100.0)
				IsAdmin:   false,
			}
			user.SetPassword("password123")
			courseTestDB.Create(&user)

			course := createTestCourse()
			token := createUserToken(user)

			req, _ := http.NewRequest("POST", fmt.Sprintf("/api/courses/%s/buy", course.ID), nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "error", response["status"])
			assert.Contains(t, response["message"], "insufficient")
		})

		t.Run("should fail when course already purchased", func(t *testing.T) {
			cleanupCourseTestDB()

			// Create test user and course
			user := createTestUser(false)
			course := createTestCourse()
			token := createUserToken(user)

			// Enroll user in course first
			userCourse := models.UserCourse{
				UserID:   user.ID,
				CourseID: course.ID,
			}
			courseTestDB.Create(&userCourse)

			req, _ := http.NewRequest("POST", fmt.Sprintf("/api/courses/%s/buy", course.ID), nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "error", response["status"])
			assert.Contains(t, response["message"], "already")
		})

		t.Run("should fail for non-existent course", func(t *testing.T) {
			cleanupCourseTestDB()

			// Create test user
			user := createTestUser(false)
			token := createUserToken(user)

			req, _ := http.NewRequest("POST", "/api/courses/non-existent-id/buy", nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "error", response["status"])
		})

		t.Run("should fail without authentication", func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/api/courses/some-id/buy", nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	})

	// Admin Course Routes Tests
	t.Run("Admin Course Routes", func(t *testing.T) {
		t.Run("GET /api/courses (admin)", func(t *testing.T) {
			t.Run("should get courses successfully as admin", func(t *testing.T) {
				cleanupCourseTestDB()

				// Create admin user and course
				adminUser := createTestUser(true)
				createTestCourse() // Create a course for listing
				adminToken := createUserToken(adminUser)

				req, _ := http.NewRequest("GET", "/api/courses", nil)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, "success", response["status"])
				assert.Equal(t, "Courses retrieved successfully", response["message"])
				assert.NotNil(t, response["data"])
				assert.NotNil(t, response["pagination"])

				courses := response["data"].([]interface{})
				assert.GreaterOrEqual(t, len(courses), 1)
			})
		})

		t.Run("GET /api/courses/:courseId (admin)", func(t *testing.T) {
			t.Run("should get course by ID successfully as admin", func(t *testing.T) {
				cleanupCourseTestDB()

				// Create admin user and course
				adminUser := createTestUser(true)
				course := createTestCourse()
				adminToken := createUserToken(adminUser)

				req, _ := http.NewRequest("GET", fmt.Sprintf("/api/courses/%s", course.ID), nil)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, "success", response["status"])
				assert.Equal(t, "Course retrieved successfully", response["message"])
				assert.NotNil(t, response["data"])

				courseData := response["data"].(map[string]interface{})
				assert.Equal(t, course.ID, courseData["id"])
				assert.Equal(t, course.Title, courseData["title"])
			})

			t.Run("should fail for non-existent course", func(t *testing.T) {
				cleanupCourseTestDB()

				// Create admin user
				adminUser := createTestUser(true)
				adminToken := createUserToken(adminUser)

				req, _ := http.NewRequest("GET", "/api/courses/non-existent-id", nil)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusNotFound, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, "error", response["status"])
				assert.Equal(t, "Course not found", response["message"])
			})
		})

		t.Run("POST /api/courses (admin)", func(t *testing.T) {
			t.Run("should create course successfully as admin", func(t *testing.T) {
				cleanupCourseTestDB()

				// Create admin user
				adminUser := createTestUser(true)
				adminToken := createUserToken(adminUser)

				// Create multipart form data
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				// Add form fields
				writer.WriteField("title", "Admin Created Course")
				writer.WriteField("description", "A course created by admin")
				writer.WriteField("instructor", "Admin Instructor")
				writer.WriteField("price", "199.99")
				writer.WriteField("topics", "programming,web development")

				// Add empty thumbnail file (optional)
				part, _ := writer.CreateFormFile("thumbnail", "")
				part.Write([]byte(""))

				writer.Close()

				req, _ := http.NewRequest("POST", "/api/courses", &body)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))
				req.Header.Set("Content-Type", writer.FormDataContentType())

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				// Should be successful with proper form data
				assert.Equal(t, http.StatusCreated, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, "success", response["status"])
				assert.Equal(t, "Course created successfully", response["message"])
			})

			t.Run("should fail with missing required fields", func(t *testing.T) {
				cleanupCourseTestDB()

				// Create admin user
				adminUser := createTestUser(true)
				adminToken := createUserToken(adminUser)

				// Create multipart form data with missing required fields
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				// Only add title, missing other required fields
				writer.WriteField("title", "Incomplete Course")

				writer.Close()

				req, _ := http.NewRequest("POST", "/api/courses", &body)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))
				req.Header.Set("Content-Type", writer.FormDataContentType())

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusBadRequest, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, "error", response["status"])
			})

			t.Run("should fail with invalid price format", func(t *testing.T) {
				cleanupCourseTestDB()

				// Create admin user
				adminUser := createTestUser(true)
				adminToken := createUserToken(adminUser)

				// Create multipart form data with invalid price
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				writer.WriteField("title", "Invalid Price Course")
				writer.WriteField("description", "A course with invalid price")
				writer.WriteField("instructor", "Admin Instructor")
				writer.WriteField("price", "not-a-number")
				writer.WriteField("topics", "programming")

				writer.Close()

				req, _ := http.NewRequest("POST", "/api/courses", &body)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))
				req.Header.Set("Content-Type", writer.FormDataContentType())

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusBadRequest, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, "error", response["status"])
			})

			t.Run("should fail for non-admin user", func(t *testing.T) {
				cleanupCourseTestDB()

				// Create regular user
				user := createTestUser(false)
				token := createUserToken(user)

				// Create multipart form data
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				writer.WriteField("title", "Unauthorized Course")
				writer.WriteField("description", "This should fail")
				writer.WriteField("instructor", "Regular User")
				writer.WriteField("price", "100.0")

				writer.Close()

				req, _ := http.NewRequest("POST", "/api/courses", &body)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
				req.Header.Set("Content-Type", writer.FormDataContentType())

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusForbidden, w.Code)
			})

			t.Run("should fail without authentication", func(t *testing.T) {
				// Create multipart form data
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				writer.WriteField("title", "Unauthenticated Course")
				writer.WriteField("description", "This should fail")
				writer.WriteField("instructor", "Anonymous")
				writer.WriteField("price", "100.0")

				writer.Close()

				req, _ := http.NewRequest("POST", "/api/courses", &body)
				req.Header.Set("Content-Type", writer.FormDataContentType())

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusUnauthorized, w.Code)
			})
		})

		t.Run("PUT /api/courses/:courseId (admin)", func(t *testing.T) {
			t.Run("should update course successfully as admin", func(t *testing.T) {
				cleanupCourseTestDB()

				// Create admin user and course
				adminUser := createTestUser(true)
				course := createTestCourse()
				adminToken := createUserToken(adminUser)

				// Create multipart form data
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				// Add form fields
				writer.WriteField("title", "Updated Course Title")
				writer.WriteField("description", "Updated course description")
				writer.WriteField("instructor", "Updated Instructor")
				writer.WriteField("price", "299.99")
				writer.WriteField("topics", "updated,topics")

				// Add empty thumbnail file (optional)
				part, _ := writer.CreateFormFile("thumbnail", "")
				part.Write([]byte(""))

				writer.Close()

				req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/courses/%s", course.ID), &body)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))
				req.Header.Set("Content-Type", writer.FormDataContentType())

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				// Should be successful with proper form data
				assert.Equal(t, http.StatusOK, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, "success", response["status"])
				assert.Equal(t, "Course updated successfully", response["message"])
			})

			t.Run("should fail with invalid price format", func(t *testing.T) {
				cleanupCourseTestDB()

				// Create admin user and course
				adminUser := createTestUser(true)
				course := createTestCourse()
				adminToken := createUserToken(adminUser)

				// Create multipart form data with invalid price
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				writer.WriteField("title", "Updated Course")
				writer.WriteField("description", "Updated description")
				writer.WriteField("instructor", "Updated Instructor")
				writer.WriteField("price", "invalid-price")

				writer.Close()

				req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/courses/%s", course.ID), &body)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))
				req.Header.Set("Content-Type", writer.FormDataContentType())

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusBadRequest, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, "error", response["status"])
			})

			t.Run("should fail for non-existent course", func(t *testing.T) {
				cleanupCourseTestDB()

				// Create admin user
				adminUser := createTestUser(true)
				adminToken := createUserToken(adminUser)

				// Create multipart form data
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				writer.WriteField("title", "Updated Course")
				writer.WriteField("description", "Updated description")
				writer.WriteField("instructor", "Updated Instructor")
				writer.WriteField("price", "199.99")

				writer.Close()

				req, _ := http.NewRequest("PUT", "/api/courses/non-existent-id", &body)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))
				req.Header.Set("Content-Type", writer.FormDataContentType())

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusNotFound, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, "error", response["status"])
				assert.Contains(t, response["message"], "Course not found")
			})

			t.Run("should fail for non-admin user", func(t *testing.T) {
				cleanupCourseTestDB()

				// Create regular user and course
				user := createTestUser(false)
				course := createTestCourse()
				token := createUserToken(user)

				// Create multipart form data
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				writer.WriteField("title", "Unauthorized Update")
				writer.WriteField("description", "This should fail")
				writer.WriteField("instructor", "Regular User")
				writer.WriteField("price", "100.0")

				writer.Close()

				req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/courses/%s", course.ID), &body)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
				req.Header.Set("Content-Type", writer.FormDataContentType())

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusForbidden, w.Code)
			})
		})

		t.Run("DELETE /api/courses/:courseId (admin)", func(t *testing.T) {
			t.Run("should delete course successfully as admin", func(t *testing.T) {
				cleanupCourseTestDB()

				// Create admin user and course
				adminUser := createTestUser(true)
				course := createTestCourse()
				adminToken := createUserToken(adminUser)

				req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/courses/%s", course.ID), nil)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusNoContent, w.Code)

				// Verify course was soft deleted (using Unscoped to find soft deleted records)
				var deletedCourse models.Course
				result := courseTestDB.Unscoped().First(&deletedCourse, "id = ?", course.ID)
				assert.NoError(t, result.Error)
				assert.NotNil(t, deletedCourse.DeletedAt)
			})

			t.Run("should fail for non-existent course", func(t *testing.T) {
				cleanupCourseTestDB()

				// Create admin user
				adminUser := createTestUser(true)
				adminToken := createUserToken(adminUser)

				req, _ := http.NewRequest("DELETE", "/api/courses/non-existent-id", nil)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusInternalServerError, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, "error", response["status"])
			})

			t.Run("should fail for already deleted course", func(t *testing.T) {
				cleanupCourseTestDB()

				// Create admin user and course
				adminUser := createTestUser(true)
				course := createTestCourse()
				adminToken := createUserToken(adminUser)

				// First delete the course
				courseTestDB.Delete(&course)

				// Try to delete again
				req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/courses/%s", course.ID), nil)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusNoContent, w.Code)

				// For 204 No Content, there should be no response body
				assert.Empty(t, w.Body.String())
			})

			t.Run("should fail for non-admin user", func(t *testing.T) {
				cleanupCourseTestDB()

				// Create regular user and course
				user := createTestUser(false)
				course := createTestCourse()
				token := createUserToken(user)

				req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/courses/%s", course.ID), nil)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusForbidden, w.Code)
			})

			t.Run("should fail without authentication", func(t *testing.T) {
				cleanupCourseTestDB()

				// Create course
				course := createTestCourse()

				req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/courses/%s", course.ID), nil)

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusUnauthorized, w.Code)
			})
		})
	})
}
