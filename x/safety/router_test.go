package safety_test

import (
	"fmt"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	safetyante "github.com/cosmos/cosmos-sdk/x/safety/ante"
	"github.com/cosmos/cosmos-sdk/x/safety/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

const (
	urlSend        = "/cosmos.bank.v1beta1.MsgSend"
	urlMultiSend   = "/cosmos.bank.v1beta1.MsgMultiSend"
	urlDelegate    = "/cosmos.staking.v1beta1.MsgDelegate"
	urlDistrParams = "/cosmos.distribution.v1beta1.MsgUpdateParams"
	urlAuthzGrant  = "/cosmos.authz.v1beta1.MsgGrant"
)

// routerBlockedMsg is the text of the error the router returns for a disabled msg.
var routerBlockedMsg = types.ErrMsgDisabled.Error()

var oneStake = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1))

// requireRouterBlocked asserts that a tx failed at execution because the
// router circuit breaker refused typeURL, with the registered ErrMsgDisabled.
func requireRouterBlocked(t *testing.T, res *abci.ExecTxResult, typeURL string) {
	t.Helper()
	require.Equal(t, types.ErrMsgDisabled.Codespace(), res.Codespace, res.Log)
	require.Equal(t, types.ErrMsgDisabled.ABCICode(), res.Code, res.Log)
	require.Contains(t, res.Log, typeURL)
}

func (f *routerFixture) msgSend() sdk.Msg {
	return banktypes.NewMsgSend(f.user.addr, f.grantee.addr, oneStake)
}

func (f *routerFixture) msgMultiSend() sdk.Msg {
	return &banktypes.MsgMultiSend{
		Inputs:  []banktypes.Input{banktypes.NewInput(f.user.addr, oneStake)},
		Outputs: []banktypes.Output{banktypes.NewOutput(f.grantee.addr, oneStake)},
	}
}

func (f *routerFixture) msgDelegate() sdk.Msg {
	return stakingtypes.NewMsgDelegate(f.user.String(), f.validator().String(), sdk.NewInt64Coin(sdk.DefaultBondDenom, 1))
}

func (f *routerFixture) msgDistrParams(communityTax string) sdk.Msg {
	params, err := f.distrKeeper.Params.Get(f.ctx())
	require.NoError(f.t, err)
	params.CommunityTax = sdkmath.LegacyMustNewDecFromStr(communityTax)
	return &distrtypes.MsgUpdateParams{Authority: f.gov, Params: params}
}

func (f *routerFixture) communityTax() sdkmath.LegacyDec {
	params, err := f.distrKeeper.Params.Get(f.ctx())
	require.NoError(f.t, err)
	return params.CommunityTax
}

func (f *routerFixture) disabled() []string {
	var out []string
	require.NoError(f.t, f.safetyKeeper.DisabledList.Walk(f.ctx(), nil, func(p string) (bool, error) {
		out = append(out, p)
		return false, nil
	}))
	return out
}

// grantSend has user grant grantee a generic authorization for MsgSend.
func (f *routerFixture) grantSend() {
	f.t.Helper()
	grant, err := authz.NewMsgGrant(f.user.addr, f.grantee.addr, authz.NewGenericAuthorization(urlSend), nil)
	require.NoError(f.t, err)
	res := f.deliverOne(f.user, grant)
	require.Zero(f.t, res.Code, res.Log)
}

// TestRouterTrippedMsgFailsInFinalizeBlock shows the router is a gate on its
// own: with no SafetyDecorator, a tripped msg passes the ante handler and
// fails when the router dispatches it, and a reset restores it.
func TestRouterTrippedMsgFailsInFinalizeBlock(t *testing.T) {
	f := newRouterFixture(t, false)

	res := f.deliverOne(f.user, f.msgSend())
	require.Zero(t, res.Code, res.Log)

	f.trip(urlSend)
	require.Equal(t, []string{urlSend}, f.disabled())

	seqBefore := f.seq(f.user)
	balBefore := f.balance(f.ctx(), f.grantee)
	res = f.deliverOne(f.user, f.msgSend())
	requireRouterBlocked(t, res, urlSend)
	require.Equal(t, balBefore, f.balance(f.ctx(), f.grantee), "blocked send moved funds")
	// without the decorator the ante handler ran, so the sequence was used
	require.Equal(t, seqBefore+1, f.seq(f.user))

	// an exact trip leaves sibling msgs alone
	res = f.deliverOne(f.user, f.msgMultiSend())
	require.Zero(t, res.Code, res.Log)

	// a tx is atomic: an allowed msg next to a tripped one is rolled back
	balBefore = f.balance(f.ctx(), f.grantee)
	res = f.deliverOne(f.user, f.msgMultiSend(), f.msgSend())
	requireRouterBlocked(t, res, urlSend)
	require.Equal(t, balBefore, f.balance(f.ctx(), f.grantee))

	f.reset(urlSend)
	require.Empty(t, f.disabled())
	res = f.deliverOne(f.user, f.msgSend())
	require.Zero(t, res.Code, res.Log)
}

