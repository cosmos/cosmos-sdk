package keeper_test

import (
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (s *KeeperTestSuite) applyValidatorSetUpdates(ctx sdk.Context, keeper *stakingkeeper.Keeper, expectedUpdatesLen int) []abci.ValidatorUpdate {
	updates, err := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	s.Require().NoError(err)
	if expectedUpdatesLen >= 0 {
		s.Require().Equal(expectedUpdatesLen, len(updates), "%v", updates)
	}
	return updates
}

func (s *KeeperTestSuite) TestValidator() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	valPubKey := PKs[0]
	valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
	valTokens := keeper.TokensFromConsensusPower(ctx, 10)

	// test how the validator is set from a purely unbonded pool
	validator := testutil.NewValidator(s.T(), valAddr, valPubKey)
	validator, _ = validator.AddTokensFromDel(valTokens)
	require.Equal(stakingtypes.Unbonded, validator.Status)
	require.Equal(valTokens, validator.Tokens)
	require.Equal(valTokens, validator.DelegatorShares.RoundInt())
	require.NoError(keeper.SetValidator(ctx, validator))
	require.NoError(keeper.SetValidatorByPowerIndex(ctx, validator))
	require.NoError(keeper.SetValidatorByConsAddr(ctx, validator))

	// ensure update
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	updates := s.applyValidatorSetUpdates(ctx, keeper, 1)
	validator, err := keeper.GetValidator(ctx, valAddr)
	require.NoError(err)
	require.Equal(validator.ABCIValidatorUpdate(keeper.PowerReduction(ctx)), updates[0])

	// after the save the validator should be bonded
	require.Equal(stakingtypes.Bonded, validator.Status)
	require.Equal(valTokens, validator.Tokens)
	require.Equal(valTokens, validator.DelegatorShares.RoundInt())

	// check each store for being saved
	consAddr, err := validator.GetConsAddr()
	require.NoError(err)
	resVal, err := keeper.GetValidatorByConsAddr(ctx, consAddr)
	require.NoError(err)
	require.True(validator.MinEqual(&resVal))

	resVals, err := keeper.GetLastValidators(ctx)
	require.NoError(err)
	require.Equal(1, len(resVals))
	require.True(validator.MinEqual(&resVals[0]))

	resVals, err = keeper.GetBondedValidatorsByPower(ctx)
	require.NoError(err)
	require.Equal(1, len(resVals))
	require.True(validator.MinEqual(&resVals[0]))

	allVals, err := keeper.GetAllValidators(ctx)
	require.NoError(err)
	require.Equal(1, len(allVals))

	// check the last validator power
	power := int64(100)
	require.NoError(keeper.SetLastValidatorPower(ctx, valAddr, power))
	resPower, err := keeper.GetLastValidatorPower(ctx, valAddr)
	require.NoError(err)
	require.Equal(power, resPower)
	require.NoError(keeper.DeleteLastValidatorPower(ctx, valAddr))
	resPower, err = keeper.GetLastValidatorPower(ctx, valAddr)
	require.NoError(err)
	require.Equal(int64(0), resPower)
}

