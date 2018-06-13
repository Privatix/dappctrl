package main

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"log"
	"math/big"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
	"github.com/sethvargo/go-password/password"
)

const (
	defaultVPNServiceID = "4b26dc82-ffb6-4ff1-99d8-f0eaac0b0532"

	jsonIdent = "    "
)

func main() {
	connStr := flag.String("connstr",
		"user=postgres dbname=dappctrl sslmode=disable",
		"PostgreSQL connection string")
	dappvpnconftpl := flag.String("dappvpnconftpl",
		"dappvpn.config.json", "Dappvpn configuration template JSON")
	dappvpnconf := flag.String("dappvpnconf",
		"dappvpn.config.json", "Dappvpn configuration file to create")
	flag.Parse()

	logger, err := util.NewLogger(util.NewLogConfig())
	if err != nil {
		log.Fatalf("failed to create logger: %s", err)
	}

	db, err := data.NewDBFromConnStr(*connStr, logger)
	if err != nil {
		logger.Fatal("failed to open db connection: %s", err)
	}
	defer data.CloseDB(db)

	id, pass := customiseProduct(logger, db, defaultVPNServiceID)
	createDappvpnConfig(logger, id, pass, *dappvpnconftpl, *dappvpnconf)
}

func randPass() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(10))
	pass, _ := password.Generate(12, int(n.Int64()), 0, false, false)
	return pass
}

func customiseProduct(logger *util.Logger,
	db *reform.DB, oldID string) (string, string) {
	prod := new(data.Product)
	if err := db.FindByPrimaryKeyTo(
		prod, oldID); err != nil {
		logger.Fatal("failed to select"+
			" Vpn Service product: %v", err)
	}

	prod.ID = util.NewUUID()

	salt, err := rand.Int(rand.Reader, big.NewInt(9*1e18))
	if err != nil {
		logger.Fatal("failed to generate salt: %v", err)
	}

	pass := randPass()

	passwordHash, err := data.HashPassword(pass, string(salt.Uint64()))
	if err != nil {
		logger.Fatal("failed to generate password hash: %v", err)
	}

	prod.Password = passwordHash
	prod.Salt = salt.Uint64()

	tx, err := db.Begin()
	if err != nil {
		logger.Fatal("failed to begin transaction: %s", err)
	}
	defer tx.Rollback()

	// update product
	if _, err := tx.Exec(`
			UPDATE products
			   SET id = $1, salt = $2, password = $3
			 WHERE id = $4;`,
		prod.ID, prod.Salt, prod.Password, oldID); err != nil {
		logger.Fatal("failed to update"+
			" Vpn Service product ID: %v", err)
	}

	if err := tx.Commit(); err != nil {
		logger.Fatal("failed to commit transaction: %s", err)
	}

	return prod.ID, pass
}

func createDappvpnConfig(logger *util.Logger,
	username, password, dappvpnconftpl, dappvpnconf string) {
	var conf map[string]interface{}
	if err := json.Unmarshal([]byte(dappvpnconftpl), &conf); err != nil {
		logger.Fatal("failed to parse dappvpn config template: %s", err)
	}

	srv, ok := conf["Server"]
	if !ok {
		logger.Fatal("no server section in dappvpn config template")
	}

	srv.(map[string]interface{})["Username"] = username
	srv.(map[string]interface{})["Password"] = password

	if err := util.WriteJSONFile(
		dappvpnconf, "", jsonIdent, &conf); err != nil {
		logger.Fatal("failed to write dappvpn config")
	}
}
