package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	codecstd "github.com/cosmos/cosmos-sdk/codec/std"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keep "github.com/cosmos/cosmos-sdk/x/supply/keeper"
	"github.com/cosmos/cosmos-sdk/x/supply/types"
)

const (
	initialPower = int64(100)
	holder       = "holder"
	multiPerm    = "multiple permissions account"
	randomPerm   = "random permission"
)

// create module accounts for testing
var (
	holderAcc     = types.NewEmptyModuleAccount(holder)
	burnerAcc     = types.NewEmptyModuleAccount(types.Burner, types.Burner)
	minterAcc     = types.NewEmptyModuleAccount(types.Minter, types.Minter)
	multiPermAcc  = types.NewEmptyModuleAccount(multiPerm, types.Burner, types.Minter, types.Staking)
	randomPermAcc = types.NewEmptyModuleAccount(randomPerm, "random")

	initTokens = sdk.TokensFromConsensusPower(initialPower)
	initCoins  = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens))
)

// nolint
func getCoinsByName(ctx sdk.Context, sk keep.Keeper, ak types.AccountKeeper, bk types.BankKeeper, moduleName string) sdk.Coins {
	moduleAddress := sk.GetModuleAddress(moduleName)
	macc := ak.GetAccount(ctx, moduleAddress)
	if macc == nil {
		return sdk.Coins(nil)
	}

	return bk.GetAllBalances(ctx, macc.GetAddress())
}

func TestSendCoins(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{Height: 1})

	// add module accounts to supply keeper
	maccPerms := simapp.GetMaccPerms()
	maccPerms[holder] = nil
	maccPerms[types.Burner] = []string{types.Burner}
	maccPerms[types.Minter] = []string{types.Minter}
	maccPerms[multiPerm] = []string{types.Burner, types.Minter, types.Staking}
	maccPerms[randomPerm] = []string{"random"}

	appCodec := codecstd.NewAppCodec(app.Codec())

	keeper := keep.NewKeeper(appCodec, app.GetKey(types.StoreKey), app.AccountKeeper, app.BankKeeper, maccPerms)
	ak := app.AccountKeeper
	bk := app.BankKeeper

	baseAcc := ak.NewAccountWithAddress(ctx, types.NewModuleAddress("baseAcc"))

	require.NoError(t, bk.SetBalances(ctx, holderAcc.GetAddress(), initCoins))
	keeper.SetSupply(ctx, types.NewSupply(initCoins))

	keeper.SetModuleAccount(ctx, holderAcc)
	keeper.SetModuleAccount(ctx, burnerAcc)
	ak.SetAccount(ctx, baseAcc)

	require.Panics(t, func() {
		keeper.SendCoinsFromModuleToModule(ctx, "", holderAcc.GetName(), initCoins) // nolint:errcheck
	})

	require.Panics(t, func() {
		keeper.SendCoinsFromModuleToModule(ctx, types.Burner, "", initCoins) // nolint:errcheck
	})

	require.Panics(t, func() {
		keeper.SendCoinsFromModuleToAccount(ctx, "", baseAcc.GetAddress(), initCoins) // nolint:errcheck
	})

	require.Error(
		t,
		keeper.SendCoinsFromModuleToAccount(ctx, holderAcc.GetName(), baseAcc.GetAddress(), initCoins.Add(initCoins...)),
	)

	require.NoError(
		t, keeper.SendCoinsFromModuleToModule(ctx, holderAcc.GetName(), types.Burner, initCoins),
	)
	require.Equal(t, sdk.Coins(nil), getCoinsByName(ctx, keeper, ak, bk, holderAcc.GetName()))
	require.Equal(t, initCoins, getCoinsByName(ctx, keeper, ak, bk, types.Burner))

	require.NoError(
		t, keeper.SendCoinsFromModuleToAccount(ctx, types.Burner, baseAcc.GetAddress(), initCoins),
	)
	require.Equal(t, sdk.Coins(nil), getCoinsByName(ctx, keeper, ak, bk, types.Burner))
	require.Equal(t, initCoins, bk.GetAllBalances(ctx, baseAcc.GetAddress()))

	require.NoError(t, keeper.SendCoinsFromAccountToModule(ctx, baseAcc.GetAddress(), types.Burner, initCoins))
	require.Equal(t, sdk.Coins(nil), bk.GetAllBalances(ctx, baseAcc.GetAddress()))
	require.Equal(t, initCoins, getCoinsByName(ctx, keeper, ak, bk, types.Burner))
}

