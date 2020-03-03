package slashing_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// Test a validator through uptime, downtime, revocation,
// unrevocation, starting height reset, and revocation again
func TestHandleAbsentValidator(t *testing.T) {
	// initial setup
	ctx, bk, sk, _, keeper := slashingkeeper.CreateTestInput(t, slashingkeeper.TestParams())
	power := int64(100)
	amt := sdk.TokensFromConsensusPower(power)
	addr, val := slashingkeeper.Addrs[0], slashingkeeper.Pks[0]
	sh := staking.NewHandler(sk)
	slh := slashing.NewHandler(keeper)

	res, err := sh(ctx, slashingkeeper.NewTestMsgCreateValidator(addr, val, amt))
	require.NoError(t, err)
	require.NotNil(t, res)

	staking.EndBlocker(ctx, sk)

	require.Equal(
		t, bk.GetAllBalances(ctx, sdk.AccAddress(addr)),
		sdk.NewCoins(sdk.NewCoin(sk.GetParams(ctx).BondDenom, slashingkeeper.InitTokens.Sub(amt))),
	)
	require.Equal(t, amt, sk.Validator(ctx, addr).GetBondedTokens())

	// will exist since the validator has been bonded
	info, found := keeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, int64(0), info.IndexOffset)
	require.Equal(t, int64(0), info.MissedBlocksCounter)
	require.Equal(t, time.Unix(0, 0).UTC(), info.JailedUntil)
	height := int64(0)

	// 1000 first blocks OK
	for ; height < keeper.SignedBlocksWindow(ctx); height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.HandleValidatorSignature(ctx, val.Address(), power, true)
	}
	info, found = keeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, int64(0), info.MissedBlocksCounter)

	// 500 blocks missed
	for ; height < keeper.SignedBlocksWindow(ctx)+(keeper.SignedBlocksWindow(ctx)-keeper.MinSignedPerWindow(ctx)); height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.HandleValidatorSignature(ctx, val.Address(), power, false)
	}
	info, found = keeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, keeper.SignedBlocksWindow(ctx)-keeper.MinSignedPerWindow(ctx), info.MissedBlocksCounter)

	// validator should be bonded still
	validator, _ := sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Bonded, validator.GetStatus())

	bondPool := sk.GetBondedPool(ctx)
	require.True(sdk.IntEq(t, amt, bk.GetBalance(ctx, bondPool.GetAddress(), sk.BondDenom(ctx)).Amount))

	// 501st block missed
	ctx = ctx.WithBlockHeight(height)
	keeper.HandleValidatorSignature(ctx, val.Address(), power, false)
	info, found = keeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	// counter now reset to zero
	require.Equal(t, int64(0), info.MissedBlocksCounter)

	// end block
	staking.EndBlocker(ctx, sk)

	// validator should have been jailed
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Unbonding, validator.GetStatus())

	slashAmt := amt.ToDec().Mul(keeper.SlashFractionDowntime(ctx)).RoundInt64()

	// validator should have been slashed
	require.Equal(t, amt.Int64()-slashAmt, validator.GetTokens().Int64())

	// 502nd block *also* missed (since the LastCommit would have still included the just-unbonded validator)
	height++
	ctx = ctx.WithBlockHeight(height)
	keeper.HandleValidatorSignature(ctx, val.Address(), power, false)
	info, found = keeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	require.Equal(t, int64(1), info.MissedBlocksCounter)

	// end block
	staking.EndBlocker(ctx, sk)

	// validator should not have been slashed any more, since it was already jailed
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, amt.Int64()-slashAmt, validator.GetTokens().Int64())

	// unrevocation should fail prior to jail expiration
	res, err = slh(ctx, types.NewMsgUnjail(addr))
	require.Error(t, err)
	require.Nil(t, res)

	// unrevocation should succeed after jail expiration
	ctx = ctx.WithBlockHeader(abci.Header{Time: time.Unix(1, 0).Add(keeper.DowntimeJailDuration(ctx))})
	res, err = slh(ctx, types.NewMsgUnjail(addr))
	require.NoError(t, err)
	require.NotNil(t, res)

	// end block
	staking.EndBlocker(ctx, sk)

	// validator should be rebonded now
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Bonded, validator.GetStatus())

	// validator should have been slashed
	require.Equal(t, amt.Int64()-slashAmt, bk.GetBalance(ctx, bondPool.GetAddress(), sk.BondDenom(ctx)).Amount.Int64())

	// Validator start height should not have been changed
	info, found = keeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, int64(0), info.StartHeight)
	// we've missed 2 blocks more than the maximum, so the counter was reset to 0 at 1 block more and is now 1
	require.Equal(t, int64(1), info.MissedBlocksCounter)

	// validator should not be immediately jailed again
	height++
	ctx = ctx.WithBlockHeight(height)
	keeper.HandleValidatorSignature(ctx, val.Address(), power, false)
	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Bonded, validator.GetStatus())

	// 500 signed blocks
	nextHeight := height + keeper.MinSignedPerWindow(ctx) + 1
	for ; height < nextHeight; height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.HandleValidatorSignature(ctx, val.Address(), power, false)
	}

	// end block
	staking.EndBlocker(ctx, sk)

	// validator should be jailed again after 500 unsigned blocks
	nextHeight = height + keeper.MinSignedPerWindow(ctx) + 1
	for ; height <= nextHeight; height++ {
		ctx = ctx.WithBlockHeight(height)
		keeper.HandleValidatorSignature(ctx, val.Address(), power, false)
	}

	// end block
	staking.EndBlocker(ctx, sk)

	validator, _ = sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Unbonding, validator.GetStatus())
}
