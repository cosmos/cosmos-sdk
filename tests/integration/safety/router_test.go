package safety_test

import (
	"fmt"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/simapp"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	safetytypes "github.com/cosmos/cosmos-sdk/x/safety/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// newGateFixture starts a SimApp chain and makes the voter the operator with
// scope "*", so the voter proposes, votes, trips and grants, and other is its
// authz grantee. With withAnte false, the ante chain runs no x/safety check
// and the router is the only gate.
func newGateFixture(t *testing.T, withAnte bool) *fixture {
	t.Helper()
	var setup func(*simapp.SimApp)
	if !withAnte {
		setup = func(app *simapp.SimApp) { app.SetAnteHandler(anteWithBreaker(t, app, nil)) }
	}
	f := newFixtureWithSetup(t, nil, nil, setup)
	f.setParams(f.voter.String(), safetytypes.Wildcard)
	return f
}

// routerBlockedMsg is the text of the error the router returns for a disabled msg.
var routerBlockedMsg = safetytypes.ErrMsgDisabled.Error()

var oneStake = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1))

// requireRouterBlocked asserts that a tx failed at execution because the
// router circuit breaker refused typeURL, with the registered ErrMsgDisabled.
func requireRouterBlocked(t *testing.T, res *abci.ExecTxResult, typeURL string) {
	t.Helper()
	require.Equal(t, safetytypes.ErrMsgDisabled.Codespace(), res.Codespace, res.Log)
	require.Equal(t, safetytypes.ErrMsgDisabled.ABCICode(), res.Code, res.Log)
	require.Contains(t, res.Log, typeURL)
}

func (f *fixture) msgSend() sdk.Msg {
	return banktypes.NewMsgSend(f.voter.addr, f.other.addr, oneStake)
}

func (f *fixture) msgMultiSend() sdk.Msg {
	return &banktypes.MsgMultiSend{
		Inputs:  []banktypes.Input{banktypes.NewInput(f.voter.addr, oneStake)},
		Outputs: []banktypes.Output{banktypes.NewOutput(f.other.addr, oneStake)},
	}
}

func (f *fixture) msgDelegate() sdk.Msg {
	return stakingtypes.NewMsgDelegate(f.voter.String(), f.validator().String(), sdk.NewInt64Coin(sdk.DefaultBondDenom, 1))
}

func (f *fixture) msgDistrParams(communityTax string) sdk.Msg {
	params, err := f.app.DistrKeeper.Params.Get(f.ctx())
	require.NoError(f.t, err)
	params.CommunityTax = sdkmath.LegacyMustNewDecFromStr(communityTax)
	return &distrtypes.MsgUpdateParams{Authority: govAddr, Params: params}
}

func (f *fixture) validator() sdk.ValAddress {
	return sdk.ValAddress(f.valSet.Validators[0].Address)
}

func (f *fixture) balanceAt(ctx sdk.Context, a account) sdkmath.Int {
	return f.app.BankKeeper.GetBalance(ctx, a.addr, sdk.DefaultBondDenom).Amount
}

func (f *fixture) communityTax() sdkmath.LegacyDec {
	params, err := f.app.DistrKeeper.Params.Get(f.ctx())
	require.NoError(f.t, err)
	return params.CommunityTax
}

// callHandler runs a router handler on a branch of ctx. A panic from the
// module handler counts as passing the breaker, since the breaker check runs
// before the module handler.
func callHandler(ctx sdk.Context, handler baseapp.MsgServiceHandler, msg sdk.Msg) (err error) {
	cacheCtx, _ := ctx.CacheContext()
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("handler panicked: %v", r)
		}
	}()
	_, err = handler(cacheCtx, msg)
	return err
}

// grantSend has the voter grant other a generic authorization for MsgSend.
func (f *fixture) grantSend() {
	f.t.Helper()
	grant, err := authz.NewMsgGrant(f.voter.addr, f.other.addr, authz.NewGenericAuthorization(msgSendURL), nil)
	require.NoError(f.t, err)
	res := f.deliver(f.voter, grant)
	require.Zero(f.t, res.Code, res.Log)
}

// TestRouterTrippedMsgFailsInFinalizeBlock shows the router is a gate on its
// own: with no circuit breaker in the ante chain, a tripped msg passes the
// ante handler and fails when the router dispatches it, and a reset restores
// it.
func TestRouterTrippedMsgFailsInFinalizeBlock(t *testing.T) {
	f := newGateFixture(t, false)

	res := f.deliver(f.voter, f.msgSend())
	require.Zero(t, res.Code, res.Log)

	f.trip(msgSendURL)
	require.Equal(t, []string{msgSendURL}, f.disabled())

	seqBefore := f.seq(f.voter)
	balBefore := f.balanceAt(f.ctx(), f.other)
	res = f.deliver(f.voter, f.msgSend())
	requireRouterBlocked(t, res, msgSendURL)
	require.Equal(t, balBefore, f.balanceAt(f.ctx(), f.other), "blocked send moved funds")
	// without the decorator the ante handler ran, so the sequence was used
	require.Equal(t, seqBefore+1, f.seq(f.voter))

	// an exact trip leaves sibling msgs alone
	res = f.deliver(f.voter, f.msgMultiSend())
	require.Zero(t, res.Code, res.Log)

	// a tx is atomic: an allowed msg next to a tripped one is rolled back
	balBefore = f.balanceAt(f.ctx(), f.other)
	res = f.deliver(f.voter, f.msgMultiSend(), f.msgSend())
	requireRouterBlocked(t, res, msgSendURL)
	require.Equal(t, balBefore, f.balanceAt(f.ctx(), f.other))

	f.reset(msgSendURL)
	require.Empty(t, f.disabled())
	res = f.deliver(f.voter, f.msgSend())
	require.Zero(t, res.Code, res.Log)
}

