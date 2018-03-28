// +build !noethtest

package lib

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"eth/contract"
	"eth/lib/tests"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"testing"
)

var (
	PrivateKey = ""
	PSCAddress = ""

	// Test sets of dummy data.
	addr1       = [20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	addr2       = [20]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}
	b32Zero     = [32]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	b32Full     = [32]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}
	u256Zero, _ = NewUint256("0")
	u256Full, _ = NewUint256("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	u192Zero, _ = NewUint192("0")
)

// fetchPSCAddress returns address of PSC is the currently active test chain.
// is case of successfully retrieved address  - caches retrieved address and returns it on the next calls,
// instead of doing redundant requests.
func fetchPSCAddress() string {
	if PSCAddress != "" {
		return PSCAddress
	}

	truffleAPI := tests.GethEthereumConfig().TruffleAPI
	response, err := http.Get(truffleAPI.Interface() + "/getPSC")
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

	PSCAddress = data["contract"].(map[string]interface{})["address"].(string)
	return PSCAddress
}

// fetchTestPrivateKey returns first available private key, that is provided by the truffle.
func fetchTestPrivateKey() string {
	if PrivateKey != "" {
		return PrivateKey
	}

	truffleAPI := tests.GethEthereumConfig().TruffleAPI
	response, err := http.Get(truffleAPI.Interface() + "/getKeys")
	if err != nil || response.StatusCode != 200 {
		log.Fatal("Can't fetch private key. It seems that test environment is broken.")
	}

	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		log.Fatal("Can't read response body. It seems that test environment is broken.")
	}

	data := make([]interface{}, 0, 0)
	json.Unmarshal(body, &data)

	PrivateKey = data[0].(map[string]interface{})["privateKey"].(string)
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
	failOnErr(err, "Failed to connect to the EthereumConf client")

	pKeyBytes, err := hex.DecodeString(fetchTestPrivateKey())
	failOnErr(err, "Failed to fetch test private key from the API")

	key, err := crypto.ToECDSA(pKeyBytes)
	failOnErr(err, "Failed parse received test private key")

	auth := bind.NewKeyedTransactor(key)

	// Events populating
	_, err = psc.ThrowEventLogChannelCreated(auth, addr1, addr2, b32Zero, big.NewInt(0), b32Full)
	failOnErr(err, "Failed to call ThrowEventLogChannelCreated")

	_, err = psc.ThrowEventLogChannelToppedUp(auth, addr1, addr2, 0, b32Full, big.NewInt(0))
	failOnErr(err, "Failed to call ThrowEventLogChannelToppedUp")

	// todo: uncomment when API would be fixed
	//_, err = psc.ThrowEventLogChannelCloseRequested(
	//	auth, addr1, addr2, 0, b32Full, big.NewInt(0))
	//failOnErr(err, "Failed to call ThrowEventLogChannelCloseRequested")

	_, err = psc.ThrowEventLogServiceOfferingCreated(auth, addr1, b32Zero, big.NewInt(0), 0)
	failOnErr(err, "Failed to call ThrowEventLogServiceOfferingCreated")

	_, err = psc.ThrowEventLogServiceOfferingDeleted(auth, b32Zero)
	failOnErr(err, "Failed to call ThrowEventLogServiceOfferingDeleted")

	_, err = psc.ThrowEventLogServiceOfferingEndpoint(auth, addr1, b32Zero, 0, b32Full)
	failOnErr(err, "Failed to call ThrowEventLogServiceOfferingEndpoint")

	_, err = psc.ThrowEventLogServiceOfferingSupplyChanged(auth, b32Zero, 0)
	failOnErr(err, "Failed to call ThrowEventLogServiceOfferingSupplyChanged")

	_, err = psc.ThrowEventLogServiceOfferingPopedUp(auth, b32Zero)
	failOnErr(err, "Failed to call ThrowEventLogServiceOfferingPopedUp")

	_, err = psc.ThrowEventLogCooperativeChannelClose(auth, addr1, addr2, 0, b32Full, big.NewInt(0))
	failOnErr(err, "Failed to call ThrowEventLogCooperativeChannelClose")

	_, err = psc.ThrowEventLogUnCooperativeChannelClose(auth, addr1, addr2, 0, b32Full, big.NewInt(0))
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
		event, err := NewEventChannelCreated([4]string{topics[0], topics[1], topics[2], topics[3]}, data)
		failOnErr(err, "Can't create EventChannelCreated")

		client := event.Client.Bytes()
		agent := event.Agent.Bytes()

		cmpBytes(client[:], addr1[:], "ChannelCreated: client is unexpected")
		cmpBytes(agent[:], addr2[:], "ChannelCreated: agent is unexpected")
		cmpU256(event.OfferingHash, u256Zero, "ChannelCreated: offering hash is unexpected")
		cmpU192(event.Deposit, u192Zero, "ChannelCreated: deposit is unexpected")
	}

	{
		topics, data := fetchEventData(EthDigestChannelToppedUp)
		event, err := NewEventChannelToppedUp([4]string{topics[0], topics[1], topics[2], topics[3]}, data)
		failOnErr(err, "Can't create EventChannelToppedUp")

		client := event.Client.Bytes()
		agent := event.Agent.Bytes()

		cmpBytes(client[:], addr1[:], "ChannelToppedUp: client address is unexpected")
		cmpBytes(agent[:], addr2[:], "ChannelToppedUp: agent address is unexpected")
		cmpU256(event.OpenBlockNumber, u256Zero, "ChannelToppedUp: open block number is unexpected")
		cmpU256(event.OfferingHash, u256Full, "ChannelToppedUp: offering hash is unexpected")
		cmpU192(event.AddedDeposit, u192Zero, "ChannelToppedUp: added deposit is unexpected")
	}

	{
		topics, data := fetchEventData(EthOfferingCreated)
		event, err := NewEventServiceOfferingCreated([3]string{topics[0], topics[1], topics[2]}, data)
		failOnErr(err, "Can't create EventOfferingCreated")

		agent := event.Agent.Bytes()
		cmpBytes(agent[:], addr1[:], "OfferingCreated: client address is unexpected")
		cmpU256(event.OfferingHash, u256Zero, "OfferingCreated: offering hash is unexpected")
		cmpU192(event.MinDeposit, u192Zero, "OfferingCreated: min deposit is unexpected")
		cmpU192(event.CurrentSupply, u192Zero, "OfferingCreated: current supply is unexpected")
	}

	{
		topics, _ := fetchEventData(EthOfferingDeleted)
		event, err := NewEventServiceOfferingDeleted([2]string{topics[0], topics[1]})
		failOnErr(err, "Can't create EventServiceOfferingDeleted")

		cmpU256(event.OfferingHash, u256Zero, "OfferingCreated: offering hash is unexpected")
	}

	{
		topics, data := fetchEventData(EthServiceOfferingEndpoint)
		event, err := NewEventServiceOfferingEndpoint([4]string{topics[0], topics[1], topics[2], topics[3]}, data)
		failOnErr(err, "Can't create EventServiceOfferingEndpoint")

		client := event.Client.Bytes()
		cmpBytes(client[:], addr1[:], "ServiceOfferingEndpoint: client address is unexpected")
		cmpU256(event.OfferingHash, u256Zero, "ServiceOfferingEndpoint: offering hash is unexpected")
		cmpU256(event.OpenBlockNumber, u256Zero, "ServiceOfferingEndpoint: open block number is unexpected")
		cmpU256(event.EndpointHash, u256Full, "ServiceOfferingEndpoint: endpoint hash is unexpected")
	}

	{
		topics, data := fetchEventData(EthServiceOfferingSupplyChanged)
		event, err := NewEventServiceOfferingSupplyChanged([2]string{topics[0], topics[1]}, data)
		failOnErr(err, "Can't create EventServiceOfferingSupplyChanged")

		cmpU256(event.OfferingHash, u256Zero, "ServiceOfferingSupplyChanged: offering hash is unexpected")
		cmpU192(event.CurrentSupply, u192Zero, "ServiceOfferingSupplyChanged: current supply is unexpected")
	}

	{
		topics, _ := fetchEventData(EthServiceOfferingPoppedUp)
		event, err := NewEventServiceOfferingPoppedUp([2]string{topics[0], topics[1]})
		failOnErr(err, "Can't create EventServiceOfferingPoppedUp")
		cmpU256(event.OfferingHash, u256Zero, "ServiceOfferingPoppedUp: offering hash is unexpected")
	}

	{
		topics, data := fetchEventData(EthCooperativeChannelClose)
		event, err := NewEventCooperativeChannelClose([4]string{topics[0], topics[1], topics[2], topics[3]}, data)
		failOnErr(err, "Can't create EventCooperativeChannelClose")

		client := event.Client.Bytes()
		agent := event.Agent.Bytes()

		cmpBytes(client[:], addr1[:], "CooperativeChannelClose: client is unexpected")
		cmpBytes(agent[:], addr2[:], "CooperativeChannelClose: agent is unexpected")
		cmpU256(event.OpenBlockNumber, u256Zero, "CooperativeChannelClose: open block number is unexpected")
		cmpU256(event.OfferingHash, u256Full, "CooperativeChannelClose: offering hash is unexpected")
		cmpU192(event.Balance, u192Zero, "CooperativeChannelClose: balance is unexpected")
	}

	{
		topics, data := fetchEventData(EthUncooperativeChannelClose)
		event, err := NewEventUnCooperativeChannelClose([4]string{topics[0], topics[1], topics[2], topics[3]}, data)
		failOnErr(err, "Can't create EventUnCooperativeChannelClose")

		client := event.Client.Bytes()
		agent := event.Agent.Bytes()

		cmpBytes(client[:], addr1[:], "UnCooperativeChannelClose: client is unexpected")
		cmpBytes(agent[:], addr2[:], "UnCooperativeChannelClose: agent is unexpected")
		cmpU256(event.OpenBlockNumber, u256Zero, "UnCooperativeChannelClose: open block number is unexpected")
		cmpU256(event.OfferingHash, u256Full, "UnCooperativeChannelClose: offering hash is unexpected")
		cmpU192(event.Balance, u192Zero, "UnCooperativeChannelClose: balance is unexpected")
	}
}

func TestNegativeLogsFetching(t *testing.T) {
	failIfNoError := func(err error, args ...interface{}) {
		if err == nil {
			t.Fatal(args)
		}
	}

	node := tests.GethEthereumConfig().Geth
	client := NewEthereumClient(node.Host, node.Port)

	_, err := client.GetLogs("", []string{"0x0"}, "", "")
	failIfNoError(err, "Error must be returned")

	_, err = client.GetLogs(fetchPSCAddress(), []string{"0x0"}, "", "")
	failIfNoError(err, "Error must be returned")

	_, err = client.GetLogs(fetchPSCAddress(), []string{"", ""}, "", "")
	failIfNoError(err, "Error must be returned")
}

func TestLogsFetchingWithBrokenNetwork(t *testing.T) {
	node := tests.GethEthereumConfig().Geth
	client := NewEthereumClient(node.Host, node.Port+1) // Note: invalid port is used.

	{
		_, err := client.GetLogs(fetchPSCAddress(), []string{EthOfferingCreated}, "", "")
		if err == nil {
			t.Fatal("Error must be returned")
		}
	}
}
