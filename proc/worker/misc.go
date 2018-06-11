package worker

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/AlekSi/pointer"
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
		err = fmt.Errorf("unable to parse hash: %v", err)
		return
	}
	ret = [common.HashLength]byte(hash)
	return
}

func (w *Worker) balanceData(job *data.Job) (*data.JobBalanceData, error) {
	balanceData := &data.JobBalanceData{}
	if err := w.unmarshalDataTo(job.Data, balanceData); err != nil {
		return nil, err
	}
	return balanceData, nil
}

func (w *Worker) publishData(job *data.Job) (*data.JobPublishData, error) {
	publishData := &data.JobPublishData{}
	if err := w.unmarshalDataTo(job.Data, publishData); err != nil {
		return nil, err
	}
	return publishData, nil
}

func (w *Worker) unmarshalDataTo(jobData []byte, v interface{}) error {
	if err := json.Unmarshal(jobData, v); err != nil {
		return fmt.Errorf("could not unmarshal data to %T: %v", v, err)
	}
	return nil
}

func (w *Worker) ethLogTx(ethLog *data.EthLog) (*types.Transaction, error) {
	hash, err := data.ToHash(ethLog.TxHash)
	if err != nil {
		return nil, fmt.Errorf("could not decode eth tx hash: %v", err)
	}

	return w.getTransaction(hash)
}

func (w *Worker) newUser(tx *types.Transaction) (*data.User, error) {
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

func (w *Worker) updateAccountBalances(acc *data.Account) error {
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

	amount, err = w.ethBalance(agentAddr)
	if err != nil {
		return err
	}

	acc.EthBalance = data.B64BigInt(data.FromBytes(amount.Bytes()))

	return w.db.Update(acc)
}

func (w *Worker) ethBalance(addr common.Address) (*big.Int, error) {
	// TODO: move timeout to conf
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	amount, err := w.ethBack.EthBalanceAt(ctx, addr)
	if err != nil {
		return nil, fmt.Errorf("could not get eth balance: %v", err)
	}

	return amount, nil
}

func (w *Worker) saveEthTX(job *data.Job, tx *types.Transaction,
	method, relatedType, relatedID, from, to string) error {
	raw, err := tx.MarshalJSON()
	if err != nil {
		return err
	}

	dtx := data.EthTx{
		ID:          util.NewUUID(),
		Hash:        data.FromBytes(tx.Hash().Bytes()),
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

	return w.db.Insert(&dtx)
}
