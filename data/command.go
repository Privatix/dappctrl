package data

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/privatix/dappctrl/data/migration"
	"github.com/privatix/dappctrl/statik"
	reform "gopkg.in/reform.v1"
)

type cmdFlag struct {
	connection string
	version    int64
}

// ExecuteCommand executes commands to manage database
// db-migrate - command to execute migration scripts
// db-init-data - command to initialize database by default values
// db-version - command to print the version of the database schema.
func ExecuteCommand(args []string) error {
	if len(args) == 0 {
		return nil
	}

	switch args[0] {
	case "db-migrate":
		f := readFlags(args)
		err := migration.Migrate(f.connection, f.version)
		if err != nil {
			panic(fmt.Sprintf("failed to run migration %s", err))
		}
		os.Exit(0)
	case "db-init-data":
		f := readFlags(args)
		if err := initData(f.connection); err != nil {
			panic(fmt.Sprintf("failed to init database %s", err))
		}
		os.Exit(0)
	case "db-version":
		f := readFlags(args)
		version, err := migration.Version(f.connection)
		if err != nil {
			msg := "failed to print database schema version"
			panic(fmt.Sprintf("%s %s", msg, err))
		}
		fmt.Println("database schema version:", version)
		os.Exit(0)
	}
	return nil
}

func readFlags(args []string) *cmdFlag {
	connStr := flag.String("conn", "", "Database connection string")
	version := flag.Int("version", 0, "Migrate to version")

	flag.CommandLine.Parse(args[1:])

	if *connStr == "" {
		panic(errors.New("connection string is not detected"))
	}

	return &cmdFlag{
		connection: *connStr,
		version:    int64(*version),
	}
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
