package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
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
	app    *simapp.SimApp
	keeper *keeper.Keeper
}

func (suite *KeeperTestSuite) SetupTest() {
	checkTx := false
	app := simapp.Setup(checkTx)
	cdc := codec.NewHybridCodec(app.Codec())

	// create new keeper so we can define custom scoping before init and seal
	keeper := keeper.NewKeeper(cdc, app.GetKey(capability.StoreKey), app.GetMemKey(capability.MemStoreKey))

	suite.app = app
	suite.ctx = app.BaseApp.NewContext(checkTx, abci.Header{Height: 1})
	suite.keeper = keeper
}

func (suite *KeeperTestSuite) TestInitializeAndSeal() {
	sk := suite.keeper.ScopeToModule(bank.ModuleName)

	caps := make([]*types.Capability, 5)
	// Get Latest Index before creating new ones to sychronize indices correctly
	prevIndex := suite.keeper.GetLatestIndex(suite.ctx)

	for i := range caps {
		cap, err := sk.NewCapability(suite.ctx, fmt.Sprintf("transfer-%d", i))
		suite.Require().NoError(err)
		suite.Require().NotNil(cap)
		suite.Require().Equal(uint64(i)+prevIndex, cap.GetIndex())

		caps[i] = cap
	}

	suite.Require().NotPanics(func() {
		suite.keeper.InitializeAndSeal(suite.ctx)
	})

	for i, cap := range caps {
		got, ok := sk.GetCapability(suite.ctx, fmt.Sprintf("transfer-%d", i))
		suite.Require().True(ok)
		suite.Require().Equal(cap, got)
		suite.Require().Equal(uint64(i)+prevIndex, got.GetIndex())
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
	suite.Require().True(cap == got, "expected memory addresses to be equal")

	got, ok = sk.GetCapability(suite.ctx, "invalid")
	suite.Require().False(ok)
	suite.Require().Nil(got)

	cap, err = sk.NewCapability(suite.ctx, "transfer")
	suite.Require().Error(err)
	suite.Require().Nil(cap)
}

func (suite *KeeperTestSuite) TestOriginalCapabilityKeeper() {
	got, ok := suite.app.ScopedIBCKeeper.GetCapability(suite.ctx, "invalid")
	suite.Require().False(ok)
	suite.Require().Nil(got)

	port, ok := suite.app.ScopedIBCKeeper.GetCapability(suite.ctx, "ports/transfer")
	suite.Require().True(ok)
	suite.Require().NotNil(port)
}

func (suite *KeeperTestSuite) TestAuthenticateCapability() {
	sk1 := suite.keeper.ScopeToModule(bank.ModuleName)
	sk2 := suite.keeper.ScopeToModule(staking.ModuleName)

	cap1, err := sk1.NewCapability(suite.ctx, "transfer")
	suite.Require().NoError(err)
	suite.Require().NotNil(cap1)

	forgedCap := types.NewCapability(0) // index should be the same index as the first capability
	suite.Require().False(sk2.AuthenticateCapability(suite.ctx, forgedCap, "transfer"))

	cap2, err := sk2.NewCapability(suite.ctx, "bond")
	suite.Require().NoError(err)
	suite.Require().NotNil(cap2)

	got, ok := sk1.GetCapability(suite.ctx, "transfer")
	suite.Require().True(ok)

	suite.Require().True(sk1.AuthenticateCapability(suite.ctx, cap1, "transfer"))
	suite.Require().True(sk1.AuthenticateCapability(suite.ctx, got, "transfer"))
	suite.Require().False(sk1.AuthenticateCapability(suite.ctx, cap1, "invalid"))
	suite.Require().False(sk1.AuthenticateCapability(suite.ctx, cap2, "transfer"))

	suite.Require().True(sk2.AuthenticateCapability(suite.ctx, cap2, "bond"))
	suite.Require().False(sk2.AuthenticateCapability(suite.ctx, cap2, "invalid"))
	suite.Require().False(sk2.AuthenticateCapability(suite.ctx, cap1, "bond"))

	sk2.ReleaseCapability(suite.ctx, cap2)
	suite.Require().False(sk2.AuthenticateCapability(suite.ctx, cap2, "bond"))

	badCap := types.NewCapability(100)
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

func (suite *KeeperTestSuite) TestGetOwners() {
	sk1 := suite.keeper.ScopeToModule(bank.ModuleName)
	sk2 := suite.keeper.ScopeToModule(staking.ModuleName)
	sk3 := suite.keeper.ScopeToModule("foo")

	sks := []keeper.ScopedKeeper{sk1, sk2, sk3}

	cap, err := sk1.NewCapability(suite.ctx, "transfer")
	suite.Require().NoError(err)
	suite.Require().NotNil(cap)

	suite.Require().NoError(sk2.ClaimCapability(suite.ctx, cap, "transfer"))
	suite.Require().NoError(sk3.ClaimCapability(suite.ctx, cap, "transfer"))

	expectedOrder := []string{bank.ModuleName, "foo", staking.ModuleName}
	// Ensure all scoped keepers can get owners
	for _, sk := range sks {
		owners, ok := sk.GetOwners(suite.ctx, "transfer")
		mods, cap, mok := sk.LookupModules(suite.ctx, "transfer")

		suite.Require().True(ok, "could not retrieve owners")
		suite.Require().NotNil(owners, "owners is nil")

		suite.Require().True(mok, "could not retrieve modules")
		suite.Require().NotNil(cap, "capability is nil")
		suite.Require().NotNil(mods, "modules is nil")

		suite.Require().Equal(len(expectedOrder), len(owners.Owners), "length of owners is unexpected")
		for i, o := range owners.Owners {
			// Require owner is in expected position
			suite.Require().Equal(expectedOrder[i], o.Module, "module is unexpected")
			suite.Require().Equal(expectedOrder[i], mods[i], "module in lookup is unexpected")
		}
	}

	// foo module releases capability
	err = sk3.ReleaseCapability(suite.ctx, cap)
	suite.Require().Nil(err, "could not release capability")

	// new expected order and scoped capabilities
	expectedOrder = []string{bank.ModuleName, staking.ModuleName}
	sks = []keeper.ScopedKeeper{sk1, sk2}

	// Ensure all scoped keepers can get owners
	for _, sk := range sks {
		owners, ok := sk.GetOwners(suite.ctx, "transfer")
		mods, cap, mok := sk.LookupModules(suite.ctx, "transfer")

		suite.Require().True(ok, "could not retrieve owners")
		suite.Require().NotNil(owners, "owners is nil")

		suite.Require().True(mok, "could not retrieve modules")
		suite.Require().NotNil(cap, "capability is nil")
		suite.Require().NotNil(mods, "modules is nil")

		suite.Require().Equal(len(expectedOrder), len(owners.Owners), "length of owners is unexpected")
		for i, o := range owners.Owners {
			// Require owner is in expected position
			suite.Require().Equal(expectedOrder[i], o.Module, "module is unexpected")
			suite.Require().Equal(expectedOrder[i], mods[i], "module in lookup is unexpected")
		}
	}

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
	got, ok := sk2.GetCapability(suite.ctx, "transfer")
	suite.Require().False(ok)
	suite.Require().Nil(got)

	suite.Require().NoError(sk1.ReleaseCapability(suite.ctx, cap1))
	got, ok = sk1.GetCapability(suite.ctx, "transfer")
	suite.Require().False(ok)
	suite.Require().Nil(got)
}

func (suite KeeperTestSuite) TestRevertCapability() {
	sk := suite.keeper.ScopeToModule(bank.ModuleName)

	ms := suite.ctx.MultiStore()

	msCache := ms.CacheMultiStore()
	cacheCtx := suite.ctx.WithMultiStore(msCache)

	capName := "revert"
	// Create capability on cached context
	cap, err := sk.NewCapability(cacheCtx, capName)
	suite.Require().NoError(err, "could not create capability")

	// Check that capability written in cached context
	gotCache, ok := sk.GetCapability(cacheCtx, capName)
	suite.Require().True(ok, "could not retrieve capability from cached context")
	suite.Require().Equal(cap, gotCache, "did not get correct capability from cached context")

	// Check that capability is NOT written to original context
	got, ok := sk.GetCapability(suite.ctx, capName)
	suite.Require().False(ok, "retrieved capability from original context before write")
	suite.Require().Nil(got, "capability not nil in original store")

	// Write to underlying memKVStore
	msCache.Write()

	got, ok = sk.GetCapability(suite.ctx, capName)
	suite.Require().True(ok, "could not retrieve capability from context")
	suite.Require().Equal(cap, got, "did not get correct capability from context")
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
