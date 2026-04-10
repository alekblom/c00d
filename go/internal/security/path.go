package security

import (
	"path/filepath"
	"strings"

	"github.com/alekblom/c00d/internal/config"
)

// ValidatePath checks if a path is within the allowed base path
func ValidatePath(path string) (string, bool) {
	fullPath := filepath.Join(config.C.BasePath, path)
	if !strings.HasPrefix(fullPath, config.C.BasePath) {
		return "", false
	}
	return fullPath, true
}

// ValidateFullPath checks if an absolute path is within the allowed base path
func ValidateFullPath(fullPath string) bool {
	return strings.HasPrefix(fullPath, config.C.BasePath)
}
