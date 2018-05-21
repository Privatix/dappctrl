package ept

import (
	"encoding/json"
	"math/rand"

	"github.com/sethvargo/go-password/password"
	"github.com/xeipuuv/gojsonschema"
)

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
		TemplateHash:           o.tmpl.Hash,
		Username:               o.ch.ID,
		Password:               genPass(),
		PaymentReceiverAddress: paymentReceiverAddress,
		ServiceEndpointAddress: serviceEndpointAddress,
		AdditionalParams:       conf,
	}, nil
}

func genPass() string {
	// Passing valid arguments, thus ignoring errors.
	// Password of length 12 with up to 5 digits and 0 symbols,
	// allowing no repeats.
	generated, _ := password.Generate(12, rand.Intn(5), 0, false, false)
	return generated
}

func config(confByte []byte) (map[string]string, error) {
	var conf map[string]string

	if err := json.Unmarshal(confByte, &conf); err != nil {
		return nil, err
	}

	return conf, nil
}
