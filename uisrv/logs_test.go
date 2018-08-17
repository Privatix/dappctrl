package uisrv

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

func insertTestLogEvents(t *testing.T, qty int, lvl log.Level) {
	logs := make([]reform.Struct, qty)
	logTime := time.Now()
	for ; qty > 0; qty-- {
		logs[qty-1] = &data.LogEvent{
			Time:    logTime,
			Level:   lvl,
			Context: []byte("{}"),
		}
	}
	data.InsertToTestDB(t, testServer.db, logs...)
}

func testGetLogs(t *testing.T, res *http.Response, page, nPage, exp int) {
	if res.StatusCode != http.StatusOK {
		t.Fatal("failed to get logs: ", res.StatusCode)
	}
	ret := paginatedReply{}
	json.NewDecoder(res.Body).Decode(&ret)
	if page != ret.Current || exp != len(ret.Items) || nPage != ret.Pages {
		t.Fatalf("expected (current, pages, items): (%d, %d, %d), "+
			"got: (%d, %d, %d) (%s)", page, nPage, exp,
			ret.Current, ret.Pages, len(ret.Items), util.Caller())
	}
}

func TestGetLogsPagination(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	insertTestLogEvents(t, 2, log.Error)

	res := getResources(t, logsPath, map[string]string{
		"page":    "1",
		"perPage": "2",
	})
	testGetLogs(t, res, 1, 1, 2)

	res = getResources(t, logsPath, map[string]string{
		"page":    "1",
		"perPage": "1",
	})
	testGetLogs(t, res, 1, 2, 1)

	res = getResources(t, logsPath, map[string]string{
		"page":    "2",
		"perPage": "1",
	})
	testGetLogs(t, res, 2, 2, 1)

	res = getResources(t, logsPath, map[string]string{
		"page":    "3",
		"perPage": "1",
	})
	testGetLogs(t, res, 3, 2, 0)
}

func TestGetLogsFilteringByLevel(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	insertTestLogEvents(t, 2, log.Error)

	// Filter by level
	res := getResources(t, logsPath, map[string]string{
		"page":    "1",
		"perPage": "2",
		"level":   string(log.Error),
	})
	testGetLogs(t, res, 1, 1, 2)

	res = getResources(t, logsPath, map[string]string{
		"page":    "1",
		"perPage": "2",
		"level":   string(log.Fatal),
	})
	testGetLogs(t, res, 1, 0, 0)
}

func dateArg(d time.Time) string {
	return d.UTC().Format(time.RFC3339)
}

func TestGetLogsByDateRange(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	insertTestLogEvents(t, 2, log.Error)

	// Filter by date range.
	res := getResources(t, logsPath, map[string]string{
		"page":     "1",
		"perPage":  "2",
		"dateFrom": dateArg(time.Now().Add(-time.Minute)),
		"dateTo":   dateArg(time.Now().Add(time.Minute)),
	})
	testGetLogs(t, res, 1, 1, 2)

	res = getResources(t, logsPath, map[string]string{
		"page":     "1",
		"perPage":  "2",
		"dateFrom": dateArg(time.Now().Add(time.Minute)),
		"dateTo":   dateArg(time.Now().Add(time.Hour)),
	})
	testGetLogs(t, res, 1, 0, 0)
}
