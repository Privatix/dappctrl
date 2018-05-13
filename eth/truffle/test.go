// +build !noethtest

package truffle

// GetTestAccounts returns all available accounts in truffle.
func (api *API) GetTestAccounts() []TestAccount {
	accounts := []TestAccount{}
	api.fetchFromTruffle("/getKeys", &accounts)
	return accounts
}
