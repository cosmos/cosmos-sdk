package keeper_test

import (
	"testing"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// Test a new validator entering the validator set
// Ensure that SigningInfo.StartHeight is set correctly
// and that they are not immediately jailed
func TestHandleNewValidator(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	addrDels := simapp.AddTestAddrsIncremental(app, ctx, 1, sdk.TokensFromConsensusPower(200))
	valAddrs := simapp.ConvertAddrsToValAddrs(addrDels)
	pks := simapp.CreateTestPubKeys(1)

	addr, val := valAddrs[0], pks[0]
	amt := sdk.TokensFromConsensusPower(100)
	sh := staking.NewHandler(app.StakingKeeper)

	ctx = ctx.WithBlockHeight(app.SlashingKeeper.SignedBlocksWindow(ctx) + 1)

	// Validator created
	res, err := sh(ctx, keeper.NewTestMsgCreateValidator(addr, val, amt))
	require.NoError(t, err)
	require.NotNil(t, res)

	staking.EndBlocker(ctx, app.StakingKeeper)

	require.Equal(
		t, app.BankKeeper.GetAllBalances(ctx, sdk.AccAddress(addr)),
		sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.GetParams(ctx).BondDenom, InitTokens.Sub(amt))),
	)
	require.Equal(t, amt, app.StakingKeeper.Validator(ctx, addr).GetBondedTokens())

	// Now a validator, for two blocks
	app.SlashingKeeper.HandleValidatorSignature(ctx, val.Address(), 100, true)
	ctx = ctx.WithBlockHeight(app.SlashingKeeper.SignedBlocksWindow(ctx) + 2)
	app.SlashingKeeper.HandleValidatorSignature(ctx, val.Address(), 100, false)

	info, found := app.SlashingKeeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(val.Address()))
	require.True(t, found)
	require.Equal(t, app.SlashingKeeper.SignedBlocksWindow(ctx)+1, info.StartHeight)
	require.Equal(t, int64(2), info.IndexOffset)
	require.Equal(t, int64(1), info.MissedBlocksCounter)
	require.Equal(t, time.Unix(0, 0).UTC(), info.JailedUntil)

	// validator should be bonded still, should not have been jailed or slashed
	validator, _ := app.StakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Bonded, validator.GetStatus())
	bondPool := app.StakingKeeper.GetBondedPool(ctx)
	expTokens := sdk.TokensFromConsensusPower(100)
	require.Equal(t, expTokens.Int64(), app.BankKeeper.GetBalance(ctx, bondPool.GetAddress(), app.StakingKeeper.BondDenom(ctx)).Amount.Int64())
}

// Test a jailed validator being "down" twice
// Ensure that they're only slashed once
func TestHandleAlreadyJailed(t *testing.T) {
	// initial setup
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})
	power := int64(100)

	amt := sdk.TokensFromConsensusPower(power)
	addrDels := simapp.AddTestAddrsIncremental(app, ctx, 1, sdk.TokensFromConsensusPower(200))
	valAddrs := simapp.ConvertAddrsToValAddrs(addrDels)
	pks := simapp.CreateTestPubKeys(1)

	addr, val := valAddrs[0], pks[0]
	sh := staking.NewHandler(app.StakingKeeper)
	res, err := sh(ctx, keeper.NewTestMsgCreateValidator(addr, val, amt))
	require.NoError(t, err)
	require.NotNil(t, res)

	staking.EndBlocker(ctx, app.StakingKeeper)

	// 1000 first blocks OK
	height := int64(0)
	for ; height < app.SlashingKeeper.SignedBlocksWindow(ctx); height++ {
		ctx = ctx.WithBlockHeight(height)
		app.SlashingKeeper.HandleValidatorSignature(ctx, val.Address(), power, true)
	}

	// 501 blocks missed
	for ; height < app.SlashingKeeper.SignedBlocksWindow(ctx)+(app.SlashingKeeper.SignedBlocksWindow(ctx)-app.SlashingKeeper.MinSignedPerWindow(ctx))+1; height++ {
		ctx = ctx.WithBlockHeight(height)
		app.SlashingKeeper.HandleValidatorSignature(ctx, val.Address(), power, false)
	}

	// end block
	staking.EndBlocker(ctx, app.StakingKeeper)

	// validator should have been jailed and slashed
	validator, _ := app.StakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, sdk.Unbonding, validator.GetStatus())

	// validator should have been slashed
	resultingTokens := amt.Sub(sdk.TokensFromConsensusPower(1))
	require.Equal(t, resultingTokens, validator.GetTokens())

	// another block missed
	ctx = ctx.WithBlockHeight(height)
	app.SlashingKeeper.HandleValidatorSignature(ctx, val.Address(), power, false)

	// validator should not have been slashed twice
	validator, _ = app.StakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(val))
	require.Equal(t, resultingTokens, validator.GetTokens())
}
