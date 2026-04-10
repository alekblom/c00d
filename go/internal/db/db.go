package db

import (
	"database/sql"
	"log"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/alekblom/c00d/internal/config"
)

// DB is the global database connection
var DB *sql.DB

// Init initializes the database connection and creates tables
func Init() {
	var err error
	dbPath := filepath.Join(config.C.DataDir, "c00d.db")
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	RunMigrations()
}
