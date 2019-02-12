package slashing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// Have to change these parameters for tests
// lest the tests take forever
func keeperTestParams() Params {
	params := DefaultParams()
	params.SignedBlocksWindow = 1000
	params.DowntimeJailDuration = 60 * 60
	return params
}

// ______________________________________________________________

// Test that a validator is slashed correctly
// when we discover evidence of infraction
func TestHandleDoubleSign(t *testing.T) {

	// initial setup
	ctx, ck, sk, _, keeper := createTestInput(t, keeperTestParams())
	// validator added pre-genesis
	ctx = ctx.WithBlockHeight(-1)
	power := int64(100)
	amt := sdk.TokensFromTendermintPower(power)
	operatorAddr, val := addrs[0], pks[0]
	got := staking.NewHandler(sk)(ctx, NewTestMsgCreateValidator(operatorAddr, val, amt))
	require.True(t, got.IsOK())
	staking.EndBlocker(ctx, sk)
	require.Equal(
		t, ck.GetCoins(ctx, sdk.AccAddress(operatorAddr)),
		sdk.Coins{sdk.NewCoin(sk.GetParams(ctx).BondDenom, initCoins.Sub(amt))},
	)
	require.Equal(t, amt, sk.Validator(ctx, operatorAddr).GetBondedTokens())

	// handle a signature to set signing info
	keeper.handleValidatorSignature(ctx, val.Address(), amt.Int64(), true)

	oldTokens := sk.Validator(ctx, operatorAddr).GetTokens()

	// double sign less than max age
	keeper.handleDoubleSign(ctx, val.Address(), 0, time.Unix(0, 0), power)

	// should be jailed
	require.True(t, sk.Validator(ctx, operatorAddr).GetJailed())

	// tokens should be decreased
	newTokens := sk.Validator(ctx, operatorAddr).GetTokens()
	require.True(t, newTokens.LT(oldTokens))

	// New evidence
	keeper.handleDoubleSign(ctx, val.Address(), 0, time.Unix(0, 0), power)

	// tokens should be the same (capped slash)
	require.True(t, sk.Validator(ctx, operatorAddr).GetTokens().Equal(newTokens))

	// Jump to past the unbonding period
	ctx = ctx.WithBlockHeader(abci.Header{Time: time.Unix(1, 0).Add(sk.GetParams(ctx).UnbondingTime)})

	// Still shouldn't be able to unjail
	msgUnjail := NewMsgUnjail(operatorAddr)
	res := handleMsgUnjail(ctx, msgUnjail, keeper)
	require.False(t, res.IsOK())

	// Should be able to unbond now
	del, _ := sk.GetDelegation(ctx, sdk.AccAddress(operatorAddr), operatorAddr)
	msgUnbond := staking.NewMsgUndelegate(sdk.AccAddress(operatorAddr), operatorAddr, del.GetShares())
	res = staking.NewHandler(sk)(ctx, msgUnbond)
	require.True(t, res.IsOK())
}

// ______________________________________________________________

// Test that a validator is slashed correctly
// when we discover evidence of infraction
func TestPastMaxEvidenceAge(t *testing.T) {

	// initial setup
	ctx, ck, sk, _, keeper := createTestInput(t, keeperTestParams())
	// validator added pre-genesis
	ctx = ctx.WithBlockHeight(-1)
	power := int64(100)
	amt := sdk.TokensFromTendermintPower(power)
	operatorAddr, val := addrs[0], pks[0]
	got := staking.NewHandler(sk)(ctx, NewTestMsgCreateValidator(operatorAddr, val, amt))
	require.True(t, got.IsOK())
	staking.EndBlocker(ctx, sk)
	require.Equal(
		t, ck.GetCoins(ctx, sdk.AccAddress(operatorAddr)),
		sdk.Coins{sdk.NewCoin(sk.GetParams(ctx).BondDenom, initCoins.Sub(amt))},
	)
	require.Equal(t, amt, sk.Validator(ctx, operatorAddr).GetBondedTokens())

	// handle a signature to set signing info
	keeper.handleValidatorSignature(ctx, val.Address(), power, true)

	ctx = ctx.WithBlockHeader(abci.Header{Time: time.Unix(1, 0).Add(keeper.MaxEvidenceAge(ctx))})

	oldPower := sk.Validator(ctx, operatorAddr).GetTendermintPower()

	// double sign past max age
	keeper.handleDoubleSign(ctx, val.Address(), 0, time.Unix(0, 0), power)

	// should still be bonded
	require.True(t, sk.Validator(ctx, operatorAddr).GetStatus() == sdk.Bonded)

	// should still have same power
	require.Equal(t, oldPower, sk.Validator(ctx, operatorAddr).GetTendermintPower())
}

