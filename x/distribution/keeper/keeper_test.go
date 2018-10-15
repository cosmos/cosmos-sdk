package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/stretchr/testify/require"
)

func TestSetGetProposerConsAddr(t *testing.T) {
	ctx, _, keeper, _, _ := CreateTestInputDefault(t, false, 0)

	keeper.SetProposerConsAddr(ctx, valConsAddr1)
	res := keeper.GetProposerConsAddr(ctx)
	require.True(t, res.Equals(valConsAddr1), "expected: %v got: %v", valConsAddr1.String(), res.String())
}

func TestSetGetPercentPrecommitVotes(t *testing.T) {
	ctx, _, keeper, _, _ := CreateTestInputDefault(t, false, 0)

	someDec := sdk.NewDec(333)
	keeper.SetPercentPrecommitVotes(ctx, someDec)
	res := keeper.GetPercentPrecommitVotes(ctx)
	require.True(sdk.DecEq(t, someDec, res))
}

func TestSetGetCommunityTax(t *testing.T) {
	ctx, _, keeper, _, _ := CreateTestInputDefault(t, false, 0)

	someDec := sdk.NewDec(333)
	keeper.SetCommunityTax(ctx, someDec)
	res := keeper.GetCommunityTax(ctx)
	require.True(sdk.DecEq(t, someDec, res))
}

func TestSetGetFeePool(t *testing.T) {
	ctx, _, keeper, _, _ := CreateTestInputDefault(t, false, 0)

	fp := types.InitialFeePool()
	fp.ValAccum.UpdateHeight = 777

	keeper.SetFeePool(ctx, fp)
	res := keeper.GetFeePool(ctx)
	require.Equal(t, fp.ValAccum, res.ValAccum)
}
