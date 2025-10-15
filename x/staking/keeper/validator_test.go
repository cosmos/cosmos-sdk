package keeper_test

import (
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttime "github.com/cometbft/cometbft/types/time"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
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

func (s *KeeperTestSuite) TestGetAllPendingUnbondingValidators() {
	testCases := []struct {
		name         string
		maxCacheSize int
		description  string
	}{
		{
			name:         "cache size < 0 (cache disabled)",
			maxCacheSize: -1,
			description:  "should always read from store when cache is not initialized",
		},
		{
			name:         "cache size = 0 (unlimited cache)",
			maxCacheSize: 0,
			description:  "should use unlimited cache with no size restrictions",
		},
		{
			name:         "cache size > unbonding queue entries",
			maxCacheSize: 10,
			description:  "should use cache when cache is large enough",
		},
		{
			name:         "cache size == unbonding queue entries",
			maxCacheSize: 2,
			description:  "should use cache when cache size matches entries",
		},
		{
			name:         "cache size < unbonding queue entries",
			maxCacheSize: 1,
			description:  "should fallback to store when cache size is exceeded",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			key := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
			storeService := runtime.NewKVStoreService(key)
			testCtx := sdktestutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
			ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
			encCfg := moduletestutil.MakeTestEncodingConfig()

			ctrl := gomock.NewController(s.T())
			accountKeeper := testutil.NewMockAccountKeeper(ctrl)
			accountKeeper.EXPECT().GetModuleAddress(stakingtypes.BondedPoolName).Return(bondedAcc.GetAddress()).AnyTimes()
			accountKeeper.EXPECT().GetModuleAddress(stakingtypes.NotBondedPoolName).Return(notBondedAcc.GetAddress()).AnyTimes()
			accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

			bankKeeper := testutil.NewMockBankKeeper(ctrl)

			keeper := stakingkeeper.NewKeeper(
				encCfg.Codec,
				storeService,
				accountKeeper,
				bankKeeper,
				authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				address.NewBech32Codec("cosmosvaloper"),
				address.NewBech32Codec("cosmosvalcons"),
				tc.maxCacheSize,
			)
			s.Require().NoError(keeper.SetParams(ctx, stakingtypes.DefaultParams()))

			blockTime := time.Now().UTC()
			blockHeight := int64(1000)
			ctx = ctx.WithBlockHeight(blockHeight).WithBlockTime(blockTime)

			// add ready to unbond validator
			valPubKey := PKs[0]
			valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
			val := testutil.NewValidator(s.T(), valAddr, valPubKey)
			val.UnbondingHeight = blockHeight
			val.UnbondingTime = blockTime
			val.Status = stakingtypes.Unbonding
			s.Require().NoError(keeper.SetValidator(ctx, val))
			s.Require().NoError(keeper.InsertUnbondingValidatorQueue(ctx, val))

			// add another unbonding validator
			valAddr1 := sdk.ValAddress(PKs[1].Address().Bytes())
			validator1 := testutil.NewValidator(s.T(), valAddr1, PKs[1])
			valUnbondingHeight1 := blockHeight - 10
			valUnbondingTime1 := blockTime.Add(-1 * time.Minute)
			validator1.UnbondingHeight = valUnbondingHeight1
			validator1.UnbondingTime = valUnbondingTime1
			validator1.Status = stakingtypes.Unbonding
			s.Require().NoError(keeper.SetValidator(ctx, validator1))
			s.Require().NoError(keeper.InsertUnbondingValidatorQueue(ctx, validator1))

			// get pending unbonding validators should return the inserted validators
			unbondingValidators, err := keeper.GetPendingUnbondingValidators(ctx, val.UnbondingTime, val.UnbondingHeight)
			s.Require().NoError(err)
			s.Require().Equal(2, len(unbondingValidators))
			s.Require().Equal(val.GetOperator(), unbondingValidators[stakingtypes.GetCacheValidatorQueueKey(val.UnbondingTime, val.UnbondingHeight)][0])
			s.Require().Equal(validator1.GetOperator(), unbondingValidators[stakingtypes.GetCacheValidatorQueueKey(validator1.UnbondingTime, validator1.UnbondingHeight)][0])

			// Test calling again to verify cache consistency
			unbondingValidators2, err := keeper.GetPendingUnbondingValidators(ctx, val.UnbondingTime, val.UnbondingHeight)
			s.Require().NoError(err)
			s.Require().Equal(len(unbondingValidators), len(unbondingValidators2), "repeated call should return same number of entries")
		})
	}
}