// TestRouterAuthzExecOfTrippedMsgFails shows a tripped msg nested in authz
// MsgExec fails at execution, because authz dispatches through the same
// router handler. MsgExec itself is not tripped.
func TestRouterAuthzExecOfTrippedMsgFails(t *testing.T) {
	f := newGateFixture(t, false)
	f.grantSend()

	exec := authz.NewMsgExec(f.other.addr, []sdk.Msg{f.msgSend()})
	res := f.deliver(f.other, &exec)
	require.Zero(t, res.Code, "exec before trip: %s", res.Log)

	f.trip(msgSendURL)

	balBefore := f.balanceAt(f.ctx(), f.voter)
	res = f.deliver(f.other, &exec)
	requireRouterBlocked(t, res, msgSendURL)
	require.Equal(t, balBefore, f.balanceAt(f.ctx(), f.voter), "granter funds moved")

	// a doubly nested exec is refused too
	inner := authz.NewMsgExec(f.other.addr, []sdk.Msg{f.msgSend()})
	outer := authz.NewMsgExec(f.other.addr, []sdk.Msg{&inner})
	res = f.deliver(f.other, &outer)
	require.NotZero(t, res.Code, res.Log)

	f.reset(msgSendURL)
	res = f.deliver(f.other, &exec)
	require.Zero(t, res.Code, "exec after reset: %s", res.Log)
}

// TestRouterGovProposalOfTrippedMsgFails shows a proposal that executes a
// tripped msg fails at execution, all of its msgs roll back, and a proposal
// of protected msgs still executes while the trip stands.
func TestRouterGovProposalOfTrippedMsgFails(t *testing.T) {
	f := newGateFixture(t, false)
	taxBefore := f.communityTax()
	govParams, err := f.app.GovKeeper.Params.Get(f.ctx())
	require.NoError(t, err)

	// both proposals are submitted while nothing is tripped, then the operator
	// trips during the voting period
	only := f.submitProposal(f.msgDistrParams("0.33"))
	newGovParams := govParams
	newGovParams.BurnVoteVeto = !govParams.BurnVoteVeto
	mixed := f.submitProposal(&govv1.MsgUpdateParams{Authority: govAddr, Params: newGovParams}, f.msgDistrParams("0.34"))
	f.trip(distrParamsURL)
	f.endVoting()

	for _, id := range []uint64{only, mixed} {
		p := f.requireProposalStatus(id, govv1.StatusFailed)
		require.Contains(t, p.FailedReason, routerBlockedMsg, "proposal %d", id)
		require.Contains(t, p.FailedReason, distrParamsURL, "proposal %d", id)
	}
	require.True(t, taxBefore.Equal(f.communityTax()), "tripped proposal changed distribution params")
	got, err := f.app.GovKeeper.Params.Get(f.ctx())
	require.NoError(t, err)
	require.Equal(t, govParams.BurnVoteVeto, got.BurnVoteVeto, "the gov msg in a failed proposal must roll back")

	// protected msgs execute through gov while the trip stands: gov params,
	// and the authority resetting the trip
	protected := f.submitProposal(
		&govv1.MsgUpdateParams{Authority: govAddr, Params: newGovParams},
		safetytypes.NewMsgReset(govAddr, distrParamsURL),
	)
	require.Equal(t, []string{distrParamsURL}, f.disabled(), "the trip must stand until the proposal executes")
	f.endVoting()
	f.requireProposalStatus(protected, govv1.StatusPassed)
	got, err = f.app.GovKeeper.Params.Get(f.ctx())
	require.NoError(t, err)
	require.Equal(t, newGovParams.BurnVoteVeto, got.BurnVoteVeto)
	require.Empty(t, f.disabled())

	// control: with the trip gone the same distribution proposal passes
	again := f.submitProposal(f.msgDistrParams("0.33"))
	f.endVoting()
	f.requireProposalStatus(again, govv1.StatusPassed)
	require.True(t, sdkmath.LegacyMustNewDecFromStr("0.33").Equal(f.communityTax()))
}

// wrapperTypeURLs is the wrapper set, spelled out so that a msg wrongly
// added to it fails the sweep below.
var wrapperTypeURLs = map[string]bool{
	"/cosmos.authz.v1beta1.MsgExec":      true,
	"/cosmos.group.v1.MsgSubmitProposal": true,
	"/cosmos.group.v1.MsgVote":           true,
	"/cosmos.group.v1.MsgExec":           true,
}

