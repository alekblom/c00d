package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/alekblom/c00d/internal/config"
	"github.com/alekblom/c00d/internal/db"
)

// Auth handles authentication requests
func Auth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "GET" {
		// Check if authenticated
		if config.C.Password == "" {
			json.NewEncoder(w).Encode(map[string]any{"authenticated": true, "required": false})
			return
		}

		cookie, err := r.Cookie("c00d_session")
		if err != nil {
			json.NewEncoder(w).Encode(map[string]any{"authenticated": false, "required": true})
			return
		}

		authenticated := db.ValidateSession(cookie.Value)
		json.NewEncoder(w).Encode(map[string]any{"authenticated": authenticated, "required": true})
		return
	}

	if r.Method == "POST" {
		var req struct {
			Password string `json:"password"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		if req.Password != config.C.Password {
			http.Error(w, `{"error":"invalid password"}`, http.StatusUnauthorized)
			return
		}

		// Create session
		token := make([]byte, 32)
		rand.Read(token)
		sessionID := hex.EncodeToString(token)

		db.CreateSession(sessionID, 24*time.Hour)

		http.SetCookie(w, &http.Cookie{
			Name:     "c00d_session",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   86400,
		})

		json.NewEncoder(w).Encode(map[string]any{"success": true})
		return
	}

	if r.Method == "DELETE" {
		cookie, _ := r.Cookie("c00d_session")
		if cookie != nil {
			db.DeleteSession(cookie.Value)
		}
		http.SetCookie(w, &http.Cookie{
			Name:   "c00d_session",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		json.NewEncoder(w).Encode(map[string]any{"success": true})
		return
	}
}
