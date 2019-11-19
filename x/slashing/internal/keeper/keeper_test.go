package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/internal/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// ______________________________________________________________

// Test that a validator is slashed correctly
// when we discover evidence of infraction
func TestHandleDoubleSign(t *testing.T) {

	// initial setup
	ctx, ck, sk, _, keeper := CreateTestInput(t, TestParams())
	// validator added pre-genesis
	ctx = ctx.WithBlockHeight(-1)
	power := int64(100)
	amt := sdk.TokensFromConsensusPower(power)
	operatorAddr, val := Addrs[0], Pks[0]
	got := staking.NewHandler(sk)(ctx, NewTestMsgCreateValidator(operatorAddr, val, amt))
	require.True(t, got.IsOK())
	staking.EndBlocker(ctx, sk)
	require.Equal(
		t, ck.GetCoins(ctx, sdk.AccAddress(operatorAddr)),
		sdk.NewCoins(sdk.NewCoin(sk.GetParams(ctx).BondDenom, InitTokens.Sub(amt))),
	)
	require.Equal(t, amt, sk.Validator(ctx, operatorAddr).GetBondedTokens())

	// handle a signature to set signing info
	keeper.HandleValidatorSignature(ctx, val.Address(), amt.Int64(), true)

	oldTokens := sk.Validator(ctx, operatorAddr).GetTokens()

	// double sign less than max age
	keeper.HandleDoubleSign(ctx, val.Address(), 0, time.Unix(0, 0), power)

	// should be jailed
	require.True(t, sk.Validator(ctx, operatorAddr).IsJailed())

	// tokens should be decreased
	newTokens := sk.Validator(ctx, operatorAddr).GetTokens()
	require.True(t, newTokens.LT(oldTokens))

	// New evidence
	keeper.HandleDoubleSign(ctx, val.Address(), 0, time.Unix(0, 0), power)

	// tokens should be the same (capped slash)
	require.True(t, sk.Validator(ctx, operatorAddr).GetTokens().Equal(newTokens))

	// Jump to past the unbonding period
	ctx = ctx.WithBlockHeader(abci.Header{Time: time.Unix(1, 0).Add(sk.GetParams(ctx).UnbondingTime)})

	// Still shouldn't be able to unjail
	require.Error(t, keeper.Unjail(ctx, operatorAddr))

	// Should be able to unbond now
	del, _ := sk.GetDelegation(ctx, sdk.AccAddress(operatorAddr), operatorAddr)
	validator, _ := sk.GetValidator(ctx, operatorAddr)

	totalBond := validator.TokensFromShares(del.GetShares()).TruncateInt()
	msgUnbond := staking.NewMsgUndelegate(sdk.AccAddress(operatorAddr), operatorAddr, sdk.NewCoin(sk.GetParams(ctx).BondDenom, totalBond))
	res := staking.NewHandler(sk)(ctx, msgUnbond)
	require.True(t, res.IsOK())
}

// ______________________________________________________________

// Test that a validator is slashed correctly
// when we discover evidence of infraction
func TestPastMaxEvidenceAge(t *testing.T) {

	// initial setup
	ctx, ck, sk, _, keeper := CreateTestInput(t, TestParams())
	// validator added pre-genesis
	ctx = ctx.WithBlockHeight(-1)
	power := int64(100)
	amt := sdk.TokensFromConsensusPower(power)
	operatorAddr, val := Addrs[0], Pks[0]
	got := staking.NewHandler(sk)(ctx, NewTestMsgCreateValidator(operatorAddr, val, amt))
	require.True(t, got.IsOK())
	staking.EndBlocker(ctx, sk)
	require.Equal(
		t, ck.GetCoins(ctx, sdk.AccAddress(operatorAddr)),
		sdk.NewCoins(sdk.NewCoin(sk.GetParams(ctx).BondDenom, InitTokens.Sub(amt))),
	)
	require.Equal(t, amt, sk.Validator(ctx, operatorAddr).GetBondedTokens())

	// handle a signature to set signing info
	keeper.HandleValidatorSignature(ctx, val.Address(), power, true)

	ctx = ctx.WithBlockHeader(abci.Header{Time: time.Unix(1, 0).Add(keeper.MaxEvidenceAge(ctx))})

	oldPower := sk.Validator(ctx, operatorAddr).GetConsensusPower()

	// double sign past max age
	keeper.HandleDoubleSign(ctx, val.Address(), 0, time.Unix(0, 0), power)

	// should still be bonded
	require.True(t, sk.Validator(ctx, operatorAddr).IsBonded())

	// should still have same power
	require.Equal(t, oldPower, sk.Validator(ctx, operatorAddr).GetConsensusPower())
}

