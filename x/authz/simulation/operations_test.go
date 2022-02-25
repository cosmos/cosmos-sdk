package simulation_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/authz/simulation"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type SimTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *simapp.SimApp
}

func (suite *SimTestSuite) SetupTest() {
	checkTx := false
	app := simapp.Setup(suite.T(), checkTx)
	suite.app = app
	suite.ctx = app.BaseApp.NewContext(checkTx, tmproto.Header{})
}

func (suite *SimTestSuite) TestWeightedOperations() {
	cdc := suite.app.AppCodec()
	appParams := make(simtypes.AppParams)

	weightedOps := simulation.WeightedOperations(appParams, cdc, suite.app.AccountKeeper,
		suite.app.BankKeeper, suite.app.AuthzKeeper, cdc,
	)

	s := rand.NewSource(3)
	r := rand.New(s)
	// setup 2 accounts
	accs := suite.getTestingAccounts(r, 2)

	expected := []struct {
		weight     int
		opMsgRoute string
	}{
		{simulation.WeightGrant, simulation.TypeMsgGrant},
		{simulation.WeightExec, simulation.TypeMsgExec},
		{simulation.WeightRevoke, simulation.TypeMsgRevoke},
	}

	for i, w := range weightedOps {
		operationMsg, _, _ := w.Op()(r, suite.app.BaseApp, suite.ctx, accs, "")
		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		suite.Require().Equal(expected[i].weight, w.Weight(), "weight should be the same")
		suite.Require().Equal(expected[i].opMsgRoute, operationMsg.Route, "route should be the same")
		suite.Require().Equal(expected[i].opMsgRoute, operationMsg.Name, "operation Msg name should be the same")
	}
}

func (suite *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, 200000)
	initCoins := sdk.NewCoins(sdk.NewCoin("stake", initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, account.Address)
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
		suite.Require().NoError(testutil.FundAccount(suite.app.BankKeeper, suite.ctx, account.Address, initCoins))
	}

	return accounts
}

func (suite *SimTestSuite) TestSimulateGrant() {
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 2)
	blockTime := time.Now().UTC()
	ctx := suite.ctx.WithBlockTime(blockTime)

	// begin a new block
	suite.app.BeginBlock(abci.RequestBeginBlock{
		Header: tmproto.Header{
			Height:  suite.app.LastBlockHeight() + 1,
			AppHash: suite.app.LastCommitID().Hash,
		},
	})

	granter := accounts[0]
	grantee := accounts[1]

	// execute operation
	op := simulation.SimulateMsgGrant(suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.AuthzKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, ctx, accounts, "")
	suite.Require().NoError(err)

	var msg authz.MsgGrant
	suite.app.LegacyAmino().UnmarshalJSON(operationMsg.Msg, &msg)
	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(granter.Address.String(), msg.Granter)
	suite.Require().Equal(grantee.Address.String(), msg.Grantee)
	suite.Require().Len(futureOperations, 0)

}

func (suite *SimTestSuite) TestSimulateRevoke() {
	// setup 3 accounts
	s := rand.NewSource(2)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 3)

	// begin a new block
	suite.app.BeginBlock(abci.RequestBeginBlock{
		Header: tmproto.Header{
			Height:  suite.app.LastBlockHeight() + 1,
			AppHash: suite.app.LastCommitID().Hash,
		}})

	initAmt := suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, 200000)
	initCoins := sdk.NewCoins(sdk.NewCoin("stake", initAmt))

	granter := accounts[0]
	grantee := accounts[1]
	authorization := banktypes.NewSendAuthorization(initCoins)

	err := suite.app.AuthzKeeper.SaveGrant(suite.ctx, grantee.Address, granter.Address, authorization, time.Now().Add(30*time.Hour))
	suite.Require().NoError(err)

	// execute operation
	op := simulation.SimulateMsgRevoke(suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.AuthzKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg authz.MsgRevoke
	suite.app.LegacyAmino().UnmarshalJSON(operationMsg.Msg, &msg)

	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(granter.Address.String(), msg.Granter)
	suite.Require().Equal(grantee.Address.String(), msg.Grantee)
	suite.Require().Equal(banktypes.SendAuthorization{}.MsgTypeURL(), msg.MsgTypeUrl)
	suite.Require().Len(futureOperations, 0)

}

func (suite *SimTestSuite) TestSimulateExec() {
	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 3)

	// begin a new block
	suite.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: suite.app.LastBlockHeight() + 1, AppHash: suite.app.LastCommitID().Hash}})

	initAmt := suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, 200000)
	initCoins := sdk.NewCoins(sdk.NewCoin("stake", initAmt))

	granter := accounts[0]
	grantee := accounts[1]
	authorization := banktypes.NewSendAuthorization(initCoins)

	err := suite.app.AuthzKeeper.SaveGrant(suite.ctx, grantee.Address, granter.Address, authorization, suite.ctx.BlockTime().Add(1*time.Hour))
	suite.Require().NoError(err)

	// execute operation
	op := simulation.SimulateMsgExec(suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.AuthzKeeper, suite.app.AppCodec())
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg authz.MsgExec

	suite.app.LegacyAmino().UnmarshalJSON(operationMsg.Msg, &msg)

	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(grantee.Address.String(), msg.Grantee)
	suite.Require().Len(futureOperations, 0)

}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}
