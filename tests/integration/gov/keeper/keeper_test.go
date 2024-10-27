package keeper_test

import (
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/simapp"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

// KeeperTestSuite only tests gov's keeper logic around tallying, since it
// relies on complex interactions with x/staking.
//
// It also uses simapp (and not a depinjected app) because we manually set a
// new app.StakingKeeper in `createValidators`.
type KeeperTestSuite struct {
	suite.Suite

	app               *simapp.SimApp
	ctx               sdk.Context
	queryClient       v1.QueryClient
	legacyQueryClient v1beta1.QueryClient
	addrs             []sdk.AccAddress
	msgSrvr           v1.MsgServer
	legacyMsgSrvr     v1beta1.MsgServer
}

func (suite *KeeperTestSuite) SetupTest() {
	app := simapp.Setup(suite.T(), false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// Populate the gov account with some coins, as the TestProposal we have
	// is a MsgSend from the gov account.
	coins := sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100000)))
	err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins)
	suite.NoError(err)
	err = app.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, coins)
	suite.NoError(err)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	v1.RegisterQueryServer(queryHelper, app.GovKeeper)
	legacyQueryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	v1beta1.RegisterQueryServer(legacyQueryHelper, keeper.NewLegacyQueryServer(app.GovKeeper))
	queryClient := v1.NewQueryClient(queryHelper)
	legacyQueryClient := v1beta1.NewQueryClient(legacyQueryHelper)

	suite.app = app
	suite.ctx = ctx
	suite.queryClient = queryClient
	suite.legacyQueryClient = legacyQueryClient
	suite.msgSrvr = keeper.NewMsgServerImpl(suite.app.GovKeeper)

	govAcct := suite.app.GovKeeper.GetGovernanceAccount(suite.ctx).GetAddress()
	suite.legacyMsgSrvr = keeper.NewLegacyMsgServerImpl(govAcct.String(), suite.msgSrvr)
	suite.addrs = simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(30000000))
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
