package ui_test

import (
	"encoding/json"
	"testing"

	"github.com/privatix/dappctrl/ui"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/util"
)

func TestTopUpChannel(t *testing.T) {
	fxt := newFixture(t)
	defer fxt.close()

	var j *data.Job
	handler.SetMockQueue(job.QueueMock(func(method int, j2 *data.Job,
		relatedIDs []string, subID string, subFunc job.SubFunc) error {
		switch method {
		case job.MockAdd:
			j = j2
		default:
			t.Fatal("unexpected queue call")
		}
		return nil
	}))

	expectResult := func(expected, actual error) {
		util.TestExpectResult(t, "TopUpChannel", expected, actual)
	}

	err := handler.TopUpChannel("wrong-password", fxt.Channel.ID, 123)
	expectResult(ui.ErrAccessDenied, err)

	err = handler.TopUpChannel(data.TestPassword, util.NewUUID(), 123)
	expectResult(ui.ErrChannelNotFound, err)

	err = handler.TopUpChannel(data.TestPassword, fxt.Channel.ID, 123)
	expectResult(nil, err)

	if j == nil || j.RelatedType != data.JobChannel ||
		j.RelatedID != fxt.Channel.ID ||
		j.Type != data.JobClientPreChannelTopUp {
		t.Fatalf("expected job not created")
	}

	// Test default gas price setup.
	var testGasPrice uint64 = 500
	deleteSetting := insertDefaultGasPriceSetting(t, testGasPrice)
	defer deleteSetting()
	handler.TopUpChannel(data.TestPassword, fxt.Channel.ID, 0)
	jdata := &data.JobPublishData{}
	json.Unmarshal(j.Data, jdata)
	if jdata.GasPrice != testGasPrice {
		t.Fatal("job with default gas price expected")
	}
}