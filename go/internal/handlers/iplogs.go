package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/alekblom/c00d/internal/config"
	"github.com/alekblom/c00d/internal/db"
)

// IPLogs handles requests to view IP logs
func IPLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "GET" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Check if logging is enabled
	if !config.ShouldLogIPs() {
		json.NewEncoder(w).Encode(map[string]any{
			"enabled": false,
			"message": "IP logging is disabled in config",
		})
		return
	}

	// Get query params
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	viewType := r.URL.Query().Get("view")

	if viewType == "unique" {
		// Return unique IPs with counts
		ips, err := db.GetUniqueIPs(limit)
		if err != nil {
			http.Error(w, `{"error":"failed to fetch logs"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"enabled": true,
			"view":    "unique",
			"ips":     ips,
		})
		return
	}

	// Return detailed logs
	logs, err := db.GetIPLogs(limit)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch logs"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"enabled": true,
		"view":    "detailed",
		"logs":    logs,
		"count":   len(logs),
	})
}
