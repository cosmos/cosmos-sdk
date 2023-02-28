package keeper_test

import (
	"testing"

	"cosmossdk.io/simapp"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

// fixture only tests gov's keeper logic around tallying, since it
// relies on complex interactions with x/staking.
//
// It also uses simapp (and not a depinjected app) because we manually set a
// new app.StakingKeeper in `createValidators`.
type fixture struct {
	app               *simapp.SimApp
	ctx               sdk.Context
	queryClient       v1.QueryClient
	legacyQueryClient v1beta1.QueryClient
	addrs             []sdk.AccAddress
	msgSrvr           v1.MsgServer
	legacyMsgSrvr     v1beta1.MsgServer
}

// initFixture uses simapp (and not a depinjected app) because we manually set a
// new app.StakingKeeper in `createValidators` which is used in most of the
// gov keeper tests.
func initFixture(t *testing.T) *fixture {
	f := &fixture{}

	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, cmtproto.Header{})

	// Populate the gov account with some coins, as the TestProposal we have
	// is a MsgSend from the gov account.
	coins := sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100000)))
	err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins)
	assert.NilError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, coins)
	assert.NilError(t, err)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	v1.RegisterQueryServer(queryHelper, app.GovKeeper)
	legacyQueryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	v1beta1.RegisterQueryServer(legacyQueryHelper, keeper.NewLegacyQueryServer(app.GovKeeper))
	queryClient := v1.NewQueryClient(queryHelper)
	legacyQueryClient := v1beta1.NewQueryClient(legacyQueryHelper)

	f.app = app
	f.ctx = ctx
	f.queryClient = queryClient
	f.legacyQueryClient = legacyQueryClient
	f.msgSrvr = keeper.NewMsgServerImpl(f.app.GovKeeper)

	govAcct := f.app.GovKeeper.GetGovernanceAccount(f.ctx).GetAddress()
	f.legacyMsgSrvr = keeper.NewLegacyMsgServerImpl(govAcct.String(), f.msgSrvr)
	f.addrs = simtestutil.AddTestAddrsIncremental(app.BankKeeper, app.StakingKeeper, ctx, 2, sdk.NewInt(30000000))

	return f
}
