package worker

import (
	"context"
	"crypto/ecdsa"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	ethutil "github.com/privatix/dappctrl/eth/util"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

func (w *Worker) accountKey(logger log.Logger, ethAddr string) (*ecdsa.PrivateKey, error) {
	acc, err := w.account(logger, ethAddr)
	if err != nil {
		return nil, err
	}

	return w.key(logger, acc.PrivateKey)
}

func (w *Worker) key(logger log.Logger, key string) (*ecdsa.PrivateKey, error) {
	ret, err := w.decryptKeyFunc(key, w.pwdGetter.Get())
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrParsePrivateKey
	}
	return ret, nil
}

func (w *Worker) toOfferingHashArr(logger log.Logger, h string) (ret [common.HashLength]byte, err error) {
	var hash common.Hash
	hash, err = data.ToHash(h)
	if err != nil {
		logger.Error(err.Error())
		err = ErrParseOfferingHash
	}
	ret = [common.HashLength]byte(hash)
	return
}

func (w *Worker) balanceData(logger log.Logger, job *data.Job) (*data.JobBalanceData, error) {
	balanceData := &data.JobBalanceData{}
	if err := w.unmarshalDataTo(logger, job.Data, balanceData); err != nil {
		return nil, err
	}
	return balanceData, nil
}

func (w *Worker) publishData(logger log.Logger, job *data.Job) (*data.JobPublishData, error) {
	publishData := &data.JobPublishData{}
	if err := w.unmarshalDataTo(logger, job.Data, publishData); err != nil {
		return nil, err
	}
	return publishData, nil
}

func (w *Worker) unmarshalDataTo(
	logger log.Logger, jobData []byte, v interface{}) error {
	if err := json.Unmarshal(jobData, v); err != nil {
		logger.Error(err.Error())
		return ErrParseJobData
	}
	return nil
}

func (w *Worker) ethLogTx(logger log.Logger, ethLog *data.EthLog) (*types.Transaction, error) {
	hash, err := data.HexToHash(ethLog.TxHash)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrParseEthLog
	}

	return w.getTransaction(logger, hash)
}

func (w *Worker) newUser(logger log.Logger, tx *types.Transaction) (*data.User, bool, error) {
	signer := &types.HomesteadSigner{}
	pubkey, err := ethutil.RecoverPubKey(signer, tx)
	if err != nil {
		logger.Error(err.Error())
		return nil, false, ErrRecoverClientPubKey
	}

	addr := data.HexFromBytes(crypto.PubkeyToAddress(*pubkey).Bytes())

	_, err = w.db.FindOneFrom(data.UserTable, "eth_addr", addr)
	if err != sql.ErrNoRows {
		return nil, false, nil
	}

	return &data.User{
		ID:        util.NewUUID(),
		EthAddr:   addr,
		PublicKey: data.FromBytes(crypto.FromECDSAPub(pubkey)),
	}, true, nil
}

func (w *Worker) addJob(logger log.Logger, jType, rType, rID string) error {
	err := job.AddSimple(w.queue, jType, rType, rID, data.JobTask)
	if err != nil {
		logger.Error(err.Error())
		return ErrAddJob
	}
	return nil
}

func (w *Worker) addJobWithData(logger log.Logger,
	jType, rType, rID string, jData interface{}) error {
	err := job.AddWithData(w.queue, jType, rType, rID, data.JobTask, jData)
	if err != nil {
		logger.Error(err.Error())
		return ErrAddJob
	}
	return nil
}

func (w *Worker) addJobWithDelay(logger log.Logger,
	jType, rType, rID string, delay time.Duration) error {
	err := job.AddWithDelay(w.queue, jType, rType, rID, data.JobTask, delay)
	if err != nil {
		logger.Error(err.Error())
		return ErrAddJob
	}
	return nil
}

func (w *Worker) updateAccountBalances(job *data.Job, jobType string) error {
	logger := w.logger.Add("method", "updateAccountBalances", "job", job)

	acc, err := w.relatedAccount(logger, job, jobType)
	if err != nil {
		return err
	}

	agentAddr, err := data.HexToAddress(acc.EthAddr)
	if err != nil {
		logger.Error(err.Error())
		return ErrParseEthAddr
	}

	amount, err := w.ethBack.PTCBalanceOf(&bind.CallOpts{}, agentAddr)
	if err != nil {
		logger.Error(err.Error())
		return ErrPTCRetrieveBalance
	}

	acc.PTCBalance = amount.Uint64()

	amount, err = w.ethBack.PSCBalanceOf(&bind.CallOpts{}, agentAddr)
	if err != nil {
		logger.Error(err.Error())
		return ErrPSCRetrieveBalance
	}

	acc.PSCBalance = amount.Uint64()

	amount, err = w.ethBalance(logger, agentAddr)
	if err != nil {
		logger.Error(err.Error())
		return ErrEthRetrieveBalance
	}

	acc.EthBalance = data.B64BigInt(data.FromBytes(amount.Bytes()))

	now := time.Now()

	acc.LastBalanceCheck = &now

	return w.saveRecord(logger, acc)
}

func (w *Worker) ethBalance(logger log.Logger, addr common.Address) (*big.Int, error) {
	amount, err := w.ethBack.EthBalanceAt(context.Background(), addr)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrEthRetrieveBalance
	}

	return amount, nil
}

func (w *Worker) saveEthTX(logger log.Logger, job *data.Job, tx *types.Transaction,
	method, relatedType, relatedID, from, to string) error {
	raw, err := tx.MarshalJSON()
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	dtx := data.EthTx{
		ID:          util.NewUUID(),
		Hash:        data.HexFromBytes(tx.Hash().Bytes()),
		Method:      method,
		Status:      data.TxSent,
		JobID:       pointer.ToString(job.ID),
		Issued:      time.Now(),
		AddrFrom:    from,
		AddrTo:      to,
		Nonce:       pointer.ToString(fmt.Sprint(tx.Nonce())),
		GasPrice:    tx.GasPrice().Uint64(),
		Gas:         tx.Gas(),
		TxRaw:       raw,
		RelatedType: relatedType,
		RelatedID:   relatedID,
	}

	err = data.Insert(w.db.Querier, &dtx)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	return nil
}

// keyFromChannelData returns the unique channel identifier
// used in a Privatix Service Contract.
func (w *Worker) keyFromChannelData(logger log.Logger, channel string) (string, error) {
	ch, err := w.channel(logger, channel)
	if err != nil {
		return "", err
	}

	offering, err := w.offering(logger, ch.Offering)
	if err != nil {
		return "", err
	}

	key, err := data.ChannelKey(ch.Client, ch.Agent,
		ch.Block, offering.Hash)
	// internal
	if err != nil {
		logger.Add("channel", ch, "offering", offering).Error(err.Error())
		return "", ErrInternal
	}
	return data.FromBytes(key), nil
}

func (w *Worker) updateRelatedOffering(job *data.Job, jobType, status string) error {
	logger := w.logger.Add("method", "updateRelatedOffering", "job", job)
	offering, err := w.relatedOffering(logger, job, jobType)
	if err != nil {
		return err
	}

	offering.OfferStatus = status

	// internal
	return w.saveRecord(logger, offering)
}

func (w *Worker) saveRecord(logger log.Logger, rec reform.Record) error {
	err := data.Save(w.db.Querier, rec)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}
	return nil
}
