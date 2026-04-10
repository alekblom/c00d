package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alekblom/c00d/internal/config"
)

// Search handles file content search requests
func Search(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Query      string `json:"query"`
		Path       string `json:"path"`
		IsRegex    bool   `json:"is_regex"`
		MaxResults int    `json:"max_results"`
		FileGlob   string `json:"file_glob"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Query == "" {
		http.Error(w, `{"error":"query is required"}`, http.StatusBadRequest)
		return
	}

	if req.MaxResults <= 0 {
		req.MaxResults = 100
	}

	searchPath := config.C.BasePath
	if req.Path != "" {
		searchPath = filepath.Join(config.C.BasePath, req.Path)
		if !strings.HasPrefix(searchPath, config.C.BasePath) {
			http.Error(w, `{"error":"access denied"}`, http.StatusForbidden)
			return
		}
	}

	var pattern *regexp.Regexp
	var err error
	if req.IsRegex {
		pattern, err = regexp.Compile(req.Query)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"invalid regex: %s"}`, err.Error()), http.StatusBadRequest)
			return
		}
	} else {
		// Case-insensitive substring search
		pattern = regexp.MustCompile("(?i)" + regexp.QuoteMeta(req.Query))
	}

	type SearchResult struct {
		File    string `json:"file"`
		Line    int    `json:"line"`
		Content string `json:"content"`
	}
	results := []SearchResult{}

	// Skip directories
	skipDirs := map[string]bool{
		"node_modules": true,
		".git":         true,
		"vendor":       true,
		".c00d":        true,
	}

	filepath.WalkDir(searchPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// Skip hidden files and directories
		if strings.HasPrefix(d.Name(), ".") && d.Name() != "." {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip certain directories
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// Check file glob if specified
		if req.FileGlob != "" {
			matched, _ := filepath.Match(req.FileGlob, d.Name())
			if !matched {
				return nil
			}
		}

		// Skip large files (> 1MB)
		info, err := d.Info()
		if err != nil || info.Size() > 1024*1024 {
			return nil
		}

		// Read and search file
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(config.C.BasePath, path)
		lines := strings.Split(string(content), "\n")

		for i, line := range lines {
			if pattern.MatchString(line) {
				results = append(results, SearchResult{
					File:    relPath,
					Line:    i + 1,
					Content: strings.TrimSpace(line),
				})
				if len(results) >= req.MaxResults {
					return io.EOF
				}
			}
		}

		return nil
	})

	json.NewEncoder(w).Encode(map[string]any{
		"results": results,
		"count":   len(results),
	})
}
