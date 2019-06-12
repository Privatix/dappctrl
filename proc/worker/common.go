package worker

import "github.com/privatix/dappctrl/data"

// CompleteServiceTransition is an end step of service status transitioning.
func (w *Worker) CompleteServiceTransition(job *data.Job) error {
	logger := w.logger.Add("method", "CompleteServiceTransition",
		"job", job)

	ch, err := w.relatedChannel(
		logger, job, data.JobCompleteServiceTransition)
	if err != nil {
		return err
	}

	logger = logger.Add("channel", ch)

	err = w.unmarshalDataTo(logger, job.Data, &ch.ServiceStatus)
	if err != nil {
		return err
	}

	return w.saveRecord(logger, w.db.Querier, ch)
}
