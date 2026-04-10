package ai

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/alekblom/c00d/internal/config"
)

// CallOllama calls the Ollama API
func CallOllama(system string, messages []map[string]string) map[string]any {
	// Prepend system message
	allMessages := append([]map[string]string{{"role": "system", "content": system}}, messages...)

	payload, _ := json.Marshal(map[string]any{
		"model":    config.C.AI.Model,
		"messages": allMessages,
		"stream":   false,
	})

	url := strings.TrimRight(config.C.AI.OllamaURL, "/") + "/api/chat"
	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return map[string]any{"success": false, "error": "Cannot connect to Ollama: " + err.Error()}
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)

	if result["message"] == nil {
		return map[string]any{"success": false, "error": "Invalid response from Ollama"}
	}

	msg := result["message"].(map[string]any)
	return map[string]any{
		"success": true,
		"content": msg["content"],
		"tokens":  0,
	}
}