// TestRouterWildcardBlocksAllButProtected trips "*" and shows that every
// msg outside the protected and wrapper sets fails in FinalizeBlock while gov
// keeps running and a gov proposal of gov, safety, consensus and upgrade msgs
// executes.
func TestRouterWildcardBlocksAllButProtected(t *testing.T) {
	f := newGateFixture(t, false)
	f.grantSend()
	f.trip(safetytypes.Wildcard)

	// sweep every registered msg type through its real router handler: the
	// breaker refuses exactly those outside the protected and wrapper sets
	ctx := f.ctx()
	registry := f.app.InterfaceRegistry()
	router := f.app.MsgServiceRouter()
	var blocked, protected int
	for _, typeURL := range registry.ListImplementations(sdk.MsgInterfaceProtoName) {
		passes := safetytypes.IsProtected(typeURL) || wrapperTypeURLs[typeURL]
		allowed, err := f.app.SafetyKeeper.IsAllowed(ctx, typeURL)
		if passes {
			require.NoError(t, err, typeURL)
		} else {
			require.ErrorIs(t, err, safetytypes.ErrMsgDisabled, typeURL)
		}
		require.Equal(t, passes, allowed, typeURL)

		handler := router.HandlerByTypeURL(typeURL)
		if handler == nil {
			continue
		}
		resolved, err := registry.Resolve(typeURL)
		require.NoError(t, err, typeURL)
		msg := resolved
		if m, ok := msg.(sdk.HasValidateBasic); ok && m.ValidateBasic() != nil {
			// the router runs ValidateBasic before the breaker, so an empty
			// msg of this type never reaches it
			continue
		}
		routerErr := callHandler(ctx, handler, msg)
		if passes {
			protected++
			if routerErr != nil {
				require.NotContains(t, routerErr.Error(), routerBlockedMsg, typeURL)
			}
			continue
		}
		blocked++
		require.Error(t, routerErr, typeURL)
		require.Contains(t, routerErr.Error(), routerBlockedMsg, typeURL)
	}
	t.Logf("router sweep under *: %d msg types blocked, %d protected or wrapper msg types passed the breaker", blocked, protected)
	require.Positive(t, blocked)
	require.Positive(t, protected)

	// real txs: bank, staking, authz, and a nested exec all fail
	for name, msg := range map[string]sdk.Msg{
		msgSendURL:      f.msgSend(),
		msgMultiSendURL: f.msgMultiSend(),
		msgDelegateURL:  f.msgDelegate(),
		authzGrantURL: func() sdk.Msg {
			m, err := authz.NewMsgGrant(f.voter.addr, f.other.addr, authz.NewGenericAuthorization(msgDelegateURL), nil)
			require.NoError(t, err)
			return m
		}(),
	} {
		res := f.deliver(f.voter, msg)
		requireRouterBlocked(t, res, name)
	}
	// MsgExec passes the breaker; the router blocks the msg it dispatches
	exec := authz.NewMsgExec(f.other.addr, []sdk.Msg{f.msgSend()})
	res := f.deliver(f.other, &exec)
	requireRouterBlocked(t, res, msgSendURL)

	// gov still takes proposals and votes, and executes protected msgs
	govParams, err := f.app.GovKeeper.Params.Get(f.ctx())
	require.NoError(t, err)
	newGovParams := govParams
	newGovParams.BurnVoteVeto = !govParams.BurnVoteVeto

	newSafetyParams := safetytypes.Params{Authority: govAddr, Operator: f.other.String(), Scope: []string{safetytypes.Wildcard}}

	cp := sims.DefaultConsensusParams
	newBlock := &cmtproto.BlockParams{MaxBytes: cp.Block.MaxBytes, MaxGas: cp.Block.MaxGas - 1}
	consensusMsg := &consensustypes.MsgUpdateParams{
		Authority: govAddr,
		Block:     newBlock,
		Evidence:  cp.Evidence,
		Validator: cp.Validator,
		Auth:      cp.Authority,
	}

	plan := upgradetypes.Plan{Name: "safety-test", Height: f.app.LastBlockHeight() + 1_000}
	upgradeMsg := &upgradetypes.MsgSoftwareUpgrade{Authority: govAddr, Plan: plan}

	id := f.submitProposal(
		&govv1.MsgUpdateParams{Authority: govAddr, Params: newGovParams},
		&safetytypes.MsgUpdateParams{Authority: govAddr, Params: newSafetyParams},
		consensusMsg,
		upgradeMsg,
	)
	failing := f.submitProposal(f.msgDistrParams("0.33"))
	f.endVoting()

	f.requireProposalStatus(id, govv1.StatusPassed)
	p := f.requireProposalStatus(failing, govv1.StatusFailed)
	require.Contains(t, p.FailedReason, routerBlockedMsg)
	require.Contains(t, p.FailedReason, distrParamsURL)

	ctx = f.ctx()
	gotGov, err := f.app.GovKeeper.Params.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, newGovParams.BurnVoteVeto, gotGov.BurnVoteVeto)
	gotSafety, err := f.app.SafetyKeeper.Params.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, newSafetyParams, gotSafety)
	gotConsensus, err := f.app.ConsensusParamsKeeper.ParamsStore.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, newBlock.MaxGas, gotConsensus.Block.MaxGas)
	gotPlan, err := f.app.UpgradeKeeper.GetUpgradePlan(ctx)
	require.NoError(t, err)
	require.Equal(t, plan.Name, gotPlan.Name)
	require.Equal(t, plan.Height, gotPlan.Height)
	require.Equal(t, []string{safetytypes.Wildcard}, f.disabled(), "MsgUpdateParams must not change the disabled list")

	// the old operator lost its role; the new operator resets "*"
	res = f.deliver(f.voter, safetytypes.NewMsgReset(f.voter.String(), safetytypes.Wildcard))
	require.Equal(t, sdkerrors.ErrUnauthorized.ABCICode(), res.Code, res.Log)
	require.Equal(t, sdkerrors.ErrUnauthorized.Codespace(), res.Codespace)

	res = f.deliver(f.other, safetytypes.NewMsgReset(f.other.String(), safetytypes.Wildcard))
	require.Zero(t, res.Code, res.Log)
	require.Empty(t, f.disabled())

	res = f.deliver(f.voter, f.msgSend())
	require.Zero(t, res.Code, res.Log)
}

