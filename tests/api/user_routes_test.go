package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"yonatan/labpro/config"
	apiAdminUserControllers "yonatan/labpro/controllers/api/admin"
	"yonatan/labpro/database"
	"yonatan/labpro/models"
	apiRoutes "yonatan/labpro/routes/api"
	"yonatan/labpro/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var userTestDB *gorm.DB

func setupUserTestDB() {
	cfg := config.LoadTestWithProjectRoot()

	var err error
	userTestDB, err = gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to test database: " + err.Error())
	}

	// Set the global database instance
	database.DB = userTestDB

	// Auto migrate the schema
	err = userTestDB.AutoMigrate(
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

func cleanupUserTestDB() {
	// Clean up test data in correct order due to foreign key constraints
	userTestDB.Exec("DELETE FROM user_module_progresses")
	userTestDB.Exec("DELETE FROM user_courses")
	userTestDB.Exec("DELETE FROM modules")
	userTestDB.Exec("DELETE FROM courses")
	userTestDB.Exec("DELETE FROM users")
}

func setupUserTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	cfg := config.LoadTestWithProjectRoot()

	// Initialize services
	userService := services.NewUserService(userTestDB)

	// Initialize controllers
	adminUserController := apiAdminUserControllers.NewUserAPIController(userService)

	api := router.Group("/api")
	apiRoutes.SetupUserRoutes(api, adminUserController, cfg)

	return router
}

func createUserTestUser(email, username string, isAdmin bool) models.User {
	user := models.User{
		Username:  username,
		Email:     email,
		FirstName: "Test",
		LastName:  "User",
		Balance:   1000.0,
		IsAdmin:   isAdmin,
	}
	user.SetPassword("password123")
	userTestDB.Create(&user)
	return user
}

func createUserTestToken(user models.User) string {
	cfg := config.LoadTestWithProjectRoot()
	authService := services.NewAuthService(cfg)
	token, _, _ := authService.Login(user.Username, "password123")
	return token
}

func TestGetUsers(t *testing.T) {
	// Setup
	setupUserTestDB()
	defer cleanupUserTestDB()
	router := setupUserTestRouter()

	// Create test admin user
	adminUser := createUserTestUser("admin@test.com", "adminuser", true)

	// Create regular test users for listing
	regularUser1 := createUserTestUser("user1@test.com", "testuser1", false)
	regularUser2 := createUserTestUser("user2@test.com", "testuser2", false)

	// Generate admin token
	adminToken := createUserTestToken(adminUser)

	tests := []struct {
		name            string
		token           string
		queryParams     string
		expectedStatus  int
		expectedUsers   bool
		checkPagination bool
	}{
		{
			name:            "Get users successfully as admin",
			token:           adminToken,
			queryParams:     "",
			expectedStatus:  http.StatusOK,
			expectedUsers:   true,
			checkPagination: true,
		},
		{
			name:            "Get users with search query",
			token:           adminToken,
			queryParams:     "?q=testuser1",
			expectedStatus:  http.StatusOK,
			expectedUsers:   true,
			checkPagination: true,
		},
		{
			name:            "Get users with pagination",
			token:           adminToken,
			queryParams:     "?page=1&limit=2",
			expectedStatus:  http.StatusOK,
			expectedUsers:   true,
			checkPagination: true,
		},
		{
			name:            "Unauthorized without token",
			token:           "",
			queryParams:     "",
			expectedStatus:  http.StatusUnauthorized,
			expectedUsers:   false,
			checkPagination: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/users"+tt.queryParams, nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedUsers {
				assert.Equal(t, "success", response["status"])
				assert.Equal(t, "Users retrieved successfully", response["message"])
				assert.NotNil(t, response["data"])

				if tt.checkPagination {
					assert.NotNil(t, response["pagination"])
				}
			}
		})
	}

	// Cleanup test users
	userTestDB.Delete(&regularUser1)
	userTestDB.Delete(&regularUser2)
	userTestDB.Delete(&adminUser)
}

