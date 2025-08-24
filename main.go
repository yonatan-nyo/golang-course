package main

import (
	"log"
	"yonatan/labpro/config"
	"yonatan/labpro/database"
	_ "yonatan/labpro/docs"
	"yonatan/labpro/router"

	"github.com/gin-gonic/gin"
)

// @title           Labpro API
// @version         1.0
// @description     This is a learning management system API server.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	database.Init(cfg.DatabaseURL)

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Setup router
	r := router.SetupRouter(cfg)

	// Set maximum multipart memory (100 MB for video uploads)
	r.MaxMultipartMemory = 100 << 20

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
