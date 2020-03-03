package slashing_test

import (
	"errors"
	"strings"
	"testing"
	"time"

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

func TestJailedValidatorDelegations(t *testing.T) {
	// initial setup
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{Time: time.Unix(0, 0)})

	pks := simapp.CreateTestPubKeys(3)
	simapp.AddTestAddrsFromPubKeys(app, ctx, pks, sdk.TokensFromConsensusPower(20))
	app.SlashingKeeper.SetParams(ctx, slashingkeeper.TestParams())

	stakingParams := app.StakingKeeper.GetParams(ctx)
	app.StakingKeeper.SetParams(ctx, stakingParams)

	// create a validator
	bondAmount := sdk.TokensFromConsensusPower(10)
	valPubKey := pks[1]
	valAddr, consAddr := sdk.ValAddress(pks[1].Address()), sdk.ConsAddress(pks[0].Address())

	msgCreateVal := slashingkeeper.NewTestMsgCreateValidator(valAddr, valPubKey, bondAmount)
	res, err := staking.NewHandler(app.StakingKeeper)(ctx, msgCreateVal)
	require.NoError(t, err)
	require.NotNil(t, res)

	// end block
	staking.EndBlocker(ctx, app.StakingKeeper)

	// set dummy signing info
	newInfo := slashing.NewValidatorSigningInfo(consAddr, 0, 0, time.Unix(0, 0), false, 0)
	app.SlashingKeeper.SetValidatorSigningInfo(ctx, consAddr, newInfo)

	// delegate tokens to the validator
	delAddr := sdk.AccAddress(pks[2].Address())
	msgDelegate := slashingkeeper.NewTestMsgDelegate(delAddr, valAddr, bondAmount)
	res, err = staking.NewHandler(app.StakingKeeper)(ctx, msgDelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	unbondAmt := sdk.NewCoin(app.StakingKeeper.GetParams(ctx).BondDenom, bondAmount)

	// unbond validator total self-delegations (which should jail the validator)
	msgUndelegate := staking.NewMsgUndelegate(sdk.AccAddress(valAddr), valAddr, unbondAmt)
	res, err = staking.NewHandler(app.StakingKeeper)(ctx, msgUndelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	err = app.StakingKeeper.CompleteUnbonding(ctx, sdk.AccAddress(valAddr), valAddr)
	require.Nil(t, err, "expected complete unbonding validator to be ok, got: %v", err)

	// verify validator still exists and is jailed
	validator, found := app.StakingKeeper.GetValidator(ctx, valAddr)
	require.True(t, found)
	require.True(t, validator.IsJailed())

	// verify the validator cannot unjail itself
	res, err = slashing.NewHandler(app.SlashingKeeper)(ctx, slashing.NewMsgUnjail(valAddr))
	require.Error(t, err)
	require.Nil(t, res)

	// self-delegate to validator
	msgSelfDelegate := slashingkeeper.NewTestMsgDelegate(sdk.AccAddress(valAddr), valAddr, bondAmount)
	res, err = staking.NewHandler(app.StakingKeeper)(ctx, msgSelfDelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	// verify the validator can now unjail itself
	res, err = slashing.NewHandler(app.SlashingKeeper)(ctx, slashing.NewMsgUnjail(valAddr))
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestInvalidMsg(t *testing.T) {
	k := slashing.Keeper{}
	h := slashing.NewHandler(k)

	res, err := h(sdk.NewContext(nil, abci.Header{}, false, nil), sdk.NewTestMsg())
	require.Error(t, err)
	require.Nil(t, res)
	require.True(t, strings.Contains(err.Error(), "unrecognized slashing message type"))
}
