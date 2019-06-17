package migration

import (
	"database/sql"

	"github.com/pressly/goose"
	"github.com/privatix/dappctrl/statik"
)

func init() {
	goose.AddMigration(Up00002, Down00002)
}

// Up00002 removes redundant system.version.db settings from DB.
func Up00002(tx *sql.Tx) error {
	query, err := statik.ReadFile("/scripts/migration/00002_remove_version_db_setting.sql")
	if err != nil {
		return err
	}
	return exec(string(query), tx)
}

// Down00002 does nothing. No rollback needed.
func Down00002(tx *sql.Tx) error {
	return nil
}
