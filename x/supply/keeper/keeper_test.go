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

func TestSupply(t *testing.T) {
	initialPower := int64(100)
	initTokens := sdk.TokensFromConsensusPower(initialPower)

	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{Height: 1})

	totalSupply := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens))
	app.SupplyKeeper.SetSupply(ctx, types.NewSupply(totalSupply))

	total := app.SupplyKeeper.GetSupply(ctx).GetTotal()

	require.Equal(t, totalSupply, total)
}

func TestValidatePermissions(t *testing.T) {
	app := simapp.Setup(false)

	// add module accounts to supply keeper
	maccPerms := simapp.GetMaccPerms()
	maccPerms[holder] = nil
	maccPerms[types.Burner] = []string{types.Burner}
	maccPerms[types.Minter] = []string{types.Minter}
	maccPerms[multiPerm] = []string{types.Burner, types.Minter, types.Staking}
	maccPerms[randomPerm] = []string{"random"}

	appCodec := codecstd.NewAppCodec(app.Codec())
	keeper := keep.NewKeeper(appCodec, app.GetKey(types.StoreKey), app.AccountKeeper, app.BankKeeper, maccPerms)

	err := keeper.ValidatePermissions(multiPermAcc)
	require.NoError(t, err)

	err = keeper.ValidatePermissions(randomPermAcc)
	require.NoError(t, err)

	// unregistered permissions
	otherAcc := types.NewEmptyModuleAccount("other", "other")
	err = app.SupplyKeeper.ValidatePermissions(otherAcc)
	require.Error(t, err)
}
