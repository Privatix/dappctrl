package ept

import "github.com/pkg/errors"

// Endpoint EndpointMessage Template errors
var (
	ErrTimeOut         = errors.New("timeout")
	ErrInvalidProdConf = errors.New("invalid product config")
	ErrOfferAccessID   = errors.New("product field offer_access_id is null")
)