// Test a validator through uptime, downtime, revocation,
// unrevocation, starting height reset, and revocation again
func TestHandleAbsentValidator(t *testing.T) {

	// initial setup
	ctx, ck, sk, _, keeper := createTestInput(t, keeperTestParams())
	power := int64(100)
	amt := sdk.TokensFromTendermintPower(power)
	addr, val := addrs[0], pks[0]
	sh := staking.NewHandler(sk)
	slh := NewHandler(keeper)
	got := sh(ctx, NewTestMsgCreateValidator(addr, val, amt))
	require.True(t, got.IsOK())
	staking.EndBlocker(ctx, sk)

	require.Equal(
		t, ck.GetCoins(ctx, sdk.AccAddress(addr)),
		sdk.Coins{sdk.NewCoin(sk.GetParams(ctx).BondDenom, initCoins.Sub(amt))},
	)
	require.Equal(t, amt, sk.Validator(ctx, addr).GetBondedTokens())

	// will exist since the validator has been bonded
	info, found := keeper.getValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, int64(0), info.IndexOffset)
	require.Equal(t, int64(0), info.MissedBlocksCounter)
	require.Equal(t, time.Unix(0, 0).UTC(), info.JailedUntil)
	height := int64(0)

	// 1000 first blocks OK
	for ; height < keeper.SignedBlocksWindow(ctx); height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.handleValidatorSignature(ctx, val.Address(), power, true)
	}
	info, found = keeper.getValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, int64(0), info.MissedBlocksCounter)

	// 500 blocks missed
	for ; height < keeper.SignedBlocksWindow(ctx)+(keeper.SignedBlocksWindow(ctx)-keeper.MinSignedPerWindow(ctx)); height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.handleValidatorSignature(ctx, val.Address(), power, false)
	}
	info, found = keeper.getValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, keeper.SignedBlocksWindow(ctx)-keeper.MinSignedPerWindow(ctx), info.MissedBlocksCounter)

	// validator should be bonded still
	validator, _ := sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Bonded, validator.GetStatus())
	pool := sk.GetPool(ctx)
	require.True(sdk.IntEq(t, amt, pool.BondedTokens))

	// 501st block missed
	ctx = ctx.WithBlockHeight(height)
	keeper.handleValidatorSignature(ctx, val.Address(), power, false)
	info, found = keeper.getValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	// counter now reset to zero
	require.Equal(t, int64(0), info.MissedBlocksCounter)

	// end block
	staking.EndBlocker(ctx, sk)

	// validator should have been jailed
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Unbonding, validator.GetStatus())

	slashAmt := sdk.NewDecFromInt(amt).Mul(keeper.SlashFractionDowntime(ctx)).RoundInt64()

	// validator should have been slashed
	require.Equal(t, amt.Int64()-slashAmt, validator.GetTokens().Int64())

	// 502nd block *also* missed (since the LastCommit would have still included the just-unbonded validator)
	height++
	ctx = ctx.WithBlockHeight(height)
	keeper.handleValidatorSignature(ctx, val.Address(), power, false)
	info, found = keeper.getValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, int64(1), info.MissedBlocksCounter)

	// end block
	staking.EndBlocker(ctx, sk)

	// validator should not have been slashed any more, since it was already jailed
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, amt.Int64()-slashAmt, validator.GetTokens().Int64())

	// unrevocation should fail prior to jail expiration
	got = slh(ctx, NewMsgUnjail(addr))
	require.False(t, got.IsOK())

	// unrevocation should succeed after jail expiration
	ctx = ctx.WithBlockHeader(abci.Header{Time: time.Unix(1, 0).Add(keeper.DowntimeJailDuration(ctx))})
	got = slh(ctx, NewMsgUnjail(addr))
	require.True(t, got.IsOK())

	// end block
	staking.EndBlocker(ctx, sk)

	// validator should be rebonded now
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Bonded, validator.GetStatus())

	// validator should have been slashed
	pool = sk.GetPool(ctx)
	require.Equal(t, amt.Int64()-slashAmt, pool.BondedTokens.Int64())

	// Validator start height should not have been changed
	info, found = keeper.getValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	// we've missed 2 blocks more than the maximum, so the counter was reset to 0 at 1 block more and is now 1
	require.Equal(t, int64(1), info.MissedBlocksCounter)

	// validator should not be immediately jailed again
	height++
	ctx = ctx.WithBlockHeight(height)
	keeper.handleValidatorSignature(ctx, val.Address(), power, false)
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Bonded, validator.GetStatus())

	// 500 signed blocks
	nextHeight := height + keeper.MinSignedPerWindow(ctx) + 1
	for ; height < nextHeight; height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.handleValidatorSignature(ctx, val.Address(), power, false)
	}

	// end block
	staking.EndBlocker(ctx, sk)

	// validator should be jailed again after 500 unsigned blocks
	nextHeight = height + keeper.MinSignedPerWindow(ctx) + 1
	for ; height <= nextHeight; height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.handleValidatorSignature(ctx, val.Address(), power, false)
	}

	// end block
	staking.EndBlocker(ctx, sk)

	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Unbonding, validator.GetStatus())
}

