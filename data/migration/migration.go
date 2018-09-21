package migration

import (
	"database/sql"

	"github.com/pressly/goose"
)

func exec(query string, tx *sql.Tx) error {
	_, err := tx.Exec(query)
	return err
}

const minVersion = int64(0)
const maxVersion = int64((1 << 63) - 1)

// Update executes migration scripts.
func Update(connStr string) error {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	return executeMigrationScripts(db)
}

func executeMigrationScripts(db *sql.DB) error {
	current, err := goose.GetDBVersion(db)
	if err != nil {
		return err
	}

	migrations, err := goose.CollectMigrations(".", minVersion, maxVersion)
	if err != nil {
		return err
	}

	last, err := migrations.Last()
	if err != nil {
		return err
	}

	if last.Version > current {
		if err := goose.Up(db, "."); err != nil {
			return err
		}
	}
	return nil
}
