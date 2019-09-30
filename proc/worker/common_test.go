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
	defer fxt.close()

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

func TestIncreaseTxGasPrice(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()

	fxt := env.newTestFixture(t,
		data.JobIncreaseTxGasPrice, data.JobTransaction)
	defer fxt.close()

	// Low gas price validation.
	jdata := &data.JobPublishData{
		GasPrice: fxt.EthTx.GasPrice - 1,
	}
	setJobData(t, fxt.DB, fxt.job, jdata)
	if err := env.worker.IncreaseTxGasPrice(fxt.job); err != ErrTxNoGasIncrease {
		t.Fatalf("wanted: %v, got: %v", ErrTxNoGasIncrease, err)
	}
	jdata.GasPrice = fxt.EthTx.GasPrice + 1
	setJobData(t, fxt.DB, fxt.job, jdata)

	// Transaction is already mined validation.
	if err := env.worker.IncreaseTxGasPrice(fxt.job); err != ErrEthTxIsMined {
		t.Fatalf("wanted: %v, got: %v", ErrEthTxIsMined, err)
	}
	env.ethBack.TxIsPending = true

	// Success case.
	if err := env.worker.IncreaseTxGasPrice(fxt.job); err != nil {
		t.Fatalf("wanted success, got: %v", err)
	}

	var tx data.EthTx
	data.FindInTestDB(t, db, &tx, "related_id", fxt.job.RelatedID)
	data.DeleteFromTestDB(t, db, &tx)
}
