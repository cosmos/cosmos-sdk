package baseapp_test

import (
	"context"
	"fmt"
	"maps"
	"sync"
	"sync/atomic"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/baseapp"
	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"
	"github.com/cosmos/cosmos-sdk/client"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// guardLookupGas is the gas meteredBreaker charges for one lookup.
const guardLookupGas = 1_000

var (
	counterURL  = sdk.MsgTypeURL(&baseapptestutil.MsgCounter{})
	counter2URL = sdk.MsgTypeURL(&baseapptestutil.MsgCounter2{})
)

// errGuardBreakerStore stands for a circuit breaker that cannot read its
// state.
var errGuardBreakerStore = errorsmod.Wrap(sdkerrors.ErrIO, "circuit breaker store unavailable")

// meteredBreaker charges guardLookupGas to the Context gas meter for every
// lookup, as a store-backed breaker such as x/safety pays for its reads, and
// counts lookups per type URL. It blocks the type URLs in blocked and returns
// errGuardBreakerStore for those in failing. With allowedOnErr set, it
// reports a failing type URL as allowed alongside the error, as
// contrib/x/circuit does when its store read fails.
type meteredBreaker struct {
	blocked      map[string]bool
	failing      map[string]bool
	allowedOnErr bool
	calls        map[string]int
}

func (b *meteredBreaker) IsAllowed(ctx context.Context, typeURL string) (bool, error) {
	sdk.UnwrapSDKContext(ctx).GasMeter().ConsumeGas(guardLookupGas, "circuit breaker lookup")
	b.calls[typeURL]++
	if b.failing[typeURL] {
		return b.allowedOnErr, errGuardBreakerStore
	}
	return !b.blocked[typeURL], nil
}

func newMeteredBreaker() *meteredBreaker {
	return &meteredBreaker{blocked: map[string]bool{}, failing: map[string]bool{}, calls: map[string]int{}}
}

// guardProbeKey is a Context value key a test sets to see that a value of
// the caller's Context reaches the msg server.
type guardProbeKey struct{}

// meteredCounterServer handles MsgCounter by writing the counter to capKey1,
// and records the type URL the router marked as checked, the
// guardProbeKey value, and whether its ctx is an sdk.Context on each call. With
// guard set it calls msgservice.CheckMsgAllowed first, as SDK msg servers do.
// Without it, it runs as msg servers ran before the guard existed.
type meteredCounterServer struct {
	guard  bool
	runs   *int
	marks  *[]string
	probes *[]any
	plain  *[]bool
}

func (s meteredCounterServer) IncrementCounter(ctx context.Context, msg *baseapptestutil.MsgCounter) (*baseapptestutil.MsgCreateCounterResponse, error) {
	*s.marks = append(*s.marks, sdk.UnwrapSDKContext(ctx).CheckedMsgTypeURL())
	*s.probes = append(*s.probes, ctx.Value(guardProbeKey{}))
	_, isSDKCtx := ctx.(sdk.Context)
	*s.plain = append(*s.plain, isSDKCtx)
	if s.guard {
		if err := msgservice.CheckMsgAllowed(ctx, msg); err != nil {
			return nil, err
		}
	}
	*s.runs++
	sdk.UnwrapSDKContext(ctx).KVStore(capKey1).Set([]byte("counter"), []byte{byte(msg.Counter)})
	return &baseapptestutil.MsgCreateCounterResponse{}, nil
}

// meteredRelayServer handles MsgCounter2 the way a module msg server reaches
// another module. With Counter 1 it dispatches a MsgCounter through the
// MsgServiceRouter, as authz, gov and group dispatch inner msgs. With Counter
// 2 it calls the MsgCounter server directly, as a module or an EVM precompile
// can.
type meteredRelayServer struct {
	guard  bool
	router *baseapp.MsgServiceRouter
	inner  meteredCounterServer
}

func (s meteredRelayServer) IncrementCounter(ctx context.Context, msg *baseapptestutil.MsgCounter2) (*baseapptestutil.MsgCreateCounterResponse, error) {
	if s.guard {
		if err := msgservice.CheckMsgAllowed(ctx, msg); err != nil {
			return nil, err
		}
	}
	inner := &baseapptestutil.MsgCounter{Counter: msg.Counter}
	switch msg.Counter {
	case 1:
		if _, err := s.router.Handler(inner)(sdk.UnwrapSDKContext(ctx), inner); err != nil {
			return nil, err
		}
	case 2:
		if _, err := s.inner.IncrementCounter(ctx, inner); err != nil {
			return nil, err
		}
	}
	return &baseapptestutil.MsgCreateCounterResponse{}, nil
}

// guardApp is a BaseApp with a meteredBreaker and the metered msg servers,
// either all guarded or all unguarded.
type guardApp struct {
	t        *testing.T
	app      *baseapp.BaseApp
	txConfig client.TxConfig
	cb       *meteredBreaker
	inner    meteredCounterServer
	runs     *int
	marks    *[]string
}

// newGuardApp returns a guardApp after InitChain, with vote extensions
// enabled from the first height. opts run after the breaker is set.
func newGuardApp(t *testing.T, guard bool, opts ...func(*baseapp.BaseApp)) *guardApp {
	t.Helper()
	g := newLoadedGuardApp(t, guard, opts...)
	_, err := g.app.InitChain(&abci.RequestInitChain{
		InitialHeight:   1,
		ConsensusParams: &cmtproto.ConsensusParams{Abci: &cmtproto.ABCIParams{VoteExtensionsEnableHeight: 1}},
	})
	require.NoError(t, err)
	return g
}

// newLoadedGuardApp returns a guardApp that has loaded its latest version but
// not run InitChain, as a node does when it restarts, so its check state is
// the one BaseApp.Init sets.
func newLoadedGuardApp(t *testing.T, guard bool, opts ...func(*baseapp.BaseApp)) *guardApp {
	t.Helper()
	cb := newMeteredBreaker()
	opts = append([]func(*baseapp.BaseApp){func(app *baseapp.BaseApp) { app.SetCircuitBreaker(cb) }}, opts...)
	suite := NewBaseAppSuite(t, opts...)
	runs := new(int)
	marks := &[]string{}
	inner := meteredCounterServer{guard: guard, runs: runs, marks: marks, probes: &[]any{}, plain: &[]bool{}}
	router := suite.baseApp.MsgServiceRouter()
	baseapptestutil.RegisterCounterServer(router, inner)
	baseapptestutil.RegisterCounter2Server(router, meteredRelayServer{guard: guard, router: router, inner: inner})
	return &guardApp{t: t, app: suite.baseApp, txConfig: suite.txConfig, cb: cb, inner: inner, runs: runs, marks: marks}
}

// encodeTx returns an encoded tx of msgs.
func (g *guardApp) encodeTx(msgs ...sdk.Msg) []byte {
	g.t.Helper()
	builder := g.txConfig.NewTxBuilder()
	require.NoError(g.t, builder.SetMsgs(msgs...))
	setTxSignature(g.t, builder, 0)
	bz, err := g.txConfig.TxEncoder()(builder.GetTx())
	require.NoError(g.t, err)
	return bz
}

// guardRun is what one exec mode reports for a tx, and the lookups and msg
// server runs it made.
type guardRun struct {
	gas       uint64
	codespace string
	code      uint32
	calls     map[string]int
	runs      int
}

func (g *guardApp) reset() {
	g.cb.calls = map[string]int{}
	*g.runs = 0
	*g.marks = nil
}

func (g *guardApp) record(gas uint64, codespace string, code uint32) guardRun {
	return guardRun{gas: gas, codespace: codespace, code: code, calls: maps.Clone(g.cb.calls), runs: *g.runs}
}

// simulate runs bz in Simulate, which leaves the app state as it was.
func (g *guardApp) simulate(bz []byte) (guardRun, error) {
	g.t.Helper()
	g.reset()
	gasInfo, _, err := g.app.Simulate(bz)
	codespace, code, _ := errorsmod.ABCIInfo(err, false)
	return g.record(gasInfo.GasUsed, codespace, code), err
}

// finalize runs bz in the app's first block.
func (g *guardApp) finalize(bz []byte) guardRun {
	g.t.Helper()
	g.reset()
	res, err := g.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1, Txs: [][]byte{bz}})
	require.NoError(g.t, err)
	require.Len(g.t, res.TxResults, 1)
	tx := res.TxResults[0]
	return g.record(uint64(tx.GasUsed), tx.Codespace, tx.Code)
}