// This function tests UpdateValidator, GetValidator, GetLastValidators, RemoveValidator
func (s *KeeperTestSuite) TestValidatorBasics() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	// construct the validators
	var validators [3]stakingtypes.Validator
	powers := []int64{9, 8, 7}
	for i, power := range powers {
		validators[i] = testutil.NewValidator(s.T(), sdk.ValAddress(PKs[i].Address().Bytes()), PKs[i])
		validators[i].Status = stakingtypes.Unbonded
		validators[i].Tokens = math.ZeroInt()
		tokens := keeper.TokensFromConsensusPower(ctx, power)

		validators[i], _ = validators[i].AddTokensFromDel(tokens)
	}

	require.Equal(keeper.TokensFromConsensusPower(ctx, 9), validators[0].Tokens)
	require.Equal(keeper.TokensFromConsensusPower(ctx, 8), validators[1].Tokens)
	require.Equal(keeper.TokensFromConsensusPower(ctx, 7), validators[2].Tokens)

	// check the empty keeper first
	_, err := keeper.GetValidator(ctx, sdk.ValAddress(PKs[0].Address().Bytes()))
	require.ErrorIs(err, stakingtypes.ErrNoValidatorFound)
	resVals, err := keeper.GetLastValidators(ctx)
	require.NoError(err)
	require.Zero(len(resVals))

	resVals, err = keeper.GetValidators(ctx, 2)
	require.NoError(err)
	require.Len(resVals, 0)

	// set and retrieve a record
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	validators[0] = stakingkeeper.TestingUpdateValidator(keeper, ctx, validators[0], true)
	require.NoError(keeper.SetValidatorByConsAddr(ctx, validators[0]))
	resVal, err := keeper.GetValidator(ctx, sdk.ValAddress(PKs[0].Address().Bytes()))
	require.NoError(err)
	require.True(validators[0].MinEqual(&resVal))

	// retrieve from consensus
	resVal, err = keeper.GetValidatorByConsAddr(ctx, sdk.ConsAddress(PKs[0].Address()))
	require.NoError(err)
	require.True(validators[0].MinEqual(&resVal))
	resVal, err = keeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(PKs[0]))
	require.NoError(err)
	require.True(validators[0].MinEqual(&resVal))

	resVals, err = keeper.GetLastValidators(ctx)
	require.NoError(err)
	require.Equal(1, len(resVals))
	require.True(validators[0].MinEqual(&resVals[0]))
	require.Equal(stakingtypes.Bonded, validators[0].Status)
	require.True(keeper.TokensFromConsensusPower(ctx, 9).Equal(validators[0].BondedTokens()))

	// modify a records, save, and retrieve
	validators[0].Status = stakingtypes.Bonded
	validators[0].Tokens = keeper.TokensFromConsensusPower(ctx, 10)
	validators[0].DelegatorShares = math.LegacyNewDecFromInt(validators[0].Tokens)
	validators[0] = stakingkeeper.TestingUpdateValidator(keeper, ctx, validators[0], true)
	resVal, err = keeper.GetValidator(ctx, sdk.ValAddress(PKs[0].Address().Bytes()))
	require.NoError(err)
	require.True(validators[0].MinEqual(&resVal))

	resVals, err = keeper.GetLastValidators(ctx)
	require.NoError(err)
	require.Equal(1, len(resVals))
	require.True(validators[0].MinEqual(&resVals[0]))

	// add other validators
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	validators[1] = stakingkeeper.TestingUpdateValidator(keeper, ctx, validators[1], true)
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	validators[2] = stakingkeeper.TestingUpdateValidator(keeper, ctx, validators[2], true)
	resVal, err = keeper.GetValidator(ctx, sdk.ValAddress(PKs[1].Address().Bytes()))
	require.NoError(err)
	require.True(validators[1].MinEqual(&resVal))
	resVal, err = keeper.GetValidator(ctx, sdk.ValAddress(PKs[2].Address().Bytes()))
	require.NoError(err)
	require.True(validators[2].MinEqual(&resVal))

	resVals, err = keeper.GetLastValidators(ctx)
	require.NoError(err)
	require.Equal(3, len(resVals))

	// remove a record

	bz, err := keeper.ValidatorAddressCodec().StringToBytes(validators[1].GetOperator())
	require.NoError(err)

	// shouldn't be able to remove if status is not unbonded
	require.EqualError(keeper.RemoveValidator(ctx, bz), "cannot call RemoveValidator on bonded or unbonding validators: failed to remove validator")

	// shouldn't be able to remove if there are still tokens left
	validators[1].Status = stakingtypes.Unbonded
	require.NoError(keeper.SetValidator(ctx, validators[1]))
	require.EqualError(keeper.RemoveValidator(ctx, bz), "attempting to remove a validator which still contains tokens: failed to remove validator")

	validators[1].Tokens = math.ZeroInt()                    // ...remove all tokens
	require.NoError(keeper.SetValidator(ctx, validators[1])) // ...set the validator
	require.NoError(keeper.RemoveValidator(ctx, bz))         // Now it can be removed.
	_, err = keeper.GetValidator(ctx, sdk.ValAddress(PKs[1].Address().Bytes()))
	require.ErrorIs(err, stakingtypes.ErrNoValidatorFound)
}

