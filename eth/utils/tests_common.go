package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

// FetchPSCAddress is utility method for fetching contract address in test environment.
func FetchPSCAddress() string {
	truffleAPI := GethEthereumConfig().TruffleAPI
	response, err := http.Get(truffleAPI.Interface() + "/getPSC")
	if err != nil || response.StatusCode != 200 {
		log.Fatal("Can't fetch PSC address. It seems that test environment is broken.")
	}

	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		log.Fatal("Can't read response body. It seems that test environment is broken.")
	}

	data := make(map[string]interface{})
	json.Unmarshal(body, &data)

	return data["contract"].(map[string]interface{})["address"].(string)
}

// FetchTestPrivateKey is utility method for fetching account private key in test environment.
func FetchTestPrivateKey() string {
	truffleAPI := GethEthereumConfig().TruffleAPI
	response, err := http.Get(truffleAPI.Interface() + "/getKeys")
	if err != nil || response.StatusCode != 200 {
		log.Fatal("Can't fetch private key. It seems that test environment is broken.")
	}

	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		log.Fatal("Can't read response body. It seems that test environment is broken.")
	}

	data := make([]interface{}, 0, 0)
	json.Unmarshal(body, &data)

	return data[0].(map[string]interface{})["privateKey"].(string)
}
