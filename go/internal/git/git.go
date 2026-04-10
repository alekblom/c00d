package git

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/alekblom/c00d/internal/config"
)

// Status returns the current git status
type Status struct {
	Branch    string              `json:"branch"`
	Ahead     string              `json:"ahead,omitempty"`
	Behind    string              `json:"behind,omitempty"`
	Staged    []FileStatus        `json:"staged"`
	Unstaged  []FileStatus        `json:"unstaged"`
	Untracked []string            `json:"untracked"`
}

// FileStatus represents a file's git status
type FileStatus struct {
	Status string `json:"status"`
	File   string `json:"file"`
}

// GetStatus returns the current git status
func GetStatus() (*Status, error) {
	status := &Status{}

	// Get current branch
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = config.C.BasePath
	branchOut, _ := cmd.Output()
	status.Branch = strings.TrimSpace(string(branchOut))

	// Get ahead/behind
	cmd = exec.Command("git", "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
	cmd.Dir = config.C.BasePath
	countOut, err := cmd.Output()
	if err == nil {
		parts := strings.Fields(string(countOut))
		if len(parts) == 2 {
			status.Ahead = parts[0]
			status.Behind = parts[1]
		}
	}

	// Get staged files
	cmd = exec.Command("git", "diff", "--cached", "--name-status")
	cmd.Dir = config.C.BasePath
	stagedOut, _ := cmd.Output()
	status.Staged = parseFileStatus(string(stagedOut))

	// Get unstaged files
	cmd = exec.Command("git", "diff", "--name-status")
	cmd.Dir = config.C.BasePath
	unstagedOut, _ := cmd.Output()
	status.Unstaged = parseFileStatus(string(unstagedOut))

	// Get untracked files
	cmd = exec.Command("git", "ls-files", "--others", "--exclude-standard")
	cmd.Dir = config.C.BasePath
	untrackedOut, _ := cmd.Output()
	status.Untracked = []string{}
	for _, line := range strings.Split(strings.TrimSpace(string(untrackedOut)), "\n") {
		if line != "" {
			status.Untracked = append(status.Untracked, line)
		}
	}

	return status, nil
}

func parseFileStatus(output string) []FileStatus {
	result := []FileStatus{}
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			result = append(result, FileStatus{
				Status: parts[0],
				File:   parts[1],
			})
		}
	}
	return result
}

// RunCommand executes a git command and returns the output
func RunCommand(args ...string) (string, int, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = config.C.BasePath

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	return output.String(), exitCode, err
}

// Stage adds a file to the staging area
func Stage(file string) (string, int, error) {
	return RunCommand("add", "--", file)
}

// Unstage removes a file from the staging area
func Unstage(file string) (string, int, error) {
	return RunCommand("restore", "--staged", "--", file)
}

// StageAll stages all changes
func StageAll() (string, int, error) {
	return RunCommand("add", "-A")
}

// UnstageAll unstages all changes
func UnstageAll() (string, int, error) {
	return RunCommand("restore", "--staged", ".")
}

// Commit creates a new commit
func Commit(message string) (string, int, error) {
	return RunCommand("commit", "-m", message)
}

// Push pushes to the remote
func Push() (string, int, error) {
	return RunCommand("push")
}

// Pull pulls from the remote
func Pull() (string, int, error) {
	return RunCommand("pull")
}

// Diff shows the diff
func Diff(staged bool, file string) (string, int, error) {
	args := []string{"diff"}
	if staged {
		args = append(args, "--staged")
	}
	if file != "" {
		args = append(args, "--", file)
	}
	return RunCommand(args...)
}

// Discard discards changes to a file
func Discard(file string) (string, int, error) {
	return RunCommand("checkout", "--", file)
}
