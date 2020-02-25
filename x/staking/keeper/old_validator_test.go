package keeper

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestFullValidatorSetPowerChange(t *testing.T) {
	ctx, _, _, keeper, _ := CreateTestInput(t, false, 1000)
	params := keeper.GetParams(ctx)
	max := 2
	params.MaxValidators = uint32(2)
	keeper.SetParams(ctx, params)

	// initialize some validators into the state
	powers := []int64{0, 100, 400, 400, 200}
	var validators [5]types.Validator
	for i, power := range powers {
		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{})
		tokens := sdk.TokensFromConsensusPower(power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)
		TestingUpdateValidator(keeper, ctx, validators[i], true)
	}
	for i := range powers {
		var found bool
		validators[i], found = keeper.GetValidator(ctx, validators[i].OperatorAddress)
		require.True(t, found)
	}
	assert.Equal(t, sdk.Unbonded, validators[0].Status)
	assert.Equal(t, sdk.Unbonding, validators[1].Status)
	assert.Equal(t, sdk.Bonded, validators[2].Status)
	assert.Equal(t, sdk.Bonded, validators[3].Status)
	assert.Equal(t, sdk.Unbonded, validators[4].Status)
	resValidators := keeper.GetBondedValidatorsByPower(ctx)
	assert.Equal(t, max, len(resValidators))
	assert.True(ValEq(t, validators[2], resValidators[0])) // in the order of txs
	assert.True(ValEq(t, validators[3], resValidators[1]))

	// test a swap in voting power

	tokens := sdk.TokensFromConsensusPower(600)
	validators[0], _ = validators[0].AddTokensFromDel(tokens)
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], true)
	resValidators = keeper.GetBondedValidatorsByPower(ctx)
	assert.Equal(t, max, len(resValidators))
	assert.True(ValEq(t, validators[0], resValidators[0]))
	assert.True(ValEq(t, validators[2], resValidators[1]))
}

func TestApplyAndReturnValidatorSetUpdatesAllNone(t *testing.T) {
	ctx, _, _, keeper, _ := CreateTestInput(t, false, 1000)

	powers := []int64{10, 20}
	var validators [2]types.Validator
	for i, power := range powers {
		valPubKey := PKs[i+1]
		valAddr := sdk.ValAddress(valPubKey.Address().Bytes())

		validators[i] = types.NewValidator(valAddr, valPubKey, types.Description{})
		tokens := sdk.TokensFromConsensusPower(power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)
	}

	// test from nothing to something
	//  tendermintUpdate set: {} -> {c1, c3}
	require.Equal(t, 0, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))
	keeper.SetValidator(ctx, validators[0])
	keeper.SetValidatorByPowerIndex(ctx, validators[0])
	keeper.SetValidator(ctx, validators[1])
	keeper.SetValidatorByPowerIndex(ctx, validators[1])

	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	assert.Equal(t, 2, len(updates))
	validators[0], _ = keeper.GetValidator(ctx, validators[0].OperatorAddress)
	validators[1], _ = keeper.GetValidator(ctx, validators[1].OperatorAddress)
	assert.Equal(t, validators[0].ABCIValidatorUpdate(), updates[1])
	assert.Equal(t, validators[1].ABCIValidatorUpdate(), updates[0])
}

func TestApplyAndReturnValidatorSetUpdatesIdentical(t *testing.T) {
	ctx, _, _, keeper, _ := CreateTestInput(t, false, 1000)

	powers := []int64{10, 20}
	var validators [2]types.Validator
	for i, power := range powers {
		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{})

		tokens := sdk.TokensFromConsensusPower(power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)

	}
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], false)
	require.Equal(t, 2, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	// test identical,
	//  tendermintUpdate set: {} -> {}
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], false)
	require.Equal(t, 0, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))
}

func TestApplyAndReturnValidatorSetUpdatesSingleValueChange(t *testing.T) {
	ctx, _, _, keeper, _ := CreateTestInput(t, false, 1000)

	powers := []int64{10, 20}
	var validators [2]types.Validator
	for i, power := range powers {

		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{})

		tokens := sdk.TokensFromConsensusPower(power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)

	}
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], false)
	require.Equal(t, 2, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	// test single value change
	//  tendermintUpdate set: {} -> {c1'}
	validators[0].Status = sdk.Bonded
	validators[0].Tokens = sdk.TokensFromConsensusPower(600)
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)

	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)

	require.Equal(t, 1, len(updates))
	require.Equal(t, validators[0].ABCIValidatorUpdate(), updates[0])
}

