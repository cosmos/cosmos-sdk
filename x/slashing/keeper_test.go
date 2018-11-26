package slashing

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

// Have to change these parameters for tests
// lest the tests take forever
func keeperTestParams() Params {
	params := DefaultParams()
	params.SignedBlocksWindow = 1000
	params.DowntimeUnbondDuration = 60 * 60
	params.DoubleSignUnbondDuration = 60 * 60
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
	amtInt := int64(100)
	operatorAddr, val, amt := addrs[0], pks[0], sdk.NewInt(amtInt)
	got := stake.NewHandler(sk)(ctx, NewTestMsgCreateValidator(operatorAddr, val, amt))
	require.True(t, got.IsOK())
	stake.EndBlocker(ctx, sk)
	require.Equal(
		t, ck.GetCoins(ctx, sdk.AccAddress(operatorAddr)),
		sdk.Coins{sdk.NewCoin(sk.GetParams(ctx).BondDenom, initCoins.Sub(amt))},
	)
	require.True(t, sdk.NewDecFromInt(amt).Equal(sk.Validator(ctx, operatorAddr).GetPower()))

	// handle a signature to set signing info
	keeper.handleValidatorSignature(ctx, val.Address(), amtInt, true)

	// double sign less than max age
	keeper.handleDoubleSign(ctx, val.Address(), 0, time.Unix(0, 0), amtInt)

	// should be jailed
	require.True(t, sk.Validator(ctx, operatorAddr).GetJailed())
	// unjail to measure power
	sk.Unjail(ctx, sdk.ConsAddress(val.Address()))
	// power should be reduced
	require.Equal(
		t, sdk.NewDecFromInt(amt).Mul(sdk.NewDec(19).Quo(sdk.NewDec(20))),
		sk.Validator(ctx, operatorAddr).GetPower(),
	)
	ctx = ctx.WithBlockHeader(abci.Header{Time: time.Unix(1, 0).Add(keeper.MaxEvidenceAge(ctx))})

	// double sign past max age
	keeper.handleDoubleSign(ctx, val.Address(), 0, time.Unix(0, 0), amtInt)
	require.Equal(
		t, sdk.NewDecFromInt(amt).Mul(sdk.NewDec(19).Quo(sdk.NewDec(20))),
		sk.Validator(ctx, operatorAddr).GetPower(),
	)
}

// Test that the amount a validator is slashed for multiple double signs
// is correctly capped by the slashing period in which they were committed
func TestSlashingPeriodCap(t *testing.T) {

	// initial setup
	ctx, ck, sk, _, keeper := createTestInput(t, DefaultParams())
	amtInt := int64(100)
	operatorAddr, amt := addrs[0], sdk.NewInt(amtInt)
	valConsPubKey, valConsAddr := pks[0], pks[0].Address()
	got := stake.NewHandler(sk)(ctx, NewTestMsgCreateValidator(operatorAddr, valConsPubKey, amt))
	require.True(t, got.IsOK())
	stake.EndBlocker(ctx, sk)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	require.Equal(
		t, ck.GetCoins(ctx, sdk.AccAddress(operatorAddr)),
		sdk.Coins{sdk.NewCoin(sk.GetParams(ctx).BondDenom, initCoins.Sub(amt))},
	)
	require.True(t, sdk.NewDecFromInt(amt).Equal(sk.Validator(ctx, operatorAddr).GetPower()))

	// handle a signature to set signing info
	keeper.handleValidatorSignature(ctx, valConsAddr, amtInt, true)

	// double sign less than max age
	keeper.handleDoubleSign(ctx, valConsAddr, 1, time.Unix(0, 0), amtInt)
	// should be jailed
	require.True(t, sk.Validator(ctx, operatorAddr).GetJailed())
	// end block
	stake.EndBlocker(ctx, sk)
	// update block height
	ctx = ctx.WithBlockHeight(int64(2))
	// unjail to measure power
	sk.Unjail(ctx, sdk.ConsAddress(valConsAddr))
	// end block
	stake.EndBlocker(ctx, sk)
	// power should be reduced
	expectedPower := sdk.NewDecFromInt(amt).Mul(sdk.NewDec(19).Quo(sdk.NewDec(20)))
	require.Equal(t, expectedPower, sk.Validator(ctx, operatorAddr).GetPower())

	// double sign again, same slashing period
	keeper.handleDoubleSign(ctx, valConsAddr, 1, time.Unix(0, 0), amtInt)
	// should be jailed
	require.True(t, sk.Validator(ctx, operatorAddr).GetJailed())
	// end block
	stake.EndBlocker(ctx, sk)
	// update block height
	ctx = ctx.WithBlockHeight(int64(3))
	// unjail to measure power
	sk.Unjail(ctx, sdk.ConsAddress(valConsAddr))
	// end block
	stake.EndBlocker(ctx, sk)
	// power should be equal, no more should have been slashed
	expectedPower = sdk.NewDecFromInt(amt).Mul(sdk.NewDec(19).Quo(sdk.NewDec(20)))
	require.Equal(t, expectedPower, sk.Validator(ctx, operatorAddr).GetPower())

	// double sign again, new slashing period
	keeper.handleDoubleSign(ctx, valConsAddr, 3, time.Unix(0, 0), amtInt)
	// should be jailed
	require.True(t, sk.Validator(ctx, operatorAddr).GetJailed())
	// unjail to measure power
	sk.Unjail(ctx, sdk.ConsAddress(valConsAddr))
	// end block
	stake.EndBlocker(ctx, sk)
	// power should be reduced
	expectedPower = sdk.NewDecFromInt(amt).Mul(sdk.NewDec(18).Quo(sdk.NewDec(20)))
	require.Equal(t, expectedPower, sk.Validator(ctx, operatorAddr).GetPower())
}

