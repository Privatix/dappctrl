package ept

import (
	"sync"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/log"
)

// Config is a configuration for Endpoint Message Service.
type Config struct {
	Timeout uint // In milliseconds.
}

type result struct {
	msg *Message
	err error
}

type req struct {
	done      chan bool
	channelID string
	callback  chan *result
}

type obj struct {
	ch    data.Channel
	offer data.Offering
	prod  data.Product
	tmpl  data.Template
}

// Message structure for Endpoint Message.
type Message struct {
	TemplateHash           data.HexString    `json:"templateHash"`
	Username               string            `json:"username"`
	Password               string            `json:"password"`
	PaymentReceiverAddress string            `json:"paymentReceiverAddress"`
	ServiceEndpointAddress string            `json:"serviceEndpointAddress"`
	AdditionalParams       map[string]string `json:"additionalParams"`
}

// Service for generation Endpoint Message.
type Service struct {
	db      *reform.DB
	logger  log.Logger
	msgChan chan *req
	payAddr string
	timeout time.Duration
}

// NewConfig creates a default Endpoint Message Service configuration.
func NewConfig() *Config {
	return &Config{1}
}

func newResult(tpl *Message, err error) *result {
	return &result{tpl, err}
}

// New function for initialize the service for generating
// the Endpoint Message.
func New(db *reform.DB, logger log.Logger, payAddr string,
	timeout uint) (*Service, error) {
	return &Service{
		db:      db,
		msgChan: make(chan *req),
		payAddr: payAddr,
		timeout: time.Duration(timeout) * time.Millisecond,
		logger:  logger.Add("type", "messages/ept.Service"),
	}, nil
}

// EndpointMessage returns the endpoint message object.
func (s *Service) EndpointMessage(channelID string) (*Message, error) {
	c := make(chan *result)
	done := make(chan bool)

	req := &req{channelID: channelID, callback: c, done: done}

	go s.processing(req)

	select {
	case result := <-c:
		return result.msg, result.err
	case <-time.After(s.timeout):
		close(done)
		return nil, ErrTimeOut
	}
}

func (s *Service) objects(done chan bool, errC chan error,
	objCh chan *obj, channelID string) {
	defer close(objCh)

	var ch data.Channel
	var offer data.Offering
	var prod data.Product
	var tmpl data.Template

	localErr := make(chan error)
	terminate := make(chan bool)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		exit := false

		go s.find(done, localErr, &ch, channelID)

		select {
		case <-done:
			terminate <- true
			exit = true
		case err := <-localErr:
			if err != nil {
				s.logger.Add("channel",
					channelID).Error(err.Error())
				errC <- err
				terminate <- true
				exit = true
			}
		}
		if exit {
			return
		}

		go s.find(done, localErr, &offer, ch.Offering)

		select {
		case <-done:
			terminate <- true
			exit = true
		case err := <-localErr:
			if err != nil {
				s.logger.Add("offering",
					ch.Offering).Error(err.Error())
				errC <- err
				terminate <- true
				exit = true
			}
		}

		if exit {
			return
		}

		go s.find(done, localErr, &prod, offer.Product)

		select {
		case <-done:
			terminate <- true
		case err := <-localErr:
			if err != nil {
				s.logger.Add("product",
					offer.Product).Error(err.Error())
				errC <- err
				terminate <- true
			}
		}

		go s.find(done, localErr, &tmpl, prod.OfferAccessID)

		select {
		case <-done:
			terminate <- true
		case err := <-localErr:
			if err != nil {
				s.logger.Add("offerAccessID",
					prod.OfferAccessID).Error(err.Error())
				errC <- err
				terminate <- true
			}
		}
	}()
	wg.Wait()

	select {
	case <-done:
	case <-terminate:
	case objCh <- &obj{prod: prod, offer: offer, ch: ch, tmpl: tmpl}:
	}
}

func (s *Service) processing(req *req) {
	resp := make(chan *result)

	go func() {
		var o *obj
		var m *Message

		exit := false

		errC := make(chan error)
		objCh := make(chan *obj)
		msgCh := make(chan *Message)

		go s.objects(req.done, errC, objCh, req.channelID)

		select {
		case <-req.done:
			exit = true
		case err := <-errC:
			resp <- newResult(nil, err)
			exit = true
		case o = <-objCh:
		}

		if exit {
			return
		}

		go s.genMsg(req.done, errC, o, msgCh)

		select {
		case <-req.done:
			exit = true
		case err := <-errC:
			resp <- newResult(nil, err)
			exit = true
		case m = <-msgCh:
		}

		if exit {
			return
		}

		if !validMsg(o.tmpl.Raw, *m) {
			select {
			case <-req.done:
			case resp <- newResult(nil, ErrInvalidFormat):
			}
			return

		}

		select {
		case <-req.done:
		case resp <- newResult(m, nil):
		}
	}()

	select {
	case <-req.done:
	case req.callback <- <-resp:
	}
}

func (s *Service) genMsg(done chan bool, errC chan error, o *obj,
	msgCh chan *Message) {
	conf, err := s.config(o.prod.Config)
	if err != nil {
		select {
		case <-done:
		case errC <- err:
		}
		return
	}

	msg, err := fillMsg(o, s.payAddr, conf)
	if err != nil {
		select {
		case <-done:
		case errC <- err:
		}
		return
	}
	select {
	case <-done:
	case msgCh <- msg:
	}
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
