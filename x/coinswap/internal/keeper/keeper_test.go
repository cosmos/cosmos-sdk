package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/coinswap/internal/types"
)

const (
	moduleName = "swap:atom:btc"
)

// test that the module account gets created with an initial
// balance of zero coins.
func TestCreateReservePool(t *testing.T) {
	ctx, keeper, _ := createTestInput(t, sdk.NewInt(0), 0)

	moduleAcc := keeper.sk.GetModuleAccount(ctx, moduleName)
	require.Nil(t, moduleAcc)

	keeper.CreateReservePool(ctx, moduleName)
	moduleAcc = keeper.sk.GetModuleAccount(ctx, moduleName)
	require.NotNil(t, moduleAcc)
	require.Equal(t, sdk.Coins{}, moduleAcc.GetCoins(), "module account has non zero balance after creation")

	// attempt to recreate existing ModuleAccount
	require.Panics(t, func() { keeper.CreateReservePool(ctx, moduleName) })
}

// test that the params can be properly set and retrieved
func TestParams(t *testing.T) {
	ctx, keeper, _ := createTestInput(t, sdk.NewInt(0), 0)

	cases := []struct {
		params types.Params
	}{
		{types.DefaultParams()},
		{types.NewParams("pineapple", types.NewFeeParam(sdk.NewInt(5), sdk.NewInt(10)))},
	}

	for _, tc := range cases {
		keeper.SetParams(ctx, tc.params)

		feeParam := keeper.GetFeeParam(ctx)
		require.Equal(t, tc.params.Fee, feeParam)

		nativeDenom := keeper.GetNativeDenom(ctx)
		require.Equal(t, tc.params.NativeDenom, nativeDenom)
	}
}

// test that non existent reserve pool returns false and
// that balance is updated.
func TestGetReservePool(t *testing.T) {
	amt := sdk.NewInt(100)
	ctx, keeper, accs := createTestInput(t, amt, 1)

	reservePool, found := keeper.GetReservePool(ctx, moduleName)
	require.False(t, found)

	keeper.CreateReservePool(ctx, moduleName)
	reservePool, found = keeper.GetReservePool(ctx, moduleName)
	require.True(t, found)

	keeper.sk.SendCoinsFromAccountToModule(ctx, accs[0].GetAddress(), moduleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt)))
	reservePool, found = keeper.GetReservePool(ctx, moduleName)
	reservePool, found = keeper.GetReservePool(ctx, moduleName)
	require.True(t, found)
	require.Equal(t, amt, reservePool.AmountOf(sdk.DefaultBondDenom))
}
