package lib

import (
	"errors"
	"fmt"
)

const (
	//
	// Contract events digests.
	// Please, see this article for the details:
	// https://codeburst.io/deep-dive-into-ethereum-logs-a8d2047c7371
	//

	EthDigestChannelCreated         = "a6153987181667023837aee39c3f1a702a16e5e146323ef10fb96844a526143c"
	EthDigestChannelToppedUp        = "392a992c1a7b756e553d8d97f43d59fafe79bc672808247debc077a6cdaba7b9"
	EthChannelCloseRequested        = "21ff66d79903f9d4ab6ab3c7c903af993e709be2ce2f4532d572925dea741cb1"
	EthOfferingCreated              = "49d573efb7cbb057727f6cadb4150ba6d5041c4fb55afe606508be636e158127"
	EthOfferingDeleted              = "21652905a07e2790c3a220d14394aee13681876bfbf38e658fa82ee5afe0c862"
	EthServiceOfferingEndpoint      = "00a7695de2bf4b4a523002334437d52e135b7a2a892d4471b5dd9005e5cd0681"
	EthServiceOfferingSupplyChanged = "1337b30376128e64c2ffd4e95d4c900b4ab42af11202b328722020216eeb46df"
	EthServiceOfferingPoppedUp      = "c8404827c21b5491a6c3dc0881307e47bfa40c3baf3d607c2d14f6bc808d4bfb"
	EthCooperativeChannelClose      = "56a4dfc7b9f93649d9142c7bef0a429decf8d3be895a3180c67a76a18d79f4ab"
	EthUncooperativeChannelClose    = "8a79bd24ee9bcfd977d6fc685befa8775c8a933f0abe82ab73b716cf419f968e"
)

// Base interface for all events, that are expected to be received from the ethereum block-chain.
type Event interface {
	Digest() string
}

// todo: here and elsewhere: To clearly deliver the meaning, I would suggest to change naming style here and in other place from NewEventServiceOfferingCreated to NewServiceOfferingCreatedEvent
type EventChannelCreated struct {
	Client            *Address // Indexed.
	Agent             *Address // Indexed.
	OfferingHash      *Uint256 // Indexed.
	Deposit           *Uint192
	AuthenticatedHash *Uint256
}

func NewEventChannelCreated(topics [4]string, hexData string) (*EventChannelCreated, error) {
	var err error
	e := &EventChannelCreated{}

	err = validateTopics(topics[:])
	err = checkEventDigest(topics[0], e.Digest(), err)

	e.Client, err = parseAddress(topics[1], err)
	e.Agent, err = parseAddress(topics[2], err)
	e.OfferingHash, err = parseTopicAsUint256(topics[3], err)

	e.Deposit, err = parseDataFieldAsUint192(hexData, 0, err)
	e.AuthenticatedHash, err = parseDataFieldAsUint256(hexData, 1, err)
	return e, err
}

func (e *EventChannelCreated) Digest() string {
	return EthDigestChannelCreated
}

type EventChannelToppedUp struct {
	Client          *Address // Indexed.
	Agent           *Address // Indexed.
	OpenBlockNumber *Uint256 // Indexed.
	OfferingHash    *Uint256
	AddedDeposit    *Uint192
}

func NewEventChannelToppedUp(topics [4]string, hexData string) (*EventChannelToppedUp, error) {
	var err error
	e := &EventChannelToppedUp{}

	err = validateTopics(topics[:])
	err = checkEventDigest(topics[0], e.Digest(), err)

	e.Client, err = parseAddress(topics[1], err)
	e.Agent, err = parseAddress(topics[2], err)
	e.OpenBlockNumber, err = parseTopicAsUint256(topics[3], err)

	e.OfferingHash, err = parseDataFieldAsUint256(hexData, 0, err)
	e.AddedDeposit, err = parseDataFieldAsUint192(hexData, 1, err)
	return e, err
}

func (e *EventChannelToppedUp) Digest() string {
	return EthDigestChannelToppedUp
}

type EventChannelCloseRequested struct {
	Client          *Address // Indexed.
	Agent           *Address // Indexed.
	OpenBlockNumber *Uint256 // Indexed.
	OfferingHash    *Uint256
	Balance         *Uint192
}

func NewEventChannelCloseRequested(topics [4]string, hexData string) (*EventChannelCloseRequested, error) {
	return nil, errors.New("not implemented") // todo: fix this when API would be fixed
}

func (e *EventChannelCloseRequested) Digest() string {
	return EthChannelCloseRequested
}

type EventOfferingCreated struct {
	Agent         *Address // Indexed.
	OfferingHash  *Uint256 // Indexed.
	MinDeposit    *Uint192
	CurrentSupply *Uint192
}

func NewEventServiceOfferingCreated(topics [3]string, hexData string) (*EventOfferingCreated, error) {
	var err error
	e := &EventOfferingCreated{}

	err = validateTopics(topics[:])
	err = checkEventDigest(topics[0], e.Digest(), err)

	e.Agent, err = parseAddress(topics[1], err)
	e.OfferingHash, err = parseTopicAsUint256(topics[2], err)

	e.MinDeposit, err = parseDataFieldAsUint192(hexData, 0, err)
	e.CurrentSupply, err = parseDataFieldAsUint192(hexData, 1, err)
	return e, err
}

