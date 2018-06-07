package eth

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

//
// Contract events digests.
// Please, see this article for the details:
// https://codeburst.io/deep-dive-into-ethereum-logs-a8d2047c7371
//
const (
	// PSC logs.
	EthDigestChannelCreated      = "a6153987181667023837aee39c3f1a702a16e5e146323ef10fb96844a526143c"
	EthDigestChannelToppedUp     = "a3b2cd532a9050531ecc674928d7704894707ede1a436bfbee86b96b83f2a5ce"
	EthChannelCloseRequested     = "b40564b1d36572b2942ad7cfc5a5a967f3ef08c82163a910dee760c5b629a32e"
	EthOfferingCreated           = "32c1913dfde418197923027c2f2260f19903a2e86a93ed83c4689ac91a96bafd"
	EthOfferingDeleted           = "c3013cd9dd5c33b95a9cc1bc076481c9a6a1970be6d7f1ed33adafad6e57d3d6"
	EthOfferingEndpoint          = "450e7ab61f9e1c40dd7c79edcba274a7a96f025fab1733b3fa1087a1b5d1db7d"
	EthOfferingPoppedUp          = "c37352067a3ca1eafcf2dc5ba537fc473509c4e4aaca729cb1dab7053ec1ffbf"
	EthCooperativeChannelClose   = "b488ea0f49970f556cf18e57588e78dcc1d3fd45c71130aa5099a79e8b06c8e7"
	EthUncooperativeChannelClose = "7418f9b30b6de272d9d54ee6822f674042c58cea183b76d5d4e7b3c933a158f6"

	// PTC logs.
	EthTokenApproval = "8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925"
	EthTokenTransfer = "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
)

// Event received from the ethereum block-chain.
type Event interface {
	Digest() string
}

// ChannelCreatedEvent implements wrapper for contract event.
// Please see contract implementation for the details.
type ChannelCreatedEvent struct {
	Agent             common.Address // Indexed.
	Client            common.Address // Indexed.
	OfferingHash      *Uint256       // Indexed.
	Deposit           *Uint192
	AuthenticatedHash *Uint256
}

// NewChannelCreatedEvent creates event of type ChannelCreatedEvent.
// Please see contract implementation for the details.
func NewChannelCreatedEvent(topics [4]string, hexData string) (*ChannelCreatedEvent, error) {
	var err error
	e := &ChannelCreatedEvent{}

	err = validateTopics(topics[:])
	err = checkEventDigest(topics[0], e.Digest(), err)

	e.Agent, err = parseAddress(topics[1], err)
	e.Client, err = parseAddress(topics[2], err)
	e.OfferingHash, err = parseTopicAsUint256(topics[3], err)

	e.Deposit, err = parseDataFieldAsUint192(hexData, 0, err)
	e.AuthenticatedHash, err = parseDataFieldAsUint256(hexData, 1, err)
	return e, err
}

// Digest returns keccak512 hash sum of the event signature.
func (e *ChannelCreatedEvent) Digest() string {
	return EthDigestChannelCreated
}

// ChannelToppedUpEvent implements wrapper for contract event.
// Please see contract implementation for the details.
type ChannelToppedUpEvent struct {
	Agent           common.Address // Indexed.
	Client          common.Address // Indexed.
	OfferingHash    *Uint256       // Indexed.
	OpenBlockNumber *Uint256
	AddedDeposit    *Uint192
}

// NewChannelToppedUpEvent creates event of type ChannelToppedUpEvent.
// Please see contract implementation for the details.
func NewChannelToppedUpEvent(topics [4]string, hexData string) (*ChannelToppedUpEvent, error) {
	var err error
	e := &ChannelToppedUpEvent{}

	err = validateTopics(topics[:])
	err = checkEventDigest(topics[0], e.Digest(), err)

	e.Agent, err = parseAddress(topics[1], err)
	e.Client, err = parseAddress(topics[2], err)
	e.OfferingHash, err = parseTopicAsUint256(topics[3], err)

	e.OpenBlockNumber, err = parseDataFieldAsUint256(hexData, 0, err)
	e.AddedDeposit, err = parseDataFieldAsUint192(hexData, 1, err)
	return e, err
}

// Digest returns keccak512 hash sum of the event signature.
func (e *ChannelToppedUpEvent) Digest() string {
	return EthDigestChannelToppedUp
}

// ChannelCloseRequestedEvent implements wrapper for contract event.
// Please see contract implementation for the details.
type ChannelCloseRequestedEvent struct {
	Client          common.Address // Indexed.
	Agent           common.Address // Indexed.
	OfferingHash    *Uint256       // Indexed.
	OpenBlockNumber *Uint256
	Balance         *Uint192
}

