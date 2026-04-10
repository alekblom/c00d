package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/alekblom/c00d/internal/config"
)

// Config returns editor and AI configuration
func Config(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]any{
		"editor": map[string]any{
			"theme":     config.C.Editor.Theme,
			"font_size": config.C.Editor.FontSize,
			"tab_size":  config.C.Editor.TabSize,
		},
		"ai": map[string]any{
			"provider": config.C.AI.Provider,
			"has_key":  config.C.AI.APIKey != "" || config.C.AI.LicenseKey != "",
		},
	})
}