// TestRouterAuthzExecOfTrippedMsgFails shows a tripped msg nested in authz
// MsgExec fails at execution, because authz dispatches through the same
// router handler. MsgExec itself is not tripped.
func TestRouterAuthzExecOfTrippedMsgFails(t *testing.T) {
	f := newRouterFixture(t, false)
	f.grantSend()

	exec := authz.NewMsgExec(f.grantee.addr, []sdk.Msg{f.msgSend()})
	res := f.deliverOne(f.grantee, &exec)
	require.Zero(t, res.Code, "exec before trip: %s", res.Log)

	f.trip(urlSend)

	balBefore := f.balance(f.ctx(), f.user)
	res = f.deliverOne(f.grantee, &exec)
	requireRouterBlocked(t, res, urlSend)
	require.Equal(t, balBefore, f.balance(f.ctx(), f.user), "granter funds moved")

	// a doubly nested exec is refused too
	inner := authz.NewMsgExec(f.grantee.addr, []sdk.Msg{f.msgSend()})
	outer := authz.NewMsgExec(f.grantee.addr, []sdk.Msg{&inner})
	res = f.deliverOne(f.grantee, &outer)
	require.NotZero(t, res.Code, res.Log)

	f.reset(urlSend)
	res = f.deliverOne(f.grantee, &exec)
	require.Zero(t, res.Code, "exec after reset: %s", res.Log)
}

// TestRouterGovProposalOfTrippedMsgFails shows a proposal that executes a
// tripped msg fails at execution, all of its msgs roll back, and a proposal
// of protected msgs still executes while the trip stands.
func TestRouterGovProposalOfTrippedMsgFails(t *testing.T) {
	f := newRouterFixture(t, false)
	taxBefore := f.communityTax()
	govParams, err := f.govKeeper.Params.Get(f.ctx())
	require.NoError(t, err)

	// both proposals are submitted while nothing is tripped, then the operator
	// trips during the voting period
	only := f.submitProposal(f.msgDistrParams("0.33"))
	newGovParams := govParams
	newGovParams.BurnVoteVeto = !govParams.BurnVoteVeto
	mixed := f.submitProposal(&govv1.MsgUpdateParams{Authority: f.gov, Params: newGovParams}, f.msgDistrParams("0.34"))
	f.trip(urlDistrParams)
	f.endVoting()

	for _, id := range []uint64{only, mixed} {
		p := f.requireProposalStatus(id, govv1.StatusFailed)
		require.Contains(t, p.FailedReason, routerBlockedMsg, "proposal %d", id)
		require.Contains(t, p.FailedReason, urlDistrParams, "proposal %d", id)
	}
	require.True(t, taxBefore.Equal(f.communityTax()), "tripped proposal changed distribution params")
	got, err := f.govKeeper.Params.Get(f.ctx())
	require.NoError(t, err)
	require.Equal(t, govParams.BurnVoteVeto, got.BurnVoteVeto, "the gov msg in a failed proposal must roll back")

	// protected msgs execute through gov while the trip stands: gov params,
	// and the authority resetting the trip
	protected := f.submitProposal(
		&govv1.MsgUpdateParams{Authority: f.gov, Params: newGovParams},
		types.NewMsgReset(f.gov, urlDistrParams),
	)
	require.Equal(t, []string{urlDistrParams}, f.disabled(), "the trip must stand until the proposal executes")
	f.endVoting()
	f.requireProposalStatus(protected, govv1.StatusPassed)
	got, err = f.govKeeper.Params.Get(f.ctx())
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
	f := newRouterFixture(t, false)
	f.grantSend()
	f.trip(types.Wildcard)

	// sweep every registered msg type through its real router handler: the
	// breaker refuses exactly those outside the protected and wrapper sets
	ctx := f.ctx()
	registry := f.registry
	router := f.app.MsgServiceRouter()
	var blocked, protected int
	for _, typeURL := range registry.ListImplementations(sdk.MsgInterfaceProtoName) {
		passes := types.IsProtected(typeURL) || wrapperTypeURLs[typeURL]
		allowed, err := f.safetyKeeper.IsAllowed(ctx, typeURL)
		if passes {
			require.NoError(t, err, typeURL)
		} else {
			require.ErrorIs(t, err, types.ErrMsgDisabled, typeURL)
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
		urlSend:      f.msgSend(),
		urlMultiSend: f.msgMultiSend(),
		urlDelegate:  f.msgDelegate(),
		urlAuthzGrant: func() sdk.Msg {
			m, err := authz.NewMsgGrant(f.user.addr, f.grantee.addr, authz.NewGenericAuthorization(urlDelegate), nil)
			require.NoError(t, err)
			return m
		}(),
	} {
		res := f.deliverOne(f.user, msg)
		requireRouterBlocked(t, res, name)
	}
	// MsgExec passes the breaker; the router blocks the msg it dispatches
	exec := authz.NewMsgExec(f.grantee.addr, []sdk.Msg{f.msgSend()})
	res := f.deliverOne(f.grantee, &exec)
	requireRouterBlocked(t, res, urlSend)

	// gov still takes proposals and votes, and executes protected msgs
	govParams, err := f.govKeeper.Params.Get(f.ctx())
	require.NoError(t, err)
	newGovParams := govParams
	newGovParams.BurnVoteVeto = !govParams.BurnVoteVeto

	newSafetyParams := types.Params{Authority: f.gov, Operator: f.grantee.String(), Scope: []string{types.Wildcard}}

	cp := sims.DefaultConsensusParams
	newBlock := &cmtproto.BlockParams{MaxBytes: cp.Block.MaxBytes, MaxGas: cp.Block.MaxGas - 1}
	consensusMsg := &consensustypes.MsgUpdateParams{
		Authority: f.gov,
		Block:     newBlock,
		Evidence:  cp.Evidence,
		Validator: cp.Validator,
		Auth:      cp.Authority,
	}

	plan := upgradetypes.Plan{Name: "safety-test", Height: f.app.LastBlockHeight() + 1_000}
	upgradeMsg := &upgradetypes.MsgSoftwareUpgrade{Authority: f.gov, Plan: plan}

	id := f.submitProposal(
		&govv1.MsgUpdateParams{Authority: f.gov, Params: newGovParams},
		&types.MsgUpdateParams{Authority: f.gov, Params: newSafetyParams},
		consensusMsg,
		upgradeMsg,
	)
	failing := f.submitProposal(f.msgDistrParams("0.33"))
	f.endVoting()

	f.requireProposalStatus(id, govv1.StatusPassed)
	p := f.requireProposalStatus(failing, govv1.StatusFailed)
	require.Contains(t, p.FailedReason, routerBlockedMsg)
	require.Contains(t, p.FailedReason, urlDistrParams)

	ctx = f.ctx()
	gotGov, err := f.govKeeper.Params.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, newGovParams.BurnVoteVeto, gotGov.BurnVoteVeto)
	gotSafety, err := f.safetyKeeper.Params.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, newSafetyParams, gotSafety)
	gotConsensus, err := f.consensusKeeper.ParamsStore.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, newBlock.MaxGas, gotConsensus.Block.MaxGas)
	gotPlan, err := f.upgradeKeeper.GetUpgradePlan(ctx)
	require.NoError(t, err)
	require.Equal(t, plan.Name, gotPlan.Name)
	require.Equal(t, plan.Height, gotPlan.Height)
	require.Equal(t, []string{types.Wildcard}, f.disabled(), "MsgUpdateParams must not change the disabled list")

	// the old operator lost its role; the new operator resets "*"
	res = f.deliverOne(f.user, types.NewMsgReset(f.user.String(), types.Wildcard))
	require.Equal(t, sdkerrors.ErrUnauthorized.ABCICode(), res.Code, res.Log)
	require.Equal(t, sdkerrors.ErrUnauthorized.Codespace(), res.Codespace)

	res = f.deliverOne(f.grantee, types.NewMsgReset(f.grantee.String(), types.Wildcard))
	require.Zero(t, res.Code, res.Log)
	require.Empty(t, f.disabled())

	res = f.deliverOne(f.user, f.msgSend())
	require.Zero(t, res.Code, res.Log)
}

