package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"yonatan/labpro/config"
	apiControllers "yonatan/labpro/controllers/api"
	"yonatan/labpro/database"
	"yonatan/labpro/models"
	apiRoutes "yonatan/labpro/routes/api"
	"yonatan/labpro/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var testDB *gorm.DB

func setupTestDB() {
	cfg := config.LoadTestWithProjectRoot()

	var err error
	testDB, err = gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to test database: " + err.Error())
	}

	// Set the global database instance
	database.DB = testDB

	// Auto migrate the schema
	err = testDB.AutoMigrate(
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

func cleanupTestDB() {
	// Clean up test data
	testDB.Exec("DELETE FROM user_module_progresses")
	testDB.Exec("DELETE FROM user_courses")
	testDB.Exec("DELETE FROM modules")
	testDB.Exec("DELETE FROM courses")
	testDB.Exec("DELETE FROM users")
}

func setupAuthTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	cfg := config.LoadTestWithProjectRoot()
	authService := services.NewAuthService(cfg)
	authController := apiControllers.NewAuthAPIController(authService)

	api := router.Group("/api")
	apiRoutes.SetupAuthRoutes(api, authController, cfg)

	return router
}

func TestAuthRoutes(t *testing.T) {
	// Setup test database
	setupTestDB()
	defer cleanupTestDB()

	router := setupAuthTestRouter()

	t.Run("POST /api/auth/register", func(t *testing.T) {
		t.Run("should register user successfully with valid data", func(t *testing.T) {
			cleanupTestDB() // Clean before test

			reqBody := map[string]interface{}{
				"username":         "testuser",
				"email":            "test@example.com",
				"first_name":       "Test",
				"last_name":        "User",
				"password":         "password123",
				"confirm_password": "password123",
			}

			jsonData, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "success", response["status"])
			assert.Equal(t, "Registration successful", response["message"])
			assert.NotNil(t, response["data"])

			data := response["data"].(map[string]interface{})
			assert.Equal(t, "testuser", data["username"])
			assert.Equal(t, "Test", data["first_name"])
			assert.Equal(t, "User", data["last_name"])
			assert.NotEmpty(t, data["id"])
		})

		t.Run("should fail when passwords don't match", func(t *testing.T) {
			cleanupTestDB() // Clean before test

			reqBody := map[string]interface{}{
				"username":         "testuser2",
				"email":            "test2@example.com",
				"first_name":       "Test",
				"last_name":        "User",
				"password":         "password123",
				"confirm_password": "differentpassword",
			}

			jsonData, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "error", response["status"])
			assert.Equal(t, "Password and confirm password do not match", response["message"])
		})

		t.Run("should fail with missing required fields", func(t *testing.T) {
			reqBody := map[string]interface{}{
				"username": "testuser3",
				// missing email, first_name, last_name, password, confirm_password
			}

			jsonData, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "error", response["status"])
			assert.Contains(t, response["message"], "required")
		})

		t.Run("should fail with duplicate username", func(t *testing.T) {
			cleanupTestDB() // Clean before test

			// Create first user
			user := models.User{
				Username:  "duplicateuser",
				Email:     "first@example.com",
				FirstName: "First",
				LastName:  "User",
			}
			user.SetPassword("password123")
			testDB.Create(&user)

			// Try to create another user with same username
			reqBody := map[string]interface{}{
				"username":         "duplicateuser",
				"email":            "second@example.com",
				"first_name":       "Second",
				"last_name":        "User",
				"password":         "password123",
				"confirm_password": "password123",
			}

			jsonData, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "error", response["status"])
			assert.Equal(t, "username already exists", response["message"])
		})

		t.Run("should fail with duplicate email", func(t *testing.T) {
			cleanupTestDB() // Clean before test

			// Create first user
			user := models.User{
				Username:  "firstuser",
				Email:     "duplicate@example.com",
				FirstName: "First",
				LastName:  "User",
			}
			user.SetPassword("password123")
			testDB.Create(&user)

			// Try to create another user with same email
			reqBody := map[string]interface{}{
				"username":         "seconduser",
				"email":            "duplicate@example.com",
				"first_name":       "Second",
				"last_name":        "User",
				"password":         "password123",
				"confirm_password": "password123",
			}

			jsonData, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "error", response["status"])
			assert.Equal(t, "email already exists", response["message"])
		})
	})

	t.Run("POST /api/auth/login", func(t *testing.T) {
		t.Run("should login successfully with valid username", func(t *testing.T) {
			cleanupTestDB() // Clean before test

			// Create test user
			user := models.User{
				Username:  "loginuser",
				Email:     "login@example.com",
				FirstName: "Login",
				LastName:  "User",
			}
			user.SetPassword("password123")
			testDB.Create(&user)

			reqBody := map[string]interface{}{
				"identifier": "loginuser",
				"password":   "password123",
			}

			jsonData, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "success", response["status"])
			assert.Equal(t, "Login successful", response["message"])
			assert.NotNil(t, response["data"])

			data := response["data"].(map[string]interface{})
			assert.Equal(t, "loginuser", data["username"])
			assert.NotEmpty(t, data["token"])
		})

		t.Run("should login successfully with valid email", func(t *testing.T) {
			cleanupTestDB() // Clean before test

			// Create test user
			user := models.User{
				Username:  "emailuser",
				Email:     "email@example.com",
				FirstName: "Email",
				LastName:  "User",
			}
			user.SetPassword("password123")
			testDB.Create(&user)

			reqBody := map[string]interface{}{
				"identifier": "email@example.com",
				"password":   "password123",
			}

			jsonData, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "success", response["status"])
			assert.Equal(t, "Login successful", response["message"])
			assert.NotNil(t, response["data"])

			data := response["data"].(map[string]interface{})
			assert.Equal(t, "emailuser", data["username"])
			assert.NotEmpty(t, data["token"])
		})

		t.Run("should fail with invalid credentials", func(t *testing.T) {
			cleanupTestDB() // Clean before test

			// Create test user
			user := models.User{
				Username:  "testuser",
				Email:     "test@example.com",
				FirstName: "Test",
				LastName:  "User",
			}
			user.SetPassword("password123")
			testDB.Create(&user)

			reqBody := map[string]interface{}{
				"identifier": "testuser",
				"password":   "wrongpassword",
			}

			jsonData, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "error", response["status"])
			assert.Equal(t, "invalid credentials", response["message"])
		})

		t.Run("should fail with non-existent user", func(t *testing.T) {
			cleanupTestDB() // Clean before test

			reqBody := map[string]interface{}{
				"identifier": "nonexistentuser",
				"password":   "password123",
			}

			jsonData, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "error", response["status"])
			assert.Equal(t, "invalid credentials", response["message"])
		})

		t.Run("should fail with missing required fields", func(t *testing.T) {
			reqBody := map[string]interface{}{
				"identifier": "testuser",
				// missing password
			}

			jsonData, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "error", response["status"])
			assert.Contains(t, response["message"], "required")
		})
	})

	t.Run("POST /api/auth/logout", func(t *testing.T) {
		t.Run("should logout successfully", func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/api/auth/logout", nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "success", response["status"])
			assert.Equal(t, "Logout successful", response["message"])
			assert.Nil(t, response["data"])
		})
	})

	t.Run("GET /api/auth/self", func(t *testing.T) {
		t.Run("should get profile successfully with valid token", func(t *testing.T) {
			cleanupTestDB() // Clean before test

			// Create test user
			user := models.User{
				Username:  "profileuser",
				Email:     "profile@example.com",
				FirstName: "Profile",
				LastName:  "User",
				Balance:   100.0,
			}
			user.SetPassword("password123")
			testDB.Create(&user)

			// Login to get token
			loginReqBody := map[string]interface{}{
				"identifier": "profileuser",
				"password":   "password123",
			}

			jsonData, _ := json.Marshal(loginReqBody)
			loginReq, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
			loginReq.Header.Set("Content-Type", "application/json")

			loginW := httptest.NewRecorder()
			router.ServeHTTP(loginW, loginReq)

			assert.Equal(t, http.StatusOK, loginW.Code)

			var loginResponse map[string]interface{}
			err := json.Unmarshal(loginW.Body.Bytes(), &loginResponse)
			assert.NoError(t, err)

			loginData := loginResponse["data"].(map[string]interface{})
			token := loginData["token"].(string)

			// Use token to get profile
			profileReq, _ := http.NewRequest("GET", "/api/auth/self", nil)
			profileReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			profileW := httptest.NewRecorder()
			router.ServeHTTP(profileW, profileReq)

			assert.Equal(t, http.StatusOK, profileW.Code)

			var profileResponse map[string]interface{}
			err = json.Unmarshal(profileW.Body.Bytes(), &profileResponse)
			assert.NoError(t, err)

			assert.Equal(t, "success", profileResponse["status"])
			assert.Equal(t, "Profile retrieved successfully", profileResponse["message"])
			assert.NotNil(t, profileResponse["data"])

			data := profileResponse["data"].(map[string]interface{})
			assert.Equal(t, "profileuser", data["username"])
			assert.Equal(t, "profile@example.com", data["email"])
			assert.Equal(t, "Profile", data["first_name"])
			assert.Equal(t, "User", data["last_name"])
			assert.Equal(t, 100.0, data["balance"])
			assert.NotEmpty(t, data["id"])
		})

		t.Run("should fail without authorization header", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/auth/self", nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})

		t.Run("should fail with invalid token", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/auth/self", nil)
			req.Header.Set("Authorization", "Bearer invalid-token")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})

		t.Run("should fail with malformed authorization header", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/auth/self", nil)
			req.Header.Set("Authorization", "InvalidFormat")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	})
}
