package config

import "github.com/pkg/errors"

// Endpoint EndpointMessage Template errors
var (
	ErrInput           = errors.New("one or more input parameters is wrong")
	ErrFilePathIsEmpty = errors.New("filePath is empty")
	ErrCertNotExist    = errors.New("certificate not exist in the config file")
	ErrCertCanNotRead  = errors.New("cannot read certificate file")
	ErrCertNotFound    = errors.New("certificate can not be found in the specified path")
)
