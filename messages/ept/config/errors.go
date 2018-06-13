package config

import "github.com/pkg/errors"

// Config errors
var (
	ErrInput           = errors.New("one or more input parameters is wrong")
	ErrFilePathIsEmpty = errors.New("filePath is empty")
	ErrCertNotExist    = errors.New("certificate doesn't exist in the config file")
	ErrCertCanNotRead  = errors.New("cannot read certificate file")
	ErrCertNotFound    = errors.New("certificate can not be found in the specified path")
	ErrCancel          = errors.New("canceled sending product config to session server")
)