func (s *KeeperTestSuite) TestInsertUnbondingValidatorQueue() {
	testCases := []struct {
		name         string
		maxCacheSize int
		description  string
	}{
		{
			name:         "cache size < 0 (cache disabled)",
			maxCacheSize: -1,
			description:  "should always write to store when cache is not initialized",
		},
		{
			name:         "cache size = 0 (unlimited cache)",
			maxCacheSize: 0,
			description:  "should use unlimited cache with no size restrictions",
		},
		{
			name:         "cache size > unbonding queue entries",
			maxCacheSize: 10,
			description:  "should use cache when cache is large enough",
		},
		{
			name:         "cache size == unbonding queue entries",
			maxCacheSize: 2,
			description:  "should use cache when cache size matches entries",
		},
		{
			name:         "cache size < unbonding queue entries",
			maxCacheSize: 1,
			description:  "should fallback to store when cache size is exceeded",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			key := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
			storeService := runtime.NewKVStoreService(key)
			testCtx := sdktestutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
			ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
			encCfg := moduletestutil.MakeTestEncodingConfig()

			ctrl := gomock.NewController(s.T())
			accountKeeper := testutil.NewMockAccountKeeper(ctrl)
			accountKeeper.EXPECT().GetModuleAddress(stakingtypes.BondedPoolName).Return(bondedAcc.GetAddress()).AnyTimes()
			accountKeeper.EXPECT().GetModuleAddress(stakingtypes.NotBondedPoolName).Return(notBondedAcc.GetAddress()).AnyTimes()
			accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

			bankKeeper := testutil.NewMockBankKeeper(ctrl)

			// Initialize keeper with specific cache size
			keeper := stakingkeeper.NewKeeper(
				encCfg.Codec,
				storeService,
				accountKeeper,
				bankKeeper,
				authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				address.NewBech32Codec("cosmosvaloper"),
				address.NewBech32Codec("cosmosvalcons"),
				tc.maxCacheSize,
			)
			s.Require().NoError(keeper.SetParams(ctx, stakingtypes.DefaultParams()))

			blockTime := time.Now().UTC()
			blockHeight := int64(1000)
			ctx = ctx.WithBlockHeight(blockHeight).WithBlockTime(blockTime)

			iterator, err := keeper.ValidatorQueueIterator(ctx, blockTime, blockHeight)
			s.Require().NoError(err)
			defer iterator.Close()
			count := 0
			for ; iterator.Valid(); iterator.Next() {
				count++
			}
			// no unbonding validator in the queue initially
			s.Require().Equal(0, count)

			// add ready to unbond validator
			valPubKey := PKs[0]
			valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
			val := testutil.NewValidator(s.T(), valAddr, valPubKey)
			val.UnbondingHeight = blockHeight
			val.UnbondingTime = blockTime
			val.Status = stakingtypes.Unbonding
			s.Require().NoError(keeper.SetValidator(ctx, val))
			s.Require().NoError(keeper.InsertUnbondingValidatorQueue(ctx, val))

			// add another unbonding validator with same unbonding time and height
			valAddr1 := sdk.ValAddress(PKs[1].Address().Bytes())
			validator1 := testutil.NewValidator(s.T(), valAddr1, PKs[1])
			valUnbondingHeight1 := blockHeight
			valUnbondingTime1 := blockTime
			validator1.UnbondingHeight = valUnbondingHeight1
			validator1.UnbondingTime = valUnbondingTime1
			validator1.Status = stakingtypes.Unbonding
			s.Require().NoError(keeper.SetValidator(ctx, validator1))
			s.Require().NoError(keeper.InsertUnbondingValidatorQueue(ctx, validator1))

			iterator1, err := keeper.ValidatorQueueIterator(ctx, blockTime, blockHeight)
			s.Require().NoError(err)
			defer iterator1.Close()
			count1 := 0
			for ; iterator1.Valid(); iterator1.Next() {
				count1++
			}

			// unbonding validator should be retrieved
			// count 1 due to same unbonding time and height
			s.Require().Equal(1, count1)

			// Verify GetUnbondingValidators returns the correct validators after insertion
			unbondingVals, err := keeper.GetUnbondingValidators(ctx, blockTime, blockHeight)
			s.Require().NoError(err)
			s.Require().Equal(2, len(unbondingVals), "should have 2 validators at same time/height")
			s.Require().Contains(unbondingVals, val.OperatorAddress)
			s.Require().Contains(unbondingVals, validator1.OperatorAddress)

			// add another unbonding validator with different unbonding time and height
			valAddr2 := sdk.ValAddress(PKs[2].Address().Bytes())
			validator2 := testutil.NewValidator(s.T(), valAddr2, PKs[2])
			valUnbondingHeight2 := blockHeight - 10
			valUnbondingTime2 := blockTime.Add(-1 * time.Minute)
			validator2.UnbondingHeight = valUnbondingHeight2
			validator2.UnbondingTime = valUnbondingTime2
			validator2.Status = stakingtypes.Unbonding
			s.Require().NoError(keeper.SetValidator(ctx, validator2))
			s.Require().NoError(keeper.InsertUnbondingValidatorQueue(ctx, validator2))

			iterator2, err := keeper.ValidatorQueueIterator(ctx, blockTime, blockHeight)
			s.Require().NoError(err)
			defer iterator2.Close()
			count2 := 0
			for ; iterator2.Valid(); iterator2.Next() {
				count2++
			}

			// unbonding validator should be retrieved
			s.Require().Equal(2, count2)

			// Verify the new validator was inserted at the correct time/height
			unbondingVals2, err := keeper.GetUnbondingValidators(ctx, validator2.UnbondingTime, validator2.UnbondingHeight)
			s.Require().NoError(err)
			s.Require().Equal(1, len(unbondingVals2), "should have 1 validator at different time/height")
			s.Require().Contains(unbondingVals2, validator2.OperatorAddress)
		})
	}
}

