package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

type MockStakingHooks struct {
	types.StakingHooksTemplate
	afterUnbondingOpInitiated func(uint64)
}

func (h MockStakingHooks) AfterUnbondingOpInitiated(_ sdk.Context, id uint64) {
	h.afterUnbondingOpInitiated(id)
}

func setup(t *testing.T, hookCalled *bool, ubdeID *uint64) (
	app *simapp.SimApp, ctx sdk.Context, bondDenom string, addrDels []sdk.AccAddress, addrVals []sdk.ValAddress,
) {
	_, app, ctx = createTestInput()

	stakingKeeper := keeper.NewKeeper(
		app.AppCodec(),
		app.GetKey(types.StoreKey),
		app.GetKey(types.StoreKey),
		app.AccountKeeper,
		app.BankKeeper,
		app.GetSubspace(types.ModuleName),
	)

	myHooks := MockStakingHooks{
		afterUnbondingOpInitiated: func(id uint64) {
			*hookCalled = true
			// save id
			*ubdeID = id
			// call back to stop unbonding
			res := app.StakingKeeper.PutUnbondingOpOnHold(ctx, id)
			require.True(t, res, "could not find unbonding op")
		},
	}

	stakingKeeper.SetHooks(
		types.NewMultiStakingHooks(myHooks),
	)

	app.StakingKeeper = stakingKeeper

	addrDels = simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(10000))
	addrVals = simapp.ConvertAddrsToValAddrs(addrDels)

	valTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)
	startTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 20)

	bondDenom = app.StakingKeeper.BondDenom(ctx)
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)

	require.NoError(t, simapp.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(bondDenom, startTokens))))
	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	// Create a validator
	validator1 := teststaking.NewValidator(t, addrVals[0], PKs[0])
	validator1, issuedShares1 := validator1.AddTokensFromDel(valTokens)
	require.Equal(t, valTokens, issuedShares1.RoundInt())

	validator1 = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validator1, true)
	require.True(sdk.IntEq(t, valTokens, validator1.BondedTokens()))
	require.True(t, validator1.IsBonded())

	// Create a delegator
	delegation := types.NewDelegation(addrDels[0], addrVals[0], issuedShares1)
	app.StakingKeeper.SetDelegation(ctx, delegation)

	// Create a validator to redelegate to
	validator2 := teststaking.NewValidator(t, addrVals[1], PKs[1])
	validator2, issuedShares2 := validator2.AddTokensFromDel(valTokens)
	require.Equal(t, valTokens, issuedShares2.RoundInt())

	validator2 = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validator2, true)
	require.Equal(t, types.Bonded, validator2.Status)
	require.True(t, validator2.IsBonded())

	return
}

func doUnbondingDelegation(
	t *testing.T, app *simapp.SimApp, ctx sdk.Context, bondDenom string, addrDels []sdk.AccAddress, addrVals []sdk.ValAddress, hookCalled *bool,
) (completionTime time.Time, bondedAmt sdk.Int, notBondedAmt sdk.Int) {
	// UNDELEGATE
	// Save original bonded and unbonded amounts
	bondedAmt1 := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	notBondedAmt1 := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	var err error
	completionTime, err = app.StakingKeeper.Undelegate(ctx, addrDels[0], addrVals[0], sdk.NewDec(1))
	require.NoError(t, err)

	// check that the unbonding actually happened
	bondedAmt2 := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	notBondedAmt2 := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	// Bonded amount is less
	require.True(sdk.IntEq(t, bondedAmt1.SubRaw(1), bondedAmt2))
	// Unbonded amount is more
	require.True(sdk.IntEq(t, notBondedAmt1.AddRaw(1), notBondedAmt2))

	// Check that the unbonding happened- we look up the entry and see that it has the correct number of shares
	unbondingDelegations := app.StakingKeeper.GetUnbondingDelegationsFromValidator(ctx, addrVals[0])
	require.Equal(t, sdk.NewInt(1), unbondingDelegations[0].Entries[0].Balance)

	// check that our hook was called
	require.True(t, *hookCalled)

	return completionTime, bondedAmt2, notBondedAmt2
}

func doRedelegation(
	t *testing.T, app *simapp.SimApp, ctx sdk.Context, addrDels []sdk.AccAddress, addrVals []sdk.ValAddress, hookCalled *bool,
) (completionTime time.Time) {
	var err error
	completionTime, err = app.StakingKeeper.BeginRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1], sdk.NewDec(1))
	require.NoError(t, err)

	// Check that the redelegation happened- we look up the entry and see that it has the correct number of shares
	redelegations := app.StakingKeeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
	require.Equal(t, 1, len(redelegations))
	require.Equal(t, sdk.NewDec(1), redelegations[0].Entries[0].SharesDst)

	// check that our hook was called
	require.True(t, *hookCalled)

	return completionTime
}

