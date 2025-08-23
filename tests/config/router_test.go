package api

import (
	"testing"
	"yonatan/labpro/config"
)

func TestConfigLoading(t *testing.T) {
	t.Run("Show Loaded Test Config", func(t *testing.T) {
		// Load test configuration
		cfg := config.LoadTestWithProjectRoot()

		t.Logf("=== LOADED TEST CONFIG ===")
		t.Logf("DatabaseURL: %s", cfg.DatabaseURL)
		t.Logf("JWTSecret: %s", cfg.JWTSecret)
		t.Logf("Port: %s", cfg.Port)
		t.Logf("Environment: %s", cfg.Environment)
		t.Logf("BaseURL: %s", cfg.BaseURL)
		t.Logf("UploadPath: %s", cfg.UploadPath)
		t.Logf("MaxFileSize: %s", cfg.MaxFileSize)
		t.Logf("========================")
	})
}