func (s *KeeperTestSuite) TestGetUnbondingValidators() {
	testCases := []struct {
		name         string
		maxCacheSize int
		description  string
	}{
		{
			name:         "cache size < 0 (cache disabled)",
			maxCacheSize: -1,
			description:  "should always read from store when cache is not initialized",
		},
		{
			name:         "cache size = 0 (unlimited cache)",
			maxCacheSize: 0,
			description:  "should use unlimited cache with no size restrictions",
		},
		{
			name:         "cache size > unbonding queue entries",
			maxCacheSize: 10,
			description:  "should use cache when cache is large enough",
		},
		{
			name:         "cache size == unbonding queue entries",
			maxCacheSize: 3,
			description:  "should use cache when cache size matches entries",
		},
		{
			name:         "cache size < unbonding queue entries",
			maxCacheSize: 1,
			description:  "should fallback to store when cache size is exceeded",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			key := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
			storeService := runtime.NewKVStoreService(key)
			testCtx := sdktestutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
			ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
			encCfg := moduletestutil.MakeTestEncodingConfig()

			ctrl := gomock.NewController(s.T())
			accountKeeper := testutil.NewMockAccountKeeper(ctrl)
			accountKeeper.EXPECT().GetModuleAddress(stakingtypes.BondedPoolName).Return(bondedAcc.GetAddress()).AnyTimes()
			accountKeeper.EXPECT().GetModuleAddress(stakingtypes.NotBondedPoolName).Return(notBondedAcc.GetAddress()).AnyTimes()
			accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

			bankKeeper := testutil.NewMockBankKeeper(ctrl)

			keeper := stakingkeeper.NewKeeper(
				encCfg.Codec,
				storeService,
				accountKeeper,
				bankKeeper,
				authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				address.NewBech32Codec("cosmosvaloper"),
				address.NewBech32Codec("cosmosvalcons"),
				tc.maxCacheSize,
			)
			s.Require().NoError(keeper.SetParams(ctx, stakingtypes.DefaultParams()))

			baseTime := time.Now().UTC()
			baseHeight := int64(1000)
			ctx = ctx.WithBlockHeight(baseHeight).WithBlockTime(baseTime)

			// Create validators unbonding at different times/heights
			// Group 1: Two validators at baseTime, baseHeight
			val1 := testutil.NewValidator(s.T(), sdk.ValAddress(PKs[0].Address().Bytes()), PKs[0])
			val1.UnbondingHeight = baseHeight
			val1.UnbondingTime = baseTime
			val1.Status = stakingtypes.Unbonding
			s.Require().NoError(keeper.SetValidator(ctx, val1))
			s.Require().NoError(keeper.InsertUnbondingValidatorQueue(ctx, val1))

			val2 := testutil.NewValidator(s.T(), sdk.ValAddress(PKs[1].Address().Bytes()), PKs[1])
			val2.UnbondingHeight = baseHeight
			val2.UnbondingTime = baseTime
			val2.Status = stakingtypes.Unbonding
			s.Require().NoError(keeper.SetValidator(ctx, val2))
			s.Require().NoError(keeper.InsertUnbondingValidatorQueue(ctx, val2))

			// Group 2: One validator at different time/height
			val3 := testutil.NewValidator(s.T(), sdk.ValAddress(PKs[2].Address().Bytes()), PKs[2])
			val3.UnbondingHeight = baseHeight + 10
			val3.UnbondingTime = baseTime.Add(1 * time.Hour)
			val3.Status = stakingtypes.Unbonding
			s.Require().NoError(keeper.SetValidator(ctx, val3))
			s.Require().NoError(keeper.InsertUnbondingValidatorQueue(ctx, val3))

			// Group 3: One validator at another different time/height
			val4 := testutil.NewValidator(s.T(), sdk.ValAddress(PKs[3].Address().Bytes()), PKs[3])
			val4.UnbondingHeight = baseHeight - 5
			val4.UnbondingTime = baseTime.Add(-30 * time.Minute)
			val4.Status = stakingtypes.Unbonding
			s.Require().NoError(keeper.SetValidator(ctx, val4))
			s.Require().NoError(keeper.InsertUnbondingValidatorQueue(ctx, val4))

			// Test 1: Get validators for group 1 (baseTime, baseHeight)
			// Should return 2 validators
			unbondingVals1, err := keeper.GetUnbondingValidators(ctx, baseTime, baseHeight)
			s.Require().NoError(err)
			s.Require().Equal(2, len(unbondingVals1), "should have 2 validators at baseTime/baseHeight")

			// Verify the correct validators are returned
			s.Require().Contains(unbondingVals1, val1.OperatorAddress)
			s.Require().Contains(unbondingVals1, val2.OperatorAddress)

			// Test 2: Get validators for group 2 (baseTime + 1 hour, baseHeight + 10)
			// Should return 1 validator
			unbondingVals2, err := keeper.GetUnbondingValidators(ctx, val3.UnbondingTime, val3.UnbondingHeight)
			s.Require().NoError(err)
			s.Require().Equal(1, len(unbondingVals2), "should have 1 validator at baseTime+1hour/baseHeight+10")
			s.Require().Contains(unbondingVals2, val3.OperatorAddress)

			// Test 3: Get validators for group 3 (baseTime - 30 min, baseHeight - 5)
			// Should return 1 validator
			unbondingVals3, err := keeper.GetUnbondingValidators(ctx, val4.UnbondingTime, val4.UnbondingHeight)
			s.Require().NoError(err)
			s.Require().Equal(1, len(unbondingVals3), "should have 1 validator at baseTime-30min/baseHeight-5")
			s.Require().Contains(unbondingVals3, val4.OperatorAddress)

			// Test 4: Get validators for a time/height with no validators
			// Should return empty slice
			emptyTime := baseTime.Add(2 * time.Hour)
			emptyHeight := baseHeight + 100
			unbondingValsEmpty, err := keeper.GetUnbondingValidators(ctx, emptyTime, emptyHeight)
			s.Require().NoError(err)
			s.Require().Equal(0, len(unbondingValsEmpty), "should have 0 validators at non-existent time/height")

			// Test 5: Call GetUnbondingValidators again to verify cache consistency
			// This ensures second call returns same results (cache hit scenario)
			unbondingVals1Again, err := keeper.GetUnbondingValidators(ctx, baseTime, baseHeight)
			s.Require().NoError(err)
			s.Require().Equal(len(unbondingVals1), len(unbondingVals1Again), "repeated call should return same number of validators")
			s.Require().ElementsMatch(unbondingVals1, unbondingVals1Again, "repeated call should return same validators")

			// Verify total validator count
			allValidators, err := keeper.GetAllValidators(ctx)
			s.Require().NoError(err)
			s.Require().Equal(4, len(allValidators), "should have 4 total validators")
		})
	}
}

