package simulation_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/cosmos/cosmos-sdk/x/feegrant/simulation"
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
	suite.ctx = app.BaseApp.NewContext(checkTx, tmproto.Header{
		Time: time.Now(),
	})

}

func (suite *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		err := testutil.FundAccount(suite.app.BankKeeper, suite.ctx, account.Address, initCoins)
		suite.Require().NoError(err)
	}

	return accounts
}

func (suite *SimTestSuite) TestWeightedOperations() {
	app, ctx := suite.app, suite.ctx
	require := suite.Require()

	ctx.WithChainID("test-chain")

	cdc := app.AppCodec()
	appParams := make(simtypes.AppParams)

	weightedOps := simulation.WeightedOperations(
		appParams, cdc, app.AccountKeeper,
		app.BankKeeper, app.FeeGrantKeeper,
	)

	s := rand.NewSource(1)
	r := rand.New(s)
	accs := suite.getTestingAccounts(r, 3)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{
			simappparams.DefaultWeightGrantAllowance,
			feegrant.MsgGrantAllowance{}.Route(),
			simulation.TypeMsgGrantAllowance,
		},
		{
			simappparams.DefaultWeightRevokeAllowance,
			feegrant.MsgRevokeAllowance{}.Route(),
			simulation.TypeMsgRevokeAllowance,
		},
	}

	for i, w := range weightedOps {
		operationMsg, _, _ := w.Op()(r, app.BaseApp, ctx, accs, ctx.ChainID())
		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		require.Equal(expected[i].weight, w.Weight(), "weight should be the same")
		require.Equal(expected[i].opMsgRoute, operationMsg.Route, "route should be the same")
		require.Equal(expected[i].opMsgName, operationMsg.Name, "operation Msg name should be the same")
	}
}

func (suite *SimTestSuite) TestSimulateMsgGrantAllowance() {
	app, ctx := suite.app, suite.ctx
	require := suite.Require()

	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 3)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgGrantAllowance(app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(err)

	var msg feegrant.MsgGrantAllowance
	suite.app.AppCodec().UnmarshalJSON(operationMsg.Msg, &msg)

	require.True(operationMsg.OK)
	require.Equal(accounts[2].Address.String(), msg.Granter)
	require.Equal(accounts[1].Address.String(), msg.Grantee)
	require.Len(futureOperations, 0)
}

func (suite *SimTestSuite) TestSimulateMsgRevokeAllowance() {
	app, ctx := suite.app, suite.ctx
	require := suite.Require()

	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 3)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: suite.app.LastBlockHeight() + 1, AppHash: suite.app.LastCommitID().Hash}})

	feeAmt := app.StakingKeeper.TokensFromConsensusPower(ctx, 200000)
	feeCoins := sdk.NewCoins(sdk.NewCoin("foo", feeAmt))

	granter, grantee := accounts[0], accounts[1]

	oneYear := ctx.BlockTime().AddDate(1, 0, 0)
	err := app.FeeGrantKeeper.GrantAllowance(
		ctx,
		granter.Address,
		grantee.Address,
		&feegrant.BasicAllowance{
			SpendLimit: feeCoins,
			Expiration: &oneYear,
		},
	)
	require.NoError(err)

	// execute operation
	op := simulation.SimulateMsgRevokeAllowance(app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(err)

	var msg feegrant.MsgRevokeAllowance
	suite.app.AppCodec().UnmarshalJSON(operationMsg.Msg, &msg)

	require.True(operationMsg.OK)
	require.Equal(granter.Address.String(), msg.Granter)
	require.Equal(grantee.Address.String(), msg.Grantee)
	require.Len(futureOperations, 0)
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}
