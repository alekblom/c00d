package ai

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/alekblom/c00d/internal/config"
)

// CallOpenAI calls the OpenAI API
func CallOpenAI(system string, messages []map[string]string) map[string]any {
	if config.C.AI.APIKey == "" {
		return map[string]any{"success": false, "error": "OpenAI API key not configured"}
	}

	// Prepend system message
	allMessages := append([]map[string]string{{"role": "system", "content": system}}, messages...)

	payload, _ := json.Marshal(map[string]any{
		"model":      config.C.AI.Model,
		"messages":   allMessages,
		"max_tokens": 4096,
	})

	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.C.AI.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)

	if result["error"] != nil {
		errMap := result["error"].(map[string]any)
		return map[string]any{"success": false, "error": errMap["message"]}
	}

	content := ""
	if choices, ok := result["choices"].([]any); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]any); ok {
			if msg, ok := choice["message"].(map[string]any); ok {
				content = msg["content"].(string)
			}
		}
	}

	tokens := 0
	if usage, ok := result["usage"].(map[string]any); ok {
		if total, ok := usage["total_tokens"].(float64); ok {
			tokens = int(total)
		}
	}

	return map[string]any{
		"success": true,
		"content": content,
		"tokens":  tokens,
	}
}
