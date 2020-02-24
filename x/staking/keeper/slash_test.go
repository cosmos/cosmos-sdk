package keeper_test

import (
	"testing"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

// initConfig creates 3 validators and bootstrap the app.
func initConfig(t *testing.T, power int64) (*simapp.SimApp, sdk.Context, []sdk.AccAddress, []sdk.ValAddress) {
	_, app, ctx := getBaseSimappWithCustomKeeper()

	addrDels := simapp.AddTestAddrsIncremental(app, ctx, 100, sdk.NewInt(10000))
	addrVals := simapp.ConvertAddrsToValAddrs(addrDels)

	amt := sdk.TokensFromConsensusPower(power)
	totalSupply := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), amt.MulRaw(int64(len(addrDels)))))

	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
	err := app.BankKeeper.SetBalances(ctx, notBondedPool.GetAddress(), totalSupply)
	require.NoError(t, err)
	app.SupplyKeeper.SetModuleAccount(ctx, notBondedPool)

	numVals := int64(3)
	bondedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), amt.MulRaw(numVals)))
	bondedPool := app.StakingKeeper.GetBondedPool(ctx)
	err = app.BankKeeper.SetBalances(ctx, bondedPool.GetAddress(), bondedCoins)
	require.NoError(t, err)
	app.SupplyKeeper.SetModuleAccount(ctx, bondedPool)

	for i := int64(0); i < numVals; i++ {
		validator := types.NewValidator(addrVals[i], PKs[i], types.Description{})
		validator, _ = validator.AddTokensFromDel(amt)
		validator = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validator, true)
		app.StakingKeeper.SetValidatorByConsAddr(ctx, validator)
	}

	return app, ctx, addrDels, addrVals
}

// tests Jail, Unjail
func TestRevocation(t *testing.T) {
	app, ctx, _, addrVals := initConfig(t, 5)

	consAddr := sdk.ConsAddress(PKs[0].Address())

	// initial state
	val, found := app.StakingKeeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.False(t, val.IsJailed())

	// test jail
	app.StakingKeeper.Jail(ctx, consAddr)
	val, found = app.StakingKeeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.True(t, val.IsJailed())

	// test unjail
	app.StakingKeeper.Unjail(ctx, consAddr)
	val, found = app.StakingKeeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.False(t, val.IsJailed())
}

// tests slashUnbondingDelegation
func TestSlashUnbondingDelegation(t *testing.T) {
	app, ctx, addrDels, addrVals := initConfig(t, 10)

	fraction := sdk.NewDecWithPrec(5, 1)

	// set an unbonding delegation with expiration timestamp (beyond which the
	// unbonding delegation shouldn't be slashed)
	ubd := types.NewUnbondingDelegation(addrDels[0], addrVals[0], 0,
		time.Unix(5, 0), sdk.NewInt(10))

	app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)

	// unbonding started prior to the infraction height, stakw didn't contribute
	slashAmount := app.StakingKeeper.SlashUnbondingDelegation(ctx, ubd, 1, fraction)
	require.Equal(t, int64(0), slashAmount.Int64())

	// after the expiration time, no longer eligible for slashing
	ctx = ctx.WithBlockHeader(abci.Header{Time: time.Unix(10, 0)})
	app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)
	slashAmount = app.StakingKeeper.SlashUnbondingDelegation(ctx, ubd, 0, fraction)
	require.Equal(t, int64(0), slashAmount.Int64())

	// test valid slash, before expiration timestamp and to which stake contributed
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
	oldUnbondedPoolBalances := app.BankKeeper.GetAllBalances(ctx, notBondedPool.GetAddress())
	ctx = ctx.WithBlockHeader(abci.Header{Time: time.Unix(0, 0)})
	app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)
	slashAmount = app.StakingKeeper.SlashUnbondingDelegation(ctx, ubd, 0, fraction)
	require.Equal(t, int64(5), slashAmount.Int64())
	ubd, found := app.StakingKeeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)

	// initial balance unchanged
	require.Equal(t, sdk.NewInt(10), ubd.Entries[0].InitialBalance)

	// balance decreased
	require.Equal(t, sdk.NewInt(5), ubd.Entries[0].Balance)
	newUnbondedPoolBalances := app.BankKeeper.GetAllBalances(ctx, notBondedPool.GetAddress())
	diffTokens := oldUnbondedPoolBalances.Sub(newUnbondedPoolBalances)
	require.Equal(t, int64(5), diffTokens.AmountOf(app.StakingKeeper.BondDenom(ctx)).Int64())
}

