package uisrv

import (
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

func testGetTransactions(t *testing.T, exp int) {
	res := getResources(t, transactionsPath, nil)
	testGetResources(t, res, exp)
}

func TestGetTransactions(t *testing.T) {
	defer setTestUserCredentials(t)()
	testGetTransactions(t, 0)
	tx := &data.EthTx{
		ID:          util.NewUUID(),
		Status:      data.TxSent,
		GasPrice:    1,
		Gas:         1,
		TxRaw:       []byte("{}"),
		RelatedType: data.JobChannel,
		RelatedID:   util.NewUUID(),
	}
	data.InsertToTestDB(t, testServer.db, tx)
	defer data.DeleteFromTestDB(t, testServer.db, tx)
	testGetTransactions(t, 1)
}
