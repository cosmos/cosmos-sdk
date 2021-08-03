package capability_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	"github.com/cosmos/cosmos-sdk/x/capability/keeper"
	"github.com/cosmos/cosmos-sdk/x/capability/types"
)

type CapabilityTestSuite struct {
	suite.Suite

	cdc    codec.Marshaler
	ctx    sdk.Context
	app    *simapp.SimApp
	keeper *keeper.Keeper
	module module.AppModule
}

func (suite *CapabilityTestSuite) SetupTest() {
	checkTx := false
	app := simapp.Setup(checkTx)
	cdc := app.AppCodec()

	// create new keeper so we can define custom scoping before init and seal
	keeper := keeper.NewKeeper(cdc, app.GetKey(types.StoreKey), app.GetMemKey(types.MemStoreKey))

	suite.app = app
	suite.ctx = app.BaseApp.NewContext(checkTx, tmproto.Header{Height: 1})
	suite.keeper = keeper
	suite.cdc = cdc
	suite.module = capability.NewAppModule(cdc, *keeper)
}

// The following test case mocks a specific bug discovered in https://github.com/cosmos/cosmos-sdk/issues/9800
// and ensures that the current code successfully fixes the issue.
func (suite *CapabilityTestSuite) TestInitializeMemStore() {
	sk1 := suite.keeper.ScopeToModule(banktypes.ModuleName)

	cap1, err := sk1.NewCapability(suite.ctx, "transfer")
	suite.Require().NoError(err)
	suite.Require().NotNil(cap1)

	// mock statesync by creating new keeper that shares persistent state but loses in-memory map
	newKeeper := keeper.NewKeeper(suite.cdc, suite.app.GetKey(types.StoreKey), suite.app.GetMemKey("testing"))
	newSk1 := newKeeper.ScopeToModule(banktypes.ModuleName)

	// Mock App startup
	ctx := suite.app.BaseApp.NewUncachedContext(false, tmproto.Header{})
	newKeeper.InitializeAndSeal(ctx)
	suite.Require().False(newKeeper.IsInitialized(ctx), "memstore initialized flag set before BeginBlock")

	// Mock app beginblock and ensure that no gas has been consumed and memstore is initialized
	ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{}).WithBlockGasMeter(sdk.NewGasMeter(50))
	prevGas := ctx.BlockGasMeter().GasConsumed()
	restartedModule := capability.NewAppModule(suite.cdc, *newKeeper)
	restartedModule.BeginBlock(ctx, abci.RequestBeginBlock{})
	suite.Require().True(newKeeper.IsInitialized(ctx), "memstore initialized flag not set")
	gasUsed := ctx.BlockGasMeter().GasConsumed()

	suite.Require().Equal(prevGas, gasUsed, "beginblocker consumed gas during execution")

	// Mock the first transaction getting capability and subsequently failing
	// by using a cached context and discarding all cached writes.
	cacheCtx, _ := ctx.CacheContext()
	_, ok := newSk1.GetCapability(cacheCtx, "transfer")
	suite.Require().True(ok)

	// Ensure that the second transaction can still receive capability even if first tx fails.
	ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{})

	cap1, ok = newSk1.GetCapability(ctx, "transfer")
	suite.Require().True(ok)

	// Ensure the capabilities don't get reinitialized on next BeginBlock
	// by testing to see if capability returns same pointer
	// also check that initialized flag is still set
	restartedModule.BeginBlock(ctx, abci.RequestBeginBlock{})
	recap, ok := newSk1.GetCapability(ctx, "transfer")
	suite.Require().True(ok)
	suite.Require().Equal(cap1, recap, "capabilities got reinitialized after second BeginBlock")
	suite.Require().True(newKeeper.IsInitialized(ctx), "memstore initialized flag not set")
}

func TestCapabilityTestSuite(t *testing.T) {
	suite.Run(t, new(CapabilityTestSuite))
}
