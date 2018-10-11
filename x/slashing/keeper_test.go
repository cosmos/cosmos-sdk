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
func init() {
	defaultSignedBlocksWindow = 1000
	defaultDowntimeUnbondDuration = 60 * 60
	defaultDoubleSignUnbondDuration = 60 * 60
}

// ______________________________________________________________

// Test that a validator is slashed correctly
// when we discover evidence of infraction
func TestHandleDoubleSign(t *testing.T) {

	// initial setup
	ctx, ck, sk, _, keeper := createTestInput(t)
	// validator added pre-genesis
	ctx = ctx.WithBlockHeight(-1)
	amtInt := int64(100)
	operatorAddr, val, amt := addrs[0], pks[0], sdk.NewInt(amtInt)
	got := stake.NewHandler(sk)(ctx, newTestMsgCreateValidator(operatorAddr, val, amt))
	require.True(t, got.IsOK())
	validatorUpdates := stake.EndBlocker(ctx, sk)
	keeper.AddValidators(ctx, validatorUpdates)
	require.Equal(t, ck.GetCoins(ctx, sdk.AccAddress(operatorAddr)), sdk.Coins{{sk.GetParams(ctx).BondDenom, initCoins.Sub(amt)}})
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
	ctx, ck, sk, _, keeper := createTestInput(t)
	amtInt := int64(100)
	operatorAddr, amt := addrs[0], sdk.NewInt(amtInt)
	valConsPubKey, valConsAddr := pks[0], pks[0].Address()
	got := stake.NewHandler(sk)(ctx, newTestMsgCreateValidator(operatorAddr, valConsPubKey, amt))
	require.True(t, got.IsOK())
	validatorUpdates := stake.EndBlocker(ctx, sk)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	keeper.AddValidators(ctx, validatorUpdates)
	require.Equal(t, ck.GetCoins(ctx, sdk.AccAddress(operatorAddr)), sdk.Coins{{sk.GetParams(ctx).BondDenom, initCoins.Sub(amt)}})
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
	ctx, ck, sk, _, keeper := createTestInput(t)
	amtInt := int64(100)
	addr, val, amt := addrs[0], pks[0], sdk.NewInt(amtInt)
	sh := stake.NewHandler(sk)
	slh := NewHandler(keeper)
	got := sh(ctx, newTestMsgCreateValidator(addr, val, amt))
	require.True(t, got.IsOK())
	validatorUpdates := stake.EndBlocker(ctx, sk)
	keeper.AddValidators(ctx, validatorUpdates)
	require.Equal(t, ck.GetCoins(ctx, sdk.AccAddress(addr)), sdk.Coins{{sk.GetParams(ctx).BondDenom, initCoins.Sub(amt)}})
	require.True(t, sdk.NewDecFromInt(amt).Equal(sk.Validator(ctx, addr).GetPower()))
	// will exist since the validator has been bonded
	info, found := keeper.getValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, int64(0), info.IndexOffset)
	require.Equal(t, int64(0), info.MissedBlocksCounter)
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
	require.Equal(t, keeper.SignedBlocksWindow(ctx)-keeper.MinSignedPerWindow(ctx)+1, info.MissedBlocksCounter)

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
	require.Equal(t, keeper.SignedBlocksWindow(ctx)-keeper.MinSignedPerWindow(ctx)+2, info.MissedBlocksCounter)

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

	// validator start height should have been changed
	info, found = keeper.getValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, height, info.StartHeight)
	// we've missed 2 blocks more than the maximum
	require.Equal(t, keeper.SignedBlocksWindow(ctx)-keeper.MinSignedPerWindow(ctx)+2, info.MissedBlocksCounter)

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
	ctx, ck, sk, _, keeper := createTestInput(t)
	addr, val, amt := addrs[0], pks[0], int64(100)
	sh := stake.NewHandler(sk)

	// 1000 first blocks not a validator
	ctx = ctx.WithBlockHeight(keeper.SignedBlocksWindow(ctx) + 1)

	// Validator created
	got := sh(ctx, newTestMsgCreateValidator(addr, val, sdk.NewInt(amt)))
	require.True(t, got.IsOK())
	validatorUpdates := stake.EndBlocker(ctx, sk)
	keeper.AddValidators(ctx, validatorUpdates)
	require.Equal(t, ck.GetCoins(ctx, sdk.AccAddress(addr)), sdk.Coins{{sk.GetParams(ctx).BondDenom, initCoins.SubRaw(amt)}})
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
	ctx, _, sk, _, keeper := createTestInput(t)
	amtInt := int64(100)
	addr, val, amt := addrs[0], pks[0], sdk.NewInt(amtInt)
	sh := stake.NewHandler(sk)
	got := sh(ctx, newTestMsgCreateValidator(addr, val, amt))
	require.True(t, got.IsOK())
	validatorUpdates := stake.EndBlocker(ctx, sk)
	keeper.AddValidators(ctx, validatorUpdates)

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