// Test a new validator entering the validator set
// Ensure that SigningInfo.StartHeight is set correctly
// and that they are not immediately jailed
func TestHandleNewValidator(t *testing.T) {
	// initial setup
	ctx, ck, sk, _, keeper := createTestInput(t, keeperTestParams())
	addr, val := addrs[0], pks[0]
	amt := sdk.TokensFromTendermintPower(100)
	sh := staking.NewHandler(sk)

	// 1000 first blocks not a validator
	ctx = ctx.WithBlockHeight(keeper.SignedBlocksWindow(ctx) + 1)

	// Validator created
	got := sh(ctx, NewTestMsgCreateValidator(addr, val, amt))
	require.True(t, got.IsOK())
	staking.EndBlocker(ctx, sk)

	require.Equal(
		t, ck.GetCoins(ctx, sdk.AccAddress(addr)),
		sdk.Coins{sdk.NewCoin(sk.GetParams(ctx).BondDenom, initCoins.Sub(amt))},
	)
	require.Equal(t, amt, sk.Validator(ctx, addr).GetBondedTokens())

	// Now a validator, for two blocks
	keeper.handleValidatorSignature(ctx, val.Address(), 100, true)
	ctx = ctx.WithBlockHeight(keeper.SignedBlocksWindow(ctx) + 2)
	keeper.handleValidatorSignature(ctx, val.Address(), 100, false)

	info, found := keeper.getValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, keeper.SignedBlocksWindow(ctx)+1, info.StartHeight)
	require.Equal(t, int64(2), info.IndexOffset)
	require.Equal(t, int64(1), info.MissedBlocksCounter)
	require.Equal(t, time.Unix(0, 0).UTC(), info.JailedUntil)

	// validator should be bonded still, should not have been jailed or slashed
	validator, _ := sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Bonded, validator.GetStatus())
	pool := sk.GetPool(ctx)
	expTokens := sdk.TokensFromTendermintPower(100)
	require.Equal(t, expTokens, pool.BondedTokens)
}

// Test a jailed validator being "down" twice
// Ensure that they're only slashed once
func TestHandleAlreadyJailed(t *testing.T) {

	// initial setup
	ctx, _, sk, _, keeper := createTestInput(t, DefaultParams())
	power := int64(100)
	amt := sdk.TokensFromTendermintPower(power)
	addr, val := addrs[0], pks[0]
	sh := staking.NewHandler(sk)
	got := sh(ctx, NewTestMsgCreateValidator(addr, val, amt))
	require.True(t, got.IsOK())
	staking.EndBlocker(ctx, sk)

	// 1000 first blocks OK
	height := int64(0)
	for ; height < keeper.SignedBlocksWindow(ctx); height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.handleValidatorSignature(ctx, val.Address(), power, true)
	}

	// 501 blocks missed
	for ; height < keeper.SignedBlocksWindow(ctx)+(keeper.SignedBlocksWindow(ctx)-keeper.MinSignedPerWindow(ctx))+1; height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.handleValidatorSignature(ctx, val.Address(), power, false)
	}

	// end block
	staking.EndBlocker(ctx, sk)

	// validator should have been jailed and slashed
	validator, _ := sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Unbonding, validator.GetStatus())

	// validator should have been slashed
	resultingTokens := amt.Sub(sdk.TokensFromTendermintPower(1))
	require.Equal(t, resultingTokens, validator.GetTokens())

	// another block missed
	ctx = ctx.WithBlockHeight(height)
	keeper.handleValidatorSignature(ctx, val.Address(), power, false)

	// validator should not have been slashed twice
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, resultingTokens, validator.GetTokens())

}