func (e *EventOfferingCreated) Digest() string {
	return EthOfferingCreated
}

type EventOfferingDeleted struct {
	OfferingHash *Uint256 // Indexed.
}

func NewEventServiceOfferingDeleted(topics [2]string) (*EventOfferingDeleted, error) {
	var err error
	e := &EventOfferingDeleted{}

	err = validateTopics(topics[:])
	err = checkEventDigest(topics[0], e.Digest(), err)

	e.OfferingHash, err = parseTopicAsUint256(topics[1], err)
	return e, err
}

func (e *EventOfferingDeleted) Digest() string {
	return EthOfferingDeleted
}

type EventOfferingEndpoint struct {
	Client          *Address // Indexed.
	OfferingHash    *Uint256 // Indexed.
	OpenBlockNumber *Uint256 // Indexed.
	EndpointHash    *Uint256
}

func NewEventServiceOfferingEndpoint(topics [4]string, hexData string) (*EventOfferingEndpoint, error) {
	var err error
	e := &EventOfferingEndpoint{}

	err = validateTopics(topics[:])
	err = checkEventDigest(topics[0], e.Digest(), err)

	e.Client, err = parseAddress(topics[1], err)
	e.OfferingHash, err = parseTopicAsUint256(topics[2], err)
	e.OpenBlockNumber, err = parseTopicAsUint256(topics[3], err)

	e.EndpointHash, err = parseDataFieldAsUint256(hexData, 0, err)
	return e, err
}

func (e *EventOfferingEndpoint) Digest() string {
	return EthServiceOfferingEndpoint
}

type EventOfferingSupplyChanged struct {
	OfferingHash  *Uint256 // Indexed.
	CurrentSupply *Uint192
}

func NewEventServiceOfferingSupplyChanged(topics [2]string, hexData string) (*EventOfferingSupplyChanged, error) {
	var err error
	e := &EventOfferingSupplyChanged{}

	err = validateTopics(topics[:])
	err = checkEventDigest(topics[0], e.Digest(), err)

	e.OfferingHash, err = parseTopicAsUint256(topics[1], err)

	e.CurrentSupply, err = parseDataFieldAsUint192(hexData, 0, err)
	return e, err
}

func (e *EventOfferingSupplyChanged) Digest() string {
	return EthServiceOfferingSupplyChanged
}

type EventOfferingPoppedUp struct {
	OfferingHash *Uint256 // Indexed.
}

func NewEventServiceOfferingPoppedUp(topics [2]string) (*EventOfferingPoppedUp, error) {
	var err error
	e := &EventOfferingPoppedUp{}

	err = validateTopics(topics[:])
	err = checkEventDigest(topics[0], e.Digest(), err)

	e.OfferingHash, err = parseTopicAsUint256(topics[1], err)
	return e, err
}

func (e *EventOfferingPoppedUp) Digest() string {
	return EthServiceOfferingPoppedUp
}

type EventCooperativeChannelClose struct {
	Client          *Address // Indexed.
	Agent           *Address // Indexed.
	OpenBlockNumber *Uint256 // Indexed.
	OfferingHash    *Uint256
	Balance         *Uint192
}

func NewEventCooperativeChannelClose(topics [4]string, hexData string) (*EventCooperativeChannelClose, error) {
	var err error
	e := &EventCooperativeChannelClose{}

	err = validateTopics(topics[:])
	err = checkEventDigest(topics[0], e.Digest(), err)

	e.Client, err = parseAddress(topics[1], err)
	e.Agent, err = parseAddress(topics[2], err)
	e.OpenBlockNumber, err = parseTopicAsUint256(topics[3], err)

	e.OfferingHash, err = parseDataFieldAsUint256(hexData, 0, err)
	e.Balance, err = parseDataFieldAsUint192(hexData, 1, err)
	return e, err
}

func (e *EventCooperativeChannelClose) Digest() string {
	return EthCooperativeChannelClose
}

type EventUncooperativeChannelClose struct {
	Client          *Address // Indexed.
	Agent           *Address // Indexed.
	OpenBlockNumber *Uint256 // Indexed.
	OfferingHash    *Uint256
	Balance         *Uint192
}

func NewEventUnCooperativeChannelClose(topics [4]string, hexData string) (*EventUncooperativeChannelClose, error) {
	var err error
	e := &EventUncooperativeChannelClose{}

	err = validateTopics(topics[:])
	err = checkEventDigest(topics[0], e.Digest(), err)

	e.Client, err = parseAddress(topics[1], err)
	e.Agent, err = parseAddress(topics[2], err)
	e.OpenBlockNumber, err = parseTopicAsUint256(topics[3], err)

	e.OfferingHash, err = parseDataFieldAsUint256(hexData, 0, err)
	e.Balance, err = parseDataFieldAsUint192(hexData, 1, err)
	return e, err
}

func (e *EventUncooperativeChannelClose) Digest() string {
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

func parseAddress(topic string, err error) (*Address, error) {
	if err != nil {
		return nil, err
	}

	return NewAddress(toAddressHex(topicToHex(topic)))
}

func parseTopicAsUint256(topic string, err error) (*Uint256, error) {
	if err != nil {
		return nil, err
	}

	return NewUint256(topic)
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
