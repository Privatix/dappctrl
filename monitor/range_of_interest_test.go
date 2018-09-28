package monitor_test

import (
	"context"
	"testing"
)

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

type rangeOfInterestTestCase struct {
	blockLimit       uint64
	freshBlocks      uint64
	lastProcessed    uint64
	latestBlock      uint64
	minConfirmations uint64
	expectedFrom     uint64
	expectedTo       uint64
}

func TestRangeOfInterest(t *testing.T) {
	for _, testcase := range []rangeOfInterestTestCase{
		{
			blockLimit:       10,
			expectedFrom:     890,
			expectedTo:       900,
			freshBlocks:      100,
			lastProcessed:    500,
			latestBlock:      1000,
			minConfirmations: 10,
		},
		{
			blockLimit:       10,
			expectedFrom:     951,
			expectedTo:       961,
			freshBlocks:      100,
			lastProcessed:    950,
			latestBlock:      1000,
			minConfirmations: 10,
		},
		{
			blockLimit:       10,
			expectedFrom:     501,
			expectedTo:       511,
			freshBlocks:      1000,
			lastProcessed:    500,
			latestBlock:      1000,
			minConfirmations: 10,
		},
	} {
		testRangeOfInterest(t, testcase)
	}
}

func testRangeOfInterest(t *testing.T, testcase rangeOfInterestTestCase) {
	cleanup := blockSettings(t, testcase.freshBlocks,
		testcase.blockLimit, testcase.lastProcessed,
		testcase.minConfirmations)
	defer cleanup()

	ethClient.HeaderByNumberResult = testcase.latestBlock
	from, to, _ := mon.RangeOfInterest(context.Background())
	if testcase.expectedFrom != from {
		t.Fatalf("wanted from block: %d, got: %d",
			testcase.expectedFrom, from)
	}

	if testcase.expectedTo != to {
		t.Fatalf("wanted to block: %d, got: %d",
			testcase.expectedTo, to)
	}
}