// Test a validator dipping in and out of the validator set
// Ensure that missed blocks are tracked correctly and that
// the start height of the signing info is reset correctly
func TestValidatorDippingInAndOut(t *testing.T) {

	// initial setup
	// keeperTestParams set the SignedBlocksWindow to 1000 and MaxMissedBlocksPerWindow to 500
	ctx, _, sk, _, keeper := createTestInput(t, keeperTestParams())
	params := sk.GetParams(ctx)
	params.MaxValidators = 1
	sk.SetParams(ctx, params)
	power := int64(100)
	amt := sdk.TokensFromTendermintPower(power)
	addr, val := addrs[0], pks[0]
	consAddr := sdk.ConsAddress(addr)
	sh := staking.NewHandler(sk)
	got := sh(ctx, NewTestMsgCreateValidator(addr, val, amt))
	require.True(t, got.IsOK())
	staking.EndBlocker(ctx, sk)

	// 100 first blocks OK
	height := int64(0)
	for ; height < int64(100); height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.handleValidatorSignature(ctx, val.Address(), power, true)
	}

	// validator kicked out of validator set
	newAmt := sdk.TokensFromTendermintPower(101)
	got = sh(ctx, NewTestMsgCreateValidator(addrs[1], pks[1], newAmt))
	require.True(t, got.IsOK())
	validatorUpdates, _ := staking.EndBlocker(ctx, sk)
	require.Equal(t, 2, len(validatorUpdates))
	validator, _ := sk.GetValidator(ctx, addr)
	require.Equal(t, sdk.Unbonding, validator.Status)

	// 600 more blocks happened
	height = int64(700)
	ctx = ctx.WithBlockHeight(height)

	// validator added back in
	delTokens := sdk.TokensFromTendermintPower(3)
	got = sh(ctx, newTestMsgDelegate(sdk.AccAddress(addrs[2]), addrs[0], delTokens))
	require.True(t, got.IsOK())
	validatorUpdates, _ = staking.EndBlocker(ctx, sk)
	require.Equal(t, 2, len(validatorUpdates))
	validator, _ = sk.GetValidator(ctx, addr)
	require.Equal(t, sdk.Bonded, validator.Status)
	newPower := int64(103)

	// validator misses a block
	keeper.handleValidatorSignature(ctx, val.Address(), newPower, false)
	height++

	// shouldn't be jailed/kicked yet
	validator, _ = sk.GetValidator(ctx, addr)
	require.Equal(t, sdk.Bonded, validator.Status)

	// validator misses 500 more blocks, 501 total
	latest := height
	for ; height < latest+500; height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.handleValidatorSignature(ctx, val.Address(), newPower, false)
	}

	// should now be jailed & kicked
	staking.EndBlocker(ctx, sk)
	validator, _ = sk.GetValidator(ctx, addr)
	require.Equal(t, sdk.Unbonding, validator.Status)

	// check all the signing information
	signInfo, found := keeper.getValidatorSigningInfo(ctx, consAddr)
	require.True(t, found)
	require.Equal(t, int64(0), signInfo.MissedBlocksCounter)
	require.Equal(t, int64(0), signInfo.IndexOffset)
	// array should be cleared
	for offset := int64(0); offset < keeper.SignedBlocksWindow(ctx); offset++ {
		missed := keeper.getValidatorMissedBlockBitArray(ctx, consAddr, offset)
		require.False(t, missed)
	}

	// some blocks pass
	height = int64(5000)
	ctx = ctx.WithBlockHeight(height)

	// validator rejoins and starts signing again
	sk.Unjail(ctx, consAddr)
	keeper.handleValidatorSignature(ctx, val.Address(), newPower, true)
	height++

	// validator should not be kicked since we reset counter/array when it was jailed
	staking.EndBlocker(ctx, sk)
	validator, _ = sk.GetValidator(ctx, addr)
	require.Equal(t, sdk.Bonded, validator.Status)

	// validator misses 501 blocks
	latest = height
	for ; height < latest+501; height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.handleValidatorSignature(ctx, val.Address(), newPower, false)
	}

	// validator should now be jailed & kicked
	staking.EndBlocker(ctx, sk)
	validator, _ = sk.GetValidator(ctx, addr)
	require.Equal(t, sdk.Unbonding, validator.Status)

}
