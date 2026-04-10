package ai

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/alekblom/c00d/internal/config"
)

// CallAnthropic calls the Anthropic API
func CallAnthropic(system string, messages []map[string]string) map[string]any {
	if config.C.AI.APIKey == "" {
		return map[string]any{"success": false, "error": "Anthropic API key not configured"}
	}

	payload, _ := json.Marshal(map[string]any{
		"model":      config.C.AI.Model,
		"max_tokens": 4096,
		"system":     system,
		"messages":   messages,
	})

	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", config.C.AI.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

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
	if contentArr, ok := result["content"].([]any); ok && len(contentArr) > 0 {
		if textBlock, ok := contentArr[0].(map[string]any); ok {
			content = textBlock["text"].(string)
		}
	}

	tokens := 0
	if usage, ok := result["usage"].(map[string]any); ok {
		if in, ok := usage["input_tokens"].(float64); ok {
			tokens += int(in)
		}
		if out, ok := usage["output_tokens"].(float64); ok {
			tokens += int(out)
		}
	}

	return map[string]any{
		"success": true,
		"content": content,
		"tokens":  tokens,
	}
}
