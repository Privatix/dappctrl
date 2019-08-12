package bc

import (
	"fmt"
	"testing"

	"github.com/privatix/dappctrl/data"
)

func blockSettings(t *testing.T, limit,
	last, confirmations uint64) func() {
	settings := []*data.Setting{
		{
			Key:   data.SettingBlockLimit,
			Name:  "block limit",
			Value: fmt.Sprint(limit),
		},
		{
			Key:   data.SettingLastProcessedBlock,
			Name:  "last scanned block",
			Value: fmt.Sprint(last),
		},
		{
			Key:   data.SettingMinConfirmations,
			Name:  "min confirmations",
			Value: fmt.Sprint(confirmations),
		},
	}

	data.InsertToTestDB(t, db, settings[0], settings[1], settings[2])
	return func() {
		defer data.DeleteFromTestDB(t, db, settings[0], settings[1], settings[2])
	}
}

func offeringsSearchBlockSettings(t *testing.T, limit, fresh, lastFrom,
	confirmations uint64) func() {
	settings := []*data.Setting{
		{
			Key:   data.SettingBlockLimit,
			Name:  "block limit",
			Value: fmt.Sprint(limit),
		},
		{
			Key:   data.SettingOfferingsFreshBlocks,
			Name:  "offering's fresh blocks",
			Value: fmt.Sprint(fresh),
		},
		{
			Key:   data.SettingLastBackSearchBlock,
			Name:  "offering's last from",
			Value: fmt.Sprint(lastFrom),
		},
		{
			Key:   data.SettingMinConfirmations,
			Name:  "min confirmations",
			Value: fmt.Sprint(confirmations),
		},
	}

	data.InsertToTestDB(t, db, settings[0], settings[1], settings[2],
		settings[3])
	return func() {
		defer data.DeleteFromTestDB(t, db, settings[0], settings[1], settings[2],
			settings[3])
	}
}

type rangeOfInterestTestCase struct {
	blockLimit       uint64
	lastProcessed    uint64
	latestBlock      uint64
	minConfirmations uint64
	expectedFrom     uint64
	expectedTo       uint64
}

type offeringsRangeOfInterestTestCase struct {
	blockLimit           uint64
	offeringsFreshBlocks uint64
	lastSearchFrom       uint64
	latestBlock          uint64
	minConfirmations     uint64
	expectedFrom         uint64
	expectedTo           uint64
}

func TestRangeOfInterest(t *testing.T) {
	for _, testcase := range []rangeOfInterestTestCase{
		{
			blockLimit:       10,
			expectedFrom:     501,
			expectedTo:       511,
			lastProcessed:    500,
			latestBlock:      1000,
			minConfirmations: 10,
		},
		{
			blockLimit:       10,
			expectedFrom:     951,
			expectedTo:       961,
			lastProcessed:    950,
			latestBlock:      1000,
			minConfirmations: 10,
		},
		{
			blockLimit:       10,
			expectedFrom:     501,
			expectedTo:       511,
			lastProcessed:    500,
			latestBlock:      1000,
			minConfirmations: 10,
		},
	} {
		testRangeOfInterest(t, testcase)
	}
}

func TestOfferingsRangeOfInterest(t *testing.T) {
	for _, tc := range []offeringsRangeOfInterestTestCase{
		{
			blockLimit:           10,
			offeringsFreshBlocks: 510,
			lastSearchFrom:       500,
			latestBlock:          1013,
			minConfirmations:     3,
			expectedFrom:         0,
			expectedTo:           0,
		},
		{
			blockLimit:           10,
			offeringsFreshBlocks: 600,
			lastSearchFrom:       500,
			latestBlock:          1003,
			minConfirmations:     3,
			expectedFrom:         489,
			expectedTo:           499,
		},
		{
			blockLimit:           10,
			offeringsFreshBlocks: 600,
			lastSearchFrom:       0,
			latestBlock:          1003,
			minConfirmations:     3,
			expectedFrom:         990,
			expectedTo:           1000,
		},
		{
			blockLimit:           10,
			offeringsFreshBlocks: 500,
			lastSearchFrom:       505,
			latestBlock:          1003,
			minConfirmations:     3,
			expectedFrom:         500,
			expectedTo:           504,
		},
	} {
		func() {
			cleanup := offeringsSearchBlockSettings(t, tc.blockLimit, tc.offeringsFreshBlocks,
				tc.lastSearchFrom, tc.minConfirmations)
			defer cleanup()
			from, to, err := offeringsRangeOfInterest(db, tc.latestBlock)
			if err != nil {
				t.Fatal(err)
			}
			if from != tc.expectedFrom {
				t.Fatalf("wanted from: %v, got: %v", tc.expectedFrom, from)
			}
			if to != tc.expectedTo {
				t.Fatalf("wanted to: %v, got: %v", tc.expectedTo, to)
			}
		}()
	}
}

func testRangeOfInterest(t *testing.T, testcase rangeOfInterestTestCase) {
	cleanup := blockSettings(t, testcase.blockLimit, testcase.lastProcessed,
		testcase.minConfirmations)
	defer cleanup()

	from, to, _ := rangeOfInterest(db, testcase.latestBlock)
	if testcase.expectedFrom != from {
		t.Fatalf("wanted from block: %d, got: %d",
			testcase.expectedFrom, from)
	}

	if testcase.expectedTo != to {
		t.Fatalf("wanted to block: %d, got: %d",
			testcase.expectedTo, to)
	}
}
