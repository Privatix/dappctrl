package migration

import (
	"database/sql"

	"github.com/pressly/goose"
	"github.com/privatix/dappctrl/statik"
)

func init() {
	goose.AddMigration(Up00005, Down00005)
}

// Up00005 deletes old unused setting.
func Up00005(tx *sql.Tx) error {
	query, err := statik.ReadFile("/scripts/migration/00005_old_setting_delete_up.sql")
	if err != nil {
		return err
	}
	return exec(string(query), tx)
}

// Down00005 does nothing.
func Down00005(tx *sql.Tx) error {
	return nil
}
