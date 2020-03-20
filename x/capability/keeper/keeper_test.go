package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/capability"
	"github.com/cosmos/cosmos-sdk/x/capability/keeper"
	"github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
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

	caps := make([]types.Capability, 5)

	for i := range caps {
		cap, err := sk.NewCapability(suite.ctx, fmt.Sprintf("transfer-%d", i))
		suite.Require().NoError(err)
		suite.Require().NotNil(cap)
		suite.Require().Equal(uint64(i), cap.GetIndex())

		caps[i] = cap
	}

	suite.Require().NotPanics(func() {
		suite.keeper.InitializeAndSeal(suite.ctx)
	})

	for i, cap := range caps {
		got, ok := sk.GetCapability(suite.ctx, fmt.Sprintf("transfer-%d", i))
		suite.Require().True(ok)
		suite.Require().Equal(cap, got)
		suite.Require().Equal(uint64(i), got.GetIndex())
	}

	suite.Require().Panics(func() {
		suite.keeper.InitializeAndSeal(suite.ctx)
	})
}

func (suite *KeeperTestSuite) TestNewCapability() {
	sk := suite.keeper.ScopeToModule(bank.ModuleName)

	cap, err := sk.NewCapability(suite.ctx, "transfer")
	suite.Require().NoError(err)
	suite.Require().NotNil(cap)

	got, ok := sk.GetCapability(suite.ctx, "transfer")
	suite.Require().True(ok)
	suite.Require().Equal(cap, got)

	cap, err = sk.NewCapability(suite.ctx, "transfer")
	suite.Require().Error(err)
	suite.Require().Nil(cap)
}

func (suite *KeeperTestSuite) TestAuthenticateCapability() {
	sk1 := suite.keeper.ScopeToModule(bank.ModuleName)
	sk2 := suite.keeper.ScopeToModule(staking.ModuleName)

	cap1, err := sk1.NewCapability(suite.ctx, "transfer")
	suite.Require().NoError(err)
	suite.Require().NotNil(cap1)

	cap2, err := sk2.NewCapability(suite.ctx, "bond")
	suite.Require().NoError(err)
	suite.Require().NotNil(cap2)

	suite.Require().True(sk1.AuthenticateCapability(suite.ctx, cap1, "transfer"))
	suite.Require().False(sk1.AuthenticateCapability(suite.ctx, cap1, "invalid"))
	suite.Require().False(sk1.AuthenticateCapability(suite.ctx, cap2, "transfer"))

	suite.Require().True(sk2.AuthenticateCapability(suite.ctx, cap2, "bond"))
	suite.Require().False(sk2.AuthenticateCapability(suite.ctx, cap2, "invalid"))
	suite.Require().False(sk2.AuthenticateCapability(suite.ctx, cap1, "bond"))
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
