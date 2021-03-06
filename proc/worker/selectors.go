package worker

import (
	"database/sql"
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/log"
)

func (w *Worker) isJobInvalid(job *data.Job, jobType, relType string) bool {
	return job.Type != jobType || job.RelatedType != relType
}

func (w *Worker) relatedOffering(logger log.Logger, job *data.Job,
	jobType string) (*data.Offering, error) {
	if w.isJobInvalid(job, jobType, data.JobOffering) {
		return nil, ErrInvalidJob
	}

	rec := &data.Offering{}
	err := data.FindByPrimaryKeyTo(w.db.Querier, rec, job.RelatedID)
	if err != nil {
		return nil, ErrOfferingNotFound
	}

	return rec, err
}

func (w *Worker) relatedChannel(logger log.Logger, job *data.Job,
	jobType string) (*data.Channel, error) {
	if w.isJobInvalid(job, jobType, data.JobChannel) {
		return nil, ErrInvalidJob
	}

	rec := &data.Channel{}
	err := data.FindByPrimaryKeyTo(w.db.Querier, rec, job.RelatedID)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrChannelNotFound
	}

	return rec, err
}

func (w *Worker) relatedEthTx(logger log.Logger, job *data.Job,
	jobType string) (*data.EthTx, error) {
	if w.isJobInvalid(job, jobType, data.JobTransaction) {
		return nil, ErrInvalidJob
	}

	rec := &data.EthTx{}
	err := data.FindByPrimaryKeyTo(w.db.Querier, rec, job.RelatedID)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrTxNotFound
	}
	return rec, nil
}

func (w *Worker) relatedEndpoint(logger log.Logger, job *data.Job,
	jobType string) (*data.Endpoint, error) {
	if w.isJobInvalid(job, jobType, data.JobEndpoint) {
		return nil, ErrInvalidJob
	}

	rec := &data.Endpoint{}
	err := data.FindByPrimaryKeyTo(w.db.Querier, rec, job.RelatedID)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrEndpointNotFound
	}

	return rec, err
}

func (w *Worker) relatedAccount(logger log.Logger, job *data.Job,
	jobType string) (*data.Account, error) {
	if w.isJobInvalid(job, jobType, data.JobAccount) {
		return nil, ErrInvalidJob
	}

	rec := &data.Account{}
	err := data.FindByPrimaryKeyTo(w.db.Querier, rec, job.RelatedID)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrAccountNotFound
	}

	return rec, err
}

func (w *Worker) ethLog(logger log.Logger, job *data.Job) (*data.JobEthLog, error) {
	jdata := &data.JobData{}
	err := json.Unmarshal(job.Data, jdata)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrEthLogNotFound
	}
	return jdata.EthLog, nil
}

func (w *Worker) channel(logger log.Logger, pk string) (*data.Channel, error) {
	channel := &data.Channel{}
	err := data.FindByPrimaryKeyTo(w.db.Querier, channel, pk)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrChannelNotFound
	}
	return channel, nil
}

func (w *Worker) endpoint(logger log.Logger, channel string) (*data.Endpoint, error) {
	endpoint := &data.Endpoint{}
	err := data.FindOneTo(w.db.Querier, endpoint, "channel", channel)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrEndpointNotFound
	}
	return endpoint, nil
}

func (w *Worker) endpointByPK(logger log.Logger, pk string) (*data.Endpoint, error) {
	endpoint := &data.Endpoint{}
	err := data.FindByPrimaryKeyTo(w.db.Querier, endpoint, pk)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrEndpointNotFound
	}
	return endpoint, nil
}

func (w *Worker) offering(logger log.Logger, pk string) (*data.Offering, error) {
	offering := &data.Offering{}
	err := data.FindByPrimaryKeyTo(w.db.Querier, offering, pk)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrOfferingNotFound
	}
	return offering, nil
}

func (w *Worker) productByPK(logger log.Logger, pk string) (*data.Product, error) {
	product := &data.Product{}
	err := data.FindByPrimaryKeyTo(w.db.Querier, product, pk)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrProductNotFound
	}
	return product, nil
}

func (w *Worker) offeringByHash(logger log.Logger,
	hash common.Hash) (*data.Offering, error) {
	hashStr := data.HexFromBytes(hash.Bytes())
	return w.offeringByHashString(logger, hashStr)
}

func (w *Worker) offeringByHashString(logger log.Logger,
	hash data.HexString) (*data.Offering, error) {
	offering := &data.Offering{}
	err := w.db.FindOneTo(offering, "hash", hash)
	if err == sql.ErrNoRows {
		return nil, ErrOfferingNotFound
	}
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}
	return offering, nil
}

func (w *Worker) account(logger log.Logger, ethAddr data.HexString) (*data.Account, error) {
	account := &data.Account{}
	err := data.FindOneTo(w.db.Querier, account, "eth_addr", ethAddr)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrAccountNotFound
	}
	return account, nil
}

func (w *Worker) accountByPK(logger log.Logger, pk string) (*data.Account, error) {
	account := &data.Account{}
	err := data.FindByPrimaryKeyTo(w.db.Querier, account, pk)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrAccountNotFound
	}
	return account, nil
}

func (w *Worker) user(logger log.Logger, ethAddr data.HexString) (*data.User, error) {
	user := &data.User{}
	err := data.FindOneTo(w.db.Querier, user, "eth_addr", ethAddr)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (w *Worker) template(logger log.Logger, pk string) (*data.Template, error) {
	template := &data.Template{}
	err := data.FindByPrimaryKeyTo(w.db.Querier, template, pk)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrTemplateNotFound
	}
	return template, nil
}

func (w *Worker) templateByHash(logger log.Logger, hash data.HexString) (*data.Template, error) {
	template := &data.Template{}
	err := data.FindOneTo(w.db.Querier, template, "hash", hash)
	if err != nil {
		logger.Add("hash", hash).Error(err.Error())
		return nil, ErrTemplateByHashNotFound
	}
	return template, nil
}
