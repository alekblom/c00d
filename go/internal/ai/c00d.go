package ai

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/alekblom/c00d/internal/config"
)

// CallC00d calls the c00d AI API
func CallC00d(system string, messages []map[string]string) map[string]any {
	payload, _ := json.Marshal(map[string]any{
		"license_key": config.C.AI.LicenseKey,
		"system":      system,
		"messages":    messages,
		"model":       config.C.AI.Model,
	})

	resp, err := http.Post("https://c00d.com/api/ai/chat", "application/json", bytes.NewReader(payload))
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)

	if result["error"] != nil {
		return map[string]any{"success": false, "error": result["error"]}
	}

	return map[string]any{
		"success": true,
		"content": result["content"],
		"tokens":  result["tokens"],
	}
}
