// +build !noagentuisrvtest

package uisrv

import (
	"net/http/httptest"
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

func getEndpoints(ch, id string) *httptest.ResponseRecorder {
	return getResources(endpointsPath,
		map[string]string{"ch_id": ch, "id": id},
		testServer.handleGetEndpoints)
}

func testGetEndpoint(t *testing.T, exp int, ch, id string) {
	res := getEndpoints(ch, id)
	testGetResources(t, res, exp)
}

func TestGetEndpoints(t *testing.T) {
	// Get empty list.
	testGetEndpoint(t, 0, "", "")

	// Get all endpoints.

	// Prepare test data.
	ch, deleteChan := createTestChannel()
	defer deleteChan()
	tplAccess := data.NewTestTemplate(data.TemplateAccess)
	endpoint := data.NewTestEndpoint(ch.ID, tplAccess.ID)
	deleteItems := insertItems(tplAccess, endpoint)
	defer deleteItems()

	testGetEndpoint(t, 1, "", "")

	// Get all by channel id.
	testGetEndpoint(t, 1, endpoint.Channel, "")
	testGetEndpoint(t, 0, util.NewUUID(), "")
	// Get by id.
	testGetEndpoint(t, 1, "", endpoint.ID)
	testGetEndpoint(t, 0, "", util.NewUUID())
}
