package slashing

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

// Test that a validator is slashed correctly
// when we discover evidence of equivocation
func TestHandleDoubleSign(t *testing.T) {

	// initial setup
	ctx, ck, sk, keeper := createTestInput(t)
	addr, val, amt := addrs[0], pks[0], int64(100)
	got := stake.NewHandler(sk)(ctx, newTestMsgCreateValidator(addr, val, amt))
	require.True(t, got.IsOK())
	stake.EndBlocker(ctx, sk)
	require.Equal(t, ck.GetCoins(ctx, addr), sdk.Coins{{sk.GetParams(ctx).BondDenom, initCoins - amt}})
	require.Equal(t, sdk.NewRat(amt), sk.Validator(ctx, addr).GetPower())

	// double sign less than max age
	keeper.handleDoubleSign(ctx, 0, 0, val)
	require.Equal(t, sdk.NewRat(amt).Mul(sdk.NewRat(19).Quo(sdk.NewRat(20))), sk.Validator(ctx, addr).GetPower())
	ctx = ctx.WithBlockHeader(abci.Header{Time: 300})

	// double sign past max age
	keeper.handleDoubleSign(ctx, 0, 0, val)
	require.Equal(t, sdk.NewRat(amt).Mul(sdk.NewRat(19).Quo(sdk.NewRat(20))), sk.Validator(ctx, addr).GetPower())
}

// Test a validator through uptime, downtime, revocation,
// unrevocation, starting height reset, and revocation again
func TestHandleAbsentValidator(t *testing.T) {

	// initial setup
	ctx, ck, sk, keeper := createTestInput(t)
	addr, val, amt := addrs[0], pks[0], int64(100)
	sh := stake.NewHandler(sk)
	slh := NewHandler(keeper)
	got := sh(ctx, newTestMsgCreateValidator(addr, val, amt))
	require.True(t, got.IsOK())
	stake.EndBlocker(ctx, sk)
	require.Equal(t, ck.GetCoins(ctx, addr), sdk.Coins{{sk.GetParams(ctx).BondDenom, initCoins - amt}})
	require.Equal(t, sdk.NewRat(amt), sk.Validator(ctx, addr).GetPower())
	info, found := keeper.getValidatorSigningInfo(ctx, val.Address())
	require.False(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, int64(0), info.IndexOffset)
	require.Equal(t, int64(0), info.SignedBlocksCounter)
	require.Equal(t, int64(0), info.JailedUntil)
	height := int64(0)

	// 1000 first blocks OK
	for ; height < 1000; height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.handleValidatorSignature(ctx, val, true)
	}
	info, found = keeper.getValidatorSigningInfo(ctx, val.Address())
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, SignedBlocksWindow, info.SignedBlocksCounter)

	// 50 blocks missed
	for ; height < 1050; height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.handleValidatorSignature(ctx, val, false)
	}
	info, found = keeper.getValidatorSigningInfo(ctx, val.Address())
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, SignedBlocksWindow-50, info.SignedBlocksCounter)

	// validator should be bonded still
	validator, _ := sk.GetValidatorByPubKey(ctx, val)
	require.Equal(t, sdk.Bonded, validator.GetStatus())
	pool := sk.GetPool(ctx)
	require.Equal(t, int64(100), pool.BondedTokens)

	// 51st block missed
	ctx = ctx.WithBlockHeight(height)
	keeper.handleValidatorSignature(ctx, val, false)
	info, found = keeper.getValidatorSigningInfo(ctx, val.Address())
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, SignedBlocksWindow-51, info.SignedBlocksCounter)

	// validator should have been revoked
	validator, _ = sk.GetValidatorByPubKey(ctx, val)
	require.Equal(t, sdk.Unbonded, validator.GetStatus())

	// unrevocation should fail prior to jail expiration
	got = slh(ctx, NewMsgUnrevoke(addr))
	require.False(t, got.IsOK())

	// unrevocation should succeed after jail expiration
	ctx = ctx.WithBlockHeader(abci.Header{Time: int64(86400 * 2)})
	got = slh(ctx, NewMsgUnrevoke(addr))
	require.True(t, got.IsOK())

	// validator should be rebonded now
	validator, _ = sk.GetValidatorByPubKey(ctx, val)
	require.Equal(t, sdk.Bonded, validator.GetStatus())

	// validator should have been slashed
	pool = sk.GetPool(ctx)
	require.Equal(t, int64(99), pool.BondedTokens)

	// validator start height should have been changed
	info, found = keeper.getValidatorSigningInfo(ctx, val.Address())
	require.True(t, found)
	require.Equal(t, height, info.StartHeight)
	require.Equal(t, SignedBlocksWindow-51, info.SignedBlocksCounter)

	// validator should not be immediately revoked again
	height++
	ctx = ctx.WithBlockHeight(height)
	keeper.handleValidatorSignature(ctx, val, false)
	validator, _ = sk.GetValidatorByPubKey(ctx, val)
	require.Equal(t, sdk.Bonded, validator.GetStatus())

	// validator should be revoked again after 100 unsigned blocks
	nextHeight := height + 100
	for ; height <= nextHeight; height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.handleValidatorSignature(ctx, val, false)
	}
	validator, _ = sk.GetValidatorByPubKey(ctx, val)
	require.Equal(t, sdk.Unbonded, validator.GetStatus())
}

// Test a new validator entering the validator set
// Ensure that SigningInfo.StartHeight is set correctly
// and that they are not immediately revoked
func TestHandleNewValidator(t *testing.T) {
	// initial setup
	ctx, ck, sk, keeper := createTestInput(t)
	addr, val, amt := addrs[0], pks[0], int64(100)
	sh := stake.NewHandler(sk)
	got := sh(ctx, newTestMsgCreateValidator(addr, val, amt))
	require.True(t, got.IsOK())
	stake.EndBlocker(ctx, sk)
	require.Equal(t, ck.GetCoins(ctx, addr), sdk.Coins{{sk.GetParams(ctx).BondDenom, initCoins - amt}})
	require.Equal(t, sdk.NewRat(amt), sk.Validator(ctx, addr).GetPower())

	// 1000 first blocks not a validator
	ctx = ctx.WithBlockHeight(1001)

	// Now a validator, for two blocks
	keeper.handleValidatorSignature(ctx, val, true)
	ctx = ctx.WithBlockHeight(1002)
	keeper.handleValidatorSignature(ctx, val, false)

	info, found := keeper.getValidatorSigningInfo(ctx, val.Address())
	require.True(t, found)
	require.Equal(t, int64(1001), info.StartHeight)
	require.Equal(t, int64(2), info.IndexOffset)
	require.Equal(t, int64(1), info.SignedBlocksCounter)
	require.Equal(t, int64(0), info.JailedUntil)

	// validator should be bonded still, should not have been revoked or slashed
	validator, _ := sk.GetValidatorByPubKey(ctx, val)
	require.Equal(t, sdk.Bonded, validator.GetStatus())
	pool := sk.GetPool(ctx)
	require.Equal(t, int64(100), pool.BondedTokens)
}