// TestRouterPrefixTrips trips prefix patterns and checks they block exactly
// the msgs under the prefix, at segment boundaries, and never protected msgs.
func TestRouterPrefixTrips(t *testing.T) {
	f := newGateFixture(t, false)

	f.trip("/cosmos.bank.*")
	res := f.deliver(f.voter, f.msgSend())
	requireRouterBlocked(t, res, msgSendURL)
	res = f.deliver(f.voter, f.msgMultiSend())
	requireRouterBlocked(t, res, msgMultiSendURL)
	res = f.deliver(f.voter, f.msgDelegate())
	require.Zero(t, res.Code, "staking is outside /cosmos.bank.*: %s", res.Log)

	// a prefix trip does not absorb an exact reset: the exact pattern is not
	// in the list
	res = f.deliver(f.voter, safetytypes.NewMsgReset(f.voter.String(), msgSendURL))
	require.Equal(t, safetytypes.ErrNotDisabled.ABCICode(), res.Code, res.Log)
	f.reset("/cosmos.bank.*")

	f.trip("/cosmos.staking.v1beta1.*")
	res = f.deliver(f.voter, f.msgDelegate())
	requireRouterBlocked(t, res, msgDelegateURL)
	res = f.deliver(f.voter, f.msgSend())
	require.Zero(t, res.Code, res.Log)
	f.reset("/cosmos.staking.v1beta1.*")

	// the widest prefix still leaves protected msgs alone: x/safety itself
	// (the reset below) and gov (a proposal and a vote)
	f.trip("/cosmos.*")
	for url, msg := range map[string]sdk.Msg{msgSendURL: f.msgSend(), msgDelegateURL: f.msgDelegate()} {
		res := f.deliver(f.voter, msg)
		requireRouterBlocked(t, res, url)
	}
	id := f.submitProposal(f.msgDistrParams("0.33"))
	f.endVoting()
	p := f.requireProposalStatus(id, govv1.StatusFailed)
	require.Contains(t, p.FailedReason, distrParamsURL)

	// tripping a disabled pattern again is rejected
	res = f.deliver(f.voter, safetytypes.NewMsgTrip(f.voter.String(), "/cosmos.*"))
	require.Equal(t, safetytypes.ErrAlreadyDisabled.ABCICode(), res.Code, res.Log)
	f.reset("/cosmos.*")

	// a prefix tripped over an existing exact entry is accepted, and the exact
	// entry keeps MsgSend blocked after the prefix is reset
	f.trip(msgSendURL)
	f.trip("/cosmos.*")
	f.reset("/cosmos.*")
	res = f.deliver(f.voter, f.msgSend())
	requireRouterBlocked(t, res, msgSendURL)
	res = f.deliver(f.voter, f.msgDelegate())
	require.Zero(t, res.Code, res.Log)
	f.reset(msgSendURL)
	res = f.deliver(f.voter, f.msgSend())
	require.Zero(t, res.Code, res.Log)
}

// TestRouterTripTakesEffectInSameBlock puts a trip and the msg it blocks in
// one block: txs before the trip run, txs after it fail, and a reset later
// in the block lifts it for the txs after the reset.
func TestRouterTripTakesEffectInSameBlock(t *testing.T) {
	f := newGateFixture(t, false)
	seq := f.seq(f.voter)
	before := f.balanceAt(f.ctx(), f.other)

	res := f.deliverTxs(
		f.encodeAt(f.voter, seq, f.msgSend()),
		f.encodeAt(f.voter, seq+1, safetytypes.NewMsgTrip(f.voter.String(), msgSendURL)),
		f.encodeAt(f.voter, seq+2, f.msgSend()),
		f.encodeAt(f.voter, seq+3, f.msgMultiSend()),
		f.encodeAt(f.voter, seq+4, safetytypes.NewMsgReset(f.voter.String(), msgSendURL)),
		f.encodeAt(f.voter, seq+5, f.msgSend()),
	)
	require.Zero(t, res[0].Code, res[0].Log)
	require.Zero(t, res[1].Code, res[1].Log)
	requireRouterBlocked(t, res[2], msgSendURL)
	require.Zero(t, res[3].Code, res[3].Log)
	require.Zero(t, res[4].Code, res[4].Log)
	require.Zero(t, res[5].Code, res[5].Log)

	// three sends of one stake succeeded: the first, the multisend, the last
	require.Equal(t, before.AddRaw(3), f.balanceAt(f.ctx(), f.other))
	require.Empty(t, f.disabled())
}

// TestAnteCheckTxRejectsTrippedMsgs wires the CircuitBreakerDecorator and shows
// CheckTx turns away a tripped top-level msg, an authz-nested one, and a
// proposal with one, before the fee is deducted or the sequence used.
func TestAnteCheckTxRejectsTrippedMsgs(t *testing.T) {
	f := newGateFixture(t, true)
	f.grantSend()
	f.trip(msgSendURL)

	exec := authz.NewMsgExec(f.other.addr, []sdk.Msg{f.msgSend()})
	deepExec := authz.NewMsgExec(f.other.addr, []sdk.Msg{&exec})

	cases := []struct {
		name   string
		signer account
		msgs   []sdk.Msg
	}{
		{"top level", f.voter, []sdk.Msg{f.msgSend()}},
		{"top level after an allowed msg", f.voter, []sdk.Msg{f.msgMultiSend(), f.msgSend()}},
		{"nested in authz MsgExec", f.other, []sdk.Msg{&exec}},
		{"nested two deep in authz MsgExec", f.other, []sdk.Msg{&deepExec}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := f.balanceAt(f.checkCtx(), tc.signer)
			seqBefore := f.app.AccountKeeper.GetAccount(f.checkCtx(), tc.signer.addr).GetSequence()

			res := f.checkTx(f.encode(tc.signer, tc.msgs...))
			require.Equal(t, safetytypes.ErrMsgDisabled.ABCICode(), res.Code, res.Log)
			require.Equal(t, safetytypes.ErrMsgDisabled.Codespace(), res.Codespace)
			require.Contains(t, res.Log, msgSendURL)

			require.Equal(t, before, f.balanceAt(f.checkCtx(), tc.signer), "no fee may be taken")
			require.Equal(t, seqBefore, f.app.AccountKeeper.GetAccount(f.checkCtx(), tc.signer.addr).GetSequence(), "sequence must not be used")
		})
	}

	// control: an allowed tx does pay its fee in CheckTx state, so the
	// balance checks above measure something
	before := f.balanceAt(f.checkCtx(), f.voter)
	res := f.checkTx(f.encode(f.voter, f.msgMultiSend()))
	require.Zero(t, res.Code, res.Log)
	require.Equal(t, before.Sub(txFee.AmountOf(sdk.DefaultBondDenom)), f.balanceAt(f.checkCtx(), f.voter))
}

