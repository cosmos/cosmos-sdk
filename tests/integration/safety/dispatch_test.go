package safety_test

import (
	"errors"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/simapp"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	safetytypes "github.com/cosmos/cosmos-sdk/x/safety/types"
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
// dispatch fails with the x/safety error while the trip stands and succeeds
// after the reset. x/safety's own tests cover the trip patterns and keeper
// calls on a depinject app.
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

	f.mustDeliver(f.operator, safetytypes.NewMsgReset(f.operator.String(), msgSendURL))
	for caller, err := range p.run(f) {
		require.NoError(t, err, caller)
	}
}
