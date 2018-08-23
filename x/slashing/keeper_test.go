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

// Test that a validator is slashed correctly
// when we discover evidence of infraction
func TestHandleDoubleSign(t *testing.T) {

	// initial setup
	ctx, ck, sk, _, keeper := createTestInput(t)
	amtInt := int64(100)
	addr, val, amt := addrs[0], pks[0], sdk.NewInt(amtInt)
	got := stake.NewHandler(sk)(ctx, newTestMsgCreateValidator(addr, val, amt))
	require.True(t, got.IsOK())
	validatorUpdates := stake.EndBlocker(ctx, sk)
	keeper.AddValidators(ctx, validatorUpdates)
	require.Equal(t, ck.GetCoins(ctx, addr), sdk.Coins{{sk.GetParams(ctx).BondDenom, initCoins.Sub(amt)}})
	require.True(t, sdk.NewRatFromInt(amt).Equal(sk.Validator(ctx, addr).GetPower()))

	// handle a signature to set signing info
	keeper.handleValidatorSignature(ctx, val.Address(), amtInt, true)

	// double sign less than max age
	keeper.handleDoubleSign(ctx, val.Address(), 0, time.Unix(0, 0), amtInt)

	// should be revoked
	require.True(t, sk.Validator(ctx, addr).GetRevoked())
	// unrevoke to measure power
	sk.Unrevoke(ctx, val)
	// power should be reduced
	require.Equal(t, sdk.NewRatFromInt(amt).Mul(sdk.NewRat(19).Quo(sdk.NewRat(20))), sk.Validator(ctx, addr).GetPower())
	ctx = ctx.WithBlockHeader(abci.Header{Time: time.Unix(1, 0).Add(keeper.MaxEvidenceAge(ctx))})

	// double sign past max age
	keeper.handleDoubleSign(ctx, val.Address(), 0, time.Unix(0, 0), amtInt)
	require.Equal(t, sdk.NewRatFromInt(amt).Mul(sdk.NewRat(19).Quo(sdk.NewRat(20))), sk.Validator(ctx, addr).GetPower())
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
	require.Equal(t, ck.GetCoins(ctx, addr), sdk.Coins{{sk.GetParams(ctx).BondDenom, initCoins.Sub(amt)}})
	require.True(t, sdk.NewRatFromInt(amt).Equal(sk.Validator(ctx, addr).GetPower()))
	info, found := keeper.getValidatorSigningInfo(ctx, sdk.ValAddress(val.Address()))
	require.False(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, int64(0), info.IndexOffset)
	require.Equal(t, int64(0), info.SignedBlocksCounter)
	// default time.Time value
	var blankTime time.Time
	require.Equal(t, blankTime, info.JailedUntil)
	height := int64(0)

	// 1000 first blocks OK
	for ; height < keeper.SignedBlocksWindow(ctx); height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.handleValidatorSignature(ctx, val.Address(), amtInt, true)
	}
	info, found = keeper.getValidatorSigningInfo(ctx, sdk.ValAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, keeper.SignedBlocksWindow(ctx), info.SignedBlocksCounter)

	// 500 blocks missed
	for ; height < keeper.SignedBlocksWindow(ctx)+(keeper.SignedBlocksWindow(ctx)-keeper.MinSignedPerWindow(ctx)); height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.handleValidatorSignature(ctx, val.Address(), amtInt, false)
	}
	info, found = keeper.getValidatorSigningInfo(ctx, sdk.ValAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, keeper.SignedBlocksWindow(ctx)-keeper.MinSignedPerWindow(ctx), info.SignedBlocksCounter)

	// validator should be bonded still
	validator, _ := sk.GetValidatorByPubKey(ctx, val)
	require.Equal(t, sdk.Bonded, validator.GetStatus())
	pool := sk.GetPool(ctx)
	require.Equal(t, int64(amtInt), pool.BondedTokens.RoundInt64())

	// 501st block missed
	ctx = ctx.WithBlockHeight(height)
	keeper.handleValidatorSignature(ctx, val.Address(), amtInt, false)
	info, found = keeper.getValidatorSigningInfo(ctx, sdk.ValAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, keeper.SignedBlocksWindow(ctx)-keeper.MinSignedPerWindow(ctx)-1, info.SignedBlocksCounter)

	// validator should have been revoked
	validator, _ = sk.GetValidatorByPubKey(ctx, val)
	require.Equal(t, sdk.Unbonded, validator.GetStatus())

	// unrevocation should fail prior to jail expiration
	got = slh(ctx, NewMsgUnrevoke(addr))
	require.False(t, got.IsOK())

	// unrevocation should succeed after jail expiration
	ctx = ctx.WithBlockHeader(abci.Header{Time: time.Unix(1, 0).Add(keeper.DowntimeUnbondDuration(ctx))})
	got = slh(ctx, NewMsgUnrevoke(addr))
	require.True(t, got.IsOK())

	// validator should be rebonded now
	validator, _ = sk.GetValidatorByPubKey(ctx, val)
	require.Equal(t, sdk.Bonded, validator.GetStatus())

	// validator should have been slashed
	pool = sk.GetPool(ctx)
	slashAmt := sdk.NewRat(amtInt).Mul(keeper.SlashFractionDowntime(ctx)).RoundInt64()
	require.Equal(t, int64(amtInt)-slashAmt, pool.BondedTokens.RoundInt64())

	// validator start height should have been changed
	info, found = keeper.getValidatorSigningInfo(ctx, sdk.ValAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, height, info.StartHeight)
	require.Equal(t, keeper.SignedBlocksWindow(ctx)-keeper.MinSignedPerWindow(ctx)-1, info.SignedBlocksCounter)

	// validator should not be immediately revoked again
	height++
	ctx = ctx.WithBlockHeight(height)
	keeper.handleValidatorSignature(ctx, val.Address(), amtInt, false)
	validator, _ = sk.GetValidatorByPubKey(ctx, val)
	require.Equal(t, sdk.Bonded, validator.GetStatus())

	// 500 signed blocks
	nextHeight := height + keeper.MinSignedPerWindow(ctx) + 1
	for ; height < nextHeight; height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.handleValidatorSignature(ctx, val.Address(), amtInt, false)
	}

	// validator should be revoked again after 500 unsigned blocks
	nextHeight = height + keeper.MinSignedPerWindow(ctx) + 1
	for ; height <= nextHeight; height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.handleValidatorSignature(ctx, val.Address(), amtInt, false)
	}
	validator, _ = sk.GetValidatorByPubKey(ctx, val)
	require.Equal(t, sdk.Unbonded, validator.GetStatus())
}