// TestAnteCheckTxRejectionLeavesTxValidAfterReset shows that a CheckTx
// rejection is not final. The rejected tx keeps its sequence, so its signed
// bytes execute if anyone broadcasts them after the reset.
func TestAnteCheckTxRejectionLeavesTxValidAfterReset(t *testing.T) {
	f := newGateFixture(t, true)
	f.trip(msgSendURL)

	bz := f.encode(f.other, banktypes.NewMsgSend(f.other.addr, f.voter.addr, oneStake))
	res := f.checkTx(bz)
	require.Equal(t, safetytypes.ErrMsgDisabled.ABCICode(), res.Code, res.Log)

	f.reset(msgSendURL)
	before := f.balanceAt(f.ctx(), f.voter)
	r := f.deliverTxs(bz)[0]
	require.Zero(t, r.Code, r.Log)
	require.Equal(t, before.Add(oneStake.AmountOf(sdk.DefaultBondDenom)), f.balanceAt(f.ctx(), f.voter))
}

// TestAnteCheckTxRejectsGovProposalOfTrippedMsg checks that the decorator opens
// wrappers exposing GetMsgs, such as gov v1 MsgSubmitProposal: a proposal with
// a tripped msg is rejected at submission, while a proposal of protected msgs
// is admitted.
func TestAnteCheckTxRejectsGovProposalOfTrippedMsg(t *testing.T) {
	f := newGateFixture(t, true)
	f.trip(distrParamsURL)

	params, err := f.app.GovKeeper.Params.Get(f.ctx())
	require.NoError(t, err)

	tripped, err := govv1.NewMsgSubmitProposal([]sdk.Msg{f.msgDistrParams("0.33")}, params.MinDeposit, f.voter.String(), "", "t", "s", false)
	require.NoError(t, err)
	before := f.balanceAt(f.checkCtx(), f.voter)
	res := f.checkTx(f.encode(f.voter, tripped))
	require.Equal(t, safetytypes.ErrMsgDisabled.ABCICode(), res.Code,
		"gov MsgSubmitProposal with a tripped msg must fail CheckTx (the decorator opens GetMsgs wrappers); log: %s", res.Log)
	require.Equal(t, before, f.balanceAt(f.checkCtx(), f.voter), "no fee may be taken")

	protected, err := govv1.NewMsgSubmitProposal([]sdk.Msg{safetytypes.NewMsgReset(govAddr, distrParamsURL)}, params.MinDeposit, f.voter.String(), "", "t", "s", false)
	require.NoError(t, err)
	res = f.checkTx(f.encode(f.voter, protected))
	require.Zero(t, res.Code, res.Log)
}

// TestAnteFinalizeBlockLeavesTrippedMsgsToRouter shows the decorator stands
// aside in FinalizeBlock. A tx that passed CheckTx before a trip and executes
// after it in the same block, or that a proposer includes later, fails at the
// router like any failed tx: it pays its fee and uses its sequence. So once
// the trip is reset, the signed bytes cannot execute again.
func TestAnteFinalizeBlockLeavesTrippedMsgsToRouter(t *testing.T) {
	f := newGateFixture(t, true)
	f.grantSend()

	feeAmt := txFee.AmountOf(sdk.DefaultBondDenom)
	userSeq := f.seq(f.voter)
	granteeSeq := f.seq(f.other)
	userBefore := f.balanceAt(f.ctx(), f.voter)
	granteeBefore := f.balanceAt(f.ctx(), f.other)
	exec := authz.NewMsgExec(f.other.addr, []sdk.Msg{f.msgSend()})

	trip := f.encodeAt(f.voter, userSeq, safetytypes.NewMsgTrip(f.voter.String(), msgSendURL))
	send := f.encodeAt(f.voter, userSeq+1, f.msgSend())
	execTx := f.encodeAt(f.other, granteeSeq, &exec)
	for _, bz := range [][]byte{trip, send, execTx} {
		c := f.checkTx(bz)
		require.Zero(t, c.Code, c.Log)
	}

	res := f.deliverTxs(trip, send, execTx)
	require.Zero(t, res[0].Code, res[0].Log)
	requireRouterBlocked(t, res[1], msgSendURL)
	requireRouterBlocked(t, res[2], msgSendURL)

	// all three txs paid their fee and used their sequence; no stake moved
	require.Equal(t, userBefore.Sub(feeAmt.MulRaw(2)), f.balanceAt(f.ctx(), f.voter))
	require.Equal(t, granteeBefore.Sub(feeAmt), f.balanceAt(f.ctx(), f.other))
	require.Equal(t, userSeq+2, f.seq(f.voter))
	require.Equal(t, granteeSeq+1, f.seq(f.other))

	// a proposer that includes a tripped tx in a later block gets the same
	later := f.encode(f.voter, f.msgSend())
	requireRouterBlocked(t, f.deliverTxs(later)[0], msgSendURL)
	require.Equal(t, userSeq+3, f.seq(f.voter))

	// after the reset, the failed txs stay failed
	f.reset(msgSendURL)
	userAfter := f.balanceAt(f.ctx(), f.voter)
	for _, bz := range [][]byte{send, execTx, later} {
		require.Equal(t, sdkerrors.ErrWrongSequence.ABCICode(), f.checkTx(bz).Code)
		r := f.deliverTxs(bz)[0]
		require.Equal(t, sdkerrors.ErrWrongSequence.ABCICode(), r.Code, r.Log)
	}
	require.Equal(t, userAfter, f.balanceAt(f.ctx(), f.voter))
	require.Equal(t, granteeBefore.Sub(feeAmt), f.balanceAt(f.ctx(), f.other))
}

