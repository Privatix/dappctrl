package main

import (
	"crypto/rand"
	"database/sql"
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
	jsonIdent  = "    "
	appVersion = "0.15.0"
)

func main() {
	connStr := flag.String("connstr",
		"user=postgres dbname=dappctrl sslmode=disable",
		"PostgreSQL connection string")
	dappvpnconftpl := flag.String("dappvpnconftpl",
		"dappvpn.config.json", "Dappvpn configuration template JSON")
	dappvpnconf := flag.String("dappvpnconf",
		"dappvpn.config.json", "Dappvpn configuration file to create")
	template := flag.String("template", "", "Offering template ID")
	agent := flag.Bool("agent", false, "Whether to install agent")
	flag.Parse()

	logger, err := util.NewLogger(util.NewLogConfig())
	if err != nil {
		log.Fatalf("failed to create logger: %s", err)
	}
	defer logger.GracefulStop()

	db, err := data.NewDBFromConnStr(*connStr)
	if err != nil {
		logger.Fatal("failed to open db connection: %s", err)
	}
	defer data.CloseDB(db)

	id, pass := customiseProduct(logger, db, *template, *agent)
	createDappvpnConfig(
		logger, id, pass, *dappvpnconftpl, *dappvpnconf, *agent)
	writeAppVersion(logger, db)
}

func randPass() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(10))
	pass, _ := password.Generate(12, int(n.Int64()), 0, false, false)
	return pass
}

func customiseProduct(logger *util.Logger,
	db *reform.DB, templateID string, agent bool) (string, string) {
	prod := new(data.Product)
	err := db.SelectOneTo(prod,
		"WHERE offer_tpl_id = $1 AND is_server = $2", templateID, agent)
	if err != nil {
		logger.Fatal("failed to find VPN service product: %v", err)
	}

	oldID := prod.ID
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
	username, password, dappvpnconftpl, dappvpnconf string, agent bool) {
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

	if !agent {
		mon, ok := conf["Monitor"]
		if !ok {
			logger.Fatal(
				"no monitor section in dappvpn config template")
		}

		mon.(map[string]interface{})["Addr"] = "localhost:7506"
	}

	if err := util.WriteJSONFile(
		dappvpnconf, "", jsonIdent, &conf); err != nil {
		logger.Fatal("failed to write dappvpn config")
	}
}

func writeAppVersion(logger *util.Logger, db *reform.DB) {
	versionSetting := &data.Setting{}
	err := db.FindOneTo(versionSetting, "key", data.SettingAppVersion)
	if err == sql.ErrNoRows {
		err = db.Insert(&data.Setting{
			Key:         data.SettingAppVersion,
			Value:       appVersion,
			Permissions: data.ReadOnly,
			Name:        "App version",
		})
	} else if err == nil {
		logger.Info("%s before update: %v",
			data.SettingAppVersion, versionSetting.Value)

		versionSetting.Value = appVersion
		err = db.Update(versionSetting)
	}

	if err != nil {
		logger.Fatal("failed to write app version: %v", err)
	}

	logger.Info("%s after update: %v",
		data.SettingAppVersion, appVersion)
}
