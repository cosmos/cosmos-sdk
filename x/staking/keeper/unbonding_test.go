package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type MockStakingHooks struct {
	testutil.MockStakingHooks
	afterUnbondingInitiated func(uint64)
}

func (h MockStakingHooks) AfterUnbondingInitiated(ctx sdk.Context, id uint64) error {
	h.afterUnbondingInitiated(id)
	return nil
}

func (s *KeeperTestSuite) SetupUnbondingTests(t *testing.T, hookCalled *bool, ubdeID *uint64) (ctx sdk.Context, bondDenom string, addrDels []sdk.AccAddress, addrVals []sdk.ValAddress) {
	ctx = s.ctx

	testHooks := &MockStakingHooks{
		afterUnbondingInitiated: func(id uint64) {
			*hookCalled = true
			// save id
			*ubdeID = id
			// call back to stop unbonding
			err := s.stakingKeeper.PutUnbondingOnHold(ctx, id)
			require.NoError(t, err)
		},
	}

	s.stakingKeeper.SetHooks(types.NewMultiStakingHooks(testHooks))

	// TODO mocks
	// addrDels = simtestutil.AddTestAddrsIncremental(s.bankKeeper, s.stakingKeeper, ctx, 2, math.NewInt(10000))
	addrVals = simtestutil.ConvertAddrsToValAddrs(addrDels)

	valTokens := s.stakingKeeper.TokensFromConsensusPower(ctx, 10)
	startTokens := s.stakingKeeper.TokensFromConsensusPower(ctx, 20)

	bondDenom = s.stakingKeeper.BondDenom(ctx)
	notBondedPool := s.stakingKeeper.GetNotBondedPool(ctx)

	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), types.BondedPoolName, types.NotBondedPoolName, startTokens)
	s.accountKeeper.SetModuleAccount(ctx, notBondedPool)

	// Create a validator
	validator1 := testutil.NewValidator(t, addrVals[0], PKs[0])
	validator1, issuedShares1 := validator1.AddTokensFromDel(valTokens)
	require.Equal(t, valTokens, issuedShares1.RoundInt())

	validator1 = stakingkeeper.TestingUpdateValidator(s.stakingKeeper, ctx, validator1, true)
	require.True(math.IntEq(t, valTokens, validator1.BondedTokens()))
	require.True(t, validator1.IsBonded())

	// Create a delegator
	delegation := types.NewDelegation(addrDels[0], addrVals[0], issuedShares1)
	s.stakingKeeper.SetDelegation(ctx, delegation)

	// Create a validator to redelegate to
	validator2 := testutil.NewValidator(t, addrVals[1], PKs[1])
	validator2, issuedShares2 := validator2.AddTokensFromDel(valTokens)
	require.Equal(t, valTokens, issuedShares2.RoundInt())

	validator2 = stakingkeeper.TestingUpdateValidator(s.stakingKeeper, ctx, validator2, true)
	require.Equal(t, types.Bonded, validator2.Status)
	require.True(t, validator2.IsBonded())

	return
}

