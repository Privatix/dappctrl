// +build !noethtest

package truffle

// GetContractTransactionHash returns hash of psc contract's transaction.
func (api *API) GetContractTransactionHash() string {
	data := make(map[string]interface{})
	api.fetchFromTruffle("/getPSC", &data)
	return data["contract"].(map[string]interface{})["transactionHash"].(string)
}

// FetchTestPrivateKey returns first available private key in truffle.
func (api *API) FetchTestPrivateKey() string {
	accounts := api.GetTestAccounts()
	return accounts[0].PrivateKey
}

// GetTestAccountAddress returns first available account's address.
func (api *API) GetTestAccountAddress() string {
	accounts := api.GetTestAccounts()
	return accounts[0].Account
}

// GetTestAccounts returns all available accounts in truffle.
func (api *API) GetTestAccounts() []TestAccount {
	accounts := []TestAccount{}
	api.fetchFromTruffle("/getKeys", &accounts)
	return accounts
}
