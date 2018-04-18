// +build !noethtest

package lib

import (
	"encoding/hex"
	"log"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"dappctrl/eth/utils"
	"bytes"
	"github.com/privatix/dappctrl/eth/contract"
	"github.com/privatix/dappctrl/eth/lib/tests"
)

var (
	PrivateKey = ""
	PSCAddress = ""

	// Test sets of dummy data.
	Addr1       = [20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	Addr2       = [20]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}
	B32Zero     = [32]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	B32Full     = [32]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}
	U256Zero, _ = NewUint256("0")
	U256Full, _ = NewUint256("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	U192Zero, _ = NewUint192("0")
)

// fetchPSCAddress returns address of PSC is the currently active test chain.
// is case of successfully retrieved address  - caches retrieved address and returns it on the next calls,
// instead of doing redundant requests.
func fetchPSCAddress() string {
	if PSCAddress != "" {
		return PSCAddress
	}

	PSCAddress = utils.FetchPSCAddress()
	return PSCAddress
}

// fetchTestPrivateKey returns first available private key, that is provided by the truffle.
func fetchTestPrivateKey() string {
	if PrivateKey != "" {
		return PrivateKey
	}

	PrivateKey = utils.FetchTestPrivateKey()
	return PrivateKey
}

func populateEvents() {
	failOnErr := func(err error, args ...interface{}) {
		if err != nil {
			log.Fatal(args, " / Error details: ", err)
		}
	}

	geth := tests.GethEthereumConfig().Geth
	conn, err := ethclient.Dial(geth.Interface())
	failOnErr(err, "Failed to connect to the EthereumConf client")

	contractAddress, err := NewAddress(fetchPSCAddress())
	failOnErr(err, "Failed to parse received contract address")

	psc, err := contract.NewPrivatixServiceContract(contractAddress.Bytes(), conn)
	failOnErr(err, "Failed to connect to the Ethereum client")

	pKeyBytes, err := hex.DecodeString(fetchTestPrivateKey())
	failOnErr(err, "Failed to fetch test private key from the API")

	key, err := crypto.ToECDSA(pKeyBytes)
	failOnErr(err, "Failed to parse received test private key")

	auth := bind.NewKeyedTransactor(key)

	// Events populating
	_, err = psc.ThrowEventLogChannelCreated(auth, Addr1, Addr2, B32Zero, big.NewInt(0), B32Full)
	failOnErr(err, "Failed to call ThrowEventLogChannelCreated")

	_, err = psc.ThrowEventLogChannelToppedUp(auth, Addr1, Addr2, B32Full, 0, big.NewInt(0))
	failOnErr(err, "Failed to call ThrowEventLogChannelToppedUp")

	_, err = psc.ThrowEventLogChannelCloseRequested(auth, Addr1, Addr2, B32Full, 0, big.NewInt(0))
	failOnErr(err, "Failed to call ThrowEventLogChannelCloseRequested")

	_, err = psc.ThrowEventLogOfferingCreated(auth, Addr1, B32Zero, big.NewInt(0), 0)
	failOnErr(err, "Failed to call ThrowEventLogOfferingCreated")

	_, err = psc.ThrowEventLogOfferingDeleted(auth, Addr1, B32Zero)
	failOnErr(err, "Failed to call ThrowEventLogOfferingDeleted")

	_, err = psc.ThrowEventLogOfferingEndpoint(auth, Addr1, Addr2, B32Zero, 0, B32Full)
	failOnErr(err, "Failed to call ThrowEventLogOfferingEndpoint")

	_, err = psc.ThrowEventLogOfferingSupplyChanged(auth, Addr1, B32Zero, 0)
	failOnErr(err, "Failed to call ThrowEventLogOfferingSupplyChanged")

	_, err = psc.ThrowEventLogOfferingPopedUp(auth, Addr1, B32Zero)
	failOnErr(err, "Failed to call ThrowEventLogOfferingPopedUp")

	_, err = psc.ThrowEventLogCooperativeChannelClose(auth, Addr1, Addr2, B32Full, 0, big.NewInt(0))
	failOnErr(err, "Failed to call ThrowEventLogCooperativeChannelClose")

	_, err = psc.ThrowEventLogUnCooperativeChannelClose(auth, Addr1, Addr2, B32Full, 0, big.NewInt(0))
	failOnErr(err, "Failed to call ThrowEventLogUnCooperativeChannelClose")
}

