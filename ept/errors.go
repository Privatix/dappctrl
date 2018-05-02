package ept

import "github.com/pkg/errors"

// Endpoint EndpointMessage Template errors
var (
	ErrInput           = errors.New("one or more input parameters is wrong")
	ErrReceiver        = errors.New("receiver format is wrong")
	ErrEndpoint        = errors.New("endpoint format is wrong")
	ErrHash            = errors.New("hash format is wrong")
	ErrFilePathIsEmpty = errors.New("filePath is empty")
	ErrCertNotExist    = errors.New("certificate not exist in the config file")
	ErrCertCanNotRead  = errors.New("cannot read certificate file")
	ErrCertNotFound    = errors.New("certificate can not be found in the specified path")
	ErrCertIsNull      = errors.New("the certificate file does not contain CERTIFICATE")
)
