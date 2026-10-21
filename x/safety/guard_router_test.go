package safety_test

import (
	"context"
	"math/rand"
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	errorsmod "cosmossdk.io/errors"

	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/x/authz"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/safety/types"
)

const (
	urlExec        = "/cosmos.authz.v1beta1.MsgExec"
	urlBankParams  = "/cosmos.bank.v1beta1.MsgUpdateParams"
	urlRelayCaller = "/MsgCounter2"
)

// errBreakerStore stands for a circuit breaker that cannot read its state.
var errBreakerStore = errorsmod.Wrap(sdkerrors.ErrIO, "circuit breaker store unavailable")

// instrumentedBreaker wraps the x/safety keeper as the BaseApp circuit
// breaker. It counts lookups per type URL, from the router and from the msg
// server guard alike, and returns errBreakerStore for the type URLs in
// failing.
type instrumentedBreaker struct {
	inner   sdk.CircuitBreaker
	calls   map[string]int
	failing map[string]bool
}

func (b *instrumentedBreaker) IsAllowed(ctx context.Context, typeURL string) (bool, error) {
	b.calls[typeURL]++
	if b.failing[typeURL] {
		return false, errBreakerStore
	}
	return b.inner.IsAllowed(ctx, typeURL)
}

func (b *instrumentedBreaker) resetCalls() { b.calls = map[string]int{} }

// instrumentBreaker installs an instrumentedBreaker around the x/safety keeper.
// BaseApp attaches the breaker to a state's Context when it creates the state,
// so an empty block runs first to give Simulate a CheckTx state that has it.
func (f *routerFixture) instrumentBreaker() *instrumentedBreaker {
	f.t.Helper()
	b := &instrumentedBreaker{inner: f.safetyKeeper, failing: map[string]bool{}}
	f.app.MsgServiceRouter().SetCircuit(b)
	f.deliver()
	require.Equal(f.t, sdk.CircuitBreaker(b), f.app.GetContextForCheckTx(nil).CircuitBreaker())
	b.resetCalls()
	return b
}

// relayServer handles the test-only MsgCounter2 the way a module msg server,
// or an EVM precompile running inside one, calls another module: it runs its
// own guard, then calls the bank msg server directly with a MsgSend of one
// stake from user to grantee, without the MsgServiceRouter.
type relayServer struct {
	bank banktypes.MsgServer
	send *banktypes.MsgSend
}

func (s relayServer) IncrementCounter(ctx context.Context, msg *baseapptestutil.MsgCounter2) (*baseapptestutil.MsgCreateCounterResponse, error) {
	if err := msgservice.CheckMsgAllowed(ctx, msg); err != nil {
		return nil, err
	}
	if _, err := s.bank.Send(ctx, s.send); err != nil {
		return nil, err
	}
	return &baseapptestutil.MsgCreateCounterResponse{}, nil
}

// registerRelay adds the relay msg server to the app's MsgServiceRouter and
// returns a relay msg signed by user.
func (f *routerFixture) registerRelay() sdk.Msg {
	f.t.Helper()
	baseapptestutil.RegisterInterfaces(f.registry)
	baseapptestutil.RegisterCounter2Server(f.app.MsgServiceRouter(), relayServer{
		bank: bankkeeper.NewMsgServerImpl(f.bankKeeper),
		send: f.msgSend().(*banktypes.MsgSend),
	})
	relay := &baseapptestutil.MsgCounter2{Signer: f.user.String()}
	require.Equal(f.t, urlRelayCaller, sdk.MsgTypeURL(relay))
	return relay
}

// grantDelegate has user grant grantee a generic authorization for
// MsgDelegate.
func (f *routerFixture) grantDelegate() {
	f.t.Helper()
	grant, err := authz.NewMsgGrant(f.user.addr, f.grantee.addr, authz.NewGenericAuthorization(urlDelegate), nil)
	require.NoError(f.t, err)
	res := f.deliverOne(f.user, grant)
	require.Zero(f.t, res.Code, res.Log)
}

// signSame signs msgs at a's committed sequence like signNext, but with the
// same memo every time, so two txs of the same msgs pay the same size gas.
func (f *routerFixture) signSame(a testAccount, msgs ...sdk.Msg) []byte {
	f.t.Helper()
	tx, err := sims.GenSignedMockTx(rand.New(rand.NewSource(1)), f.txConfig, msgs, nil, sims.DefaultGenTxGas, "",
		[]uint64{f.accNum(a)}, []uint64{f.seq(a)}, a.priv)
	require.NoError(f.t, err)
	bz, err := f.txConfig.TxEncoder()(tx)
	require.NoError(f.t, err)
	return bz
}

