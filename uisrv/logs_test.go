package uisrv

import (
	"testing"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
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

func TestGetLogsPagination(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	insertTestLogEvents(t, 2, log.Error)

	res := getResources(t, logsPath, map[string]string{
		"page":    "1",
		"perPage": "2",
	})
	testGetResources(t, res, 2)

	res = getResources(t, logsPath, map[string]string{
		"page":    "1",
		"perPage": "1",
	})
	testGetResources(t, res, 1)

	res = getResources(t, logsPath, map[string]string{
		"page":    "2",
		"perPage": "1",
	})
	testGetResources(t, res, 1)

	res = getResources(t, logsPath, map[string]string{
		"page":    "3",
		"perPage": "1",
	})
	testGetResources(t, res, 0)
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
	testGetResources(t, res, 2)

	res = getResources(t, logsPath, map[string]string{
		"page":    "1",
		"perPage": "2",
		"level":   string(log.Fatal),
	})
	testGetResources(t, res, 0)
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
	testGetResources(t, res, 2)

	res = getResources(t, logsPath, map[string]string{
		"page":     "1",
		"perPage":  "2",
		"dateFrom": dateArg(time.Now().Add(time.Minute)),
		"dateTo":   dateArg(time.Now().Add(time.Hour)),
	})
	testGetResources(t, res, 0)
}
