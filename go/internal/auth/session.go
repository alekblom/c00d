package auth

import (
	"time"

	"github.com/alekblom/c00d/internal/db"
)

// StartCleanupRoutine starts a background goroutine to clean up expired sessions
func StartCleanupRoutine() {
	go func() {
		for {
			time.Sleep(time.Hour)
			db.CleanupExpiredSessions()
		}
	}()
}