// runBoth runs bz in Simulate and then in FinalizeBlock, requires both to
// report the same gas, result, lookups and msg server runs, and returns the
// FinalizeBlock run and the Simulate error.
func (g *guardApp) runBoth(bz []byte) (guardRun, error) {
	g.t.Helper()
	sim, simErr := g.simulate(bz)
	fin := g.finalize(bz)
	require.Equal(g.t, sim, fin, "Simulate and FinalizeBlock must match")
	return fin, simErr
}

func guardSigner() string {
	_, _, addr := testdata.KeyTestPubAddr()
	return addr.String()
}

func TestCircuitBreakerOnStateContexts(t *testing.T) {
	g := newGuardApp(t, true)

	require.Equal(t, sdk.CircuitBreaker(g.cb), getFinalizeBlockStateCtx(g.app).CircuitBreaker())

	plain := NewBaseAppSuite(t).baseApp
	require.Nil(t, getCheckStateCtx(plain).CircuitBreaker())
}

// TestCircuitBreakerOnTestHelperContexts shows that the Contexts BaseApp's
// test helpers return have the breaker, as the Contexts BaseApp builds for
// txs do, so a test that calls a msg server on one exercises the guard.
func TestCircuitBreakerOnTestHelperContexts(t *testing.T) {
	g := newGuardApp(t, true)
	g.cb.blocked[counterURL] = true
	plain := NewBaseAppSuite(t).baseApp
	_, err := plain.InitChain(&abci.RequestInitChain{InitialHeight: 1})
	require.NoError(t, err)
	header := cmtproto.Header{Height: 2}

	helpers := map[string]func(app *baseapp.BaseApp) sdk.Context{
		"NewContext check":           func(app *baseapp.BaseApp) sdk.Context { return app.NewContext(true) },
		"NewContext finalize":        func(app *baseapp.BaseApp) sdk.Context { return app.NewContext(false) },
		"NewContextLegacy":           func(app *baseapp.BaseApp) sdk.Context { return app.NewContextLegacy(false, header) },
		"NewUncachedContext":         func(app *baseapp.BaseApp) sdk.Context { return app.NewUncachedContext(false, header) },
		"NewNextBlockContext":        func(app *baseapp.BaseApp) sdk.Context { return app.NewNextBlockContext(header) },
		"GetContextForCheckTx":       func(app *baseapp.BaseApp) sdk.Context { return app.GetContextForCheckTx(nil) },
		"GetContextForFinalizeBlock": func(app *baseapp.BaseApp) sdk.Context { return app.GetContextForFinalizeBlock(nil) },
	}
	for name, newCtx := range helpers {
		t.Run(name, func(t *testing.T) {
			ctx := newCtx(g.app)
			require.Equal(t, sdk.CircuitBreaker(g.cb), ctx.CircuitBreaker())
			require.ErrorIs(t, g.directCallErr(ctx), sdkerrors.ErrUnauthorized)

			require.Nil(t, newCtx(plain).CircuitBreaker())
		})
	}
}