// simulate runs msgs signed by a at its committed sequence in Simulate, with
// or without the guard's breaker. It returns the gas used and the error.
func (f *routerFixture) simulate(a testAccount, guardOff bool, msgs ...sdk.Msg) (uint64, error) {
	f.t.Helper()
	f.guardBreakerOff = guardOff
	defer func() { f.guardBreakerOff = false }()
	gasInfo, _, err := f.app.Simulate(f.signSame(a, msgs...))
	return gasInfo.GasUsed, err
}

// finalize runs msgs signed by a in their own block, with or without the
// guard's breaker, and returns the gas used, code, codespace and log.
func (f *routerFixture) finalize(a testAccount, guardOff bool, msgs ...sdk.Msg) (uint64, uint32, string, string) {
	f.t.Helper()
	f.guardBreakerOff = guardOff
	defer func() { f.guardBreakerOff = false }()
	res := f.deliver(f.signSame(a, msgs...))[0]
	return uint64(res.GasUsed), res.Code, res.Codespace, res.Log
}

// requireSameFailure asserts that Simulate and FinalizeBlock refused a tx
// with the same registered error, which wraps want.
func requireSameFailure(t *testing.T, want *errorsmod.Error, simErr error, code uint32, codespace, log string) {
	t.Helper()
	require.ErrorIs(t, simErr, want)
	simCodespace, simCode, _ := errorsmod.ABCIInfo(simErr, false)
	require.Equal(t, want.Codespace(), codespace, log)
	require.Equal(t, want.ABCICode(), code, log)
	require.Equal(t, codespace, simCodespace)
	require.Equal(t, code, simCode)
}

// newGasFixture returns a fixture with the SafetyDecorator, user's grants of
// MsgSend and MsgDelegate to grantee, and an unrelated trip, so lookups read
// a disabled list that is not empty.
func newGasFixture(t *testing.T) *routerFixture {
	t.Helper()
	f := newRouterFixture(t, true)
	f.grantSend()
	f.grantDelegate()
	f.trip(urlMultiSend)
	return f
}

// TestGuardRoutedMsgGasMatchesBreakerFreeGuard shows that the msg server
// guard adds no gas to msgs the router dispatches, top level or nested in
// authz MsgExec, in Simulate and in FinalizeBlock. The baseline runs the same
// tx with the breaker removed from the tx Context, so the router checks each
// msg as before and every guard is a no-op.
//
// Simulate compares both runs on the same state. FinalizeBlock commits, so
// the baseline runs in a second fixture built by the same steps, and the
// Simulate gas of the two fixtures shows that their states match.
func TestGuardRoutedMsgGasMatchesBreakerFreeGuard(t *testing.T) {
	guardedApp := newGasFixture(t)
	baselineApp := newGasFixture(t)

	testCases := []struct {
		name string
		tx   func(f *routerFixture) (testAccount, []sdk.Msg)
	}{
		{
			name: "bank send",
			tx:   func(f *routerFixture) (testAccount, []sdk.Msg) { return f.user, []sdk.Msg{f.msgSend()} },
		},
		{
			name: "staking delegate",
			tx:   func(f *routerFixture) (testAccount, []sdk.Msg) { return f.user, []sdk.Msg{f.msgDelegate()} },
		},
		{
			name: "two top-level msgs",
			tx: func(f *routerFixture) (testAccount, []sdk.Msg) {
				return f.user, []sdk.Msg{f.msgSend(), f.msgDelegate()}
			},
		},
		{
			name: "authz exec of send and delegate",
			tx: func(f *routerFixture) (testAccount, []sdk.Msg) {
				exec := authz.NewMsgExec(f.grantee.addr, []sdk.Msg{f.msgSend(), f.msgDelegate()})
				return f.grantee, []sdk.Msg{&exec}
			},
		},
		{
			name: "authz exec nested in authz exec",
			tx: func(f *routerFixture) (testAccount, []sdk.Msg) {
				inner := authz.NewMsgExec(f.grantee.addr, []sdk.Msg{f.msgSend(), f.msgDelegate()})
				outer := authz.NewMsgExec(f.grantee.addr, []sdk.Msg{&inner})
				return f.grantee, []sdk.Msg{&outer}
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			signer, msgs := tc.tx(guardedApp)
			guarded, err := guardedApp.simulate(signer, false, msgs...)
			require.NoError(t, err)
			baseline, err := guardedApp.simulate(signer, true, msgs...)
			require.NoError(t, err)
			require.Equal(t, baseline, guarded, "Simulate")

			baseSigner, baseMsgs := tc.tx(baselineApp)
			baseline, err = baselineApp.simulate(baseSigner, true, baseMsgs...)
			require.NoError(t, err)
			require.Equal(t, baseline, guarded, "the two fixtures must start from matching states")

			guardedFinal, code, _, log := guardedApp.finalize(signer, false, msgs...)
			require.Zero(t, code, log)
			baselineFinal, code, _, log := baselineApp.finalize(baseSigner, true, baseMsgs...)
			require.Zero(t, code, log)
			require.Equal(t, baselineFinal, guardedFinal, "FinalizeBlock")
		})
	}
}

