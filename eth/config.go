package eth

// PSCPeriods psc periods.
type PSCPeriods struct {
	PopUp     uint32
	Challenge uint32
	Remove    uint32
}

// Config is a configuration for Ethereum client.
type Config struct {
	CheckTimeout uint64
	Contract     struct {
		PTCAddrHex string
		PSCAddrHex string
		Periods    *PSCPeriods
	}
	GethURL    string
	Timeout    uint64
	HTTPClient *httpClientConf
}

type httpClientConf struct {
	DialTimeout           uint64
	TLSHandshakeTimeout   uint64
	ResponseHeaderTimeout uint64
	RequestTimeout        uint64
	IdleConnTimeout       uint64
	KeepAliveTimeout      uint64
}

// NewConfig creates a default Ethereum client configuration.
func NewConfig() *Config {
	return &Config{
		CheckTimeout: 20,
		Timeout:      10,
		HTTPClient: &httpClientConf{
			DialTimeout:           5,
			TLSHandshakeTimeout:   2,
			ResponseHeaderTimeout: 8,
			RequestTimeout:        10,
			IdleConnTimeout:       30,
			KeepAliveTimeout:      60,
		},
	}
}
