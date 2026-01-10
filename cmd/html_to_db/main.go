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

	// Извлечение данных из HTML файлов и сохранение в БД
	log.Println("Extracting data from HTML files...")
	// TODO: Implement HTML parsing and data extraction
	log.Println("Data import completed")
}
