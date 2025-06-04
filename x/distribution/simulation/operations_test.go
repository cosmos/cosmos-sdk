package simulation_test

import (
	"math/rand"
	"testing"

	abci "github.com/cometbft/cometbft/v2/abci/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/simulation"
	distrtestutil "github.com/cosmos/cosmos-sdk/x/distribution/testutil"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TestWeightedOperations tests the weights of the operations.
func (suite *SimTestSuite) TestWeightedOperations() {
	appParams := make(simtypes.AppParams)

	weightedOps := simulation.WeightedOperations(appParams, suite.cdc, suite.txConfig, suite.accountKeeper,
		suite.bankKeeper, suite.distrKeeper, suite.stakingKeeper)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accs := suite.getTestingAccounts(r, 3)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{simulation.DefaultWeightMsgSetWithdrawAddress, types.ModuleName, sdk.MsgTypeURL(&types.MsgSetWithdrawAddress{})},
		{simulation.DefaultWeightMsgWithdrawDelegationReward, types.ModuleName, sdk.MsgTypeURL(&types.MsgWithdrawDelegatorReward{})},
		{simulation.DefaultWeightMsgWithdrawValidatorCommission, types.ModuleName, sdk.MsgTypeURL(&types.MsgWithdrawValidatorCommission{})},
		{simulation.DefaultWeightMsgFundCommunityPool, types.ModuleName, sdk.MsgTypeURL(&types.MsgFundCommunityPool{})},
	}

	for i, w := range weightedOps {
		operationMsg, _, err := w.Op()(r, suite.app.BaseApp, suite.ctx, accs, "")
		suite.Require().NoError(err)

		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		suite.Require().Equal(expected[i].weight, w.Weight(), "weight should be the same")
		suite.Require().Equal(expected[i].opMsgRoute, operationMsg.Route, "route should be the same")
		suite.Require().Equal(expected[i].opMsgName, operationMsg.Name, "operation Msg name should be the same")
	}
}

// TestSimulateMsgSetWithdrawAddress tests the normal scenario of a valid message of type TypeMsgSetWithdrawAddress.
// Abonormal scenarios, where the message is created by an errors, are not tested here.
func (suite *SimTestSuite) TestSimulateMsgSetWithdrawAddress() {
	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 3)

	_, err := suite.app.FinalizeBlock(&abci.FinalizeBlockRequest{
		Height: suite.app.LastBlockHeight() + 1,
		Hash:   suite.app.LastCommitID().Hash,
	})
	suite.Require().NoError(err)

	// execute operation
	op := simulation.SimulateMsgSetWithdrawAddress(suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.distrKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg types.MsgSetWithdrawAddress
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	suite.Require().NoError(err)
	suite.Require().True(operationMsg.OK)
	suite.Require().Equal("cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", msg.DelegatorAddress)
	suite.Require().Equal("cosmos1p8wcgrjr4pjju90xg6u9cgq55dxwq8j7u4x9a0", msg.WithdrawAddress)
	suite.Require().Equal(sdk.MsgTypeURL(&types.MsgSetWithdrawAddress{}), sdk.MsgTypeURL(&msg))
	suite.Require().Len(futureOperations, 0)
}

// TestSimulateMsgWithdrawDelegatorReward tests the normal scenario of a valid message
// of type TypeMsgWithdrawDelegatorReward.
// Abonormal scenarios, where the message is created by an errors, are not tested here.
func (suite *SimTestSuite) TestSimulateMsgWithdrawDelegatorReward() {
	// setup 3 accounts
	s := rand.NewSource(4)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 3)

	// setup accounts[0] as validator
	validator0 := suite.getTestingValidator0(accounts)

	// setup delegation
	delTokens := sdk.TokensFromConsensusPower(2, sdk.DefaultPowerReduction)
	validator0, issuedShares := validator0.AddTokensFromDel(delTokens)
	delegator := accounts[1]

	delegation := stakingtypes.NewDelegation(delegator.Address.String(), validator0.GetOperator(), issuedShares)
	suite.Require().NoError(suite.stakingKeeper.SetDelegation(suite.ctx, delegation))
	valBz, err := address.NewBech32Codec("cosmosvaloper").StringToBytes(validator0.GetOperator())
	suite.Require().NoError(err)
	suite.Require().NoError(suite.distrKeeper.SetDelegatorStartingInfo(suite.ctx, valBz, delegator.Address, types.NewDelegatorStartingInfo(2, math.LegacyOneDec(), 200)))

	suite.setupValidatorRewards(valBz)

	_, err = suite.app.FinalizeBlock(&abci.FinalizeBlockRequest{
		Height: suite.app.LastBlockHeight() + 1,
		Hash:   suite.app.LastCommitID().Hash,
	})
	suite.Require().NoError(err)

	// execute operation
	op := simulation.SimulateMsgWithdrawDelegatorReward(suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.distrKeeper, suite.stakingKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg types.MsgWithdrawDelegatorReward
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	suite.Require().NoError(err)
	suite.Require().True(operationMsg.OK)
	suite.Require().Equal("cosmosvaloper1l4s054098kk9hmr5753c6k3m2kw65h686d3mhr", msg.ValidatorAddress)
	suite.Require().Equal("cosmos1d6u7zhjwmsucs678d7qn95uqajd4ucl9jcjt26", msg.DelegatorAddress)
	suite.Require().Equal(sdk.MsgTypeURL(&types.MsgWithdrawDelegatorReward{}), sdk.MsgTypeURL(&msg))
	suite.Require().Len(futureOperations, 0)
}

