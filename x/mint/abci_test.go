package mint_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/simapp"
)

func TestMintCalcBlocksYear(t *testing.T) {
	app := simapp.Setup(false)

	firstBlockTime := time.Date(2020, 2, 1, 12, 30, 0, 0, time.UTC)
	initHeader := types.Header{
		Height: 1,
		Time:   firstBlockTime,
	}

	// First Block
	app.BeginBlock(
		types.RequestBeginBlock{Header: initHeader},
	)
	ctx := app.NewContext(false, types.Header{})
	// assert block time
	minter := app.MintKeeper.GetMinter(ctx)
	require.Equal(t, 5*time.Second, minter.AverageBlockTime)
	require.Equal(t, initHeader.Time, minter.GetLastBlockTimestamp())
	app.Commit()

	// Second Block after 8 seconds.
	secondBlockHeader := types.Header{Height: 2, Time: firstBlockTime.Add(time.Second * 8)}
	app.BeginBlock(
		types.RequestBeginBlock{Header: secondBlockHeader},
	)

	// Then Average should be 5 + 8 / 2 (blockTime1 + blockTime2 / height))
	ctx = app.NewContext(false, types.Header{})
	minter = app.MintKeeper.GetMinter(ctx)
	require.Equal(t, secondBlockHeader.Time, minter.GetLastBlockTimestamp())
	require.Equal(t, (5*time.Second+8*time.Second)/2, minter.AverageBlockTime)
}
