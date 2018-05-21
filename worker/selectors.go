package worker

import (
	"github.com/privatix/dappctrl/data"
	reform "gopkg.in/reform.v1"
)

func (w *Worker) relatedAndValidate(rec reform.Record, job *data.Job, jobType, relType string) error {
	if job.Type != jobType || job.RelatedType != relType {
		return ErrInvalidJob
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

func (w *Worker) ethLog(job *data.Job) (log *data.EthLog, err error) {
	log = &data.EthLog{}
	err = w.db.FindOneTo(log, "job", job.ID)
	return
}

func (w *Worker) endpoint(channel string) (endp *data.Endpoint, err error) {
	endp = &data.Endpoint{}
	err = w.db.FindOneTo(endp, "channel", channel)
	return
}

func (w *Worker) offering(pk string) (offering *data.Offering, err error) {
	offering = &data.Offering{}
	err = w.db.FindByPrimaryKeyTo(offering, pk)
	return
}

func (w *Worker) account(ethAddr string) (account *data.Account, err error) {
	account = &data.Account{}
	err = w.db.FindOneTo(account, "eth_addr", ethAddr)
	return
}

func (w *Worker) user(ethAddr string) (user *data.User, err error) {
	user = &data.User{}
	err = w.db.FindOneTo(user, "eth_addr", ethAddr)
	return
}

func (w *Worker) template(pk string) (template *data.Template, err error) {
	template = &data.Template{}
	err = w.db.FindByPrimaryKeyTo(template, pk)
	return
}

func (w *Worker) templateByHash(hash string) (template *data.Template, err error) {
	template = &data.Template{}
	err = w.db.FindOneTo(template, "hash", hash)
	return
}
