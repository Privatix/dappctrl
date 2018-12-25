package sess_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/sess"
	"github.com/privatix/dappctrl/util"
)

func subscribeConnChange(client *rpc.Client, product, password string,
	ch chan *sess.ConnChangeResult) (*rpc.ClientSubscription, error) {
	return client.Subscribe(context.Background(),
		"sess", ch, "connChange", product, password)
}

func setChannelServiceStatus(t *testing.T,
	fxt *data.TestFixture, status string) {
	fxt.Channel.ServiceStatus = status
	data.SaveToTestDB(t, db, fxt.Channel)
}

var connChangeEvents = []struct{ t, s string }{
	{data.JobClientPreServiceSuspend, data.ServiceSuspended},     // ignored
	{data.JobClientPreServiceSuspend, data.ServiceSuspending},    // ok
	{data.JobClientPreServiceUnsuspend, data.ServiceActive},      // ignored
	{data.JobClientPreServiceUnsuspend, data.ServiceActivating},  // ok
	{data.JobClientPreServiceTerminate, data.ServiceTerminated},  // ignored
	{data.JobClientPreServiceTerminate, data.ServiceTerminating}, // ok
}

var connChangeStatuses = []string{
	sess.ConnStop,
	sess.ConnStart,
	sess.ConnStop,
}

func TestConnChange(t *testing.T) {
	fxt := newTestFixture(t)
	defer fxt.Close()

	setChannelServiceStatus(t, fxt, data.ServicePending)

	unsubscribed := false
	mtx := sync.Mutex{}

	j := &data.Job{
		Type:      data.JobClientPreServiceUnsuspend,
		RelatedID: fxt.Channel.ID,
	}

	client, _ := newClient(job.QueueMock(func(method int, tx *reform.TX,
		j2 *data.Job, jobTypes []string, subID string,
		subFunc job.SubFunc) error {
		mtx.Lock()
		defer mtx.Unlock()

		switch method {
		case job.MockSubscribe:
			go func() {
				time.Sleep(time.Millisecond)

				subFunc(j, errors.New("some error")) // ignored

				for _, cs := range connChangeEvents {
					j.Type = cs.t
					setChannelServiceStatus(t, fxt, cs.s)
					subFunc(j, nil)
				}
			}()
		case job.MockUnsubscribe:
			unsubscribed = true
		default:
			t.Fatal("unexpected queue call")
		}

		return nil
	}))

	ch := make(chan *sess.ConnChangeResult)

	_, err := subscribeConnChange(client,
		"bad-channel", data.TestPassword, ch)
	util.TestExpectResult(t, "ConnChange", sess.ErrAccessDenied, err)

	_, err = subscribeConnChange(client,
		fxt.Product.ID, "bad-password", ch)
	util.TestExpectResult(t, "ConnChange", sess.ErrAccessDenied, err)

	sub, err := subscribeConnChange(client,
		fxt.Product.ID, data.TestPassword, ch)
	util.TestExpectResult(t, "ConnChange", nil, err)

	for i, v := range connChangeStatuses {
		ret := <-ch
		if ret.Channel != fxt.Channel.ID || ret.Status != v {
			t.Fatalf("wrong data for notification %d", i)
		}

	}

	sub.Unsubscribe()
	time.Sleep(time.Millisecond)

	mtx.Lock()
	defer mtx.Unlock()
	if !unsubscribed {
		t.Fatal("didn't unsubscribe")
	}
}
