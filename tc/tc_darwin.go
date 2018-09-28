package tc

// Config is a traffic control configuration.
type Config struct {
}

// NewConfig creates a default configuration.
func NewConfig() *Config {
	return &Config{}
}

// SetRateLimit sets a rate limit for a given client IP address on a given
// network interface.
func (tc *TrafficControl) SetRateLimit(
	clientIP string, upMbits, downMbits float32) error {
	return nil
}

// UnsetRateLimit removes a rate limit for a given client IP address on a given
// network interface.
func (tc *TrafficControl) UnsetRateLimit(clientIP string) error {
	return nil
}
