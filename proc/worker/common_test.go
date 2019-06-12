package worker

import (
	"testing"

	"github.com/privatix/dappctrl/data"
)

func TestClientCompleteServiceTransition(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()

	fxt := env.newTestFixture(t,
		data.JobClientPreServiceUnsuspend, data.JobChannel)
	defer fxt.Close()

	fxt.job.Type = data.JobCompleteServiceTransition

	transitions := map[string]string{
		data.ServiceActivating:  data.ServiceActive,
		data.ServiceSuspending:  data.ServiceSuspended,
		data.ServiceTerminating: data.ServiceTerminated,
	}

	for k, v := range transitions {
		fxt.Channel.ServiceStatus = k
		env.updateInTestDB(t, fxt.Channel)

		setJobData(t, db, fxt.job, v)
		runJob(t, env.worker.CompleteServiceTransition, fxt.job)

		var ch data.Channel
		env.findTo(t, &ch, fxt.Channel.ID)

		if ch.ServiceStatus != v {
			t.Fatalf("expected %s service status, but got %s",
				v, ch.ServiceStatus)
		}
	}
}
