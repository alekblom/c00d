package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/alekblom/c00d/internal/config"
)

// File handles single file operations (read, write, delete, rename)
func File(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := r.URL.Query().Get("path")
	fullPath := filepath.Join(config.C.BasePath, path)

	// Security check
	if !strings.HasPrefix(fullPath, config.C.BasePath) {
		http.Error(w, `{"error":"access denied"}`, http.StatusForbidden)
		return
	}

	switch r.Method {
	case "GET":
		// Read file
		content, err := os.ReadFile(fullPath)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"path":    path,
			"content": string(content),
		})

	case "POST", "PUT":
		// Write file
		var req struct {
			Content string `json:"content"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		// Ensure directory exists
		os.MkdirAll(filepath.Dir(fullPath), 0755)

		err := os.WriteFile(fullPath, []byte(req.Content), 0644)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"success": true})

	case "DELETE":
		// Delete file
		err := os.RemoveAll(fullPath)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"success": true})

	case "PATCH":
		// Rename file
		var req struct {
			NewPath string `json:"new_path"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		if req.NewPath == "" {
			http.Error(w, `{"error":"new_path is required"}`, http.StatusBadRequest)
			return
		}

		newFullPath := filepath.Join(config.C.BasePath, req.NewPath)
		// Security check for new path
		if !strings.HasPrefix(newFullPath, config.C.BasePath) {
			http.Error(w, `{"error":"access denied"}`, http.StatusForbidden)
			return
		}

		// Ensure parent directory exists
		os.MkdirAll(filepath.Dir(newFullPath), 0755)

		err := os.Rename(fullPath, newFullPath)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"success": true, "new_path": req.NewPath})
	}
}