func (s *KeeperTestSuite) TestUpdateValidatorByPowerIndex() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	valPubKey := PKs[0]
	valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
	valTokens := keeper.TokensFromConsensusPower(ctx, 100)

	// add a validator
	validator := testutil.NewValidator(s.T(), valAddr, PKs[0])
	validator, delSharesCreated := validator.AddTokensFromDel(valTokens)
	require.Equal(stakingtypes.Unbonded, validator.Status)
	require.Equal(valTokens, validator.Tokens)

	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	validator, err := keeper.GetValidator(ctx, valAddr)
	require.NoError(err)
	require.Equal(valTokens, validator.Tokens)

	power := stakingtypes.GetValidatorsByPowerIndexKey(validator, keeper.PowerReduction(ctx), keeper.ValidatorAddressCodec())
	require.True(stakingkeeper.ValidatorByPowerIndexExists(ctx, keeper, power))

	// burn half the delegator shares
	require.NoError(keeper.DeleteValidatorByPowerIndex(ctx, validator))
	validator, burned := validator.RemoveDelShares(delSharesCreated.Quo(math.LegacyNewDec(2)))
	require.Equal(keeper.TokensFromConsensusPower(ctx, 50), burned)
	stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true) // update the validator, possibly kicking it out
	require.False(stakingkeeper.ValidatorByPowerIndexExists(ctx, keeper, power))

	validator, err = keeper.GetValidator(ctx, valAddr)
	require.NoError(err)

	power = stakingtypes.GetValidatorsByPowerIndexKey(validator, keeper.PowerReduction(ctx), keeper.ValidatorAddressCodec())
	require.True(stakingkeeper.ValidatorByPowerIndexExists(ctx, keeper, power))

	// set new validator by power index
	require.NoError(keeper.DeleteValidatorByPowerIndex(ctx, validator))
	require.False(stakingkeeper.ValidatorByPowerIndexExists(ctx, keeper, power))
	require.NoError(keeper.SetNewValidatorByPowerIndex(ctx, validator))
	require.True(stakingkeeper.ValidatorByPowerIndexExists(ctx, keeper, power))
}

func (s *KeeperTestSuite) TestApplyAndReturnValidatorSetUpdatesPowerDecrease() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	powers := []int64{100, 100}
	var validators [2]stakingtypes.Validator

	for i, power := range powers {
		validators[i] = testutil.NewValidator(s.T(), sdk.ValAddress(PKs[i].Address().Bytes()), PKs[i])
		tokens := keeper.TokensFromConsensusPower(ctx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)

	}

	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	validators[0] = stakingkeeper.TestingUpdateValidator(keeper, ctx, validators[0], false)
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	validators[1] = stakingkeeper.TestingUpdateValidator(keeper, ctx, validators[1], false)
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	s.applyValidatorSetUpdates(ctx, keeper, 2)

	// check initial power
	require.Equal(int64(100), validators[0].GetConsensusPower(keeper.PowerReduction(ctx)))
	require.Equal(int64(100), validators[1].GetConsensusPower(keeper.PowerReduction(ctx)))

	// test multiple value change
	// tendermintUpdate set: {c1, c3} -> {c1', c3'}
	delTokens1 := keeper.TokensFromConsensusPower(ctx, 20)
	delTokens2 := keeper.TokensFromConsensusPower(ctx, 30)
	validators[0], _ = validators[0].RemoveDelShares(math.LegacyNewDecFromInt(delTokens1))
	validators[1], _ = validators[1].RemoveDelShares(math.LegacyNewDecFromInt(delTokens2))
	validators[0] = stakingkeeper.TestingUpdateValidator(keeper, ctx, validators[0], false)
	validators[1] = stakingkeeper.TestingUpdateValidator(keeper, ctx, validators[1], false)

	// power has changed
	require.Equal(int64(80), validators[0].GetConsensusPower(keeper.PowerReduction(ctx)))
	require.Equal(int64(70), validators[1].GetConsensusPower(keeper.PowerReduction(ctx)))

	// CometBFT updates should reflect power change
	updates := s.applyValidatorSetUpdates(ctx, keeper, 2)
	require.Equal(validators[0].ABCIValidatorUpdate(keeper.PowerReduction(ctx)), updates[0])
	require.Equal(validators[1].ABCIValidatorUpdate(keeper.PowerReduction(ctx)), updates[1])
}