// TestRouterPrefixTrips trips prefix patterns and checks they block exactly
// the msgs under the prefix, at segment boundaries, and never protected msgs.
func TestRouterPrefixTrips(t *testing.T) {
	f := newRouterFixture(t, false)

	f.trip("/cosmos.bank.*")
	res := f.deliverOne(f.user, f.msgSend())
	requireRouterBlocked(t, res, urlSend)
	res = f.deliverOne(f.user, f.msgMultiSend())
	requireRouterBlocked(t, res, urlMultiSend)
	res = f.deliverOne(f.user, f.msgDelegate())
	require.Zero(t, res.Code, "staking is outside /cosmos.bank.*: %s", res.Log)

	// a prefix trip does not absorb an exact reset: the exact pattern is not
	// in the list
	res = f.deliverOne(f.user, types.NewMsgReset(f.user.String(), urlSend))
	require.Equal(t, types.ErrNotDisabled.ABCICode(), res.Code, res.Log)
	f.reset("/cosmos.bank.*")

	f.trip("/cosmos.staking.v1beta1.*")
	res = f.deliverOne(f.user, f.msgDelegate())
	requireRouterBlocked(t, res, urlDelegate)
	res = f.deliverOne(f.user, f.msgSend())
	require.Zero(t, res.Code, res.Log)
	f.reset("/cosmos.staking.v1beta1.*")

	// the widest prefix still leaves protected msgs alone: x/safety itself
	// (the reset below) and gov (a proposal and a vote)
	f.trip("/cosmos.*")
	for url, msg := range map[string]sdk.Msg{urlSend: f.msgSend(), urlDelegate: f.msgDelegate()} {
		res := f.deliverOne(f.user, msg)
		requireRouterBlocked(t, res, url)
	}
	id := f.submitProposal(f.msgDistrParams("0.33"))
	f.endVoting()
	p := f.requireProposalStatus(id, govv1.StatusFailed)
	require.Contains(t, p.FailedReason, urlDistrParams)

	// tripping a disabled pattern again is rejected
	res = f.deliverOne(f.user, types.NewMsgTrip(f.user.String(), "/cosmos.*"))
	require.Equal(t, types.ErrAlreadyDisabled.ABCICode(), res.Code, res.Log)
	f.reset("/cosmos.*")

	// a prefix tripped over an existing exact entry is accepted, and the exact
	// entry keeps MsgSend blocked after the prefix is reset
	f.trip(urlSend)
	f.trip("/cosmos.*")
	f.reset("/cosmos.*")
	res = f.deliverOne(f.user, f.msgSend())
	requireRouterBlocked(t, res, urlSend)
	res = f.deliverOne(f.user, f.msgDelegate())
	require.Zero(t, res.Code, res.Log)
	f.reset(urlSend)
	res = f.deliverOne(f.user, f.msgSend())
	require.Zero(t, res.Code, res.Log)
}

