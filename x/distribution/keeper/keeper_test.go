package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/stretchr/testify/require"
)

func TestSetGetPreviousProposerConsAddr(t *testing.T) {
	ctx, _, keeper, _, _ := CreateTestInputDefault(t, false, 0)

	keeper.SetPreviousProposerConsAddr(ctx, valConsAddr1)
	res := keeper.GetPreviousProposerConsAddr(ctx)
	require.True(t, res.Equals(valConsAddr1), "expected: %v got: %v", valConsAddr1.String(), res.String())
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
	fp.TotalValAccum.UpdateHeight = 777

	keeper.SetFeePool(ctx, fp)
	res := keeper.GetFeePool(ctx)
	require.Equal(t, fp.TotalValAccum, res.TotalValAccum)
}
