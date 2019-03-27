package worker

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

func TestPreAccountAddBalanceApprove(t *testing.T) {
	// check PTC balance PTC.balanceOf()
	// PTC.increaseApproval()
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobPreAccountAddBalanceApprove,
		data.JobAccount)
	defer env.close()
	defer fixture.close()

	var transferAmount int64 = 10

	setJobData(t, fixture.DB, fixture.job, data.JobBalanceData{
		Amount: uint64(transferAmount),
	})

	env.ethBack.BalancePTC = big.NewInt(transferAmount)
	env.ethBack.BalanceEth = big.NewInt(999999)

	runJob(t, env.worker.PreAccountAddBalanceApprove, fixture.job)

	agentAddr := data.TestToAddress(t, fixture.Account.EthAddr)

	env.ethBack.TestCalled(t, "PTCIncreaseApproval", agentAddr,
		env.gasConf.PTC.Approve,
		conf.pscAddr,
		big.NewInt(transferAmount))

	noCallerAddr := common.BytesToAddress([]byte{})
	env.ethBack.TestCalled(t, "PTCBalanceOf", noCallerAddr, 0,
		agentAddr)

	// Test eth transaction was recorded.
	env.deleteEthTx(t, fixture.job.ID)

	testCommonErrors(t, env.worker.PreAccountAddBalanceApprove,
		*fixture.job)
}

func TestPreAccountAddBalance(t *testing.T) {
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobPreAccountAddBalance,
		data.JobAccount)
	defer env.close()
	defer fixture.close()

	transferAmount := int64(10)
	approveJob := data.NewTestJob(data.JobPreAccountAddBalanceApprove,
		data.JobUser, data.JobAccount)
	approveJob.RelatedID = fixture.Account.ID
	env.insertToTestDB(t, approveJob)
	setJobData(t, fixture.DB, approveJob, &data.JobBalanceData{
		Amount: uint64(transferAmount),
	})
	defer env.deleteFromTestDB(t, approveJob)

	txHash := data.HexString("d238f7")

	setJobData(t, db, fixture.job, &data.JobData{
		EthLog: &data.JobEthLog{
			TxHash: txHash,
		},
	})

	tx := &data.EthTx{
		ID:          util.NewUUID(),
		JobID:       &approveJob.ID,
		Hash:        txHash,
		RelatedType: data.JobAccount,
		RelatedID:   fixture.Account.ID,
		Status:      data.TxSent,
		Gas:         1,
		GasPrice:    uint64(1),
		TxRaw:       []byte("{}"),
	}
	env.insertToTestDB(t, tx)

	defer env.deleteFromTestDB(t, tx)

	runJob(t, env.worker.PreAccountAddBalance, fixture.job)

	agentAddr := data.TestToAddress(t, fixture.Account.EthAddr)

	env.ethBack.TestCalled(t, "PSCAddBalanceERC20", agentAddr,
		env.gasConf.PSC.AddBalanceERC20, uint64(transferAmount))

	// Test eth transaction was recorded.
	env.deleteEthTx(t, fixture.job.ID)

	testCommonErrors(t, env.worker.PreAccountAddBalance, *fixture.job)
}

func TestPreAccountReturnBalance(t *testing.T) {
	// check PSC balance PSC.balanceOf()
	// PSC.returnBalanceERC20()
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobPreAccountReturnBalance,
		data.JobAccount)
	defer env.close()
	defer fixture.close()

	var amount int64 = 10

	setJobData(t, fixture.DB, fixture.job, &data.JobBalanceData{
		Amount: uint64(amount),
	})

	env.ethBack.BalancePSC = uint64(amount)
	env.ethBack.BalanceEth = big.NewInt(999999)

	runJob(t, env.worker.PreAccountReturnBalance, fixture.job)

	agentAddr := data.TestToAddress(t, fixture.Account.EthAddr)

	noCallerAddr := common.BytesToAddress([]byte{})
	env.ethBack.TestCalled(t, "PSCBalanceOf", noCallerAddr, 0, agentAddr)

	env.ethBack.TestCalled(t, "PSCReturnBalanceERC20", agentAddr,
		env.gasConf.PSC.ReturnBalanceERC20, uint64(amount))

	// Test eth transaction was recorded.
	env.deleteEthTx(t, fixture.job.ID)

	testCommonErrors(t, env.worker.PreAccountReturnBalance, *fixture.job)
}

func TestAfterAccountAddBalance(t *testing.T) {
	// update balance in DB.accounts.ptc_balance
	env := newWorkerTest(t)
	defer env.close()
	testAccountBalancesUpdate(t, env, env.worker.AfterAccountAddBalance,
		data.JobAfterAccountAddBalance)
}

func TestAfterAccountReturnBalance(t *testing.T) {
	// Test update balance in DB.accounts.psc_balance
	env := newWorkerTest(t)
	defer env.close()
	testAccountBalancesUpdate(t, env,
		env.worker.AfterAccountReturnBalance,
		data.JobAfterAccountReturnBalance)
}

