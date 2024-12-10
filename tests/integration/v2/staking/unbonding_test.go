package staking

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"

	"cosmossdk.io/core/header"
	"cosmossdk.io/math"
	banktestutil "cosmossdk.io/x/bank/testutil"
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	"cosmossdk.io/x/staking/testutil"
	"cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetupUnbondingTests creates two validators and setup mocked staking hooks for testing unbonding
func SetupUnbondingTests(t *testing.T, f *fixture, hookCalled *bool, ubdeID *uint64) (bondDenom string, addrDels []sdk.AccAddress, addrVals []sdk.ValAddress) {
	t.Helper()
	// setup hooks
	mockCtrl := gomock.NewController(t)
	mockStackingHooks := testutil.NewMockStakingHooks(mockCtrl)
	mockStackingHooks.EXPECT().AfterDelegationModified(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockStackingHooks.EXPECT().AfterUnbondingInitiated(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx sdk.Context, id uint64) error {
		*hookCalled = true
		// save id
		*ubdeID = id
		// call back to stop unbonding
		err := f.stakingKeeper.PutUnbondingOnHold(f.ctx, id)
		assert.NilError(t, err)

		return nil
	}).AnyTimes()
	mockStackingHooks.EXPECT().AfterValidatorBeginUnbonding(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockStackingHooks.EXPECT().AfterValidatorBonded(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockStackingHooks.EXPECT().AfterValidatorCreated(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockStackingHooks.EXPECT().AfterValidatorRemoved(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockStackingHooks.EXPECT().BeforeDelegationCreated(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockStackingHooks.EXPECT().BeforeDelegationRemoved(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockStackingHooks.EXPECT().BeforeDelegationSharesModified(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockStackingHooks.EXPECT().BeforeValidatorModified(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockStackingHooks.EXPECT().BeforeValidatorSlashed(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockStackingHooks.EXPECT().AfterConsensusPubKeyUpdate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	f.stakingKeeper.SetHooks(types.NewMultiStakingHooks(mockStackingHooks))

	addrDels = simtestutil.AddTestAddrsIncremental(f.bankKeeper, f.stakingKeeper, f.ctx, 2, math.NewInt(10000))
	addrVals = simtestutil.ConvertAddrsToValAddrs(addrDels)

	valTokens := f.stakingKeeper.TokensFromConsensusPower(f.ctx, 10)
	startTokens := f.stakingKeeper.TokensFromConsensusPower(f.ctx, 20)

	bondDenom, err := f.stakingKeeper.BondDenom(f.ctx)
	assert.NilError(t, err)
	notBondedPool := f.stakingKeeper.GetNotBondedPool(f.ctx)

	assert.NilError(t, banktestutil.FundModuleAccount(f.ctx, f.bankKeeper, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(bondDenom, startTokens))))
	f.accountKeeper.SetModuleAccount(f.ctx, notBondedPool)

	// Create a validator
	validator1 := testutil.NewValidator(t, addrVals[0], PKs[0])
	validator1, issuedShares1 := validator1.AddTokensFromDel(valTokens)
	assert.DeepEqual(t, valTokens, issuedShares1.RoundInt())

	validator1 = stakingkeeper.TestingUpdateValidator(f.stakingKeeper, f.ctx, validator1, true)
	assert.Assert(math.IntEq(t, valTokens, validator1.BondedTokens()))
	assert.Assert(t, validator1.IsBonded())

	// Create a delegator
	delegation := types.NewDelegation(addrDels[0].String(), addrVals[0].String(), issuedShares1)
	assert.NilError(t, f.stakingKeeper.SetDelegation(f.ctx, delegation))

	// Create a validator to redelegate to
	validator2 := testutil.NewValidator(t, addrVals[1], PKs[1])
	validator2, issuedShares2 := validator2.AddTokensFromDel(valTokens)
	assert.DeepEqual(t, valTokens, issuedShares2.RoundInt())

	validator2 = stakingkeeper.TestingUpdateValidator(f.stakingKeeper, f.ctx, validator2, true)
	assert.Equal(t, types.Bonded, validator2.Status)
	assert.Assert(t, validator2.IsBonded())

	return bondDenom, addrDels, addrVals
}

func doUnbondingDelegation(
	t *testing.T,
	stakingKeeper *stakingkeeper.Keeper,
	bankKeeper types.BankKeeper,
	ctx context.Context,
	bondDenom string,
	addrDels []sdk.AccAddress,
	addrVals []sdk.ValAddress,
	hookCalled *bool,
) (completionTime time.Time, bondedAmt, notBondedAmt math.Int) {
	t.Helper()
	// UNDELEGATE
	// Save original bonded and unbonded amounts
	bondedAmt1 := bankKeeper.GetBalance(ctx, stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	notBondedAmt1 := bankKeeper.GetBalance(ctx, stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	var err error
	undelegateAmount := math.LegacyNewDec(1)
	completionTime, undelegatedAmount, err := stakingKeeper.Undelegate(ctx, addrDels[0], addrVals[0], undelegateAmount)
	assert.NilError(t, err)
	assert.Assert(t, undelegateAmount.Equal(math.LegacyNewDecFromInt(undelegatedAmount)))
	// check that the unbonding actually happened
	bondedAmt2 := bankKeeper.GetBalance(ctx, stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	notBondedAmt2 := bankKeeper.GetBalance(ctx, stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	// Bonded amount is less
	assert.Assert(math.IntEq(t, bondedAmt1.SubRaw(1), bondedAmt2))
	// Unbonded amount is more
	assert.Assert(math.IntEq(t, notBondedAmt1.AddRaw(1), notBondedAmt2))

	// Check that the unbonding happened- we look up the entry and see that it has the correct number of shares
	unbondingDelegations, err := stakingKeeper.GetUnbondingDelegationsFromValidator(ctx, addrVals[0])
	assert.NilError(t, err)
	assert.DeepEqual(t, math.NewInt(1), unbondingDelegations[0].Entries[0].Balance)

	// check that our hook was called
	assert.Assert(t, *hookCalled)

	return completionTime, bondedAmt2, notBondedAmt2
}

func doRedelegation(
	t *testing.T,
	stakingKeeper *stakingkeeper.Keeper,
	ctx context.Context,
	addrDels []sdk.AccAddress,
	addrVals []sdk.ValAddress,
	hookCalled *bool,
) (completionTime time.Time) {
	t.Helper()
	var err error
	completionTime, err = stakingKeeper.BeginRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1], math.LegacyNewDec(1))
	assert.NilError(t, err)

	// Check that the redelegation happened- we look up the entry and see that it has the correct number of shares
	redelegations, err := stakingKeeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
	assert.NilError(t, err)
	assert.Equal(t, 1, len(redelegations))
	assert.DeepEqual(t, math.LegacyNewDec(1), redelegations[0].Entries[0].SharesDst)

	// check that our hook was called
	assert.Assert(t, *hookCalled)

	return completionTime
}

func doValidatorUnbonding(
	t *testing.T,
	stakingKeeper *stakingkeeper.Keeper,
	ctx context.Context,
	addrVal sdk.ValAddress,
	hookCalled *bool,
) (validator types.Validator) {
	t.Helper()
	validator, found := stakingKeeper.GetValidator(ctx, addrVal)
	assert.Assert(t, found)
	// Check that status is bonded
	assert.Equal(t, types.BondStatus(3), validator.Status)

	validator, err := stakingKeeper.BeginUnbondingValidator(ctx, validator)
	assert.NilError(t, err)

	// Check that status is unbonding
	assert.Equal(t, types.BondStatus(2), validator.Status)

	// check that our hook was called
	assert.Assert(t, *hookCalled)

	return validator
}

func TestValidatorUnbondingOnHold1(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	var (
		hookCalled bool
		ubdeID     uint64
	)

	_, _, addrVals := SetupUnbondingTests(t, f, &hookCalled, &ubdeID)

	// Start unbonding first validator
	validator := doValidatorUnbonding(t, f.stakingKeeper, f.ctx, addrVals[0], &hookCalled)

	completionTime := validator.UnbondingTime
	completionHeight := validator.UnbondingHeight

	// CONSUMER CHAIN'S UNBONDING PERIOD ENDS - STOPPED UNBONDING CAN NOW COMPLETE
	err := f.stakingKeeper.UnbondingCanComplete(f.ctx, ubdeID)
	assert.NilError(t, err)

	// Try to unbond validator
	assert.NilError(t, f.stakingKeeper.UnbondAllMatureValidators(f.ctx))

	// Check that validator unbonding is not complete (is not mature yet)
	validator, found := f.stakingKeeper.GetValidator(f.ctx, addrVals[0])
	assert.Assert(t, found)
	assert.Equal(t, types.Unbonding, validator.Status)
	unbondingVals, err := f.stakingKeeper.GetUnbondingValidators(f.ctx, completionTime, completionHeight)
	assert.NilError(t, err)
	assert.Equal(t, 1, len(unbondingVals))
	assert.Equal(t, validator.OperatorAddress, unbondingVals[0])

	// PROVIDER CHAIN'S UNBONDING PERIOD ENDS - BUT UNBONDING CANNOT COMPLETE
	f.ctx = integration.SetHeaderInfo(f.ctx, header.Info{
		Time:   completionTime.Add(time.Duration(1)),
		Height: completionHeight + 1,
	})
	assert.NilError(t, f.stakingKeeper.UnbondAllMatureValidators(f.ctx))

	// Check that validator unbonding is complete
	validator, found = f.stakingKeeper.GetValidator(f.ctx, addrVals[0])
	assert.Assert(t, found)
	assert.Equal(t, types.Unbonded, validator.Status)
	unbondingVals, err = f.stakingKeeper.GetUnbondingValidators(f.ctx, completionTime, completionHeight)
	assert.NilError(t, err)
	assert.Equal(t, 0, len(unbondingVals))
}

func TestValidatorUnbondingOnHold2(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	var (
		hookCalled bool
		ubdeID     uint64
		ubdeIDs    []uint64
	)

	_, _, addrVals := SetupUnbondingTests(t, f, &hookCalled, &ubdeID)

	// Start unbonding first validator
	validator1 := doValidatorUnbonding(t, f.stakingKeeper, f.ctx, addrVals[0], &hookCalled)
	ubdeIDs = append(ubdeIDs, ubdeID)

	// Reset hookCalled flag
	hookCalled = false

	// Start unbonding second validator
	validator2 := doValidatorUnbonding(t, f.stakingKeeper, f.ctx, addrVals[1], &hookCalled)
	ubdeIDs = append(ubdeIDs, ubdeID)

	// Check that there are two unbonding operations
	assert.Equal(t, 2, len(ubdeIDs))

	// Check that both validators have same unbonding time
	assert.Equal(t, validator1.UnbondingTime, validator2.UnbondingTime)

	completionTime := validator1.UnbondingTime
	completionHeight := validator1.UnbondingHeight

	// PROVIDER CHAIN'S UNBONDING PERIOD ENDS - BUT UNBONDING CANNOT COMPLETE
	f.ctx = integration.SetHeaderInfo(f.ctx, header.Info{
		Time:   completionTime.Add(time.Duration(1)),
		Height: completionHeight + 1,
	})
	assert.NilError(t, f.stakingKeeper.UnbondAllMatureValidators(f.ctx))

	// Check that unbonding is not complete for both validators
	validator1, found := f.stakingKeeper.GetValidator(f.ctx, addrVals[0])
	assert.Assert(t, found)
	assert.Equal(t, types.Unbonding, validator1.Status)
	validator2, found = f.stakingKeeper.GetValidator(f.ctx, addrVals[1])
	assert.Assert(t, found)
	assert.Equal(t, types.Unbonding, validator2.Status)
	unbondingVals, err := f.stakingKeeper.GetUnbondingValidators(f.ctx, completionTime, completionHeight)
	assert.NilError(t, err)
	assert.Equal(t, 2, len(unbondingVals))
	assert.Equal(t, validator1.OperatorAddress, unbondingVals[0])
	assert.Equal(t, validator2.OperatorAddress, unbondingVals[1])

	// CONSUMER CHAIN'S UNBONDING PERIOD ENDS - STOPPED UNBONDING CAN NOW COMPLETE
	err = f.stakingKeeper.UnbondingCanComplete(f.ctx, ubdeIDs[0])
	assert.NilError(t, err)

	// Try again to unbond validators
	assert.NilError(t, f.stakingKeeper.UnbondAllMatureValidators(f.ctx))

	// Check that unbonding is complete for validator1, but not for validator2
	validator1, found = f.stakingKeeper.GetValidator(f.ctx, addrVals[0])
	assert.Assert(t, found)
	assert.Equal(t, types.Unbonded, validator1.Status)
	validator2, found = f.stakingKeeper.GetValidator(f.ctx, addrVals[1])
	assert.Assert(t, found)
	assert.Equal(t, types.Unbonding, validator2.Status)
	unbondingVals, err = f.stakingKeeper.GetUnbondingValidators(f.ctx, completionTime, completionHeight)
	assert.NilError(t, err)
	assert.Equal(t, 1, len(unbondingVals))
	assert.Equal(t, validator2.OperatorAddress, unbondingVals[0])

	// Unbonding for validator2 can complete
	err = f.stakingKeeper.UnbondingCanComplete(f.ctx, ubdeIDs[1])
	assert.NilError(t, err)

	// Try again to unbond validators
	assert.NilError(t, f.stakingKeeper.UnbondAllMatureValidators(f.ctx))

	// Check that unbonding is complete for validator2
	validator2, found = f.stakingKeeper.GetValidator(f.ctx, addrVals[1])
	assert.Assert(t, found)
	assert.Equal(t, types.Unbonded, validator2.Status)
	unbondingVals, err = f.stakingKeeper.GetUnbondingValidators(f.ctx, completionTime, completionHeight)
	assert.NilError(t, err)
	assert.Equal(t, 0, len(unbondingVals))
}

func TestRedelegationOnHold1(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	var (
		hookCalled bool
		ubdeID     uint64
	)

	// _, app, ctx := createTestInput(t)
	_, addrDels, addrVals := SetupUnbondingTests(t, f, &hookCalled, &ubdeID)
	completionTime := doRedelegation(t, f.stakingKeeper, f.ctx, addrDels, addrVals, &hookCalled)

	// CONSUMER CHAIN'S UNBONDING PERIOD ENDS - BUT UNBONDING CANNOT COMPLETE
	err := f.stakingKeeper.UnbondingCanComplete(f.ctx, ubdeID)
	assert.NilError(t, err)

	// Redelegation is not complete - still exists
	redelegations, err := f.stakingKeeper.GetRedelegationsFromSrcValidator(f.ctx, addrVals[0])
	assert.NilError(t, err)
	assert.Equal(t, 1, len(redelegations))

	// PROVIDER CHAIN'S UNBONDING PERIOD ENDS - STOPPED UNBONDING CAN NOW COMPLETE
	f.ctx = integration.SetHeaderInfo(f.ctx, header.Info{Time: completionTime})
	_, err = f.stakingKeeper.CompleteRedelegation(f.ctx, addrDels[0], addrVals[0], addrVals[1])
	assert.NilError(t, err)

	// Redelegation is complete and record is gone
	redelegations, err = f.stakingKeeper.GetRedelegationsFromSrcValidator(f.ctx, addrVals[0])
	assert.NilError(t, err)
	assert.Equal(t, 0, len(redelegations))
}

func TestRedelegationOnHold2(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	var (
		hookCalled bool
		ubdeID     uint64
	)

	// _, app, ctx := createTestInput(t)
	_, addrDels, addrVals := SetupUnbondingTests(t, f, &hookCalled, &ubdeID)
	completionTime := doRedelegation(t, f.stakingKeeper, f.ctx, addrDels, addrVals, &hookCalled)

	// PROVIDER CHAIN'S UNBONDING PERIOD ENDS - BUT UNBONDING CANNOT COMPLETE
	f.ctx = integration.SetHeaderInfo(f.ctx, header.Info{Time: completionTime})
	_, err := f.stakingKeeper.CompleteRedelegation(f.ctx, addrDels[0], addrVals[0], addrVals[1])
	assert.NilError(t, err)

	// Redelegation is not complete - still exists
	redelegations, err := f.stakingKeeper.GetRedelegationsFromSrcValidator(f.ctx, addrVals[0])
	assert.NilError(t, err)
	assert.Equal(t, 1, len(redelegations))

	// CONSUMER CHAIN'S UNBONDING PERIOD ENDS - STOPPED UNBONDING CAN NOW COMPLETE
	err = f.stakingKeeper.UnbondingCanComplete(f.ctx, ubdeID)
	assert.NilError(t, err)

	// Redelegation is complete and record is gone
	redelegations, err = f.stakingKeeper.GetRedelegationsFromSrcValidator(f.ctx, addrVals[0])
	assert.NilError(t, err)
	assert.Equal(t, 0, len(redelegations))
}

func TestUnbondingDelegationOnHold1(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	var (
		hookCalled bool
		ubdeID     uint64
	)

	// _, app, ctx := createTestInput(t)
	bondDenom, addrDels, addrVals := SetupUnbondingTests(t, f, &hookCalled, &ubdeID)
	for _, addr := range addrDels {
		acc := f.accountKeeper.NewAccountWithAddress(f.ctx, addr)
		f.accountKeeper.SetAccount(f.ctx, acc)
	}
	completionTime, bondedAmt1, notBondedAmt1 := doUnbondingDelegation(t, f.stakingKeeper, f.bankKeeper, f.ctx, bondDenom, addrDels, addrVals, &hookCalled)

	// CONSUMER CHAIN'S UNBONDING PERIOD ENDS - BUT UNBONDING CANNOT COMPLETE
	err := f.stakingKeeper.UnbondingCanComplete(f.ctx, ubdeID)
	assert.NilError(t, err)

	bondedAmt3 := f.bankKeeper.GetBalance(f.ctx, f.stakingKeeper.GetBondedPool(f.ctx).GetAddress(), bondDenom).Amount
	notBondedAmt3 := f.bankKeeper.GetBalance(f.ctx, f.stakingKeeper.GetNotBondedPool(f.ctx).GetAddress(), bondDenom).Amount

	// Bonded and unbonded amounts are the same as before because the completionTime has not yet passed and so the
	// unbondingDelegation has not completed
	assert.Assert(math.IntEq(t, bondedAmt1, bondedAmt3))
	assert.Assert(math.IntEq(t, notBondedAmt1, notBondedAmt3))

	// PROVIDER CHAIN'S UNBONDING PERIOD ENDS - STOPPED UNBONDING CAN NOW COMPLETE
	f.ctx = integration.SetHeaderInfo(f.ctx, header.Info{Time: completionTime})
	_, err = f.stakingKeeper.CompleteUnbonding(f.ctx, addrDels[0], addrVals[0])
	assert.NilError(t, err)

	// Check that the unbonding was finally completed
	bondedAmt5 := f.bankKeeper.GetBalance(f.ctx, f.stakingKeeper.GetBondedPool(f.ctx).GetAddress(), bondDenom).Amount
	notBondedAmt5 := f.bankKeeper.GetBalance(f.ctx, f.stakingKeeper.GetNotBondedPool(f.ctx).GetAddress(), bondDenom).Amount

	assert.Assert(math.IntEq(t, bondedAmt1, bondedAmt5))
	// Not bonded amount back to what it was originally
	assert.Assert(math.IntEq(t, notBondedAmt1.SubRaw(1), notBondedAmt5))
}

func TestUnbondingDelegationOnHold2(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	var (
		hookCalled bool
		ubdeID     uint64
	)

	// _, app, ctx := createTestInput(t)
	bondDenom, addrDels, addrVals := SetupUnbondingTests(t, f, &hookCalled, &ubdeID)
	for _, addr := range addrDels {
		acc := f.accountKeeper.NewAccountWithAddress(f.ctx, addr)
		f.accountKeeper.SetAccount(f.ctx, acc)
	}
	completionTime, bondedAmt1, notBondedAmt1 := doUnbondingDelegation(t, f.stakingKeeper, f.bankKeeper, f.ctx, bondDenom, addrDels, addrVals, &hookCalled)

	// PROVIDER CHAIN'S UNBONDING PERIOD ENDS - BUT UNBONDING CANNOT COMPLETE
	f.ctx = integration.SetHeaderInfo(f.ctx, header.Info{Time: completionTime})
	_, err := f.stakingKeeper.CompleteUnbonding(f.ctx, addrDels[0], addrVals[0])
	assert.NilError(t, err)

	bondedAmt3 := f.bankKeeper.GetBalance(f.ctx, f.stakingKeeper.GetBondedPool(f.ctx).GetAddress(), bondDenom).Amount
	notBondedAmt3 := f.bankKeeper.GetBalance(f.ctx, f.stakingKeeper.GetNotBondedPool(f.ctx).GetAddress(), bondDenom).Amount

	// Bonded and unbonded amounts are the same as before because the completionTime has not yet passed and so the
	// unbondingDelegation has not completed
	assert.Assert(math.IntEq(t, bondedAmt1, bondedAmt3))
	assert.Assert(math.IntEq(t, notBondedAmt1, notBondedAmt3))

	// CONSUMER CHAIN'S UNBONDING PERIOD ENDS - STOPPED UNBONDING CAN NOW COMPLETE
	err = f.stakingKeeper.UnbondingCanComplete(f.ctx, ubdeID)
	assert.NilError(t, err)

	// Check that the unbonding was finally completed
	bondedAmt5 := f.bankKeeper.GetBalance(f.ctx, f.stakingKeeper.GetBondedPool(f.ctx).GetAddress(), bondDenom).Amount
	notBondedAmt5 := f.bankKeeper.GetBalance(f.ctx, f.stakingKeeper.GetNotBondedPool(f.ctx).GetAddress(), bondDenom).Amount

	assert.Assert(math.IntEq(t, bondedAmt1, bondedAmt5))
	// Not bonded amount back to what it was originally
	assert.Assert(math.IntEq(t, notBondedAmt1.SubRaw(1), notBondedAmt5))
}
