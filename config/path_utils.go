package config

import (
	"os"
	"path/filepath"
)

// getProjectRoot finds the project root by looking for go.mod file
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

// LoadWithProjectRoot loads config ensuring we're in the project root directory
func LoadWithProjectRoot(envFiles ...string) *Config {
	// Save current directory
	originalDir, _ := os.Getwd()

	// Change to project root if we can find it
	if projectRoot := getProjectRoot(); projectRoot != "" {
		os.Chdir(projectRoot)
		defer os.Chdir(originalDir) // Restore original directory
	}

	return Load(envFiles...)
}

// LoadTestWithProjectRoot loads test config ensuring we're in the project root directory
func LoadTestWithProjectRoot() *Config {
	return LoadWithProjectRoot(".env.test")
}