func TestApplyAndReturnValidatorSetUpdatesMultipleValueChange(t *testing.T) {
	ctx, _, _, keeper, _ := CreateTestInput(t, false, 1000)

	powers := []int64{10, 20}
	var validators [2]types.Validator
	for i, power := range powers {

		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{})

		tokens := sdk.TokensFromConsensusPower(power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)

	}
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], false)
	require.Equal(t, 2, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	// test multiple value change
	//  tendermintUpdate set: {c1, c3} -> {c1', c3'}
	delTokens1 := sdk.TokensFromConsensusPower(190)
	delTokens2 := sdk.TokensFromConsensusPower(80)
	validators[0], _ = validators[0].AddTokensFromDel(delTokens1)
	validators[1], _ = validators[1].AddTokensFromDel(delTokens2)
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], false)

	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 2, len(updates))
	require.Equal(t, validators[0].ABCIValidatorUpdate(), updates[0])
	require.Equal(t, validators[1].ABCIValidatorUpdate(), updates[1])
}

func TestApplyAndReturnValidatorSetUpdatesInserted(t *testing.T) {
	ctx, _, _, keeper, _ := CreateTestInput(t, false, 1000)

	powers := []int64{10, 20, 5, 15, 25}
	var validators [5]types.Validator
	for i, power := range powers {

		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{})

		tokens := sdk.TokensFromConsensusPower(power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)

	}

	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], false)
	require.Equal(t, 2, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	// test validtor added at the beginning
	//  tendermintUpdate set: {} -> {c0}
	keeper.SetValidator(ctx, validators[2])
	keeper.SetValidatorByPowerIndex(ctx, validators[2])
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	validators[2], _ = keeper.GetValidator(ctx, validators[2].OperatorAddress)
	require.Equal(t, 1, len(updates))
	require.Equal(t, validators[2].ABCIValidatorUpdate(), updates[0])

	// test validtor added at the beginning
	//  tendermintUpdate set: {} -> {c0}
	keeper.SetValidator(ctx, validators[3])
	keeper.SetValidatorByPowerIndex(ctx, validators[3])
	updates = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	validators[3], _ = keeper.GetValidator(ctx, validators[3].OperatorAddress)
	require.Equal(t, 1, len(updates))
	require.Equal(t, validators[3].ABCIValidatorUpdate(), updates[0])

	// test validtor added at the end
	//  tendermintUpdate set: {} -> {c0}
	keeper.SetValidator(ctx, validators[4])
	keeper.SetValidatorByPowerIndex(ctx, validators[4])
	updates = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	validators[4], _ = keeper.GetValidator(ctx, validators[4].OperatorAddress)
	require.Equal(t, 1, len(updates))
	require.Equal(t, validators[4].ABCIValidatorUpdate(), updates[0])
}

func TestApplyAndReturnValidatorSetUpdatesWithCliffValidator(t *testing.T) {
	ctx, _, _, keeper, _ := CreateTestInput(t, false, 1000)
	params := types.DefaultParams()
	params.MaxValidators = 2
	keeper.SetParams(ctx, params)

	powers := []int64{10, 20, 5}
	var validators [5]types.Validator
	for i, power := range powers {

		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{})

		tokens := sdk.TokensFromConsensusPower(power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)

	}
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], false)
	require.Equal(t, 2, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	// test validator added at the end but not inserted in the valset
	//  tendermintUpdate set: {} -> {}
	TestingUpdateValidator(keeper, ctx, validators[2], false)
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 0, len(updates))

	// test validator change its power and become a gotValidator (pushing out an existing)
	//  tendermintUpdate set: {}     -> {c0, c4}
	require.Equal(t, 0, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	tokens := sdk.TokensFromConsensusPower(10)
	validators[2], _ = validators[2].AddTokensFromDel(tokens)
	keeper.SetValidator(ctx, validators[2])
	keeper.SetValidatorByPowerIndex(ctx, validators[2])
	updates = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	validators[2], _ = keeper.GetValidator(ctx, validators[2].OperatorAddress)
	require.Equal(t, 2, len(updates), "%v", updates)
	require.Equal(t, validators[0].ABCIValidatorUpdateZero(), updates[1])
	require.Equal(t, validators[2].ABCIValidatorUpdate(), updates[0])
}

func TestApplyAndReturnValidatorSetUpdatesPowerDecrease(t *testing.T) {
	ctx, _, _, keeper, _ := CreateTestInput(t, false, 1000)

	powers := []int64{100, 100}
	var validators [2]types.Validator
	for i, power := range powers {

		validators[i] = types.NewValidator(sdk.ValAddress(Addrs[i]), PKs[i], types.Description{})

		tokens := sdk.TokensFromConsensusPower(power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)

	}
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], false)
	require.Equal(t, 2, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	// check initial power
	require.Equal(t, int64(100), validators[0].GetConsensusPower())
	require.Equal(t, int64(100), validators[1].GetConsensusPower())

	// test multiple value change
	//  tendermintUpdate set: {c1, c3} -> {c1', c3'}
	delTokens1 := sdk.TokensFromConsensusPower(20)
	delTokens2 := sdk.TokensFromConsensusPower(30)
	validators[0], _ = validators[0].RemoveDelShares(delTokens1.ToDec())
	validators[1], _ = validators[1].RemoveDelShares(delTokens2.ToDec())
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], false)

	// power has changed
	require.Equal(t, int64(80), validators[0].GetConsensusPower())
	require.Equal(t, int64(70), validators[1].GetConsensusPower())

	// Tendermint updates should reflect power change
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 2, len(updates))
	require.Equal(t, validators[0].ABCIValidatorUpdate(), updates[0])
	require.Equal(t, validators[1].ABCIValidatorUpdate(), updates[1])
}

