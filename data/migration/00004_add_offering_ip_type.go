package migration

import (
	"database/sql"

	"github.com/pressly/goose"
	"github.com/privatix/dappctrl/statik"
)

func init() {
	goose.AddMigration(Up00004, Down00004)
}

// Up00004 adds ip type.
func Up00004(tx *sql.Tx) error {
	query, err := statik.ReadFile("/scripts/migration/00004_add_offering_ip_type_up.sql")
	if err != nil {
		return err
	}
	return exec(string(query), tx)
}

// Down00004 drops ip type.
func Down00004(tx *sql.Tx) error {
	query, err := statik.ReadFile("/scripts/migration/00004_add_offering_ip_type_down.sql")
	if err != nil {
		return err
	}
	return exec(string(query), tx)
}