// TestCircuitBreakerOnQueryContext covers queries that execute msg servers,
// such as an EVM eth_call or gas estimate whose precompile calls bank Send.
func TestCircuitBreakerOnQueryContext(t *testing.T) {
	g := newGuardApp(t, true)
	_, err := g.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.NoError(t, err)
	_, err = g.app.Commit()
	require.NoError(t, err)

	abciCtx, err := g.app.CreateQueryContext(0, false)
	require.NoError(t, err)
	grpcCtx, err := g.app.CreateQueryContextWithCheckHeader(0, false, false)
	require.NoError(t, err)

	msg := &baseapptestutil.MsgCounter{Counter: 1}
	g.cb.blocked[counterURL] = true
	for name, ctx := range map[string]sdk.Context{"abci": abciCtx, "grpc": grpcCtx} {
		require.Equal(t, sdk.CircuitBreaker(g.cb), ctx.CircuitBreaker(), name)
		_, err = g.inner.IncrementCounter(ctx, msg)
		require.ErrorIs(t, err, sdkerrors.ErrUnauthorized, name)
	}
	require.Zero(t, *g.runs)

	plain := NewBaseAppSuite(t).baseApp
	_, err = plain.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1})
	require.NoError(t, err)
	_, err = plain.Commit()
	require.NoError(t, err)
	plainCtx, err := plain.CreateQueryContext(0, false)
	require.NoError(t, err)
	require.Nil(t, plainCtx.CircuitBreaker())
}

