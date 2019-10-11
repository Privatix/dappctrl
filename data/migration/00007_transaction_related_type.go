package migration

import (
	"database/sql"

	"github.com/pressly/goose"
	"github.com/privatix/dappctrl/statik"
)

func init() {
	goose.AddMigration(Up00007, Down00007)
}

// Up00007 deletes old unused setting.
func Up00007(tx *sql.Tx) error {
	query, err := statik.ReadFile("/scripts/migration/00007_transaction_related_type_up.sql")
	if err != nil {
		return err
	}
	return exec(string(query), tx)
}

// Down00007 does nothing.
func Down00007(tx *sql.Tx) error {
	return nil
}
