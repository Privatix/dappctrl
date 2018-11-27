package data

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
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
// db-create - command to create database
// db-migrate - command to execute migration scripts
// db-init-data - command to initialize database by default values
// db-version - command to print the version of the database schema.
func ExecuteCommand(args []string) error {
	if len(args) == 0 {
		return nil
	}

	switch args[0] {
	case "db-create":
		f := readFlags(args)
		if err := createDatabase(f.connection); err != nil {
			panic("failed to create database: " + err.Error())
		}
		os.Exit(0)
	case "db-migrate":
		f := readFlags(args)
		err := migration.Migrate(f.connection, f.version)
		if err != nil {
			panic("failed to run migration: " + err.Error())
		}
		os.Exit(0)
	case "db-init-data":
		f := readFlags(args)
		if err := initData(f.connection); err != nil {
			panic("failed to init database: " + err.Error())
		}
		os.Exit(0)
	case "db-version":
		f := readFlags(args)
		version, err := migration.Version(f.connection)
		if err != nil {
			msg := "failed to print database schema version: "
			panic(msg + err.Error())
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

func createDatabase(connStr string) error {
	db, err := NewDBFromConnStr(connStr)
	if err != nil {
		return err
	}
	defer CloseDB(db)

	file, err := statik.ReadFile("/scripts/create_db.sql")
	if err != nil {
		return err
	}

	s := string(file)
	re := regexp.MustCompile(`(?m)\\connect \w*`)
	separator := re.FindString(s)

	i := strings.LastIndex(s, separator)
	createStatements := s[:i]

	for _, query := range strings.Split(createStatements, ";") {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}

	// switch to dappctrl database and configurate
	re = regexp.MustCompile(`(?m)dbname=\w*`)
	connStr = re.ReplaceAllString(connStr, `dbname=`+separator[8:])
	conn, err := NewDBFromConnStr(connStr)
	if err != nil {
		return err
	}
	defer CloseDB(conn)

	configStatements := s[i+len(separator):]
	for _, query := range strings.Split(configStatements, ";") {
		if _, err := conn.Exec(query); err != nil {
			return err
		}
	}

	return nil
}
