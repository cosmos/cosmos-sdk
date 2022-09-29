package keeper_test

import (
	"testing"
	"time"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

type StakingHooksTemplate struct{}

var _ types.StakingHooks = StakingHooksTemplate{}

func (h StakingHooksTemplate) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) error {
	return nil
}
func (h StakingHooksTemplate) BeforeValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) error {
	return nil
}
func (h StakingHooksTemplate) AfterValidatorRemoved(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h StakingHooksTemplate) AfterValidatorBonded(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h StakingHooksTemplate) AfterValidatorBeginUnbonding(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h StakingHooksTemplate) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h StakingHooksTemplate) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h StakingHooksTemplate) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h StakingHooksTemplate) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h StakingHooksTemplate) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) error {
	return nil
}
func (h StakingHooksTemplate) AfterUnbondingInitiated(ctx sdk.Context, id uint64) error {
	return nil
}

type MockStakingHooks struct {
	StakingHooksTemplate
	afterUnbondingInitiated func(uint64)
}

func (h MockStakingHooks) AfterUnbondingInitiated(ctx sdk.Context, id uint64) error {
	h.afterUnbondingInitiated(id)
	return nil
}

func (s *KeeperTestSuite) SetupUnbondingTests(t *testing.T, hookCalled *bool, ubdeID *uint64) (
	ctx sdk.Context, bondDenom string, addrDels []sdk.AccAddress, addrVals []sdk.ValAddress,
) {
	ctx, keeper := s.ctx, s.stakingKeeper
	// _, app, ctx = createTestInput()

	// stakingKeeper := keeper.NewKeeper(
	// 	app.AppCodec(),
	// 	app.GetKey(types.StoreKey),
	// 	app.AccountKeeper,
	// 	app.BankKeeper,
	// 	app.GetSubspace(types.ModuleName),
	// )

	myHooks := MockStakingHooks{
		afterUnbondingInitiated: func(id uint64) {
			*hookCalled = true
			// save id
			*ubdeID = id
			// call back to stop unbonding
			err := keeper.PutUnbondingOnHold(ctx, id)
			require.NoError(t, err)
		},
	}

	keeper.SetHooks(
		types.NewMultiStakingHooks(myHooks),
	)

	addrDels = simtestutil.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(10000))
	addrVals = simtestutil.ConvertAddrsToValAddrs(addrDels)

	valTokens := keeper.TokensFromConsensusPower(ctx, 10)
	startTokens := keeper.TokensFromConsensusPower(ctx, 20)

	bondDenom = keeper.BondDenom(ctx)
	notBondedPool := keeper.GetNotBondedPool(ctx)

	// require.NoError(t, banktestutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(bondDenom, startTokens))))
	s.bankKeeper.SendCoinsFromModuleToModule(ctx, types.NotBondedPoolName, types.BondedPoolName, startTokens)
	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	// Create a validator
	validator1 := teststaking.NewValidator(t, addrVals[0], PKs[0])
	validator1, issuedShares1 := validator1.AddTokensFromDel(valTokens)
	require.Equal(t, valTokens, issuedShares1.RoundInt())

	validator1 = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator1, true)
	require.True(sdk.IntEq(t, valTokens, validator1.BondedTokens()))
	require.True(t, validator1.IsBonded())

	// Create a delegator
	delegation := types.NewDelegation(addrDels[0], addrVals[0], issuedShares1)
	keeper.SetDelegation(ctx, delegation)

	// Create a validator to redelegate to
	validator2 := teststaking.NewValidator(t, addrVals[1], PKs[1])
	validator2, issuedShares2 := validator2.AddTokensFromDel(valTokens)
	require.Equal(t, valTokens, issuedShares2.RoundInt())

	validator2 = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator2, true)
	require.Equal(t, types.Bonded, validator2.Status)
	require.True(t, validator2.IsBonded())

	return
}