func TestApplyAndReturnValidatorSetUpdatesNewValidator(t *testing.T) {
	ctx, _, _, keeper, _ := CreateTestInput(t, false, 1000)
	params := keeper.GetParams(ctx)
	params.MaxValidators = uint32(3)

	keeper.SetParams(ctx, params)

	powers := []int64{100, 100}
	var validators [2]types.Validator

	// initialize some validators into the state
	for i, power := range powers {

		valPubKey := PKs[i+1]
		valAddr := sdk.ValAddress(valPubKey.Address().Bytes())

		validators[i] = types.NewValidator(valAddr, valPubKey, types.Description{})
		tokens := sdk.TokensFromConsensusPower(power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)

		keeper.SetValidator(ctx, validators[i])
		keeper.SetValidatorByPowerIndex(ctx, validators[i])
	}

	// verify initial Tendermint updates are correct
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, len(validators), len(updates))
	validators[0], _ = keeper.GetValidator(ctx, validators[0].OperatorAddress)
	validators[1], _ = keeper.GetValidator(ctx, validators[1].OperatorAddress)
	require.Equal(t, validators[0].ABCIValidatorUpdate(), updates[0])
	require.Equal(t, validators[1].ABCIValidatorUpdate(), updates[1])

	require.Equal(t, 0, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	// update initial validator set
	for i, power := range powers {

		keeper.DeleteValidatorByPowerIndex(ctx, validators[i])
		tokens := sdk.TokensFromConsensusPower(power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)

		keeper.SetValidator(ctx, validators[i])
		keeper.SetValidatorByPowerIndex(ctx, validators[i])
	}

	// add a new validator that goes from zero power, to non-zero power, back to
	// zero power
	valPubKey := PKs[len(validators)+1]
	valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
	amt := sdk.NewInt(100)

	validator := types.NewValidator(valAddr, valPubKey, types.Description{})
	validator, _ = validator.AddTokensFromDel(amt)

	keeper.SetValidator(ctx, validator)

	validator, _ = validator.RemoveDelShares(amt.ToDec())
	keeper.SetValidator(ctx, validator)
	keeper.SetValidatorByPowerIndex(ctx, validator)

	// add a new validator that increases in power
	valPubKey = PKs[len(validators)+2]
	valAddr = sdk.ValAddress(valPubKey.Address().Bytes())

	validator = types.NewValidator(valAddr, valPubKey, types.Description{})
	tokens := sdk.TokensFromConsensusPower(500)
	validator, _ = validator.AddTokensFromDel(tokens)
	keeper.SetValidator(ctx, validator)
	keeper.SetValidatorByPowerIndex(ctx, validator)

	// verify initial Tendermint updates are correct
	updates = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	validator, _ = keeper.GetValidator(ctx, validator.OperatorAddress)
	validators[0], _ = keeper.GetValidator(ctx, validators[0].OperatorAddress)
	validators[1], _ = keeper.GetValidator(ctx, validators[1].OperatorAddress)
	require.Equal(t, len(validators)+1, len(updates))
	require.Equal(t, validator.ABCIValidatorUpdate(), updates[0])
	require.Equal(t, validators[0].ABCIValidatorUpdate(), updates[1])
	require.Equal(t, validators[1].ABCIValidatorUpdate(), updates[2])
}

