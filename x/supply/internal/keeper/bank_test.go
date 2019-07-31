package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply/internal/types"
)

const initialPower = int64(100)

var (
	holderAcc     = types.NewEmptyModuleAccount(holder)
	burnerAcc     = types.NewEmptyModuleAccount(types.Burner, types.Burner)
	minterAcc     = types.NewEmptyModuleAccount(types.Minter, types.Minter)
	multiPermAcc  = types.NewEmptyModuleAccount(multiPerm, types.Burner, types.Minter, types.Staking)
	randomPermAcc = types.NewEmptyModuleAccount(randomPerm, "random")

	initTokens = sdk.TokensFromConsensusPower(initialPower)
	initCoins  = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens))
)

func getCoinsByName(ctx sdk.Context, k Keeper, moduleName string) sdk.Coins {
	moduleAddress := k.GetModuleAddress(moduleName)
	macc := k.ak.GetAccount(ctx, moduleAddress)
	if macc == nil {
		return sdk.Coins(nil)
	}
	return macc.GetCoins()
}

func TestSendCoins(t *testing.T) {
	nAccs := int64(4)
	ctx, ak, keeper := createTestInput(t, false, initialPower, nAccs)

	baseAcc := ak.NewAccountWithAddress(ctx, types.NewModuleAddress("baseAcc"))

	err := holderAcc.SetCoins(initCoins)
	require.NoError(t, err)

	keeper.SetModuleAccount(ctx, holderAcc)
	keeper.SetModuleAccount(ctx, burnerAcc)
	ak.SetAccount(ctx, baseAcc)

	err = keeper.SendCoinsFromModuleToModule(ctx, "", holderAcc.GetName(), initCoins)
	require.Error(t, err)

	require.Panics(t, func() {
		keeper.SendCoinsFromModuleToModule(ctx, types.Burner, "", initCoins)
	})

	err = keeper.SendCoinsFromModuleToAccount(ctx, "", baseAcc.GetAddress(), initCoins)
	require.Error(t, err)

	err = keeper.SendCoinsFromModuleToAccount(ctx, holderAcc.GetName(), baseAcc.GetAddress(), initCoins.Add(initCoins))
	require.Error(t, err)

	err = keeper.SendCoinsFromModuleToModule(ctx, holderAcc.GetName(), types.Burner, initCoins)
	require.NoError(t, err)
	require.Equal(t, sdk.Coins(nil), getCoinsByName(ctx, keeper, holderAcc.GetName()))
	require.Equal(t, initCoins, getCoinsByName(ctx, keeper, types.Burner))

	err = keeper.SendCoinsFromModuleToAccount(ctx, types.Burner, baseAcc.GetAddress(), initCoins)
	require.NoError(t, err)
	require.Equal(t, sdk.Coins(nil), getCoinsByName(ctx, keeper, types.Burner))

	require.Equal(t, initCoins, keeper.ak.GetAccount(ctx, baseAcc.GetAddress()).GetCoins())

	err = keeper.SendCoinsFromAccountToModule(ctx, baseAcc.GetAddress(), types.Burner, initCoins)
	require.NoError(t, err)
	require.Equal(t, sdk.Coins(nil), keeper.ak.GetAccount(ctx, baseAcc.GetAddress()).GetCoins())
	require.Equal(t, initCoins, getCoinsByName(ctx, keeper, types.Burner))
}

func TestMintCoins(t *testing.T) {
	nAccs := int64(4)
	ctx, _, keeper := createTestInput(t, false, initialPower, nAccs)

	keeper.SetModuleAccount(ctx, burnerAcc)
	keeper.SetModuleAccount(ctx, minterAcc)
	keeper.SetModuleAccount(ctx, multiPermAcc)
	keeper.SetModuleAccount(ctx, randomPermAcc)

	initialSupply := keeper.GetSupply(ctx)

	require.Error(t, keeper.MintCoins(ctx, "", initCoins), "no module account")
	require.Panics(t, func() { keeper.MintCoins(ctx, types.Burner, initCoins) }, "invalid permission")
	require.Panics(t, func() { keeper.MintCoins(ctx, types.Minter, sdk.Coins{sdk.Coin{"denom", sdk.NewInt(-10)}}) }, "insufficient coins") //nolint

	require.Panics(t, func() { keeper.MintCoins(ctx, randomPerm, initCoins) })

	err := keeper.MintCoins(ctx, types.Minter, initCoins)
	require.NoError(t, err)
	require.Equal(t, initCoins, getCoinsByName(ctx, keeper, types.Minter))
	require.Equal(t, initialSupply.GetTotal().Add(initCoins), keeper.GetSupply(ctx).GetTotal())

	// test same functionality on module account with multiple permissions
	initialSupply = keeper.GetSupply(ctx)

	err = keeper.MintCoins(ctx, multiPermAcc.GetName(), initCoins)
	require.NoError(t, err)
	require.Equal(t, initCoins, getCoinsByName(ctx, keeper, multiPermAcc.GetName()))
	require.Equal(t, initialSupply.GetTotal().Add(initCoins), keeper.GetSupply(ctx).GetTotal())

	require.Panics(t, func() { keeper.MintCoins(ctx, types.Burner, initCoins) })
}

func TestBurnCoins(t *testing.T) {
	nAccs := int64(4)
	ctx, _, keeper := createTestInput(t, false, initialPower, nAccs)

	require.NoError(t, burnerAcc.SetCoins(initCoins))
	keeper.SetModuleAccount(ctx, burnerAcc)

	initialSupply := keeper.GetSupply(ctx)
	initialSupply = initialSupply.Inflate(initCoins)
	keeper.SetSupply(ctx, initialSupply)

	require.Error(t, keeper.BurnCoins(ctx, "", initCoins), "no module account")
	require.Panics(t, func() { keeper.BurnCoins(ctx, types.Minter, initCoins) }, "invalid permission")
	require.Panics(t, func() { keeper.BurnCoins(ctx, randomPerm, initialSupply.GetTotal()) }, "random permission")
	require.Panics(t, func() { keeper.BurnCoins(ctx, types.Burner, initialSupply.GetTotal()) }, "insufficient coins")

	err := keeper.BurnCoins(ctx, types.Burner, initCoins)
	require.NoError(t, err)
	require.Equal(t, sdk.Coins(nil), getCoinsByName(ctx, keeper, types.Burner))
	require.Equal(t, initialSupply.GetTotal().Sub(initCoins), keeper.GetSupply(ctx).GetTotal())

	// test same functionality on module account with multiple permissions
	initialSupply = keeper.GetSupply(ctx)
	initialSupply = initialSupply.Inflate(initCoins)
	keeper.SetSupply(ctx, initialSupply)

	require.NoError(t, multiPermAcc.SetCoins(initCoins))
	keeper.SetModuleAccount(ctx, multiPermAcc)

	err = keeper.BurnCoins(ctx, multiPermAcc.GetName(), initCoins)
	require.NoError(t, err)
	require.Equal(t, sdk.Coins(nil), getCoinsByName(ctx, keeper, multiPermAcc.GetName()))
	require.Equal(t, initialSupply.GetTotal().Sub(initCoins), keeper.GetSupply(ctx).GetTotal())
}
