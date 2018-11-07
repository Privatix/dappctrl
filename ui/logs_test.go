package ui_test

import (
	"testing"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/ui"
	"github.com/privatix/dappctrl/util/log"
)

type getLogsTestData struct {
	offset     uint
	limit      uint
	searchText string
	level      []string
	dateFrom   string
	dataTo     string
	totalItems int
	exp        int
}

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
	data.InsertToTestDB(t, db, logs...)
}

func assertResult(t *testing.T, res *ui.GetLogsResult, err error,
	totalItems, exp int, assertErrEqual func(error, error)) {
	assertErrEqual(nil, err)

	if len(res.Items) != exp || res.TotalItems != totalItems {
		t.Fatalf("wanted (items, total items): (%d, %d), got (%d, %d)",
			exp, totalItems, len(res.Items), res.TotalItems)
	}
}

func dateArg(d time.Time) string {
	return d.UTC().Format(time.RFC3339)
}

func TestGetLogs(t *testing.T) {
	defer data.CleanTestTable(t, db, data.LogEventView)

	fxt, assertErrEqual := newTest(t, "GetLogs")
	defer fxt.close()

	insertTestLogEvents(t, 2, log.Error)

	data.InsertToTestDB(t, db, &data.LogEvent{
		Message: "foo foo",
		Level:   log.Info,
		Context: []byte("{}"),
	})

	data.InsertToTestDB(t, db, &data.LogEvent{
		Level:   log.Debug,
		Context: []byte("{\"foo\": \"bar\"}"),
	})

	_, err := handler.GetLogs("wrong-password",
		[]string{string(log.Error)}, "", "", "", 2, 1)
	assertErrEqual(ui.ErrAccessDenied, err)

	testData := []*getLogsTestData{
		// Test pagination.
		{1, 2, "", []string{string(log.Error)}, "", "", 2, 1},
		{1, 1, "", []string{string(log.Error)}, "", "", 2, 1},
		{2, 1, "", []string{string(log.Error)}, "", "", 2, 0},
		{0, 3, "", []string{string(log.Error)}, "", "", 2, 2},
		{0, 0, "", []string{string(log.Error)}, "", "", 2, 2},
		{1, 0, "", []string{string(log.Error)}, "", "", 2, 1},
		// Test filtering by level.
		{1, 2, "", []string{string(log.Error)}, "", "", 2, 1},
		{1, 2, "", []string{string(log.Fatal)}, "", "", 0, 0},
		// Test filtering by date range.
		{1, 2, "", []string{string(log.Error)},
			dateArg(time.Now().Add(-time.Minute)),
			dateArg(time.Now().Add(time.Minute)), 2, 1},
		{1, 2, "", []string{string(log.Fatal)},
			dateArg(time.Now().Add(time.Minute)),
			dateArg(time.Now().Add(time.Hour)), 0, 0},
		// Test filtering by msg text.
		{0, 1, "fo", []string{string(log.Info)}, "", "", 1, 1},
		{0, 1, "do", nil, "", "", 0, 0},
		// Test filtering by context.
		{0, 1, "ba", nil, "", "", 1, 1},
		{0, 1, "foo", nil, "", "", 1, 1},
		{0, 1, "foo foo", nil, "", "", 1, 1},
	}

	for _, v := range testData {
		res, err := handler.GetLogs(
			data.TestPassword, v.level, v.searchText,
			v.dateFrom, v.dataTo, v.offset, v.limit)
		assertResult(t, res, err, v.totalItems, v.exp, assertErrEqual)
	}
}