// TestAnteFinalizeBlockAdmitsGovProposalOfTrippedMsg shows the limit of
// rejection at submission. A gov proposal with a tripped msg that passed
// CheckTx before the trip and executes after it in the same block is created,
// since the router admits the protected gov MsgSubmitProposal and gov does not
// run the msgs it stores. The router blocks the tripped msg when the proposal
// executes while the trip stands, so the proposal fails.
func TestAnteFinalizeBlockAdmitsGovProposalOfTrippedMsg(t *testing.T) {
	f := newGateFixture(t, true)
	params, err := f.app.GovKeeper.Params.Get(f.ctx())
	require.NoError(t, err)
	taxBefore := f.communityTax()
	id, err := f.app.GovKeeper.ProposalID.Peek(f.ctx())
	require.NoError(t, err)

	seq := f.seq(f.voter)
	trip := f.encodeAt(f.voter, seq, safetytypes.NewMsgTrip(f.voter.String(), distrParamsURL))
	submit, err := govv1.NewMsgSubmitProposal([]sdk.Msg{f.msgDistrParams("0.33")}, params.MinDeposit, f.voter.String(), "", "t", "s", false)
	require.NoError(t, err)
	submitTx := f.encodeAt(f.voter, seq+1, submit)
	vote := f.encodeAt(f.voter, seq+2, govv1.NewMsgVote(f.voter.addr, id, govv1.OptionYes, ""))
	for _, bz := range [][]byte{trip, submitTx} {
		c := f.checkTx(bz)
		require.Zero(t, c.Code, c.Log)
	}

	res := f.deliverTxs(trip, submitTx, vote)
	for _, r := range res {
		require.Zero(t, r.Code, r.Log)
	}
	f.requireProposalStatus(id, govv1.StatusVotingPeriod)

	// the same proposal now fails CheckTx
	again := f.checkTx(f.encode(f.voter, submit))
	require.Equal(t, safetytypes.ErrMsgDisabled.ABCICode(), again.Code, again.Log)

	f.endVoting()
	p := f.requireProposalStatus(id, govv1.StatusFailed)
	require.Contains(t, p.FailedReason, distrParamsURL)
	require.True(t, taxBefore.Equal(f.communityTax()), "tripped proposal changed distribution params")
}

// TestAnteNestingDepthLimit sends authz MsgExec chains through CheckTx: every
// chain the tx decoder accepts passes the decorator, and the first one deeper
// fails to decode before any decorator runs.
func TestAnteNestingDepthLimit(t *testing.T) {
	f := newGateFixture(t, true)

	nest := func(wrappers int) sdk.Msg {
		msg := f.msgSend()
		for range wrappers {
			exec := authz.NewMsgExec(f.voter.addr, []sdk.Msg{msg})
			msg = &exec
		}
		return msg
	}

	// the tx body's Any is one unpack level, and each MsgExec adds one
	deepest := codectypes.MaxUnpackAnyRecursionDepth - 1
	res := f.checkTx(f.encode(f.voter, nest(deepest)))
	require.Zero(t, res.Code, res.Log)

	res = f.checkTx(f.encode(f.other, nest(deepest+1)))
	require.Equal(t, sdkerrors.ErrTxDecode.ABCICode(), res.Code, res.Log)
	require.Equal(t, sdkerrors.ErrTxDecode.Codespace(), res.Codespace)
}

// TestAntePassesProtectedMsgsUnderWildcard trips "*" with the decorator wired
// and shows CheckTx admits protected msgs and rejects everything else.
func TestAntePassesProtectedMsgsUnderWildcard(t *testing.T) {
	f := newGateFixture(t, true)
	f.trip(safetytypes.Wildcard)

	res := f.checkTx(f.encode(f.voter, f.msgDelegate()))
	require.Equal(t, safetytypes.ErrMsgDisabled.ABCICode(), res.Code, res.Log)

	params, err := f.app.GovKeeper.Params.Get(f.ctx())
	require.NoError(t, err)
	submit, err := govv1.NewMsgSubmitProposal([]sdk.Msg{safetytypes.NewMsgReset(govAddr, safetytypes.Wildcard)}, params.MinDeposit, f.voter.String(), "", "t", "s", false)
	require.NoError(t, err)
	res = f.checkTx(f.encode(f.voter, submit))
	require.Zero(t, res.Code, res.Log)

	// the authority resets "*" through gov while it stands
	id := f.submitProposal(safetytypes.NewMsgReset(govAddr, safetytypes.Wildcard))
	f.endVoting()
	f.requireProposalStatus(id, govv1.StatusPassed)
	require.Empty(t, f.disabled())

	res = f.checkTx(f.encode(f.voter, f.msgDelegate()))
	require.Zero(t, res.Code, res.Log)
}

// requireWrapped requires a tx with msgs, signed by a, to pass, or with
// blockedURL set, to fail on that type URL: in CheckTx when the decorator is
// wired, and at the router in FinalizeBlock. A tx CheckTx rejects is still
// delivered, since the decorator stands aside in FinalizeBlock and the router
// must refuse it on its own.
func (f *fixture) requireWrapped(withAnte bool, a account, blockedURL string, msgs ...sdk.Msg) {
	f.t.Helper()
	bz := f.encode(a, msgs...)
	if withAnte {
		res := f.checkTx(bz)
		if blockedURL == "" {
			require.Zero(f.t, res.Code, res.Log)
		} else {
			require.Equal(f.t, safetytypes.ErrMsgDisabled.ABCICode(), res.Code, res.Log)
			require.Equal(f.t, safetytypes.ErrMsgDisabled.Codespace(), res.Codespace, res.Log)
			require.Contains(f.t, res.Log, blockedURL)
		}
	}
	res := f.deliverTxs(bz)[0]
	if blockedURL == "" {
		require.Zero(f.t, res.Code, res.Log)
		return
	}
	requireRouterBlocked(f.t, res, blockedURL)
}