// TestRouterMarksCheckedMsg shows that the router marks the msg it
// dispatches, and only that msg: an inner msg dispatched through the router
// gets its own mark, while a msg server called directly from inside a routed
// msg sees the outer mark, so its guard runs the lookup.
func TestRouterMarksCheckedMsg(t *testing.T) {
	testCases := []struct {
		name    string
		msg     sdk.Msg
		expMark string
	}{
		{name: "top-level msg", msg: &baseapptestutil.MsgCounter{Counter: 1}, expMark: counterURL},
		{name: "inner msg through the router", msg: &baseapptestutil.MsgCounter2{Counter: 1}, expMark: counterURL},
		{name: "inner msg server called directly", msg: &baseapptestutil.MsgCounter2{Counter: 2}, expMark: counter2URL},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := newGuardApp(t, true)
			ctx := getFinalizeBlockStateCtx(g.app)

			_, err := g.app.MsgServiceRouter().Handler(tc.msg)(ctx, tc.msg)
			require.NoError(t, err)
			require.Equal(t, []string{tc.expMark}, *g.marks)
			require.Equal(t, 1, g.cb.calls[counterURL], "the MsgCounter server must not repeat the router lookup")

			g.reset()
			g.cb.blocked[counterURL] = true
			_, err = g.app.MsgServiceRouter().Handler(tc.msg)(ctx, tc.msg)
			require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)
			require.Zero(t, *g.runs, "the blocked msg server ran")
		})
	}
}

func TestRouterWithoutBreakerClearsMark(t *testing.T) {
	g := newGuardApp(t, true)
	cdc := codectestutil.CodecOptions{}.NewCodec()
	baseapptestutil.RegisterInterfaces(cdc.InterfaceRegistry())
	router := baseapp.NewMsgServiceRouter()
	router.SetInterfaceRegistry(cdc.InterfaceRegistry())
	baseapptestutil.RegisterCounterServer(router, g.inner)

	ctx := getFinalizeBlockStateCtx(g.app).WithCheckedMsgTypeURL(counterURL)
	msg := &baseapptestutil.MsgCounter{Counter: 1}

	_, err := router.Handler(msg)(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, []string{""}, *g.marks)
	require.Equal(t, 1, g.cb.calls[counterURL], "a mark set outside the router must not skip the check")
}

// TestGuardRoutedGasMatchesUnguardedServers shows that guarded msg servers
// charge exactly the gas unguarded ones do, in Simulate and in FinalizeBlock,
// for a top-level msg and for msgs a wrapper dispatches through the router.
// The router checks each msg once either way.
func TestGuardRoutedGasMatchesUnguardedServers(t *testing.T) {
	signer := guardSigner()
	testCases := []struct {
		name     string
		msgs     []sdk.Msg
		expCalls map[string]int
		expRuns  int
	}{
		{
			name:     "top-level msg",
			msgs:     []sdk.Msg{&baseapptestutil.MsgCounter{Counter: 1, Signer: signer}},
			expCalls: map[string]int{counterURL: 1},
			expRuns:  1,
		},
		{
			name:     "inner msg dispatched through the router",
			msgs:     []sdk.Msg{&baseapptestutil.MsgCounter2{Counter: 1, Signer: signer}},
			expCalls: map[string]int{counter2URL: 1, counterURL: 1},
			expRuns:  1,
		},
		{
			name: "wrappers around a top-level msg",
			msgs: []sdk.Msg{
				&baseapptestutil.MsgCounter2{Counter: 1, Signer: signer},
				&baseapptestutil.MsgCounter{Counter: 3, Signer: signer},
				&baseapptestutil.MsgCounter2{Counter: 1, Signer: signer},
			},
			expCalls: map[string]int{counter2URL: 2, counterURL: 3},
			expRuns:  3,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			guarded := newGuardApp(t, true)
			plain := newGuardApp(t, false)
			bz := guarded.encodeTx(tc.msgs...)

			got, err := guarded.runBoth(bz)
			require.NoError(t, err)
			want, err := plain.runBoth(bz)
			require.NoError(t, err)

			require.Zero(t, got.code)
			require.Equal(t, tc.expCalls, got.calls)
			require.Equal(t, tc.expRuns, got.runs)
			require.Equal(t, want, got, "the guard must add no lookups and no gas to routed msgs")
		})
	}
}

