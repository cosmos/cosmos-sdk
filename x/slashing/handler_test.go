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
	msg := newTestMsgCreateValidator(sdk.ValAddress(addr), val, amt)
	got := stake.NewHandler(sk)(ctx, msg)
	require.True(t, got.IsOK())
	stake.EndBlocker(ctx, sk)
	require.Equal(t, ck.GetCoins(ctx, addr), sdk.Coins{{sk.GetParams(ctx).BondDenom, initCoins.Sub(amt)}})
	require.True(t, sdk.NewDecFromInt(amt).Equal(sk.Validator(ctx, sdk.ValAddress(addr)).GetPower()))

	// assert non-jailed validator can't be unjailed
	got = slh(ctx, NewMsgUnjail(sdk.ValAddress(addr)))
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
	valAddr, valPubKey, bondAmount := sdk.ValAddress(addrs[0]), pks[0], sdk.NewInt(amount)
	msgCreateVal := newTestMsgCreateValidator(valAddr, valPubKey, bondAmount)
	got := stake.NewHandler(stakeKeeper)(ctx, msgCreateVal)
	require.True(t, got.IsOK(), "expected create validator msg to be ok, got: %v", got)

	// set dummy signing info
	newInfo := ValidatorSigningInfo{
		StartHeight:         int64(0),
		IndexOffset:         int64(0),
		JailedUntil:         time.Unix(0, 0),
		SignedBlocksCounter: int64(0),
	}
	slashingKeeper.setValidatorSigningInfo(ctx, valAddr, newInfo)

	// delegate tokens to the validator
	delAddr := addrs[1]
	msgDelegate := newTestMsgDelegate(delAddr, valAddr, bondAmount)
	got = stake.NewHandler(stakeKeeper)(ctx, msgDelegate)
	require.True(t, got.IsOK(), "expected delegation to be ok, got %v", got)

	unbondShares := sdk.NewDec(10)

	// unbond validator total self-delegations (which should jail the validator)
	msgBeginUnbonding := stake.NewMsgBeginUnbonding(sdk.AccAddress(valAddr), valAddr, unbondShares)
	got = stake.NewHandler(stakeKeeper)(ctx, msgBeginUnbonding)
	require.True(t, got.IsOK(), "expected begin unbonding validator msg to be ok, got: %v", got)

	msgCompleteUnbonding := stake.NewMsgCompleteUnbonding(sdk.AccAddress(valAddr), valAddr)
	got = stake.NewHandler(stakeKeeper)(ctx, msgCompleteUnbonding)
	require.True(t, got.IsOK(), "expected complete unbonding validator msg to be ok, got: %v", got)

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
