package proc

import (
	"time"

	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
)

type transition = map[string]struct{}

var (
	activateTransitions = transition{
		data.ServicePending:   struct{}{},
		data.ServiceSuspended: struct{}{},
	}
	suspendTransitions = transition{
		data.ServiceActive: struct{}{},
	}
	terminateTransitions = transition{
		data.ServicePending:   struct{}{},
		data.ServiceActive:    struct{}{},
		data.ServiceSuspended: struct{}{},
	}
)

func checkJobExists(tx *reform.TX, rel, ty string) error {
	var err error
	if len(ty) != 0 {
		_, err = tx.SelectOneFrom(data.JobTable,
			"WHERE status = $1 AND related_id = $2 AND type = $3",
			data.JobActive, rel, ty)
		if err == nil {
			return ErrSameJobExists
		}
	} else {
		_, err = tx.SelectOneFrom(data.JobTable,
			"WHERE status = $1 AND related_id = $2",
			data.JobActive, rel)
		if err == nil {
			return ErrActiveJobsExist
		}
	}

	if err == reform.ErrNoRows {
		return nil
	}

	return err
}

func cancelJobs(tx *reform.TX, rel string) error {
	_, err := tx.Exec(`
		UPDATE jobs
		   SET status = $1
		 WHERE status = $2 AND related_id = $3`,
		data.JobCanceled, data.JobActive, rel)
	return err
}

func (p *Processor) alterServiceStatus(id, jobCreator, jobType,
	jobTypeToCheck string, trans transition, cancel bool) (string, error) {
	tx, err := p.queue.DB().Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var ch data.Channel
	err = tx.SelectOneTo(&ch, "WHERE id = $1 FOR UPDATE", id)
	if err != nil {
		return "", err
	}

	if _, ok := trans[ch.ServiceStatus]; !ok {
		return "", ErrBadServiceStatus
	}

	if err := checkJobExists(tx, ch.ID, jobTypeToCheck); err != nil {
		return "", err
	}

	if cancel {
		if err := cancelJobs(tx, ch.ID); err != nil {
			return "", err
		}
	}

	j := &data.Job{
		Type:        jobType,
		RelatedID:   ch.ID,
		RelatedType: data.JobChannel,
		CreatedAt:   time.Now(),
		CreatedBy:   jobCreator,
		Data:        []byte("{}"),
	}

	if err = p.queue.Add(j); err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return j.ID, nil
}

// SuspendChannel tries to suspend a given channel.
func (p *Processor) SuspendChannel(id, jobCreator string) (string, error) {
	return p.alterServiceStatus(id, jobCreator,
		data.JobAgentPreServiceSuspend, "", suspendTransitions, false)
}

// ActivateChannel tries to activate a given channel.
func (p *Processor) ActivateChannel(id, jobCreator string) (string, error) {
	return p.alterServiceStatus(id, jobCreator,
		data.JobAgentPreServiceUnsuspend, "", activateTransitions,
		false)
}

// TerminateChannel tries to terminate a given channel.
func (p *Processor) TerminateChannel(id, jobCreator string) (string, error) {
	return p.alterServiceStatus(id, jobCreator,
		data.JobAgentPreServiceTerminate,
		data.JobAgentPreServiceTerminate, terminateTransitions, true)
}
