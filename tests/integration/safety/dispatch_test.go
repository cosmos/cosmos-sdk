package safety_test

import (
	"errors"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/simapp"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	safetytypes "github.com/cosmos/cosmos-sdk/x/safety/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// dispatchProbe dispatches a MsgSend from SimApp's PreBlocker, BeginBlocker
// and ante handler while armed, the way an EVM precompile or another module
// would run a bank msg it does not own, and records each result by caller.
// It dispatches on a branch of the caller's context, so it changes no state.
type dispatchProbe struct {
	app   *simapp.SimApp
	msg   *banktypes.MsgSend
	armed bool
	errs  map[string]error
}

func (p *dispatchProbe) dispatch(caller string, ctx sdk.Context) {
	if !p.armed {
		return
	}
	cacheCtx, _ := ctx.CacheContext()
	resp, err := msgservice.Dispatch[*banktypes.MsgSendResponse](cacheCtx, p.app.MsgServiceRouter(), p.msg)
	if err == nil && resp == nil {
		err = errors.New("no response")
	}
	p.errs[caller] = err
}

func (p *dispatchProbe) setup(app *simapp.SimApp) {
	p.app = app
	app.SetPreBlocker(func(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
		p.dispatch("preblock", ctx)
		return app.PreBlocker(ctx, req)
	})
	app.SetBeginBlocker(func(ctx sdk.Context) (sdk.BeginBlock, error) {
		p.dispatch("beginblock", ctx)
		return app.BeginBlocker(ctx)
	})
	ante := app.AnteHandler()
	app.SetAnteHandler(func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		if ctx.ExecMode() == sdk.ExecModeFinalize {
			p.dispatch("finalize tx", ctx)
		}
		return ante(ctx, tx, simulate)
	})
}

// run arms the probe for one block with a tx in it and a query context on
// the committed state, and returns the results.
func (p *dispatchProbe) run(f *fixture) map[string]error {
	f.t.Helper()
	p.errs = map[string]error{}
	p.armed = true
	f.finalize(f.encode(f.other, send(f.other, f.user)))
	queryCtx, err := f.app.CreateQueryContext(f.app.LastBlockHeight(), false)
	require.NoError(f.t, err)
	p.dispatch("query", queryCtx)
	p.armed = false
	require.Len(f.t, p.errs, 4)
	return p.errs
}

// TestDispatchInSimApp trips MsgSend on a SimApp chain and dispatches one
// from the block hooks, a FinalizeBlock tx and a query context. Every
// dispatch fails with the x/safety error while the trip stands, a direct
// bank msg server call does not, and every dispatch succeeds after the
// reset.
func TestDispatchInSimApp(t *testing.T) {
	p := &dispatchProbe{}
	f := newFixtureWithSetup(t, nil, nil, p.setup)
	p.msg = send(f.user, f.other)
	f.setParams(f.operator.String(), "*")

	for caller, err := range p.run(f) {
		require.NoError(t, err, caller)
	}

	f.mustDeliver(f.operator, safetytypes.NewMsgTrip(f.operator.String(), msgSendURL))
	for caller, err := range p.run(f) {
		require.ErrorIs(t, err, safetytypes.ErrMsgDisabled, caller)
	}

	ctx, _ := f.ctx().CacheContext()
	_, err := bankkeeper.NewMsgServerImpl(f.app.BankKeeper).Send(ctx, p.msg)
	require.NoError(t, err, "a direct msg server call skips the trip")

	f.mustDeliver(f.operator, safetytypes.NewMsgReset(f.operator.String(), msgSendURL))
	for caller, err := range p.run(f) {
		require.NoError(t, err, caller)
	}
}

// branch returns a cached branch of ctx with its own gas meter and event
// manager.
func branch(ctx sdk.Context) sdk.Context {
	cacheCtx, _ := ctx.CacheContext()
	return cacheCtx.WithGasMeter(storetypes.NewInfiniteGasMeter()).WithEventManager(sdk.NewEventManager())
}

func (f *fixture) dispatchSend(ctx sdk.Context) (*banktypes.MsgSendResponse, error) {
	return msgservice.Dispatch[*banktypes.MsgSendResponse](ctx, f.app.MsgServiceRouter(), f.msgSend())
}