// NewChannelCloseRequestedEvent creates event of type ChannelCloseRequestedEvent.
// Please see contract implementation for the details.
func NewChannelCloseRequestedEvent(topics [4]string, hexData string) (*ChannelCloseRequestedEvent, error) {
	var err error
	e := &ChannelCloseRequestedEvent{}

	err = validateTopics(topics[:])
	err = checkEventDigest(topics[0], e.Digest(), err)

	e.Agent, err = parseAddress(topics[1], err)
	e.Client, err = parseAddress(topics[2], err)
	e.OfferingHash, err = parseTopicAsUint256(topics[3], err)

	e.OpenBlockNumber, err = parseDataFieldAsUint256(hexData, 0, err)
	e.Balance, err = parseDataFieldAsUint192(hexData, 1, err)
	return e, err
}

// Digest returns keccak512 hash sum of the event signature.
func (e *ChannelCloseRequestedEvent) Digest() string {
	return EthChannelCloseRequested
}

// OfferingCreatedEvent implements wrapper for contract event.
// Please see contract implementation for the details.
type OfferingCreatedEvent struct {
	Agent         common.Address // Indexed.
	OfferingHash  *Uint256       // Indexed.
	MinDeposit    *Uint192
	CurrentSupply *Uint256
}

// NewOfferingCreatedEvent creates event of type OfferingCreatedEvent.
// Please see contract implementation for the details.
func NewOfferingCreatedEvent(topics [4]string, hexData string) (*OfferingCreatedEvent, error) {
	var err error
	e := &OfferingCreatedEvent{}

	err = validateTopics(topics[:])
	err = checkEventDigest(topics[0], e.Digest(), err)

	e.Agent, err = parseAddress(topics[1], err)
	e.OfferingHash, err = parseTopicAsUint256(topics[2], err)
	e.MinDeposit, err = parseTopicAsUint192(topics[3], err)

	e.CurrentSupply, err = parseDataFieldAsUint256(hexData, 0, err)
	return e, err
}

// Digest returns keccak512 hash sum of the event signature.
func (e *OfferingCreatedEvent) Digest() string {
	return EthOfferingCreated
}

// OfferingDeletedEvent implements wrapper for contract event.
// Please see contract implementation for the details.
type OfferingDeletedEvent struct {
	Agent        common.Address // Indexed.
	OfferingHash *Uint256       // Indexed.
}

// NewOfferingDeletedEvent creates event of type OfferingDeletedEvent..
// Please see contract implementation for the details.
func NewOfferingDeletedEvent(topics [3]string) (*OfferingDeletedEvent, error) {
	var err error
	e := &OfferingDeletedEvent{}

	err = validateTopics(topics[:])
	err = checkEventDigest(topics[0], e.Digest(), err)

	e.Agent, err = parseAddress(topics[1], err)
	e.OfferingHash, err = parseTopicAsUint256(topics[2], err)
	return e, err
}

// Digest returns keccak512 hash sum of the event signature.
func (e *OfferingDeletedEvent) Digest() string {
	return EthOfferingDeleted
}

// OfferingEndpointEvent implements wrapper for contract event.
// Please see contract implementation for the details.
type OfferingEndpointEvent struct {
	Agent           common.Address // Indexed.
	Client          common.Address // Indexed.
	OfferingHash    *Uint256       // Indexed.
	OpenBlockNumber *Uint256
	EndpointHash    *Uint256
}

// NewOfferingEndpointEvent creates event of type OfferingEndpointEvent.
// Please see contract implementation for the details.
func NewOfferingEndpointEvent(topics [4]string, hexData string) (*OfferingEndpointEvent, error) {
	var err error
	e := &OfferingEndpointEvent{}

	err = validateTopics(topics[:])
	err = checkEventDigest(topics[0], e.Digest(), err)

	e.Agent, err = parseAddress(topics[1], err)
	e.Client, err = parseAddress(topics[2], err)
	e.OfferingHash, err = parseTopicAsUint256(topics[3], err)

	e.OpenBlockNumber, err = parseDataFieldAsUint256(hexData, 0, err)
	e.EndpointHash, err = parseDataFieldAsUint256(hexData, 1, err)
	return e, err
}

// Digest returns keccak512 hash sum of the event signature.
func (e *OfferingEndpointEvent) Digest() string {
	return EthOfferingEndpoint
}

// OfferingPoppedUpEvent implements wrapper for contract event.
// Please see contract implementation for the details.
type OfferingPoppedUpEvent struct {
	Agent        common.Address // Indexed.
	OfferingHash *Uint256       // Indexed.
}

// NewOfferingPoppedUpEvent creates event of type OfferingPoppedUpEvent.
// Please see contract implementation for the details.
func NewOfferingPoppedUpEvent(topics [3]string) (*OfferingPoppedUpEvent, error) {
	var err error
	e := &OfferingPoppedUpEvent{}

	err = validateTopics(topics[:])
	err = checkEventDigest(topics[0], e.Digest(), err)

	e.Agent, err = parseAddress(topics[1], err)
	e.OfferingHash, err = parseTopicAsUint256(topics[2], err)
	return e, err
}

// Digest returns keccak512 hash sum of the event signature.
func (e *OfferingPoppedUpEvent) Digest() string {
	return EthOfferingPoppedUp
}

