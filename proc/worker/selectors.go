package worker

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
)

func (w *Worker) validateJob(job *data.Job, jobType, relType string) error {
	if job.Type != jobType || job.RelatedType != relType {
		return ErrInvalidJob
	}
	return nil
}

func (w *Worker) relatedAndValidate(rec reform.Record, job *data.Job, jobType, relType string) error {
	if err := w.validateJob(job, jobType, relType); err != nil {
		return err
	}
	return w.db.FindByPrimaryKeyTo(rec, job.RelatedID)
}

func (w *Worker) relatedOffering(job *data.Job, jobType string) (*data.Offering, error) {
	rec := &data.Offering{}
	err := w.relatedAndValidate(rec, job, jobType, data.JobOfferring)
	return rec, err
}

func (w *Worker) relatedChannel(job *data.Job, jobType string) (*data.Channel, error) {
	rec := &data.Channel{}
	err := w.relatedAndValidate(rec, job, jobType, data.JobChannel)
	return rec, err
}

func (w *Worker) relatedEndpoint(job *data.Job, jobType string) (*data.Endpoint, error) {
	rec := &data.Endpoint{}
	err := w.relatedAndValidate(rec, job, jobType, data.JobEndpoint)
	return rec, err
}

func (w *Worker) relatedAccount(job *data.Job, jobType string) (*data.Account, error) {
	rec := &data.Account{}
	err := w.relatedAndValidate(rec, job, jobType, data.JobAccount)
	return rec, err
}

func (w *Worker) ethLog(job *data.Job) (*data.EthLog, error) {
	log := &data.EthLog{}
	err := w.db.FindOneTo(log, "job", job.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find %T: %v", log, err)
	}
	return log, nil
}

func (w *Worker) endpoint(channel string) (*data.Endpoint, error) {
	endpoint := &data.Endpoint{}
	err := w.db.FindOneTo(endpoint, "channel", channel)
	if err != nil {
		return nil, fmt.Errorf("failed to find %T: %v", endpoint, err)
	}
	return endpoint, nil
}

func (w *Worker) offering(pk string) (*data.Offering, error) {
	offering := &data.Offering{}
	err := w.db.FindByPrimaryKeyTo(offering, pk)
	if err != nil {
		return nil, fmt.Errorf("failed to find %T: %v", offering, err)
	}
	return offering, nil
}

func (w *Worker) offeringByHash(hash common.Hash) (*data.Offering, error) {
	offering := &data.Offering{}
	hashB64 := data.FromBytes(hash.Bytes())
	err := w.db.FindOneTo(offering, "hash", hashB64)
	if err != nil {
		return nil, fmt.Errorf("failed to find %T by hash: %v",
			offering, err)
	}
	return offering, nil
}

func (w *Worker) account(ethAddr string) (*data.Account, error) {
	account := &data.Account{}
	err := w.db.FindOneTo(account, "eth_addr", ethAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to find %T: %v", account, err)
	}
	return account, nil
}

func (w *Worker) user(ethAddr string) (*data.User, error) {
	user := &data.User{}
	err := w.db.FindOneTo(user, "eth_addr", ethAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to find %T by eth addr: %v",
			user, err)
	}
	return user, nil
}

func (w *Worker) template(pk string) (*data.Template, error) {
	template := &data.Template{}
	err := w.db.FindByPrimaryKeyTo(template, pk)
	if err != nil {
		return nil, fmt.Errorf("failed to find %T: %v", template, err)
	}
	return template, nil
}

func (w *Worker) templateByHash(hash string) (*data.Template, error) {
	template := &data.Template{}
	err := w.db.FindOneTo(template, "hash", hash)
	if err != nil {
		return nil, fmt.Errorf("failed to find %T by hash: %v",
			template, err)
	}
	return template, nil
}
