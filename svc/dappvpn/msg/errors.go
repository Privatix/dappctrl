package msg

import (
	"github.com/privatix/dappctrl/util/errors"
)

// Errors.
const (
	// CRC16("github.com/privatix/dappctrl/svc/dappvpn/msg") = 0xE21C
	ErrServiceEndpointAddr errors.Error = 0xE21C<<8 + iota
	ErrDecodeParams
	ErrCreateDir
	ErrCreateConfig
	ErrParseConfigTemplate
	ErrCombineConfigTemplate
	ErrCreateAccessFile
	ErrReadConfigFile
	ErrReadLineFromConfigFile
	ErrReadCAFile
	ErrCAFormat
	ErrContextIsDone
)

var errMsgs = errors.Messages{
	ErrServiceEndpointAddr: "service endpoint address" +
		" of an invalid format",
	ErrDecodeParams:        "failed to decode additional params from JSON",
	ErrCreateDir:           "failed to make directory",
	ErrCreateConfig:        "failed to create client configuration file",
	ErrParseConfigTemplate: "failed to parse template for config",
	ErrCombineConfigTemplate: "failed to combine template and " +
		"config params",
	ErrCreateAccessFile: "failed to create client access file",
	ErrReadConfigFile:   "failed to read vpn configuration file",
	ErrReadLineFromConfigFile: "failed to read line from vpn " +
		"configuration file",
	ErrReadCAFile: "failed to read Certificate Authority file",
	ErrCAFormat: "certificate authority can not be found " +
		"in the specified path",
	ErrContextIsDone: "context is done",
}

func init() { errors.InjectMessages(errMsgs) }
