package keeper

import (
	"sort"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/stretchr/testify/require"
)

func TestHistoricalInfo(t *testing.T) {
	ctx, _, keeper, _ := CreateTestInput(t, false, 10)
	var validators []types.Validator

	for i, valAddr := range addrVals {
		validators = append(validators, types.NewValidator(valAddr, PKs[i], types.Description{}))
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