// Test a new validator entering the validator set
// Ensure that SigningInfo.StartHeight is set correctly
// and that they are not immediately revoked
func TestHandleNewValidator(t *testing.T) {
	// initial setup
	ctx, ck, sk, _, keeper := createTestInput(t)
	addr, val, amt := addrs[0], pks[0], int64(100)
	sh := stake.NewHandler(sk)
	got := sh(ctx, newTestMsgCreateValidator(addr, val, sdk.NewInt(amt)))
	require.True(t, got.IsOK())
	validatorUpdates := stake.EndBlocker(ctx, sk)
	keeper.AddValidators(ctx, validatorUpdates)
	require.Equal(t, ck.GetCoins(ctx, addr), sdk.Coins{{sk.GetParams(ctx).BondDenom, initCoins.SubRaw(amt)}})
	require.Equal(t, sdk.NewRat(amt), sk.Validator(ctx, addr).GetPower())

	// 1000 first blocks not a validator
	ctx = ctx.WithBlockHeight(keeper.SignedBlocksWindow(ctx) + 1)

	// Now a validator, for two blocks
	keeper.handleValidatorSignature(ctx, val.Address(), 100, true)
	ctx = ctx.WithBlockHeight(keeper.SignedBlocksWindow(ctx) + 2)
	keeper.handleValidatorSignature(ctx, val.Address(), 100, false)

	info, found := keeper.getValidatorSigningInfo(ctx, sdk.ValAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(keeper.SignedBlocksWindow(ctx)+1), info.StartHeight)
	require.Equal(t, int64(2), info.IndexOffset)
	require.Equal(t, int64(1), info.SignedBlocksCounter)
	require.Equal(t, time.Unix(0, 0).UTC(), info.JailedUntil)

	// validator should be bonded still, should not have been revoked or slashed
	validator, _ := sk.GetValidatorByPubKey(ctx, val)
	require.Equal(t, sdk.Bonded, validator.GetStatus())
	pool := sk.GetPool(ctx)
	require.Equal(t, int64(100), pool.BondedTokens.RoundInt64())
}

// Test a revoked validator being "down" twice
// Ensure that they're only slashed once
func TestHandleAlreadyRevoked(t *testing.T) {

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

	// validator should have been revoked and slashed
	validator, _ := sk.GetValidatorByPubKey(ctx, val)
	require.Equal(t, sdk.Unbonded, validator.GetStatus())

	// validator should have been slashed
	require.Equal(t, int64(amtInt-1), validator.GetTokens().RoundInt64())

	// another block missed
	ctx = ctx.WithBlockHeight(height)
	keeper.handleValidatorSignature(ctx, val.Address(), amtInt, false)

	// validator should not have been slashed twice
	validator, _ = sk.GetValidatorByPubKey(ctx, val)
	require.Equal(t, int64(amtInt-1), validator.GetTokens().RoundInt64())

}