func TestApplyAndReturnValidatorSetUpdatesBondTransition(t *testing.T) {
	ctx, _, _, keeper, _ := CreateTestInput(t, false, 1000)
	params := keeper.GetParams(ctx)
	params.MaxValidators = uint32(2)

	keeper.SetParams(ctx, params)

	powers := []int64{100, 200, 300}
	var validators [3]types.Validator

	// initialize some validators into the state
	for i, power := range powers {
		moniker := fmt.Sprintf("%d", i)
		valPubKey := PKs[i+1]
		valAddr := sdk.ValAddress(valPubKey.Address().Bytes())

		validators[i] = types.NewValidator(valAddr, valPubKey, types.Description{Moniker: moniker})
		tokens := sdk.TokensFromConsensusPower(power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)
		keeper.SetValidator(ctx, validators[i])
		keeper.SetValidatorByPowerIndex(ctx, validators[i])
	}

	// verify initial Tendermint updates are correct
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 2, len(updates))
	validators[2], _ = keeper.GetValidator(ctx, validators[2].OperatorAddress)
	validators[1], _ = keeper.GetValidator(ctx, validators[1].OperatorAddress)
	require.Equal(t, validators[2].ABCIValidatorUpdate(), updates[0])
	require.Equal(t, validators[1].ABCIValidatorUpdate(), updates[1])

	require.Equal(t, 0, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	// delegate to validator with lowest power but not enough to bond
	ctx = ctx.WithBlockHeight(1)

	var found bool
	validators[0], found = keeper.GetValidator(ctx, validators[0].OperatorAddress)
	require.True(t, found)

	keeper.DeleteValidatorByPowerIndex(ctx, validators[0])
	tokens := sdk.TokensFromConsensusPower(1)
	validators[0], _ = validators[0].AddTokensFromDel(tokens)
	keeper.SetValidator(ctx, validators[0])
	keeper.SetValidatorByPowerIndex(ctx, validators[0])

	// verify initial Tendermint updates are correct
	require.Equal(t, 0, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))

	// create a series of events that will bond and unbond the validator with
	// lowest power in a single block context (height)
	ctx = ctx.WithBlockHeight(2)

	validators[1], found = keeper.GetValidator(ctx, validators[1].OperatorAddress)
	require.True(t, found)

	keeper.DeleteValidatorByPowerIndex(ctx, validators[0])
	validators[0], _ = validators[0].RemoveDelShares(validators[0].DelegatorShares)
	keeper.SetValidator(ctx, validators[0])
	keeper.SetValidatorByPowerIndex(ctx, validators[0])
	updates = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 0, len(updates))

	keeper.DeleteValidatorByPowerIndex(ctx, validators[1])
	tokens = sdk.TokensFromConsensusPower(250)
	validators[1], _ = validators[1].AddTokensFromDel(tokens)
	keeper.SetValidator(ctx, validators[1])
	keeper.SetValidatorByPowerIndex(ctx, validators[1])

	// verify initial Tendermint updates are correct
	updates = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(updates))
	require.Equal(t, validators[1].ABCIValidatorUpdate(), updates[0])

	require.Equal(t, 0, len(keeper.ApplyAndReturnValidatorSetUpdates(ctx)))
}

func TestUpdateValidatorCommission(t *testing.T) {
	ctx, _, _, keeper, _ := CreateTestInput(t, false, 1000)
	ctx = ctx.WithBlockHeader(abci.Header{Time: time.Now().UTC()})

	commission1 := types.NewCommissionWithTime(
		sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(3, 1),
		sdk.NewDecWithPrec(1, 1), time.Now().UTC().Add(time.Duration(-1)*time.Hour),
	)
	commission2 := types.NewCommission(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(3, 1), sdk.NewDecWithPrec(1, 1))

	val1 := types.NewValidator(addrVals[0], PKs[0], types.Description{})
	val2 := types.NewValidator(addrVals[1], PKs[1], types.Description{})

	val1, _ = val1.SetInitialCommission(commission1)
	val2, _ = val2.SetInitialCommission(commission2)

	keeper.SetValidator(ctx, val1)
	keeper.SetValidator(ctx, val2)

	testCases := []struct {
		validator   types.Validator
		newRate     sdk.Dec
		expectedErr bool
	}{
		{val1, sdk.ZeroDec(), true},
		{val2, sdk.NewDecWithPrec(-1, 1), true},
		{val2, sdk.NewDecWithPrec(4, 1), true},
		{val2, sdk.NewDecWithPrec(3, 1), true},
		{val2, sdk.NewDecWithPrec(2, 1), false},
	}

	for i, tc := range testCases {
		commission, err := keeper.UpdateValidatorCommission(ctx, tc.validator, tc.newRate)

		if tc.expectedErr {
			require.Error(t, err, "expected error for test case #%d with rate: %s", i, tc.newRate)
		} else {
			tc.validator.Commission = commission
			keeper.SetValidator(ctx, tc.validator)
			val, found := keeper.GetValidator(ctx, tc.validator.OperatorAddress)

			require.True(t, found,
				"expected to find validator for test case #%d with rate: %s", i, tc.newRate,
			)
			require.NoError(t, err,
				"unexpected error for test case #%d with rate: %s", i, tc.newRate,
			)
			require.Equal(t, tc.newRate, val.Commission.Rate,
				"expected new validator commission rate for test case #%d with rate: %s", i, tc.newRate,
			)
			require.Equal(t, ctx.BlockHeader().Time, val.Commission.UpdateTime,
				"expected new validator commission update time for test case #%d with rate: %s", i, tc.newRate,
			)
		}
	}
}