func TestRedelegationOnHold1(t *testing.T) {
	var hookCalled bool
	var ubdeID uint64
	app, ctx, _, addrDels, addrVals := setup(t, &hookCalled, &ubdeID)
	completionTime := doRedelegation(t, app, ctx, addrDels, addrVals, &hookCalled)

	// CONSUMER CHAIN'S UNBONDING PERIOD ENDS - UNBONDING CANNOT COMPLETE
	err := app.StakingKeeper.UnbondingOpCanComplete(ctx, ubdeID)
	require.NoError(t, err)

	// Redelegation is not complete - still exists
	redelegations := app.StakingKeeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
	require.Equal(t, 1, len(redelegations))

	// PROVIDER CHAIN'S UNBONDING PERIOD ENDS - COMPLETE STOPPED UNBONDING
	ctx = ctx.WithBlockTime(completionTime)
	_, err = app.StakingKeeper.CompleteRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.NoError(t, err)

	// Redelegation is complete and record is gone
	redelegations = app.StakingKeeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
	require.Equal(t, 0, len(redelegations))
}

func TestRedelegationOnHold2(t *testing.T) {
	var hookCalled bool
	var ubdeID uint64
	app, ctx, _, addrDels, addrVals := setup(t, &hookCalled, &ubdeID)
	completionTime := doRedelegation(t, app, ctx, addrDels, addrVals, &hookCalled)

	// PROVIDER CHAIN'S UNBONDING PERIOD ENDS - UNBONDING CANNOT COMPLETE
	ctx = ctx.WithBlockTime(completionTime)
	_, err := app.StakingKeeper.CompleteRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.NoError(t, err)

	// Redelegation is not complete - still exists
	redelegations := app.StakingKeeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
	require.Equal(t, 1, len(redelegations))

	// CONSUMER CHAIN'S UNBONDING PERIOD ENDS - COMPLETE STOPPED UNBONDING
	err = app.StakingKeeper.UnbondingOpCanComplete(ctx, ubdeID)
	require.NoError(t, err)

	// Redelegation is complete and record is gone
	redelegations = app.StakingKeeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
	require.Equal(t, 0, len(redelegations))
}

func TestUnbondingDelegationOnHold1(t *testing.T) {
	var hookCalled bool
	var ubdeID uint64
	app, ctx, bondDenom, addrDels, addrVals := setup(t, &hookCalled, &ubdeID)
	completionTime, bondedAmt1, notBondedAmt1 := doUnbondingDelegation(t, app, ctx, bondDenom, addrDels, addrVals, &hookCalled)

	// CONSUMER CHAIN'S UNBONDING PERIOD ENDS - UNBONDING CANNOT COMPLETE
	err := app.StakingKeeper.UnbondingOpCanComplete(ctx, ubdeID)
	require.NoError(t, err)

	bondedAmt3 := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	notBondedAmt3 := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// Bonded and unbonded amounts are the same as before because the completionTime has not yet passed and so the
	// unbondingDelegation has not completed
	require.True(sdk.IntEq(t, bondedAmt1, bondedAmt3))
	require.True(sdk.IntEq(t, notBondedAmt1, notBondedAmt3))

	// PROVIDER CHAIN'S UNBONDING PERIOD ENDS - COMPLETE STOPPED UNBONDING
	ctx = ctx.WithBlockTime(completionTime)
	_, err = app.StakingKeeper.CompleteUnbonding(ctx, addrDels[0], addrVals[0])
	require.NoError(t, err)

	// Check that the unbonding was finally completed
	bondedAmt5 := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	notBondedAmt5 := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	require.True(sdk.IntEq(t, bondedAmt1, bondedAmt5))
	// Not bonded amount back to what it was originaly
	require.True(sdk.IntEq(t, notBondedAmt1.SubRaw(1), notBondedAmt5))
}

func TestUnbondingDelegationOnHold2(t *testing.T) {
	var hookCalled bool
	var ubdeID uint64
	app, ctx, bondDenom, addrDels, addrVals := setup(t, &hookCalled, &ubdeID)
	completionTime, bondedAmt1, notBondedAmt1 := doUnbondingDelegation(t, app, ctx, bondDenom, addrDels, addrVals, &hookCalled)

	// PROVIDER CHAIN'S UNBONDING PERIOD ENDS - UNBONDING CANNOT COMPLETE
	ctx = ctx.WithBlockTime(completionTime)
	_, err := app.StakingKeeper.CompleteUnbonding(ctx, addrDels[0], addrVals[0])
	require.NoError(t, err)

	bondedAmt3 := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	notBondedAmt3 := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// Bonded and unbonded amounts are the same as before because the completionTime has not yet passed and so the
	// unbondingDelegation has not completed
	require.True(sdk.IntEq(t, bondedAmt1, bondedAmt3))
	require.True(sdk.IntEq(t, notBondedAmt1, notBondedAmt3))

	// CONSUMER CHAIN'S UNBONDING PERIOD ENDS - COMPLETE STOPPED UNBONDING
	err = app.StakingKeeper.UnbondingOpCanComplete(ctx, ubdeID)
	require.NoError(t, err)

	// Check that the unbonding was finally completed
	bondedAmt5 := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	notBondedAmt5 := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	require.True(sdk.IntEq(t, bondedAmt1, bondedAmt5))
	// Not bonded amount back to what it was originaly
	require.True(sdk.IntEq(t, notBondedAmt1.SubRaw(1), notBondedAmt5))
}
