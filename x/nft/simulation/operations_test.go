package simulation_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletypes "github.com/cosmos/cosmos-sdk/types/module"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/nft"
	"github.com/cosmos/cosmos-sdk/x/nft/keeper"
	nftkeeper "github.com/cosmos/cosmos-sdk/x/nft/keeper"
	"github.com/cosmos/cosmos-sdk/x/nft/module"
	"github.com/cosmos/cosmos-sdk/x/nft/simulation"
	nfttestutil "github.com/cosmos/cosmos-sdk/x/nft/testutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmtime "github.com/tendermint/tendermint/libs/time"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type SimTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	baseApp       *baseapp.BaseApp
	accountKeeper *nfttestutil.MockAccountKeeper
	bankKeeper    *nfttestutil.MockBankKeeper
	nftKeeper     nftkeeper.Keeper
	encCfg        moduletestutil.TestEncodingConfig
}

func (suite *SimTestSuite) SetupTest() {
	key := sdk.NewKVStoreKey(nft.StoreKey)
	// suite setup
	addrs := simtestutil.CreateIncrementalAccounts(3)
	suite.encCfg = moduletestutil.MakeTestEncodingConfig(module.AppModuleBasic{})

	// gomock initializations
	ctrl := gomock.NewController(suite.T())
	suite.accountKeeper = nfttestutil.NewMockAccountKeeper(ctrl)
	suite.bankKeeper = nfttestutil.NewMockBankKeeper(ctrl)

	suite.accountKeeper.EXPECT().GetModuleAddress(nft.ModuleName).Return(addrs[0]).AnyTimes()

	testCtx := testutil.DefaultContextWithDB(suite.T(), key, sdk.NewTransientStoreKey("transient_test"))
	suite.baseApp = baseapp.NewBaseApp(
		"nft",
		log.NewNopLogger(),
		testCtx.DB,
		suite.encCfg.TxConfig.TxDecoder(),
	)

	suite.baseApp.SetCMS(testCtx.CMS)

	suite.baseApp.SetInterfaceRegistry(suite.encCfg.InterfaceRegistry)
	suite.ctx = testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})

	suite.nftKeeper = keeper.NewKeeper(key, suite.encCfg.Codec, suite.accountKeeper, suite.bankKeeper)
	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.encCfg.InterfaceRegistry)
	nft.RegisterQueryServer(queryHelper, suite.nftKeeper)

	cfg := moduletypes.NewConfigurator(suite.encCfg.Codec, suite.baseApp.MsgServiceRouter(), suite.baseApp.GRPCQueryRouter())

	appModule := module.NewAppModule(suite.encCfg.Codec, suite.nftKeeper, suite.accountKeeper, suite.bankKeeper, suite.encCfg.InterfaceRegistry)
	appModule.RegisterServices(cfg)
	appModule.RegisterInterfaces(suite.encCfg.InterfaceRegistry)

}

func (suite *SimTestSuite) TestWeightedOperations() {
	weightedOps := simulation.WeightedOperations(
		suite.encCfg.InterfaceRegistry,
		make(simtypes.AppParams),
		suite.encCfg.Codec,
		suite.accountKeeper,
		suite.bankKeeper,
		suite.nftKeeper,
	)

	// begin new block
	suite.baseApp.BeginBlock(abci.RequestBeginBlock{
		Header: tmproto.Header{
			Height:  suite.baseApp.LastBlockHeight() + 1,
			AppHash: suite.baseApp.LastCommitID().Hash,
		},
	})

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accs := suite.getTestingAccounts(r, 3)

	suite.accountKeeper.EXPECT().GetAccount(suite.ctx, accs[2].Address).Return(authtypes.NewBaseAccount(accs[2].Address, accs[2].PubKey, 0, 0)).Times(1)
	suite.bankKeeper.EXPECT().SpendableCoins(suite.ctx, accs[2].Address).Return(sdk.Coins{sdk.NewInt64Coin("stake", 10)}).Times(1)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{simulation.WeightSend, simulation.TypeMsgSend, simulation.TypeMsgSend},
	}

	for i, w := range weightedOps {
		operationMsg, _, err := w.Op()(r, suite.baseApp, suite.ctx, accs, "")
		suite.Require().NoError(err)

		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		suite.Require().Equal(expected[i].weight, w.Weight(), "weight should be the same")
		suite.Require().Equal(expected[i].opMsgRoute, operationMsg.Route, "route should be the same")
		suite.Require().Equal(expected[i].opMsgName, operationMsg.Name, "operation Msg name should be the same")
	}
}

func (suite *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)
	return accounts
}

func (suite *SimTestSuite) TestSimulateMsgSend() {
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 2)
	blockTime := time.Now().UTC()
	ctx := suite.ctx.WithBlockTime(blockTime)

	acc := authtypes.NewBaseAccount(accounts[0].Address, accounts[0].PubKey, 0, 0)
	suite.accountKeeper.EXPECT().GetAccount(ctx, accounts[0].Address).Return(acc).Times(1)
	suite.bankKeeper.EXPECT().SpendableCoins(ctx, accounts[0].Address).Return(sdk.Coins{sdk.NewInt64Coin("stake", 10)}).Times(1)

	// begin new block
	suite.baseApp.BeginBlock(abci.RequestBeginBlock{
		Header: tmproto.Header{
			Height:  suite.baseApp.LastBlockHeight() + 1,
			AppHash: suite.baseApp.LastCommitID().Hash,
		},
	})

	// execute operation
	registry := suite.encCfg.InterfaceRegistry
	op := simulation.SimulateMsgSend(codec.NewProtoCodec(registry), suite.accountKeeper, suite.bankKeeper, suite.nftKeeper)
	operationMsg, futureOperations, err := op(r, suite.baseApp, ctx, accounts, "")
	suite.Require().NoError(err)

	var msg nft.MsgSend
	suite.encCfg.Codec.UnmarshalJSON(operationMsg.Msg, &msg)
	suite.Require().True(operationMsg.OK)
	suite.Require().Len(futureOperations, 0)
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}
