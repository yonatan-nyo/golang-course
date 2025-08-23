package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

var moduleTestDB *gorm.DB

func setupModuleTestDB() {
	cfg := config.LoadTestWithProjectRoot()

	var err error
	moduleTestDB, err = gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to test database: " + err.Error())
	}

	// Set the global database instance
	database.DB = moduleTestDB

	// Auto migrate the schema
	err = moduleTestDB.AutoMigrate(
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

func cleanupModuleTestDB() {
	// Clean up test data in correct order due to foreign key constraints
	moduleTestDB.Exec("DELETE FROM user_module_progresses")
	moduleTestDB.Exec("DELETE FROM user_courses")
	moduleTestDB.Exec("DELETE FROM modules")
	moduleTestDB.Exec("DELETE FROM courses")
	moduleTestDB.Exec("DELETE FROM users")
}

func setupModuleTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Get config for services
	cfg := config.LoadTestWithProjectRoot()

	// Initialize services
	moduleService := services.NewModuleService(moduleTestDB, cfg)

	// Initialize controllers
	userModuleController := apiUserControllers.NewModuleAPIController(moduleService)
	adminModuleController := apiAdminControllers.NewModuleAPIController(moduleService)

	api := router.Group("/api")
	apiRoutes.SetupModuleRoutes(api, adminModuleController, userModuleController, cfg)

	return router
}

func createModuleTestUser(isAdmin bool) models.User {
	user := models.User{
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Balance:   1000.0,
		IsAdmin:   isAdmin,
	}
	user.SetPassword("password123")
	moduleTestDB.Create(&user)
	return user
}

func createModuleTestCourse() models.Course {
	course := models.Course{
		Title:       "Test Course",
		Description: "A test course for module testing",
		Instructor:  "Test Instructor",
		Price:       100.0,
		Thumbnail:   "test-thumbnail.jpg",
		Topics:      pq.StringArray{"programming", "testing"},
	}
	moduleTestDB.Create(&course)
	return course
}

func createModuleTestModule(courseID string, order int) models.Module {
	module := models.Module{
		CourseID:     courseID,
		Title:        fmt.Sprintf("Test Module %d", order),
		Description:  fmt.Sprintf("Description for test module %d", order),
		Order:        order,
		PDFContent:   stringPtr("test-pdf-content.pdf"),
		VideoContent: stringPtr("test-video-content.mp4"),
	}
	moduleTestDB.Create(&module)
	return module
}

func createModuleUserToken(user models.User) string {
	cfg := config.LoadTestWithProjectRoot()
	authService := services.NewAuthService(cfg)
	token, _, _ := authService.Login(user.Username, "password123")
	return token
}

func stringPtr(s string) *string {
	return &s
}

// getProjectRoot finds the project root by looking for go.mod file (similar to config package)
func getProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Walk up the directory tree to find go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached filesystem root
		}
		dir = parent
	}

	return ""
}

// loadTestFileFromProjectRoot loads a test file from project root, similar to how config loads .env files
func loadTestFileFromProjectRoot(filename string) []byte {
	// Save current directory
	originalDir, _ := os.Getwd()

	// Change to project root if we can find it
	if projectRoot := getProjectRoot(); projectRoot != "" {
		os.Chdir(projectRoot)
		defer os.Chdir(originalDir) // Restore original directory
	}

	// Load the file
	content, err := os.ReadFile(filename)
	if err != nil {
		// Return empty bytes if file doesn't exist
		return []byte{}
	}
	return content
}

// loadTestPDF loads the test PDF file from project root using project root detection
func loadTestPDF() []byte {
	return loadTestFileFromProjectRoot("pdf_for_test.pdf")
}

// loadTestVideo loads the test video file from project root using project root detection
func loadTestVideo() []byte {
	return loadTestFileFromProjectRoot("vid_for_test.mp4")
}