// Test a validator through uptime, downtime, revocation,
// unrevocation, starting height reset, and revocation again
func TestHandleAbsentValidator(t *testing.T) {

	// initial setup
	ctx, ck, sk, _, keeper := createTestInput(t, keeperTestParams())
	amtInt := int64(100)
	addr, val, amt := addrs[0], pks[0], sdk.NewInt(amtInt)
	sh := stake.NewHandler(sk)
	slh := NewHandler(keeper)
	got := sh(ctx, NewTestMsgCreateValidator(addr, val, amt))
	require.True(t, got.IsOK())
	stake.EndBlocker(ctx, sk)

	require.Equal(
		t, ck.GetCoins(ctx, sdk.AccAddress(addr)),
		sdk.Coins{sdk.NewCoin(sk.GetParams(ctx).BondDenom, initCoins.Sub(amt))},
	)
	require.True(t, sdk.NewDecFromInt(amt).Equal(sk.Validator(ctx, addr).GetPower()))

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
		keeper.handleValidatorSignature(ctx, val.Address(), amtInt, true)
	}
	info, found = keeper.getValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, int64(0), info.MissedBlocksCounter)

	// 500 blocks missed
	for ; height < keeper.SignedBlocksWindow(ctx)+(keeper.SignedBlocksWindow(ctx)-keeper.MinSignedPerWindow(ctx)); height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.handleValidatorSignature(ctx, val.Address(), amtInt, false)
	}
	info, found = keeper.getValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, keeper.SignedBlocksWindow(ctx)-keeper.MinSignedPerWindow(ctx), info.MissedBlocksCounter)

	// validator should be bonded still
	validator, _ := sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Bonded, validator.GetStatus())
	pool := sk.GetPool(ctx)
	require.Equal(t, amtInt, pool.BondedTokens.RoundInt64())

	// 501st block missed
	ctx = ctx.WithBlockHeight(height)
	keeper.handleValidatorSignature(ctx, val.Address(), amtInt, false)
	info, found = keeper.getValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	// counter now reset to zero
	require.Equal(t, int64(0), info.MissedBlocksCounter)

	// end block
	stake.EndBlocker(ctx, sk)

	// validator should have been jailed
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Unbonding, validator.GetStatus())

	slashAmt := sdk.NewDec(amtInt).Mul(keeper.SlashFractionDowntime(ctx)).RoundInt64()

	// validator should have been slashed
	require.Equal(t, amtInt-slashAmt, validator.GetTokens().RoundInt64())

	// 502nd block *also* missed (since the LastCommit would have still included the just-unbonded validator)
	height++
	ctx = ctx.WithBlockHeight(height)
	keeper.handleValidatorSignature(ctx, val.Address(), amtInt, false)
	info, found = keeper.getValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, int64(1), info.MissedBlocksCounter)

	// end block
	stake.EndBlocker(ctx, sk)

	// validator should not have been slashed any more, since it was already jailed
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, amtInt-slashAmt, validator.GetTokens().RoundInt64())

	// 502nd block *double signed* (oh no!)
	keeper.handleDoubleSign(ctx, val.Address(), height, ctx.BlockHeader().Time, amtInt)

	// validator should have been slashed
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	secondSlashAmt := sdk.NewDec(amtInt).Mul(keeper.SlashFractionDoubleSign(ctx)).RoundInt64()
	require.Equal(t, amtInt-slashAmt-secondSlashAmt, validator.GetTokens().RoundInt64())

	// unrevocation should fail prior to jail expiration
	got = slh(ctx, NewMsgUnjail(addr))
	require.False(t, got.IsOK())

	// unrevocation should succeed after jail expiration
	ctx = ctx.WithBlockHeader(abci.Header{Time: time.Unix(1, 0).Add(keeper.DowntimeUnbondDuration(ctx))})
	got = slh(ctx, NewMsgUnjail(addr))
	require.True(t, got.IsOK())

	// end block
	stake.EndBlocker(ctx, sk)

	// validator should be rebonded now
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Bonded, validator.GetStatus())

	// validator should have been slashed
	pool = sk.GetPool(ctx)
	require.Equal(t, amtInt-slashAmt-secondSlashAmt, pool.BondedTokens.RoundInt64())

	// validator start height should not have been changed
	info, found = keeper.getValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	// we've missed 2 blocks more than the maximum, so the counter was reset to 0 at 1 block more and is now 1
	require.Equal(t, int64(1), info.MissedBlocksCounter)

	// validator should not be immediately jailed again
	height++
	ctx = ctx.WithBlockHeight(height)
	keeper.handleValidatorSignature(ctx, val.Address(), amtInt, false)
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Bonded, validator.GetStatus())

	// 500 signed blocks
	nextHeight := height + keeper.MinSignedPerWindow(ctx) + 1
	for ; height < nextHeight; height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.handleValidatorSignature(ctx, val.Address(), amtInt, false)
	}

	// end block
	stake.EndBlocker(ctx, sk)

	// validator should be jailed again after 500 unsigned blocks
	nextHeight = height + keeper.MinSignedPerWindow(ctx) + 1
	for ; height <= nextHeight; height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.handleValidatorSignature(ctx, val.Address(), amtInt, false)
	}

	// end block
	stake.EndBlocker(ctx, sk)

	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Unbonding, validator.GetStatus())
}

