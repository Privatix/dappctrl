package country

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// UndefinedCountry is default country code.
// This means that a country is not defined.
const UndefinedCountry = "ZZ"

// Config is the configuration for obtaining a country code.
type Config struct {
	Field   string
	Timeout uint64 // in milliseconds.
	// In a url template there should be an pattern {{ip}}.
	// Pattern {{ip}} will be replaced by a ip address of the agent.
	URLTemplate string
}

// NewConfig creates new configuration for obtaining a country code.
func NewConfig() *Config {
	return &Config{
		Timeout: 30,
	}
}

func do(ctx context.Context, req *http.Request) (*http.Response, error) {
	res, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			return nil, err
		}
	}
	return res, nil
}

// GetCountry returns country code by ip.
// Parses response in JSON format and returns a value of the field.
func GetCountry(timeout uint64, url, field string) (string, error) {
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Millisecond*time.Duration(timeout))
	defer cancel()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	res, err := do(ctx, req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	object := make(map[string]interface{})
	if err := json.NewDecoder(res.Body).Decode(&object); err != nil {
		return "", err
	}

	f, ok := object[field]
	if !ok {
		return "", ErrMissingRequiredField
	}

	country, ok := f.(string)
	if !ok {
		return "", ErrBadCountryValueType
	}

	return strings.TrimSpace(strings.ToUpper(country)), nil
}
