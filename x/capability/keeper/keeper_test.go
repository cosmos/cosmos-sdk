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

	suite.Require().Panics(func() {
		_ = suite.keeper.ScopeToModule(staking.ModuleName)
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

	got, ok = sk.GetCapability(suite.ctx, "invalid")
	suite.Require().False(ok)
	suite.Require().Nil(got)

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

	forgedCap := types.NewCapabilityKey(0) // index should be the same index as the first capability
	suite.Require().False(sk2.AuthenticateCapability(suite.ctx, forgedCap, "transfer"))

	cap2, err := sk2.NewCapability(suite.ctx, "bond")
	suite.Require().NoError(err)
	suite.Require().NotNil(cap2)

	suite.Require().True(sk1.AuthenticateCapability(suite.ctx, cap1, "transfer"))
	suite.Require().False(sk1.AuthenticateCapability(suite.ctx, cap1, "invalid"))
	suite.Require().False(sk1.AuthenticateCapability(suite.ctx, cap2, "transfer"))

	suite.Require().True(sk2.AuthenticateCapability(suite.ctx, cap2, "bond"))
	suite.Require().False(sk2.AuthenticateCapability(suite.ctx, cap2, "invalid"))
	suite.Require().False(sk2.AuthenticateCapability(suite.ctx, cap1, "bond"))

	badCap := types.NewCapabilityKey(100)
	suite.Require().False(sk1.AuthenticateCapability(suite.ctx, badCap, "transfer"))
	suite.Require().False(sk2.AuthenticateCapability(suite.ctx, badCap, "bond"))
}

func (suite *KeeperTestSuite) TestClaimCapability() {
	sk1 := suite.keeper.ScopeToModule(bank.ModuleName)
	sk2 := suite.keeper.ScopeToModule(staking.ModuleName)

	cap, err := sk1.NewCapability(suite.ctx, "transfer")
	suite.Require().NoError(err)
	suite.Require().NotNil(cap)

	suite.Require().Error(sk1.ClaimCapability(suite.ctx, cap, "transfer"))
	suite.Require().NoError(sk2.ClaimCapability(suite.ctx, cap, "transfer"))

	got, ok := sk1.GetCapability(suite.ctx, "transfer")
	suite.Require().True(ok)
	suite.Require().Equal(cap, got)

	got, ok = sk2.GetCapability(suite.ctx, "transfer")
	suite.Require().True(ok)
	suite.Require().Equal(cap, got)
}

func (suite *KeeperTestSuite) TestReleaseCapability() {
	sk1 := suite.keeper.ScopeToModule(bank.ModuleName)
	sk2 := suite.keeper.ScopeToModule(staking.ModuleName)

	cap1, err := sk1.NewCapability(suite.ctx, "transfer")
	suite.Require().NoError(err)
	suite.Require().NotNil(cap1)

	suite.Require().NoError(sk2.ClaimCapability(suite.ctx, cap1, "transfer"))

	cap2, err := sk2.NewCapability(suite.ctx, "bond")
	suite.Require().NoError(err)
	suite.Require().NotNil(cap2)

	suite.Require().Error(sk1.ReleaseCapability(suite.ctx, cap2))

	suite.Require().NoError(sk2.ReleaseCapability(suite.ctx, cap1))
	got, ok := sk1.GetCapability(suite.ctx, "transfer")
	suite.Require().False(ok)
	suite.Require().Nil(got)

	suite.Require().NoError(sk1.ReleaseCapability(suite.ctx, cap1))
	got, ok = sk2.GetCapability(suite.ctx, "transfer")
	suite.Require().False(ok)
	suite.Require().Nil(got)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
