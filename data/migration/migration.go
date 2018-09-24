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

// Version returns the version of the database schema.
func Version(connStr string) (int64, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return 0, err
	}
	defer db.Close()
	return goose.GetDBVersion(db)
}

// Migrate executes migration scripts.
func Migrate(connStr string, version int64) error {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	if version == 0 {
		migrations, err := goose.CollectMigrations(".", minVersion, maxVersion)
		if err != nil {
			return err
		}

		last, err := migrations.Last()
		if err != nil {
			return err
		}
		version = last.Version
	}

	return executeMigrationScripts(db, version)
}

func executeMigrationScripts(db *sql.DB, version int64) error {
	current, err := goose.GetDBVersion(db)
	if err != nil {
		return err
	}

	// upgrade database schema
	if version > current {
		if err := goose.UpTo(db, ".", version); err != nil {
			return err
		}
	}

	// downgrade database schema
	if version < current {
		if err := goose.DownTo(db, ".", version); err != nil {
			return err
		}
	}
	return nil
}
