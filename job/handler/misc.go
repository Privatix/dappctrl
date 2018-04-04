package handler

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/privatix/dappctrl/data"
	ethutil "github.com/privatix/dappctrl/eth/util"
	"github.com/privatix/dappctrl/util"
)

func (w *Handler) key(key string) (*ecdsa.PrivateKey, error) {
	return w.decryptKeyFunc(key, w.pwdGetter.Get())
}

func (w *Handler) toHashArr(h string) (ret [common.HashLength]byte, err error) {
	var hash common.Hash
	hash, err = data.ToHash(h)
	if err != nil {
		err = fmt.Errorf("unable to parse hash: %v", err)
		return
	}
	ret = [common.HashLength]byte(hash)
	return
}

func (w *Handler) balanceData(job *data.Job) (*data.JobBalanceData, error) {
	balanceData := &data.JobBalanceData{}
	if err := json.Unmarshal(job.Data, balanceData); err != nil {
		return nil, fmt.Errorf("could not unmarshal data to %T: %v",
			balanceData, err)
	}
	return balanceData, nil
}

func (w *Handler) ethLogTx(ethLog *data.EthLog) (*types.Transaction, error) {
	hash, err := data.ToHash(ethLog.TxHash)
	if err != nil {
		return nil, fmt.Errorf("could not decode eth tx hash: %v", err)
	}

	return w.getTransaction(hash)
}

func (w *Handler) newUser(tx *types.Transaction) (*data.User, error) {
	signer := &types.HomesteadSigner{}
	pubkey, err := ethutil.RecoverPubKey(signer, tx)
	if err != nil {
		err = fmt.Errorf("could not recover client's pub key: %v", err)
		return nil, err
	}

	addr := crypto.PubkeyToAddress(*pubkey)

	return &data.User{
		ID:        util.NewUUID(),
		EthAddr:   data.FromBytes(addr.Bytes()),
		PublicKey: data.FromBytes(crypto.FromECDSAPub(pubkey)),
	}, nil
}

func (w *Handler) addJob(jType, rType, rID string) error {
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

func (w *Handler) updateAccountBalances(job *data.Job, jobType string) error {
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
		return fmt.Errorf("could not get ptc balance: %v", err)
	}

	acc.PTCBalance = amount.Uint64()

	amount, err = w.ethBack.PSCBalanceOf(&bind.CallOpts{}, agentAddr)
	if err != nil {
		return fmt.Errorf("could not get psc balance: %v", err)
	}

	acc.PSCBalance = amount.Uint64()

	return w.db.Update(acc)
}