// TestGuardDirectCallPaysOneLookup shows that a msg server called directly
// from another msg server pays for exactly one lookup, in Simulate and in
// FinalizeBlock, which unguarded servers skip.
func TestGuardDirectCallPaysOneLookup(t *testing.T) {
	guarded := newGuardApp(t, true)
	plain := newGuardApp(t, false)
	bz := guarded.encodeTx(&baseapptestutil.MsgCounter2{Counter: 2, Signer: guardSigner()})

	got, err := guarded.runBoth(bz)
	require.NoError(t, err)
	want, err := plain.runBoth(bz)
	require.NoError(t, err)

	require.Equal(t, map[string]int{counter2URL: 1, counterURL: 1}, got.calls)
	require.Equal(t, map[string]int{counter2URL: 1}, want.calls)
	require.Equal(t, want.gas+guardLookupGas, got.gas)
	require.Equal(t, 1, got.runs)
}

// TestGuardMarkDoesNotLeakIntoDirectCall blocks MsgCounter and shows that the
// router's mark for an outer MsgCounter2 does not let a direct call of the
// MsgCounter server through, in Simulate and in FinalizeBlock. Unguarded
// servers run the blocked msg, which is the gap the guard closes.
func TestGuardMarkDoesNotLeakIntoDirectCall(t *testing.T) {
	signer := guardSigner()
	direct := &baseapptestutil.MsgCounter2{Counter: 2, Signer: signer}
	routed := &baseapptestutil.MsgCounter2{Counter: 1, Signer: signer}

	guarded := newGuardApp(t, true)
	guarded.cb.blocked[counterURL] = true
	got, err := guarded.runBoth(guarded.encodeTx(direct))
	require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)
	require.ErrorContains(t, err, counterURL)
	require.Equal(t, sdkerrors.ErrUnauthorized.ABCICode(), got.code)
	require.Equal(t, sdkerrors.ErrUnauthorized.Codespace(), got.codespace)
	require.Equal(t, map[string]int{counter2URL: 1, counterURL: 1}, got.calls)
	require.Zero(t, got.runs, "the blocked msg server ran")

	plain := newGuardApp(t, false)
	plain.cb.blocked[counterURL] = true
	got, err = plain.runBoth(plain.encodeTx(direct))
	require.NoError(t, err)
	require.Equal(t, 1, got.runs)

	// a msg dispatched through the router is refused with or without the guard
	for _, g := range []*guardApp{newGuardApp(t, true), newGuardApp(t, false)} {
		g.cb.blocked[counterURL] = true
		got, err = g.runBoth(g.encodeTx(routed))
		require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)
		require.Equal(t, map[string]int{counter2URL: 1, counterURL: 1}, got.calls)
		require.Zero(t, got.runs)
	}
}

// TestGuardFailsClosedOnBreakerError shows that the guard and the router
// refuse a msg when the breaker returns an error for it, and return that
// error, whether the breaker reports the msg as blocked or as allowed. It
// covers a direct call outside any tx, a direct call inside a hybrid handler,
// a direct call inside a routed msg in Simulate and FinalizeBlock, and the
// router's own check of a top-level msg and of a hybrid handler call.
func TestGuardFailsClosedOnBreakerError(t *testing.T) {
	for _, allowedOnErr := range []bool{false, true} {
		t.Run(fmt.Sprintf("allowed %t", allowedOnErr), func(t *testing.T) {
			g := newGuardApp(t, true)
			g.cb.allowedOnErr = allowedOnErr
			g.cb.failing[counterURL] = true
			resp := &baseapptestutil.MsgCreateCounterResponse{}

			_, err := g.inner.IncrementCounter(getFinalizeBlockStateCtx(g.app), &baseapptestutil.MsgCounter{Counter: 1})
			require.ErrorIs(t, err, errGuardBreakerStore)
			require.Zero(t, *g.runs)

			direct := &baseapptestutil.MsgCounter2{Counter: 2, Signer: guardSigner()}
			hybrid := g.app.MsgServiceRouter().HybridHandlerByMsgName("MsgCounter2")
			require.NotNil(t, hybrid)
			err = hybrid(getFinalizeBlockStateCtx(g.app), direct, resp)
			require.ErrorIs(t, err, errGuardBreakerStore)
			require.Zero(t, *g.runs)

			routedHybrid := g.app.MsgServiceRouter().HybridHandlerByMsgName("MsgCounter")
			require.NotNil(t, routedHybrid)
			err = routedHybrid(getFinalizeBlockStateCtx(g.app), &baseapptestutil.MsgCounter{Counter: 1}, resp)
			require.ErrorIs(t, err, errGuardBreakerStore)
			require.Zero(t, *g.runs)

			got, err := g.runBoth(g.encodeTx(direct))
			require.ErrorIs(t, err, errGuardBreakerStore)
			require.Equal(t, sdkerrors.ErrIO.Codespace(), got.codespace)
			require.Equal(t, sdkerrors.ErrIO.ABCICode(), got.code)
			require.Equal(t, map[string]int{counter2URL: 1, counterURL: 1}, got.calls)
			require.Zero(t, got.runs)

			got, err = g.runBoth(g.encodeTx(&baseapptestutil.MsgCounter{Counter: 1, Signer: guardSigner()}))
			require.ErrorIs(t, err, errGuardBreakerStore)
			require.Equal(t, sdkerrors.ErrIO.ABCICode(), got.code)
			require.Equal(t, map[string]int{counterURL: 1}, got.calls)
			require.Zero(t, got.runs, "the router let the msg through")
		})
	}
}

