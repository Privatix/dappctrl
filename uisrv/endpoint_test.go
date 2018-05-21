// +build !noagentuisrvtest

package uisrv

import (
	"net/http"
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

func getEndpoints(t *testing.T, ch, id string) *http.Response {
	return getResources(t, endpointsPath,
		map[string]string{"ch_id": ch, "id": id})
}

func testGetEndpoint(t *testing.T, exp int, ch, id string) {
	res := getEndpoints(t, ch, id)
	testGetResources(t, res, exp)
}

func TestGetEndpoints(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	// Get empty list.
	testGetEndpoint(t, 0, "", "")

	// Get all endpoints.

	// Prepare test data.
	ch := createTestChannel(t)
	tplAccess := data.NewTestTemplate(data.TemplateAccess)
	endpoint := data.NewTestEndpoint(ch.ID, tplAccess.ID)
	insertItems(t, tplAccess, endpoint)

	testGetEndpoint(t, 1, "", "")

	// Get all by channel id.
	testGetEndpoint(t, 1, endpoint.Channel, "")
	testGetEndpoint(t, 0, util.NewUUID(), "")
	// Get by id.
	testGetEndpoint(t, 1, "", endpoint.ID)
	testGetEndpoint(t, 0, "", util.NewUUID())
}