// TestSimulateMsgWithdrawValidatorCommission tests the normal scenario of a valid message
// of type TypeMsgWithdrawValidatorCommission.
// Abonormal scenarios, where the message is created by an errors, are not tested here.
func (suite *SimTestSuite) TestSimulateMsgWithdrawValidatorCommission() {
	suite.testSimulateMsgWithdrawValidatorCommission("atoken")
	suite.testSimulateMsgWithdrawValidatorCommission("tokenxxx")
}

// all the checks in this function should not fail if we change the tokenName
func (suite *SimTestSuite) testSimulateMsgWithdrawValidatorCommission(tokenName string) {
	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 3)

	// setup accounts[0] as validator
	validator0 := suite.getTestingValidator0(accounts)

	// set module account coins
	distrAcc := suite.distrKeeper.GetDistributionAccount(suite.ctx)
	suite.Require().NoError(banktestutil.FundModuleAccount(suite.ctx, suite.bankKeeper, distrAcc.GetName(), sdk.NewCoins(
		sdk.NewCoin(tokenName, math.NewInt(10)),
		sdk.NewCoin("stake", math.NewInt(5)),
	)))
	suite.accountKeeper.SetModuleAccount(suite.ctx, distrAcc)

	// set outstanding rewards
	valCommission := sdk.NewDecCoins(
		sdk.NewDecCoinFromDec(tokenName, math.LegacyNewDec(5).Quo(math.LegacyNewDec(2))),
		sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(1).Quo(math.LegacyNewDec(1))),
	)
	valCodec := address.NewBech32Codec("cosmosvaloper")

	val0, err := valCodec.StringToBytes(validator0.GetOperator())
	suite.Require().NoError(err)

	genVal0, err := valCodec.StringToBytes(suite.genesisVals[0].GetOperator())
	suite.Require().NoError(err)

	suite.Require().NoError(suite.distrKeeper.SetValidatorOutstandingRewards(suite.ctx, val0, types.ValidatorOutstandingRewards{Rewards: valCommission}))
	suite.Require().NoError(suite.distrKeeper.SetValidatorOutstandingRewards(suite.ctx, genVal0, types.ValidatorOutstandingRewards{Rewards: valCommission}))

	// setup validator accumulated commission
	suite.Require().NoError(suite.distrKeeper.SetValidatorAccumulatedCommission(suite.ctx, val0, types.ValidatorAccumulatedCommission{Commission: valCommission}))
	suite.Require().NoError(suite.distrKeeper.SetValidatorAccumulatedCommission(suite.ctx, genVal0, types.ValidatorAccumulatedCommission{Commission: valCommission}))

	_, err = suite.app.FinalizeBlock(&abci.FinalizeBlockRequest{
		Height: suite.app.LastBlockHeight() + 1,
		Hash:   suite.app.LastCommitID().Hash,
	})
	suite.Require().NoError(err)

	// execute operation
	op := simulation.SimulateMsgWithdrawValidatorCommission(suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.distrKeeper, suite.stakingKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	if !operationMsg.OK {
		suite.Require().Equal("could not find account", operationMsg.Comment)
	} else {
		suite.Require().NoError(err)

		var msg types.MsgWithdrawValidatorCommission
		err = proto.Unmarshal(operationMsg.Msg, &msg)
		suite.Require().NoError(err)
		suite.Require().True(operationMsg.OK)
		suite.Require().Equal("cosmosvaloper1tnh2q55v8wyygtt9srz5safamzdengsn9dsd7z", msg.ValidatorAddress)
		suite.Require().Equal(sdk.MsgTypeURL(&types.MsgWithdrawValidatorCommission{}), sdk.MsgTypeURL(&msg))
		suite.Require().Len(futureOperations, 0)
	}
}

