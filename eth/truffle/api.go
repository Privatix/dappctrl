package truffle

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

// API communicates with truffle test api.
type API string

// FetchPSCAddress returns psc contract address hex from truffle.
func (api *API) FetchPSCAddress() string {
	return api.fetchAddr("/getPSC")
}

// FetchPTCAddress returns ptc address hex from truffle.
func (api *API) FetchPTCAddress() string {
	return api.fetchAddr("/getPrix")
}

func (api *API) fetchAddr(path string) string {
	data := map[string]interface{}{}
	api.fetchFromTruffle(path, &data)
	return data["contract"].(map[string]interface{})["address"].(string)
}

var truffleFetchResults = map[string][]byte{}

func (api *API) fetchFromTruffle(path string, v interface{}) {
	reply, ok := truffleFetchResults[path]
	if !ok {
		response, err := http.Get(string(*api) + path)
		if err != nil || response.StatusCode != http.StatusOK {
			log.Fatal("can't fetch, check test environment")
		}
		reply, err = ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatalf("can't read truffle response: %v", err)
		}
	}

	if err := json.Unmarshal(reply, v); err != nil {
		log.Fatalf("can't parse truffle response: %v", err)
	}

	truffleFetchResults[path] = reply
}