// TestGuardGovProposalMsgsGasMatchesBreakerFreeGuard shows that the guard adds
// no gas to the msgs gov executes from a passed proposal, which gov
// dispatches through the router in its EndBlocker.
func TestGuardGovProposalMsgsGasMatchesBreakerFreeGuard(t *testing.T) {
	f := newRouterFixture(t, false)
	b := f.instrumentBreaker()

	bankParams := banktypes.DefaultParams()
	bankParams.DefaultSendEnabled = false
	id := f.submitProposal(f.msgDistrParams("0.2"), &banktypes.MsgUpdateParams{Authority: f.gov, Params: bankParams})
	end := f.proposal(id).VotingEndTime.Add(1)

	endBlock := func(guardOff bool) uint64 {
		ctx := f.app.NewNextBlockContext(cmtproto.Header{Height: f.app.LastBlockHeight() + 1, Time: end})
		require.Equal(t, sdk.CircuitBreaker(b), ctx.CircuitBreaker())
		ctx, _ = ctx.CacheContext()
		ctx = ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())
		if guardOff {
			ctx = ctx.WithCircuitBreaker(nil)
		}
		b.resetCalls()
		require.NoError(t, gov.EndBlocker(ctx, f.govKeeper))

		p, err := f.govKeeper.Proposals.Get(ctx, id)
		require.NoError(t, err)
		require.Equal(t, govv1.StatusPassed, p.Status, p.FailedReason)
		require.Equal(t, map[string]int{urlDistrParams: 1, urlBankParams: 1}, b.calls,
			"the router checks each proposal msg once and the guard repeats none")
		return ctx.GasMeter().GasConsumed()
	}
	require.Equal(t, endBlock(true), endBlock(false))
}

// TestAuthzExecInnerMsgsCheckedOnce counts circuit breaker lookups and shows
// that each msg in an authz MsgExec, at any depth, is looked up exactly once,
// by the router, the same way in Simulate and in FinalizeBlock.
func TestAuthzExecInnerMsgsCheckedOnce(t *testing.T) {
	f := newRouterFixture(t, false)
	f.grantSend()
	f.grantDelegate()
	b := f.instrumentBreaker()

	execSendDelegate := authz.NewMsgExec(f.grantee.addr, []sdk.Msg{f.msgSend(), f.msgDelegate()})
	inner := authz.NewMsgExec(f.grantee.addr, []sdk.Msg{f.msgSend(), f.msgSend()})
	execNested := authz.NewMsgExec(f.grantee.addr, []sdk.Msg{&inner, f.msgDelegate()})

	testCases := []struct {
		name string
		msg  sdk.Msg
		exp  map[string]int
	}{
		{name: "send and delegate", msg: &execSendDelegate, exp: map[string]int{urlExec: 1, urlSend: 1, urlDelegate: 1}},
		{name: "nested exec", msg: &execNested, exp: map[string]int{urlExec: 2, urlSend: 2, urlDelegate: 1}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b.resetCalls()
			_, err := f.simulate(f.grantee, false, tc.msg)
			require.NoError(t, err)
			require.Equal(t, tc.exp, b.calls, "Simulate")

			b.resetCalls()
			_, code, _, log := f.finalize(f.grantee, false, tc.msg)
			require.Zero(t, code, log)
			require.Equal(t, tc.exp, b.calls, "FinalizeBlock")
		})
	}
}