// TestSimulateMsgFundCommunityPool tests the normal scenario of a valid message of type TypeMsgFundCommunityPool.
// Abonormal scenarios, where the message is created by an errors, are not tested here.
func (suite *SimTestSuite) TestSimulateMsgFundCommunityPool() {
	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 3)

	_, err := suite.app.FinalizeBlock(&abci.FinalizeBlockRequest{
		Height: suite.app.LastBlockHeight() + 1,
		Hash:   suite.app.LastCommitID().Hash,
	})
	suite.Require().NoError(err)

	// execute operation
	op := simulation.SimulateMsgFundCommunityPool(suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.distrKeeper, suite.stakingKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg types.MsgFundCommunityPool
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	suite.Require().NoError(err)
	suite.Require().True(operationMsg.OK)
	suite.Require().Equal("4896096stake", msg.Amount.String())
	suite.Require().Equal("cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", msg.Depositor)
	suite.Require().Equal(sdk.MsgTypeURL(&types.MsgFundCommunityPool{}), sdk.MsgTypeURL(&msg))
	suite.Require().Len(futureOperations, 0)
}

type SimTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	app         *runtime.App
	genesisVals []stakingtypes.Validator

	txConfig      client.TxConfig
	cdc           codec.Codec
	stakingKeeper *stakingkeeper.Keeper
	accountKeeper authkeeper.AccountKeeper
	bankKeeper    bankkeeper.Keeper
	distrKeeper   keeper.Keeper
}

func (suite *SimTestSuite) SetupTest() {
	var (
		appBuilder *runtime.AppBuilder
		err        error
	)
	suite.app, err = simtestutil.Setup(
		depinject.Configs(
			distrtestutil.AppConfig,
			depinject.Supply(log.NewNopLogger()),
		),
		&suite.accountKeeper,
		&suite.bankKeeper,
		&suite.cdc,
		&appBuilder,
		&suite.stakingKeeper,
		&suite.distrKeeper,
		&suite.txConfig,
	)

	suite.NoError(err)

	suite.ctx = suite.app.NewContext(false)

	genesisVals, err := suite.stakingKeeper.GetAllValidators(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(genesisVals, 1)
	suite.genesisVals = genesisVals
}

func (suite *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 200)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, account.Address)
		suite.accountKeeper.SetAccount(suite.ctx, acc)
		suite.Require().NoError(banktestutil.FundAccount(suite.ctx, suite.bankKeeper, account.Address, initCoins))
	}

	return accounts
}

func (suite *SimTestSuite) getTestingValidator0(accounts []simtypes.Account) stakingtypes.Validator {
	commission0 := stakingtypes.NewCommission(math.LegacyZeroDec(), math.LegacyOneDec(), math.LegacyOneDec())
	return suite.getTestingValidator(accounts, commission0, 0)
}

func (suite *SimTestSuite) getTestingValidator(accounts []simtypes.Account, commission stakingtypes.Commission, n int) stakingtypes.Validator {
	require := suite.Require()
	account := accounts[n]
	valPubKey := account.PubKey
	valAddr := sdk.ValAddress(account.PubKey.Address().Bytes())
	validator, err := stakingtypes.NewValidator(valAddr.String(), valPubKey, stakingtypes.
		Description{})
	require.NoError(err)
	validator, err = validator.SetInitialCommission(commission)
	require.NoError(err)
	validator.DelegatorShares = math.LegacyNewDec(100)
	validator.Tokens = math.NewInt(1000000)

	require.NoError(suite.stakingKeeper.SetValidator(suite.ctx, validator))

	return validator
}

func (suite *SimTestSuite) setupValidatorRewards(valAddress sdk.ValAddress) {
	decCoins := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyOneDec())}
	historicalRewards := types.NewValidatorHistoricalRewards(decCoins, 2)
	suite.Require().NoError(suite.distrKeeper.SetValidatorHistoricalRewards(suite.ctx, valAddress, 2, historicalRewards))
	// setup current revards
	currentRewards := types.NewValidatorCurrentRewards(decCoins, 3)
	suite.Require().NoError(suite.distrKeeper.SetValidatorCurrentRewards(suite.ctx, valAddress, currentRewards))
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}