func (s *KeeperTestSuite) TestUpdateValidatorCommission() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	// Set MinCommissionRate to 0.05
	params, err := keeper.GetParams(ctx)
	require.NoError(err)
	params.MinCommissionRate = math.LegacyNewDecWithPrec(5, 2)
	require.NoError(keeper.SetParams(ctx, params))

	commission1 := stakingtypes.NewCommissionWithTime(
		math.LegacyNewDecWithPrec(1, 1), math.LegacyNewDecWithPrec(3, 1),
		math.LegacyNewDecWithPrec(1, 1), time.Now().UTC().Add(time.Duration(-1)*time.Hour),
	)
	commission2 := stakingtypes.NewCommission(math.LegacyNewDecWithPrec(1, 1), math.LegacyNewDecWithPrec(3, 1), math.LegacyNewDecWithPrec(1, 1))

	val1 := testutil.NewValidator(s.T(), sdk.ValAddress(PKs[0].Address().Bytes()), PKs[0])
	val2 := testutil.NewValidator(s.T(), sdk.ValAddress(PKs[1].Address().Bytes()), PKs[1])

	val1, _ = val1.SetInitialCommission(commission1)
	val2, _ = val2.SetInitialCommission(commission2)

	require.NoError(keeper.SetValidator(ctx, val1))
	require.NoError(keeper.SetValidator(ctx, val2))

	testCases := []struct {
		validator   stakingtypes.Validator
		newRate     math.LegacyDec
		expectedErr bool
	}{
		{val1, math.LegacyZeroDec(), true},
		{val2, math.LegacyNewDecWithPrec(-1, 1), true},
		{val2, math.LegacyNewDecWithPrec(4, 1), true},
		{val2, math.LegacyNewDecWithPrec(3, 1), true},
		{val2, math.LegacyNewDecWithPrec(1, 2), true},
		{val2, math.LegacyNewDecWithPrec(2, 1), false},
	}

	for i, tc := range testCases {
		commission, err := keeper.UpdateValidatorCommission(ctx, tc.validator, tc.newRate)

		if tc.expectedErr {
			require.Error(err, "expected error for test case #%d with rate: %s", i, tc.newRate)
		} else {
			require.NoError(err,
				"unexpected error for test case #%d with rate: %s", i, tc.newRate,
			)

			tc.validator.Commission = commission
			err = keeper.SetValidator(ctx, tc.validator)
			require.NoError(err)

			bz, err := keeper.ValidatorAddressCodec().StringToBytes(tc.validator.GetOperator())
			require.NoError(err)

			val, err := keeper.GetValidator(ctx, bz)
			require.NoError(err,
				"expected to find validator for test case #%d with rate: %s", i, tc.newRate,
			)

			require.Equal(tc.newRate, val.Commission.Rate,
				"expected new validator commission rate for test case #%d with rate: %s", i, tc.newRate,
			)
			require.Equal(ctx.BlockHeader().Time, val.Commission.UpdateTime,
				"expected new validator commission update time for test case #%d with rate: %s", i, tc.newRate,
			)
		}
	}
}

func (s *KeeperTestSuite) TestValidatorToken() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	valPubKey := PKs[0]
	valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
	addTokens := keeper.TokensFromConsensusPower(ctx, 10)
	delTokens := keeper.TokensFromConsensusPower(ctx, 5)

	validator := testutil.NewValidator(s.T(), valAddr, valPubKey)
	validator, _, err := keeper.AddValidatorTokensAndShares(ctx, validator, addTokens)
	require.NoError(err)
	require.Equal(addTokens, validator.Tokens)
	validator, _ = keeper.GetValidator(ctx, valAddr)
	require.Equal(math.LegacyNewDecFromInt(addTokens), validator.DelegatorShares)

	_, _, err = keeper.RemoveValidatorTokensAndShares(ctx, validator, math.LegacyNewDecFromInt(delTokens))
	require.NoError(err)
	validator, _ = keeper.GetValidator(ctx, valAddr)
	require.Equal(delTokens, validator.Tokens)
	require.True(validator.DelegatorShares.Equal(math.LegacyNewDecFromInt(delTokens)))

	_, err = keeper.RemoveValidatorTokens(ctx, validator, delTokens)
	require.NoError(err)
	validator, _ = keeper.GetValidator(ctx, valAddr)
	require.True(validator.Tokens.IsZero())
}