func testAccountBalancesUpdate(t *testing.T, env *workerTest,
	worker func(*data.Job) error, jobType string) {
	// update balances in DB.accounts.psc_balance and DB.account.ptc_balance

	fixture := env.newTestFixture(t, jobType, data.JobAccount)
	defer fixture.close()

	env.ethBack.BalanceEth = big.NewInt(2)
	env.ethBack.BalancePTC = big.NewInt(100)
	env.ethBack.BalancePSC = 200

	runJob(t, worker, fixture.job)

	checkBalances(t, env, fixture.Account.ID, 200, 100, 2)

	testCommonErrors(t, worker, *fixture.job)
}

func checkBalances(t *testing.T, env *workerTest, accID string, psc, ptc, eth uint) {
	account := &data.Account{}
	env.findTo(t, account, accID)
	if account.PTCBalance != 100 {
		t.Fatalf("wrong ptc balance, wanted: %v, got: %v", 100,
			account.PTCBalance)
	}
	if account.PSCBalance != 200 {
		t.Fatalf("wrong psc balance, wanted: %v, got: %v", 200,
			account.PSCBalance)
	}
	if strings.TrimSpace(string(account.EthBalance)) !=
		string(data.FromBytes(env.ethBack.BalanceEth.Bytes())) {
		t.Logf("%v!=%v", string(account.EthBalance),
			data.FromBytes(env.ethBack.BalanceEth.Bytes()))
		t.Fatal("wrong eth balance")
	}
}

func testAfterChannelTopUp(t *testing.T, agent bool) {
	var jobType string

	if agent {
		jobType = data.JobAgentAfterChannelTopUp
	} else {
		jobType = data.JobClientAfterChannelTopUp
	}

	// Add deposit to channels.total_deposit
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, jobType,
		data.JobChannel)
	defer env.close()
	defer fixture.close()

	block := fixture.Channel.Block
	addedDeposit := big.NewInt(1)

	eventData, err := logChannelTopUpDataArguments.Pack(block, addedDeposit)
	if err != nil {
		t.Fatal(err)
	}

	agentAddr := data.TestToAddress(t, fixture.Channel.Agent)
	clientAddr := data.TestToAddress(t, fixture.Channel.Client)
	offeringHash := data.TestToHash(t, fixture.Offering.Hash)
	topics := data.LogTopics{
		// First topic is always the event.
		common.BytesToHash([]byte{}),
		common.BytesToHash(agentAddr.Bytes()),
		common.BytesToHash(clientAddr.Bytes()),
		offeringHash,
	}
	if err != nil {
		t.Fatal(err)
	}

	setJobData(t, db, fixture.job, &data.JobData{
		EthLog: &data.JobEthLog{
			Data:   eventData,
			Topics: topics,
		},
	})

	var job func(*data.Job) error

	if agent {
		job = env.worker.AgentAfterChannelTopUp
	} else {
		job = env.worker.ClientAfterChannelTopUp
	}
	runJob(t, job, fixture.job)

	channel := new(data.Channel)
	env.findTo(t, channel, fixture.Channel.ID)

	diff := channel.TotalDeposit - fixture.Channel.TotalDeposit
	if diff != addedDeposit.Uint64() {
		t.Fatal("total deposit not updated")
	}

	if !agent {
		env.deleteJob(t, data.JobAccountUpdateBalances,
			data.JobAccount, fixture.UserAcc.ID)
	}

	testCommonErrors(t, job, *fixture.job)
}

func TestDecrementCurrentSupply(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()
	fxt := env.newTestFixture(t,
		data.JobDecrementCurrentSupply, data.JobOffering)
	defer fxt.close()

	runJob(t, env.worker.DecrementCurrentSupply, fxt.job)

	offering := &data.Offering{}
	env.findTo(t, offering, fxt.Offering.ID)

	if offering.CurrentSupply+1 != fxt.Offering.CurrentSupply {
		t.Fatal("offering's current supply was not decremented")
	}

	testCommonErrors(t, env.worker.DecrementCurrentSupply, *fxt.job)
}

func TestIncrementCurrentSupply(t *testing.T) {
	env := newWorkerTest(t)
	defer env.close()
	fxt := env.newTestFixture(t,
		data.JobIncrementCurrentSupply, data.JobOffering)
	defer fxt.close()

	fxt.Offering.CurrentSupply--
	env.updateInTestDB(t, fxt.Offering)

	runJob(t, env.worker.IncrementCurrentSupply, fxt.job)

	offering := &data.Offering{}
	env.findTo(t, offering, fxt.Offering.ID)

	if offering.CurrentSupply-1 != fxt.Offering.CurrentSupply {
		t.Fatal("offering's current supply was not incremented")
	}

	testCommonErrors(t, env.worker.IncrementCurrentSupply, *fxt.job)
}
