package slashing_test

import (
	"errors"
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/require"
)

func TestCannotUnjailUnlessJailed(t *testing.T) {
	// initial setup
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	pks := simapp.CreateTestPubKeys(1)
	simapp.AddTestAddrsFromPubKeys(app, ctx, pks, sdk.TokensFromConsensusPower(200))

	slh := slashing.NewHandler(app.SlashingKeeper)
	amt := sdk.TokensFromConsensusPower(100)
	addr, val := sdk.ValAddress(pks[0].Address()), pks[0]

	msg := slashingkeeper.NewTestMsgCreateValidator(addr, val, amt)
	res, err := staking.NewHandler(app.StakingKeeper)(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	staking.EndBlocker(ctx, app.StakingKeeper)

	require.Equal(
		t, app.BankKeeper.GetAllBalances(ctx, sdk.AccAddress(addr)),
		sdk.Coins{sdk.NewCoin(app.StakingKeeper.GetParams(ctx).BondDenom, slashingkeeper.InitTokens.Sub(amt))},
	)
	require.Equal(t, amt, app.StakingKeeper.Validator(ctx, addr).GetBondedTokens())

	// assert non-jailed validator can't be unjailed
	res, err = slh(ctx, slashing.NewMsgUnjail(addr))
	require.Error(t, err)
	require.Nil(t, res)
	require.True(t, errors.Is(slashing.ErrValidatorNotJailed, err))
}

func TestCannotUnjailUnlessMeetMinSelfDelegation(t *testing.T) {
	// initial setup
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})
	pks := simapp.CreateTestPubKeys(1)
	simapp.AddTestAddrsFromPubKeys(app, ctx, pks, sdk.TokensFromConsensusPower(200))

	slh := slashing.NewHandler(app.SlashingKeeper)
	amtInt := int64(100)
	addr, val, amt := sdk.ValAddress(pks[0].Address()), pks[0], sdk.TokensFromConsensusPower(amtInt)
	msg := slashingkeeper.NewTestMsgCreateValidator(addr, val, amt)
	msg.MinSelfDelegation = amt

	res, err := staking.NewHandler(app.StakingKeeper)(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	staking.EndBlocker(ctx, app.StakingKeeper)

	require.Equal(
		t, app.BankKeeper.GetAllBalances(ctx, sdk.AccAddress(addr)),
		sdk.Coins{sdk.NewCoin(app.StakingKeeper.GetParams(ctx).BondDenom, slashingkeeper.InitTokens.Sub(amt))},
	)

	unbondAmt := sdk.NewCoin(app.StakingKeeper.GetParams(ctx).BondDenom, sdk.OneInt())
	undelegateMsg := staking.NewMsgUndelegate(sdk.AccAddress(addr), addr, unbondAmt)
	res, err = staking.NewHandler(app.StakingKeeper)(ctx, undelegateMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	require.True(t, app.StakingKeeper.Validator(ctx, addr).IsJailed())

	// assert non-jailed validator can't be unjailed
	res, err = slh(ctx, slashing.NewMsgUnjail(addr))
	require.Error(t, err)
	require.Nil(t, res)
	require.True(t, errors.Is(slashing.ErrSelfDelegationTooLowToUnjail, err))
}
