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

	initHeader := types.Header{
		Height: 1,
		Time:   time.Date(2020, 2, 1, 12, 30, 0, 0, time.UTC),
	}

	ctx := app.NewContext(false, initHeader)

	// First Block
	app.BeginBlock(
		types.RequestBeginBlock{Header: ctx.BlockHeader()},
	)
	app.EndBlock(types.RequestEndBlock{Height: initHeader.Height})

	secondBlockHeader := types.Header{Height: 2, Time: time.Date(2020, 2, 1, 12, 30, 0, 0, time.UTC).Add(time.Second * 5)}
	// Second Block
	ctx.WithBlockHeader(secondBlockHeader)

	app.BeginBlock(
		types.RequestBeginBlock{Header: ctx.BlockHeader()},
	)

	minter := app.MintKeeper.GetMinter(ctx)
	require.Equal(t, initHeader.Time, minter.GetLastBlockTimestamp())
}
