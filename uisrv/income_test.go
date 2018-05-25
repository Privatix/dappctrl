// +build !noethtest

package uisrv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

func getIncome(t *testing.T, entity, id string) uint64 {
	url := fmt.Sprintf("http://:%s@%s/%s?%s=%s", testPassword,
		testServer.conf.Addr, incomePath, entity, id)
	r, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to get income: %v. %s", err, util.Caller())
	}
	var income uint64
	if err := json.NewDecoder(r.Body).Decode(&income); err != nil {
		t.Fatalf("could not decode income: %v, %s", err, util.Caller())
	}
	return income
}

func TestIncome(t *testing.T) {
	fixture := data.NewTestFixture(t, testServer.db)
	defer fixture.Close()
	resetCred := setTestUserCredentials(t)
	defer resetCred()

	ch1 := *fixture.Channel
	ch1.ID = util.NewUUID()
	ch1.ReceiptBalance = 10

	ch2 := *fixture.Channel
	ch2.ID = util.NewUUID()
	ch2.ReceiptBalance = 20

	data.InsertToTestDB(t, testServer.db, &ch1, &ch2)
	defer data.DeleteFromTestDB(t, testServer.db, &ch1, &ch2)

	expectedIncome := fixture.Channel.ReceiptBalance
	expectedIncome += ch1.ReceiptBalance + ch2.ReceiptBalance

	var incomeReturned uint64

	fail := func(wanted uint64) {
		t.Fatalf("wanted %v, got %v. %s", wanted, incomeReturned,
			util.Caller())
	}

	// By offering id.
	incomeReturned = getIncome(t, usagesByOfferingID, fixture.Offering.ID)
	if expectedIncome != incomeReturned {
		fail(expectedIncome)
	}
	incomeReturned = getIncome(t, usagesByOfferingID, util.NewUUID())
	if expectedIncome == incomeReturned {
		fail(0)
	}

	// By product id.
	incomeReturned = getIncome(t, usagesByProductID, fixture.Product.ID)
	if expectedIncome != incomeReturned {
		fail(expectedIncome)
	}
	incomeReturned = getIncome(t, usagesByProductID, util.NewUUID())
	if expectedIncome == incomeReturned {
		fail(0)
	}
}
