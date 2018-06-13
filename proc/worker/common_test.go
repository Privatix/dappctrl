package worker

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/privatix/dappctrl/data"
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

	fixture.setJobData(t, data.JobBalanceData{
		Amount: uint(transferAmount),
	})

	env.ethBack.balancePTC = big.NewInt(transferAmount)
	env.ethBack.balanceEth = big.NewInt(999999)

	runJob(t, env.worker.PreAccountAddBalanceApprove, fixture.job)

	agentAddr := data.TestToAddress(t, fixture.Account.EthAddr)

	env.ethBack.testCalled(t, "PTCIncreaseApproval", agentAddr,
		env.gasConf.PTC.Approve,
		conf.pscAddr,
		big.NewInt(transferAmount))

	noCallerAddr := common.BytesToAddress([]byte{})
	env.ethBack.testCalled(t, "PTCBalanceOf", noCallerAddr, 0,
		agentAddr)

	// Test eth transaction was recorded.
	env.deleteEthTx(t, fixture.job.ID)

	testCommonErrors(t, env.worker.PreAccountAddBalanceApprove,
		*fixture.job)
}

func TestPreAccountAddBalance(t *testing.T) {
	// PSC.addBalanceERC20()
	env := newWorkerTest(t)
	fixture := env.newTestFixture(t, data.JobPreAccountAddBalance,
		data.JobAccount)
	defer env.close()
	defer fixture.close()

	var transferAmount int64 = 10

	fixture.setJobData(t, data.JobBalanceData{
		Amount: uint(transferAmount),
	})

	runJob(t, env.worker.PreAccountAddBalance, fixture.job)

	agentAddr := data.TestToAddress(t, fixture.Account.EthAddr)

	env.ethBack.testCalled(t, "PSCAddBalanceERC20", agentAddr,
		env.gasConf.PSC.AddBalanceERC20, big.NewInt(transferAmount))

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

	fixture.setJobData(t, &data.JobBalanceData{
		Amount: uint(amount),
	})

	env.ethBack.balancePSC = big.NewInt(amount)
	env.ethBack.balanceEth = big.NewInt(999999)

	runJob(t, env.worker.PreAccountReturnBalance, fixture.job)

	agentAddr := data.TestToAddress(t, fixture.Account.EthAddr)

	noCallerAddr := common.BytesToAddress([]byte{})
	env.ethBack.testCalled(t, "PSCBalanceOf", noCallerAddr, 0, agentAddr)

	env.ethBack.testCalled(t, "PSCReturnBalanceERC20", agentAddr,
		env.gasConf.PSC.ReturnBalanceERC20, big.NewInt(amount))

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

func TestAccountAddCheckBalancee(t *testing.T) {
	t.Skip("TODO")
	// env := newWorkerTest(t)
	// defer env.close()
	// testAccountBalancesUpdate(t, env, env.worker.AccountAddCheckBalance,
	// 	data.JobAccountAddCheckBalance)
}

func testAccountBalancesUpdate(t *testing.T, env *workerTest,
	worker func(*data.Job) error, jobType string) {
	// update balances in DB.accounts.psc_balance and DB.account.ptc_balance

	fixture := env.newTestFixture(t, jobType, data.JobAccount)
	defer fixture.close()

	env.ethBack.balanceEth = big.NewInt(2)
	env.ethBack.balancePTC = big.NewInt(100)
	env.ethBack.balancePSC = big.NewInt(200)

	runJob(t, worker, fixture.job)

	account := &data.Account{}
	env.findTo(t, account, fixture.Account.ID)
	if account.PTCBalance != 100 {
		t.Fatalf("wrong ptc balance, wanted: %v, got: %v", 100,
			account.PTCBalance)
	}
	if account.PSCBalance != 200 {
		t.Fatalf("wrong psc balance, wanted: %v, got: %v", 200,
			account.PSCBalance)
	}
	if strings.TrimSpace(string(account.EthBalance)) !=
		data.FromBytes(env.ethBack.balanceEth.Bytes()) {
		t.Logf("%v!=%v", string(account.EthBalance),
			data.FromBytes(env.ethBack.balanceEth.Bytes()))
		t.Fatal("wrong eth balance")
	}

	testCommonErrors(t, worker, *fixture.job)
}