// CooperativeChannelCloseEvent implements wrapper for contract event.
// Please see contract implementation for the details.
type CooperativeChannelCloseEvent struct {
	Agent           common.Address // Indexed.
	Client          common.Address // Indexed.
	OfferingHash    *Uint256       // Indexed.
	OpenBlockNumber *Uint256
	Balance         *Uint192
}

// NewCooperativeChannelCloseEvent creates event of type CooperativeChannelCloseEvent.
// Please see contract implementation for the details.
func NewCooperativeChannelCloseEvent(topics [4]string, hexData string) (*CooperativeChannelCloseEvent, error) {
	var err error
	e := &CooperativeChannelCloseEvent{}

	err = validateTopics(topics[:])
	err = checkEventDigest(topics[0], e.Digest(), err)

	e.Agent, err = parseAddress(topics[1], err)
	e.Client, err = parseAddress(topics[2], err)
	e.OfferingHash, err = parseTopicAsUint256(topics[3], err)

	e.OpenBlockNumber, err = parseDataFieldAsUint256(hexData, 0, err)
	e.Balance, err = parseDataFieldAsUint192(hexData, 1, err)
	return e, err
}

// Digest returns keccak512 hash sum of the event signature.
func (e *CooperativeChannelCloseEvent) Digest() string {
	return EthCooperativeChannelClose
}

// UncooperativeChannelCloseEvent implements wrapper for contract event.
// Please see contract implementation for the details.
type UncooperativeChannelCloseEvent struct {
	Agent           common.Address // Indexed.
	Client          common.Address // Indexed.
	OfferingHash    *Uint256       // Indexed.
	OpenBlockNumber *Uint256
	Balance         *Uint192
}

// NewUnCooperativeChannelCloseEvent creates event of type NewUnCooperativeChannelCloseEvent.
// Please see contract implementation for the details.
func NewUnCooperativeChannelCloseEvent(topics [4]string, hexData string) (*UncooperativeChannelCloseEvent, error) {
	var err error
	e := &UncooperativeChannelCloseEvent{}

	err = validateTopics(topics[:])
	err = checkEventDigest(topics[0], e.Digest(), err)

	e.Agent, err = parseAddress(topics[1], err)
	e.Client, err = parseAddress(topics[2], err)
	e.OfferingHash, err = parseTopicAsUint256(topics[3], err)

	e.OpenBlockNumber, err = parseDataFieldAsUint256(hexData, 0, err)
	e.Balance, err = parseDataFieldAsUint192(hexData, 1, err)
	return e, err
}

// Digest returns keccak512 hash sum of the event signature.
func (e *UncooperativeChannelCloseEvent) Digest() string {
	return EthUncooperativeChannelClose
}

func errorUnexpectedEventType(receivedDigest, expectedDigest string) error {
	return fmt.Errorf("unexpected event type occurred: %s, but %s is expected", receivedDigest, expectedDigest)
}

func validateTopics(topics []string) error {
	for _, topic := range topics {
		if len(topic) != 66 { // "0x" + 64 bytes.
			return errors.New("Invalid topic occurred: " + topic)
		}

		if topic[:2] != "0x" {
			return errors.New("Invalid topic occurred: " + topic)
		}
	}

	return nil
}

func checkEventDigest(topic string, expectedDigest string, err error) error {
	if err != nil {
		return err
	}

	digestHex := topicToHex(topic)
	if digestHex != expectedDigest {
		return errorUnexpectedEventType(digestHex, expectedDigest)
	}

	return nil
}

func parseAddress(topic string, inErr error) (addr common.Address, err error) {
	if inErr != nil {
		err = inErr
		return
	}

	return NewAddress(toAddressHex(topicToHex(topic)))
}

func parseTopicAsUint256(topic string, err error) (*Uint256, error) {
	if err != nil {
		return nil, err
	}

	return NewUint256(topic)
}

func parseTopicAsUint192(topic string, err error) (*Uint192, error) {
	if err != nil {
		return nil, err
	}

	return NewUint192(get192BitsDataField(topic, 0))
}

func parseDataFieldAsUint256(hexData string, offset uint8, err error) (*Uint256, error) {
	if err != nil {
		return nil, err
	}

	return NewUint256(get256BitsDataField(hexData, offset))
}

func parseDataFieldAsUint192(hexData string, offset uint8, err error) (*Uint192, error) {
	if err != nil {
		return nil, err
	}

	return NewUint192(get192BitsDataField(hexData, offset))
}

func topicToHex(topic string) string {
	if len(topic) <= 2 {
		return ""
	}
	return topic[2:]
}

func toAddressHex(hex string) string {
	if len(hex) <= 24 {
		return ""
	}
	return hex[24:]
}

func get256BitsDataField(hexData string, offset uint8) string {
	offsetFrom := 2 + (offset * 64) // skipping "0x"
	offsetTo := offsetFrom + 64
	if len(hexData) < int(offsetTo) {
		return ""
	}

	return "0x" + hexData[offsetFrom:offsetTo]
}

func get192BitsDataField(hexData string, offset uint8) string {
	dataField := get256BitsDataField(hexData, offset)
	if len(dataField) < 18 {
		return ""
	}

	return dataField[18:]
}
