package somc

// SOMC codes are 2^0, 2^1, 2^2 etc.
const (
	torCode    uint8 = 1
	directCode uint8 = 2
)

// TorAgentConfig is agent's config to run SOMC over Tor.
type TorAgentConfig struct {
	Hostname string
}

// NewTorAgentConfig creates config with default values.
func NewTorAgentConfig() *TorAgentConfig {
	return &TorAgentConfig{
		Hostname: "",
	}
}

// DirectAgentConfig is agent's config to directly serve SOMC.
type DirectAgentConfig struct {
	Addr string
}

// NewDirectAgentConig creates config with default values.
func NewDirectAgentConfig() *DirectAgentConfig {
	return &DirectAgentConfig{
		Addr: "",
	}
}

// TorClientConfig is client's config to connect to SOMCs on Tor.
type TorClientConfig struct {
	Socks uint
}

// NewTorClientConfig creates config with default values.
func NewTorClientConfig() *TorClientConfig {
	return &TorClientConfig{
		Socks: 9050,
	}
}