// TestRouterTripTakesEffectInSameBlock puts a trip and the msg it blocks in
// one block: txs before the trip run, txs after it fail, and a reset later
// in the block lifts it for the txs after the reset.
func TestRouterTripTakesEffectInSameBlock(t *testing.T) {
	f := newRouterFixture(t, false)
	seq := f.seq(f.user)
	before := f.balance(f.ctx(), f.grantee)

	res := f.deliver(
		f.sign(f.user, seq, nil, f.msgSend()),
		f.sign(f.user, seq+1, nil, types.NewMsgTrip(f.user.String(), urlSend)),
		f.sign(f.user, seq+2, nil, f.msgSend()),
		f.sign(f.user, seq+3, nil, f.msgMultiSend()),
		f.sign(f.user, seq+4, nil, types.NewMsgReset(f.user.String(), urlSend)),
		f.sign(f.user, seq+5, nil, f.msgSend()),
	)
	require.Zero(t, res[0].Code, res[0].Log)
	require.Zero(t, res[1].Code, res[1].Log)
	requireRouterBlocked(t, res[2], urlSend)
	require.Zero(t, res[3].Code, res[3].Log)
	require.Zero(t, res[4].Code, res[4].Log)
	require.Zero(t, res[5].Code, res[5].Log)

	// three sends of one stake succeeded: the first, the multisend, the last
	require.Equal(t, before.AddRaw(3), f.balance(f.ctx(), f.grantee))
	require.Empty(t, f.disabled())
}

// TestRouterGatesHybridHandlers checks that a hybrid handler runs the
// x/safety check, as the routed handler does.
func TestRouterGatesHybridHandlers(t *testing.T) {
	f := newRouterFixture(t, false)
	handler := f.app.MsgServiceRouter().HybridHandlerByMsgName(urlSend[1:])
	require.NotNil(t, handler)

	send := banktypes.NewMsgSend(sdk.AccAddress("from________________"), f.grantee.addr, oneStake)
	f.trip(urlSend)
	require.ErrorIs(t, handler(f.ctx(), send, &banktypes.MsgSendResponse{}), types.ErrMsgDisabled)

	f.reset(urlSend)
	require.ErrorIs(t, handler(f.ctx(), send, &banktypes.MsgSendResponse{}), sdkerrors.ErrInsufficientFunds)
}

// TestAnteCheckTxRejectsTrippedMsgs wires the SafetyDecorator and shows
// CheckTx turns away a tripped top-level msg, an authz-nested one, and a
// proposal with one, before the fee is deducted or the sequence used.
func TestAnteCheckTxRejectsTrippedMsgs(t *testing.T) {
	f := newRouterFixture(t, true)
	f.grantSend()
	f.trip(urlSend)

	fee := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000))
	exec := authz.NewMsgExec(f.grantee.addr, []sdk.Msg{f.msgSend()})
	deepExec := authz.NewMsgExec(f.grantee.addr, []sdk.Msg{&exec})

	cases := []struct {
		name   string
		signer testAccount
		msgs   []sdk.Msg
	}{
		{"top level", f.user, []sdk.Msg{f.msgSend()}},
		{"top level after an allowed msg", f.user, []sdk.Msg{f.msgMultiSend(), f.msgSend()}},
		{"nested in authz MsgExec", f.grantee, []sdk.Msg{&exec}},
		{"nested two deep in authz MsgExec", f.grantee, []sdk.Msg{&deepExec}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := f.balance(f.checkCtx(), tc.signer)
			seqBefore := f.accountKeeper.GetAccount(f.checkCtx(), tc.signer.addr).GetSequence()

			res := f.checkTx(f.signNext(tc.signer, fee, tc.msgs...))
			require.Equal(t, types.ErrMsgDisabled.ABCICode(), res.Code, res.Log)
			require.Equal(t, types.ErrMsgDisabled.Codespace(), res.Codespace)
			require.Contains(t, res.Log, urlSend)

			require.Equal(t, before, f.balance(f.checkCtx(), tc.signer), "no fee may be taken")
			require.Equal(t, seqBefore, f.accountKeeper.GetAccount(f.checkCtx(), tc.signer.addr).GetSequence(), "sequence must not be used")
		})
	}

	// control: an allowed tx does pay its fee in CheckTx state, so the
	// balance checks above measure something
	before := f.balance(f.checkCtx(), f.user)
	res := f.checkTx(f.signNext(f.user, fee, f.msgMultiSend()))
	require.Zero(t, res.Code, res.Log)
	require.Equal(t, before.Sub(fee.AmountOf(sdk.DefaultBondDenom)), f.balance(f.checkCtx(), f.user))
}