// TestDispatchBlockedByTrips dispatches a MsgSend under an exact, a prefix
// and a "*" trip. Each fails with the x/safety error, while a direct call
// to the bank msg server under the same trip moves the funds.
func TestDispatchBlockedByTrips(t *testing.T) {
	for _, pattern := range []string{msgSendURL, "/cosmos.bank.*", safetytypes.Wildcard} {
		t.Run(pattern, func(t *testing.T) {
			f := newGateFixture(t, false)
			f.trip(pattern)

			resp, err := f.dispatchSend(branch(f.ctx()))
			require.ErrorIs(t, err, safetytypes.ErrMsgDisabled)
			require.ErrorContains(t, err, msgSendURL)
			require.Nil(t, resp)

			// a direct msg server call skips the router, so the trip does
			// not stop it
			ctx := branch(f.ctx())
			before := f.balanceAt(ctx, f.other)
			_, err = bankkeeper.NewMsgServerImpl(f.app.BankKeeper).Send(ctx, f.msgSend().(*banktypes.MsgSend))
			require.NoError(t, err)
			require.Equal(t, before.AddRaw(1), f.balanceAt(ctx, f.other))
		})
	}
}

// TestDispatchPrefixTripLeavesOtherMsgs dispatches a MsgDelegate under a
// bank prefix trip, which does not cover it.
func TestDispatchPrefixTripLeavesOtherMsgs(t *testing.T) {
	f := newGateFixture(t, false)
	f.trip("/cosmos.bank.*")

	resp, err := msgservice.Dispatch[*stakingtypes.MsgDelegateResponse](branch(f.ctx()), f.app.MsgServiceRouter(), f.msgDelegate())
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestDispatchProtectedUnderWildcard dispatches a protected msg, the
// operator's reset, while "*" stands.
func TestDispatchProtectedUnderWildcard(t *testing.T) {
	f := newGateFixture(t, false)
	f.trip(safetytypes.Wildcard)

	ctx := branch(f.ctx())
	resp, err := msgservice.Dispatch[*safetytypes.MsgResetResponse](ctx, f.app.MsgServiceRouter(), safetytypes.NewMsgReset(f.voter.String(), safetytypes.Wildcard))
	require.NoError(t, err)
	require.NotNil(t, resp)

	has, err := f.app.SafetyKeeper.DisabledList.Has(ctx, safetytypes.Wildcard)
	require.NoError(t, err)
	require.False(t, has)
}

// TestDispatchMatchesRouterPath runs a MsgSend through its router handler
// and through Dispatch on two branches of the same state. Both use the same
// gas, and Dispatch emits the events the router handler returns.
func TestDispatchMatchesRouterPath(t *testing.T) {
	f := newGateFixture(t, false)
	ctx := f.ctx()

	routed := branch(ctx)
	res, err := f.app.MsgServiceRouter().Handler(f.msgSend())(routed, f.msgSend())
	require.NoError(t, err)
	require.NotEmpty(t, res.Events)

	dispatched := branch(ctx)
	resp, err := f.dispatchSend(dispatched)
	require.NoError(t, err)
	require.NotNil(t, resp)

	require.NotZero(t, routed.GasMeter().GasConsumed())
	require.Equal(t, routed.GasMeter().GasConsumed(), dispatched.GasMeter().GasConsumed())
	require.Equal(t, res.Events, dispatched.EventManager().ABCIEvents())
	require.Equal(t, f.balanceAt(routed, f.other), f.balanceAt(dispatched, f.other))
}

// TestDispatchGasMatchesNoBreaker dispatches an allowed MsgSend with x/safety
// as the breaker and runs the same msg through its router handler with no
// breaker set. The breaker check charges no gas, so both use the same gas.
func TestDispatchGasMatchesNoBreaker(t *testing.T) {
	f := newGateFixture(t, false)
	ctx := f.ctx()

	dispatched := branch(ctx)
	_, err := f.dispatchSend(dispatched)
	require.NoError(t, err)

	router := f.app.MsgServiceRouter()
	router.SetCircuit(nil)
	t.Cleanup(func() { router.SetCircuit(f.app.SafetyKeeper) })
	routed := branch(ctx)
	_, err = router.Handler(f.msgSend())(routed, f.msgSend())
	require.NoError(t, err)

	require.NotZero(t, routed.GasMeter().GasConsumed())
	require.Equal(t, routed.GasMeter().GasConsumed(), dispatched.GasMeter().GasConsumed())
}
