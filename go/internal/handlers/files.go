package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/alekblom/c00d/internal/config"
)

// Files handles directory listing requests
func Files(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	}

	fullPath := filepath.Join(config.C.BasePath, path)

	// Security: ensure path is within base
	if !strings.HasPrefix(fullPath, config.C.BasePath) {
		http.Error(w, `{"error":"access denied"}`, http.StatusForbidden)
		return
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	type FileInfo struct {
		Name    string `json:"name"`
		Path    string `json:"path"`
		IsDir   bool   `json:"is_dir"`
		Size    int64  `json:"size"`
		ModTime string `json:"mod_time"`
	}

	files := make([]FileInfo, 0)
	for _, entry := range entries {
		// Skip hidden files starting with .
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		// Skip common large/irrelevant directories
		if entry.IsDir() && (entry.Name() == "node_modules" || entry.Name() == "vendor" || entry.Name() == ".git") {
			continue
		}

		info, _ := entry.Info()
		relPath := filepath.Join(path, entry.Name())

		files = append(files, FileInfo{
			Name:    entry.Name(),
			Path:    relPath,
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().Format(time.RFC3339),
		})
	}

	// Sort: directories first, then by name
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return files[i].Name < files[j].Name
	})

	json.NewEncoder(w).Encode(map[string]any{
		"path":  path,
		"files": files,
	})
}
