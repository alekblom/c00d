package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"

	"github.com/alekblom/c00d/internal/config"
)

// Git handles git operations
func Git(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Action  string `json:"action"`
		File    string `json:"file"`
		Message string `json:"message"`
		Staged  bool   `json:"staged"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	var cmd *exec.Cmd
	var output bytes.Buffer

	switch req.Action {
	case "status":
		// Get comprehensive status
		result := map[string]any{}

		// Get current branch
		cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		cmd.Dir = config.C.BasePath
		branchOut, _ := cmd.Output()
		result["branch"] = strings.TrimSpace(string(branchOut))

		// Get ahead/behind
		cmd = exec.Command("git", "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
		cmd.Dir = config.C.BasePath
		countOut, err := cmd.Output()
		if err == nil {
			parts := strings.Fields(string(countOut))
			if len(parts) == 2 {
				result["ahead"] = parts[0]
				result["behind"] = parts[1]
			}
		}

		// Get staged files
		cmd = exec.Command("git", "diff", "--cached", "--name-status")
		cmd.Dir = config.C.BasePath
		stagedOut, _ := cmd.Output()
		result["staged"] = parseGitStatus(string(stagedOut))

		// Get unstaged files
		cmd = exec.Command("git", "diff", "--name-status")
		cmd.Dir = config.C.BasePath
		unstagedOut, _ := cmd.Output()
		result["unstaged"] = parseGitStatus(string(unstagedOut))

		// Get untracked files
		cmd = exec.Command("git", "ls-files", "--others", "--exclude-standard")
		cmd.Dir = config.C.BasePath
		untrackedOut, _ := cmd.Output()
		untracked := []string{}
		for _, line := range strings.Split(strings.TrimSpace(string(untrackedOut)), "\n") {
			if line != "" {
				untracked = append(untracked, line)
			}
		}
		result["untracked"] = untracked

		json.NewEncoder(w).Encode(result)
		return

	case "stage":
		if req.File == "" {
			http.Error(w, `{"error":"file is required"}`, http.StatusBadRequest)
			return
		}
		cmd = exec.Command("git", "add", "--", req.File)

	case "unstage":
		if req.File == "" {
			http.Error(w, `{"error":"file is required"}`, http.StatusBadRequest)
			return
		}
		cmd = exec.Command("git", "restore", "--staged", "--", req.File)

	case "stage_all":
		cmd = exec.Command("git", "add", "-A")

	case "unstage_all":
		cmd = exec.Command("git", "restore", "--staged", ".")

	case "commit":
		if req.Message == "" {
			http.Error(w, `{"error":"message is required"}`, http.StatusBadRequest)
			return
		}
		cmd = exec.Command("git", "commit", "-m", req.Message)

	case "push":
		cmd = exec.Command("git", "push")

	case "pull":
		cmd = exec.Command("git", "pull")

	case "diff":
		args := []string{"diff"}
		if req.Staged {
			args = append(args, "--staged")
		}
		if req.File != "" {
			args = append(args, "--", req.File)
		}
		cmd = exec.Command("git", args...)

	case "discard":
		if req.File == "" {
			http.Error(w, `{"error":"file is required"}`, http.StatusBadRequest)
			return
		}
		cmd = exec.Command("git", "checkout", "--", req.File)

	default:
		http.Error(w, `{"error":"invalid action"}`, http.StatusBadRequest)
		return
	}

	cmd.Dir = config.C.BasePath
	cmd.Stdout = &output
	cmd.Stderr = &output

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	json.NewEncoder(w).Encode(map[string]any{
		"success":   exitCode == 0,
		"output":    output.String(),
		"exit_code": exitCode,
	})
}

func parseGitStatus(output string) []map[string]string {
	result := []map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			result = append(result, map[string]string{
				"status": parts[0],
				"file":   parts[1],
			})
		}
	}
	return result
}
