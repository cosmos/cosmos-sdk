package keeper_test

import (
	"sort"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/simapp"

	"github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestHistoricalInfo(t *testing.T) {
	_, app, ctx := createTestInput()

	addrDels := simapp.AddTestAddrsIncremental(app, ctx, 50, sdk.NewInt(0))
	addrVals := simapp.ConvertAddrsToValAddrs(addrDels)

	validators := make([]types.Validator, len(addrVals))

	for i, valAddr := range addrVals {
		validators[i] = types.NewValidator(valAddr, PKs[i], types.Description{})
	}

	hi := types.NewHistoricalInfo(ctx.BlockHeader(), validators)

	app.StakingKeeper.SetHistoricalInfo(ctx, 2, hi)

	recv, found := app.StakingKeeper.GetHistoricalInfo(ctx, 2)
	require.True(t, found, "HistoricalInfo not found after set")
	require.Equal(t, hi, recv, "HistoricalInfo not equal")
	require.True(t, sort.IsSorted(types.Validators(recv.Valset)), "HistoricalInfo validators is not sorted")

	app.StakingKeeper.DeleteHistoricalInfo(ctx, 2)

	recv, found = app.StakingKeeper.GetHistoricalInfo(ctx, 2)
	require.False(t, found, "HistoricalInfo found after delete")
	require.Equal(t, types.HistoricalInfo{}, recv, "HistoricalInfo is not empty")
}

func TestTrackHistoricalInfo(t *testing.T) {
	_, app, ctx := createTestInput()

	addrDels := simapp.AddTestAddrsIncremental(app, ctx, 50, sdk.NewInt(0))
	addrVals := simapp.ConvertAddrsToValAddrs(addrDels)

	// set historical entries in params to 5
	params := types.DefaultParams()
	params.HistoricalEntries = 5
	app.StakingKeeper.SetParams(ctx, params)

	// set historical info at 5, 4 which should be pruned
	// and check that it has been stored
	h4 := abci.Header{
		ChainID: "HelloChain",
		Height:  4,
	}
	h5 := abci.Header{
		ChainID: "HelloChain",
		Height:  5,
	}
	valSet := []types.Validator{
		types.NewValidator(addrVals[0], PKs[0], types.Description{}),
		types.NewValidator(addrVals[1], PKs[1], types.Description{}),
	}
	hi4 := types.NewHistoricalInfo(h4, valSet)
	hi5 := types.NewHistoricalInfo(h5, valSet)
	app.StakingKeeper.SetHistoricalInfo(ctx, 4, hi4)
	app.StakingKeeper.SetHistoricalInfo(ctx, 5, hi5)
	recv, found := app.StakingKeeper.GetHistoricalInfo(ctx, 4)
	require.True(t, found)
	require.Equal(t, hi4, recv)
	recv, found = app.StakingKeeper.GetHistoricalInfo(ctx, 5)
	require.True(t, found)
	require.Equal(t, hi5, recv)

	// Set last validators in keeper
	val1 := types.NewValidator(addrVals[2], PKs[2], types.Description{})
	app.StakingKeeper.SetValidator(ctx, val1)
	app.StakingKeeper.SetLastValidatorPower(ctx, val1.OperatorAddress, 10)
	val2 := types.NewValidator(addrVals[3], PKs[3], types.Description{})
	vals := []types.Validator{val1, val2}
	sort.Sort(types.Validators(vals))
	app.StakingKeeper.SetValidator(ctx, val2)
	app.StakingKeeper.SetLastValidatorPower(ctx, val2.OperatorAddress, 8)

	// Set Header for BeginBlock context
	header := abci.Header{
		ChainID: "HelloChain",
		Height:  10,
	}
	ctx = ctx.WithBlockHeader(header)

	app.StakingKeeper.TrackHistoricalInfo(ctx)

	// Check HistoricalInfo at height 10 is persisted
	expected := types.HistoricalInfo{
		Header: header,
		Valset: vals,
	}
	recv, found = app.StakingKeeper.GetHistoricalInfo(ctx, 10)
	require.True(t, found, "GetHistoricalInfo failed after BeginBlock")
	require.Equal(t, expected, recv, "GetHistoricalInfo returned eunexpected result")

	// Check HistoricalInfo at height 5, 4 is pruned
	recv, found = app.StakingKeeper.GetHistoricalInfo(ctx, 4)
	require.False(t, found, "GetHistoricalInfo did not prune earlier height")
	require.Equal(t, types.HistoricalInfo{}, recv, "GetHistoricalInfo at height 4 is not empty after prune")
	recv, found = app.StakingKeeper.GetHistoricalInfo(ctx, 5)
	require.False(t, found, "GetHistoricalInfo did not prune first prune height")
	require.Equal(t, types.HistoricalInfo{}, recv, "GetHistoricalInfo at height 5 is not empty after prune")
}