// TestAnteCheckTxRejectionLeavesTxValidAfterReset shows that a CheckTx
// rejection is not final. The rejected tx keeps its sequence, so its signed
// bytes execute if anyone broadcasts them after the reset.
func TestAnteCheckTxRejectionLeavesTxValidAfterReset(t *testing.T) {
	f := newRouterFixture(t, true)
	f.trip(urlSend)

	bz := f.signNext(f.grantee, nil, banktypes.NewMsgSend(f.grantee.addr, f.user.addr, oneStake))
	res := f.checkTx(bz)
	require.Equal(t, types.ErrMsgDisabled.ABCICode(), res.Code, res.Log)

	f.reset(urlSend)
	before := f.balance(f.ctx(), f.user)
	r := f.deliver(bz)[0]
	require.Zero(t, r.Code, r.Log)
	require.Equal(t, before.Add(oneStake.AmountOf(sdk.DefaultBondDenom)), f.balance(f.ctx(), f.user))
}

// TestAnteCheckTxRejectsGovProposalOfTrippedMsg checks that the decorator opens
// wrappers exposing GetMsgs, such as gov v1 MsgSubmitProposal: a proposal with
// a tripped msg is rejected at submission, while a proposal of protected msgs
// is admitted.
func TestAnteCheckTxRejectsGovProposalOfTrippedMsg(t *testing.T) {
	f := newRouterFixture(t, true)
	f.trip(urlDistrParams)

	params, err := f.govKeeper.Params.Get(f.ctx())
	require.NoError(t, err)
	fee := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000))

	tripped, err := govv1.NewMsgSubmitProposal([]sdk.Msg{f.msgDistrParams("0.33")}, params.MinDeposit, f.user.String(), "", "t", "s", false)
	require.NoError(t, err)
	before := f.balance(f.checkCtx(), f.user)
	res := f.checkTx(f.signNext(f.user, fee, tripped))
	require.Equal(t, types.ErrMsgDisabled.ABCICode(), res.Code,
		"gov MsgSubmitProposal with a tripped msg must fail CheckTx (the decorator opens GetMsgs wrappers); log: %s", res.Log)
	require.Equal(t, before, f.balance(f.checkCtx(), f.user), "no fee may be taken")

	protected, err := govv1.NewMsgSubmitProposal([]sdk.Msg{types.NewMsgReset(f.gov, urlDistrParams)}, params.MinDeposit, f.user.String(), "", "t", "s", false)
	require.NoError(t, err)
	res = f.checkTx(f.signNext(f.user, fee, protected))
	require.Zero(t, res.Code, res.Log)
}

// TestAnteFinalizeBlockLeavesTrippedMsgsToRouter shows the decorator stands
// aside in FinalizeBlock. A tx that passed CheckTx before a trip and executes
// after it in the same block, or that a proposer includes later, fails at the
// router like any failed tx: it pays its fee and uses its sequence. So once
// the trip is reset, the signed bytes cannot execute again.
func TestAnteFinalizeBlockLeavesTrippedMsgsToRouter(t *testing.T) {
	f := newRouterFixture(t, true)
	f.grantSend()

	fee := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000))
	feeAmt := fee.AmountOf(sdk.DefaultBondDenom)
	userSeq := f.seq(f.user)
	granteeSeq := f.seq(f.grantee)
	userBefore := f.balance(f.ctx(), f.user)
	granteeBefore := f.balance(f.ctx(), f.grantee)
	exec := authz.NewMsgExec(f.grantee.addr, []sdk.Msg{f.msgSend()})

	trip := f.sign(f.user, userSeq, fee, types.NewMsgTrip(f.user.String(), urlSend))
	send := f.sign(f.user, userSeq+1, fee, f.msgSend())
	execTx := f.sign(f.grantee, granteeSeq, fee, &exec)
	for _, bz := range [][]byte{trip, send, execTx} {
		c := f.checkTx(bz)
		require.Zero(t, c.Code, c.Log)
	}

	res := f.deliver(trip, send, execTx)
	require.Zero(t, res[0].Code, res[0].Log)
	requireRouterBlocked(t, res[1], urlSend)
	requireRouterBlocked(t, res[2], urlSend)

	// all three txs paid their fee and used their sequence; no stake moved
	require.Equal(t, userBefore.Sub(feeAmt.MulRaw(2)), f.balance(f.ctx(), f.user))
	require.Equal(t, granteeBefore.Sub(feeAmt), f.balance(f.ctx(), f.grantee))
	require.Equal(t, userSeq+2, f.seq(f.user))
	require.Equal(t, granteeSeq+1, f.seq(f.grantee))

	// a proposer that includes a tripped tx in a later block gets the same
	later := f.signNext(f.user, fee, f.msgSend())
	requireRouterBlocked(t, f.deliver(later)[0], urlSend)
	require.Equal(t, userSeq+3, f.seq(f.user))

	// after the reset, the failed txs stay failed
	f.reset(urlSend)
	userAfter := f.balance(f.ctx(), f.user)
	for _, bz := range [][]byte{send, execTx, later} {
		require.Equal(t, sdkerrors.ErrWrongSequence.ABCICode(), f.checkTx(bz).Code)
		r := f.deliver(bz)[0]
		require.Equal(t, sdkerrors.ErrWrongSequence.ABCICode(), r.Code, r.Log)
	}
	require.Equal(t, userAfter, f.balance(f.ctx(), f.user))
	require.Equal(t, granteeBefore.Sub(feeAmt), f.balance(f.ctx(), f.grantee))
}

