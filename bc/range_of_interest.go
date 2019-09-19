package bc

import (
	"github.com/privatix/dappctrl/data"
	reform "gopkg.in/reform.v1"
)

func rangeOfInterest(db *reform.DB, latestBlock uint64) (first uint64, last uint64, err error) {
	unreliableNum, err := data.GetUint64Setting(db, data.SettingMinConfirmations)
	if err != nil {
		return 0, 0, err
	}

	limitNum, err := data.GetUint64Setting(db, data.SettingBlockLimit)
	if err != nil {
		return first, last, nil
	}

	first, err = data.GetUint64Setting(db, data.SettingLastProcessedBlock)
	if err != nil {
		return 0, 0, err
	}

	if first == 0 {
		first = safeSub(safeSub(latestBlock, unreliableNum), limitNum)
	}

	first = first + 1

	last = safeSub(latestBlock, unreliableNum)

	if limitNum != 0 && last > first && (last-first) > limitNum {
		last = first + limitNum
	}

	return first, last, nil
}

func offeringsRangeOfInterest(db *reform.DB, upBlock uint64) (uint64, uint64, error) {
	unreliableNum, err := data.GetUint64Setting(db, data.SettingMinConfirmations)
	if err != nil {
		return 0, 0, err
	}

	limitNum, err := data.GetUint64Setting(db, data.SettingBlockLimit)
	if err != nil {
		return 0, 0, err
	}

	freshNum, err := data.GetUint64Setting(db, data.SettingOfferingsFreshBlocks)
	if err != nil {
		return 0, 0, err
	}

	lastFrom, err := data.GetUint64Setting(db, data.SettingLastBackSearchBlock)
	if err != nil {
		return 0, 0, err
	}

	if lastFrom == 0 {
		lastFrom = safeSub(upBlock, unreliableNum) + 1
	}

	from := safeSub(safeSub(lastFrom, limitNum), 1)
	last := safeSub(lastFrom, 1)

	if last <= safeSub(upBlock, freshNum) {
		return 0, 0, nil
	}

	if from < safeSub(upBlock, freshNum) {
		from = safeSub(safeSub(upBlock, freshNum), unreliableNum)
	}

	if last <= from {
		return 0, 0, nil
	}

	return from, last, nil
}

func safeSub(a, b uint64) uint64 {
	if a > b {
		return a - b
	}
	return 0
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}
