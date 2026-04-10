package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alekblom/c00d/internal/config"
	"github.com/alekblom/c00d/internal/db"
)

// Terminal handles command execution requests
func Terminal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Command string `json:"command"`
		Cwd     string `json:"cwd"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Command == "" {
		http.Error(w, `{"error":"no command"}`, http.StatusBadRequest)
		return
	}

	// Determine working directory
	cwd := config.C.BasePath
	if req.Cwd != "" {
		cwd = filepath.Join(config.C.BasePath, req.Cwd)
		if !strings.HasPrefix(cwd, config.C.BasePath) {
			cwd = config.C.BasePath
		}
	}

	// Save to history
	db.DB.Exec("INSERT INTO command_history (command) VALUES (?)", req.Command)

	// Execute command
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", req.Command)
	} else {
		cmd = exec.Command("sh", "-c", req.Command)
	}
	cmd.Dir = cwd

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	output := stdout.String()
	if stderr.Len() > 0 {
		if output != "" {
			output += "\n"
		}
		output += stderr.String()
	}

	json.NewEncoder(w).Encode(map[string]any{
		"output":    output,
		"exit_code": exitCode,
	})
}
