// +build !nouitest

package ui

import (
	"errors"
	"testing"
	"time"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/util"
)

func TestAcceptOffering(t *testing.T) {
	fxt := newFixture(t)
	defer fxt.close()

	var j *data.Job
	handler.queue = job.QueueMock(func(method int, j2 *data.Job,
		relatedID, subID string, subFunc job.SubFunc) error {
		switch method {
		case job.MockAdd:
			j = j2
		case job.MockSubscribe:
			go func() {
				time.Sleep(time.Millisecond)
				subFunc(j, nil)
				subFunc(j, errors.New("some error"))
			}()
		}
		return nil
	})

	ch := make(chan *JobResult)

	sub, err := subscribe(client, ch, "acceptOffering",
		"wrong-password", fxt.UserAcc.ID, fxt.Offering.ID, 12345)
	util.TestExpectResult(t, "AcceptOffering", ErrAccessDenied, err)

	sub, err = subscribe(client, ch, "acceptOffering",
		data.TestPassword, util.NewUUID(), fxt.Offering.ID, 12345)
	util.TestExpectResult(t, "AcceptOffering", ErrAccountNotFound, err)

	sub, err = subscribe(client, ch, "acceptOffering",
		data.TestPassword, fxt.UserAcc.ID, util.NewUUID(), 12345)
	util.TestExpectResult(t, "AcceptOffering", ErrOfferingNotFound, err)

	sub, err = subscribe(client, ch, "acceptOffering",
		data.TestPassword, fxt.UserAcc.ID, fxt.Offering.ID, 12345)
	util.TestExpectResult(t, "AcceptOffering", nil, err)
	defer sub.Unsubscribe()

	res := <-ch
	if res.Type != data.JobClientPreChannelCreate || res.Result != nil {
		t.Fatalf("wrong data for the first notification")
	}

	res = <-ch
	if res.Type != data.JobClientPreChannelCreate ||
		res.Result.Message != "some error" {
		t.Fatalf("wrong data for the second notification")
	}
}