func TestMintCoins(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{Height: 1})

	// add module accounts to supply keeper
	maccPerms := simapp.GetMaccPerms()
	maccPerms[holder] = nil
	maccPerms[types.Burner] = []string{types.Burner}
	maccPerms[types.Minter] = []string{types.Minter}
	maccPerms[multiPerm] = []string{types.Burner, types.Minter, types.Staking}
	maccPerms[randomPerm] = []string{"random"}

	appCodec := codecstd.NewAppCodec(app.Codec())

	keeper := keep.NewKeeper(appCodec, app.GetKey(types.StoreKey), app.AccountKeeper, app.BankKeeper, maccPerms)
	ak := app.AccountKeeper
	bk := app.BankKeeper

	keeper.SetModuleAccount(ctx, burnerAcc)
	keeper.SetModuleAccount(ctx, minterAcc)
	keeper.SetModuleAccount(ctx, multiPermAcc)
	keeper.SetModuleAccount(ctx, randomPermAcc)

	initialSupply := keeper.GetSupply(ctx)

	require.Panics(t, func() { keeper.MintCoins(ctx, "", initCoins) }, "no module account")            // nolint:errcheck
	require.Panics(t, func() { keeper.MintCoins(ctx, types.Burner, initCoins) }, "invalid permission") // nolint:errcheck
	err := keeper.MintCoins(ctx, types.Minter, sdk.Coins{sdk.Coin{Denom: "denom", Amount: sdk.NewInt(-10)}})
	require.Error(t, err, "insufficient coins")

	require.Panics(t, func() { keeper.MintCoins(ctx, randomPerm, initCoins) }) // nolint:errcheck

	err = keeper.MintCoins(ctx, types.Minter, initCoins)
	require.NoError(t, err)
	require.Equal(t, initCoins, getCoinsByName(ctx, keeper, ak, bk, types.Minter))
	require.Equal(t, initialSupply.GetTotal().Add(initCoins...), keeper.GetSupply(ctx).GetTotal())

	// test same functionality on module account with multiple permissions
	initialSupply = keeper.GetSupply(ctx)

	err = keeper.MintCoins(ctx, multiPermAcc.GetName(), initCoins)
	require.NoError(t, err)
	require.Equal(t, initCoins, getCoinsByName(ctx, keeper, ak, bk, multiPermAcc.GetName()))
	require.Equal(t, initialSupply.GetTotal().Add(initCoins...), keeper.GetSupply(ctx).GetTotal())

	require.Panics(t, func() { keeper.MintCoins(ctx, types.Burner, initCoins) }) // nolint:errcheck
}

func TestBurnCoins(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{Height: 1})

	// add module accounts to supply keeper
	maccPerms := simapp.GetMaccPerms()
	maccPerms[holder] = nil
	maccPerms[types.Burner] = []string{types.Burner}
	maccPerms[types.Minter] = []string{types.Minter}
	maccPerms[multiPerm] = []string{types.Burner, types.Minter, types.Staking}
	maccPerms[randomPerm] = []string{"random"}

	appCodec := codecstd.NewAppCodec(app.Codec())

	keeper := keep.NewKeeper(appCodec, app.GetKey(types.StoreKey), app.AccountKeeper, app.BankKeeper, maccPerms)
	ak := app.AccountKeeper
	bk := app.BankKeeper

	require.NoError(t, bk.SetBalances(ctx, burnerAcc.GetAddress(), initCoins))
	keeper.SetSupply(ctx, types.NewSupply(initCoins))
	keeper.SetModuleAccount(ctx, burnerAcc)

	initialSupply := keeper.GetSupply(ctx)
	initialSupply.Inflate(initCoins)
	keeper.SetSupply(ctx, initialSupply)

	require.Panics(t, func() { keeper.BurnCoins(ctx, "", initCoins) }, "no module account")                        // nolint:errcheck
	require.Panics(t, func() { keeper.BurnCoins(ctx, types.Minter, initCoins) }, "invalid permission")             // nolint:errcheck
	require.Panics(t, func() { keeper.BurnCoins(ctx, randomPerm, initialSupply.GetTotal()) }, "random permission") // nolint:errcheck
	err := keeper.BurnCoins(ctx, types.Burner, initialSupply.GetTotal())
	require.Error(t, err, "insufficient coins")

	err = keeper.BurnCoins(ctx, types.Burner, initCoins)
	require.NoError(t, err)
	require.Equal(t, sdk.Coins(nil), getCoinsByName(ctx, keeper, ak, bk, types.Burner))
	require.Equal(t, initialSupply.GetTotal().Sub(initCoins), keeper.GetSupply(ctx).GetTotal())

	// test same functionality on module account with multiple permissions
	initialSupply = keeper.GetSupply(ctx)
	initialSupply.Inflate(initCoins)
	keeper.SetSupply(ctx, initialSupply)

	require.NoError(t, bk.SetBalances(ctx, multiPermAcc.GetAddress(), initCoins))
	keeper.SetModuleAccount(ctx, multiPermAcc)

	err = keeper.BurnCoins(ctx, multiPermAcc.GetName(), initCoins)
	require.NoError(t, err)
	require.Equal(t, sdk.Coins(nil), getCoinsByName(ctx, keeper, ak, bk, multiPermAcc.GetName()))
	require.Equal(t, initialSupply.GetTotal().Sub(initCoins), keeper.GetSupply(ctx).GetTotal())
}
