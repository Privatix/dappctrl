package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/tool/rating/deals"
	utillog "github.com/privatix/dappctrl/util/log"
)

// Config is a program configuration.
type Config struct {
	Eth eth.Config
}

func main() {
	var in, out, configFile string
	flag.StringVar(&configFile, "config", "config.json", "--config config.json")
	flag.StringVar(&in, "in", "in.json", "--in in.json")
	flag.StringVar(&out, "out", "out.json", "--out out.json")
	flag.Parse()

	// Create a logger.
	logger, err := utillog.NewStderrLogger(utillog.NewWriterConfig())
	if err != nil {
		panic(err)
	}

	// Parse program input, eg deals.
	ratingDeals := make([]deals.Deal, 0)
	readJSONFile(in, &ratingDeals)

	// Parse config.
	var conf Config
	readJSONFile(configFile, &conf)

	// Create deals accounts.
	clients, agents := createAccounts(ratingDeals)

	// Ask user to transfer ptc and eth.
	for _, acc := range agents {
		fmt.Println(acc.EthAddr)
	}
	for _, acc := range clients {
		fmt.Println(acc.EthAddr)
	}
	fmt.Println("Send eth and prix to these addresses")
	fmt.Println("Only when finished, press 'Enter' to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	// Create eth client.
	ethback := eth.NewBackend(&conf.Eth, logger)

	// Transfer everythin to PSC.
	var wg sync.WaitGroup
	wg.Add(len(agents) + len(clients))
	for _, acc := range agents {
		go func(acc deals.Account) {
			if err := deals.TransferToPSC(ethback, acc); err != nil {
				log.Panic(err)
			}
			wg.Done()
		}(acc)
	}
	for _, acc := range clients {
		go func(acc deals.Account) {
			if err := deals.TransferToPSC(ethback, acc); err != nil {
				log.Panic(err)
			}
			wg.Done()
		}(acc)
	}

	logger.Info("Wait for transfers to complete...")
	wg.Wait()

	logger.Info("Making deals...")

	// Make program output.
	allAccs := make(map[string][]deals.Account)
	allAccs["clients"] = clients
	allAccs["agents"] = agents

	// Perform deals and if all good, save accounts to output file.
	deals.Make(ethback, agents, clients, ratingDeals)

	if f, err := os.Create(out); err != nil {
		log.Panic(err)
	} else if err := json.NewEncoder(f).Encode(&allAccs); err != nil {
		log.Panic(err)
	}

	// Wait for stdout to finish.
	logger.Info("All done.")
}

func readJSONFile(f string, v interface{}) {
	if f, err := os.Open(f); err != nil {
		log.Panic(err)
	} else if err := json.NewDecoder(f).Decode(v); err != nil {
		log.Panic(err)
	}
}

func createAccounts(ratingDeals []deals.Deal) ([]deals.Account, []deals.Account) {
	var countClients, countAgents int
	for _, d := range ratingDeals {
		if d.Agent+1 > countAgents {
			countAgents = d.Agent + 1
		}
		if d.Client+1 > countClients {
			countClients = d.Client + 1
		}
	}

	clients, err := deals.CreateAccounts(countClients)
	if err != nil {
		log.Panic(err)
	}
	agents, err := deals.CreateAccounts(countAgents)
	if err != nil {
		log.Panic(err)
	}

	return clients, agents
}
