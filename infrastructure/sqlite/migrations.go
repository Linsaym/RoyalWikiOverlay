package sqlite

import (
	"database/sql"
	"embed"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(db *sql.DB) error {
	files, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return err
	}

	for _, file := range files {
		content, err := migrationsFS.ReadFile("migrations/" + file.Name())
		if err != nil {
			return err
		}

		if _, err := db.Exec(string(content)); err != nil {
			return err
		}
	}

	return nil
}
