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

func getLogEvents(t *testing.T, params map[string]string) *http.Response {
	return getResources(t, logsPath, params)
}

func testGetLogEventsResult(t *testing.T, res *http.Response, page, nPage, exp int) {
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

	res := getLogEvents(t, map[string]string{
		"page":    "1",
		"perPage": "2",
	})
	testGetLogEventsResult(t, res, 1, 1, 2)

	res = getLogEvents(t, map[string]string{
		"page":    "1",
		"perPage": "1",
	})
	testGetLogEventsResult(t, res, 1, 2, 1)

	res = getLogEvents(t, map[string]string{
		"page":    "2",
		"perPage": "1",
	})
	testGetLogEventsResult(t, res, 2, 2, 1)

	res = getLogEvents(t, map[string]string{
		"page":    "3",
		"perPage": "1",
	})
	testGetLogEventsResult(t, res, 3, 2, 0)
}

func TestGetLogsFilteringByLevel(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	insertTestLogEvents(t, 2, log.Error)

	// Filter by level
	res := getLogEvents(t, map[string]string{
		"page":    "1",
		"perPage": "2",
		"level":   string(log.Error),
	})
	testGetLogEventsResult(t, res, 1, 1, 2)

	res = getLogEvents(t, map[string]string{
		"page":    "1",
		"perPage": "2",
		"level":   string(log.Fatal),
	})
	testGetLogEventsResult(t, res, 1, 0, 0)
}

func dateArg(d time.Time) string {
	return d.UTC().Format(time.RFC3339)
}

func TestGetLogsByDateRange(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	insertTestLogEvents(t, 2, log.Error)

	// Filter by date range.
	res := getLogEvents(t, map[string]string{
		"page":     "1",
		"perPage":  "2",
		"dateFrom": dateArg(time.Now().Add(-time.Minute)),
		"dateTo":   dateArg(time.Now().Add(time.Minute)),
	})
	testGetLogEventsResult(t, res, 1, 1, 2)

	res = getLogEvents(t, map[string]string{
		"page":     "1",
		"perPage":  "2",
		"dateFrom": dateArg(time.Now().Add(time.Minute)),
		"dateTo":   dateArg(time.Now().Add(time.Hour)),
	})
	testGetLogEventsResult(t, res, 1, 0, 0)
}

func TestGetLogsByMsgText(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)
	data.InsertToTestDB(t, testServer.db, &data.LogEvent{
		Message: "foo",
		Level:   log.Error,
		Context: []byte("{}"),
	})

	res := getLogEvents(t, map[string]string{
		"page":       "1",
		"perPage":    "1",
		"searchText": "fo",
	})
	testGetLogEventsResult(t, res, 1, 1, 1)

	res = getLogEvents(t, map[string]string{
		"page":       "1",
		"perPage":    "1",
		"searchText": "do",
	})
	testGetLogEventsResult(t, res, 1, 0, 0)
}

func TestGetLogsByContext(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	data.InsertToTestDB(t, testServer.db, &data.LogEvent{
		Level:   log.Error,
		Context: []byte("{\"foo\": \"bar\"}"),
	})

	res := getLogEvents(t, map[string]string{
		"page":       "1",
		"perPage":    "1",
		"searchText": "ba",
	})
	testGetLogEventsResult(t, res, 1, 1, 1)

	res = getLogEvents(t, map[string]string{
		"page":       "1",
		"perPage":    "1",
		"searchText": "foo",
	})
	testGetLogEventsResult(t, res, 1, 0, 0)
}
