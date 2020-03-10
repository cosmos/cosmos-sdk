package mint_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/simapp"
)

func TestBlocksPerYearCalculation(t *testing.T) {
	app := simapp.Setup(false)

	firstBlockTime := time.Date(2020, 2, 1, 12, 30, 0, 0, time.UTC)
	initHeader := types.Header{
		Height: 1,
		Time:   firstBlockTime,
	}

	// LastBlockTimestamp is not set.
	ctx := app.NewContext(false, types.Header{})
	minter := app.MintKeeper.GetMinter(ctx)
	require.Equal(t, time.Time{}, minter.GetLastBlockTimestamp())

	// First Block
	app.BeginBlock(
		types.RequestBeginBlock{Header: initHeader},
	)
	ctx = app.NewContext(false, types.Header{})

	// just BlockTimeStamp is saved
	minter = app.MintKeeper.GetMinter(ctx)
	require.Equal(t, time.Duration(0), minter.AverageBlockTime)
	require.Equal(t, initHeader.Time, minter.GetLastBlockTimestamp())
	app.Commit()

	// Second Block after 8 seconds.
	secondBlockHeader := types.Header{Height: 2, Time: firstBlockTime.Add(time.Second * 8)}
	app.BeginBlock(
		types.RequestBeginBlock{Header: secondBlockHeader},
	)

	// Then Average should be 8 / 2 (blockTime2 / height-1))
	ctx = app.NewContext(false, types.Header{})
	minter = app.MintKeeper.GetMinter(ctx)
	require.Equal(t, secondBlockHeader.Time, minter.GetLastBlockTimestamp())
	require.Equal(t, (8*time.Second)/1, minter.AverageBlockTime)
	app.Commit()

	thirdBlockHeader := types.Header{Height: 3, Time: secondBlockHeader.Time.Add(time.Second * 4)}
	app.BeginBlock(
		types.RequestBeginBlock{Header: thirdBlockHeader},
	)
	ctx = app.NewContext(false, types.Header{})
	minter = app.MintKeeper.GetMinter(ctx)
	require.Equal(t, thirdBlockHeader.Time, minter.GetLastBlockTimestamp())
	require.Equal(t, (8*time.Second+time.Second*4)/2, minter.AverageBlockTime)
}