// Test a new validator entering the validator set
// Ensure that SigningInfo.StartHeight is set correctly
// and that they are not immediately jailed
func TestHandleNewValidator(t *testing.T) {
	// initial setup
	ctx, ck, sk, _, keeper := createTestInput(t, keeperTestParams())
	addr, val, amt := addrs[0], pks[0], int64(100)
	sh := stake.NewHandler(sk)

	// 1000 first blocks not a validator
	ctx = ctx.WithBlockHeight(keeper.SignedBlocksWindow(ctx) + 1)

	// Validator created
	got := sh(ctx, NewTestMsgCreateValidator(addr, val, sdk.NewInt(amt)))
	require.True(t, got.IsOK())
	stake.EndBlocker(ctx, sk)

	require.Equal(
		t, ck.GetCoins(ctx, sdk.AccAddress(addr)),
		sdk.Coins{sdk.NewCoin(sk.GetParams(ctx).BondDenom, initCoins.SubRaw(amt))},
	)
	require.Equal(t, sdk.NewDec(amt), sk.Validator(ctx, addr).GetPower())

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
	require.Equal(t, int64(100), pool.BondedTokens.RoundInt64())
}

// Test a jailed validator being "down" twice
// Ensure that they're only slashed once
func TestHandleAlreadyJailed(t *testing.T) {

	// initial setup
	ctx, _, sk, _, keeper := createTestInput(t, DefaultParams())
	amtInt := int64(100)
	addr, val, amt := addrs[0], pks[0], sdk.NewInt(amtInt)
	sh := stake.NewHandler(sk)
	got := sh(ctx, NewTestMsgCreateValidator(addr, val, amt))
	require.True(t, got.IsOK())
	stake.EndBlocker(ctx, sk)

	// 1000 first blocks OK
	height := int64(0)
	for ; height < keeper.SignedBlocksWindow(ctx); height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.handleValidatorSignature(ctx, val.Address(), amtInt, true)
	}

	// 501 blocks missed
	for ; height < keeper.SignedBlocksWindow(ctx)+(keeper.SignedBlocksWindow(ctx)-keeper.MinSignedPerWindow(ctx))+1; height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.handleValidatorSignature(ctx, val.Address(), amtInt, false)
	}

	// end block
	stake.EndBlocker(ctx, sk)

	// validator should have been jailed and slashed
	validator, _ := sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Unbonding, validator.GetStatus())

	// validator should have been slashed
	require.Equal(t, amtInt-1, validator.GetTokens().RoundInt64())

	// another block missed
	ctx = ctx.WithBlockHeight(height)
	keeper.handleValidatorSignature(ctx, val.Address(), amtInt, false)

	// validator should not have been slashed twice
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, amtInt-1, validator.GetTokens().RoundInt64())

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
	amtInt := int64(100)
	addr, val, amt := addrs[0], pks[0], sdk.NewInt(amtInt)
	consAddr := sdk.ConsAddress(addr)
	sh := stake.NewHandler(sk)
	got := sh(ctx, NewTestMsgCreateValidator(addr, val, amt))
	require.True(t, got.IsOK())
	stake.EndBlocker(ctx, sk)

	// 100 first blocks OK
	height := int64(0)
	for ; height < int64(100); height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.handleValidatorSignature(ctx, val.Address(), amtInt, true)
	}

	// validator kicked out of validator set
	newAmt := int64(101)
	got = sh(ctx, NewTestMsgCreateValidator(addrs[1], pks[1], sdk.NewInt(newAmt)))
	require.True(t, got.IsOK())
	validatorUpdates := stake.EndBlocker(ctx, sk)
	require.Equal(t, 2, len(validatorUpdates))
	validator, _ := sk.GetValidator(ctx, addr)
	require.Equal(t, sdk.Unbonding, validator.Status)

	// 600 more blocks happened
	height = int64(700)
	ctx = ctx.WithBlockHeight(height)

	// validator added back in
	got = sh(ctx, newTestMsgDelegate(sdk.AccAddress(addrs[2]), addrs[0], sdk.NewInt(2)))
	require.True(t, got.IsOK())
	validatorUpdates = stake.EndBlocker(ctx, sk)
	require.Equal(t, 2, len(validatorUpdates))
	validator, _ = sk.GetValidator(ctx, addr)
	require.Equal(t, sdk.Bonded, validator.Status)
	newAmt = int64(102)

	// validator misses a block
	keeper.handleValidatorSignature(ctx, val.Address(), newAmt, false)
	height++

	// shouldn't be jailed/kicked yet
	validator, _ = sk.GetValidator(ctx, addr)
	require.Equal(t, sdk.Bonded, validator.Status)

	// validator misses 500 more blocks, 501 total
	latest := height
	for ; height < latest+500; height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.handleValidatorSignature(ctx, val.Address(), newAmt, false)
	}

	// should now be jailed & kicked
	stake.EndBlocker(ctx, sk)
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
	keeper.handleValidatorSignature(ctx, val.Address(), newAmt, true)
	height++

	// validator should not be kicked since we reset counter/array when it was jailed
	stake.EndBlocker(ctx, sk)
	validator, _ = sk.GetValidator(ctx, addr)
	require.Equal(t, sdk.Bonded, validator.Status)

	// validator misses 501 blocks
	latest = height
	for ; height < latest+501; height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.handleValidatorSignature(ctx, val.Address(), newAmt, false)
	}

	// validator should now be jailed & kicked
	stake.EndBlocker(ctx, sk)
	validator, _ = sk.GetValidator(ctx, addr)
	require.Equal(t, sdk.Unbonding, validator.Status)

}
