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

func getUsage(t *testing.T, entity, id string) uint64 {
	url := fmt.Sprintf("http://:%s@%s/%s?%s=%s", testPassword,
		testServer.conf.Addr, usagePath, entity, id)
	r, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to get usage: %v. %s", err, util.Caller())
	}
	var usage uint64
	if err := json.NewDecoder(r.Body).Decode(&usage); err != nil {
		t.Fatalf("could not decode usage: %v, %s", err, util.Caller())
	}
	return usage
}

func TestUsage(t *testing.T) {
	fixture := data.NewTestFixture(t, testServer.db)
	defer fixture.Close()
	resetCred := setTestUserCredentials(t)
	defer resetCred()

	sess1 := data.NewTestSession(fixture.Channel.ID)
	sess1.UnitsUsed = 10

	sess2 := data.NewTestSession(fixture.Channel.ID)
	sess2.UnitsUsed = 20

	data.InsertToTestDB(t, testServer.db, sess1, sess2)
	defer data.DeleteFromTestDB(t, testServer.db, sess1, sess2)

	expectedUsage := sess1.UnitsUsed + sess2.UnitsUsed

	var usageReturned uint64

	fail := func(wanted uint64) {
		t.Fatalf("wanted %v, got %v. %s", wanted, usageReturned,
			util.Caller())
	}

	// By channel id.
	usageReturned = getUsage(t, usagesByChannelID, fixture.Channel.ID)
	if expectedUsage != usageReturned {
		fail(expectedUsage)
	}
	usageReturned = getUsage(t, usagesByChannelID, util.NewUUID())
	if expectedUsage == usageReturned {
		fail(0)
	}

	// By offering id.
	usageReturned = getUsage(t, usagesByOfferingID, fixture.Offering.ID)
	if expectedUsage != usageReturned {
		fail(expectedUsage)
	}
	usageReturned = getUsage(t, usagesByOfferingID, util.NewUUID())
	if expectedUsage == usageReturned {
		fail(0)
	}

	// By product id.
	usageReturned = getUsage(t, usagesByProductID, fixture.Product.ID)
	if expectedUsage != usageReturned {
		fail(expectedUsage)
	}
	usageReturned = getUsage(t, usagesByProductID, util.NewUUID())
	if expectedUsage == usageReturned {
		fail(0)
	}
}
