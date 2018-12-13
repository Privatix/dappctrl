package looper

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/eth/contract"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

var (
	conf struct {
		DB  *data.DBConfig
		Eth *eth.Config
		Job *job.Config
		Log *log.WriterConfig
	}
	logger             log.Logger
	db                 *reform.DB
	serviceContractABI abi.ABI
	ethBackend         *eth.TestEthBackend
)

func createJob() *data.Job {
	return &data.Job{
		Type:        data.JobAgentPreOfferingPopUp,
		RelatedID:   util.NewUUID(),
		RelatedType: data.JobOffering,
		CreatedBy:   data.JobUser,
		CreatedAt:   time.Now(),
		Data:        []byte("{}"),
	}
}

func expectedJobs(t *testing.T, iterationCh chan time.Time,
	doneCh chan struct{}, relatedID string, exp int) []reform.Struct {
	iterationCh <- time.Now()
	<-doneCh

	res, err := db.FindAllFrom(
		data.JobTable, "related_id", relatedID)
	if err != nil {
		t.Fatal(err)
	}

	if len(res) != exp {
		t.Fatalf("expected jobs numder:%d, got: %d",
			exp, len(res))
	}
	return res
}

func TestLoop(t *testing.T) {
	j := createJob()

	f := func() []*data.Job {
		return []*data.Job{j}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	queue := job.NewQueue(conf.Job, logger, db, nil)
	defer queue.Close()

	c := make(chan time.Time)
	tik := &time.Ticker{C: c}
	done := make(chan struct{})

	go loop(ctx, tik, db, queue, f, logger, done)

	jobs := expectedJobs(t, c, done, j.RelatedID, 1)
	jobFromBD := jobs[0].(*data.Job)
	defer db.Delete(jobFromBD)

	expectedJobs(t, c, done, j.RelatedID, 1)

	jobFromBD.Status = data.JobDone
	err := db.Save(jobFromBD)
	if err != nil {
		t.Fatal(err)
	}

	jobs = expectedJobs(t, c, done, j.RelatedID, 2)
	for _, j := range jobs {
		err = db.Delete(j.(*data.Job))
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestMain(m *testing.M) {
	conf.DB = data.NewDBConfig()
	conf.Eth = eth.NewConfig()
	conf.Job = job.NewConfig()
	conf.Log = log.NewWriterConfig()
	args := &util.TestArgs{
		Conf: &conf,
	}
	util.ReadTestArgs(args)

	var err error
	logger, err = log.NewTestLogger(conf.Log, args.Verbose)
	if err != nil {
		panic(err)
	}

	db = data.NewTestDB(conf.DB)

	serviceContractABI, err = abi.JSON(strings.NewReader(
		contract.PrivatixServiceContractABI))
	if err != nil {
		panic(err)
	}

	ethBackend = eth.NewTestEthBackend(common.HexToAddress("0x1"))

	os.Exit(m.Run())
}
