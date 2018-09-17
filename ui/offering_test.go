package ui_test

import (
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/ui"
	"github.com/privatix/dappctrl/util"
)

func TestAcceptOffering(t *testing.T) {
	fxt, assertMatchErr := newTest(t, "AcceptOffering")
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

	_, err := handler.AcceptOffering(
		"wrong-password", fxt.UserAcc.ID, fxt.Offering.ID, 12345)
	assertMatchErr(ui.ErrAccessDenied, err)

	_, err = handler.AcceptOffering(
		data.TestPassword, util.NewUUID(), fxt.Offering.ID, 12345)
	assertMatchErr(ui.ErrAccountNotFound, err)

	_, err = handler.AcceptOffering(
		data.TestPassword, fxt.UserAcc.ID, util.NewUUID(), 12345)
	assertMatchErr(ui.ErrOfferingNotFound, err)

	res, err := handler.AcceptOffering(
		data.TestPassword, fxt.UserAcc.ID, fxt.Offering.ID, 12345)
	assertMatchErr(nil, err)

	if res == nil || j == nil || j.RelatedType != data.JobChannel ||
		j.RelatedID != *res ||
		j.Type != data.JobClientPreChannelCreate {
		t.Fatalf("wrong result data")
	}
}
