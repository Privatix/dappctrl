package deals

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
)

// Deal types.
const (
	TypeCoop   = "coop"
	TypeUncoop = "uncoop"
)

// Deal is a detailed deals closed over blockchain between given angent and client.
type Deal struct {
	Agent      int     `json:"agent"`
	Client     int     `json:"client"`
	PrixClosed float64 `json:"prixClosed"`
	Type       string  `json:"type"`
}

// Account is an ethereum account
type Account struct {
	EthAddr    data.HexString
	PrivateKey data.HexString
}

// Make makes and closes deals.
func Make(back eth.Backend, agents, clients []Account, deals []Deal) {
	var wg sync.WaitGroup
	wg.Add(len(deals))
	var mu sync.Mutex
	nonces := make(map[data.HexString]uint64, 0)
	for _, deal := range deals {
		go func(d Deal) {
			if err := proccessDeal(back, agents, clients, nonces, &mu, d); err != nil {
				log.Panic(err)
			}
			wg.Done()
		}(deal)
	}

	wg.Wait()
}

func proccessDeal(back eth.Backend, agents, clients []Account, nonces map[data.HexString]uint64, mu *sync.Mutex, d Deal) error {
	agent := agents[d.Agent]
	client := clients[d.Client]
	balance := uint64((d.PrixClosed * 10e6))

	// Random offering hash.
	randomBytes := make([]byte, common.HashLength)
	rand.Read(randomBytes)
	offerHash := [common.HashLength]byte(common.BytesToHash(randomBytes))

	agentAddr, err := data.HexToAddress(agent.EthAddr)
	if err != nil {
		return fmt.Errorf("could not parse agent address: %v", err)
	}

	clientAddr, err := data.HexToAddress(client.EthAddr)
	if err != nil {
		return fmt.Errorf("could not convert clients addr: %v", err)
	}

	// Register offering.
	log.Println(fmt.Sprintf("registering service offering for agent `%s`...", agent.EthAddr))
	if auth, err := newAuth(agent); err != nil {
		return fmt.Errorf("could not use agents key: %v", err)
	} else {
		mu.Lock()
		if _, ok := nonces[agent.EthAddr]; !ok {
			val, err := back.PendingNonceAt(context.Background(), agentAddr)
			if err != nil {
				mu.Unlock()
				return fmt.Errorf("could not get pending nonce: %v", err)
			}
			nonces[agent.EthAddr] = val
		}
		auth.Nonce = new(big.Int).SetUint64(nonces[agent.EthAddr])
		nonces[agent.EthAddr]++
		mu.Unlock()
		if _, err := back.RegisterServiceOffering(auth, offerHash, balance, 1, 0, data.FromBytes([]byte{})); err != nil {
			return fmt.Errorf("could not register offering: %v", err)
		}
	}

	// Wait for offering registeration.
	time.Sleep(time.Minute)

	// Accept offering.
	log.Println("accepting service offering...")

	clientAuth, err := newAuth(client)
	if err != nil {
		return fmt.Errorf("could not use clients key: %v", err)
	}
	mu.Lock()
	if _, ok := nonces[client.EthAddr]; !ok {
		val, err := back.PendingNonceAt(context.Background(), clientAddr)
		if err != nil {
			mu.Unlock()
			return fmt.Errorf("could not get pending nonce: %v", err)
		}
		nonces[client.EthAddr] = val
	}
	clientAuth.Nonce = new(big.Int).SetUint64(nonces[client.EthAddr])
	nonces[client.EthAddr]++
	mu.Unlock()
	if _, err := back.PSCCreateChannel(clientAuth, agentAddr, offerHash, balance); err != nil {
		return fmt.Errorf("could not accept offering: %v", err)
	}
	blockStart, err := back.LatestBlockNumber(context.Background())
	if err != nil {
		return err
	}

	// Wait for accept.
	time.Sleep(time.Minute)

	logs, err := back.FilterLogs(context.Background(), ethereum.FilterQuery{
		FromBlock: blockStart,
		Addresses: []common.Address{back.PSCAddress()},
		Topics:    [][]common.Hash{{eth.ServiceChannelCreated}, {agentAddr.Hash()}, {clientAddr.Hash()}, {common.BytesToHash(offerHash[:])}},
	})
	if len(logs) != 1 {
		return fmt.Errorf(fmt.Sprintf("wanted only channel created log, got %v", logs))
	}
	block := logs[0].BlockNumber

	hash := eth.BalanceProofHash(back.PSCAddress(), agentAddr, uint32(block), offerHash, balance)

	clientKey, err := crypto.HexToECDSA(string(client.PrivateKey))
	if err != nil {
		return fmt.Errorf("could not convert clients key: %v", err)
	}

	balanceSig, err := crypto.Sign(hash, clientKey)
	if err != nil {
		return fmt.Errorf("could not sign balance")
	}

	// Close channel.
	if d.Type == TypeCoop {
		log.Println("closing channel cooperativelly...")
		// Cooperative close.
		closingHash := eth.BalanceClosingHash(clientAddr, back.PSCAddress(), uint32(block), offerHash, balance)
		if agentKey, err := crypto.HexToECDSA(string(agent.PrivateKey)); err != nil {
			return fmt.Errorf("could not convert agents key: %v", err)
		} else if agentAuth, err := newAuth(agent); err != nil {
			return fmt.Errorf("could not use clients key: %v", err)
		} else if closingSig, err := crypto.Sign(closingHash, agentKey); err != nil {
			return fmt.Errorf("could not sign closing hash: %v", err)
		} else {
			mu.Lock()
			agentAuth.Nonce = new(big.Int).SetUint64(nonces[agent.EthAddr])
			nonces[agent.EthAddr]++
			mu.Unlock()
			tx, err := back.CooperativeClose(agentAuth, agentAddr, uint32(block), offerHash, balance, balanceSig, closingSig)
			if err != nil {
				return fmt.Errorf("could not cooperative close: %v", err)
			}
			fmt.Printf("coop close tx: %s\n", tx.Hash().Hex())
		}
	} else if d.Type == TypeUncoop {
		log.Println("closing channel uncoperativelly...")
		if clientAuth, err := newAuth(client); err != nil {
			return fmt.Errorf("could not use clients key: %v", err)
		} else {
			mu.Lock()
			clientAuth.Nonce = new(big.Int).SetUint64(nonces[client.EthAddr])
			nonces[client.EthAddr]++
			mu.Unlock()
			if _, err := back.PSCUncooperativeClose(clientAuth, agentAddr, uint32(block), offerHash, balance); err != nil {
				return fmt.Errorf("could not uncoop close channel: %v", err)
			}
		}
		// Challange period is 20.
		time.Sleep(20 * 16 * time.Second) // 16 for each block.
		if clientAuth, err := newAuth(client); err != nil {
			return fmt.Errorf("could not use clients key: %v", err)
		} else {
			mu.Lock()
			clientAuth.Nonce = new(big.Int).SetUint64(nonces[client.EthAddr])
			nonces[client.EthAddr]++
			mu.Unlock()
			if _, err := back.PSCSettle(clientAuth, agentAddr, uint32(block), offerHash); err != nil {
				return fmt.Errorf("could not settle channel: %v", err)
			}
		}
	} else {
		return fmt.Errorf("unknown deal type: %v", d.Type)
	}
	return nil
}