// Test a new validator entering the validator set
// Ensure that SigningInfo.StartHeight is set correctly
// and that they are not immediately jailed
func TestHandleNewValidator(t *testing.T) {
	// initial setup
	ctx, ck, sk, _, keeper := CreateTestInput(t, TestParams())
	addr, val := Addrs[0], Pks[0]
	amt := sdk.TokensFromConsensusPower(100)
	sh := staking.NewHandler(sk)

	// 1000 first blocks not a validator
	ctx = ctx.WithBlockHeight(keeper.SignedBlocksWindow(ctx) + 1)

	// Validator created
	got := sh(ctx, NewTestMsgCreateValidator(addr, val, amt))
	require.True(t, got.IsOK())
	staking.EndBlocker(ctx, sk)

	require.Equal(
		t, ck.GetCoins(ctx, sdk.AccAddress(addr)),
		sdk.NewCoins(sdk.NewCoin(sk.GetParams(ctx).BondDenom, InitTokens.Sub(amt))),
	)
	require.Equal(t, amt, sk.Validator(ctx, addr).GetBondedTokens())

	// Now a validator, for two blocks
	keeper.HandleValidatorSignature(ctx, val.Address(), 100, true)
	ctx = ctx.WithBlockHeight(keeper.SignedBlocksWindow(ctx) + 2)
	keeper.HandleValidatorSignature(ctx, val.Address(), 100, false)

	info, found := keeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, keeper.SignedBlocksWindow(ctx)+1, info.StartHeight)
	require.Equal(t, int64(2), info.IndexOffset)
	require.Equal(t, int64(1), info.MissedBlocksCounter)
	require.Equal(t, time.Unix(0, 0).UTC(), info.JailedUntil)

	// validator should be bonded still, should not have been jailed or slashed
	validator, _ := sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Bonded, validator.GetStatus())
	bondPool := sk.GetBondedPool(ctx)
	expTokens := sdk.TokensFromConsensusPower(100)
	require.Equal(t, expTokens.Int64(), bondPool.GetCoins().AmountOf(sk.BondDenom(ctx)).Int64())
}

// Test a jailed validator being "down" twice
// Ensure that they're only slashed once
func TestHandleAlreadyJailed(t *testing.T) {

	// initial setup
	ctx, _, sk, _, keeper := CreateTestInput(t, types.DefaultParams())
	power := int64(100)
	amt := sdk.TokensFromConsensusPower(power)
	addr, val := Addrs[0], Pks[0]
	sh := staking.NewHandler(sk)
	got := sh(ctx, NewTestMsgCreateValidator(addr, val, amt))
	require.True(t, got.IsOK())
	staking.EndBlocker(ctx, sk)

	// 1000 first blocks OK
	height := int64(0)
	for ; height < keeper.SignedBlocksWindow(ctx); height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.HandleValidatorSignature(ctx, val.Address(), power, true)
	}

	// 501 blocks missed
	for ; height < keeper.SignedBlocksWindow(ctx)+(keeper.SignedBlocksWindow(ctx)-keeper.MinSignedPerWindow(ctx))+1; height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.HandleValidatorSignature(ctx, val.Address(), power, false)
	}

	// end block
	staking.EndBlocker(ctx, sk)

	// validator should have been jailed and slashed
	validator, _ := sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Unbonding, validator.GetStatus())

	// validator should have been slashed
	resultingTokens := amt.Sub(sdk.TokensFromConsensusPower(1))
	require.Equal(t, resultingTokens, validator.GetTokens())

	// another block missed
	ctx = ctx.WithBlockHeight(height)
	keeper.HandleValidatorSignature(ctx, val.Address(), power, false)

	// validator should not have been slashed twice
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, resultingTokens, validator.GetTokens())

}