func (s *KeeperTestSuite) TestUnbondingValidator() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	valPubKey := PKs[0]
	valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
	validator := testutil.NewValidator(s.T(), valAddr, valPubKey)
	addTokens := keeper.TokensFromConsensusPower(ctx, 10)

	// set unbonding validator
	endTime := time.Now()
	endHeight := ctx.BlockHeight() + 10
	require.NoError(keeper.SetUnbondingValidatorsQueue(ctx, endTime, endHeight, []string{valAddr.String()}))

	resVals, err := keeper.GetUnbondingValidators(ctx, endTime, endHeight)
	require.NoError(err)
	require.Equal(1, len(resVals))
	require.Equal(valAddr.String(), resVals[0])

	// add another unbonding validator
	valAddr1 := sdk.ValAddress(PKs[1].Address().Bytes())
	validator1 := testutil.NewValidator(s.T(), valAddr1, PKs[1])
	validator1.UnbondingHeight = endHeight
	validator1.UnbondingTime = endTime
	require.NoError(keeper.InsertUnbondingValidatorQueue(ctx, validator1))

	resVals, err = keeper.GetUnbondingValidators(ctx, endTime, endHeight)
	require.NoError(err)
	require.Equal(2, len(resVals))

	// delete unbonding validator from the queue
	require.NoError(keeper.DeleteValidatorQueue(ctx, validator1))
	resVals, err = keeper.GetUnbondingValidators(ctx, endTime, endHeight)
	require.NoError(err)
	require.Equal(1, len(resVals))
	require.Equal(valAddr.String(), resVals[0])

	// check unbonding mature validators
	ctx = ctx.WithBlockHeight(endHeight).WithBlockTime(endTime)
	err = keeper.UnbondAllMatureValidators(ctx)
	require.EqualError(err, "validator in the unbonding queue was not found: validator does not exist")

	require.NoError(keeper.SetValidator(ctx, validator))
	ctx = ctx.WithBlockHeight(endHeight).WithBlockTime(endTime)

	err = keeper.UnbondAllMatureValidators(ctx)
	require.EqualError(err, "unexpected validator in unbonding queue; status was not unbonding")

	validator.Status = stakingtypes.Unbonding
	require.NoError(keeper.SetValidator(ctx, validator))
	require.NoError(keeper.UnbondAllMatureValidators(ctx))
	validator, err = keeper.GetValidator(ctx, valAddr)
	require.ErrorIs(err, stakingtypes.ErrNoValidatorFound)

	require.NoError(keeper.SetUnbondingValidatorsQueue(ctx, endTime, endHeight, []string{valAddr.String()}))
	validator = testutil.NewValidator(s.T(), valAddr, valPubKey)
	validator, _ = validator.AddTokensFromDel(addTokens)
	validator.Status = stakingtypes.Unbonding
	require.NoError(keeper.SetValidator(ctx, validator))
	require.NoError(keeper.UnbondAllMatureValidators(ctx))
	validator, err = keeper.GetValidator(ctx, valAddr)
	require.NoError(err)
	require.Equal(stakingtypes.Unbonded, validator.Status)
}

