package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply/types"
)

const (
	foo = "foo"
	bar = "bar"
	minter = "minter"
	base = "baseAcc"

	initialPower = int64(100)
)

var (
	fooAcc = types.NewModuleHolderAccount(foo)
	barAcc = types.NewModuleHolderAccount(bar)
	minterAcc = types.NewModuleMinterAccount(minter)

	initTokens = sdk.TokensFromTendermintPower(initialPower)
	initCoins = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens))
)

func TestSendCoins(t *testing.T) {
	nAccs := int64(4)
	ctx, ak, keeper := createTestInput(t, false, initialPower, nAccs)

	baseAcc := ak.NewAccountWithAddress(ctx, sdk.AccAddress(crypto.AddressHash([]byte("baseAcc"))))

	err := fooAcc.SetCoins(initCoins)
	require.NoError(t, err)

	keeper.SetModuleAccount(ctx, fooAcc)
	keeper.SetModuleAccount(ctx, barAcc)
	ak.SetAccount(ctx, baseAcc)

	err = keeper.SendCoinsModuleToModule(ctx, "", bar, initCoins)
	require.Error(t, err)

	err = keeper.SendCoinsModuleToModule(ctx, foo, "", initCoins)
	require.Error(t, err)

	err = keeper.SendCoinsModuleToAccount(ctx, "", baseAcc.GetAddress(), initCoins)
	require.Error(t, err)

	err = keeper.SendCoinsModuleToAccount(ctx, foo, baseAcc.GetAddress(), initCoins.Add(initCoins))
	require.Error(t, err)

	err = keeper.SendCoinsModuleToModule(ctx, foo, bar, initCoins)
	require.NoError(t, err)
	require.Equal(t, sdk.Coins(nil), keeper.GetCoinsByName(ctx, foo))
	require.Equal(t, initCoins, keeper.GetCoinsByName(ctx, bar))

	err = keeper.SendCoinsModuleToAccount(ctx, bar, baseAcc.GetAddress(), initCoins)
	require.NoError(t, err)
	require.Equal(t, sdk.Coins(nil), keeper.GetCoinsByName(ctx, bar))
	require.Equal(t, initCoins, keeper.GetCoins(ctx, baseAcc.GetAddress()))

	err = keeper.SendCoinsAccountToModule(ctx, baseAcc.GetAddress(), foo, initCoins)
	require.NoError(t, err)
	require.Equal(t, sdk.Coins(nil), keeper.GetCoins(ctx, baseAcc.GetAddress()))
	require.Equal(t, initCoins, keeper.GetCoinsByName(ctx, foo))
}

func TestMintCoins(t *testing.T) {
	nAccs := int64(4)
	ctx, _, keeper := createTestInput(t, false, initialPower, nAccs)

	keeper.SetModuleAccount(ctx, fooAcc)
	keeper.SetModuleAccount(ctx, minterAcc)

	initialSupply := keeper.GetSupply(ctx)

	err := keeper.MintCoins(ctx, "", initCoins)
	require.Error(t, err)

	err = keeper.MintCoins(ctx, minter, initCoins)
	require.NoError(t, err)
	require.Equal(t, initCoins, keeper.GetCoinsByName(ctx, minter))
	require.Equal(t, initialSupply.Total.Add(initCoins), keeper.GetSupply(ctx).Total)

	require.Panics(t,  func() { keeper.MintCoins(ctx, foo, initCoins) })
}

func TestBurnCoins(t *testing.T) {
	nAccs := int64(4)
	ctx, _, keeper := createTestInput(t, false, initialPower, nAccs)

	err := fooAcc.SetCoins(initCoins)
	require.NoError(t, err)
	keeper.SetModuleAccount(ctx, fooAcc)

	initialSupply := keeper.GetSupply(ctx)
	initialSupply.Inflate(initCoins)
	keeper.SetSupply(ctx, initialSupply)

	err = keeper.BurnCoins(ctx, "", initCoins)
	require.Error(t, err)

	err = keeper.BurnCoins(ctx, foo, initialSupply.Total)
	require.Error(t, err)

	err = keeper.BurnCoins(ctx, foo, initCoins)
	require.NoError(t, err)
	require.Equal(t, sdk.Coins(nil), keeper.GetCoinsByName(ctx, foo))
	require.Equal(t, initialSupply.Total.Sub(initCoins), keeper.GetSupply(ctx).Total)
}
