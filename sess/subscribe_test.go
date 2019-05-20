package sess_test

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/sess"
	"github.com/privatix/dappctrl/util"
)

func TestConnChange(t *testing.T) {
	t.Run("ErrAccessDenied", func(t *testing.T) {
		fxt := newTestFixture(t)
		defer fxt.Close()

		ch := make(chan *sess.ConnChangeResult)
		_, err := client.Subscribe(context.Background(),
			"sess", ch, "connChange", "bad-product", data.TestPassword)
		util.TestExpectResult(t, "ConnChange", sess.ErrAccessDenied, err)

		_, err = client.Subscribe(context.Background(),
			"sess", ch, "connChange", fxt.Product.ID, "bad-password")
		util.TestExpectResult(t, "ConnChange", sess.ErrAccessDenied, err)
	})

	t.Run("SubscribeAndUnsubscribe", func(t *testing.T) {
		fxt := newTestFixture(t)
		defer fxt.Close()

		var mu sync.Mutex
		callMethods := make([]int, 0)

		client, _ := newClient(job.QueueMock(func(method int, _ *reform.TX,
			_ *data.Job, _ []string, _ string, _ job.SubFunc) error {
			mu.Lock()
			defer mu.Unlock()
			callMethods = append(callMethods, method)
			return nil
		}))

		ch := make(chan *sess.ConnChangeResult)
		sub, err := client.Subscribe(context.Background(), "sess", ch, "connChange",
			fxt.Product.ID, data.TestPassword)
		util.TestExpectResult(t, "ConnChange", nil, err)

		sub.Unsubscribe()
		want := []int{job.MockSubscribe, job.MockUnsubscribe}
		if !reflect.DeepEqual(callMethods, want) {
			t.Fatalf("callMethods=%v, want %v", callMethods, want)
		}
	})

	t.Run("Jobs", func(t *testing.T) {
		fxt := newTestFixture(t)
		defer fxt.Close()

		fxt.Channel.ServiceStatus = data.ServicePending
		data.SaveToTestDB(t, db, fxt.Channel)

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
					subFunc(j, errors.New("some error")) // ignored

					for _, cs := range []struct{ t, s string }{
						// Client jobs.
						{data.JobClientEndpointGet, data.ServiceSuspended},      // ignored.
						{data.JobClientEndpointGet, data.ServicePending},        // ignored.
						{data.JobClientPreServiceUnsuspend, data.ServiceActive}, // ignored.
						{data.JobClientPreServiceUnsuspend, data.ServiceActivating},
						{data.JobClientPreServiceSuspend, data.ServiceSuspended}, // ignored.
						{data.JobClientPreServiceSuspend, data.ServiceSuspending},
						{data.JobClientPreServiceTerminate, data.ServiceTerminated}, // ignored.
						{data.JobClientPreServiceTerminate, data.ServiceTerminating},
						// Agent jobs.
						{data.JobAgentPreEndpointMsgCreate, data.ServiceActive}, // ignored
						{data.JobAgentPreEndpointMsgCreate, data.ServiceActivating},
						{data.JobAgentPreEndpointMsgCreate, data.ServiceSuspended}, // ignored.
						{data.JobAgentPreServiceUnsuspend, data.ServiceActive},     // ignored.
						{data.JobAgentPreServiceUnsuspend, data.ServiceActivating},
						{data.JobAgentPreServiceSuspend, data.ServiceSuspended}, // ignored.
						{data.JobAgentPreServiceSuspend, data.ServiceSuspending},
						{data.JobAgentPreServiceTerminate, data.ServiceTerminated}, // ignored.
						{data.JobAgentPreServiceTerminate, data.ServiceTerminating},
					} {
						j.Type = cs.t
						fxt.Channel.ServiceStatus = cs.s
						data.SaveToTestDB(t, db, fxt.Channel)
						subFunc(j, nil)
					}
				}()
			case job.MockUnsubscribe:
			default:
				t.Fatal("unexpected queue call")
			}

			return nil
		}))

		ch := make(chan *sess.ConnChangeResult)
		sub, err := client.Subscribe(context.Background(),
			"sess", ch, "connChange", fxt.Product.ID, data.TestPassword)
		util.TestExpectResult(t, "ConnChange", nil, err)
		defer sub.Unsubscribe()

		for i, wantStatus := range []string{
			// Client job statuses.
			sess.ConnStart,
			sess.ConnStop,
			sess.ConnStop,
			// Agent job statuses.
			sess.ConnStart,
			sess.ConnStart,
			sess.ConnStop,
			sess.ConnStop,
		} {
			ret := <-ch
			if ret.Channel != fxt.Channel.ID || ret.Status != wantStatus {
				t.Fatalf("ConnChange(#%d)=%s, want %s", i, ret.Status, wantStatus)
			}
		}
	})
}