func (s *KeeperTestSuite) TestUnbondAllMatureValidators() {
	testCases := []struct {
		name                   string
		maxCacheSize           int
		numUnbondingValidators int
	}{
		{
			name:                   "cache size < 0 (cache disabled)",
			maxCacheSize:           -1,
			numUnbondingValidators: 3,
		},
		{
			name:                   "cache size = 0 (unlimited cache)",
			maxCacheSize:           0,
			numUnbondingValidators: 3,
		},
		{
			name:                   "cache size > unbonding validators",
			maxCacheSize:           3,
			numUnbondingValidators: 2,
		},
		{
			name:                   "cache size == unbonding validators",
			maxCacheSize:           2,
			numUnbondingValidators: 2,
		},
		{
			name:                   "cache size < unbonding validators",
			maxCacheSize:           1,
			numUnbondingValidators: 2,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			key := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
			storeService := runtime.NewKVStoreService(key)
			testCtx := sdktestutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
			ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
			encCfg := moduletestutil.MakeTestEncodingConfig()

			ctrl := gomock.NewController(s.T())
			accountKeeper := testutil.NewMockAccountKeeper(ctrl)
			accountKeeper.EXPECT().GetModuleAddress(stakingtypes.BondedPoolName).Return(bondedAcc.GetAddress()).AnyTimes()
			accountKeeper.EXPECT().GetModuleAddress(stakingtypes.NotBondedPoolName).Return(notBondedAcc.GetAddress()).AnyTimes()
			accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

			bankKeeper := testutil.NewMockBankKeeper(ctrl)

			// Initialize keeper with specific cache size
			keeper := stakingkeeper.NewKeeper(
				encCfg.Codec,
				storeService,
				accountKeeper,
				bankKeeper,
				authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				address.NewBech32Codec("cosmosvaloper"),
				address.NewBech32Codec("cosmosvalcons"),
				tc.maxCacheSize,
			)
			s.Require().NoError(keeper.SetParams(ctx, stakingtypes.DefaultParams()))

			blockTime := time.Now().UTC()
			blockHeight := int64(1000)
			ctx = ctx.WithBlockHeight(blockHeight).WithBlockTime(blockTime)

			// Create multiple unbonding validators that are ready to unbond
			for i := 0; i < tc.numUnbondingValidators; i++ {
				valPubKey := PKs[i]
				valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
				val := testutil.NewValidator(s.T(), valAddr, valPubKey)
				val.UnbondingHeight = blockHeight
				val.UnbondingTime = blockTime
				val.Status = stakingtypes.Unbonding
				s.Require().NoError(keeper.SetValidator(ctx, val))
				s.Require().NoError(keeper.InsertUnbondingValidatorQueue(ctx, val))
			}

			// Verify we have the expected number of validators before unbonding
			allValidators, err := keeper.GetAllValidators(ctx)
			s.Require().NoError(err)
			s.Require().Equal(tc.numUnbondingValidators, len(allValidators), "should have all validators before unbonding")

			// Verify GetUnbondingValidators returns the expected number of validators.
			// In this case, it should return all validators as all validators are unbonding at the same height and time.
			unbondingValidators, err := keeper.GetUnbondingValidators(ctx, blockTime, blockHeight)
			s.Require().NoError(err)
			s.Require().Equal(tc.numUnbondingValidators, len(unbondingValidators))

			err = keeper.UnbondAllMatureValidators(ctx)
			s.Require().NoError(err)

			// Verify all validators were unbonded (removed from the store)
			allValidatorsAfter, err := keeper.GetAllValidators(ctx)
			s.Require().NoError(err)
			s.Require().Equal(0, len(allValidatorsAfter))
		})
	}
}

