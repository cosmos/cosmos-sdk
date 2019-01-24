package auth

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	emptyCoins = sdk.Coins{}
	oneCoin    = sdk.Coins{sdk.NewInt64Coin("foocoin", 1)}
	twoCoins   = sdk.Coins{sdk.NewInt64Coin("foocoin", 2)}
)

func TestFeeCollectionKeeperGetSet(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx

	// no coins initially
	currFees := input.fck.GetCollectedFees(ctx)
	require.True(t, currFees.IsEqual(emptyCoins))

	// set feeCollection to oneCoin
	input.fck.setCollectedFees(ctx, oneCoin)

	// check that it is equal to oneCoin
	require.True(t, input.fck.GetCollectedFees(ctx).IsEqual(oneCoin))
}

func TestFeeCollectionKeeperAdd(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx

	// no coins initially
	require.True(t, input.fck.GetCollectedFees(ctx).IsEqual(emptyCoins))

	// add oneCoin and check that pool is now oneCoin
	input.fck.AddCollectedFees(ctx, oneCoin)
	require.True(t, input.fck.GetCollectedFees(ctx).IsEqual(oneCoin))

	// add oneCoin again and check that pool is now twoCoins
	input.fck.AddCollectedFees(ctx, oneCoin)
	require.True(t, input.fck.GetCollectedFees(ctx).IsEqual(twoCoins))
}

func TestFeeCollectionKeeperClear(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx

	// set coins initially
	input.fck.setCollectedFees(ctx, twoCoins)
	require.True(t, input.fck.GetCollectedFees(ctx).IsEqual(twoCoins))

	// clear fees and see that pool is now empty
	input.fck.ClearCollectedFees(ctx)
	require.True(t, input.fck.GetCollectedFees(ctx).IsEqual(emptyCoins))
}
