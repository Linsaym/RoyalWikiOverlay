package sqlite

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func Open(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	_, err = db.Exec(`PRAGMA foreign_keys = ON;`)
	if err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	return db, nil
}