// TestGuardSimulateDuringCommit runs Simulate while blocks commit, as gRPC
// Simulate does alongside consensus, and requires every run to refuse a
// blocked msg server called directly. Commit replaces the check state that
// Simulate reads, so the new state must have its breaker before Simulate can
// see it.
func TestGuardSimulateDuringCommit(t *testing.T) {
	g := newGuardApp(t, true)
	g.cb.blocked[counterURL] = true
	bz := g.encodeTx(&baseapptestutil.MsgCounter2{Counter: 2, Signer: guardSigner()})
	_, _, err := g.app.Simulate(bz)
	require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)

	var (
		stop    atomic.Bool
		allowed atomic.Int64
		wg      sync.WaitGroup
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for !stop.Load() {
			if _, _, err := g.app.Simulate(bz); err == nil {
				allowed.Add(1)
			}
		}
	}()
	for h := int64(1); h <= 5_000; h++ {
		_, err := g.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: h})
		require.NoError(t, err)
		_, err = g.app.Commit()
		require.NoError(t, err)
	}
	stop.Store(true)
	wg.Wait()
	require.Zero(t, allowed.Load(), "Simulate ran a blocked msg server")
}

// TestGuardSimulateAfterRestart covers the check state BaseApp.Init sets, which
// Simulate reads on a restarted node until its first Commit.
func TestGuardSimulateAfterRestart(t *testing.T) {
	g := newLoadedGuardApp(t, true)
	require.Equal(t, sdk.CircuitBreaker(g.cb), getCheckStateCtx(g.app).CircuitBreaker())

	g.cb.blocked[counterURL] = true
	_, _, err := g.app.Simulate(g.encodeTx(&baseapptestutil.MsgCounter2{Counter: 2, Signer: guardSigner()}))
	require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)
	require.Zero(t, *g.runs)
}

// commitBlock finalizes and commits an empty block at height.
func (g *guardApp) commitBlock(height int64) {
	g.t.Helper()
	_, err := g.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: height})
	require.NoError(g.t, err)
	_, err = g.app.Commit()
	require.NoError(g.t, err)
}

// directCallErr calls the guarded MsgCounter server directly on ctx, as a
// handler that reaches a precompile would, and returns its error.
func (g *guardApp) directCallErr(ctx sdk.Context) error {
	_, err := g.inner.IncrementCounter(ctx, &baseapptestutil.MsgCounter{Counter: 1})
	return err
}

// TestGuardInVoteExtensionHandlers shows that a msg server called directly
// from the ExtendVote and VerifyVoteExtension handlers is checked against the
// breaker at the initial height and at every later one, where BaseApp builds
// the handler Context from the committed state.
func TestGuardInVoteExtensionHandlers(t *testing.T) {
	var (
		g                 *guardApp
		extendErr, verErr error
	)
	g = newGuardApp(t, true, func(app *baseapp.BaseApp) {
		app.SetExtendVoteHandler(func(ctx sdk.Context, _ *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
			extendErr = g.directCallErr(ctx)
			return &abci.ResponseExtendVote{VoteExtension: []byte("ext")}, nil
		})
		app.SetVerifyVoteExtensionHandler(func(ctx sdk.Context, _ *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
			verErr = g.directCallErr(ctx)
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
		})
	})
	g.cb.blocked[counterURL] = true

	for height := int64(1); height <= 3; height++ {
		extendErr, verErr = nil, nil
		_, err := g.app.ExtendVote(context.Background(), &abci.RequestExtendVote{Height: height})
		require.NoError(t, err)
		_, err = g.app.VerifyVoteExtension(&abci.RequestVerifyVoteExtension{Height: height, VoteExtension: []byte("ext")})
		require.NoError(t, err)
		require.ErrorIs(t, extendErr, sdkerrors.ErrUnauthorized, "ExtendVote at height %d", height)
		require.ErrorIs(t, verErr, sdkerrors.ErrUnauthorized, "VerifyVoteExtension at height %d", height)
		require.Zero(t, *g.runs, "a blocked msg server ran at height %d", height)
		g.commitBlock(height)
	}
}

