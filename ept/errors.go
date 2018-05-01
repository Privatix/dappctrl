package ept

import "github.com/pkg/errors"

// Endpoint Message Template errors
var (
	ErrInput           = errors.New("one or more input parameters is wrong")
	ErrReceiver        = errors.New("receiver format is wrong")
	ErrEndpoint        = errors.New("endpoint format is wrong")
	ErrHash            = errors.New("hash format is wrong")
	ErrFilePathIsEmpty = errors.New("filePath is empty")
	ErrParsingLines    = errors.New("parsing lines from the file")
	ErrCertNotExist    = errors.New("certificate not exist in the config file")
	ErrCertCanNotRead  = errors.New("cannot read certificate file")
	ErrCertIsNull      = errors.New("the certificate file does not contain CERTIFICATE")
)