func TestUpdateUserBalance(t *testing.T) {
	// Setup
	setupUserTestDB()
	defer cleanupUserTestDB()
	router := setupUserTestRouter()

	// Create test admin user
	adminUser := createUserTestUser("admin@test.com", "adminuser", true)

	// Create regular test user to update balance
	targetUser := createUserTestUser("target@test.com", "targetuser", false)
	targetUser.Balance = 100.0
	userTestDB.Save(&targetUser)

	// Generate admin token
	adminToken := createUserTestToken(adminUser)

	tests := []struct {
		name            string
		userID          string
		token           string
		requestBody     map[string]interface{}
		expectedStatus  int
		expectedSuccess bool
		expectedBalance *float64
	}{
		{
			name:            "Update user balance successfully with positive increment",
			userID:          targetUser.ID,
			token:           adminToken,
			requestBody:     map[string]interface{}{"increment": 50.0},
			expectedStatus:  http.StatusOK,
			expectedSuccess: true,
			expectedBalance: &[]float64{150.0}[0],
		},
		{
			name:            "Update user balance successfully with negative increment",
			userID:          targetUser.ID,
			token:           adminToken,
			requestBody:     map[string]interface{}{"increment": -25.0},
			expectedStatus:  http.StatusOK,
			expectedSuccess: true,
			expectedBalance: &[]float64{125.0}[0],
		},
		{
			name:            "Missing increment in request body",
			userID:          targetUser.ID,
			token:           adminToken,
			requestBody:     map[string]interface{}{},
			expectedStatus:  http.StatusBadRequest,
			expectedSuccess: false,
		},
		{
			name:            "Unauthorized without token",
			userID:          targetUser.ID,
			token:           "",
			requestBody:     map[string]interface{}{"increment": 50.0},
			expectedStatus:  http.StatusUnauthorized,
			expectedSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/users/"+tt.userID+"/balance", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedSuccess {
				assert.Equal(t, "success", response["status"])
				assert.Equal(t, "User balance updated successfully", response["message"])
				assert.NotNil(t, response["data"])

				userData := response["data"].(map[string]interface{})
				assert.Equal(t, targetUser.ID, userData["id"])
				assert.Equal(t, targetUser.Username, userData["username"])

				if tt.expectedBalance != nil {
					actualBalance := userData["balance"].(float64)
					assert.InDelta(t, *tt.expectedBalance, actualBalance, 0.01)
				}
			} else if w.Code == http.StatusBadRequest {
				assert.Equal(t, "error", response["status"])
			}
		})
	}

	// Cleanup test users
	userTestDB.Delete(&targetUser)
	userTestDB.Delete(&adminUser)
}