// TestWrapperRecoveryPaths shows that "*" and prefix trips leave authz
// MsgExec running while the router still checks every msg it dispatches, so a
// vote or a reset cast through a grantee succeeds under "*", and that only an
// exact trip of MsgExec blocks it.
func TestWrapperRecoveryPaths(t *testing.T) {
	for _, withAnte := range []bool{false, true} {
		t.Run(fmt.Sprintf("ante=%v", withAnte), func(t *testing.T) {
			f := newGateFixture(t, withAnte)
			for _, url := range []string{sdk.MsgTypeURL(&safetytypes.MsgReset{}), sdk.MsgTypeURL(&govv1.MsgVote{}), msgSendURL, authzRevokeURL} {
				grant, err := authz.NewMsgGrant(f.voter.addr, f.other.addr, authz.NewGenericAuthorization(url), nil)
				require.NoError(t, err)
				res := f.deliver(f.voter, grant)
				require.Zero(t, res.Code, res.Log)
			}
			execOf := func(msg sdk.Msg) *authz.MsgExec {
				e := authz.NewMsgExec(f.other.addr, []sdk.Msg{msg})
				return &e
			}
			run := func(blockedURL string, inner sdk.Msg) {
				t.Helper()
				f.requireWrapped(withAnte, f.other, blockedURL, execOf(inner))
			}
			revoke := authz.NewMsgRevoke(f.voter.addr, f.other.addr, msgSendURL)

			// selective trip: MsgExec runs, the router still checks what it dispatches
			f.trip("/cosmos.bank.*")
			run(msgSendURL, f.msgSend())
			id := f.submitProposal(safetytypes.NewMsgReset(govAddr, "/cosmos.bank.*"))
			run("", govv1.NewMsgVote(f.voter.addr, id, govv1.OptionYes, ""))
			run("", safetytypes.NewMsgReset(f.voter.String(), "/cosmos.bank.*"))
			require.Empty(t, f.disabled())

			// a trip of MsgExec's own package leaves MsgExec running and blocks
			// the authz msgs it dispatches
			f.trip("/cosmos.authz.*")
			run("", f.msgSend())
			run(authzRevokeURL, &revoke)
			f.reset("/cosmos.authz.*")

			// "*" leaves MsgExec running too: an unprotected inner msg is
			// blocked, a vote and the operator's reset through the grantee succeed
			f.trip(safetytypes.Wildcard)
			run(msgSendURL, f.msgSend())
			id = f.submitProposal(safetytypes.NewMsgReset(govAddr, safetytypes.Wildcard))
			run("", govv1.NewMsgVote(f.voter.addr, id, govv1.OptionNo, ""))
			run("", safetytypes.NewMsgReset(f.voter.String(), safetytypes.Wildcard))
			require.Empty(t, f.disabled())

			// an exact trip of MsgExec blocks it, and with it every reset sent
			// through the grantee; the operator resets directly
			f.trip(safetytypes.Wildcard, authzExecURL)
			run(authzExecURL, safetytypes.NewMsgReset(f.voter.String(), safetytypes.Wildcard))
			run(authzExecURL, govv1.NewMsgVote(f.voter.addr, id, govv1.OptionYes, ""))
			require.ElementsMatch(t, []string{safetytypes.Wildcard, authzExecURL}, f.disabled())
			f.reset(safetytypes.Wildcard, authzExecURL)
			require.Empty(t, f.disabled())
			run("", f.msgSend())
		})
	}
}

// TestOperatorCutsOffGrantee shows the operator stopping an authz grantee
// that trips through MsgExec. "*" does not block MsgExec, so the operator
// pins MsgExec by its exact URL while "*" stands, then resets "*" so the
// granter can revoke the grant, and the grantee cannot trip again in between.
func TestOperatorCutsOffGrantee(t *testing.T) {
	for _, withAnte := range []bool{false, true} {
		t.Run(fmt.Sprintf("ante=%v", withAnte), func(t *testing.T) {
			f := newGateFixture(t, withAnte)
			urlTrip := sdk.MsgTypeURL(&safetytypes.MsgTrip{})
			grant, err := authz.NewMsgGrant(f.voter.addr, f.other.addr, authz.NewGenericAuthorization(urlTrip), nil)
			require.NoError(t, err)
			res := f.deliver(f.voter, grant)
			require.Zero(t, res.Code, res.Log)

			retrip := func(p string) []byte {
				e := authz.NewMsgExec(f.other.addr, []sdk.Msg{safetytypes.NewMsgTrip(f.voter.String(), p)})
				return f.encode(f.other, &e)
			}
			res = f.deliverTxs(retrip(safetytypes.Wildcard))[0]
			require.Zero(t, res.Code, res.Log)
			require.Equal(t, []string{safetytypes.Wildcard}, f.disabled())

			// MsgExec still runs while "*" stands: the grantee's trip reaches
			// the msg server, which rejects a pattern already in the list
			res = f.deliverTxs(retrip(safetytypes.Wildcard))[0]
			require.Equal(t, safetytypes.ErrAlreadyDisabled.ABCICode(), res.Code, res.Log)
			require.Equal(t, safetytypes.ErrAlreadyDisabled.Codespace(), res.Codespace, res.Log)

			// pin MsgExec under "*"; CheckTx admits it, since MsgTrip is protected
			pin := f.encode(f.voter, safetytypes.NewMsgTrip(f.voter.String(), authzExecURL))
			cres := f.checkTx(pin)
			require.Zero(t, cres.Code, cres.Log)
			res = f.deliverTxs(pin)[0]
			require.Zero(t, res.Code, res.Log)
			require.ElementsMatch(t, []string{safetytypes.Wildcard, authzExecURL}, f.disabled())
			requireRouterBlocked(t, f.deliverTxs(retrip("/cosmos.bank.*"))[0], authzExecURL)

			// reset "*" so the granter can revoke; the grantee races the
			// revoke in the same block and loses
			f.reset(safetytypes.Wildcard)
			revoke := authz.NewMsgRevoke(f.voter.addr, f.other.addr, urlTrip)
			results := f.deliverTxs(retrip(safetytypes.Wildcard), f.encode(f.voter, &revoke))
			requireRouterBlocked(t, results[0], authzExecURL)
			require.Zero(t, results[1].Code, results[1].Log)

			f.reset(authzExecURL)
			res = f.deliverTxs(retrip(safetytypes.Wildcard))[0]
			require.NotZero(t, res.Code, "a revoked grantee must not trip")
			require.Empty(t, f.disabled())
		})
	}
}

