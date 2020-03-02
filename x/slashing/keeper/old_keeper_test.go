package keeper

import (
	"github.com/stretchr/testify/require"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

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
	res, err := sh(ctx, NewTestMsgCreateValidator(addr, val, amt))
	require.NoError(t, err)
	require.NotNil(t, res)

	staking.EndBlocker(ctx, sk)

	// 100 first blocks OK
	height := int64(0)
	for ; height < int64(100); height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.HandleValidatorSignature(ctx, val.Address(), power, true)
	}

	// kick first validator out of validator set
	newAmt := sdk.TokensFromConsensusPower(101)
	res, err = sh(ctx, NewTestMsgCreateValidator(Addrs[1], Pks[1], newAmt))
	require.NoError(t, err)
	require.NotNil(t, res)

	validatorUpdates := staking.EndBlocker(ctx, sk)
	require.Equal(t, 2, len(validatorUpdates))
	validator, _ := sk.GetValidator(ctx, addr)
	require.Equal(t, sdk.Unbonding, validator.Status)

	// 600 more blocks happened
	height = int64(700)
	ctx = ctx.WithBlockHeight(height)

	// validator added back in
	delTokens := sdk.TokensFromConsensusPower(50)
	res, err = sh(ctx, NewTestMsgDelegate(sdk.AccAddress(Addrs[2]), Addrs[0], delTokens))
	require.NoError(t, err)
	require.NotNil(t, res)

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