// TestAnteFinalizeBlockAdmitsGovProposalOfTrippedMsg shows the limit of
// rejection at submission. A gov proposal with a tripped msg that passed
// CheckTx before the trip and executes after it in the same block is created,
// since the router admits the protected gov MsgSubmitProposal and gov does not
// run the msgs it stores. The router blocks the tripped msg when the proposal
// executes while the trip stands, so the proposal fails.
func TestAnteFinalizeBlockAdmitsGovProposalOfTrippedMsg(t *testing.T) {
	f := newRouterFixture(t, true)
	params, err := f.govKeeper.Params.Get(f.ctx())
	require.NoError(t, err)
	fee := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000))
	taxBefore := f.communityTax()
	id, err := f.govKeeper.ProposalID.Peek(f.ctx())
	require.NoError(t, err)

	seq := f.seq(f.user)
	trip := f.sign(f.user, seq, fee, types.NewMsgTrip(f.user.String(), urlDistrParams))
	submit, err := govv1.NewMsgSubmitProposal([]sdk.Msg{f.msgDistrParams("0.33")}, params.MinDeposit, f.user.String(), "", "t", "s", false)
	require.NoError(t, err)
	submitTx := f.sign(f.user, seq+1, fee, submit)
	vote := f.sign(f.user, seq+2, fee, govv1.NewMsgVote(f.user.addr, id, govv1.OptionYes, ""))
	for _, bz := range [][]byte{trip, submitTx} {
		c := f.checkTx(bz)
		require.Zero(t, c.Code, c.Log)
	}

	res := f.deliver(trip, submitTx, vote)
	for _, r := range res {
		require.Zero(t, r.Code, r.Log)
	}
	f.requireProposalStatus(id, govv1.StatusVotingPeriod)

	// the same proposal now fails CheckTx
	again := f.checkTx(f.signNext(f.user, fee, submit))
	require.Equal(t, types.ErrMsgDisabled.ABCICode(), again.Code, again.Log)

	f.endVoting()
	p := f.requireProposalStatus(id, govv1.StatusFailed)
	require.Contains(t, p.FailedReason, urlDistrParams)
	require.True(t, taxBefore.Equal(f.communityTax()), "tripped proposal changed distribution params")
}

// TestAnteNestingDepthLimit sends authz MsgExec chains through CheckTx: a
// chain at the depth limit passes, one level deeper is rejected.
func TestAnteNestingDepthLimit(t *testing.T) {
	f := newRouterFixture(t, true)

	nest := func(wrappers int) sdk.Msg {
		msg := f.msgSend()
		for range wrappers {
			exec := authz.NewMsgExec(f.user.addr, []sdk.Msg{msg})
			msg = &exec
		}
		return msg
	}

	res := f.checkTx(f.signNext(f.user, nil, nest(safetyante.MaxNestedMsgDepth)))
	require.Zero(t, res.Code, res.Log)

	res = f.checkTx(f.signNext(f.grantee, nil, nest(safetyante.MaxNestedMsgDepth+1)))
	require.Equal(t, types.ErrMaxDepth.ABCICode(), res.Code, res.Log)
	require.Equal(t, types.ErrMaxDepth.Codespace(), res.Codespace)
}

// TestAntePassesProtectedMsgsUnderWildcard trips "*" with the decorator wired
// and shows CheckTx admits protected msgs and rejects everything else.
func TestAntePassesProtectedMsgsUnderWildcard(t *testing.T) {
	f := newRouterFixture(t, true)
	f.trip(types.Wildcard)

	res := f.checkTx(f.signNext(f.user, nil, f.msgDelegate()))
	require.Equal(t, types.ErrMsgDisabled.ABCICode(), res.Code, res.Log)

	params, err := f.govKeeper.Params.Get(f.ctx())
	require.NoError(t, err)
	submit, err := govv1.NewMsgSubmitProposal([]sdk.Msg{types.NewMsgReset(f.gov, types.Wildcard)}, params.MinDeposit, f.user.String(), "", "t", "s", false)
	require.NoError(t, err)
	res = f.checkTx(f.signNext(f.user, nil, submit))
	require.Zero(t, res.Code, res.Log)

	// the authority resets "*" through gov while it stands
	id := f.submitProposal(types.NewMsgReset(f.gov, types.Wildcard))
	f.endVoting()
	f.requireProposalStatus(id, govv1.StatusPassed)
	require.Empty(t, f.disabled())

	res = f.checkTx(f.signNext(f.user, nil, f.msgDelegate()))
	require.Zero(t, res.Code, res.Log)
}

const (
	urlAuthzExec   = "/cosmos.authz.v1beta1.MsgExec"
	urlAuthzRevoke = "/cosmos.authz.v1beta1.MsgRevoke"
)

