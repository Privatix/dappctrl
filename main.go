package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/privatix/dappctrl/agent/uisrv"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/eth/contract"
	"github.com/privatix/dappctrl/eth/truffle"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/pay"
	"github.com/privatix/dappctrl/sesssrv"
	"github.com/privatix/dappctrl/somc"
	"github.com/privatix/dappctrl/util"
)

type ethConfig struct {
	Contract struct {
		PTCAddr string
		PSCAddr string
	}
	GethURL       string
	TruffleAPIURL string
}
type config struct {
	AgentServer   *uisrv.Config
	Eth           *ethConfig
	DB            *data.DBConfig
	Job           *job.Config
	Log           *util.LogConfig
	PayServer     *pay.Config
	SessionServer *sesssrv.Config
	SOMC          *somc.Config
}

func newConfig() *config {
	return &config{
		DB:   data.NewDBConfig(),
		Job:  job.NewConfig(),
		Log:  util.NewLogConfig(),
		SOMC: somc.NewConfig(),
	}
}

func fetchPSCAddress(truffleAPI string) string {
	response, err := http.Get(truffleAPI + "/getPSC")
	if err != nil || response.StatusCode != http.StatusOK {
		log.Fatal("Can't fetch PSC address. It seems that test environment is broken.")
	}

	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		log.Fatal("Can't read response body. It seems that test environment is broken.")
	}

	data := make(map[string]interface{})
	json.Unmarshal(body, &data)

	return data["contract"].(map[string]interface{})["address"].(string)
}

func fetchPTCAddress(truffleAPI string) string {
	response, err := http.Get(truffleAPI + "/getPrix")
	if err != nil || response.StatusCode != 200 {
		log.Fatal("Can't fetch PSC address. It seems that test environment is broken.")
	}

	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		log.Fatal("Can't read response body. It seems that test environment is broken.")
	}

	data := make(map[string]interface{})
	json.Unmarshal(body, &data)

	return data["contract"].(map[string]interface{})["address"].(string)
}

func readConfig(conf *config) {
	fconfig := flag.String(
		"config", "dappctrl.config.json", "Configuration file")
	flag.Parse()
	if err := util.ReadJSONFile(*fconfig, &conf); err != nil {
		log.Fatalf("failed to read configuration: %s", err)
	}
	// If test truffle api is specified, pull and update contract addresses.
	if conf.Eth.TruffleAPIURL != "" {
		api := truffle.API(conf.Eth.TruffleAPIURL)
		conf.Eth.Contract.PSCAddr = api.FetchPSCAddress()
		conf.Eth.Contract.PTCAddr = api.FetchPTCAddress()
	}
}

func main() {
	conf := newConfig()
	readConfig(conf)

	logger, err := util.NewLogger(conf.Log)
	if err != nil {
		log.Fatalf("failed to create logger: %s", err)
	}

	db, err := data.NewDB(conf.DB, logger)
	if err != nil {
		logger.Fatal("failed to open db connection: %s", err)
	}
	defer data.CloseDB(db)

	gethConn, err := ethclient.Dial(conf.Eth.GethURL)
	if err != nil {
		logger.Fatal("failed to dial geth node: %v", err)
	}

	ethClient := eth.NewEthereumClient(conf.Eth.GethURL)

	ptcAddr := common.StringToAddress(conf.Eth.Contract.PTCAddr)
	ptc, err := contract.NewPrivatixTokenContract(ptcAddr, gethConn)
	if err != nil {
		logger.Fatal("failed to create ptc instance: %v", err)
	}

	pscAddr := common.StringToAddress(conf.Eth.Contract.PSCAddr)

	psc, err := contract.NewPrivatixServiceContract(pscAddr, gethConn)
	if err != nil {
		logger.Fatal("failed to create psc intance: %v", err)
	}

	uiSrv := uisrv.NewServer(conf.AgentServer, logger, db, ethClient, ptc, psc)
	go func() {
		logger.Fatal("failed to run agent server: %s\n",
			uiSrv.ListenAndServe())
	}()

	paySrv := pay.NewServer(conf.PayServer, logger, db)
	go func() {
		logger.Fatal("failed to start pay server: %s",
			paySrv.ListenAndServe())
	}()

	sess := sesssrv.NewServer(conf.SessionServer, logger, db)
	go func() {
		logger.Fatal("failed to start session server: %s",
			sess.ListenAndServe())
	}()

	queue := job.NewQueue(conf.Job, logger, db, jobHandlers)
	logger.Fatal("failed to process job queue: %s", queue.Process())
}
