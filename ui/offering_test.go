package ui_test

import (
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/ui"
	"github.com/privatix/dappctrl/util"
)

func TestAcceptOffering(t *testing.T) {
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
		util.TestExpectResult(t, "AcceptOffering", expected, actual)
	}

	_, err := handler.AcceptOffering(
		"wrong-password", fxt.UserAcc.ID, fxt.Offering.ID, 12345)
	expectResult(ui.ErrAccessDenied, err)

	_, err = handler.AcceptOffering(
		data.TestPassword, util.NewUUID(), fxt.Offering.ID, 12345)
	expectResult(ui.ErrAccountNotFound, err)

	_, err = handler.AcceptOffering(
		data.TestPassword, fxt.UserAcc.ID, util.NewUUID(), 12345)
	expectResult(ui.ErrOfferingNotFound, err)

	res, err := handler.AcceptOffering(
		data.TestPassword, fxt.UserAcc.ID, fxt.Offering.ID, 12345)
	expectResult(nil, err)

	if res == nil || j == nil || j.RelatedType != data.JobChannel ||
		j.RelatedID != res.Channel ||
		j.Type != data.JobClientPreChannelCreate {
		t.Fatalf("wrong result data")
	}
}