// requireWrapped requires a tx with msgs, signed by a, to pass, or with
// blockedURL set, to fail on that type URL: in CheckTx when the decorator is
// wired, and at the router in FinalizeBlock. A tx CheckTx rejects is still
// delivered, since the decorator stands aside in FinalizeBlock and the router
// must refuse it on its own.
func (f *routerFixture) requireWrapped(withAnte bool, a testAccount, blockedURL string, msgs ...sdk.Msg) {
	f.t.Helper()
	bz := f.signNext(a, nil, msgs...)
	if withAnte {
		res := f.checkTx(bz)
		if blockedURL == "" {
			require.Zero(f.t, res.Code, res.Log)
		} else {
			require.Equal(f.t, types.ErrMsgDisabled.ABCICode(), res.Code, res.Log)
			require.Equal(f.t, types.ErrMsgDisabled.Codespace(), res.Codespace, res.Log)
			require.Contains(f.t, res.Log, blockedURL)
		}
	}
	res := f.deliver(bz)[0]
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
			f := newRouterFixture(t, withAnte)
			for _, url := range []string{sdk.MsgTypeURL(&types.MsgReset{}), sdk.MsgTypeURL(&govv1.MsgVote{}), urlSend, urlAuthzRevoke} {
				grant, err := authz.NewMsgGrant(f.user.addr, f.grantee.addr, authz.NewGenericAuthorization(url), nil)
				require.NoError(t, err)
				res := f.deliverOne(f.user, grant)
				require.Zero(t, res.Code, res.Log)
			}
			execOf := func(msg sdk.Msg) *authz.MsgExec {
				e := authz.NewMsgExec(f.grantee.addr, []sdk.Msg{msg})
				return &e
			}
			run := func(blockedURL string, inner sdk.Msg) {
				t.Helper()
				f.requireWrapped(withAnte, f.grantee, blockedURL, execOf(inner))
			}
			revoke := authz.NewMsgRevoke(f.user.addr, f.grantee.addr, urlSend)

			// selective trip: MsgExec runs, the router still checks what it dispatches
			f.trip("/cosmos.bank.*")
			run(urlSend, f.msgSend())
			id := f.submitProposal(types.NewMsgReset(f.gov, "/cosmos.bank.*"))
			run("", govv1.NewMsgVote(f.user.addr, id, govv1.OptionYes, ""))
			run("", types.NewMsgReset(f.user.String(), "/cosmos.bank.*"))
			require.Empty(t, f.disabled())

			// a trip of MsgExec's own package leaves MsgExec running and blocks
			// the authz msgs it dispatches
			f.trip("/cosmos.authz.*")
			run("", f.msgSend())
			run(urlAuthzRevoke, &revoke)
			f.reset("/cosmos.authz.*")

			// "*" leaves MsgExec running too: an unprotected inner msg is
			// blocked, a vote and the operator's reset through the grantee succeed
			f.trip(types.Wildcard)
			run(urlSend, f.msgSend())
			id = f.submitProposal(types.NewMsgReset(f.gov, types.Wildcard))
			run("", govv1.NewMsgVote(f.user.addr, id, govv1.OptionNo, ""))
			run("", types.NewMsgReset(f.user.String(), types.Wildcard))
			require.Empty(t, f.disabled())

			// an exact trip of MsgExec blocks it, and with it every reset sent
			// through the grantee; the operator resets directly
			f.trip(types.Wildcard, urlAuthzExec)
			run(urlAuthzExec, types.NewMsgReset(f.user.String(), types.Wildcard))
			run(urlAuthzExec, govv1.NewMsgVote(f.user.addr, id, govv1.OptionYes, ""))
			require.ElementsMatch(t, []string{types.Wildcard, urlAuthzExec}, f.disabled())
			f.reset(types.Wildcard, urlAuthzExec)
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
			f := newRouterFixture(t, withAnte)
			urlTrip := sdk.MsgTypeURL(&types.MsgTrip{})
			grant, err := authz.NewMsgGrant(f.user.addr, f.grantee.addr, authz.NewGenericAuthorization(urlTrip), nil)
			require.NoError(t, err)
			res := f.deliverOne(f.user, grant)
			require.Zero(t, res.Code, res.Log)

			retrip := func(p string) []byte {
				e := authz.NewMsgExec(f.grantee.addr, []sdk.Msg{types.NewMsgTrip(f.user.String(), p)})
				return f.signNext(f.grantee, nil, &e)
			}
			res = f.deliver(retrip(types.Wildcard))[0]
			require.Zero(t, res.Code, res.Log)
			require.Equal(t, []string{types.Wildcard}, f.disabled())

			// MsgExec still runs while "*" stands: the grantee's trip reaches
			// the msg server, which rejects a pattern already in the list
			res = f.deliver(retrip(types.Wildcard))[0]
			require.Equal(t, types.ErrAlreadyDisabled.ABCICode(), res.Code, res.Log)
			require.Equal(t, types.ErrAlreadyDisabled.Codespace(), res.Codespace, res.Log)

			// pin MsgExec under "*"; CheckTx admits it, since MsgTrip is protected
			pin := f.signNext(f.user, nil, types.NewMsgTrip(f.user.String(), urlAuthzExec))
			cres := f.checkTx(pin)
			require.Zero(t, cres.Code, cres.Log)
			res = f.deliver(pin)[0]
			require.Zero(t, res.Code, res.Log)
			require.ElementsMatch(t, []string{types.Wildcard, urlAuthzExec}, f.disabled())
			requireRouterBlocked(t, f.deliver(retrip("/cosmos.bank.*"))[0], urlAuthzExec)

			// reset "*" so the granter can revoke; the grantee races the
			// revoke in the same block and loses
			f.reset(types.Wildcard)
			revoke := authz.NewMsgRevoke(f.user.addr, f.grantee.addr, urlTrip)
			results := f.deliver(retrip(types.Wildcard), f.signNext(f.user, nil, &revoke))
			requireRouterBlocked(t, results[0], urlAuthzExec)
			require.Zero(t, results[1].Code, results[1].Log)

			f.reset(urlAuthzExec)
			res = f.deliver(retrip(types.Wildcard))[0]
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
			f := newRouterFixture(t, withAnte)
			urlDeposit := sdk.MsgTypeURL(&govv1.MsgDeposit{})
			grant, err := authz.NewMsgGrant(f.user.addr, f.grantee.addr, authz.NewGenericAuthorization(urlDeposit), nil)
			require.NoError(t, err)
			res := f.deliverOne(f.user, grant)
			require.Zero(t, res.Code, res.Log)

			// a proposal stays in its deposit period
			params, err := f.govKeeper.Params.Get(f.ctx())
			require.NoError(t, err)
			id, err := f.govKeeper.ProposalID.Peek(f.ctx())
			require.NoError(t, err)
			half := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, params.MinDeposit[0].Amount.QuoRaw(2)))
			submit, err := govv1.NewMsgSubmitProposal([]sdk.Msg{types.NewMsgReset(f.gov, types.Wildcard)}, half, f.grantee.String(), "", "t", "s", false)
			require.NoError(t, err)
			res = f.deliverOne(f.grantee, submit)
			require.Zero(t, res.Code, res.Log)

			f.trip(types.Wildcard)
			amt := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, params.MinDeposit[0].Amount.QuoRaw(8)))
			deposit := func() *authz.MsgExec {
				e := authz.NewMsgExec(f.grantee.addr, []sdk.Msg{govv1.NewMsgDeposit(f.user.addr, id, amt)})
				return &e
			}
			before := f.balance(f.ctx(), f.user)
			f.requireWrapped(withAnte, f.grantee, "", deposit())
			require.Equal(t, before.Sub(amt.AmountOf(sdk.DefaultBondDenom)), f.balance(f.ctx(), f.user))

			revoke := authz.NewMsgRevoke(f.user.addr, f.grantee.addr, urlDeposit)
			f.requireWrapped(withAnte, f.user, urlAuthzRevoke, &revoke)

			f.trip(urlAuthzExec)
			before = f.balance(f.ctx(), f.user)
			f.requireWrapped(withAnte, f.grantee, urlAuthzExec, deposit())
			require.Equal(t, before, f.balance(f.ctx(), f.user))
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
	f := newRouterFixture(t, false)
	override := f.grantee

	cp, err := f.consensusKeeper.ParamsStore.Get(f.ctx())
	require.NoError(t, err)
	setOverride := func(authority, to string) *consensustypes.MsgUpdateParams {
		return &consensustypes.MsgUpdateParams{
			Authority: authority, Block: cp.Block, Evidence: cp.Evidence,
			Validator: cp.Validator, Abci: cp.Abci,
			Auth: &cmtproto.AuthorityParams{Authority: to},
		}
	}
	f.requireProposalStatus(f.submitProposal(setOverride(f.gov, override.String())), govv1.StatusVotingPeriod)
	f.endVoting()

	grant, err := authz.NewMsgGrant(override.addr, f.user.addr, authz.NewGenericAuthorization(sdk.MsgTypeURL(&types.MsgReset{})), nil)
	require.NoError(t, err)
	res := f.deliverOne(override, grant)
	require.Zero(t, res.Code, res.Log)

	params, err := f.safetyKeeper.Params.Get(f.ctx())
	require.NoError(t, err)
	params.Authority = f.gov
	res = f.deliverOne(override, &types.MsgUpdateParams{Authority: override.String(), Params: params})
	require.Zero(t, res.Code, res.Log)

	f.trip(urlAuthzExec)
	exec := authz.NewMsgExec(f.user.addr, []sdk.Msg{types.NewMsgReset(override.String(), urlAuthzExec)})
	requireRouterBlocked(t, f.deliverOne(f.user, &exec), urlAuthzExec)

	ids := []uint64{
		f.submitProposal(types.NewMsgReset(f.gov, urlAuthzExec)),
		f.submitProposal(setOverride(f.gov, "")),
		f.submitProposal(&upgradetypes.MsgSoftwareUpgrade{
			Authority: f.gov,
			Plan:      upgradetypes.Plan{Name: "recover", Height: f.app.LastBlockHeight() + 100},
		}),
	}
	f.endVoting()
	for _, id := range ids {
		p := f.requireProposalStatus(id, govv1.StatusFailed)
		require.Contains(t, p.FailedReason, "unauthorized")
	}
	require.Equal(t, []string{urlAuthzExec}, f.disabled())

	// after the operator's reset, the override resets through authz again
	f.reset(urlAuthzExec)
	f.trip(urlSend)
	exec = authz.NewMsgExec(f.user.addr, []sdk.Msg{types.NewMsgReset(override.String(), urlSend)})
	res = f.deliverOne(f.user, &exec)
	require.Zero(t, res.Code, res.Log)
	require.Empty(t, f.disabled())
}
