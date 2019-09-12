package migration

import (
	"database/sql"

	"github.com/pressly/goose"
	"github.com/privatix/dappctrl/statik"
)

func init() {
	goose.AddMigration(Up00006, Down00006)
}

// Up00006 deletes old unused setting.
func Up00006(tx *sql.Tx) error {
	query, err := statik.ReadFile("/scripts/migration/00006_remove_version_setting_up.sql")
	if err != nil {
		return err
	}
	return exec(string(query), tx)
}

// Down00006 does nothing.
func Down00006(tx *sql.Tx) error {
	return nil
}
