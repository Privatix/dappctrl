package worker

import (
	"crypto/ecdsa"
	"encoding/json"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/privatix/dappctrl/data"
	ethutil "github.com/privatix/dappctrl/eth/util"
	"github.com/privatix/dappctrl/util"
)

func (w *Worker) key(key string) (*ecdsa.PrivateKey, error) {
	return w.decryptKeyFunc(key, w.pwdGetter.Get())
}

func (w *Worker) toHashArr(h string) (ret [common.HashLength]byte, err error) {
	var hash common.Hash
	hash, err = data.ToHash(h)
	if err != nil {
		return
	}
	ret = [common.HashLength]byte(hash)
	return
}

func (w *Worker) balanceData(job *data.Job) (*data.JobBalanceData, error) {
	balanceData := &data.JobBalanceData{}
	if err := json.Unmarshal(job.Data, balanceData); err != nil {
		return nil, err
	}
	return balanceData, nil
}

func (w *Worker) ethLogTx(job *data.Job) (*types.Transaction, error) {
	ethLog, err := w.ethLog(job)
	if err != nil {
		return nil, err
	}

	hash, err := data.ToHash(ethLog.TxHash)
	if err != nil {
		return nil, err
	}

	return w.getTransaction(hash)
}

func (w *Worker) newUser(tx *types.Transaction) (*data.User, error) {
	signer := &types.HomesteadSigner{}
	pubkey, err := ethutil.RecoverPubKey(signer, tx)
	if err != nil {
		return nil, err
	}

	addr := crypto.PubkeyToAddress(*pubkey)

	return &data.User{
		ID:        util.NewUUID(),
		EthAddr:   data.FromBytes(addr.Bytes()),
		PublicKey: data.FromBytes(crypto.FromECDSAPub(pubkey)),
	}, nil
}

func (w *Worker) addJob(jType, rType, rID string) error {
	return w.queue.Add(&data.Job{
		ID:          util.NewUUID(),
		Status:      data.JobActive,
		RelatedType: rType,
		RelatedID:   rID,
		Type:        jType,
		CreatedAt:   time.Now(),
		CreatedBy:   data.JobTask,
		Data:        []byte("{}"),
	})
}

func (w *Worker) updateAccountBalances(job *data.Job, jobType string) error {
	acc, err := w.relatedAccount(job, jobType)
	if err != nil {
		return err
	}

	agentAddr, err := data.ToAddress(acc.EthAddr)
	if err != nil {
		return err
	}

	amount, err := w.ethBack.PTCBalanceOf(&bind.CallOpts{}, agentAddr)
	if err != nil {
		return err
	}

	acc.PTCBalance = amount.Uint64()

	amount, err = w.ethBack.PSCBalanceOf(&bind.CallOpts{}, agentAddr)
	if err != nil {
		return err
	}

	acc.PSCBalance = amount.Uint64()

	return w.db.Update(acc)
}
