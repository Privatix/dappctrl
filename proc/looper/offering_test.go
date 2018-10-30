package looper

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/AlekSi/pointer"

	"github.com/privatix/dappctrl/data"
)

func expectedResult(t *testing.T, exp int,
	timeNowFunc func() time.Time) []*data.Job {
	jobs := AutoOfferingPopUp(logger, serviceContractABI, db, ethBackend,
		timeNowFunc, conf.Eth.Contract.Periods.PopUp)

	if len(jobs) != exp {
		t.Fatalf("the right amount of jobs: %d,"+
			" got %d", exp, len(jobs))
	}
	return jobs
}

func TestAutoOfferingPopUp(t *testing.T) {
	jobs := expectedResult(t, 0, time.Now)

	fxt := data.NewTestFixture(t, db)
	defer fxt.Close()

	autoPopUpSetting := &data.Setting{
		Key:   data.SettingOfferingAutoPopUp,
		Value: "true",
		Name:  "autopopup",
	}

	data.InsertToTestDB(t, db, autoPopUpSetting)
	defer data.DeleteFromTestDB(t, db, autoPopUpSetting)

	// Setting offering.autopopup not initialized.
	jobs = expectedResult(t, 0, time.Now)

	fxt.Offering.OfferStatus = data.OfferRegistered
	fxt.Offering.AutoPopUp = pointer.ToBool(true)
	data.SaveToTestDB(t, db, fxt.Offering)

	ethBackend.GasPrice = big.NewInt(1)
	ethBackend.EstimatedGas = 2
	ethBackend.OfferUpdateBlockNumber = 3
	ethBackend.BlockNumber = big.NewInt(6)

	// Not enough ETH.
	jobs = expectedResult(t, 0, time.Now)

	ethBackend.BalanceEth = big.NewInt(100)

	timePoint := time.Now()
	timeNowFunc := func() time.Time {
		return timePoint
	}

	jobs = expectedResult(t, 1, timeNowFunc)

	resultJob := jobs[0]

	if resultJob.Type != data.JobAgentPreOfferingPopUp {
		t.Fatalf("wrong Type, expected: %s, got: %s",
			data.JobAgentPreOfferingPopUp, resultJob.Type)
	}

	if resultJob.RelatedType != data.JobOffering {
		t.Fatalf("wrong RelatedType, expected: %s, got: %s",
			data.JobOffering, resultJob.RelatedType)
	}

	if resultJob.RelatedID != fxt.Offering.ID {
		t.Fatalf("wrong RelatedID, expected: %s, got: %s",
			fxt.Offering.ID, resultJob.RelatedID)
	}

	if resultJob.CreatedBy != data.JobUser {
		t.Fatalf("wrong CreatedBy, expected: %s, got: %s",
			data.JobUser, resultJob.CreatedBy)
	}

	var jobData data.JobPublishData
	err := json.Unmarshal(resultJob.Data, &jobData)
	if err != nil {
		t.Fatal(err)
	}

	if jobData.GasPrice != ethBackend.GasPrice.Uint64() {
		t.Fatalf("wrong GasPrice, expected: %d, got: %d",
			ethBackend.GasPrice.Uint64(), jobData.GasPrice)
	}

	// popUpBlock = offering.lastUpdateBlock(in blockchain) + popUpPeriod
	// delayBlocks = lastBlock - popUpBlock
	// Time to create 2 blocks is  30 sec.
	expectedTime := timePoint.Add(30 * time.Second)
	if !resultJob.NotBefore.Equal(expectedTime) {
		t.Fatalf("wrong NotBefore time, expected: %s, got: %s",
			expectedTime.String(), resultJob.NotBefore.String())
	}
}