// TestGuardChecksMsgServerCalledFromMsgServer shows that the mark the router
// sets for the msg it dispatches does not cover a different msg that its msg
// server sends to another msg server directly: the bank guard checks the
// MsgSend, and a trip of MsgSend stops it, in Simulate and in FinalizeBlock
// alike, while the router let the outer msg through.
func TestGuardChecksMsgServerCalledFromMsgServer(t *testing.T) {
	f := newRouterFixture(t, true)
	b := f.instrumentBreaker()
	relay := f.registerRelay()

	before := f.balance(f.ctx(), f.grantee)
	_, err := f.simulate(f.user, false, relay)
	require.NoError(t, err)
	require.Equal(t, map[string]int{urlRelayCaller: 1, urlSend: 1}, b.calls,
		"the router checks the relay msg and the bank guard checks the direct MsgSend")
	_, code, _, log := f.finalize(f.user, false, relay)
	require.Zero(t, code, log)
	require.Equal(t, before.AddRaw(1), f.balance(f.ctx(), f.grantee))

	// the direct call pays its own lookup, unlike a routed msg
	guarded, err := f.simulate(f.user, false, relay)
	require.NoError(t, err)
	baseline, err := f.simulate(f.user, true, relay)
	require.NoError(t, err)
	require.Greater(t, guarded, baseline)

	f.trip(urlSend)
	before = f.balance(f.ctx(), f.grantee)
	b.resetCalls()
	_, simErr := f.simulate(f.user, false, relay)
	require.ErrorContains(t, simErr, urlSend)
	require.Equal(t, map[string]int{urlRelayCaller: 1, urlSend: 1}, b.calls)
	_, code, codespace, log := f.finalize(f.user, false, relay)
	requireSameFailure(t, types.ErrMsgDisabled, simErr, code, codespace, log)
	require.Contains(t, log, urlSend)
	require.Equal(t, before, f.balance(f.ctx(), f.grantee), "the tripped direct MsgSend moved funds")

	// without the guard's breaker the direct MsgSend runs despite the trip,
	// which is the gap the guard closes
	_, err = f.simulate(f.user, true, relay)
	require.NoError(t, err)

	f.reset(urlSend)
	f.trip(types.Wildcard)
	_, simErr = f.simulate(f.user, false, relay)
	_, code, codespace, log = f.finalize(f.user, false, relay)
	requireSameFailure(t, types.ErrMsgDisabled, simErr, code, codespace, log)
	require.Contains(t, log, urlRelayCaller, "under * the router refuses the relay msg itself")
}

// TestGuardFailsClosedOnBreakerError shows that when the circuit breaker
// returns an error for a msg, the guard refuses the msg with that error, in a
// direct call, in Simulate and in FinalizeBlock, and no funds move.
func TestGuardFailsClosedOnBreakerError(t *testing.T) {
	f := newRouterFixture(t, true)
	b := f.instrumentBreaker()
	relay := f.registerRelay()
	b.failing[urlSend] = true

	// a direct call on the block Context, as an EVM precompile makes
	ctx := f.blockCtx()
	before := f.balance(ctx, f.grantee)
	err := f.directCalls()[urlSend](ctx)
	require.ErrorIs(t, err, sdkerrors.ErrIO)
	require.ErrorContains(t, err, errBreakerStore.Error())
	require.Equal(t, before, f.balance(ctx, f.grantee))

	// a direct call from inside a routed msg
	before = f.balance(f.ctx(), f.grantee)
	_, simErr := f.simulate(f.user, false, relay)
	_, code, codespace, log := f.finalize(f.user, false, relay)
	requireSameFailure(t, sdkerrors.ErrIO, simErr, code, codespace, log)
	require.Equal(t, before, f.balance(f.ctx(), f.grantee))

	// the router fails closed on the same error for a routed MsgSend
	_, simErr = f.simulate(f.user, false, f.msgSend())
	_, code, codespace, log = f.finalize(f.user, false, f.msgSend())
	requireSameFailure(t, sdkerrors.ErrIO, simErr, code, codespace, log)
	require.Equal(t, before, f.balance(f.ctx(), f.grantee))

	// a breaker error for another msg leaves MsgSend alone
	b.failing = map[string]bool{urlDelegate: true}
	_, code, _, log = f.finalize(f.user, false, relay)
	require.Zero(t, code, log)
	require.Equal(t, before.AddRaw(1), f.balance(f.ctx(), f.grantee))
}
