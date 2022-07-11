package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmtime "github.com/tendermint/tendermint/libs/time"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletypes "github.com/cosmos/cosmos-sdk/types/module"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	"github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/feegrant/simulation"
	feegranttestutil "github.com/cosmos/cosmos-sdk/x/feegrant/testutil"
)

type SimTestSuite struct {
	suite.Suite

	baseApp        *baseapp.BaseApp
	ctx            sdk.Context
	feegrantKeeper keeper.Keeper
	accountKeeper  *feegranttestutil.MockAccountKeeper
	bankKeeper     *feegranttestutil.MockBankKeeper
	encCfg         moduletestutil.TestEncodingConfig
}

func (suite *SimTestSuite) SetupTest() {
	key := sdk.NewKVStoreKey(feegrant.StoreKey)
	suite.encCfg = moduletestutil.MakeTestEncodingConfig(module.AppModuleBasic{})

	ctrl := gomock.NewController(suite.T())
	suite.accountKeeper = feegranttestutil.NewMockAccountKeeper(ctrl)
	suite.bankKeeper = feegranttestutil.NewMockBankKeeper(ctrl)

	testCtx := testutil.DefaultContextWithDB(suite.T(), key, sdk.NewTransientStoreKey("transient_test"))
	suite.baseApp = baseapp.NewBaseApp(
		"feegrant",
		log.NewNopLogger(),
		testCtx.DB,
		suite.encCfg.TxConfig.TxDecoder(),
	)

	suite.baseApp.SetCMS(testCtx.CMS)
	suite.baseApp.SetInterfaceRegistry(suite.encCfg.InterfaceRegistry)

	suite.ctx = testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})
	suite.feegrantKeeper = keeper.NewKeeper(suite.encCfg.Codec, key, suite.accountKeeper)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.encCfg.InterfaceRegistry)
	feegrant.RegisterQueryServer(queryHelper, suite.feegrantKeeper)

	cfg := moduletypes.NewConfigurator(suite.encCfg.Codec, suite.baseApp.MsgServiceRouter(), suite.baseApp.GRPCQueryRouter())

	appModule := module.NewAppModule(suite.encCfg.Codec, suite.accountKeeper, suite.bankKeeper, suite.feegrantKeeper, suite.encCfg.InterfaceRegistry)
	appModule.RegisterServices(cfg)
	appModule.RegisterInterfaces(suite.encCfg.InterfaceRegistry)
}

func (suite *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)
	initAmt := sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	for _, acc := range accounts {
		suite.accountKeeper.EXPECT().GetAccount(gomock.Any(), acc.Address).Return(authtypes.NewBaseAccountWithAddress(acc.Address)).AnyTimes()
		suite.bankKeeper.EXPECT().SpendableCoins(gomock.Any(), acc.Address).Return(initCoins).AnyTimes()
	}

	return accounts
}

func (suite *SimTestSuite) TestWeightedOperations() {
	require := suite.Require()

	suite.ctx.WithChainID("test-chain")

	appParams := make(simtypes.AppParams)

	weightedOps := simulation.WeightedOperations(
		suite.encCfg.InterfaceRegistry,
		appParams, suite.encCfg.Codec, suite.accountKeeper,
		suite.bankKeeper, suite.feegrantKeeper,
	)

	// begin new block
	suite.baseApp.BeginBlock(abci.RequestBeginBlock{
		Header: tmproto.Header{
			Height:  suite.baseApp.LastBlockHeight() + 1,
			AppHash: suite.baseApp.LastCommitID().Hash,
		},
	})

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
			feegrant.MsgGrantAllowance{}.Route(),
			simulation.TypeMsgGrantAllowance,
		},
		{
			simulation.DefaultWeightRevokeAllowance,
			feegrant.MsgRevokeAllowance{}.Route(),
			simulation.TypeMsgRevokeAllowance,
		},
	}

	for i, w := range weightedOps {
		operationMsg, _, err := w.Op()(r, suite.baseApp, suite.ctx, accs, suite.ctx.ChainID())
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
	app, ctx := suite.baseApp, suite.ctx
	require := suite.Require()

	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 3)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgGrantAllowance(codec.NewProtoCodec(suite.encCfg.InterfaceRegistry), suite.accountKeeper, suite.bankKeeper, suite.feegrantKeeper)
	operationMsg, futureOperations, err := op(r, app, ctx, accounts, "")
	require.NoError(err)

	var msg feegrant.MsgGrantAllowance
	suite.encCfg.Amino.UnmarshalJSON(operationMsg.Msg, &msg)

	require.True(operationMsg.OK)
	require.Equal(accounts[2].Address.String(), msg.Granter)
	require.Equal(accounts[1].Address.String(), msg.Grantee)
	require.Len(futureOperations, 0)
}

func (suite *SimTestSuite) TestSimulateMsgRevokeAllowance() {
	app, ctx := suite.baseApp, suite.ctx
	require := suite.Require()

	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 3)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: suite.baseApp.LastBlockHeight() + 1, AppHash: suite.baseApp.LastCommitID().Hash}})

	feeAmt := sdk.TokensFromConsensusPower(200000, sdk.DefaultPowerReduction)
	feeCoins := sdk.NewCoins(sdk.NewCoin("foo", feeAmt))

	granter, grantee := accounts[0], accounts[1]

	oneYear := ctx.BlockTime().AddDate(1, 0, 0)
	err := suite.feegrantKeeper.GrantAllowance(
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
	op := simulation.SimulateMsgRevokeAllowance(codec.NewProtoCodec(suite.encCfg.InterfaceRegistry), suite.accountKeeper, suite.bankKeeper, suite.feegrantKeeper)
	operationMsg, futureOperations, err := op(r, app, ctx, accounts, "")
	require.NoError(err)

	var msg feegrant.MsgRevokeAllowance
	suite.encCfg.Amino.UnmarshalJSON(operationMsg.Msg, &msg)

	require.True(operationMsg.OK)
	require.Equal(granter.Address.String(), msg.Granter)
	require.Equal(grantee.Address.String(), msg.Grantee)
	require.Len(futureOperations, 0)
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}