// TestGuardInProposalHandlers shows that a msg server called directly from
// the PrepareProposal and ProcessProposal handlers is checked against the
// breaker, at the initial height and at later ones.
func TestGuardInProposalHandlers(t *testing.T) {
	var (
		g                   *guardApp
		prepErr, processErr error
	)
	g = newGuardApp(t, true, func(app *baseapp.BaseApp) {
		app.SetPrepareProposal(func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
			prepErr = g.directCallErr(ctx)
			return &abci.ResponsePrepareProposal{Txs: req.Txs}, nil
		})
		app.SetProcessProposal(func(ctx sdk.Context, _ *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
			processErr = g.directCallErr(ctx)
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
		})
	})
	g.cb.blocked[counterURL] = true

	for height := int64(1); height <= 3; height++ {
		prepErr, processErr = nil, nil
		_, err := g.app.PrepareProposal(&abci.RequestPrepareProposal{Height: height, MaxTxBytes: 1_000})
		require.NoError(t, err)
		_, err = g.app.ProcessProposal(&abci.RequestProcessProposal{Height: height})
		require.NoError(t, err)
		require.ErrorIs(t, prepErr, sdkerrors.ErrUnauthorized, "PrepareProposal at height %d", height)
		require.ErrorIs(t, processErr, sdkerrors.ErrUnauthorized, "ProcessProposal at height %d", height)
		require.Zero(t, *g.runs, "a blocked msg server ran at height %d", height)
		g.commitBlock(height)
	}
}

// TestGuardInBlockHooks shows that a msg server called directly from the
// PreBlock, BeginBlock and EndBlock hooks is checked against the breaker,
// both when FinalizeBlock creates the block state and when a validator runs
// ProcessProposal first, which resets the block state FinalizeBlock then
// uses.
func TestGuardInBlockHooks(t *testing.T) {
	for _, processFirst := range []bool{false, true} {
		t.Run(fmt.Sprintf("ProcessProposal first %t", processFirst), func(t *testing.T) {
			var g *guardApp
			errs := map[string]error{}
			g = newGuardApp(t, true, func(app *baseapp.BaseApp) {
				app.SetPreBlocker(func(ctx sdk.Context, _ *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
					errs["PreBlock"] = g.directCallErr(ctx)
					return &sdk.ResponsePreBlock{}, nil
				})
				app.SetBeginBlocker(func(ctx sdk.Context) (sdk.BeginBlock, error) {
					errs["BeginBlock"] = g.directCallErr(ctx)
					return sdk.BeginBlock{}, nil
				})
				app.SetEndBlocker(func(ctx sdk.Context) (sdk.EndBlock, error) {
					errs["EndBlock"] = g.directCallErr(ctx)
					return sdk.EndBlock{}, nil
				})
			})
			g.cb.blocked[counterURL] = true

			for height := int64(1); height <= 3; height++ {
				clear(errs)
				if processFirst {
					_, err := g.app.ProcessProposal(&abci.RequestProcessProposal{Height: height})
					require.NoError(t, err)
				}
				g.commitBlock(height)
				require.Len(t, errs, 3)
				for hook, err := range errs {
					require.ErrorIs(t, err, sdkerrors.ErrUnauthorized, "%s at height %d", hook, height)
				}
				require.Zero(t, *g.runs, "a blocked msg server ran at height %d", height)
			}
		})
	}
}

