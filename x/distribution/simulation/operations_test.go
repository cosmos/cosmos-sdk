package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/distribution/simulation"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TestWeightedOperations tests the weights of the operations.
func (suite *SimTestSuite) TestWeightedOperations() {
	cdc := suite.app.AppCodec()
	appParams := make(simtypes.AppParams)

	weightesOps := simulation.WeightedOperations(appParams, cdc, suite.app.AccountKeeper,
		suite.app.BankKeeper, suite.app.DistrKeeper, suite.app.StakingKeeper,
	)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accs := suite.getTestingAccounts(r, 3)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{simappparams.DefaultWeightMsgSetWithdrawAddress, types.ModuleName, types.TypeMsgSetWithdrawAddress},
		{simappparams.DefaultWeightMsgWithdrawDelegationReward, types.ModuleName, types.TypeMsgWithdrawDelegatorReward},
		{simappparams.DefaultWeightMsgWithdrawValidatorCommission, types.ModuleName, types.TypeMsgWithdrawValidatorCommission},
		{simappparams.DefaultWeightMsgFundCommunityPool, types.ModuleName, types.TypeMsgFundCommunityPool},
	}

	for i, w := range weightesOps {
		operationMsg, _, _ := w.Op()(r, suite.app.BaseApp, suite.ctx, accs, "")
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

	// begin a new block
	suite.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: suite.app.LastBlockHeight() + 1, AppHash: suite.app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgSetWithdrawAddress(suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.DistrKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg types.MsgSetWithdrawAddress
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	suite.Require().True(operationMsg.OK)
	suite.Require().Equal("cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", msg.DelegatorAddress)
	suite.Require().Equal("cosmos1p8wcgrjr4pjju90xg6u9cgq55dxwq8j7u4x9a0", msg.WithdrawAddress)
	suite.Require().Equal(types.TypeMsgSetWithdrawAddress, msg.Type())
	suite.Require().Equal(types.ModuleName, msg.Route())
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
	delTokens := sdk.TokensFromConsensusPower(2)
	validator0, issuedShares := validator0.AddTokensFromDel(delTokens)
	delegator := accounts[1]
	delegation := stakingtypes.NewDelegation(delegator.Address, validator0.GetOperator(), issuedShares)
	suite.app.StakingKeeper.SetDelegation(suite.ctx, delegation)
	suite.app.DistrKeeper.SetDelegatorStartingInfo(suite.ctx, validator0.GetOperator(), delegator.Address, distrtypes.NewDelegatorStartingInfo(2, sdk.OneDec(), 200))

	suite.setupValidatorRewards(validator0.GetOperator())

	// begin a new block
	suite.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: suite.app.LastBlockHeight() + 1, AppHash: suite.app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgWithdrawDelegatorReward(suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.DistrKeeper, suite.app.StakingKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg types.MsgWithdrawDelegatorReward
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	suite.Require().True(operationMsg.OK)
	suite.Require().Equal("cosmosvaloper1l4s054098kk9hmr5753c6k3m2kw65h686d3mhr", msg.ValidatorAddress)
	suite.Require().Equal("cosmos1d6u7zhjwmsucs678d7qn95uqajd4ucl9jcjt26", msg.DelegatorAddress)
	suite.Require().Equal(types.TypeMsgWithdrawDelegatorReward, msg.Type())
	suite.Require().Equal(types.ModuleName, msg.Route())
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
	distrAcc := suite.app.DistrKeeper.GetDistributionAccount(suite.ctx)
	err := suite.app.BankKeeper.SetBalances(suite.ctx, distrAcc.GetAddress(), sdk.NewCoins(
		sdk.NewCoin(tokenName, sdk.NewInt(10)),
		sdk.NewCoin("stake", sdk.NewInt(5)),
	))
	suite.Require().NoError(err)
	suite.app.AccountKeeper.SetModuleAccount(suite.ctx, distrAcc)

	// set outstanding rewards
	valCommission := sdk.NewDecCoins(
		sdk.NewDecCoinFromDec(tokenName, sdk.NewDec(5).Quo(sdk.NewDec(2))),
		sdk.NewDecCoinFromDec("stake", sdk.NewDec(1).Quo(sdk.NewDec(1))),
	)

	suite.app.DistrKeeper.SetValidatorOutstandingRewards(suite.ctx, validator0.GetOperator(), types.ValidatorOutstandingRewards{Rewards: valCommission})

	// setup validator accumulated commission
	suite.app.DistrKeeper.SetValidatorAccumulatedCommission(suite.ctx, validator0.GetOperator(), types.ValidatorAccumulatedCommission{Commission: valCommission})

	// begin a new block
	suite.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: suite.app.LastBlockHeight() + 1, AppHash: suite.app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgWithdrawValidatorCommission(suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.DistrKeeper, suite.app.StakingKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg types.MsgWithdrawValidatorCommission
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	suite.Require().True(operationMsg.OK)
	suite.Require().Equal("cosmosvaloper1tnh2q55v8wyygtt9srz5safamzdengsn9dsd7z", msg.ValidatorAddress)
	suite.Require().Equal(types.TypeMsgWithdrawValidatorCommission, msg.Type())
	suite.Require().Equal(types.ModuleName, msg.Route())
	suite.Require().Len(futureOperations, 0)
}

// TestSimulateMsgFundCommunityPool tests the normal scenario of a valid message of type TypeMsgFundCommunityPool.
// Abonormal scenarios, where the message is created by an errors, are not tested here.
func (suite *SimTestSuite) TestSimulateMsgFundCommunityPool() {
	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 3)

	// begin a new block
	suite.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: suite.app.LastBlockHeight() + 1, AppHash: suite.app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgFundCommunityPool(suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.DistrKeeper, suite.app.StakingKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg types.MsgFundCommunityPool
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	suite.Require().True(operationMsg.OK)
	suite.Require().Equal("4896096stake", msg.Amount.String())
	suite.Require().Equal("cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", msg.Depositor)
	suite.Require().Equal(types.TypeMsgFundCommunityPool, msg.Type())
	suite.Require().Equal(types.ModuleName, msg.Route())
	suite.Require().Len(futureOperations, 0)
}

type SimTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *simapp.SimApp
}

func (suite *SimTestSuite) SetupTest() {
	checkTx := false
	app := simapp.Setup(checkTx)
	suite.app = app
	suite.ctx = app.BaseApp.NewContext(checkTx, tmproto.Header{})
}

func (suite *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := sdk.TokensFromConsensusPower(200)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, account.Address)
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
		err := suite.app.BankKeeper.SetBalances(suite.ctx, account.Address, initCoins)
		suite.Require().NoError(err)
	}

	return accounts
}