// TestWildcardLeavesGranteeProtectedMsgs shows that under "*" a grantee can
// still run protected msgs through MsgExec, such as a gov deposit from the
// granter's funds, while the granter cannot revoke, and that an exact trip of
// MsgExec stops it.
func TestWildcardLeavesGranteeProtectedMsgs(t *testing.T) {
	for _, withAnte := range []bool{false, true} {
		t.Run(fmt.Sprintf("ante=%v", withAnte), func(t *testing.T) {
			f := newGateFixture(t, withAnte)
			urlDeposit := sdk.MsgTypeURL(&govv1.MsgDeposit{})
			grant, err := authz.NewMsgGrant(f.voter.addr, f.other.addr, authz.NewGenericAuthorization(urlDeposit), nil)
			require.NoError(t, err)
			res := f.deliver(f.voter, grant)
			require.Zero(t, res.Code, res.Log)

			// a proposal stays in its deposit period
			params, err := f.app.GovKeeper.Params.Get(f.ctx())
			require.NoError(t, err)
			id, err := f.app.GovKeeper.ProposalID.Peek(f.ctx())
			require.NoError(t, err)
			half := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, params.MinDeposit[0].Amount.QuoRaw(2)))
			submit, err := govv1.NewMsgSubmitProposal([]sdk.Msg{safetytypes.NewMsgReset(govAddr, safetytypes.Wildcard)}, half, f.other.String(), "", "t", "s", false)
			require.NoError(t, err)
			res = f.deliver(f.other, submit)
			require.Zero(t, res.Code, res.Log)

			f.trip(safetytypes.Wildcard)
			amt := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, params.MinDeposit[0].Amount.QuoRaw(8)))
			deposit := func() *authz.MsgExec {
				e := authz.NewMsgExec(f.other.addr, []sdk.Msg{govv1.NewMsgDeposit(f.voter.addr, id, amt)})
				return &e
			}
			before := f.balanceAt(f.ctx(), f.voter)
			f.requireWrapped(withAnte, f.other, "", deposit())
			require.Equal(t, before.Sub(amt.AmountOf(sdk.DefaultBondDenom)), f.balanceAt(f.ctx(), f.voter))

			revoke := authz.NewMsgRevoke(f.voter.addr, f.other.addr, urlDeposit)
			f.requireWrapped(withAnte, f.voter, authzRevokeURL, &revoke)

			f.trip(authzExecURL)
			before = f.balanceAt(f.ctx(), f.voter)
			f.requireWrapped(withAnte, f.other, authzExecURL, deposit())
			require.Equal(t, before, f.balanceAt(f.ctx(), f.voter))
		})
	}
}

// TestOverrideAuthorityWrapperLockout shows the limit of on-chain recovery
// when the consensus AuthorityParams override names an account that acts only
// through authz. The override, not params.authority, is the x/safety
// authority, so handing params.authority to gov gives gov nothing. Once the
// operator trips authz MsgExec exactly, gov can neither reset the trip, clear
// the override nor schedule an upgrade, because each checks the override.
// Only the operator's own reset restores the path.
func TestOverrideAuthorityWrapperLockout(t *testing.T) {
	f := newGateFixture(t, false)
	override := f.other

	cp, err := f.app.ConsensusParamsKeeper.ParamsStore.Get(f.ctx())
	require.NoError(t, err)
	setOverride := func(authority, to string) *consensustypes.MsgUpdateParams {
		return &consensustypes.MsgUpdateParams{
			Authority: authority, Block: cp.Block, Evidence: cp.Evidence,
			Validator: cp.Validator, Abci: cp.Abci,
			Auth: &cmtproto.AuthorityParams{Authority: to},
		}
	}
	f.requireProposalStatus(f.submitProposal(setOverride(govAddr, override.String())), govv1.StatusVotingPeriod)
	f.endVoting()

	grant, err := authz.NewMsgGrant(override.addr, f.voter.addr, authz.NewGenericAuthorization(sdk.MsgTypeURL(&safetytypes.MsgReset{})), nil)
	require.NoError(t, err)
	res := f.deliver(override, grant)
	require.Zero(t, res.Code, res.Log)

	params, err := f.app.SafetyKeeper.Params.Get(f.ctx())
	require.NoError(t, err)
	params.Authority = govAddr
	res = f.deliver(override, &safetytypes.MsgUpdateParams{Authority: override.String(), Params: params})
	require.Zero(t, res.Code, res.Log)

	f.trip(authzExecURL)
	exec := authz.NewMsgExec(f.voter.addr, []sdk.Msg{safetytypes.NewMsgReset(override.String(), authzExecURL)})
	requireRouterBlocked(t, f.deliver(f.voter, &exec), authzExecURL)

	ids := []uint64{
		f.submitProposal(safetytypes.NewMsgReset(govAddr, authzExecURL)),
		f.submitProposal(setOverride(govAddr, "")),
		f.submitProposal(&upgradetypes.MsgSoftwareUpgrade{
			Authority: govAddr,
			Plan:      upgradetypes.Plan{Name: "recover", Height: f.app.LastBlockHeight() + 100},
		}),
	}
	f.endVoting()
	for _, id := range ids {
		p := f.requireProposalStatus(id, govv1.StatusFailed)
		require.Contains(t, p.FailedReason, "unauthorized")
	}
	require.Equal(t, []string{authzExecURL}, f.disabled())

	// after the operator's reset, the override resets through authz again
	f.reset(authzExecURL)
	f.trip(msgSendURL)
	exec = authz.NewMsgExec(f.voter.addr, []sdk.Msg{safetytypes.NewMsgReset(override.String(), msgSendURL)})
	res = f.deliver(f.voter, &exec)
	require.Zero(t, res.Code, res.Log)
	require.Empty(t, f.disabled())
}