func (s *KeeperTestSuite) TestUnbondingValidatorQueueCacheRecovery() {
	// This test verifies that when the cache is initially too small (exceeded),
	// and then entries are dequeued, the cache can recover and be used again
	// Cache size is based on the number of unique time+height keys, not individual validators
	key := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := sdktestutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	ctrl := gomock.NewController(s.T())
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)
	accountKeeper.EXPECT().GetModuleAddress(stakingtypes.BondedPoolName).Return(bondedAcc.GetAddress()).AnyTimes()
	accountKeeper.EXPECT().GetModuleAddress(stakingtypes.NotBondedPoolName).Return(notBondedAcc.GetAddress()).AnyTimes()
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// Initialize keeper with small cache size (2 time+height keys)
	maxCacheSize := 2
	keeper := stakingkeeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		address.NewBech32Codec("cosmosvaloper"),
		address.NewBech32Codec("cosmosvalcons"),
		maxCacheSize,
	)
	params := stakingtypes.DefaultParams()
	params.UnbondingTime = 1 * time.Hour // Use 1 hour so we can create different timestamps
	s.Require().NoError(keeper.SetParams(ctx, params))

	baseTime := time.Now().UTC()
	baseHeight := int64(1000)

	// Create validators at 5 different time+height combinations (exceeds cache size of 2)
	numKeys := 5
	validators := make([]stakingtypes.Validator, numKeys)
	unbondingTimes := make([]time.Time, numKeys)
	unbondingHeights := make([]int64, numKeys)

	for i := 0; i < numKeys; i++ {
		// Create unique time+height combinations
		unbondingTimes[i] = baseTime.Add(time.Duration(i) * time.Hour)
		unbondingHeights[i] = baseHeight + int64(i*10)

		valPubKey := PKs[i]
		valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
		validators[i] = testutil.NewValidator(s.T(), valAddr, valPubKey)
		validators[i].UnbondingHeight = unbondingHeights[i]
		validators[i].UnbondingTime = unbondingTimes[i]
		validators[i].Status = stakingtypes.Unbonding
		s.Require().NoError(keeper.SetValidator(ctx, validators[i]))
		s.Require().NoError(keeper.InsertUnbondingValidatorQueue(ctx, validators[i]))
	}

	// Verify all validators were created
	for i := 0; i < numKeys; i++ {
		_, err := keeper.GetValidator(ctx, sdk.ValAddress(PKs[i].Address().Bytes()))
		s.Require().NoError(err)
	}

	// At this point, cache should be exceeded (5 keys > maxCacheSize of 2)
	// GetPendingUnbondingValidators should still work, but will read from store
	ctx = ctx.WithBlockTime(unbondingTimes[numKeys-1]).WithBlockHeight(unbondingHeights[numKeys-1])
	allValidators, err := keeper.GetPendingUnbondingValidators(ctx, unbondingTimes[numKeys-1], unbondingHeights[numKeys-1])
	s.Require().NoError(err)
	s.Require().Equal(5, len(allValidators), "should have 5 different time+height keys")

	// Mature and unbond the first 3 validators (removing 3 keys)
	// Fast-forward to time when first 3 validators are mature
	ctx = ctx.WithBlockTime(unbondingTimes[2]).WithBlockHeight(unbondingHeights[2])
	err = keeper.UnbondAllMatureValidators(ctx)
	s.Require().NoError(err)

	// Verify the first 3 are now unbonded (not in unbonding queue)
	for i := 0; i < 3; i++ {
		// Verify not in queue
		vals, err := keeper.GetUnbondingValidators(ctx, unbondingTimes[i], unbondingHeights[i])
		s.Require().NoError(err)
		s.Require().Equal(0, len(vals), "should have no validators at this time+height")
	}

	// Verify the last 2 are still unbonding
	for i := 3; i < numKeys; i++ {
		valAddr := sdk.ValAddress(PKs[i].Address().Bytes())
		validator, err := keeper.GetValidator(ctx, valAddr)
		s.Require().NoError(err)
		s.Require().Equal(stakingtypes.Unbonding, validator.Status, "validator should still be unbonding")

		// Verify still in queue
		vals, err := keeper.GetUnbondingValidators(ctx, unbondingTimes[i], unbondingHeights[i])
		s.Require().NoError(err)
		s.Require().Equal(1, len(vals), "should have 1 validator at this time+height")
	}

	// Now only 2 time+height keys remain
	// This fits in the cache (2 keys == maxCacheSize)
	// GetPendingUnbondingValidators should now be able to use the cache
	ctx = ctx.WithBlockTime(unbondingTimes[4]).WithBlockHeight(unbondingHeights[4])
	remainingValidators, err := keeper.GetPendingUnbondingValidators(ctx, unbondingTimes[4], unbondingHeights[4])
	s.Require().NoError(err)
	s.Require().Equal(2, len(remainingValidators), "should have 2 time+height keys in cache")

	// Unbond the remaining 2 validators
	// Fast-forward to time when all remaining validators are mature
	ctx = ctx.WithBlockTime(unbondingTimes[4]).WithBlockHeight(unbondingHeights[4])
	err = keeper.UnbondAllMatureValidators(ctx)
	s.Require().NoError(err)

	// Verify all validators are now unbonded
	for i := 3; i < numKeys; i++ {
		// Verify not in queue
		vals, err := keeper.GetUnbondingValidators(ctx, unbondingTimes[i], unbondingHeights[i])
		s.Require().NoError(err)
		s.Require().Equal(0, len(vals), "should have no validators in queue")
	}
}

