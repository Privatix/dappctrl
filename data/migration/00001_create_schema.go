package migration

import (
	"database/sql"

	"github.com/pressly/goose"
	"github.com/privatix/dappctrl/statik"
)

func init() {
	goose.AddMigration(Up00001, Down00001)
}

// Up00001 will be executed as part of a forward migration.
func Up00001(tx *sql.Tx) error {
	query, err := statik.ReadFile("/scripts/migration/00001_schema_up.sql")
	if err != nil {
		return err
	}
	return exec(string(query), tx)
}

// Down00001 will be executed as part of a rollback.
func Down00001(tx *sql.Tx) error {
	query, err := statik.ReadFile("/scripts/migration/00001_schema_down.sql")
	if err != nil {
		return err
	}
	return exec(string(query), tx)
}
