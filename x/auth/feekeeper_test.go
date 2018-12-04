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

	// no coins initially
	currFees := input.fck.GetCollectedFees(input.ctx)
	require.True(t, currFees.IsEqual(emptyCoins))

	// set feeCollection to oneCoin
	input.fck.setCollectedFees(input.ctx, oneCoin)

	// check that it is equal to oneCoin
	require.True(t, input.fck.GetCollectedFees(input.ctx).IsEqual(oneCoin))
}

func TestFeeCollectionKeeperAdd(t *testing.T) {
	input := setupTestInput()

	// no coins initially
	require.True(t, input.fck.GetCollectedFees(input.ctx).IsEqual(emptyCoins))

	// add oneCoin and check that pool is now oneCoin
	input.fck.AddCollectedFees(input.ctx, oneCoin)
	require.True(t, input.fck.GetCollectedFees(input.ctx).IsEqual(oneCoin))

	// add oneCoin again and check that pool is now twoCoins
	input.fck.AddCollectedFees(input.ctx, oneCoin)
	require.True(t, input.fck.GetCollectedFees(input.ctx).IsEqual(twoCoins))
}

func TestFeeCollectionKeeperClear(t *testing.T) {
	input := setupTestInput()

	// set coins initially
	input.fck.setCollectedFees(input.ctx, twoCoins)
	require.True(t, input.fck.GetCollectedFees(input.ctx).IsEqual(twoCoins))

	// clear fees and see that pool is now empty
	input.fck.ClearCollectedFees(input.ctx)
	require.True(t, input.fck.GetCollectedFees(input.ctx).IsEqual(emptyCoins))
}
