package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/simapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestUnbondingDelegationsMaxEntries(t *testing.T) {
	_, app, ctx := createTestInput(t)

	addrDels := simapp.AddTestAddrsIncremental(app, ctx, 1, sdk.NewInt(10000))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	startTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)

	bondDenom := app.StakingKeeper.BondDenom(ctx)
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)

	require.NoError(t, banktestutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(bondDenom, startTokens))))
	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	// create a validator and a delegator to that validator
	validator := testutil.NewValidator(t, addrVals[0], PKs[0])

	validator, issuedShares := validator.AddTokensFromDel(startTokens)
	require.Equal(t, startTokens, issuedShares.RoundInt())

	validator = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validator, true)
	require.True(math.IntEq(t, startTokens, validator.BondedTokens()))
	require.True(t, validator.IsBonded())

	delegation := types.NewDelegation(addrDels[0], addrVals[0], issuedShares)
	app.StakingKeeper.SetDelegation(ctx, delegation)

	maxEntries := app.StakingKeeper.MaxEntries(ctx)

	oldBonded := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	oldNotBonded := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// should all pass
	var completionTime time.Time
	for i := int64(0); i < int64(maxEntries); i++ {
		var err error
		ctx = ctx.WithBlockHeight(i)
		completionTime, err = app.StakingKeeper.Undelegate(ctx, addrDels[0], addrVals[0], math.LegacyNewDec(1))
		require.NoError(t, err)
	}

	newBonded := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	require.True(math.IntEq(t, newBonded, oldBonded.SubRaw(int64(maxEntries))))
	require.True(math.IntEq(t, newNotBonded, oldNotBonded.AddRaw(int64(maxEntries))))

	oldBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	oldNotBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// an additional unbond should fail due to max entries
	_, err := app.StakingKeeper.Undelegate(ctx, addrDels[0], addrVals[0], math.LegacyNewDec(1))
	require.Error(t, err)

	newBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	require.True(math.IntEq(t, newBonded, oldBonded))
	require.True(math.IntEq(t, newNotBonded, oldNotBonded))

	// mature unbonding delegations
	ctx = ctx.WithBlockTime(completionTime)
	_, err = app.StakingKeeper.CompleteUnbonding(ctx, addrDels[0], addrVals[0])
	require.NoError(t, err)

	newBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	require.True(math.IntEq(t, newBonded, oldBonded))
	require.True(math.IntEq(t, newNotBonded, oldNotBonded.SubRaw(int64(maxEntries))))

	oldNotBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// unbonding  should work again
	_, err = app.StakingKeeper.Undelegate(ctx, addrDels[0], addrVals[0], math.LegacyNewDec(1))
	require.NoError(t, err)

	newBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	require.True(math.IntEq(t, newBonded, oldBonded.SubRaw(1)))
	require.True(math.IntEq(t, newNotBonded, oldNotBonded.AddRaw(1)))
}