func doUnbondingDelegation(
	t *testing.T,
	stakingKeeper *stakingkeeper.Keeper,
	bankKeeper types.BankKeeper,
	ctx sdk.Context,
	bondDenom string,
	addrDels []sdk.AccAddress,
	addrVals []sdk.ValAddress,
	hookCalled *bool,
) (completionTime time.Time, bondedAmt math.Int, notBondedAmt math.Int) {
	// UNDELEGATE
	// Save original bonded and unbonded amounts
	// TODO mock
	bondedAmt1 := bankKeeper.GetBalance(ctx, stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	notBondedAmt1 := bankKeeper.GetBalance(ctx, stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	var err error
	completionTime, err = stakingKeeper.Undelegate(ctx, addrDels[0], addrVals[0], sdk.NewDec(1))
	require.NoError(t, err)

	// check that the unbonding actually happened
	bondedAmt2 := bankKeeper.GetBalance(ctx, stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	notBondedAmt2 := bankKeeper.GetBalance(ctx, stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	// Bonded amount is less
	require.True(math.IntEq(t, bondedAmt1.SubRaw(1), bondedAmt2))
	// Unbonded amount is more
	require.True(math.IntEq(t, notBondedAmt1.AddRaw(1), notBondedAmt2))

	// Check that the unbonding happened- we look up the entry and see that it has the correct number of shares
	unbondingDelegations := stakingKeeper.GetUnbondingDelegationsFromValidator(ctx, addrVals[0])
	require.Equal(t, math.NewInt(1), unbondingDelegations[0].Entries[0].Balance)

	// check that our hook was called
	require.True(t, *hookCalled)

	return completionTime, bondedAmt2, notBondedAmt2
}

func doRedelegation(
	t *testing.T,
	stakingKeeper *stakingkeeper.Keeper,
	ctx sdk.Context,
	addrDels []sdk.AccAddress,
	addrVals []sdk.ValAddress,
	hookCalled *bool,
) (completionTime time.Time) {
	var err error
	completionTime, err = stakingKeeper.BeginRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1], sdk.NewDec(1))
	require.NoError(t, err)

	// Check that the redelegation happened- we look up the entry and see that it has the correct number of shares
	redelegations := stakingKeeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
	require.Equal(t, 1, len(redelegations))
	require.Equal(t, sdk.NewDec(1), redelegations[0].Entries[0].SharesDst)

	// check that our hook was called
	require.True(t, *hookCalled)

	return completionTime
}

func doValidatorUnbonding(
	t *testing.T,
	stakingKeeper *stakingkeeper.Keeper,
	ctx sdk.Context,
	addrVal sdk.ValAddress,
	hookCalled *bool,
) (validator types.Validator) {
	validator, found := stakingKeeper.GetValidator(ctx, addrVal)
	require.True(t, found)
	// Check that status is bonded
	require.Equal(t, types.BondStatus(3), validator.Status)

	validator, err := stakingKeeper.BeginUnbondingValidator(ctx, validator)
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

	ctx, _, _, addrVals := s.SetupUnbondingTests(t, &hookCalled, &ubdeID)

	// Start unbonding first validator
	validator := doValidatorUnbonding(t, s.stakingKeeper, ctx, addrVals[0], &hookCalled)

	completionTime := validator.UnbondingTime
	completionHeight := validator.UnbondingHeight

	// CONSUMER CHAIN'S UNBONDING PERIOD ENDS - STOPPED UNBONDING CAN NOW COMPLETE
	err := s.stakingKeeper.UnbondingCanComplete(ctx, ubdeID)
	require.NoError(t, err)

	// Try to unbond validator
	s.stakingKeeper.UnbondAllMatureValidators(ctx)

	// Check that validator unbonding is not complete (is not mature yet)
	validator, found := s.stakingKeeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.Equal(t, types.Unbonding, validator.Status)
	unbondingVals := s.stakingKeeper.GetUnbondingValidators(ctx, completionTime, completionHeight)
	require.Equal(t, 1, len(unbondingVals))
	require.Equal(t, validator.OperatorAddress, unbondingVals[0])

	// PROVIDER CHAIN'S UNBONDING PERIOD ENDS - BUT UNBONDING CANNOT COMPLETE
	ctx = ctx.WithBlockTime(completionTime.Add(time.Duration(1)))
	ctx = ctx.WithBlockHeight(completionHeight + 1)
	s.stakingKeeper.UnbondAllMatureValidators(ctx)

	// Check that validator unbonding is complete
	validator, found = s.stakingKeeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.Equal(t, types.Unbonded, validator.Status)
	unbondingVals = s.stakingKeeper.GetUnbondingValidators(ctx, completionTime, completionHeight)
	require.Equal(t, 0, len(unbondingVals))
}

func (s *KeeperTestSuite) TestValidatorUnbondingOnHold2(t *testing.T) {
	var hookCalled bool
	var ubdeID uint64
	var ubdeIDs []uint64

	ctx, _, _, addrVals := s.SetupUnbondingTests(t, &hookCalled, &ubdeID)

	// Start unbonding first validator
	validator1 := doValidatorUnbonding(t, s.stakingKeeper, ctx, addrVals[0], &hookCalled)
	ubdeIDs = append(ubdeIDs, ubdeID)

	// Reset hookCalled flag
	hookCalled = false

	// Start unbonding second validator
	validator2 := doValidatorUnbonding(t, s.stakingKeeper, s.ctx, addrVals[1], &hookCalled)
	ubdeIDs = append(ubdeIDs, ubdeID)

	// Check that there are two unbonding operations
	require.Equal(t, 2, len(ubdeIDs))

	// Check that both validators have same unbonding time
	require.Equal(t, validator1.UnbondingTime, validator2.UnbondingTime)

	completionTime := validator1.UnbondingTime
	completionHeight := validator1.UnbondingHeight

	// PROVIDER CHAIN'S UNBONDING PERIOD ENDS - BUT UNBONDING CANNOT COMPLETE
	ctx = ctx.WithBlockTime(completionTime.Add(time.Duration(1)))
	ctx = ctx.WithBlockHeight(completionHeight + 1)
	s.stakingKeeper.UnbondAllMatureValidators(ctx)

	// Check that unbonding is not complete for both validators
	validator1, found := s.stakingKeeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.Equal(t, types.Unbonding, validator1.Status)
	validator2, found = s.stakingKeeper.GetValidator(ctx, addrVals[1])
	require.True(t, found)
	require.Equal(t, types.Unbonding, validator2.Status)
	unbondingVals := s.stakingKeeper.GetUnbondingValidators(ctx, completionTime, completionHeight)
	require.Equal(t, 2, len(unbondingVals))
	require.Equal(t, validator1.OperatorAddress, unbondingVals[0])
	require.Equal(t, validator2.OperatorAddress, unbondingVals[1])

	// CONSUMER CHAIN'S UNBONDING PERIOD ENDS - STOPPED UNBONDING CAN NOW COMPLETE
	err := s.stakingKeeper.UnbondingCanComplete(ctx, ubdeIDs[0])
	require.NoError(t, err)

	// Try again to unbond validators
	s.stakingKeeper.UnbondAllMatureValidators(ctx)

	// Check that unbonding is complete for validator1, but not for validator2
	validator1, found = s.stakingKeeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.Equal(t, types.Unbonded, validator1.Status)
	validator2, found = s.stakingKeeper.GetValidator(ctx, addrVals[1])
	require.True(t, found)
	require.Equal(t, types.Unbonding, validator2.Status)
	unbondingVals = s.stakingKeeper.GetUnbondingValidators(ctx, completionTime, completionHeight)
	require.Equal(t, 1, len(unbondingVals))
	require.Equal(t, validator2.OperatorAddress, unbondingVals[0])

	// Unbonding for validator2 can complete
	err = s.stakingKeeper.UnbondingCanComplete(ctx, ubdeIDs[1])
	require.NoError(t, err)

	// Try again to unbond validators
	s.stakingKeeper.UnbondAllMatureValidators(ctx)

	// Check that unbonding is complete for validator2
	validator2, found = s.stakingKeeper.GetValidator(ctx, addrVals[1])
	require.True(t, found)
	require.Equal(t, types.Unbonded, validator2.Status)
	unbondingVals = s.stakingKeeper.GetUnbondingValidators(ctx, completionTime, completionHeight)
	require.Equal(t, 0, len(unbondingVals))
}

func (s *KeeperTestSuite) TestRedelegationOnHold1(t *testing.T) {
	var hookCalled bool
	var ubdeID uint64

	ctx, _, addrDels, addrVals := s.SetupUnbondingTests(t, &hookCalled, &ubdeID)
	completionTime := doRedelegation(t, s.stakingKeeper, ctx, addrDels, addrVals, &hookCalled)

	// CONSUMER CHAIN'S UNBONDING PERIOD ENDS - BUT UNBONDING CANNOT COMPLETE
	err := s.stakingKeeper.UnbondingCanComplete(ctx, ubdeID)
	require.NoError(t, err)

	// Redelegation is not complete - still exists
	redelegations := s.stakingKeeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
	require.Equal(t, 1, len(redelegations))

	// PROVIDER CHAIN'S UNBONDING PERIOD ENDS - STOPPED UNBONDING CAN NOW COMPLETE
	ctx = ctx.WithBlockTime(completionTime)
	_, err = s.stakingKeeper.CompleteRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.NoError(t, err)

	// Redelegation is complete and record is gone
	redelegations = s.stakingKeeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
	require.Equal(t, 0, len(redelegations))
}

func (s *KeeperTestSuite) TestRedelegationOnHold2(t *testing.T) {
	var hookCalled bool
	var ubdeID uint64

	ctx, _, addrDels, addrVals := s.SetupUnbondingTests(t, &hookCalled, &ubdeID)
	completionTime := doRedelegation(t, s.stakingKeeper, ctx, addrDels, addrVals, &hookCalled)

	// PROVIDER CHAIN'S UNBONDING PERIOD ENDS - BUT UNBONDING CANNOT COMPLETE
	ctx = ctx.WithBlockTime(completionTime)
	_, err := s.stakingKeeper.CompleteRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.NoError(t, err)

	// Redelegation is not complete - still exists
	redelegations := s.stakingKeeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
	require.Equal(t, 1, len(redelegations))

	// CONSUMER CHAIN'S UNBONDING PERIOD ENDS - STOPPED UNBONDING CAN NOW COMPLETE
	err = s.stakingKeeper.UnbondingCanComplete(ctx, ubdeID)
	require.NoError(t, err)

	// Redelegation is complete and record is gone
	redelegations = s.stakingKeeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
	require.Equal(t, 0, len(redelegations))
}

func (s *KeeperTestSuite) TestUnbondingDelegationOnHold1(t *testing.T) {
	var hookCalled bool
	var ubdeID uint64

	ctx, bondDenom, addrDels, addrVals := s.SetupUnbondingTests(t, &hookCalled, &ubdeID)
	completionTime, bondedAmt1, notBondedAmt1 := doUnbondingDelegation(t, s.stakingKeeper, s.bankKeeper, ctx, bondDenom, addrDels, addrVals, &hookCalled)

	// CONSUMER CHAIN'S UNBONDING PERIOD ENDS - BUT UNBONDING CANNOT COMPLETE
	err := s.stakingKeeper.UnbondingCanComplete(ctx, ubdeID)
	require.NoError(t, err)

	// TODO mock
	bondedAmt3 := s.bankKeeper.GetBalance(ctx, s.stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	notBondedAmt3 := s.bankKeeper.GetBalance(ctx, s.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// Bonded and unbonded amounts are the same as before because the completionTime has not yet passed and so the
	// unbondingDelegation has not completed
	require.True(math.IntEq(t, bondedAmt1, bondedAmt3))
	require.True(math.IntEq(t, notBondedAmt1, notBondedAmt3))

	// PROVIDER CHAIN'S UNBONDING PERIOD ENDS - STOPPED UNBONDING CAN NOW COMPLETE
	ctx = ctx.WithBlockTime(completionTime)
	_, err = s.stakingKeeper.CompleteUnbonding(ctx, addrDels[0], addrVals[0])
	require.NoError(t, err)

	// TODO mock
	// Check that the unbonding was finally completed
	bondedAmt5 := s.bankKeeper.GetBalance(ctx, s.stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	notBondedAmt5 := s.bankKeeper.GetBalance(ctx, s.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	require.True(math.IntEq(t, bondedAmt1, bondedAmt5))
	// Not bonded amount back to what it was originaly
	require.True(math.IntEq(t, notBondedAmt1.SubRaw(1), notBondedAmt5))
}

func (s *KeeperTestSuite) TestUnbondingDelegationOnHold2(t *testing.T) {
	var hookCalled bool
	var ubdeID uint64

	ctx, bondDenom, addrDels, addrVals := s.SetupUnbondingTests(t, &hookCalled, &ubdeID)
	completionTime, bondedAmt1, notBondedAmt1 := doUnbondingDelegation(t, s.stakingKeeper, s.bankKeeper, ctx, bondDenom, addrDels, addrVals, &hookCalled)

	// PROVIDER CHAIN'S UNBONDING PERIOD ENDS - BUT UNBONDING CANNOT COMPLETE
	ctx = ctx.WithBlockTime(completionTime)
	_, err := s.stakingKeeper.CompleteUnbonding(ctx, addrDels[0], addrVals[0])
	require.NoError(t, err)

	// TODO mock
	bondedAmt3 := s.bankKeeper.GetBalance(ctx, s.stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	notBondedAmt3 := s.bankKeeper.GetBalance(ctx, s.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// Bonded and unbonded amounts are the same as before because the completionTime has not yet passed and so the
	// unbondingDelegation has not completed
	require.True(math.IntEq(t, bondedAmt1, bondedAmt3))
	require.True(math.IntEq(t, notBondedAmt1, notBondedAmt3))

	// CONSUMER CHAIN'S UNBONDING PERIOD ENDS - STOPPED UNBONDING CAN NOW COMPLETE
	err = s.stakingKeeper.UnbondingCanComplete(ctx, ubdeID)
	require.NoError(t, err)

	// TODO mock
	// Check that the unbonding was finally completed
	bondedAmt5 := s.bankKeeper.GetBalance(ctx, s.stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	notBondedAmt5 := s.bankKeeper.GetBalance(ctx, s.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	require.True(math.IntEq(t, bondedAmt1, bondedAmt5))
	// Not bonded amount back to what it was originaly
	require.True(math.IntEq(t, notBondedAmt1.SubRaw(1), notBondedAmt5))
}
