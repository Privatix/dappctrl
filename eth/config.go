package eth

// Config is a configuration for Ethereum client.
type Config struct {
	Contract struct {
		PTCAddrHex string
		PSCAddrHex string
	}
	GethURL string
	Timeout uint64
}

// NewConfig creates a default Ethereum client configuration.
func NewConfig() *Config {
	return &Config{
		Timeout: 10,
	}
}
