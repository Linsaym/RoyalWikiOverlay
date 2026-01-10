package main

import (
	"log"
	"os"
	"path/filepath"

	"RoyalWikiOverlay/infrastructure/sqlite"
)

func main() {
	dbPath := filepath.Join("data", "app.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		log.Fatalf("failed to create data dir: %v", err)
	}

	db, err := sqlite.Open(dbPath)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := sqlite.ApplyMigrations(db); err != nil {
		log.Fatalf("failed to apply migrations: %v", err)
	}

	log.Printf("db initialized at %s", dbPath)
}