func TestGetUserByID(t *testing.T) {
	setupUserTestDB()
	defer cleanupUserTestDB()

	router := setupUserTestRouter()

	t.Run("Get user by ID successfully as admin", func(t *testing.T) {
		cleanupUserTestDB()

		// Create admin user and target user
		adminUser := createUserTestUser("admin@test.com", "adminuser", true)
		targetUser := createUserTestUser("target@test.com", "targetuser", false)
		adminToken := createUserTestToken(adminUser)

		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/users/%s", targetUser.ID), nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "success", response["status"])
		assert.Equal(t, "User retrieved successfully", response["message"])

		userData := response["data"].(map[string]interface{})
		assert.Equal(t, targetUser.ID, userData["id"])
		assert.Equal(t, targetUser.Username, userData["username"])
	})

	t.Run("Return 404 for non-existent user", func(t *testing.T) {
		cleanupUserTestDB()

		adminUser := createUserTestUser("admin@test.com", "adminuser", true)
		adminToken := createUserTestToken(adminUser)

		req, _ := http.NewRequest("GET", "/api/users/non-existent-id", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "error", response["status"])
		assert.Equal(t, "User not found", response["message"])
	})

	t.Run("Unauthorized without token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/users/some-id", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Forbidden for non-admin user", func(t *testing.T) {
		cleanupUserTestDB()

		regularUser := createUserTestUser("user@test.com", "regularuser", false)
		targetUser := createUserTestUser("target@test.com", "targetuser", false)
		userToken := createUserTestToken(regularUser)

		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/users/%s", targetUser.ID), nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userToken))

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestUpdateUser(t *testing.T) {
	setupUserTestDB()
	defer cleanupUserTestDB()

	router := setupUserTestRouter()

	t.Run("Update user successfully as admin", func(t *testing.T) {
		cleanupUserTestDB()

		// Create admin user and target user
		adminUser := createUserTestUser("admin@test.com", "adminuser", true)
		targetUser := createUserTestUser("target@test.com", "targetuser", false)
		adminToken := createUserTestToken(adminUser)

		updateData := map[string]interface{}{
			"email":      "updated@example.com",
			"username":   "updateduser",
			"first_name": "Updated",
			"last_name":  "User",
		}

		jsonData, _ := json.Marshal(updateData)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/users/%s", targetUser.ID), bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "success", response["status"])
		assert.Equal(t, "User updated successfully", response["message"])

		userData := response["data"].(map[string]interface{})
		assert.Equal(t, "updateduser", userData["username"])
		assert.Equal(t, "Updated", userData["first_name"])
		assert.Equal(t, "User", userData["last_name"])
	})

	t.Run("Update user with password", func(t *testing.T) {
		cleanupUserTestDB()

		adminUser := createUserTestUser("admin@test.com", "adminuser", true)
		targetUser := createUserTestUser("target@test.com", "targetuser", false)
		adminToken := createUserTestToken(adminUser)

		updateData := map[string]interface{}{
			"email":      "updated2@example.com",
			"username":   "updateduser2",
			"first_name": "Updated2",
			"last_name":  "User2",
			"password":   "newpassword123",
		}

		jsonData, _ := json.Marshal(updateData)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/users/%s", targetUser.ID), bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Fail with invalid email format", func(t *testing.T) {
		cleanupUserTestDB()

		adminUser := createUserTestUser("admin@test.com", "adminuser", true)
		targetUser := createUserTestUser("target@test.com", "targetuser", false)
		adminToken := createUserTestToken(adminUser)

		updateData := map[string]interface{}{
			"email":      "invalid-email",
			"username":   "updateduser",
			"first_name": "Updated",
			"last_name":  "User",
		}

		jsonData, _ := json.Marshal(updateData)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/users/%s", targetUser.ID), bytes.NewBuffer(jsonData))
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

	t.Run("Fail with missing required fields", func(t *testing.T) {
		cleanupUserTestDB()

		adminUser := createUserTestUser("admin@test.com", "adminuser", true)
		targetUser := createUserTestUser("target@test.com", "targetuser", false)
		adminToken := createUserTestToken(adminUser)

		updateData := map[string]interface{}{
			"email": "updated@example.com",
			// Missing username, first_name, last_name
		}

		jsonData, _ := json.Marshal(updateData)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/users/%s", targetUser.ID), bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Fail for non-existent user", func(t *testing.T) {
		cleanupUserTestDB()

		adminUser := createUserTestUser("admin@test.com", "adminuser", true)
		adminToken := createUserTestToken(adminUser)

		updateData := map[string]interface{}{
			"email":      "updated@example.com",
			"username":   "updateduser",
			"first_name": "Updated",
			"last_name":  "User",
		}

		jsonData, _ := json.Marshal(updateData)
		req, _ := http.NewRequest("PUT", "/api/users/non-existent-id", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "error", response["status"])
		assert.Contains(t, response["message"], "Failed to update user")
	})

	t.Run("Unauthorized without token", func(t *testing.T) {
		updateData := map[string]interface{}{
			"email":      "updated@example.com",
			"username":   "updateduser",
			"first_name": "Updated",
			"last_name":  "User",
		}

		jsonData, _ := json.Marshal(updateData)
		req, _ := http.NewRequest("PUT", "/api/users/some-id", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Forbidden for non-admin user", func(t *testing.T) {
		cleanupUserTestDB()

		regularUser := createUserTestUser("user@test.com", "regularuser", false)
		targetUser := createUserTestUser("target@test.com", "targetuser", false)
		userToken := createUserTestToken(regularUser)

		updateData := map[string]interface{}{
			"email":      "updated@example.com",
			"username":   "updateduser",
			"first_name": "Updated",
			"last_name":  "User",
		}

		jsonData, _ := json.Marshal(updateData)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/users/%s", targetUser.ID), bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userToken))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestDeleteUser(t *testing.T) {
	// Setup
	setupUserTestDB()
	defer cleanupUserTestDB()
	router := setupUserTestRouter()

	// Create test admin user
	adminUser := createUserTestUser("admin@test.com", "adminuser", true)

	// Generate admin token
	adminToken := createUserTestToken(adminUser)

	tests := []struct {
		name           string
		userID         string
		token          string
		expectedStatus int
		setupUser      bool
	}{
		{
			name:           "Delete user successfully",
			userID:         "", // Will be set in test
			token:          adminToken,
			expectedStatus: http.StatusNoContent,
			setupUser:      true,
		},
		{
			name:           "Admin cannot delete themselves",
			userID:         "", // Will be set to adminUser.ID
			token:          adminToken,
			expectedStatus: http.StatusBadRequest,
			setupUser:      false,
		},
		{
			name:           "Unauthorized without token",
			userID:         "some-user-id",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			setupUser:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var targetUser models.User

			// Setup user for deletion if needed
			if tt.setupUser {
				targetUser = createUserTestUser(fmt.Sprintf("delete%d@test.com", len(tt.name)), fmt.Sprintf("deleteuser%d", len(tt.name)), false)
				tt.userID = targetUser.ID
			} else if tt.name == "Admin cannot delete themselves" {
				tt.userID = adminUser.ID
			}

			req, _ := http.NewRequest("DELETE", "/api/users/"+tt.userID, nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if w.Code == http.StatusBadRequest {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "error", response["status"])
				assert.Equal(t, "Cannot delete your own account", response["message"])
			}

			// Verify user was actually deleted for successful cases
			if tt.setupUser && w.Code == http.StatusNoContent {
				var deletedUser models.User
				result := userTestDB.First(&deletedUser, "id = ?", targetUser.ID)
				assert.Error(t, result.Error) // Should not find the user
			}
		})
	}

	// Cleanup admin user
	userTestDB.Delete(&adminUser)
}