// tests slashRedelegation
func TestSlashRedelegation(t *testing.T) {
	app, ctx, addrDels, addrVals := initConfig(t, 10)
	fraction := sdk.NewDecWithPrec(5, 1)

	// add bonded tokens to pool for (re)delegations
	startCoins := sdk.NewCoins(sdk.NewInt64Coin(app.StakingKeeper.BondDenom(ctx), 15))
	bondedPool := app.StakingKeeper.GetBondedPool(ctx)
	balances := app.BankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())

	require.NoError(t, app.BankKeeper.SetBalances(ctx, bondedPool.GetAddress(), balances.Add(startCoins...)))
	app.SupplyKeeper.SetModuleAccount(ctx, bondedPool)

	// set a redelegation with an expiration timestamp beyond which the
	// redelegation shouldn't be slashed
	rd := types.NewRedelegation(addrDels[0], addrVals[0], addrVals[1], 0,
		time.Unix(5, 0), sdk.NewInt(10), sdk.NewDec(10))

	app.StakingKeeper.SetRedelegation(ctx, rd)

	// set the associated delegation
	del := types.NewDelegation(addrDels[0], addrVals[1], sdk.NewDec(10))
	app.StakingKeeper.SetDelegation(ctx, del)

	// started redelegating prior to the current height, stake didn't contribute to infraction
	validator, found := app.StakingKeeper.GetValidator(ctx, addrVals[1])
	require.True(t, found)
	slashAmount := app.StakingKeeper.SlashRedelegation(ctx, validator, rd, 1, fraction)
	require.Equal(t, int64(0), slashAmount.Int64())

	// after the expiration time, no longer eligible for slashing
	ctx = ctx.WithBlockHeader(abci.Header{Time: time.Unix(10, 0)})
	app.StakingKeeper.SetRedelegation(ctx, rd)
	validator, found = app.StakingKeeper.GetValidator(ctx, addrVals[1])
	require.True(t, found)
	slashAmount = app.StakingKeeper.SlashRedelegation(ctx, validator, rd, 0, fraction)
	require.Equal(t, int64(0), slashAmount.Int64())

	balances = app.BankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())

	// test valid slash, before expiration timestamp and to which stake contributed
	ctx = ctx.WithBlockHeader(abci.Header{Time: time.Unix(0, 0)})
	app.StakingKeeper.SetRedelegation(ctx, rd)
	validator, found = app.StakingKeeper.GetValidator(ctx, addrVals[1])
	require.True(t, found)
	slashAmount = app.StakingKeeper.SlashRedelegation(ctx, validator, rd, 0, fraction)
	require.Equal(t, int64(5), slashAmount.Int64())
	rd, found = app.StakingKeeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.True(t, found)
	require.Len(t, rd.Entries, 1)

	// end block
	updates := app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(updates))

	// initialbalance unchanged
	require.Equal(t, sdk.NewInt(10), rd.Entries[0].InitialBalance)

	// shares decreased
	del, found = app.StakingKeeper.GetDelegation(ctx, addrDels[0], addrVals[1])
	require.True(t, found)
	require.Equal(t, int64(5), del.Shares.RoundInt64())

	// pool bonded tokens should decrease
	burnedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), slashAmount))
	require.Equal(t, balances.Sub(burnedCoins), app.BankKeeper.GetAllBalances(ctx, bondedPool.GetAddress()))
}
