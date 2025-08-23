package router

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

// getAbsolutePath converts a relative path to absolute path from project root
func getAbsolutePath(relativePath string) string {
	if projectRoot := getProjectRoot(); projectRoot != "" {
		return filepath.Join(projectRoot, relativePath)
	}
	return relativePath
}