// CreateAccounts just creates accounts.
func CreateAccounts(n int) ([]Account, error) {
	ret := make([]Account, n)

	for i := range ret {
		acc, err := newAccount()
		if err != nil {
			return nil, err
		}
		ret[i] = acc
	}

	return ret, nil
}

func newAccount() (acc Account, err error) {
	pk, err := crypto.GenerateKey()
	if err != nil {
		return acc, err
	}

	acc.PrivateKey = data.HexFromBytes(crypto.FromECDSA(pk))
	acc.EthAddr = data.HexFromBytes(crypto.PubkeyToAddress(pk.PublicKey).Bytes())

	return
}

// TransferToPSC transfers all PTC to PSC.
func TransferToPSC(back eth.Backend, acc Account) error {
	addr, err := data.HexToAddress(acc.EthAddr)
	if err != nil {
		return fmt.Errorf("could not convert account hex address to ethereum address: %v", err)
	}

	amount, err := back.PTCBalanceOf(new(bind.CallOpts), addr)
	if err != nil {
		return fmt.Errorf("could not get accounts ptc balance: %v", err)
	}
	if amount.Uint64() == 0 {
		return fmt.Errorf("address=`%v` PTC balance is zerro", acc.EthAddr)
	}

	if auth, err := newAuth(acc); err != nil {
		return fmt.Errorf("could not use acc key: %v", err)
	} else if _, err := back.PTCIncreaseApproval(auth,
		back.PSCAddress(), amount); err != nil {
		return fmt.Errorf("could not increase approval: %v", err)
	}

	// Assuming approve transaction passes after a minute.
	time.Sleep(time.Minute)

	if auth, err := newAuth(acc); err != nil {
		return fmt.Errorf("could not use acc key: %v", err)
	} else if _, err := back.PSCAddBalanceERC20(auth, amount.Uint64()); err != nil {
		return fmt.Errorf("could not add balance: %v", err)
	}
	return nil
}

func newAuth(acc Account) (*bind.TransactOpts, error) {
	key, err := crypto.HexToECDSA(string(acc.PrivateKey))
	if err != nil {
		return nil, fmt.Errorf("could not convert key: %v", err)
	}

	auth := bind.NewKeyedTransactor(key)
	auth.GasLimit = 200000
	auth.GasPrice = new(big.Int).SetUint64(20000000000)
	return auth, nil
}
