package keeper

import (
	"sort"
	"testing"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/stretchr/testify/require"
)

func TestHistoricalInfo(t *testing.T) {
	ctx, _, keeper, _ := CreateTestInput(t, false, 10)
	validators := make([]types.Validator, len(addrVals))

	for i, valAddr := range addrVals {
		validators[i] = types.NewValidator(valAddr, PKs[i], types.Description{})
	}

	hi := types.NewHistoricalInfo(ctx.BlockHeader(), validators)

	keeper.SetHistoricalInfo(ctx, 2, hi)

	recv, found := keeper.GetHistoricalInfo(ctx, 2)
	require.True(t, found, "HistoricalInfo not found after set")
	require.Equal(t, hi, recv, "HistoricalInfo not equal")
	require.True(t, sort.IsSorted(types.Validators(recv.ValSet)), "HistoricalInfo validators is not sorted")

	keeper.DeleteHistoricalInfo(ctx, 2)

	recv, found = keeper.GetHistoricalInfo(ctx, 2)
	require.False(t, found, "HistoricalInfo found after delete")
	require.Equal(t, types.HistoricalInfo{}, recv, "HistoricalInfo is not empty")
}

func TestTrackHistoricalInfo(t *testing.T) {
	ctx, _, k, _ := CreateTestInput(t, false, 10)

	// set historical entries in params to 5
	params := types.DefaultParams()
	params.HistoricalEntries = 5
	k.SetParams(ctx, params)

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
		types.NewValidator(sdk.ValAddress(Addrs[0]), PKs[0], types.Description{}),
		types.NewValidator(sdk.ValAddress(Addrs[1]), PKs[1], types.Description{}),
	}
	hi4 := types.NewHistoricalInfo(h4, valSet)
	hi5 := types.NewHistoricalInfo(h5, valSet)
	k.SetHistoricalInfo(ctx, 4, hi4)
	k.SetHistoricalInfo(ctx, 5, hi5)
	recv, found := k.GetHistoricalInfo(ctx, 4)
	require.True(t, found)
	require.Equal(t, hi4, recv)
	recv, found = k.GetHistoricalInfo(ctx, 5)
	require.True(t, found)
	require.Equal(t, hi5, recv)

	// Set last validators in keeper
	val1 := types.NewValidator(sdk.ValAddress(Addrs[2]), PKs[2], types.Description{})
	k.SetValidator(ctx, val1)
	k.SetLastValidatorPower(ctx, val1.OperatorAddress, 10)
	val2 := types.NewValidator(sdk.ValAddress(Addrs[3]), PKs[3], types.Description{})
	vals := []types.Validator{val1, val2}
	sort.Sort(types.Validators(vals))
	k.SetValidator(ctx, val2)
	k.SetLastValidatorPower(ctx, val2.OperatorAddress, 8)

	// Set Header for BeginBlock context
	header := abci.Header{
		ChainID: "HelloChain",
		Height:  10,
	}
	ctx = ctx.WithBlockHeader(header)

	k.TrackHistoricalInfo(ctx)

	// Check HistoricalInfo at height 10 is persisted
	expected := types.HistoricalInfo{
		Header: header,
		ValSet: vals,
	}
	recv, found = k.GetHistoricalInfo(ctx, 10)
	require.True(t, found, "GetHistoricalInfo failed after BeginBlock")
	require.Equal(t, expected, recv, "GetHistoricalInfo returned eunexpected result")

	// Check HistoricalInfo at height 5, 4 is pruned
	recv, found = k.GetHistoricalInfo(ctx, 4)
	require.False(t, found, "GetHistoricalInfo did not prune earlier height")
	require.Equal(t, types.HistoricalInfo{}, recv, "GetHistoricalInfo at height 4 is not empty after prune")
	recv, found = k.GetHistoricalInfo(ctx, 5)
	require.False(t, found, "GetHistoricalInfo did not prune first prune height")
	require.Equal(t, types.HistoricalInfo{}, recv, "GetHistoricalInfo at height 5 is not empty after prune")
}