// Test a validator dipping in and out of the validator set
// Ensure that missed blocks are tracked correctly and that
// the start height of the signing info is reset correctly
func TestValidatorDippingInAndOut(t *testing.T) {

	// initial setup
	// TestParams set the SignedBlocksWindow to 1000 and MaxMissedBlocksPerWindow to 500
	ctx, _, sk, _, keeper := CreateTestInput(t, TestParams())
	params := sk.GetParams(ctx)
	params.MaxValidators = 1
	sk.SetParams(ctx, params)
	power := int64(100)
	amt := sdk.TokensFromConsensusPower(power)
	addr, val := Addrs[0], Pks[0]
	consAddr := sdk.ConsAddress(addr)
	sh := staking.NewHandler(sk)
	got := sh(ctx, NewTestMsgCreateValidator(addr, val, amt))
	require.True(t, got.IsOK())
	staking.EndBlocker(ctx, sk)

	// 100 first blocks OK
	height := int64(0)
	for ; height < int64(100); height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.HandleValidatorSignature(ctx, val.Address(), power, true)
	}

	// kick first validator out of validator set
	newAmt := sdk.TokensFromConsensusPower(101)
	got = sh(ctx, NewTestMsgCreateValidator(Addrs[1], Pks[1], newAmt))
	require.True(t, got.IsOK())
	validatorUpdates := staking.EndBlocker(ctx, sk)
	require.Equal(t, 2, len(validatorUpdates))
	validator, _ := sk.GetValidator(ctx, addr)
	require.Equal(t, sdk.Unbonding, validator.Status)

	// 600 more blocks happened
	height = int64(700)
	ctx = ctx.WithBlockHeight(height)

	// validator added back in
	delTokens := sdk.TokensFromConsensusPower(50)
	got = sh(ctx, NewTestMsgDelegate(sdk.AccAddress(Addrs[2]), Addrs[0], delTokens))
	require.True(t, got.IsOK())
	validatorUpdates = staking.EndBlocker(ctx, sk)
	require.Equal(t, 2, len(validatorUpdates))
	validator, _ = sk.GetValidator(ctx, addr)
	require.Equal(t, sdk.Bonded, validator.Status)
	newPower := int64(150)

	// validator misses a block
	keeper.HandleValidatorSignature(ctx, val.Address(), newPower, false)
	height++

	// shouldn't be jailed/kicked yet
	validator, _ = sk.GetValidator(ctx, addr)
	require.Equal(t, sdk.Bonded, validator.Status)

	// validator misses 500 more blocks, 501 total
	latest := height
	for ; height < latest+500; height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.HandleValidatorSignature(ctx, val.Address(), newPower, false)
	}

	// should now be jailed & kicked
	staking.EndBlocker(ctx, sk)
	validator, _ = sk.GetValidator(ctx, addr)
	require.Equal(t, sdk.Unbonding, validator.Status)

	// check all the signing information
	signInfo, found := keeper.GetValidatorSigningInfo(ctx, consAddr)
	require.True(t, found)
	require.Equal(t, int64(0), signInfo.MissedBlocksCounter)
	require.Equal(t, int64(0), signInfo.IndexOffset)
	// array should be cleared
	for offset := int64(0); offset < keeper.SignedBlocksWindow(ctx); offset++ {
		missed := keeper.GetValidatorMissedBlockBitArray(ctx, consAddr, offset)
		require.False(t, missed)
	}

	// some blocks pass
	height = int64(5000)
	ctx = ctx.WithBlockHeight(height)

	// validator rejoins and starts signing again
	sk.Unjail(ctx, consAddr)
	keeper.HandleValidatorSignature(ctx, val.Address(), newPower, true)
	height++

	// validator should not be kicked since we reset counter/array when it was jailed
	staking.EndBlocker(ctx, sk)
	validator, _ = sk.GetValidator(ctx, addr)
	require.Equal(t, sdk.Bonded, validator.Status)

	// validator misses 501 blocks
	latest = height
	for ; height < latest+501; height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.HandleValidatorSignature(ctx, val.Address(), newPower, false)
	}

	// validator should now be jailed & kicked
	staking.EndBlocker(ctx, sk)
	validator, _ = sk.GetValidator(ctx, addr)
	require.Equal(t, sdk.Unbonding, validator.Status)

}
