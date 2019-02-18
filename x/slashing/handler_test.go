package slashing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

func TestCannotUnjailUnlessJailed(t *testing.T) {
	// initial setup
	ctx, ck, sk, _, keeper := createTestInput(t, DefaultParams())
	slh := NewHandler(keeper)
	amt := sdk.TokensFromTendermintPower(100)
	addr, val := addrs[0], pks[0]
	msg := NewTestMsgCreateValidator(addr, val, amt)
	got := staking.NewHandler(sk)(ctx, msg)
	require.True(t, got.IsOK(), "%v", got)
	staking.EndBlocker(ctx, sk)

	require.Equal(
		t, ck.GetCoins(ctx, sdk.AccAddress(addr)),
		sdk.Coins{sdk.NewCoin(sk.GetParams(ctx).BondDenom, initCoins.Sub(amt))},
	)
	require.Equal(t, amt, sk.Validator(ctx, addr).GetBondedTokens())

	// assert non-jailed validator can't be unjailed
	got = slh(ctx, NewMsgUnjail(addr))
	require.False(t, got.IsOK(), "allowed unjail of non-jailed validator")
	require.EqualValues(t, CodeValidatorNotJailed, got.Code)
	require.EqualValues(t, DefaultCodespace, got.Codespace)
}

func TestCannotUnjailUnlessMeetMinSelfDelegation(t *testing.T) {
	// initial setup
	ctx, ck, sk, _, keeper := createTestInput(t, DefaultParams())
	slh := NewHandler(keeper)
	amtInt := int64(100)
	addr, val, amt := addrs[0], pks[0], sdk.TokensFromTendermintPower(amtInt)
	msg := NewTestMsgCreateValidator(addr, val, amt)
	msg.MinSelfDelegation = amt
	got := staking.NewHandler(sk)(ctx, msg)
	require.True(t, got.IsOK())
	staking.EndBlocker(ctx, sk)

	require.Equal(
		t, ck.GetCoins(ctx, sdk.AccAddress(addr)),
		sdk.Coins{sdk.NewCoin(sk.GetParams(ctx).BondDenom, initCoins.Sub(amt))},
	)

	undelegateMsg := staking.NewMsgUndelegate(sdk.AccAddress(addr), addr, sdk.OneDec())
	got = staking.NewHandler(sk)(ctx, undelegateMsg)

	require.True(t, sk.Validator(ctx, addr).GetJailed())

	// assert non-jailed validator can't be unjailed
	got = slh(ctx, NewMsgUnjail(addr))
	require.False(t, got.IsOK(), "allowed unjail of validator with less than MinSelfDelegation")
	require.EqualValues(t, CodeValidatorNotJailed, got.Code)
	require.EqualValues(t, DefaultCodespace, got.Codespace)
}

func TestJailedValidatorDelegations(t *testing.T) {
	ctx, _, stakingKeeper, _, slashingKeeper := createTestInput(t, DefaultParams())

	stakingParams := stakingKeeper.GetParams(ctx)
	stakingParams.UnbondingTime = 0
	stakingKeeper.SetParams(ctx, stakingParams)

	// create a validator
	bondAmount := sdk.TokensFromTendermintPower(10)
	valPubKey := pks[0]
	valAddr, consAddr := addrs[1], sdk.ConsAddress(addrs[0])

	msgCreateVal := NewTestMsgCreateValidator(valAddr, valPubKey, bondAmount)
	got := staking.NewHandler(stakingKeeper)(ctx, msgCreateVal)
	require.True(t, got.IsOK(), "expected create validator msg to be ok, got: %v", got)

	// end block
	staking.EndBlocker(ctx, stakingKeeper)

	// set dummy signing info
	newInfo := NewValidatorSigningInfo(0, 0, time.Unix(0, 0), false, 0)
	slashingKeeper.SetValidatorSigningInfo(ctx, consAddr, newInfo)

	// delegate tokens to the validator
	delAddr := sdk.AccAddress(addrs[2])
	msgDelegate := newTestMsgDelegate(delAddr, valAddr, bondAmount)
	got = staking.NewHandler(stakingKeeper)(ctx, msgDelegate)
	require.True(t, got.IsOK(), "expected delegation to be ok, got %v", got)

	unbondShares := sdk.NewDecFromInt(bondAmount)

	// unbond validator total self-delegations (which should jail the validator)
	msgUndelegate := staking.NewMsgUndelegate(sdk.AccAddress(valAddr), valAddr, unbondShares)
	got = staking.NewHandler(stakingKeeper)(ctx, msgUndelegate)
	require.True(t, got.IsOK(), "expected begin unbonding validator msg to be ok, got: %v", got)

	err := stakingKeeper.CompleteUnbonding(ctx, sdk.AccAddress(valAddr), valAddr)
	require.Nil(t, err, "expected complete unbonding validator to be ok, got: %v", err)

	// verify validator still exists and is jailed
	validator, found := stakingKeeper.GetValidator(ctx, valAddr)
	require.True(t, found)
	require.True(t, validator.GetJailed())

	// verify the validator cannot unjail itself
	got = NewHandler(slashingKeeper)(ctx, NewMsgUnjail(valAddr))
	require.False(t, got.IsOK(), "expected jailed validator to not be able to unjail, got: %v", got)

	// self-delegate to validator
	msgSelfDelegate := newTestMsgDelegate(sdk.AccAddress(valAddr), valAddr, bondAmount)
	got = staking.NewHandler(stakingKeeper)(ctx, msgSelfDelegate)
	require.True(t, got.IsOK(), "expected delegation to not be ok, got %v", got)

	// verify the validator can now unjail itself
	got = NewHandler(slashingKeeper)(ctx, NewMsgUnjail(valAddr))
	require.True(t, got.IsOK(), "expected jailed validator to be able to unjail, got: %v", got)
}