func (suite *SimTestSuite) getTestingValidator0(accounts []simtypes.Account) stakingtypes.Validator {
	commission0 := stakingtypes.NewCommission(sdk.ZeroDec(), sdk.OneDec(), sdk.OneDec())
	return suite.getTestingValidator(accounts, commission0, 0)
}

func (suite *SimTestSuite) getTestingValidator(accounts []simtypes.Account, commission stakingtypes.Commission, n int) stakingtypes.Validator {
	require := suite.Require()
	account := accounts[n]
	valPubKey := account.PubKey
	valAddr := sdk.ValAddress(account.PubKey.Address().Bytes())
	validator, err := stakingtypes.NewValidator(valAddr, valPubKey, stakingtypes.
		Description{})
	require.NoError(err)
	validator, err = validator.SetInitialCommission(commission)
	require.NoError(err)
	validator.DelegatorShares = sdk.NewDec(100)
	validator.Tokens = sdk.NewInt(1000000)

	suite.app.StakingKeeper.SetValidator(suite.ctx, validator)

	return validator
}

func (suite *SimTestSuite) setupValidatorRewards(valAddress sdk.ValAddress) {
	decCoins := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.OneDec())}
	historicalRewards := distrtypes.NewValidatorHistoricalRewards(decCoins, 2)
	suite.app.DistrKeeper.SetValidatorHistoricalRewards(suite.ctx, valAddress, 2, historicalRewards)
	// setup current revards
	currentRewards := distrtypes.NewValidatorCurrentRewards(decCoins, 3)
	suite.app.DistrKeeper.SetValidatorCurrentRewards(suite.ctx, valAddress, currentRewards)

}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}