func TestNormalLogsFetching(t *testing.T) {
	populateEvents()
	node := tests.GethEthereumConfig().Geth
	client := NewEthereumClient(node.Host, node.Port)

	failOnErr := func(err error, args ...interface{}) {
		if err != nil {
			t.Fatal(args, " / Error details: ", err)
		}
	}

	cmpBytes := func(a, b []byte, errorMessage string) {
		if bytes.Compare(a, b) != 0 {
			t.Fatal(errorMessage)
		}
	}

	cmpU256 := func(a, b *Uint256, errorMessage string) {
		if a.String() != b.String() {
			t.Fatal(errorMessage)
		}
	}

	cmpU192 := func(a, b *Uint192, errorMessage string) {
		if a.String() != b.String() {
			t.Fatal(errorMessage)
		}
	}

	fetchEventData := func(eventDigest string) ([]string, string) {
		response, err := client.GetLogs(
			fetchPSCAddress(),
			[]string{"0x" + eventDigest}, "", "")

		failOnErr(err, "Can't call API: ", err, " Event digest: ", eventDigest)
		if len(response.Result) == 0 {
			t.Fatal("Can't fetch result. Event digest: ", eventDigest)
		}

		return response.Result[0].Topics, response.Result[0].Data
	}

	{
		topics, data := fetchEventData(EthDigestChannelCreated)
		event, err := NewChannelCreatedEvent([4]string{topics[0], topics[1], topics[2], topics[3]}, data)
		failOnErr(err, "Can't create ChannelCreatedEvent")

		agent := event.Agent.Bytes()
		client := event.Client.Bytes()

		cmpBytes(agent[:], Addr1[:], "ChannelCreated: agent is unexpected")
		cmpBytes(client[:], Addr2[:], "ChannelCreated: client is unexpected")
		cmpU256(event.OfferingHash, U256Zero, "ChannelCreated: offering hash is unexpected")
		cmpU192(event.Deposit, U192Zero, "ChannelCreated: deposit is unexpected")
	}

	{
		topics, data := fetchEventData(EthDigestChannelToppedUp)
		event, err := NewChannelToppedUpEvent([4]string{topics[0], topics[1], topics[2], topics[3]}, data)
		failOnErr(err, "Can't create ChannelToppedUpEvent")

		agent := event.Agent.Bytes()
		client := event.Client.Bytes()

		cmpBytes(agent[:], Addr1[:], "ChannelToppedUp: agent is unexpected")
		cmpBytes(client[:], Addr2[:], "ChannelToppedUp: client is unexpected")
		cmpU256(event.OpenBlockNumber, U256Zero, "ChannelToppedUp: open block number is unexpected")
		cmpU256(event.OfferingHash, U256Full, "ChannelToppedUp: offering hash is unexpected")
		cmpU192(event.AddedDeposit, U192Zero, "ChannelToppedUp: added deposit is unexpected")
	}

	{
		topics, data := fetchEventData(EthOfferingCreated)
		event, err := NewOfferingCreatedEvent([4]string{topics[0], topics[1], topics[2], topics[3]}, data)
		failOnErr(err, "Can't create OfferingCreatedEvent")

		agent := event.Agent.Bytes()
		cmpBytes(agent[:], Addr1[:], "OfferingCreated: agent address is unexpected")
		cmpU256(event.OfferingHash, U256Zero, "OfferingCreated: offering hash is unexpected")
		cmpU192(event.MinDeposit, U192Zero, "OfferingCreated: min deposit is unexpected")
		cmpU256(event.CurrentSupply, U256Zero, "OfferingCreated: current supply is unexpected")
	}

	{
		topics, _ := fetchEventData(EthOfferingDeleted)
		event, err := NewOfferingDeletedEvent([3]string{topics[0], topics[1], topics[2]})
		failOnErr(err, "Can't create EventOfferingDeleted")

		agent := event.Agent.Bytes()
		cmpBytes(agent[:], Addr1[:], "OfferingDeleted: agent address is unexpected")
		cmpU256(event.OfferingHash, U256Zero, "OfferingDeleted: offering hash is unexpected")
	}

	{
		topics, data := fetchEventData(EthOfferingEndpoint)
		event, err := NewOfferingEndpointEvent([4]string{topics[0], topics[1], topics[2], topics[3]}, data)
		failOnErr(err, "Can't create EventOfferingEndpoint")

		agent := event.Agent.Bytes()
		client := event.Client.Bytes()
		cmpBytes(agent[:], Addr1[:], "OfferingEndpoint: agent address is unexpected")
		cmpBytes(client[:], Addr2[:], "OfferingEndpoint: client address is unexpected")
		cmpU256(event.OfferingHash, U256Zero, "OfferingEndpoint: offering hash is unexpected")
		cmpU256(event.OpenBlockNumber, U256Zero, "OfferingEndpoint: open block number is unexpected")
		cmpU256(event.EndpointHash, U256Full, "OfferingEndpoint: endpoint hash is unexpected")
	}

	{
		topics, data := fetchEventData(EthOfferingSupplyChanged)
		event, err := NewOfferingSupplyChangedEvent([3]string{topics[0], topics[1], topics[2]}, data)
		failOnErr(err, "Can't create EventOfferingSupplyChanged")

		agent := event.Agent.Bytes()
		cmpBytes(agent[:], Addr1[:], "OfferingSupplyChanged: agent address is unexpected")
		cmpU256(event.OfferingHash, U256Zero, "OfferingSupplyChanged: offering hash is unexpected")
		cmpU192(event.CurrentSupply, U192Zero, "OfferingSupplyChanged: current supply is unexpected")
	}

	{
		topics, _ := fetchEventData(EthOfferingPoppedUp)
		event, err := NewOfferingPoppedUpEvent([3]string{topics[0], topics[1], topics[2]})
		failOnErr(err, "Can't create EventOfferingPoppedUp")

		agent := event.Agent.Bytes()
		cmpBytes(agent[:], Addr1[:], "OfferingPoppedUp: agent address is unexpected")
		cmpU256(event.OfferingHash, U256Zero, "OfferingPoppedUp: offering hash is unexpected")
	}

	{
		topics, data := fetchEventData(EthCooperativeChannelClose)
		event, err := NewCooperativeChannelCloseEvent([4]string{topics[0], topics[1], topics[2], topics[3]}, data)
		failOnErr(err, "Can't create CooperativeChannelCloseEvent")

		client := event.Client.Bytes()
		agent := event.Agent.Bytes()

		cmpBytes(agent[:], Addr1[:], "CooperativeChannelClose: client is unexpected")
		cmpBytes(client[:], Addr2[:], "CooperativeChannelClose: agent is unexpected")
		cmpU256(event.OfferingHash, U256Full, "CooperativeChannelClose: offering hash is unexpected")
		cmpU256(event.OpenBlockNumber, U256Zero, "CooperativeChannelClose: open block number is unexpected")
		cmpU192(event.Balance, U192Zero, "CooperativeChannelClose: balance is unexpected")
	}

	{
		topics, data := fetchEventData(EthUncooperativeChannelClose)
		event, err := NewUnCooperativeChannelCloseEvent([4]string{topics[0], topics[1], topics[2], topics[3]}, data)
		failOnErr(err, "Can't create EventUnCooperativeChannelClose")

		client := event.Client.Bytes()
		agent := event.Agent.Bytes()

		cmpBytes(agent[:], Addr1[:], "CooperativeChannelClose: client is unexpected")
		cmpBytes(client[:], Addr2[:], "CooperativeChannelClose: agent is unexpected")
		cmpU256(event.OfferingHash, U256Full, "UnCooperativeChannelClose: offering hash is unexpected")
		cmpU256(event.OpenBlockNumber, U256Zero, "UnCooperativeChannelClose: open block number is unexpected")
		cmpU192(event.Balance, U192Zero, "UnCooperativeChannelClose: balance is unexpected")
	}
}
