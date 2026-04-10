package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/alekblom/c00d/internal/auth"
	"github.com/alekblom/c00d/internal/config"
	"github.com/alekblom/c00d/internal/db"
	"github.com/alekblom/c00d/internal/handlers"
)

//go:embed frontend/*
var frontendFS embed.FS

func main() {
	// Parse flags
	configFile := flag.String("config", "config.yaml", "Path to config file")
	port := flag.Int("port", 0, "Port to listen on (overrides config)")
	basePath := flag.String("path", "", "Base path for file browsing (overrides config)")
	flag.Parse()

	// Load config
	config.Load(*configFile)

	// Override with flags
	if *port > 0 {
		config.C.Port = *port
	}
	if *basePath != "" {
		config.C.BasePath = *basePath
	}

	// Apply defaults
	config.ApplyDefaults()

	// Ensure data directory exists
	os.MkdirAll(config.C.DataDir, 0755)

	// Initialize database
	db.Init()

	// Start session cleanup routine
	auth.StartCleanupRoutine()

	// Set frontend filesystem for handler
	handlers.FrontendFS = frontendFS

	// Setup routes
	mux := http.NewServeMux()

	// Helper to wrap handlers with logging + auth
	withAuth := func(h http.HandlerFunc) http.HandlerFunc {
		return auth.WithLogging(auth.Middleware(h))
	}

	// API routes (with logging and auth)
	mux.HandleFunc("/api/files", withAuth(handlers.Files))
	mux.HandleFunc("/api/file", withAuth(handlers.File))
	mux.HandleFunc("/api/terminal", withAuth(handlers.Terminal))
	mux.HandleFunc("/api/ai", withAuth(handlers.AI))
	mux.HandleFunc("/api/config", withAuth(handlers.Config))
	mux.HandleFunc("/api/git", withAuth(handlers.Git))
	mux.HandleFunc("/api/search", withAuth(handlers.Search))
	mux.HandleFunc("/api/iplogs", withAuth(handlers.IPLogs))
	mux.HandleFunc("/api/auth", auth.WithLogging(handlers.Auth))

	// Frontend (with logging only, no auth required)
	mux.HandleFunc("/", auth.WithLogging(handlers.Frontend))

	// Start server
	addr := fmt.Sprintf(":%d", config.C.Port)
	fmt.Printf("\n  c00d is running!\n\n")
	fmt.Printf("  Local:   http://localhost%s\n", addr)
	if ip := getOutboundIP(); ip != "" {
		fmt.Printf("  Network: http://%s%s\n", ip, addr)
	}
	fmt.Printf("  Path:    %s\n\n", config.C.BasePath)

	if config.C.Password != "" {
		fmt.Printf("  Password protection: enabled\n")
	} else {
		fmt.Printf("  Warning: No password set. Add 'password' to config.yaml\n")
	}

	if config.ShouldLogIPs() {
		fmt.Printf("  IP logging: enabled\n\n")
	} else {
		fmt.Printf("  IP logging: disabled\n\n")
	}

	log.Fatal(http.ListenAndServe(addr, mux))
}

func getOutboundIP() string {
	cmd := exec.Command("hostname", "-I")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	for _, ip := range strings.Fields(string(output)) {
		if strings.Contains(ip, ".") && !strings.HasPrefix(ip, "127.") {
			return ip
		}
	}
	return ""
}
