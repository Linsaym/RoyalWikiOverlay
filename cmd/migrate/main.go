package main

import (
	"log"
	"os"

	"RoyalWikiOverlay/infrastructure/sqlite"
)

func main() {
	dbPath := "data/app.db"

	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatal(err)
	}

	db, err := sqlite.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := sqlite.RunMigrations(db); err != nil {
		log.Fatal(err)
	}

	log.Println("Database migrations completed successfully")
}