func (s *KeeperTestSuite) TestSortValidatorQueueKeysByAscendingTimestampOrder() {
	require := s.Require()

	currentTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	oneHourLater := currentTime.Add(1 * time.Hour)
	oneHourBefore := currentTime.Add(-1 * time.Hour)

	keys := []string{
		stakingtypes.GetCacheValidatorQueueKey(oneHourLater, 1000),
		stakingtypes.GetCacheValidatorQueueKey(oneHourBefore, 500),
		stakingtypes.GetCacheValidatorQueueKey(currentTime, 750),
		stakingtypes.GetCacheValidatorQueueKey(oneHourBefore, 600),
		stakingtypes.GetCacheValidatorQueueKey(oneHourLater, 900),
	}

	stakingtypes.SortValidatorQueueKeysByAscendingTimestampOrder(keys)

	// Verify sorting is correct - should be sorted by timestamp ascending order
	for i := 0; i < len(keys)-1; i++ {
		t1, _, err := stakingtypes.ParseCacheValidatorQueueKey(keys[i])
		require.NoError(err)
		t2, _, err := stakingtypes.ParseCacheValidatorQueueKey(keys[i+1])
		require.NoError(err)

		// Current entry should be before or equal to next entry
		require.True(t1.Before(t2) || t1.Equal(t2), "timestamps should be in ascending order")

	}

	firstTime, _, err := stakingtypes.ParseCacheValidatorQueueKey(keys[0])
	require.NoError(err)
	require.Equal(oneHourBefore, firstTime)

	lastTime, _, err := stakingtypes.ParseCacheValidatorQueueKey(keys[len(keys)-1])
	require.NoError(err)
	require.Equal(oneHourLater, lastTime)
}

func (s *KeeperTestSuite) TestGetAndParseCacheValidatorQueueKey() {
	require := s.Require()

	blockTime := time.Now().UTC()
	blockHeight := int64(1000)
	key := stakingtypes.GetCacheValidatorQueueKey(blockTime, blockHeight)
	time, height, err := stakingtypes.ParseCacheValidatorQueueKey(key)
	require.NoError(err)
	require.Equal(blockTime, time)
	require.Equal(blockHeight, height)
}