func doUnbondingDelegation(
	t *testing.T, app *simtestutil.SimApp, ctx sdk.Context, bondDenom string, addrDels []sdk.AccAddress, addrVals []sdk.ValAddress, hookCalled *bool,
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
	t *testing.T, app *simtestutil.SimApp, ctx sdk.Context, addrDels []sdk.AccAddress, addrVals []sdk.ValAddress, hookCalled *bool,
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

func doValidatorUnbonding(
	t *testing.T, app *simtestutil.SimApp, ctx sdk.Context, addrVal sdk.ValAddress, hookCalled *bool,
) (validator types.Validator) {
	validator, found := app.StakingKeeper.GetValidator(ctx, addrVal)
	require.True(t, found)
	// Check that status is bonded
	require.Equal(t, types.BondStatus(3), validator.Status)

	validator, err := app.StakingKeeper.BeginUnbondingValidator(ctx, validator)
	require.NoError(t, err)

	// Check that status is unbonding
	require.Equal(t, types.BondStatus(2), validator.Status)

	// check that our hook was called
	require.True(t, *hookCalled)

	return validator
}

func (s *KeeperTestSuite) TestValidatorUnbondingOnHold1(t *testing.T) {
	var hookCalled bool
	var ubdeID uint64

	app, ctx, _, _, addrVals := s.SetupUnbondingTests(t, &hookCalled, &ubdeID)

	// Start unbonding first validator
	validator := doValidatorUnbonding(t, app, ctx, addrVals[0], &hookCalled)

	completionTime := validator.UnbondingTime
	completionHeight := validator.UnbondingHeight

	// CONSUMER CHAIN'S UNBONDING PERIOD ENDS - STOPPED UNBONDING CAN NOW COMPLETE
	err := app.StakingKeeper.UnbondingCanComplete(ctx, ubdeID)
	require.NoError(t, err)

	// Try to unbond validator
	app.StakingKeeper.UnbondAllMatureValidators(ctx)

	// Check that validator unbonding is not complete (is not mature yet)
	validator, found := app.StakingKeeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.Equal(t, types.Unbonding, validator.Status)
	unbondingVals := app.StakingKeeper.GetUnbondingValidators(ctx, completionTime, completionHeight)
	require.Equal(t, 1, len(unbondingVals))
	require.Equal(t, validator.OperatorAddress, unbondingVals[0])

	// PROVIDER CHAIN'S UNBONDING PERIOD ENDS - BUT UNBONDING CANNOT COMPLETE
	ctx = ctx.WithBlockTime(completionTime.Add(time.Duration(1)))
	ctx = ctx.WithBlockHeight(completionHeight + 1)
	app.StakingKeeper.UnbondAllMatureValidators(ctx)

	// Check that validator unbonding is complete
	validator, found = app.StakingKeeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.Equal(t, types.Unbonded, validator.Status)
	unbondingVals = app.StakingKeeper.GetUnbondingValidators(ctx, completionTime, completionHeight)
	require.Equal(t, 0, len(unbondingVals))
}

// func TestValidatorUnbondingOnHold2(t *testing.T) {
// 	var hookCalled bool
// 	var ubdeID uint64
// 	var ubdeIDs []uint64
// 	app, ctx, _, _, addrVals := setup(t, &hookCalled, &ubdeID)

// 	// Start unbonding first validator
// 	validator1 := doValidatorUnbonding(t, app, ctx, addrVals[0], &hookCalled)
// 	ubdeIDs = append(ubdeIDs, ubdeID)

// 	// Reset hookCalled flag
// 	hookCalled = false

// 	// Start unbonding second validator
// 	validator2 := doValidatorUnbonding(t, app, ctx, addrVals[1], &hookCalled)
// 	ubdeIDs = append(ubdeIDs, ubdeID)

// 	// Check that there are two unbonding operations
// 	require.Equal(t, 2, len(ubdeIDs))

// 	// Check that both validators have same unbonding time
// 	require.Equal(t, validator1.UnbondingTime, validator2.UnbondingTime)

// 	completionTime := validator1.UnbondingTime
// 	completionHeight := validator1.UnbondingHeight

// 	// PROVIDER CHAIN'S UNBONDING PERIOD ENDS - BUT UNBONDING CANNOT COMPLETE
// 	ctx = ctx.WithBlockTime(completionTime.Add(time.Duration(1)))
// 	ctx = ctx.WithBlockHeight(completionHeight + 1)
// 	app.StakingKeeper.UnbondAllMatureValidators(ctx)

// 	// Check that unbonding is not complete for both validators
// 	validator1, found := app.StakingKeeper.GetValidator(ctx, addrVals[0])
// 	require.True(t, found)
// 	require.Equal(t, types.Unbonding, validator1.Status)
// 	validator2, found = app.StakingKeeper.GetValidator(ctx, addrVals[1])
// 	require.True(t, found)
// 	require.Equal(t, types.Unbonding, validator2.Status)
// 	unbondingVals := app.StakingKeeper.GetUnbondingValidators(ctx, completionTime, completionHeight)
// 	require.Equal(t, 2, len(unbondingVals))
// 	require.Equal(t, validator1.OperatorAddress, unbondingVals[0])
// 	require.Equal(t, validator2.OperatorAddress, unbondingVals[1])

// 	// CONSUMER CHAIN'S UNBONDING PERIOD ENDS - STOPPED UNBONDING CAN NOW COMPLETE
// 	err := app.StakingKeeper.UnbondingCanComplete(ctx, ubdeIDs[0])
// 	require.NoError(t, err)

// 	// Try again to unbond validators
// 	app.StakingKeeper.UnbondAllMatureValidators(ctx)

// 	// Check that unbonding is complete for validator1, but not for validator2
// 	validator1, found = app.StakingKeeper.GetValidator(ctx, addrVals[0])
// 	require.True(t, found)
// 	require.Equal(t, types.Unbonded, validator1.Status)
// 	validator2, found = app.StakingKeeper.GetValidator(ctx, addrVals[1])
// 	require.True(t, found)
// 	require.Equal(t, types.Unbonding, validator2.Status)
// 	unbondingVals = app.StakingKeeper.GetUnbondingValidators(ctx, completionTime, completionHeight)
// 	require.Equal(t, 1, len(unbondingVals))
// 	require.Equal(t, validator2.OperatorAddress, unbondingVals[0])

// 	// Unbonding for validator2 can complete
// 	err = app.StakingKeeper.UnbondingCanComplete(ctx, ubdeIDs[1])
// 	require.NoError(t, err)

// 	// Try again to unbond validators
// 	app.StakingKeeper.UnbondAllMatureValidators(ctx)

// 	// Check that unbonding is complete for validator2
// 	validator2, found = app.StakingKeeper.GetValidator(ctx, addrVals[1])
// 	require.True(t, found)
// 	require.Equal(t, types.Unbonded, validator2.Status)
// 	unbondingVals = app.StakingKeeper.GetUnbondingValidators(ctx, completionTime, completionHeight)
// 	require.Equal(t, 0, len(unbondingVals))
// }

// func TestRedelegationOnHold1(t *testing.T) {
// 	var hookCalled bool
// 	var ubdeID uint64
// 	app, ctx, _, addrDels, addrVals := setup(t, &hookCalled, &ubdeID)
// 	completionTime := doRedelegation(t, app, ctx, addrDels, addrVals, &hookCalled)

// 	// CONSUMER CHAIN'S UNBONDING PERIOD ENDS - BUT UNBONDING CANNOT COMPLETE
// 	err := app.StakingKeeper.UnbondingCanComplete(ctx, ubdeID)
// 	require.NoError(t, err)

// 	// Redelegation is not complete - still exists
// 	redelegations := app.StakingKeeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
// 	require.Equal(t, 1, len(redelegations))

// 	// PROVIDER CHAIN'S UNBONDING PERIOD ENDS - STOPPED UNBONDING CAN NOW COMPLETE
// 	ctx = ctx.WithBlockTime(completionTime)
// 	_, err = app.StakingKeeper.CompleteRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
// 	require.NoError(t, err)

// 	// Redelegation is complete and record is gone
// 	redelegations = app.StakingKeeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
// 	require.Equal(t, 0, len(redelegations))
// }

// func TestRedelegationOnHold2(t *testing.T) {
// 	var hookCalled bool
// 	var ubdeID uint64
// 	app, ctx, _, addrDels, addrVals := setup(t, &hookCalled, &ubdeID)
// 	completionTime := doRedelegation(t, app, ctx, addrDels, addrVals, &hookCalled)

// 	// PROVIDER CHAIN'S UNBONDING PERIOD ENDS - BUT UNBONDING CANNOT COMPLETE
// 	ctx = ctx.WithBlockTime(completionTime)
// 	_, err := app.StakingKeeper.CompleteRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
// 	require.NoError(t, err)

// 	// Redelegation is not complete - still exists
// 	redelegations := app.StakingKeeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
// 	require.Equal(t, 1, len(redelegations))

// 	// CONSUMER CHAIN'S UNBONDING PERIOD ENDS - STOPPED UNBONDING CAN NOW COMPLETE
// 	err = app.StakingKeeper.UnbondingCanComplete(ctx, ubdeID)
// 	require.NoError(t, err)

// 	// Redelegation is complete and record is gone
// 	redelegations = app.StakingKeeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
// 	require.Equal(t, 0, len(redelegations))
// }

// func TestUnbondingDelegationOnHold1(t *testing.T) {
// 	var hookCalled bool
// 	var ubdeID uint64
// 	app, ctx, bondDenom, addrDels, addrVals := setup(t, &hookCalled, &ubdeID)
// 	completionTime, bondedAmt1, notBondedAmt1 := doUnbondingDelegation(t, app, ctx, bondDenom, addrDels, addrVals, &hookCalled)

// 	// CONSUMER CHAIN'S UNBONDING PERIOD ENDS - BUT UNBONDING CANNOT COMPLETE
// 	err := app.StakingKeeper.UnbondingCanComplete(ctx, ubdeID)
// 	require.NoError(t, err)

// 	bondedAmt3 := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
// 	notBondedAmt3 := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

// 	// Bonded and unbonded amounts are the same as before because the completionTime has not yet passed and so the
// 	// unbondingDelegation has not completed
// 	require.True(sdk.IntEq(t, bondedAmt1, bondedAmt3))
// 	require.True(sdk.IntEq(t, notBondedAmt1, notBondedAmt3))

// 	// PROVIDER CHAIN'S UNBONDING PERIOD ENDS - STOPPED UNBONDING CAN NOW COMPLETE
// 	ctx = ctx.WithBlockTime(completionTime)
// 	_, err = app.StakingKeeper.CompleteUnbonding(ctx, addrDels[0], addrVals[0])
// 	require.NoError(t, err)

// 	// Check that the unbonding was finally completed
// 	bondedAmt5 := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
// 	notBondedAmt5 := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

// 	require.True(sdk.IntEq(t, bondedAmt1, bondedAmt5))
// 	// Not bonded amount back to what it was originaly
// 	require.True(sdk.IntEq(t, notBondedAmt1.SubRaw(1), notBondedAmt5))
// }

// func TestUnbondingDelegationOnHold2(t *testing.T) {
// 	var hookCalled bool
// 	var ubdeID uint64
// 	app, ctx, bondDenom, addrDels, addrVals := setup(t, &hookCalled, &ubdeID)
// 	completionTime, bondedAmt1, notBondedAmt1 := doUnbondingDelegation(t, app, ctx, bondDenom, addrDels, addrVals, &hookCalled)

// 	// PROVIDER CHAIN'S UNBONDING PERIOD ENDS - BUT UNBONDING CANNOT COMPLETE
// 	ctx = ctx.WithBlockTime(completionTime)
// 	_, err := app.StakingKeeper.CompleteUnbonding(ctx, addrDels[0], addrVals[0])
// 	require.NoError(t, err)

// 	bondedAmt3 := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
// 	notBondedAmt3 := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

// 	// Bonded and unbonded amounts are the same as before because the completionTime has not yet passed and so the
// 	// unbondingDelegation has not completed
// 	require.True(sdk.IntEq(t, bondedAmt1, bondedAmt3))
// 	require.True(sdk.IntEq(t, notBondedAmt1, notBondedAmt3))

// 	// CONSUMER CHAIN'S UNBONDING PERIOD ENDS - STOPPED UNBONDING CAN NOW COMPLETE
// 	err = app.StakingKeeper.UnbondingCanComplete(ctx, ubdeID)
// 	require.NoError(t, err)

// 	// Check that the unbonding was finally completed
// 	bondedAmt5 := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
// 	notBondedAmt5 := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

// 	require.True(sdk.IntEq(t, bondedAmt1, bondedAmt5))
// 	// Not bonded amount back to what it was originaly
// 	require.True(sdk.IntEq(t, notBondedAmt1.SubRaw(1), notBondedAmt5))
// }
