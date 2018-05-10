package ept

import (
	"encoding/json"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/pay"
	"github.com/privatix/dappctrl/util"
)

type result struct {
	tplID string
	err   error
}

type req struct {
	done      chan bool
	channelID string
	callback  chan *result
}

// Service for generation Endpoint Message Template
type Service struct {
	db      *reform.DB
	msgChan chan *req
	payAddr string
}

func newResult(tplID string, err error) *result {
	return &result{tplID, err}
}

// New function for initialize the service for generating
// the Endpoint Message Template
func New(db *reform.DB, payConfig *pay.Config) *Service {
	return &Service{db, make(chan *req), payConfig.Addr}
}

// EndpointMessageTemplate creates a new endpoint message template
// in the database, returns the template ID
func (s *Service) EndpointMessageTemplate(channelID string,
	timeout time.Duration) (string, error) {
	c := make(chan *result)
	done := make(chan bool)

	req := &req{channelID: channelID, callback: c, done: done}

	go s.processing(req)

	select {
	case result := <-c:
		return result.tplID, result.err
	case <-time.After(timeout):
		close(done)
		return "", ErrTimeOut
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
				resp <- newResult("", err)
				return
			}
		}

		go s.find(req.done, errC, &offer, ch.Offering)

		select {
		case <-req.done:
			return
		case err := <-errC:
			if err != nil {
				resp <- newResult("", err)
				return
			}
		}

		go s.find(req.done, errC, &prod, offer.Product)

		select {
		case <-req.done:
			return
		case err := <-errC:
			if err != nil {
				resp <- newResult("", err)
				return
			}
		}

		if prod.OfferAccessID == nil {
			resp <- newResult("", ErrOfferAccessID)
			return
		}

		if !isValidJSON(prod.Config, &map[string]string{}) {
			resp <- newResult("", ErrInvalidProdConf)
			return
		}

		endMsgTemp := &data.EndpointMessageTemplate{
			ID:                     util.NewUUID(),
			TemplateHash:           *prod.OfferAccessID,
			Username:               &ch.ID,
			Password:               &ch.Password,
			PaymentReceiverAddress: s.payAddr,
			ServiceEndpointAddress: "",
			AdditionalParams:       prod.Config,
		}

		go s.insert(req.done, errC, endMsgTemp)

		select {
		case <-req.done:
			return
		case err := <-errC:
			if err != nil {
				resp <- newResult("", err)
				return
			}

			resp <- newResult(endMsgTemp.ID, nil)
		}
	}()

	select {
	case <-req.done:
		return
	case req.callback <- <-resp:
	}
}

func isValidJSON(data []byte, v interface{}) bool {
	if json.Unmarshal(data, v) != nil {
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

func (s *Service) insert(done chan bool, errC chan error,
	record reform.Struct) {
	err := s.db.Insert(record)

	select {
	case <-done:
		return
	case errC <- err:
	}
}
