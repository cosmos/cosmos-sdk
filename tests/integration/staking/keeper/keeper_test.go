package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	"cosmossdk.io/simapp"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// fixture uses simapp (and not a depinjected app) because we manually set a
// new app.StakingKeeper in `createValidators`.
type fixture struct {
	app         *simapp.SimApp
	ctx         sdk.Context
	addrs       []sdk.AccAddress
	vals        []types.Validator
	queryClient types.QueryClient
	msgServer   types.MsgServer
	amt1        math.Int
	amt2        math.Int
}

// initFixture uses simapp (and not a depinjected app) because we manually set a
// new app.StakingKeeper in `createValidators` which is used in most of the
// staking keeper tests.
func initFixture(t *testing.T) *fixture {
	f := &fixture{}

	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	querier := keeper.Querier{Keeper: app.StakingKeeper}

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, querier)
	queryClient := types.NewQueryClient(queryHelper)

	f.msgServer = keeper.NewMsgServerImpl(app.StakingKeeper)

	addrs, _, validators := createValidators(t, ctx, app, []int64{9, 8, 7})
	header := tmproto.Header{
		ChainID: "HelloChain",
		Height:  5,
	}

	// sort a copy of the validators, so that original validators does not
	// have its order changed
	sortedVals := make([]types.Validator, len(validators))
	copy(sortedVals, validators)
	hi := types.NewHistoricalInfo(header, sortedVals, app.StakingKeeper.PowerReduction(ctx))
	app.StakingKeeper.SetHistoricalInfo(ctx, 5, &hi)

	f.app, f.ctx, f.queryClient, f.addrs, f.vals = app, ctx, queryClient, addrs, validators

	f.amt1 = f.app.StakingKeeper.TokensFromConsensusPower(f.ctx, 101)
	f.amt2 = f.app.StakingKeeper.TokensFromConsensusPower(f.ctx, 102)

	return f
}
