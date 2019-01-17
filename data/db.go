package data

import (
	"database/sql"
	"strings"

	// Load Go Postgres driver.
	_ "github.com/lib/pq"
	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/dialects/postgresql"
)

// DBConfig is a DB configuration.
type DBConfig struct {
	Conn     map[string]string
	MaxOpen  int
	MaxIddle int
}

// ConnStr composes a data connection string.
func (c DBConfig) ConnStr() string {
	comps := []string{}
	for k, v := range c.Conn {
		comps = append(comps, k+"="+v)
	}
	return strings.Join(comps, " ")
}

// NewDBConfig creates a default DB configuration.
func NewDBConfig() *DBConfig {
	return &DBConfig{
		Conn: map[string]string{
			"dbname":  "dappctrl",
			"sslmode": "disable",
		},
	}
}

func newReform(conn *sql.DB) *reform.DB {
	dummy := func(format string, args ...interface{}) {}

	return reform.NewDB(conn,
		postgresql.Dialect, reform.NewPrintfLogger(dummy))
}

func dbConnect(connStr string) (*sql.DB, error) {
	conn, err := sql.Open("postgres", connStr)
	if err == nil {
		err = conn.Ping()
	}
	return conn, err
}

// NewDBFromConnStr connects to db and returns db instance.
func NewDBFromConnStr(connStr string) (*reform.DB, error) {
	conn, err := dbConnect(connStr)
	if err != nil {
		return nil, err
	}

	return newReform(conn), nil
}

// NewDB creates a new data connection handle.
func NewDB(conf *DBConfig) (*reform.DB, error) {
	conn, err := dbConnect(conf.ConnStr())
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(conf.MaxOpen)
	conn.SetMaxIdleConns(conf.MaxIddle)
	return newReform(conn), nil
}

// CloseDB closes database connection.
func CloseDB(db *reform.DB) {
	db.DBInterface().(*sql.DB).Close()
}
