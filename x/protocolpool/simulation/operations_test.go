package simulation_test

import (
	"math/rand"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/x/protocolpool/keeper"
	"cosmossdk.io/x/protocolpool/simulation"
	pooltestutil "cosmossdk.io/x/protocolpool/testutil"
	"cosmossdk.io/x/protocolpool/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
)

type SimTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *runtime.App

	txConfig      client.TxConfig
	cdc           codec.Codec
	accountKeeper authkeeper.AccountKeeper
	bankKeeper    bankkeeper.Keeper
	poolKeeper    keeper.Keeper
}

func (suite *SimTestSuite) SetupTest() {
	var (
		appBuilder *runtime.AppBuilder
		err        error
	)
	suite.app, err = simtestutil.Setup(
		depinject.Configs(
			pooltestutil.AppConfig,
			depinject.Supply(log.NewNopLogger()),
		),
		&suite.accountKeeper,
		&suite.bankKeeper,
		&suite.cdc,
		&appBuilder,
		&suite.poolKeeper,
		&suite.txConfig,
	)

	suite.NoError(err)

	suite.ctx = suite.app.BaseApp.NewContext(false)
}

// TestWeightedOperations tests the weights of the operations.
func (suite *SimTestSuite) TestWeightedOperations() {
	appParams := make(simtypes.AppParams)

	weightedOps := simulation.WeightedOperations(appParams, suite.cdc, suite.txConfig, suite.accountKeeper,
		suite.bankKeeper, suite.poolKeeper)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accs := suite.getTestingAccounts(r, 3)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
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

// TestSimulateMsgFundCommunityPool tests the normal scenario of a valid message of type TypeMsgFundCommunityPool.
// Abonormal scenarios, where the message is created by an errors, are not tested here.
func (suite *SimTestSuite) TestSimulateMsgFundCommunityPool() {
	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 3)

	_, err := suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: suite.app.LastBlockHeight() + 1,
		Hash:   suite.app.LastCommitID().Hash,
	})
	suite.Require().NoError(err)

	// execute operation
	op := simulation.SimulateMsgFundCommunityPool(suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.poolKeeper)
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

func (suite *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(200)))

	// add coins to the accounts
	for _, account := range accounts {
		acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, account.Address)
		suite.accountKeeper.SetAccount(suite.ctx, acc)
		suite.Require().NoError(banktestutil.FundAccount(suite.ctx, suite.bankKeeper, account.Address, initCoins))
	}

	return accounts
}
