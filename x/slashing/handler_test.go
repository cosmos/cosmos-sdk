package slashing_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestCannotUnjailUnlessJailed(t *testing.T) {
	// initial setup
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	pks := simapp.CreateTestPubKeys(1)
	simapp.AddTestAddrsFromPubKeys(app, ctx, pks, sdk.TokensFromConsensusPower(200))

	slh := slashing.NewHandler(app.SlashingKeeper)
	amt := sdk.TokensFromConsensusPower(100)
	addr, val := sdk.ValAddress(pks[0].Address()), pks[0]

	msg := keeper.NewTestMsgCreateValidator(addr, val, amt)
	res, err := staking.NewHandler(app.StakingKeeper)(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	staking.EndBlocker(ctx, app.StakingKeeper)

	require.Equal(
		t, app.BankKeeper.GetAllBalances(ctx, sdk.AccAddress(addr)),
		sdk.Coins{sdk.NewCoin(app.StakingKeeper.GetParams(ctx).BondDenom, keeper.InitTokens.Sub(amt))},
	)
	require.Equal(t, amt, app.StakingKeeper.Validator(ctx, addr).GetBondedTokens())

	// assert non-jailed validator can't be unjailed
	res, err = slh(ctx, types.NewMsgUnjail(addr))
	require.Error(t, err)
	require.Nil(t, res)
	require.True(t, errors.Is(types.ErrValidatorNotJailed, err))
}

func TestCannotUnjailUnlessMeetMinSelfDelegation(t *testing.T) {
	// initial setup
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	pks := simapp.CreateTestPubKeys(1)
	simapp.AddTestAddrsFromPubKeys(app, ctx, pks, sdk.TokensFromConsensusPower(200))

	slh := slashing.NewHandler(app.SlashingKeeper)
	amtInt := int64(100)
	addr, val, amt := sdk.ValAddress(pks[0].Address()), pks[0], sdk.TokensFromConsensusPower(amtInt)
	msg := keeper.NewTestMsgCreateValidator(addr, val, amt)
	msg.MinSelfDelegation = amt

	res, err := staking.NewHandler(app.StakingKeeper)(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	staking.EndBlocker(ctx, app.StakingKeeper)

	require.Equal(
		t, app.BankKeeper.GetAllBalances(ctx, sdk.AccAddress(addr)),
		sdk.Coins{sdk.NewCoin(app.StakingKeeper.GetParams(ctx).BondDenom, keeper.InitTokens.Sub(amt))},
	)

	unbondAmt := sdk.NewCoin(app.StakingKeeper.GetParams(ctx).BondDenom, sdk.OneInt())
	undelegateMsg := stakingtypes.NewMsgUndelegate(sdk.AccAddress(addr), addr, unbondAmt)
	res, err = staking.NewHandler(app.StakingKeeper)(ctx, undelegateMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	require.True(t, app.StakingKeeper.Validator(ctx, addr).IsJailed())

	// assert non-jailed validator can't be unjailed
	res, err = slh(ctx, types.NewMsgUnjail(addr))
	require.Error(t, err)
	require.Nil(t, res)
	require.True(t, errors.Is(types.ErrSelfDelegationTooLowToUnjail, err))
}

func TestJailedValidatorDelegations(t *testing.T) {
	// initial setup
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Time: time.Unix(0, 0)})

	pks := simapp.CreateTestPubKeys(3)
	simapp.AddTestAddrsFromPubKeys(app, ctx, pks, sdk.TokensFromConsensusPower(20))
	app.SlashingKeeper.SetParams(ctx, keeper.TestParams())

	stakingParams := app.StakingKeeper.GetParams(ctx)
	app.StakingKeeper.SetParams(ctx, stakingParams)

	// create a validator
	bondAmount := sdk.TokensFromConsensusPower(10)
	valPubKey := pks[1]
	valAddr, consAddr := sdk.ValAddress(pks[1].Address()), sdk.ConsAddress(pks[0].Address())

	msgCreateVal := keeper.NewTestMsgCreateValidator(valAddr, valPubKey, bondAmount)
	res, err := staking.NewHandler(app.StakingKeeper)(ctx, msgCreateVal)
	require.NoError(t, err)
	require.NotNil(t, res)

	// end block
	staking.EndBlocker(ctx, app.StakingKeeper)

	// set dummy signing info
	newInfo := types.NewValidatorSigningInfo(consAddr, 0, 0, time.Unix(0, 0), false, 0)
	app.SlashingKeeper.SetValidatorSigningInfo(ctx, consAddr, newInfo)

	// delegate tokens to the validator
	delAddr := sdk.AccAddress(pks[2].Address())
	msgDelegate := keeper.NewTestMsgDelegate(delAddr, valAddr, bondAmount)
	res, err = staking.NewHandler(app.StakingKeeper)(ctx, msgDelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	unbondAmt := sdk.NewCoin(app.StakingKeeper.GetParams(ctx).BondDenom, bondAmount)

	// unbond validator total self-delegations (which should jail the validator)
	msgUndelegate := stakingtypes.NewMsgUndelegate(sdk.AccAddress(valAddr), valAddr, unbondAmt)
	res, err = staking.NewHandler(app.StakingKeeper)(ctx, msgUndelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	_, err = app.StakingKeeper.CompleteUnbonding(ctx, sdk.AccAddress(valAddr), valAddr)
	require.Nil(t, err, "expected complete unbonding validator to be ok, got: %v", err)

	// verify validator still exists and is jailed
	validator, found := app.StakingKeeper.GetValidator(ctx, valAddr)
	require.True(t, found)
	require.True(t, validator.IsJailed())

	// verify the validator cannot unjail itself
	res, err = slashing.NewHandler(app.SlashingKeeper)(ctx, types.NewMsgUnjail(valAddr))
	require.Error(t, err)
	require.Nil(t, res)

	// self-delegate to validator
	msgSelfDelegate := keeper.NewTestMsgDelegate(sdk.AccAddress(valAddr), valAddr, bondAmount)
	res, err = staking.NewHandler(app.StakingKeeper)(ctx, msgSelfDelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	// verify the validator can now unjail itself
	res, err = slashing.NewHandler(app.SlashingKeeper)(ctx, types.NewMsgUnjail(valAddr))
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestInvalidMsg(t *testing.T) {
	k := keeper.Keeper{}
	h := slashing.NewHandler(k)

	res, err := h(sdk.NewContext(nil, tmproto.Header{}, false, nil), testdata.NewTestMsg())
	require.Error(t, err)
	require.Nil(t, res)
	require.True(t, strings.Contains(err.Error(), "unrecognized slashing message type"))
}

// Test a validator through uptime, downtime, revocation,
// unrevocation, starting height reset, and revocation again
func TestHandleAbsentValidator(t *testing.T) {
	// initial setup
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Time: time.Unix(0, 0)})

	pks := simapp.CreateTestPubKeys(1)
	simapp.AddTestAddrsFromPubKeys(app, ctx, pks, sdk.TokensFromConsensusPower(200))
	app.SlashingKeeper.SetParams(ctx, keeper.TestParams())

	power := int64(100)
	amt := sdk.TokensFromConsensusPower(power)
	addr, val := sdk.ValAddress(pks[0].Address()), pks[0]
	sh := staking.NewHandler(app.StakingKeeper)
	slh := slashing.NewHandler(app.SlashingKeeper)

	res, err := sh(ctx, keeper.NewTestMsgCreateValidator(addr, val, amt))
	require.NoError(t, err)
	require.NotNil(t, res)

	staking.EndBlocker(ctx, app.StakingKeeper)

	require.Equal(
		t, app.BankKeeper.GetAllBalances(ctx, sdk.AccAddress(addr)),
		sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.GetParams(ctx).BondDenom, keeper.InitTokens.Sub(amt))),
	)
	require.Equal(t, amt, app.StakingKeeper.Validator(ctx, addr).GetBondedTokens())

	// will exist since the validator has been bonded
	info, found := app.SlashingKeeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, int64(0), info.IndexOffset)
	require.Equal(t, int64(0), info.MissedBlocksCounter)
	require.Equal(t, time.Unix(0, 0).UTC(), info.JailedUntil)
	height := int64(0)

	// 1000 first blocks OK
	for ; height < app.SlashingKeeper.SignedBlocksWindow(ctx); height++ {
		ctx = ctx.WithBlockHeight(height)
		app.SlashingKeeper.HandleValidatorSignature(ctx, val.Address(), power, true)
	}
	info, found = app.SlashingKeeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, int64(0), info.MissedBlocksCounter)

	// 500 blocks missed
	for ; height < app.SlashingKeeper.SignedBlocksWindow(ctx)+(app.SlashingKeeper.SignedBlocksWindow(ctx)-app.SlashingKeeper.MinSignedPerWindow(ctx)); height++ {
		ctx = ctx.WithBlockHeight(height)
		app.SlashingKeeper.HandleValidatorSignature(ctx, val.Address(), power, false)
	}
	info, found = app.SlashingKeeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, app.SlashingKeeper.SignedBlocksWindow(ctx)-app.SlashingKeeper.MinSignedPerWindow(ctx), info.MissedBlocksCounter)

	// validator should be bonded still
	validator, _ := app.StakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, stakingtypes.Bonded, validator.GetStatus())

	bondPool := app.StakingKeeper.GetBondedPool(ctx)
	require.True(sdk.IntEq(t, amt, app.BankKeeper.GetBalance(ctx, bondPool.GetAddress(), app.StakingKeeper.BondDenom(ctx)).Amount))

	// 501st block missed
	ctx = ctx.WithBlockHeight(height)
	app.SlashingKeeper.HandleValidatorSignature(ctx, val.Address(), power, false)
	info, found = app.SlashingKeeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	// counter now reset to zero
	require.Equal(t, int64(0), info.MissedBlocksCounter)

	// end block
	staking.EndBlocker(ctx, app.StakingKeeper)

	// validator should have been jailed
	validator, _ = app.StakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, stakingtypes.Unbonding, validator.GetStatus())

	slashAmt := amt.ToDec().Mul(app.SlashingKeeper.SlashFractionDowntime(ctx)).RoundInt64()

	// validator should have been slashed
	require.Equal(t, amt.Int64()-slashAmt, validator.GetTokens().Int64())

	// 502nd block *also* missed (since the LastCommit would have still included the just-unbonded validator)
	height++
	ctx = ctx.WithBlockHeight(height)
	app.SlashingKeeper.HandleValidatorSignature(ctx, val.Address(), power, false)
	info, found = app.SlashingKeeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, int64(1), info.MissedBlocksCounter)

	// end block
	staking.EndBlocker(ctx, app.StakingKeeper)

	// validator should not have been slashed any more, since it was already jailed
	validator, _ = app.StakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, amt.Int64()-slashAmt, validator.GetTokens().Int64())

	// unrevocation should fail prior to jail expiration
	res, err = slh(ctx, types.NewMsgUnjail(addr))
	require.Error(t, err)
	require.Nil(t, res)

	// unrevocation should succeed after jail expiration
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: time.Unix(1, 0).Add(app.SlashingKeeper.DowntimeJailDuration(ctx))})
	res, err = slh(ctx, types.NewMsgUnjail(addr))
	require.NoError(t, err)
	require.NotNil(t, res)

	// end block
	staking.EndBlocker(ctx, app.StakingKeeper)

	// validator should be rebonded now
	validator, _ = app.StakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, stakingtypes.Bonded, validator.GetStatus())

	// validator should have been slashed
	require.Equal(t, amt.Int64()-slashAmt, app.BankKeeper.GetBalance(ctx, bondPool.GetAddress(), app.StakingKeeper.BondDenom(ctx)).Amount.Int64())

	// Validator start height should not have been changed
	info, found = app.SlashingKeeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	// we've missed 2 blocks more than the maximum, so the counter was reset to 0 at 1 block more and is now 1
	require.Equal(t, int64(1), info.MissedBlocksCounter)

	// validator should not be immediately jailed again
	height++
	ctx = ctx.WithBlockHeight(height)
	app.SlashingKeeper.HandleValidatorSignature(ctx, val.Address(), power, false)
	validator, _ = app.StakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, stakingtypes.Bonded, validator.GetStatus())

	// 500 signed blocks
	nextHeight := height + app.SlashingKeeper.MinSignedPerWindow(ctx) + 1
	for ; height < nextHeight; height++ {
		ctx = ctx.WithBlockHeight(height)
		app.SlashingKeeper.HandleValidatorSignature(ctx, val.Address(), power, false)
	}

	// end block
	staking.EndBlocker(ctx, app.StakingKeeper)

	// validator should be jailed again after 500 unsigned blocks
	nextHeight = height + app.SlashingKeeper.MinSignedPerWindow(ctx) + 1
	for ; height <= nextHeight; height++ {
		ctx = ctx.WithBlockHeight(height)
		app.SlashingKeeper.HandleValidatorSignature(ctx, val.Address(), power, false)
	}

	// end block
	staking.EndBlocker(ctx, app.StakingKeeper)

	validator, _ = app.StakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, stakingtypes.Unbonding, validator.GetStatus())
}
