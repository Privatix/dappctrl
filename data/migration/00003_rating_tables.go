package migration

import (
	"database/sql"

	"github.com/pressly/goose"
	"github.com/privatix/dappctrl/statik"
)

func init() {
	goose.AddMigration(Up00003, Down00003)
}

// Up00003 creates rating tables.
func Up00003(tx *sql.Tx) error {
	query, err := statik.ReadFile("/scripts/migration/00003_rating_tables_up.sql")
	if err != nil {
		return err
	}
	return exec(string(query), tx)
}

// Down00003 destroys rating tables.
func Down00003(tx *sql.Tx) error {
	query, err := statik.ReadFile("/scripts/migration/00003_rating_tables_down.sql")
	if err != nil {
		return err
	}
	return exec(string(query), tx)
}
