package ui_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/ui"
	"github.com/privatix/dappctrl/util"
)

func TestObjectChange(t *testing.T) {
	fxt, assertErrEqual := newTest(t, "ObjectChange")
	defer fxt.close()

	unsubscribed := false
	mtx := sync.Mutex{}

	j1 := &data.Job{RelatedID: fxt.Channel.ID}
	j2 := &data.Job{RelatedID: util.NewUUID()}
	handler.SetMockQueue(job.QueueMock(func(method int, tx *reform.TX,
		j3 *data.Job, relatedIDs []string, subID string,
		subFunc job.SubFunc) error {
		mtx.Lock()
		defer mtx.Unlock()

		switch method {
		case job.MockSubscribe:
			go func() {
				time.Sleep(time.Millisecond)
				subFunc(j1, nil)
				subFunc(j2, errors.New("some error"))
			}()
		case job.MockUnsubscribe:
			unsubscribed = true
		default:
			t.Fatal("unexpected queue call")
		}
		return nil
	}))

	ch := make(chan *ui.ObjectChangeResult)
	_, err := subscribe(client, ch, "objectChange",
		"bad-password", data.JobChannel, nil)
	assertErrEqual(ui.ErrAccessDenied, err)

	_, err = subscribe(client, ch, "objectChange",
		data.TestPassword, "bad-object-type", nil)
	assertErrEqual(ui.ErrBadObjectType, err)

	sub, err := subscribe(client, ch, "objectChange", data.TestPassword,
		data.JobChannel, []string{j1.RelatedID, j2.RelatedID})
	assertErrEqual(nil, err)

	var ch2 data.Channel

	ret := <-ch
	util.TestUnmarshalJSON(t, []byte(ret.Object), &ch2)
	if ret.Object == nil || ch2.ID != j1.RelatedID ||
		ret.Job == nil || ret.Error != nil {
		t.Fatalf("wrong data for the first notification")
	}

	ret = <-ch
	util.TestUnmarshalJSON(t, []byte(ret.Object), &ch2)
	if ret.Object != nil || ret.Job == nil ||
		ret.Error.Message != "some error" {
		t.Fatalf("wrong data for the second notification")
	}

	sub.Unsubscribe()
	time.Sleep(time.Millisecond)

	mtx.Lock()
	defer mtx.Unlock()
	if !unsubscribed {
		t.Fatal("didn't unsubscribe")
	}
}
