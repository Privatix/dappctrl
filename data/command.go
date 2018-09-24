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

type cmdFlag int

// cmdFlags enum
const (
	ConnectionString cmdFlag = iota
	Version
)

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
		m := readFlags(args)
		connStr := m[ConnectionString]
		version, _ := strconv.ParseInt(m[Version], 10, 64)
		if err := migration.Migrate(connStr, version); err != nil {
			panic(fmt.Sprintf("failed to run migration %s", err))
		}
		os.Exit(0)
	case "db-init-data":
		connStr := readFlags(args)[ConnectionString]
		if err := initData(connStr); err != nil {
			panic(fmt.Sprintf("failed to init database %s", err))
		}
		os.Exit(0)
	case "db-version":
		connStr := readFlags(args)[ConnectionString]
		version, err := migration.Version(connStr)
		if err != nil {
			msg := "failed to print database schema version"
			panic(fmt.Sprintf("%s %s", msg, err))
		}
		fmt.Println("database schema version:", version)
		os.Exit(0)
	}
	return nil
}

func readFlags(args []string) map[cmdFlag]string {
	connStr := flag.String("conn", "", "Database connection string")
	version := flag.Int("version", 0, "Migrate to version")

	flag.CommandLine.Parse(args[1:])

	if *connStr == "" {
		panic(errors.New("connection string is not detected"))
	}

	m := make(map[cmdFlag]string)

	m[ConnectionString] = *connStr
	m[Version] = strconv.Itoa(*version)

	return m
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
