package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL   string
	RedisAddr     string
	RedisPassword string
	JWTSecret     string
	Port          string
	Environment   string
	BaseURL       string
	UploadPath    string
	MaxFileSize   string
}

func Load(envFiles ...string) *Config {
	// Determine which env file to load
	envFile := ".env" // default
	if len(envFiles) > 0 && envFiles[0] != "" {
		envFile = envFiles[0]
	}

	// Load specified env file
	if err := godotenv.Load(envFile); err != nil {
		log.Printf("Warning: %s file not found or could not be loaded: %v", envFile, err)
	}

	return &Config{
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://user:password@localhost/labpro_db?sslmode=disable"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key"),
		Port:          getEnv("PORT", "8080"),
		Environment:   getEnv("ENVIRONMENT", "development"),
		BaseURL:       getEnv("BASE_URL", "http://localhost:8080"),
		UploadPath:    getEnv("UPLOAD_PATH", "./uploads"),
		MaxFileSize:   getEnv("MAX_FILE_SIZE", "10485760"),
	}
}

// LoadTest loads configuration specifically for testing using .env.test
func LoadTest() *Config {
	return Load(".env.test")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
