package keeper_test

import (
	"testing"

	"cosmossdk.io/simapp"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

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
}

// initFixture uses simapp (and not a depinjected app) because we manually set a
// new app.StakingKeeper in `createValidators` which is used in most of the
// staking keeper tests.
func initFixture(t *testing.T) *fixture {
	f := &fixture{}

	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, cmtproto.Header{})

	querier := keeper.Querier{Keeper: app.StakingKeeper}

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, querier)
	queryClient := types.NewQueryClient(queryHelper)

	f.msgServer = keeper.NewMsgServerImpl(app.StakingKeeper)

	addrs, _, validators := createValidators(t, ctx, app, []int64{9, 8, 7})
	header := cmtproto.Header{
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

	return f
}
