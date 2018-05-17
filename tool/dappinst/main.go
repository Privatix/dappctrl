package main

import (
	"encoding/json"
	"flag"
	"log"
	"strings"

	"github.com/AlekSi/pointer"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

const (
	jsoffer = `{
	    "title": "Person",
	        "type": "object",
	        "properties": {
	            "firstName": {"type": "string"},
	            "lastName": {"type": "string"},
	            "age": {
	                "description": "Age in years",
	                "type": "integer",
	                "minimum": 0
	            }
	    },
	    "required": ["firstName", "lastName"]
	}`

	jsendp = `{
	    "title": "Person",
	        "type": "object",
	        "properties": {
	            "firstName": {"type": "string"},
	            "lastName": {"type": "string"},
	            "age": {
	                "description": "Age in years",
	                "type": "integer",
	                "minimum": 0
	            }
	    },
	    "required": ["firstName", "lastName"]
	}`

	jsconf = `{}`

	hash = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

	salt         = 6012867121110302348
	passwordHash = "7U9gC4AZsSZ9E8NabVkw8nHRlFCJe0o_Yh9qMlIaGAg="
	password     = "secret"

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

	id := insertProduct(logger, db)
	createDappvpnConfig(logger, id, *dappvpnconftpl, *dappvpnconf)
}

func minifyJSON(json string) []byte {
	for _, v := range []string{"\t", " ", "\n"} {
		json = strings.Replace(json, v, "", -1)
	}
	return []byte(json)
}

func insertProduct(logger *util.Logger, db *reform.DB) string {
	tx, err := db.Begin()
	if err != nil {
		logger.Fatal("failed to begin transaction: %s", err)
	}
	defer tx.Rollback()

	offer := data.Template{
		ID:   util.NewUUID(),
		Hash: hash,
		Raw:  minifyJSON(jsoffer),
		Kind: data.TemplateOffer,
	}
	if err := tx.Insert(&offer); err != nil {
		logger.Fatal("failed to insert offer template: %s", err)
	}

	access := data.Template{
		ID:   util.NewUUID(),
		Hash: hash,
		Raw:  minifyJSON(jsendp),
		Kind: data.TemplateAccess,
	}
	if err := tx.Insert(&access); err != nil {
		logger.Fatal("failed to insert access template: %s", err)
	}

	prod := data.Product{
		ID:            util.NewUUID(),
		Name:          "OpenVPN server",
		OfferTplID:    pointer.ToString(offer.ID),
		OfferAccessID: pointer.ToString(access.ID),
		UsageRepType:  data.ProductUsageTotal,
		IsServer:      true,
		Salt:          salt,
		Password:      passwordHash,
		ClientIdent:   data.ClientIdentByChannelID,
		Config:        minifyJSON(jsconf),
	}
	if err := tx.Insert(&prod); err != nil {
		logger.Fatal("failed to insert product: %s", err)
	}

	if err := tx.Commit(); err != nil {
		logger.Fatal("failed to commit transaction: %s", err)
	}

	return prod.ID
}

func createDappvpnConfig(logger *util.Logger,
	username, dappvpnconftpl, dappvpnconf string) {
	var conf map[string]interface{}
	if err := json.Unmarshal([]byte(dappvpnconftpl), &conf); err != nil {
		logger.Fatal("failed to parse dappvpn config template: %s", err)
	}

	srv, ok := conf["Server"]
	if !ok {
		logger.Fatal("no server section in dappvpn config template")
	}

	srv.(map[string]interface{})["Username"] = username

	if err := util.WriteJSONFile(
		dappvpnconf, "", jsonIdent, &conf); err != nil {
		logger.Fatal("failed to write dappvpn config")
	}
}
