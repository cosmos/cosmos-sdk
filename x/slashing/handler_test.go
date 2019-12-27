package slashing

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/internal/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

func TestCannotUnjailUnlessJailed(t *testing.T) {
	// initial setup
	ctx, ck, sk, _, keeper := slashingkeeper.CreateTestInput(t, DefaultParams())
	slh := NewHandler(keeper)
	amt := sdk.TokensFromConsensusPower(100)
	addr, val := slashingkeeper.Addrs[0], slashingkeeper.Pks[0]

	msg := slashingkeeper.NewTestMsgCreateValidator(addr, val, amt)
	res, err := staking.NewHandler(sk)(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	staking.EndBlocker(ctx, sk)

	require.Equal(
		t, ck.GetCoins(ctx, sdk.AccAddress(addr)),
		sdk.Coins{sdk.NewCoin(sk.GetParams(ctx).BondDenom, slashingkeeper.InitTokens.Sub(amt))},
	)
	require.Equal(t, amt, sk.Validator(ctx, addr).GetBondedTokens())

	// assert non-jailed validator can't be unjailed
	res, err = slh(ctx, NewMsgUnjail(addr))
	require.Error(t, err)
	require.Nil(t, res)
	require.True(t, errors.Is(ErrValidatorNotJailed, err))
}

func TestCannotUnjailUnlessMeetMinSelfDelegation(t *testing.T) {
	// initial setup
	ctx, ck, sk, _, keeper := slashingkeeper.CreateTestInput(t, DefaultParams())
	slh := NewHandler(keeper)
	amtInt := int64(100)
	addr, val, amt := slashingkeeper.Addrs[0], slashingkeeper.Pks[0], sdk.TokensFromConsensusPower(amtInt)
	msg := slashingkeeper.NewTestMsgCreateValidator(addr, val, amt)
	msg.MinSelfDelegation = amt

	res, err := staking.NewHandler(sk)(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, res)

	staking.EndBlocker(ctx, sk)

	require.Equal(
		t, ck.GetCoins(ctx, sdk.AccAddress(addr)),
		sdk.Coins{sdk.NewCoin(sk.GetParams(ctx).BondDenom, slashingkeeper.InitTokens.Sub(amt))},
	)

	unbondAmt := sdk.NewCoin(sk.GetParams(ctx).BondDenom, sdk.OneInt())
	undelegateMsg := staking.NewMsgUndelegate(sdk.AccAddress(addr), addr, unbondAmt)
	res, err = staking.NewHandler(sk)(ctx, undelegateMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	require.True(t, sk.Validator(ctx, addr).IsJailed())

	// assert non-jailed validator can't be unjailed
	res, err = slh(ctx, NewMsgUnjail(addr))
	require.Error(t, err)
	require.Nil(t, res)
	require.True(t, errors.Is(ErrSelfDelegationTooLowToUnjail, err))
}

func TestJailedValidatorDelegations(t *testing.T) {
	ctx, _, stakingKeeper, _, slashingKeeper := slashingkeeper.CreateTestInput(t, DefaultParams())

	stakingParams := stakingKeeper.GetParams(ctx)
	stakingKeeper.SetParams(ctx, stakingParams)

	// create a validator
	bondAmount := sdk.TokensFromConsensusPower(10)
	valPubKey := slashingkeeper.Pks[0]
	valAddr, consAddr := slashingkeeper.Addrs[1], sdk.ConsAddress(slashingkeeper.Addrs[0])

	msgCreateVal := slashingkeeper.NewTestMsgCreateValidator(valAddr, valPubKey, bondAmount)
	res, err := staking.NewHandler(stakingKeeper)(ctx, msgCreateVal)
	require.NoError(t, err)
	require.NotNil(t, res)

	// end block
	staking.EndBlocker(ctx, stakingKeeper)

	// set dummy signing info
	newInfo := NewValidatorSigningInfo(consAddr, 0, 0, time.Unix(0, 0), false, 0)
	slashingKeeper.SetValidatorSigningInfo(ctx, consAddr, newInfo)

	// delegate tokens to the validator
	delAddr := sdk.AccAddress(slashingkeeper.Addrs[2])
	msgDelegate := slashingkeeper.NewTestMsgDelegate(delAddr, valAddr, bondAmount)
	res, err = staking.NewHandler(stakingKeeper)(ctx, msgDelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	unbondAmt := sdk.NewCoin(stakingKeeper.GetParams(ctx).BondDenom, bondAmount)

	// unbond validator total self-delegations (which should jail the validator)
	msgUndelegate := staking.NewMsgUndelegate(sdk.AccAddress(valAddr), valAddr, unbondAmt)
	res, err = staking.NewHandler(stakingKeeper)(ctx, msgUndelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	err = stakingKeeper.CompleteUnbonding(ctx, sdk.AccAddress(valAddr), valAddr)
	require.Nil(t, err, "expected complete unbonding validator to be ok, got: %v", err)

	// verify validator still exists and is jailed
	validator, found := stakingKeeper.GetValidator(ctx, valAddr)
	require.True(t, found)
	require.True(t, validator.IsJailed())

	// verify the validator cannot unjail itself
	res, err = NewHandler(slashingKeeper)(ctx, NewMsgUnjail(valAddr))
	require.Error(t, err)
	require.Nil(t, res)

	// self-delegate to validator
	msgSelfDelegate := slashingkeeper.NewTestMsgDelegate(sdk.AccAddress(valAddr), valAddr, bondAmount)
	res, err = staking.NewHandler(stakingKeeper)(ctx, msgSelfDelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	// verify the validator can now unjail itself
	res, err = NewHandler(slashingKeeper)(ctx, NewMsgUnjail(valAddr))
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestInvalidMsg(t *testing.T) {
	k := Keeper{}
	h := NewHandler(k)

	res, err := h(sdk.NewContext(nil, abci.Header{}, false, nil), sdk.NewTestMsg())
	require.Error(t, err)
	require.Nil(t, res)
	require.True(t, strings.Contains(err.Error(), "unrecognized slashing message type"))
}

// Test a validator through uptime, downtime, revocation,
// unrevocation, starting height reset, and revocation again
func TestHandleAbsentValidator(t *testing.T) {

	// initial setup
	ctx, ck, sk, _, keeper := slashingkeeper.CreateTestInput(t, slashingkeeper.TestParams())
	power := int64(100)
	amt := sdk.TokensFromConsensusPower(power)
	addr, val := slashingkeeper.Addrs[0], slashingkeeper.Pks[0]
	sh := staking.NewHandler(sk)
	slh := NewHandler(keeper)

	res, err := sh(ctx, slashingkeeper.NewTestMsgCreateValidator(addr, val, amt))
	require.NoError(t, err)
	require.NotNil(t, res)

	staking.EndBlocker(ctx, sk)

	require.Equal(
		t, ck.GetCoins(ctx, sdk.AccAddress(addr)),
		sdk.NewCoins(sdk.NewCoin(sk.GetParams(ctx).BondDenom, slashingkeeper.InitTokens.Sub(amt))),
	)
	require.Equal(t, amt, sk.Validator(ctx, addr).GetBondedTokens())

	// will exist since the validator has been bonded
	info, found := keeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, int64(0), info.IndexOffset)
	require.Equal(t, int64(0), info.MissedBlocksCounter)
	require.Equal(t, time.Unix(0, 0).UTC(), info.JailedUntil)
	height := int64(0)

	// 1000 first blocks OK
	for ; height < keeper.SignedBlocksWindow(ctx); height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.HandleValidatorSignature(ctx, val.Address(), power, true)
	}
	info, found = keeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, int64(0), info.MissedBlocksCounter)

	// 500 blocks missed
	for ; height < keeper.SignedBlocksWindow(ctx)+(keeper.SignedBlocksWindow(ctx)-keeper.MinSignedPerWindow(ctx)); height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.HandleValidatorSignature(ctx, val.Address(), power, false)
	}
	info, found = keeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, keeper.SignedBlocksWindow(ctx)-keeper.MinSignedPerWindow(ctx), info.MissedBlocksCounter)

	// validator should be bonded still
	validator, _ := sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Bonded, validator.GetStatus())
	bondPool := sk.GetBondedPool(ctx)
	require.True(sdk.IntEq(t, amt, bondPool.GetCoins().AmountOf(sk.BondDenom(ctx))))

	// 501st block missed
	ctx = ctx.WithBlockHeight(height)
	keeper.HandleValidatorSignature(ctx, val.Address(), power, false)
	info, found = keeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	// counter now reset to zero
	require.Equal(t, int64(0), info.MissedBlocksCounter)

	// end block
	staking.EndBlocker(ctx, sk)

	// validator should have been jailed
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Unbonding, validator.GetStatus())

	slashAmt := amt.ToDec().Mul(keeper.SlashFractionDowntime(ctx)).RoundInt64()

	// validator should have been slashed
	require.Equal(t, amt.Int64()-slashAmt, validator.GetTokens().Int64())

	// 502nd block *also* missed (since the LastCommit would have still included the just-unbonded validator)
	height++
	ctx = ctx.WithBlockHeight(height)
	keeper.HandleValidatorSignature(ctx, val.Address(), power, false)
	info, found = keeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, int64(1), info.MissedBlocksCounter)

	// end block
	staking.EndBlocker(ctx, sk)

	// validator should not have been slashed any more, since it was already jailed
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, amt.Int64()-slashAmt, validator.GetTokens().Int64())

	// unrevocation should fail prior to jail expiration
	res, err = slh(ctx, types.NewMsgUnjail(addr))
	require.Error(t, err)
	require.Nil(t, res)

	// unrevocation should succeed after jail expiration
	ctx = ctx.WithBlockHeader(abci.Header{Time: time.Unix(1, 0).Add(keeper.DowntimeJailDuration(ctx))})
	res, err = slh(ctx, types.NewMsgUnjail(addr))
	require.NoError(t, err)
	require.NotNil(t, res)

	// end block
	staking.EndBlocker(ctx, sk)

	// validator should be rebonded now
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Bonded, validator.GetStatus())

	// validator should have been slashed
	bondPool = sk.GetBondedPool(ctx)
	require.Equal(t, amt.Int64()-slashAmt, bondPool.GetCoins().AmountOf(sk.BondDenom(ctx)).Int64())

	// Validator start height should not have been changed
	info, found = keeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	// we've missed 2 blocks more than the maximum, so the counter was reset to 0 at 1 block more and is now 1
	require.Equal(t, int64(1), info.MissedBlocksCounter)

	// validator should not be immediately jailed again
	height++
	ctx = ctx.WithBlockHeight(height)
	keeper.HandleValidatorSignature(ctx, val.Address(), power, false)
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Bonded, validator.GetStatus())

	// 500 signed blocks
	nextHeight := height + keeper.MinSignedPerWindow(ctx) + 1
	for ; height < nextHeight; height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.HandleValidatorSignature(ctx, val.Address(), power, false)
	}

	// end block
	staking.EndBlocker(ctx, sk)

	// validator should be jailed again after 500 unsigned blocks
	nextHeight = height + keeper.MinSignedPerWindow(ctx) + 1
	for ; height <= nextHeight; height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.HandleValidatorSignature(ctx, val.Address(), power, false)
	}

	// end block
	staking.EndBlocker(ctx, sk)

	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Unbonding, validator.GetStatus())
}
