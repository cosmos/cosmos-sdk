package simulation_test

import (
	"math/rand"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codecaddress "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkapp "github.com/cosmos/cosmos-sdk/app"
	testapp "github.com/cosmos/cosmos-sdk/testutil/testapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	"github.com/cosmos/cosmos-sdk/x/feegrant/simulation"
)

type SimTestSuite struct {
	suite.Suite

	app               *sdkapp.SDKApp
	ctx               sdk.Context
	feegrantKeeper    keeper.Keeper
	interfaceRegistry codectypes.InterfaceRegistry
	txConfig          client.TxConfig
	accountKeeper     authkeeper.AccountKeeper
	bankKeeper        bankkeeper.Keeper
	cdc               codec.Codec
	legacyAmino       *codec.LegacyAmino
}

func (suite *SimTestSuite) SetupTest() {
	ta := testapp.Setup(suite.T())
	suite.app = ta
	suite.accountKeeper = ta.AccountKeeper
	suite.bankKeeper = ta.BankKeeper
	suite.cdc = ta.AppCodec()
	suite.legacyAmino = ta.LegacyAmino()
	suite.txConfig = ta.TxConfig()
	suite.interfaceRegistry = ta.InterfaceRegistry()
	suite.feegrantKeeper = *ta.FeeGrantKeeper
	suite.ctx = testapp.NewContext(ta).WithBlockTime(time.Now().UTC())
}

func (suite *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)
	initAmt := sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, account.Address)
		suite.accountKeeper.SetAccount(suite.ctx, acc)
		err := banktestutil.FundAccount(suite.ctx, suite.bankKeeper, account.Address, initCoins)
		suite.Require().NoError(err)
	}

	return accounts
}

func (suite *SimTestSuite) TestWeightedOperations() {
	require := suite.Require()

	suite.ctx.WithChainID("test-chain")

	appParams := make(simtypes.AppParams)

	weightedOps := simulation.WeightedOperations(
		suite.interfaceRegistry,
		appParams, suite.cdc, suite.txConfig, suite.accountKeeper,
		suite.bankKeeper, suite.feegrantKeeper, codecaddress.NewBech32Codec("cosmos"),
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
			simulation.DefaultWeightGrantAllowance,
			feegrant.ModuleName,
			sdk.MsgTypeURL(&feegrant.MsgGrantAllowance{}),
		},
		{
			simulation.DefaultWeightRevokeAllowance,
			feegrant.ModuleName,
			sdk.MsgTypeURL(&feegrant.MsgRevokeAllowance{}),
		},
	}

	for i, w := range weightedOps {
		operationMsg, _, err := w.Op()(r, suite.app.BaseApp, suite.ctx, accs, suite.ctx.ChainID())
		require.NoError(err)

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

	//  new block
	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: app.LastBlockHeight() + 1})
	require.NoError(err)

	// execute operation
	op := simulation.SimulateMsgGrantAllowance(codec.NewProtoCodec(suite.interfaceRegistry), suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.feegrantKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, ctx.ChainID())
	require.NoError(err)

	var msg feegrant.MsgGrantAllowance
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	require.NoError(err)
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
	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: suite.app.LastBlockHeight() + 1, Hash: suite.app.LastCommitID().Hash})
	suite.Require().NoError(err)

	feeAmt := sdk.TokensFromConsensusPower(200000, sdk.DefaultPowerReduction)
	feeCoins := sdk.NewCoins(sdk.NewCoin("foo", feeAmt))

	granter, grantee := accounts[0], accounts[1]

	oneYear := ctx.BlockTime().AddDate(1, 0, 0)
	err = suite.feegrantKeeper.GrantAllowance(
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
	op := simulation.SimulateMsgRevokeAllowance(codec.NewProtoCodec(suite.interfaceRegistry), suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.feegrantKeeper, codecaddress.NewBech32Codec("cosmos"))
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, ctx.ChainID())
	require.NoError(err)

	var msg feegrant.MsgRevokeAllowance
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	require.NoError(err)
	require.True(operationMsg.OK)
	require.Equal(granter.Address.String(), msg.Granter)
	require.Equal(grantee.Address.String(), msg.Grantee)
	require.Len(futureOperations, 0)
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}