// TestUnbondAllMatureValidators_PendingSlotCleanup verifies that pending slots are properly
// cleaned up when validators are unbonded, ensuring consistency between queue keys and
// pending slot index.
func (s *KeeperTestSuite) TestUnbondAllMatureValidators_PendingSlotCleanup() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	// Create two different slots
	slot1Time := ctx.BlockTime().Add(time.Hour)
	slot1Height := ctx.BlockHeight() + 10

	slot2Time := ctx.BlockTime().Add(2 * time.Hour)
	slot2Height := ctx.BlockHeight() + 20

	// Create multiple validators with different unbonding times/heights
	valPubKey0 := PKs[0]
	valAddr0 := sdk.ValAddress(valPubKey0.Address().Bytes())
	validator0 := testutil.NewValidator(s.T(), valAddr0, valPubKey0)
	validator0, _ = validator0.AddTokensFromDel(keeper.TokensFromConsensusPower(ctx, 10))
	validator0.Status = stakingtypes.Unbonding
	validator0.UnbondingTime = slot1Time
	validator0.UnbondingHeight = slot1Height

	valPubKey1 := PKs[1]
	valAddr1 := sdk.ValAddress(valPubKey1.Address().Bytes())
	validator1 := testutil.NewValidator(s.T(), valAddr1, valPubKey1)
	validator1, _ = validator1.AddTokensFromDel(keeper.TokensFromConsensusPower(ctx, 10))
	validator1.Status = stakingtypes.Unbonding
	validator1.UnbondingTime = slot1Time
	validator1.UnbondingHeight = slot1Height

	valPubKey2 := PKs[2]
	valAddr2 := sdk.ValAddress(valPubKey2.Address().Bytes())
	validator2 := testutil.NewValidator(s.T(), valAddr2, valPubKey2)
	validator2, _ = validator2.AddTokensFromDel(keeper.TokensFromConsensusPower(ctx, 10))
	validator2.Status = stakingtypes.Unbonding
	validator2.UnbondingTime = slot2Time
	validator2.UnbondingHeight = slot2Height

	// Set up validators in the store
	require.NoError(keeper.SetValidator(ctx, validator0))
	require.NoError(keeper.SetValidator(ctx, validator1))
	require.NoError(keeper.SetValidator(ctx, validator2))

	// Add validators to different slots
	// Slot 1: validator0 and validator1
	require.NoError(keeper.SetUnbondingValidatorsQueue(ctx, slot1Time, slot1Height, []string{
		valAddr0.String(),
		valAddr1.String(),
	}))

	// Slot 2: validator2
	require.NoError(keeper.SetUnbondingValidatorsQueue(ctx, slot2Time, slot2Height, []string{
		valAddr2.String(),
	}))

	// Verify pending slots are populated
	slots, err := keeper.GetValidatorQueuePendingSlots(ctx)
	require.NoError(err)
	require.Len(slots, 2)

	// Verify both slots are in pending
	foundSlot1 := false
	foundSlot2 := false
	for _, slot := range slots {
		if slot.Time.Equal(slot1Time) && slot.Height == slot1Height {
			foundSlot1 = true
		}
		if slot.Time.Equal(slot2Time) && slot.Height == slot2Height {
			foundSlot2 = true
		}
	}
	require.True(foundSlot1, "slot1 should be in pending slots")
	require.True(foundSlot2, "slot2 should be in pending slots")

	// Advance time and height to make slot1 mature
	ctx = ctx.WithBlockTime(slot1Time).WithBlockHeight(slot1Height)

	// Unbond mature validators (slot1 should be processed)
	require.NoError(keeper.UnbondAllMatureValidators(ctx))

	// Verify slot1 validators are unbonded
	val0, err := keeper.GetValidator(ctx, valAddr0)
	require.NoError(err)
	require.Equal(stakingtypes.Unbonded, val0.Status)

	val1, err := keeper.GetValidator(ctx, valAddr1)
	require.NoError(err)
	require.Equal(stakingtypes.Unbonded, val1.Status)

	// Verify slot1 is removed from pending slots (since it became empty)
	slots, err = keeper.GetValidatorQueuePendingSlots(ctx)
	require.NoError(err)
	require.Len(slots, 1, "slot1 should be removed from pending after becoming empty")

	// Verify slot2 is still in pending (not mature yet)
	foundSlot2 = false
	for _, slot := range slots {
		if slot.Time.Equal(slot2Time) && slot.Height == slot2Height {
			foundSlot2 = true
		}
	}
	require.True(foundSlot2, "slot2 should still be in pending slots")

	// Verify slot1 queue key is deleted (GetUnbondingValidators should return empty)
	vals, err := keeper.GetUnbondingValidators(ctx, slot1Time, slot1Height)
	require.NoError(err)
	require.Empty(vals, "slot1 queue key should be deleted")

	// Advance to make slot2 mature
	ctx = ctx.WithBlockTime(slot2Time).WithBlockHeight(slot2Height)

	// Unbond mature validators (slot2 should be processed)
	require.NoError(keeper.UnbondAllMatureValidators(ctx))

	// Verify slot2 validator is unbonded
	val2, err := keeper.GetValidator(ctx, valAddr2)
	require.NoError(err)
	require.Equal(stakingtypes.Unbonded, val2.Status)

	// Verify slot2 is removed from pending slots
	slots, err = keeper.GetValidatorQueuePendingSlots(ctx)
	require.NoError(err)
	require.Empty(slots, "all slots should be removed from pending after processing")

	// Verify slot2 queue key is deleted (GetUnbondingValidators should return empty)
	vals, err = keeper.GetUnbondingValidators(ctx, slot2Time, slot2Height)
	require.NoError(err)
	require.Empty(vals, "slot2 queue key should be deleted")
}

