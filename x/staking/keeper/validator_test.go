package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func (suite *KeeperTestSuite) applyValidatorSetUpdates(ctx sdk.Context, keeper *stakingkeeper.Keeper, expectedUpdatesLen int) []abci.ValidatorUpdate {
	updates, err := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	suite.Require().NoError(err)
	if expectedUpdatesLen >= 0 {
		suite.Require().Equal(expectedUpdatesLen, len(updates), "%v", updates)
	}
	return updates
}

func (suite *KeeperTestSuite) TestValidator() {
	ctx, keeper := suite.ctx, suite.stakingKeeper
	require := suite.Require()

	valPubKey := PKs[0]
	valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
	valTokens := keeper.TokensFromConsensusPower(ctx, 10)
	valCoins := sdk.NewCoins(sdk.NewCoin(keeper.BondDenom(ctx), valTokens))

	// test how the validator is set from a purely unbonbed pool
	validator := teststaking.NewValidator(suite.T(), valAddr, valPubKey)
	validator, _ = validator.AddTokensFromDel(valTokens)
	require.Equal(stakingtypes.Unbonded, validator.Status)
	require.Equal(valTokens, validator.Tokens)
	require.Equal(valTokens, validator.DelegatorShares.RoundInt())
	keeper.SetValidator(ctx, validator)
	keeper.SetValidatorByPowerIndex(ctx, validator)

	// ensure update
	suite.bankKeeper.EXPECT().SendCoinsFromModuleToModule(ctx, stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, valCoins)
	updates := suite.applyValidatorSetUpdates(ctx, &keeper, 1)
	validator, found := keeper.GetValidator(ctx, valAddr)
	require.True(found)
	require.Equal(validator.ABCIValidatorUpdate(keeper.PowerReduction(ctx)), updates[0])

	// after the save the validator should be bonded
	require.Equal(stakingtypes.Bonded, validator.Status)
	require.Equal(valTokens, validator.Tokens)
	require.Equal(valTokens, validator.DelegatorShares.RoundInt())

	// Check each store for being saved
	resVals := keeper.GetLastValidators(ctx)
	require.Equal(1, len(resVals))
	require.True(validator.MinEqual(&resVals[0]))

	resVals = keeper.GetBondedValidatorsByPower(ctx)
	require.Equal(1, len(resVals))
	require.True(validator.MinEqual(&resVals[0]))

	allVals := keeper.GetAllValidators(ctx)
	require.Equal(1, len(allVals))
}
