package srv

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// GetURL returns a server URL for a given path.
func GetURL(conf *Config, path string) string {
	var proto = "http"
	if conf.TLS != nil {
		proto += "s"
	}

	return proto + "://" + conf.Addr + path
}

// NewHTTPRequest creates a new HTTP request from a given server request.
func NewHTTPRequest(conf *Config, method,
	path string, req *Request) (*http.Request, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	return http.NewRequest(
		method, GetURL(conf, path), bytes.NewReader(data))
}

// Send sends an HTTP request and returns a server response.
func Send(req *http.Request) (*Response, error) {
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var resp2 Response
	if err = json.NewDecoder(resp.Body).Decode(&resp2); err != nil {
		return nil, err
	}

	return &resp2, nil
}
