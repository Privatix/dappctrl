package uisrv

import (
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

func testGetTransactions(t *testing.T, exp int, relType, relID string) {
	params := make(map[string]string)
	if relType != "" {
		params["relatedType"] = relType
	}
	if relID != "" {
		params["relatedID"] = relID
	}
	res := getResources(t, transactionsPath, params)
	testGetResources(t, res, exp)
}

func TestGetTransactions(t *testing.T) {
	defer setTestUserCredentials(t)()
	testGetTransactions(t, 0, "", "")
	testRelID := util.NewUUID()
	tx := &data.EthTx{
		ID:          util.NewUUID(),
		Status:      data.TxSent,
		GasPrice:    1,
		Gas:         1,
		TxRaw:       []byte("{}"),
		RelatedType: data.JobChannel,
		RelatedID:   testRelID,
	}
	data.InsertToTestDB(t, testServer.db, tx)
	defer data.DeleteFromTestDB(t, testServer.db, tx)
	testGetTransactions(t, 1, "", "")
	testGetTransactions(t, 0, data.JobAccount, "")
	testGetTransactions(t, 0, "", util.NewUUID())
	testGetTransactions(t, 1, data.JobChannel, testRelID)
}
