package safety_test

import (
	"context"
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/safety/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const urlWithdrawReward = "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward"

// blockCtx returns a branch of the Context BaseApp builds for the next block,
// the Context an EVM precompile or a block hook receives. BaseApp attaches the
// circuit breaker to it. Writes to the branch are discarded.
func (f *routerFixture) blockCtx() sdk.Context {
	f.t.Helper()
	ctx := f.app.NewNextBlockContext(cmtproto.Header{Height: f.app.LastBlockHeight() + 1, Time: f.blockTime})
	require.NotNil(f.t, ctx.CircuitBreaker())
	ctx, _ = ctx.CacheContext()
	return ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())
}

// directCalls calls SDK msg servers the way cosmos/evm precompiles do: as Go
// methods, without the MsgServiceRouter.
func (f *routerFixture) directCalls() map[string]func(context.Context) error {
	bank := bankkeeper.NewMsgServerImpl(f.bankKeeper)
	staking := stakingkeeper.NewMsgServerImpl(f.stakingKeeper)
	distr := distrkeeper.NewMsgServerImpl(f.distrKeeper)
	return map[string]func(context.Context) error{
		urlSend: func(ctx context.Context) error {
			_, err := bank.Send(ctx, f.msgSend().(*banktypes.MsgSend))
			return err
		},
		urlMultiSend: func(ctx context.Context) error {
			_, err := bank.MultiSend(ctx, f.msgMultiSend().(*banktypes.MsgMultiSend))
			return err
		},
		urlDelegate: func(ctx context.Context) error {
			_, err := staking.Delegate(ctx, f.msgDelegate().(*stakingtypes.MsgDelegate))
			return err
		},
		urlWithdrawReward: func(ctx context.Context) error {
			_, err := distr.WithdrawDelegatorReward(ctx, distrtypes.NewMsgWithdrawDelegatorReward(f.user.String(), f.validator().String()))
			return err
		},
	}
}

// TestDirectMsgServerCallsCheckTrips shows that a selective trip stops a msg
// server called directly, as an EVM precompile calls it, while the msg servers
// it does not cover still run, and that "*" stops them all.
func TestDirectMsgServerCallsCheckTrips(t *testing.T) {
	f := newRouterFixture(t, false)
	calls := f.directCalls()

	for url, call := range calls {
		require.NoError(t, call(f.blockCtx()), url)
	}

	f.trip(urlSend, urlDelegate)
	for url, call := range calls {
		err := call(f.blockCtx())
		if url == urlSend || url == urlDelegate {
			require.ErrorIs(t, err, types.ErrMsgDisabled, url)
			require.ErrorContains(t, err, url)
		} else {
			require.NoError(t, err, url)
		}
	}

	f.reset(urlSend, urlDelegate)
	f.trip(types.Wildcard)
	for url, call := range calls {
		require.ErrorIs(t, call(f.blockCtx()), types.ErrMsgDisabled, url)
	}
}

// TestDirectGovVotePassesUnderWildcard shows that a direct call of a protected
// msg server, as the gov precompile makes, still runs under "*".
func TestDirectGovVotePassesUnderWildcard(t *testing.T) {
	f := newRouterFixture(t, false)
	f.trip(types.Wildcard)
	id := f.submitProposal(f.msgDistrParams("0.1"))

	gov := govkeeper.NewMsgServerImpl(f.govKeeper)
	_, err := gov.Vote(f.blockCtx(), govv1.NewMsgVote(f.user.addr, id, govv1.OptionNo, ""))
	require.NoError(t, err)
}

// TestRouterDispatchPaysNoGuardLookups shows that the msg server guard adds
// no gas to a msg the router dispatches, since the router marks the msg as
// checked, while a direct call pays for its own lookups.
func TestRouterDispatchPaysNoGuardLookups(t *testing.T) {
	f := newRouterFixture(t, false)
	f.trip(urlDelegate)
	msg := f.msgSend()
	handler := f.app.MsgServiceRouter().Handler(msg)

	routed := func(ctx sdk.Context) uint64 {
		_, err := handler(ctx, msg)
		require.NoError(t, err)
		return ctx.GasMeter().GasConsumed()
	}
	// The router checks with its own breaker either way; only the guard
	// reads the breaker from the Context.
	require.Equal(t, routed(f.blockCtx().WithCircuitBreaker(nil)), routed(f.blockCtx()))

	direct := func(ctx sdk.Context) uint64 {
		require.NoError(t, f.directCalls()[urlSend](ctx))
		return ctx.GasMeter().GasConsumed()
	}
	require.Greater(t, direct(f.blockCtx()), direct(f.blockCtx().WithCircuitBreaker(nil)))
}

// TestDirectMsgServerCallsCheckTripsOnQueries shows that a query Context, on
// which cosmos/evm runs eth_call and eth_estimateGas with their precompiles,
// checks direct msg server calls against the trips at the queried height.
func TestDirectMsgServerCallsCheckTripsOnQueries(t *testing.T) {
	f := newRouterFixture(t, false)
	f.trip(urlSend)
	calls := f.directCalls()

	queryCtx := func() sdk.Context {
		ctx, err := f.app.CreateQueryContext(0, false)
		require.NoError(t, err)
		require.NotNil(t, ctx.CircuitBreaker())
		return ctx
	}
	err := calls[urlSend](queryCtx())
	require.ErrorIs(t, err, types.ErrMsgDisabled)
	require.ErrorContains(t, err, urlSend)
	require.NoError(t, calls[urlDelegate](queryCtx()))
}