// TestUnbondAllMatureValidators_PendingSlotCleanup_MultipleValidatorsInSlot verifies
// that when a slot has multiple validators, the slot is only removed from pending
// when all validators are unbonded.
func (s *KeeperTestSuite) TestUnbondAllMatureValidators_PendingSlotCleanup_MultipleValidatorsInSlot() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	slotTime := ctx.BlockTime().Add(time.Hour)
	slotHeight := ctx.BlockHeight() + 10

	valPubKey0 := PKs[0]
	valAddr0 := sdk.ValAddress(valPubKey0.Address().Bytes())
	validator0 := testutil.NewValidator(s.T(), valAddr0, valPubKey0)
	validator0, _ = validator0.AddTokensFromDel(keeper.TokensFromConsensusPower(ctx, 10))
	validator0.Status = stakingtypes.Unbonding
	validator0.UnbondingTime = slotTime
	validator0.UnbondingHeight = slotHeight

	valPubKey1 := PKs[1]
	valAddr1 := sdk.ValAddress(valPubKey1.Address().Bytes())
	validator1 := testutil.NewValidator(s.T(), valAddr1, valPubKey1)
	validator1, _ = validator1.AddTokensFromDel(keeper.TokensFromConsensusPower(ctx, 10))
	validator1.Status = stakingtypes.Unbonding
	validator1.UnbondingTime = slotTime
	validator1.UnbondingHeight = slotHeight

	require.NoError(keeper.SetValidator(ctx, validator0))
	require.NoError(keeper.SetValidator(ctx, validator1))

	// Add both validators to the same slot
	require.NoError(keeper.SetUnbondingValidatorsQueue(ctx, slotTime, slotHeight, []string{
		valAddr0.String(),
		valAddr1.String(),
	}))

	// Verify slot is in pending
	slots, err := keeper.GetValidatorQueuePendingSlots(ctx)
	require.NoError(err)
	require.Len(slots, 1)

	// Advance to make slot mature
	ctx = ctx.WithBlockTime(slotTime).WithBlockHeight(slotHeight)

	// Unbond mature validators - both should be processed
	require.NoError(keeper.UnbondAllMatureValidators(ctx))

	// Verify both validators are unbonded
	val0, err := keeper.GetValidator(ctx, valAddr0)
	require.NoError(err)
	require.Equal(stakingtypes.Unbonded, val0.Status)

	val1, err := keeper.GetValidator(ctx, valAddr1)
	require.NoError(err)
	require.Equal(stakingtypes.Unbonded, val1.Status)

	// Verify slot is removed from pending (all validators unbonded)
	slots, err = keeper.GetValidatorQueuePendingSlots(ctx)
	require.NoError(err)
	require.Empty(slots, "slot should be removed after all validators are unbonded")
}

// TestUnbondAllMatureValidators_PendingSlotCleanup_AlreadyDeletedSlot verifies
// that already-deleted slots are properly cleaned up from pending.
func (s *KeeperTestSuite) TestUnbondAllMatureValidators_PendingSlotCleanup_AlreadyDeletedSlot() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	slotTime := ctx.BlockTime().Add(time.Hour)
	slotHeight := ctx.BlockHeight() + 10

	// Manually add a slot to pending without creating the queue entry
	// This simulates a scenario where the queue key was deleted but pending slot wasn't updated
	require.NoError(keeper.AddValidatorQueuePendingSlot(ctx, slotTime, slotHeight))

	// Verify slot is in pending
	slots, err := keeper.GetValidatorQueuePendingSlots(ctx)
	require.NoError(err)
	require.Len(slots, 1)

	// Advance to make slot "mature"
	ctx = ctx.WithBlockTime(slotTime).WithBlockHeight(slotHeight)

	// UnbondAllMatureValidators should handle the missing queue key gracefully
	require.NoError(keeper.UnbondAllMatureValidators(ctx))

	// Verify the orphaned pending slot is cleaned up
	slots, err = keeper.GetValidatorQueuePendingSlots(ctx)
	require.NoError(err)
	require.Empty(slots, "orphaned pending slot should be cleaned up")
}
