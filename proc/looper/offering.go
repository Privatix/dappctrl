package looper

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/proc/adapter"
	"github.com/privatix/dappctrl/util/log"
)

// AutoOfferingPopUp creates AgentPreOfferingPopUp jobs for active offerings.
// Taken into account a current balance of agent ETH. If ETH is not enough to
// pop up all offerings, then no one of them is popped up. Function calculates
// the pop up time.
func AutoOfferingPopUp(logger log.Logger, abi abi.ABI, db *reform.DB,
	ethBack adapter.EthBackend, timeNowFunc func() time.Time,
	period uint32) []*data.Job {
	logger = logger.Add("method", "AutoOfferingPopUp")

	logger.Debug("started AutoOfferingPopUp")
	do, err := data.ReadBoolSetting(
		db.Querier, data.SettingOfferingAutoPopUp)
	if err != nil {
		logger.Warn(err.Error())
	}

	var jobs []*data.Job

	if do {
		jobs = autoOfferingPopUp(logger, abi, db, ethBack,
			timeNowFunc, period)
		logger.Debug(fmt.Sprintf("found %d offerings to pop upped",
			len(jobs)))
	}

	return jobs
}

func calcDelayToOfferingPopUp(offeringHash common.Hash,
	ethBack adapter.EthBackend, popUpPeriod uint32,
	lastBlock *big.Int) (delay time.Duration, err error) {
	_, _, _, _, lastUpdateBlock, _, err :=
		ethBack.PSCGetOfferingInfo(&bind.CallOpts{}, offeringHash)
	if err != nil {
		return 0, err
	}

	popUpBlock := uint64(lastUpdateBlock + popUpPeriod)

	if popUpBlock > lastBlock.Uint64() {
		delayBlocks := popUpBlock - lastBlock.Uint64()
		delay = time.Duration(delayBlocks) * BlockTime
	}

	return delay, err
}

func calcPriceToOfferingPopUp(abi abi.ABI, ethBack adapter.EthBackend,
	gasPrice *big.Int) *big.Int {
	input, err := abi.Pack("popupServiceOffering", common.Hash{})
	if err != nil {
		return nil
	}

	msg := ethereum.CallMsg{From: common.Address{},
		To: &common.Address{}, Data: input}

	gas, err := ethBack.EstimateGas(context.Background(), msg)
	if err != nil {
		return nil
	}

	// popupPrice = gas * gasPrice
	return new(big.Int).Mul(new(big.Int).SetUint64(gas), gasPrice)
}

func agentHaveEnoughMoney(logger log.Logger, agent common.Address,
	price *big.Int, ethBack adapter.EthBackend) bool {
	balance, err := ethBack.EthBalanceAt(context.Background(), agent)
	if err != nil {
		logger.Error(err.Error())
		return false
	}
	if balance.Cmp(price) == -1 {
		logger.Warn(fmt.Sprintf("not enough ETH: available %s,"+
			" necessary %s", balance.String(), price.String()))
		return false
	}
	return true
}

func jobOfferingPopUpData(gasPrice *big.Int) ([]byte, error) {
	jobData := &data.JobPublishData{GasPrice: gasPrice.Uint64()}
	raw, err := json.Marshal(jobData)
	if err != nil {
		return nil, err
	}
	return raw, err
}

func findOfferingsToPopUp(logger log.Logger, db *reform.DB) []reform.Struct {
	offerings, err := db.SelectAllFrom(data.OfferingTable,
		`WHERE offer_status in ('registered', 'popped_up')
			AND agent IN (SELECT eth_addr FROM accounts)
			AND (SELECT in_use FROM accounts WHERE eth_addr = agent)
			AND auto_pop_up`)
	if err != nil {
		logger.Error(err.Error())
		return nil
	}
	return offerings
}

func autoOfferingPopUp(logger log.Logger, abi abi.ABI, db *reform.DB,
	ethBack adapter.EthBackend, timeNowFunc func() time.Time,
	period uint32) []*data.Job {
	offerings := findOfferingsToPopUp(logger, db)
	if len(offerings) == 0 {
		logger.Debug("no offerings to pop up")
		return nil
	}

	gasPrice, err := ethBack.SuggestGasPrice(context.Background())
	if err != nil {
		logger.Error(err.Error())
		return nil
	}

	popupPrice := calcPriceToOfferingPopUp(abi, ethBack, gasPrice)

	// price = popupPrice * len(offerings)
	price := new(big.Int).Mul(popupPrice, big.NewInt(int64(len(offerings))))

	// TODO(maxim) We assume that a agent (account) is one.
	agentAddr, err := data.HexToAddress(offerings[0].(*data.Offering).Agent)
	if err != nil {
		logger.Error(err.Error())
		return nil
	}

	if !agentHaveEnoughMoney(logger, agentAddr, price, ethBack) {
		return nil
	}

	jobData, err := jobOfferingPopUpData(gasPrice)
	if err != nil {
		logger.Error(err.Error())
		return nil
	}

	lastBlock, err := ethBack.LatestBlockNumber(context.Background())
	if err != nil {
		logger.Error(err.Error())
		return nil
	}

	var result []*data.Job
	for _, v := range offerings {
		hash, err := data.ToHash(v.(*data.Offering).Hash)
		if err != nil {
			logger.Error(err.Error())
			return nil
		}

		delay, err := calcDelayToOfferingPopUp(
			hash, ethBack, period, lastBlock)
		if err != nil {
			logger.Error(err.Error())
			return nil
		}

		j := &data.Job{
			Type:        data.JobAgentPreOfferingPopUp,
			RelatedType: data.JobOffering,
			RelatedID:   v.(*data.Offering).ID,
			CreatedBy:   data.JobUser,
			Data:        jobData,
			NotBefore:   timeNowFunc().Add(delay),
		}

		result = append(result, j)
	}
	return result
}
