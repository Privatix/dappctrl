package message

// Errors
var (
	ErrInput           = "one or more input parameters is wrong"
	ErrReceiver        = "receiver format is wrong"
	ErrEndpoint        = "endpoint format is wrong"
	ErrFilePathIsEmpty = "filePath is empty"
	ErrParsingLines    = "parsing lines from the file"
	ErrCertNotExist    = "certificate not exist in the config file"
	ErrCertCanNotRead  = "cannot read certificate file"
	ErrCertIsNull      = "the certificate file does not contain CERTIFICATE"
)
