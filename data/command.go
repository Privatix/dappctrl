package data

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/privatix/dappctrl/data/migration"
	"github.com/privatix/dappctrl/statik"
	reform "gopkg.in/reform.v1"
)

// ExecuteCommand executes commands to manage database
// db-migrate - command for execute migration scripts
// db-init-data - command for initialize database by default values.
func ExecuteCommand(args []string) error {
	if len(args) == 0 {
		return nil
	}

	switch args[0] {
	case "db-migrate":
		connStr := getConnectionString(args)
		if err := migration.Update(connStr); err != nil {
			panic(fmt.Sprintf("failed to run migration %s", err))
		}
		os.Exit(1)
	case "db-init-data":
		connStr := getConnectionString(args)
		if err := initData(connStr); err != nil {
			panic(fmt.Sprintf("failed to init database %s", err))
		}
		os.Exit(1)
	}
	return nil
}

func getConnectionString(args []string) string {
	connStr := flag.String("conn", "", "Database connection string")

	flag.CommandLine.Parse(args[1:])

	if *connStr == "" {
		panic(errors.New("connection string is not detected"))
	}
	return *connStr
}

func initData(connStr string) error {
	db, err := NewDBFromConnStr(connStr)
	if err != nil {
		return err
	}
	defer CloseDB(db)

	file, err := statik.ReadFile("/scripts/prod_data.sql")
	if err != nil {
		return err
	}

	statements := strings.Split(string(file), ";")
	return db.InTransaction(func(tx *reform.TX) error {
		for _, query := range statements {
			if strings.HasSuffix(strings.ToLower(query), "transaction") {
				continue
			}
			if _, err = tx.Exec(query); err != nil {
				return err
			}
		}
		return nil
	})
}
