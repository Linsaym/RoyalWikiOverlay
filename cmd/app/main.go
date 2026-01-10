package main

import (
	"log"
	"os"
)

func main() {
	//dbPath := "data/app.db"

	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatal(err)
	}

	// Запуск приложения
	log.Println("Application started")
}
