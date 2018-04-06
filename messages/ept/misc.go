package ept

import (
	"encoding/json"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

func validMsg(schema []byte, msg Message) bool {
	sch := gojsonschema.NewBytesLoader(schema)
	loader := gojsonschema.NewGoLoader(msg)

	result, err := gojsonschema.Validate(sch, loader)
	if err != nil || !result.Valid() || len(result.Errors()) != 0 {
		fmt.Printf("%+v\n", result.Errors())
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
		TemplateHash:           o.tmpl.Hash,
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
