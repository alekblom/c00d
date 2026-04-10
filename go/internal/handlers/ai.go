package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/alekblom/c00d/internal/ai"
	"github.com/alekblom/c00d/internal/config"
	"github.com/alekblom/c00d/internal/db"
)

// AI handles AI chat requests
func AI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Message     string `json:"message"`
		ContextCode string `json:"context_code"`
		ContextFile string `json:"context_file"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	// Check rate limit for free tier
	if config.C.AI.LicenseKey == "" && config.C.AI.APIKey == "" && config.C.AI.Provider != "ollama" {
		today := time.Now().Format("2006-01-02")
		var count int
		db.DB.QueryRow("SELECT request_count FROM ai_usage WHERE date = ?", today).Scan(&count)
		if count >= 20 {
			json.NewEncoder(w).Encode(map[string]any{
				"success": false,
				"error":   "Daily free limit reached (20 requests). Add a license key or API key for unlimited access.",
			})
			return
		}
	}

	// Build system prompt
	systemPrompt := `You are an expert programming assistant integrated into c00d IDE.
Help users with coding tasks: explain code, fix bugs, write tests, suggest improvements.
Be concise and provide code examples when helpful.`

	if req.ContextCode != "" {
		systemPrompt += "\n\nUser is currently viewing this code"
		if req.ContextFile != "" {
			systemPrompt += " from file: " + req.ContextFile
		}
		systemPrompt += ":\n```\n" + req.ContextCode + "\n```"
	}

	// Save user message
	db.DB.Exec("INSERT INTO ai_history (role, content, file_context) VALUES (?, ?, ?)",
		"user", req.Message, req.ContextFile)

	// Get recent history
	messages := []map[string]string{}
	rows, _ := db.DB.Query(`SELECT role, content FROM ai_history ORDER BY id DESC LIMIT 10`)
	defer rows.Close()
	for rows.Next() {
		var role, content string
		rows.Scan(&role, &content)
		messages = append([]map[string]string{{"role": role, "content": content}}, messages...)
	}

	// Call AI provider
	var result map[string]any
	switch config.C.AI.Provider {
	case "ollama":
		result = ai.CallOllama(systemPrompt, messages)
	case "anthropic":
		result = ai.CallAnthropic(systemPrompt, messages)
	case "openai":
		result = ai.CallOpenAI(systemPrompt, messages)
	default:
		result = ai.CallC00d(systemPrompt, messages)
	}

	if result["success"] == true {
		// Save response
		db.DB.Exec("INSERT INTO ai_history (role, content, file_context, tokens) VALUES (?, ?, ?, ?)",
			"assistant", result["content"], req.ContextFile, result["tokens"])

		// Update usage
		today := time.Now().Format("2006-01-02")
		db.DB.Exec(`INSERT INTO ai_usage (date, request_count, token_count) VALUES (?, 1, ?)
			ON CONFLICT(date) DO UPDATE SET request_count = request_count + 1, token_count = token_count + ?`,
			today, result["tokens"], result["tokens"])
	}

	json.NewEncoder(w).Encode(result)
}
