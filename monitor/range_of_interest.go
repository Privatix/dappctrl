package monitor

import (
	"context"

	"github.com/privatix/dappctrl/data"
)

func (m *Monitor) rangeOfInterest(
	ctx context.Context) (first uint64, last uint64, err error) {
	logger := m.logger.Add("method", "rangeOfInterest")

	unreliableNum, err := data.GetUint64Setting(m.db, data.SettingMinConfirmations)
	if err != nil {
		logger.Add("setting", data.SettingMinConfirmations).Error(err.Error())
		return 0, 0, err
	}

	freshNum, err := data.GetUint64Setting(m.db, data.SettingFreshBlocks)
	if err != nil {
		logger.Add("setting", data.SettingFreshBlocks).Error(err.Error())
		return 0, 0, err
	}

	first, err = m.getLastProcessedBlockNumber()
	if err != nil {
		return 0, 0, err
	}

	first = first + 1

	latestBlock, err := m.getLatestBlockNumber(ctx)
	if err != nil {
		return 0, 0, err
	}

	last = safeSub(latestBlock, unreliableNum)

	if freshNum != 0 {
		first = max(first, safeSub(last, freshNum))
	}

	limitNum, err := data.GetUint64Setting(m.db, data.SettingBlockLimit)
	if err != nil {
		m.logger.Add("setting", data.SettingBlockLimit).Warn(err.Error())
		return first, last, nil
	}

	if limitNum != 0 && last > first && (last-first) > limitNum {
		last = first + limitNum
	}

	return first, last, nil
}

func (m *Monitor) getLastProcessedBlockNumber() (uint64, error) {
	v, err := data.GetUint64Setting(m.db, data.SettingLastProcessedBlock)
	if err != nil {
		m.logger.Error(err.Error())
		return 0, err
	}
	return v, nil
}

func (m *Monitor) getLatestBlockNumber(ctx context.Context) (uint64, error) {
	header, err := m.eth.HeaderByNumber(ctx, nil)
	if err != nil {
		m.logger.Error(err.Error())
		return 0, ErrFailedToGetHeaderByNumber
	}

	return header.Number.Uint64(), err
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