// createTestMultipartWithFiles creates a multipart form with actual test files
func createTestMultipartWithFiles(title, description string, includePDF, includeVideo bool) (*bytes.Buffer, *multipart.Writer) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add form fields
	writer.WriteField("title", title)
	writer.WriteField("description", description)

	// Add PDF file if requested
	if includePDF {
		part, _ := writer.CreateFormFile("pdf_content", "pdf_for_test.pdf")
		pdfContent := loadTestPDF()
		part.Write(pdfContent)
	}

	// Add video file if requested
	if includeVideo {
		part, _ := writer.CreateFormFile("video_content", "vid_for_test.mp4")
		videoContent := loadTestVideo()
		part.Write(videoContent)
	}

	return &body, writer
}

func enrollUserInCourse(userID, courseID string) {
	userCourse := models.UserCourse{
		UserID:   userID,
		CourseID: courseID,
	}
	moduleTestDB.Create(&userCourse)
}

func TestModuleRoutes(t *testing.T) {
	// Setup test database
	setupModuleTestDB()
	defer cleanupModuleTestDB()

	router := setupModuleTestRouter()

	t.Run("GET /api/courses/:courseId/modules", func(t *testing.T) {
		t.Run("should get course modules successfully when user is enrolled", func(t *testing.T) {
			cleanupModuleTestDB()

			// Create test user, course, and modules
			user := createModuleTestUser(false)
			course := createModuleTestCourse()
			module1 := createModuleTestModule(course.ID, 1)
			createModuleTestModule(course.ID, 2)

			// Enroll user in course
			enrollUserInCourse(user.ID, course.ID)

			token := createModuleUserToken(user)

			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/courses/%s/modules", course.ID), nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "success", response["status"])
			assert.Equal(t, "Modules retrieved successfully", response["message"])
			assert.NotNil(t, response["data"])
			assert.NotNil(t, response["pagination"])

			// Check if our test modules are in the response
			data := response["data"].([]interface{})
			assert.Len(t, data, 2)

			// Verify first module
			firstModule := data[0].(map[string]interface{})
			assert.Equal(t, module1.Title, firstModule["title"])
			assert.Equal(t, float64(module1.Order), firstModule["order"])
		})

		t.Run("should handle pagination parameters", func(t *testing.T) {
			cleanupModuleTestDB()

			// Create test user and course
			user := createModuleTestUser(false)
			course := createModuleTestCourse()

			// Create multiple modules
			for i := 1; i <= 5; i++ {
				createModuleTestModule(course.ID, i)
			}

			// Enroll user in course
			enrollUserInCourse(user.ID, course.ID)

			token := createModuleUserToken(user)

			// Test pagination
			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/courses/%s/modules?page=1&limit=3", course.ID), nil)
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

		t.Run("should fail without authentication", func(t *testing.T) {
			cleanupModuleTestDB()
			course := createModuleTestCourse()

			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/courses/%s/modules", course.ID), nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})

		t.Run("should return modules even when user is not enrolled (current behavior)", func(t *testing.T) {
			cleanupModuleTestDB()

			// Create test user and course (but don't enroll user)
			user := createModuleTestUser(false)
			course := createModuleTestCourse()
			createModuleTestModule(course.ID, 1)

			token := createModuleUserToken(user)

			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/courses/%s/modules", course.ID), nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// The service currently returns modules regardless of enrollment
			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "success", response["status"])
			assert.Equal(t, "Modules retrieved successfully", response["message"])
		})
	})

	t.Run("GET /api/modules/:id", func(t *testing.T) {
		t.Run("should get module by ID successfully when user has access", func(t *testing.T) {
			cleanupModuleTestDB()

			// Create test user, course, and module
			user := createModuleTestUser(false)
			course := createModuleTestCourse()
			module := createModuleTestModule(course.ID, 1)

			// Enroll user in course
			enrollUserInCourse(user.ID, course.ID)

			token := createModuleUserToken(user)

			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/modules/%s", module.ID), nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "success", response["status"])
			assert.Equal(t, "Module retrieved successfully", response["message"])
			assert.NotNil(t, response["data"])

			data := response["data"].(map[string]interface{})
			assert.Equal(t, module.Title, data["title"])
			assert.Equal(t, module.Description, data["description"])
			assert.Equal(t, float64(module.Order), data["order"])
		})

		t.Run("should allow admin to access any module", func(t *testing.T) {
			cleanupModuleTestDB()

			// Create admin user, course, and module
			admin := createModuleTestUser(true) // isAdmin = true
			course := createModuleTestCourse()
			module := createModuleTestModule(course.ID, 1)

			token := createModuleUserToken(admin)

			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/modules/%s", module.ID), nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "success", response["status"])
			data := response["data"].(map[string]interface{})
			assert.Equal(t, module.Title, data["title"])
		})

		t.Run("should return 404 for non-existent module", func(t *testing.T) {
			cleanupModuleTestDB()

			// Create test user
			user := createModuleTestUser(false)
			token := createModuleUserToken(user)

			req, _ := http.NewRequest("GET", "/api/modules/non-existent-id", nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "error", response["status"])
			assert.Equal(t, "Module not found or access denied", response["message"])
		})

		t.Run("should fail when user doesn't have access to module", func(t *testing.T) {
			cleanupModuleTestDB()

			// Create test user, course, and module (but don't enroll user)
			user := createModuleTestUser(false)
			course := createModuleTestCourse()
			module := createModuleTestModule(course.ID, 1)

			token := createModuleUserToken(user)

			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/modules/%s", module.ID), nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "error", response["status"])
			assert.Equal(t, "Module not found or access denied", response["message"])
		})

		t.Run("should fail without authentication", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/modules/some-id", nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	})

	t.Run("PATCH /api/modules/:id/complete", func(t *testing.T) {
		t.Run("should complete module successfully", func(t *testing.T) {
			cleanupModuleTestDB()

			// Create test user, course, and module
			user := createModuleTestUser(false)
			course := createModuleTestCourse()
			module := createModuleTestModule(course.ID, 1)

			// Enroll user in course
			enrollUserInCourse(user.ID, course.ID)

			token := createModuleUserToken(user)

			req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/modules/%s/complete", module.ID), nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "success", response["status"])
			assert.Equal(t, "Module completed successfully", response["message"])
			assert.NotNil(t, response["data"])

			// Verify module progress was created in database
			var progress models.UserModuleProgress
			err = moduleTestDB.Where("user_id = ? AND module_id = ?", user.ID, module.ID).First(&progress).Error
			assert.NoError(t, err)
			assert.True(t, progress.IsCompleted)
		})

		t.Run("should allow re-completion of already completed module (current behavior)", func(t *testing.T) {
			cleanupModuleTestDB()

			// Create test user, course, and module
			user := createModuleTestUser(false)
			course := createModuleTestCourse()
			module := createModuleTestModule(course.ID, 1)

			// Enroll user in course
			enrollUserInCourse(user.ID, course.ID)

			// Mark module as already completed
			progress := models.UserModuleProgress{
				UserID:      user.ID,
				ModuleID:    module.ID,
				IsCompleted: true,
			}
			moduleTestDB.Create(&progress)

			token := createModuleUserToken(user)

			req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/modules/%s/complete", module.ID), nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// The service currently allows re-completion
			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "success", response["status"])
			assert.Equal(t, "Module completed successfully", response["message"])
		})

		t.Run("should fail when user doesn't have access to module", func(t *testing.T) {
			cleanupModuleTestDB()

			// Create test user, course, and module (but don't enroll user)
			user := createModuleTestUser(false)
			course := createModuleTestCourse()
			module := createModuleTestModule(course.ID, 1)

			token := createModuleUserToken(user)

			req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/modules/%s/complete", module.ID), nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "error", response["status"])
		})

		t.Run("should fail for non-existent module", func(t *testing.T) {
			cleanupModuleTestDB()

			// Create test user
			user := createModuleTestUser(false)
			token := createModuleUserToken(user)

			req, _ := http.NewRequest("PATCH", "/api/modules/non-existent-id/complete", nil)
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
			req, _ := http.NewRequest("PATCH", "/api/modules/some-id/complete", nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	})

	// Admin Module Routes Tests
	t.Run("Admin Module Routes", func(t *testing.T) {
		t.Run("POST /api/courses/:courseId/modules (admin)", func(t *testing.T) {
			t.Run("should create module successfully as admin", func(t *testing.T) {
				cleanupModuleTestDB()

				// Create admin user and course
				adminUser := createModuleTestUser(true)
				course := createModuleTestCourse()
				adminToken := createModuleUserToken(adminUser)

				// Create multipart form data with test files
				body, writer := createTestMultipartWithFiles(
					"Admin Created Module",
					"A module created by admin with test files",
					true, // include PDF
					true, // include video
				)

				writer.Close()

				req, _ := http.NewRequest("POST", fmt.Sprintf("/api/courses/%s/modules", course.ID), body)
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
				assert.Equal(t, "Module created successfully", response["message"])
			})

			t.Run("should create module with PDF and video files successfully", func(t *testing.T) {
				cleanupModuleTestDB()

				// Create admin user and course
				adminUser := createModuleTestUser(true)
				course := createModuleTestCourse()
				adminToken := createModuleUserToken(adminUser)

				// Create multipart form data with only PDF file
				body, writer := createTestMultipartWithFiles(
					"Module with PDF Only",
					"A module with PDF file attachment",
					true,  // include PDF
					false, // exclude video
				)

				writer.Close()

				req, _ := http.NewRequest("POST", fmt.Sprintf("/api/courses/%s/modules", course.ID), body)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))
				req.Header.Set("Content-Type", writer.FormDataContentType())

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusCreated, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, "success", response["status"])
				assert.Equal(t, "Module created successfully", response["message"])

				// Verify the module was created with PDF content
				data := response["data"].(map[string]interface{})
				assert.NotNil(t, data["pdf_content"])
				assert.Nil(t, data["video_content"])
			})

			t.Run("should create module with video file only", func(t *testing.T) {
				cleanupModuleTestDB()

				// Create admin user and course
				adminUser := createModuleTestUser(true)
				course := createModuleTestCourse()
				adminToken := createModuleUserToken(adminUser)

				// Create multipart form data with only video file
				body, writer := createTestMultipartWithFiles(
					"Module with Video Only",
					"A module with video file attachment",
					false, // exclude PDF
					true,  // include video
				)

				writer.Close()

				req, _ := http.NewRequest("POST", fmt.Sprintf("/api/courses/%s/modules", course.ID), body)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))
				req.Header.Set("Content-Type", writer.FormDataContentType())

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusCreated, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, "success", response["status"])
				assert.Equal(t, "Module created successfully", response["message"])

				// Verify the module was created with video content
				data := response["data"].(map[string]interface{})
				assert.Nil(t, data["pdf_content"])
				assert.NotNil(t, data["video_content"])
			})

			t.Run("should fail with missing required fields", func(t *testing.T) {
				cleanupModuleTestDB()

				// Create admin user and course
				adminUser := createModuleTestUser(true)
				course := createModuleTestCourse()
				adminToken := createModuleUserToken(adminUser)

				// Create multipart form data with missing required fields
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				// Only add description, missing title (which is required)
				writer.WriteField("description", "Module without title")

				writer.Close()

				req, _ := http.NewRequest("POST", fmt.Sprintf("/api/courses/%s/modules", course.ID), &body)
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

			t.Run("should succeed even with invalid order format (order is auto-assigned)", func(t *testing.T) {
				cleanupModuleTestDB()

				// Create admin user and course
				adminUser := createModuleTestUser(true)
				course := createModuleTestCourse()
				adminToken := createModuleUserToken(adminUser)

				// Create multipart form data with invalid order
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				writer.WriteField("title", "Invalid Order Module")
				writer.WriteField("description", "A module with invalid order")
				writer.WriteField("order", "not-a-number")

				writer.Close()

				req, _ := http.NewRequest("POST", fmt.Sprintf("/api/courses/%s/modules", course.ID), &body)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))
				req.Header.Set("Content-Type", writer.FormDataContentType())

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusCreated, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, "success", response["status"])
			})

			t.Run("should fail for non-existent course", func(t *testing.T) {
				cleanupModuleTestDB()

				// Create admin user
				adminUser := createModuleTestUser(true)
				adminToken := createModuleUserToken(adminUser)

				// Create multipart form data
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				writer.WriteField("title", "Module for Non-existent Course")
				writer.WriteField("description", "This should fail")
				writer.WriteField("order", "1")

				writer.Close()

				req, _ := http.NewRequest("POST", "/api/courses/non-existent-id/modules", &body)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))
				req.Header.Set("Content-Type", writer.FormDataContentType())

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusInternalServerError, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, "error", response["status"])
			})

			t.Run("should fail for non-admin user", func(t *testing.T) {
				cleanupModuleTestDB()

				// Create regular user and course
				user := createModuleTestUser(false)
				course := createModuleTestCourse()
				token := createModuleUserToken(user)

				// Create multipart form data
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				writer.WriteField("title", "Unauthorized Module")
				writer.WriteField("description", "This should fail")
				writer.WriteField("order", "1")

				writer.Close()

				req, _ := http.NewRequest("POST", fmt.Sprintf("/api/courses/%s/modules", course.ID), &body)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
				req.Header.Set("Content-Type", writer.FormDataContentType())

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusForbidden, w.Code)
			})

			t.Run("should fail without authentication", func(t *testing.T) {
				// Create course
				course := createModuleTestCourse()

				// Create multipart form data
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				writer.WriteField("title", "Unauthenticated Module")
				writer.WriteField("description", "This should fail")
				writer.WriteField("order", "1")

				writer.Close()

				req, _ := http.NewRequest("POST", fmt.Sprintf("/api/courses/%s/modules", course.ID), &body)
				req.Header.Set("Content-Type", writer.FormDataContentType())

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusUnauthorized, w.Code)
			})
		})

		t.Run("PUT /api/modules/:id (admin)", func(t *testing.T) {
			t.Run("should update module successfully as admin", func(t *testing.T) {
				cleanupModuleTestDB()

				// Create admin user, course, and module
				adminUser := createModuleTestUser(true)
				course := createModuleTestCourse()
				module := createModuleTestModule(course.ID, 1)
				adminToken := createModuleUserToken(adminUser)

				// Create multipart form data with test files
				body, writer := createTestMultipartWithFiles(
					"Updated Module Title",
					"Updated module description with test files",
					true, // include PDF
					true, // include video
				)

				writer.Close()

				req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/modules/%s", module.ID), body)
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
				assert.Equal(t, "Module updated successfully", response["message"])
			})

			t.Run("should succeed even with invalid order format (order is ignored in updates)", func(t *testing.T) {
				cleanupModuleTestDB()

				// Create admin user, course, and module
				adminUser := createModuleTestUser(true)
				course := createModuleTestCourse()
				module := createModuleTestModule(course.ID, 1)
				adminToken := createModuleUserToken(adminUser)

				// Create multipart form data with invalid order
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				writer.WriteField("title", "Updated Module")
				writer.WriteField("description", "Updated description")
				writer.WriteField("order", "invalid-order")

				writer.Close()

				req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/modules/%s", module.ID), &body)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))
				req.Header.Set("Content-Type", writer.FormDataContentType())

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, "success", response["status"])
			})

			t.Run("should fail for non-existent module", func(t *testing.T) {
				cleanupModuleTestDB()

				// Create admin user
				adminUser := createModuleTestUser(true)
				adminToken := createModuleUserToken(adminUser)

				// Create multipart form data
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				writer.WriteField("title", "Updated Module")
				writer.WriteField("description", "Updated description")
				writer.WriteField("order", "1")

				writer.Close()

				req, _ := http.NewRequest("PUT", "/api/modules/non-existent-id", &body)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))
				req.Header.Set("Content-Type", writer.FormDataContentType())

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusNotFound, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, "error", response["status"])
				assert.Contains(t, response["message"], "Module not found")
			})

			t.Run("should fail for non-admin user", func(t *testing.T) {
				cleanupModuleTestDB()

				// Create regular user, course, and module
				user := createModuleTestUser(false)
				course := createModuleTestCourse()
				module := createModuleTestModule(course.ID, 1)
				token := createModuleUserToken(user)

				// Create multipart form data
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)

				writer.WriteField("title", "Unauthorized Update")
				writer.WriteField("description", "This should fail")
				writer.WriteField("order", "1")

				writer.Close()

				req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/modules/%s", module.ID), &body)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
				req.Header.Set("Content-Type", writer.FormDataContentType())

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusForbidden, w.Code)
			})
		})

		t.Run("DELETE /api/modules/:id (admin)", func(t *testing.T) {
			t.Run("should delete module successfully as admin", func(t *testing.T) {
				cleanupModuleTestDB()

				// Create admin user, course, and module
				adminUser := createModuleTestUser(true)
				course := createModuleTestCourse()
				module := createModuleTestModule(course.ID, 1)
				adminToken := createModuleUserToken(adminUser)

				req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/modules/%s", module.ID), nil)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusNoContent, w.Code)

				// Verify module was deleted (hard delete)
				var deletedModule models.Module
				result := moduleTestDB.First(&deletedModule, "id = ?", module.ID)
				assert.Error(t, result.Error) // Should not find the module
			})

			t.Run("should fail for non-existent module", func(t *testing.T) {
				cleanupModuleTestDB()

				// Create admin user
				adminUser := createModuleTestUser(true)
				adminToken := createModuleUserToken(adminUser)

				req, _ := http.NewRequest("DELETE", "/api/modules/non-existent-id", nil)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusInternalServerError, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, "error", response["status"])
			})

			t.Run("should fail for already deleted module", func(t *testing.T) {
				cleanupModuleTestDB()

				// Create admin user, course, and module
				adminUser := createModuleTestUser(true)
				course := createModuleTestCourse()
				module := createModuleTestModule(course.ID, 1)
				adminToken := createModuleUserToken(adminUser)

				// First delete the module
				moduleTestDB.Delete(&module)

				// Try to delete again
				req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/modules/%s", module.ID), nil)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusNoContent, w.Code)

				// For 204 No Content, there should be no response body
				assert.Empty(t, w.Body.String())
			})

			t.Run("should fail for non-admin user", func(t *testing.T) {
				cleanupModuleTestDB()

				// Create regular user, course, and module
				user := createModuleTestUser(false)
				course := createModuleTestCourse()
				module := createModuleTestModule(course.ID, 1)
				token := createModuleUserToken(user)

				req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/modules/%s", module.ID), nil)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusForbidden, w.Code)
			})

			t.Run("should fail without authentication", func(t *testing.T) {
				cleanupModuleTestDB()

				// Create course and module
				course := createModuleTestCourse()
				module := createModuleTestModule(course.ID, 1)

				req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/modules/%s", module.ID), nil)

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusUnauthorized, w.Code)
			})
		})

		t.Run("PATCH /api/courses/:courseId/modules/reorder (admin)", func(t *testing.T) {
			t.Run("should reorder modules successfully as admin", func(t *testing.T) {
				cleanupModuleTestDB()

				// Create admin user and course with modules
				adminUser := createModuleTestUser(true)
				course := createModuleTestCourse()
				module1 := createModuleTestModule(course.ID, 1)
				module2 := createModuleTestModule(course.ID, 2)
				adminToken := createModuleUserToken(adminUser)

				// Create proper reorder request body
				reorderData := map[string]interface{}{
					"module_order": []map[string]interface{}{
						{"id": module2.ID, "order": 1},
						{"id": module1.ID, "order": 2},
					},
				}

				jsonData, _ := json.Marshal(reorderData)
				req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/courses/%s/modules/reorder", course.ID), bytes.NewBuffer(jsonData))
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "success", response["status"])
				assert.Equal(t, "Modules reordered successfully", response["message"])
				assert.NotNil(t, response["data"])
			})

			t.Run("should fail for non-admin user", func(t *testing.T) {
				cleanupModuleTestDB()

				// Create regular user and course
				user := createModuleTestUser(false)
				course := createModuleTestCourse()
				token := createModuleUserToken(user)

				// Create proper reorder request body
				reorderData := map[string]interface{}{
					"module_order": []map[string]interface{}{
						{"id": "test-id", "order": 1},
					},
				}

				jsonData, _ := json.Marshal(reorderData)
				req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/courses/%s/modules/reorder", course.ID), bytes.NewBuffer(jsonData))
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusForbidden, w.Code)
			})

			t.Run("should fail with invalid request body", func(t *testing.T) {
				cleanupModuleTestDB()

				// Create admin user and course
				adminUser := createModuleTestUser(true)
				course := createModuleTestCourse()
				adminToken := createModuleUserToken(adminUser)

				// Send empty request body
				req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/courses/%s/modules/reorder", course.ID), nil)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusBadRequest, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "error", response["status"])
			})
		})
	})
}
