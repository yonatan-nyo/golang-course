package main

import (
	"log"
	"yonatan/labpro/config"
	"yonatan/labpro/database"
	"yonatan/labpro/router"

	"github.com/gin-gonic/gin"
)

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
	r := router.SetupRouter()

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
