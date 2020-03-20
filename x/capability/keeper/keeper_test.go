package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/capability"
	"github.com/cosmos/cosmos-sdk/x/capability/keeper"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx    sdk.Context
	keeper *keeper.Keeper
	app    *simapp.SimApp
}

func (suite *KeeperTestSuite) SetupTest() {
	checkTx := false
	app := simapp.Setup(checkTx)

	// create new keeper so we can define custom scoping before init and seal
	keeper := keeper.NewKeeper(
		app.Codec(), app.GetKey(capability.StoreKey), app.GetMemKey(capability.MemStoreKey),
	)

	suite.ctx = app.BaseApp.NewContext(checkTx, abci.Header{Height: 1})
	suite.keeper = keeper
	suite.app = app
}

func (suite *KeeperTestSuite) TestInitializeAndSeal() {
	sk := suite.keeper.ScopeToModule(bank.ModuleName)

	cap, err := sk.NewCapability(suite.ctx, "transfer")
	suite.Require().NoError(err)
	suite.Require().NotNil(cap)

	suite.Require().NotPanics(func() {
		suite.keeper.InitializeAndSeal(suite.ctx)
	})

	got, ok := sk.GetCapability(suite.ctx, "transfer")
	suite.Require().True(ok)
	suite.Require().Equal(cap, got)

	suite.Require().Panics(func() {
		suite.keeper.InitializeAndSeal(suite.ctx)
	})
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
