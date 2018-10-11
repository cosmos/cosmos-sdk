package slashing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

func TestCannotUnjailUnlessJailed(t *testing.T) {
	// initial setup
	ctx, ck, sk, _, keeper := createTestInput(t)
	slh := NewHandler(keeper)
	amtInt := int64(100)
	addr, val, amt := addrs[0], pks[0], sdk.NewInt(amtInt)
	msg := newTestMsgCreateValidator(addr, val, amt)
	got := stake.NewHandler(sk)(ctx, msg)
	require.True(t, got.IsOK())
	stake.EndBlocker(ctx, sk)
	require.Equal(t, ck.GetCoins(ctx, sdk.AccAddress(addr)), sdk.Coins{{sk.GetParams(ctx).BondDenom, initCoins.Sub(amt)}})
	require.True(t, sdk.NewDecFromInt(amt).Equal(sk.Validator(ctx, addr).GetPower()))

	// assert non-jailed validator can't be unjailed
	got = slh(ctx, NewMsgUnjail(addr))
	require.False(t, got.IsOK(), "allowed unjail of non-jailed validator")
	require.Equal(t, sdk.ToABCICode(DefaultCodespace, CodeValidatorNotJailed), got.Code)
}

func TestJailedValidatorDelegations(t *testing.T) {
	ctx, _, stakeKeeper, _, slashingKeeper := createTestInput(t)

	stakeParams := stakeKeeper.GetParams(ctx)
	stakeParams.UnbondingTime = 0
	stakeKeeper.SetParams(ctx, stakeParams)

	// create a validator
	amount := int64(10)
	valPubKey, bondAmount := pks[0], sdk.NewInt(amount)
	valAddr, consAddr := addrs[1], sdk.ConsAddress(addrs[0])

	msgCreateVal := newTestMsgCreateValidator(valAddr, valPubKey, bondAmount)
	got := stake.NewHandler(stakeKeeper)(ctx, msgCreateVal)
	require.True(t, got.IsOK(), "expected create validator msg to be ok, got: %v", got)

	// end block
	stake.EndBlocker(ctx, stakeKeeper)

	// set dummy signing info
	newInfo := ValidatorSigningInfo{
		StartHeight:         int64(0),
		IndexOffset:         int64(0),
		JailedUntil:         time.Unix(0, 0),
		MissedBlocksCounter: int64(0),
	}
	slashingKeeper.setValidatorSigningInfo(ctx, consAddr, newInfo)

	// delegate tokens to the validator
	delAddr := sdk.AccAddress(addrs[2])
	msgDelegate := newTestMsgDelegate(delAddr, valAddr, bondAmount)
	got = stake.NewHandler(stakeKeeper)(ctx, msgDelegate)
	require.True(t, got.IsOK(), "expected delegation to be ok, got %v", got)

	unbondShares := sdk.NewDec(10)

	// unbond validator total self-delegations (which should jail the validator)
	msgBeginUnbonding := stake.NewMsgBeginUnbonding(sdk.AccAddress(valAddr), valAddr, unbondShares)
	got = stake.NewHandler(stakeKeeper)(ctx, msgBeginUnbonding)
	require.True(t, got.IsOK(), "expected begin unbonding validator msg to be ok, got: %v", got)

	err := stakeKeeper.CompleteUnbonding(ctx, sdk.AccAddress(valAddr), valAddr)
	require.Nil(t, err, "expected complete unbonding validator to be ok, got: %v", err)

	// verify validator still exists and is jailed
	validator, found := stakeKeeper.GetValidator(ctx, valAddr)
	require.True(t, found)
	require.True(t, validator.GetJailed())

	// verify the validator cannot unjail itself
	got = NewHandler(slashingKeeper)(ctx, NewMsgUnjail(valAddr))
	require.False(t, got.IsOK(), "expected jailed validator to not be able to unjail, got: %v", got)

	// self-delegate to validator
	msgSelfDelegate := newTestMsgDelegate(sdk.AccAddress(valAddr), valAddr, bondAmount)
	got = stake.NewHandler(stakeKeeper)(ctx, msgSelfDelegate)
	require.True(t, got.IsOK(), "expected delegation to not be ok, got %v", got)

	// verify the validator can now unjail itself
	got = NewHandler(slashingKeeper)(ctx, NewMsgUnjail(valAddr))
	require.True(t, got.IsOK(), "expected jailed validator to be able to unjail, got: %v", got)
}
