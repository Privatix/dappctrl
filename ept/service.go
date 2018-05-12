package ept

import (
	"encoding/json"
	"time"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/pay"
)

type result struct {
	msg *Message
	err error
}

type req struct {
	done      chan bool
	channelID string
	callback  chan *result
}

// Message structure for Endpoint Message
type Message struct {
	TemplateHash           string            `json:"templateHash"`
	Username               string            `json:"username"`
	Password               string            `json:"password"`
	PaymentReceiverAddress string            `json:"paymentReceiverAddress"`
	ServiceEndpointAddress string            `json:"serviceEndpointAddress"`
	AdditionalParams       map[string]string `json:"additionalParams"`
}

// Service for generation Endpoint Message
type Service struct {
	db      *reform.DB
	msgChan chan *req
	payAddr string
}

func newResult(tpl *Message, err error) *result {
	return &result{tpl, err}
}

// New function for initialize the service for generating
// the Endpoint Message
func New(db *reform.DB, payConfig *pay.Config) *Service {
	return &Service{db, make(chan *req), payConfig.Addr}
}

// EndpointMessage returns the endpoint message object
func (s *Service) EndpointMessage(channelID string,
	timeout time.Duration) (*Message, error) {
	c := make(chan *result)
	done := make(chan bool)

	req := &req{channelID: channelID, callback: c, done: done}

	go s.processing(req)

	select {
	case result := <-c:
		return result.msg, result.err
	case <-time.After(timeout):
		close(done)
		return nil, ErrTimeOut
	}
}

func (s *Service) processing(req *req) {
	resp := make(chan *result)

	go func() {
		var ch data.Channel
		var offer data.Offering
		var prod data.Product

		errC := make(chan error)

		go s.find(req.done, errC, &ch, req.channelID)

		select {
		case <-req.done:
			return
		case err := <-errC:
			if err != nil {
				resp <- newResult(nil, err)
				return
			}
		}

		go s.find(req.done, errC, &offer, ch.Offering)

		select {
		case <-req.done:
			return
		case err := <-errC:
			if err != nil {
				resp <- newResult(nil, err)
				return
			}
		}

		go s.find(req.done, errC, &prod, offer.Product)

		select {
		case <-req.done:
			return
		case err := <-errC:
			if err != nil {
				resp <- newResult(nil, err)
				return
			}
		}

		var conf map[string]string

		if err := json.Unmarshal(prod.Config, &conf); err != nil {
			select {
			case <-req.done:
				return
			case resp <- newResult(nil, err):
				return
			}
		}

		if prod.OfferAccessID == nil {
			select {
			case <-req.done:
				return
			case resp <- newResult(nil, ErrInvalidFormat):
				return
			}
		}

		msg := Message{
			TemplateHash:           *prod.OfferAccessID,
			Username:               ch.ID,
			Password:               ch.Password,
			PaymentReceiverAddress: s.payAddr,
			ServiceEndpointAddress: "",
			AdditionalParams:       conf,
		}

		var temp data.Template

		go s.find(req.done, errC, &temp, *prod.OfferAccessID)

		select {
		case <-req.done:
			return
		case err := <-errC:
			if err != nil {
				resp <- newResult(nil, err)
				return
			}
		}

		if !valid(temp.Raw, msg) {
			select {
			case <-req.done:
				return
			case resp <- newResult(&msg, ErrInvalidFormat):
			}
		}

		select {
		case <-req.done:
			return
		case resp <- newResult(&msg, nil):
		}
	}()

	select {
	case <-req.done:
		return
	case req.callback <- <-resp:
	}
}

func valid(schema []byte, msg Message) bool {
	sch := gojsonschema.NewBytesLoader(schema)
	loader := gojsonschema.NewGoLoader(msg)

	result, err := gojsonschema.Validate(sch, loader)
	if err != nil || !result.Valid() || len(result.Errors()) != 0 {
		return false
	}
	return true
}

func (s *Service) find(done chan bool, errC chan error,
	record reform.Record, pk interface{}) {
	err := s.db.FindByPrimaryKeyTo(record, pk)

	select {
	case <-done:
		return
	case errC <- err:
	}
}
