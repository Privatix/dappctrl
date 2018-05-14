package ept

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/pay"
)

const (
	invalidChannel  = "invalid channel"
	invalidOffering = "invalid offering"
	invalidProduct  = "invalid product"
	invalidTemplate = "invalid template"
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

type obj struct {
	ch    data.Channel
	offer data.Offering
	prod  data.Product
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
func New(db *reform.DB, payConfig *pay.Config) (*Service, error) {
	if db == nil || payConfig == nil {
		return nil, ErrInput
	}

	return &Service{db, make(chan *req), payConfig.Addr}, nil
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

func (s *Service) objects(done chan bool, errC chan error,
	objCh chan *obj, channelID string) {
	defer close(objCh)

	var ch data.Channel
	var offer data.Offering
	var prod data.Product

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
				errC <- errWrapper(err, invalidChannel)
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
				errC <- errWrapper(err, invalidOffering)
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
			exit = true
		case err := <-localErr:
			if err != nil {
				errC <- errWrapper(err, invalidProduct)
				terminate <- true
				exit = true
			}
		}

		if exit {
			return
		}
	}()
	wg.Wait()

	select {
	case <-done:
	case <-terminate:
	case objCh <- &obj{prod: prod, offer: offer, ch: ch}:
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

		tempCh := make(chan data.Template)

		var temp data.Template

		go s.findTemp(req.done, errC, tempCh, *o.prod.OfferAccessID)

		select {
		case <-req.done:
			exit = true
		case temp = <-tempCh:
		case err := <-errC:
			resp <- newResult(nil, err)
			exit = true
		}

		if exit {
			return
		}

		if !validMsg(temp.Raw, *m) {
			select {
			case <-req.done:
				exit = true
			case resp <- newResult(nil, ErrInvalidFormat):
				exit = true
			}
			if exit {
				return
			}
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
	conf, err := config(o.prod.Config)
	if err != nil {
		select {
		case <-done:
		case errC <- err:
		}
		return
	}

	msg, err := fillMsg(o, s.payAddr, "", conf)
	if err != nil {
		select {
		case <-done:
		case errC <- errWrapper(err, invalidTemplate):
		}
		return
	}
	select {
	case <-done:
	case msgCh <- msg:
	}
}

func (s *Service) findTemp(done chan bool, errC chan error,
	tempCh chan data.Template, id string) {
	var temp data.Template

	localErrC := make(chan error)

	go s.find(done, localErrC, &temp, id)

	select {
	case <-done:
		return
	case err := <-localErrC:
		if err != nil {
			errC <- errWrapper(err, invalidTemplate)
		}
		tempCh <- temp
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

func validMsg(schema []byte, msg Message) bool {
	sch := gojsonschema.NewBytesLoader(schema)
	loader := gojsonschema.NewGoLoader(msg)

	result, err := gojsonschema.Validate(sch, loader)
	if err != nil || !result.Valid() || len(result.Errors()) != 0 {
		return false
	}
	return true
}

func fillMsg(o *obj, paymentReceiverAddress, serviceEndpointAddress string,
	conf map[string]string) (*Message, error) {

	if o.prod.OfferAccessID == nil {
		return nil, ErrProdOfferAccessID
	}

	return &Message{
		TemplateHash:           *o.prod.OfferAccessID,
		Username:               o.ch.ID,
		Password:               o.ch.Password,
		PaymentReceiverAddress: paymentReceiverAddress,
		ServiceEndpointAddress: serviceEndpointAddress,
		AdditionalParams:       conf,
	}, nil
}

func config(confByte []byte) (map[string]string, error) {
	var conf map[string]string

	if err := json.Unmarshal(confByte, &conf); err != nil {
		return nil, err
	}

	return conf, nil
}