// TestGuardHybridHandlerChecksOnce shows that a msg dispatched through a
// hybrid handler pays for one lookup and runs under its own mark, both on a
// plain Context and inside a routed msg whose mark names a different msg.
func TestGuardHybridHandlerChecksOnce(t *testing.T) {
	for _, outer := range []string{"", counter2URL} {
		// A gRPC-style caller hands the handler a plain context.Context with
		// the sdk.Context behind sdk.SdkContextKey.
		for _, wrapped := range []bool{false, true} {
			t.Run(fmt.Sprintf("outer mark %q wrapped %t", outer, wrapped), func(t *testing.T) {
				wrap := func(sdkCtx sdk.Context) context.Context {
					if !wrapped {
						return sdkCtx
					}
					probe := context.WithValue(context.Background(), guardProbeKey{}, "caller value")
					return context.WithValue(probe, sdk.SdkContextKey, sdkCtx)
				}

				g := newGuardApp(t, true)
				ctx := getFinalizeBlockStateCtx(g.app).WithCheckedMsgTypeURL(outer)
				hybrid := g.app.MsgServiceRouter().HybridHandlerByMsgName("MsgCounter")
				require.NotNil(t, hybrid)

				gasBefore := ctx.GasMeter().GasConsumed()
				err := hybrid(wrap(ctx), &baseapptestutil.MsgCounter{Counter: 1}, &baseapptestutil.MsgCreateCounterResponse{})
				require.NoError(t, err)
				require.Equal(t, map[string]int{counterURL: 1}, g.cb.calls)
				require.Equal(t, []string{counterURL}, *g.marks)
				require.Equal(t, 1, *g.runs)
				if wrapped {
					require.Equal(t, []any{"caller value"}, *g.inner.probes, "the mark dropped the caller's Context values")
				}
				// A caller that passes an sdk.Context hands the handler an
				// sdk.Context, so a handler or interceptor that type-asserts
				// ctx.(sdk.Context) still works.
				require.Equal(t, []bool{!wrapped}, *g.inner.plain, "the mark changed the handler's ctx type")

				// the same call with an unguarded server pays the same gas
				plain := newGuardApp(t, false)
				plainCtx := getFinalizeBlockStateCtx(plain.app).WithCheckedMsgTypeURL(outer)
				plainBefore := plainCtx.GasMeter().GasConsumed()
				err = plain.app.MsgServiceRouter().HybridHandlerByMsgName("MsgCounter")(wrap(plainCtx), &baseapptestutil.MsgCounter{Counter: 1}, &baseapptestutil.MsgCreateCounterResponse{})
				require.NoError(t, err)
				require.Equal(t, plainCtx.GasMeter().GasConsumed()-plainBefore, ctx.GasMeter().GasConsumed()-gasBefore)
			})
		}
	}
}

// TestBypassContextSkipsEveryCheck shows that on a Context with the circuit
// breaker bypass set, as x/upgrade passes an upgrade handler, a blocked msg
// runs through the router, through a hybrid handler, through a msg server
// that dispatches it through the router, and in a direct call, with no
// breaker lookup, while the same Context without the bypass refuses it.
func TestBypassContextSkipsEveryCheck(t *testing.T) {
	g := newGuardApp(t, true)
	g.cb.blocked[counterURL] = true
	g.cb.blocked[counter2URL] = true
	base := getFinalizeBlockStateCtx(g.app)
	counter := &baseapptestutil.MsgCounter{Counter: 1, Signer: guardSigner()}
	relay := &baseapptestutil.MsgCounter2{Counter: 1, Signer: guardSigner()}
	router := g.app.MsgServiceRouter()

	calls := map[string]func(ctx sdk.Context) error{
		"router": func(ctx sdk.Context) error {
			_, err := router.Handler(counter)(ctx, counter)
			return err
		},
		"hybrid": func(ctx sdk.Context) error {
			return router.HybridHandlerByMsgName("MsgCounter")(ctx, counter, &baseapptestutil.MsgCreateCounterResponse{})
		},
		"wrapper": func(ctx sdk.Context) error {
			_, err := router.Handler(relay)(ctx, relay)
			return err
		},
		"direct": g.directCallErr,
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			g.reset()
			require.ErrorIs(t, call(base), sdkerrors.ErrUnauthorized)
			require.NotEmpty(t, g.cb.calls)

			g.reset()
			require.NoError(t, call(base.WithCircuitBreakerBypass(true)))
			require.Empty(t, g.cb.calls)
			require.Equal(t, 1, *g.runs)
		})
	}
}
