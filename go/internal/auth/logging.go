package auth

import (
	"net/http"

	"github.com/alekblom/c00d/internal/config"
	"github.com/alekblom/c00d/internal/db"
)

// LoggingMiddleware logs client IP addresses if enabled in config
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if config.ShouldLogIPs() {
			db.LogIP(r)
		}
		next(w, r)
	}
}

// WithLogging wraps a handler with IP logging
func WithLogging(next http.HandlerFunc) http.HandlerFunc {
	return LoggingMiddleware(next)
}
