package main

import (
	"log"
	"os"

	"RoyalWikiOverlay/infrastructure/sqlite"
)

func main() {
	log.Println("✅ Процесс может занять до минуты, дождитесь конца")
	const dbPath = "data/royalwiki.db"
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatalf("create data dir: %v", err)
	}

	db, err := sqlite.Open(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := sqlite.RunMigrations(db); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	log.Println("✅ SQLite migrations applied successfully")
}
