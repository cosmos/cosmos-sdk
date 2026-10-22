package safety_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	safetytypes "github.com/cosmos/cosmos-sdk/x/safety/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TestDefaultGenesis checks the x/safety state SimApp starts with: gov is the
// authority, there is no operator, the operator scope is empty, and nothing
// is disabled.
func TestDefaultGenesis(t *testing.T) {
	f := newFixture(t, nil, nil)

	p := f.params()
	require.Equal(t, govAddr, p.Authority)
	require.Empty(t, p.Operator)
	require.Empty(t, f.disabled())
	require.True(t, f.isAllowed(msgSendURL))
	require.Empty(t, p.Scope, "default operator scope must be empty")
}

// TestOperatorLifecycle walks the main flow: gov appoints an operator with a
// bank scope, the operator trips MsgSend, MsgSend is rejected both as a
// top-level msg and nested in authz MsgExec, and the operator resets it.
func TestOperatorLifecycle(t *testing.T) {
	f := newFixture(t, nil, nil)

	// the user lets the operator spend on its behalf, for the authz paths
	grant, err := authz.NewMsgGrant(f.user.addr, f.operator.addr, authz.NewGenericAuthorization(msgSendURL), nil)
	require.NoError(t, err)
	f.mustDeliver(f.user, grant)

	res := f.mustPass(&safetytypes.MsgUpdateParams{
		Authority: govAddr,
		Params:    safetytypes.Params{Authority: govAddr, Operator: f.operator.String(), Scope: []string{"/cosmos.bank.*"}},
	})
	updated := typedEvents[*safetytypes.EventParamsUpdated](t, res.Events)
	require.Len(t, updated, 1)
	require.Equal(t, govAddr, updated[0].Authority)
	require.Equal(t, f.operator.String(), updated[0].Params.Operator)
	require.Equal(t, []string{"/cosmos.bank.*"}, f.params().Scope)

	// before the trip, sends work
	f.mustDeliver(f.user, send(f.user, f.other))

	tripRes := f.mustDeliver(f.operator, safetytypes.NewMsgTrip(f.operator.String(), msgSendURL))
	trips := typedEvents[*safetytypes.EventTrip](t, tripRes.Events)
	require.Len(t, trips, 1)
	require.Equal(t, f.operator.String(), trips[0].Operator)
	require.Equal(t, []string{msgSendURL}, trips[0].Patterns)
	require.Equal(t, []string{msgSendURL}, f.disabled())
	require.False(t, f.isAllowed(msgSendURL))
	require.True(t, f.isAllowed("/cosmos.bank.v1beta1.MsgMultiSend"))

	otherBefore := f.balance(f.other.addr)

	t.Run("top-level MsgSend is rejected", func(t *testing.T) {
		f.t = t
		f.requireRejected(f.user, safetytypes.ErrMsgDisabled, send(f.user, f.other))
	})
	t.Run("MsgSend nested in authz MsgExec is rejected", func(t *testing.T) {
		f.t = t
		exec := authz.NewMsgExec(f.operator.addr, []sdk.Msg{send(f.user, f.other)})
		f.requireRejected(f.operator, safetytypes.ErrMsgDisabled, &exec)
	})
	t.Run("MsgSend in a tx with an allowed msg is rejected", func(t *testing.T) {
		f.t = t
		f.requireRejected(f.user, safetytypes.ErrMsgDisabled,
			stakingtypes.NewMsgDelegate(f.user.String(), sdk.ValAddress(f.valSet.Validators[0].Address).String(), sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)),
			send(f.user, f.other))
	})
	f.t = t
	require.Equal(t, otherBefore, f.balance(f.other.addr), "no tripped send may succeed")

	// unrelated bank msgs keep working
	f.mustDeliver(f.user, multiSend(f.user, f.other))

	resetRes := f.mustDeliver(f.operator, safetytypes.NewMsgReset(f.operator.String(), msgSendURL))
	resets := typedEvents[*safetytypes.EventReset](t, resetRes.Events)
	require.Len(t, resets, 1)
	require.Equal(t, f.operator.String(), resets[0].Sender)
	require.Equal(t, []string{msgSendURL}, resets[0].Patterns)
	require.Empty(t, f.disabled())
	require.True(t, f.isAllowed(msgSendURL))

	before := f.balance(f.other.addr)
	f.mustDeliver(f.user, send(f.user, f.other))
	exec := authz.NewMsgExec(f.operator.addr, []sdk.Msg{send(f.user, f.other)})
	f.mustDeliver(f.operator, &exec)
	require.Equal(t, before.AddRaw(2), f.balance(f.other.addr))
}

// TestRouterGateWithinTx checks the authoritative router check: a msg tripped
// earlier in the same tx passes the ante handler but is blocked when the
// router dispatches it, both top level and via authz, and the whole tx,
// including the trip, is rolled back.
func TestRouterGateWithinTx(t *testing.T) {
	f := newFixture(t, nil, nil)
	grant, err := authz.NewMsgGrant(f.user.addr, f.operator.addr, authz.NewGenericAuthorization(msgSendURL), nil)
	require.NoError(t, err)
	f.mustDeliver(f.user, grant)
	f.setParams(f.operator.String(), "/cosmos.bank.*")

	before := f.balance(f.other.addr)
	trip := safetytypes.NewMsgTrip(f.operator.String(), msgSendURL)

	res := f.deliver(f.operator, trip, send(f.operator, f.other))
	require.NotZero(t, res.Code, res.Log)
	require.Contains(t, res.Log, msgSendURL)
	require.Equal(t, safetytypes.ErrMsgDisabled.Codespace(), res.Codespace, res.Log)
	require.Equal(t, safetytypes.ErrMsgDisabled.ABCICode(), res.Code, res.Log)

	exec := authz.NewMsgExec(f.operator.addr, []sdk.Msg{send(f.user, f.other)})
	res = f.deliver(f.operator, trip, &exec)
	require.NotZero(t, res.Code, res.Log)
	require.Contains(t, res.Log, msgSendURL)
	require.Equal(t, safetytypes.ErrMsgDisabled.Codespace(), res.Codespace, res.Log)
	require.Equal(t, safetytypes.ErrMsgDisabled.ABCICode(), res.Code, res.Log)

	require.Empty(t, f.disabled(), "the failed txs must not leave the trip behind")
	require.Equal(t, before, f.balance(f.other.addr))
}

// TestOperatorScope checks that the operator cannot trip or reset outside its
// scope, cannot trip protected or unregistered msgs, cannot change params,
// and that a failing pattern aborts the whole trip.
func TestOperatorScope(t *testing.T) {
	f := newFixture(t, nil, nil)
	f.setParams(f.operator.String(), "/cosmos.bank.*")

	f.requireFails(f.operator, safetytypes.ErrOutOfScope, safetytypes.NewMsgTrip(f.operator.String(), msgDelegateURL))
	f.requireFails(f.operator, safetytypes.ErrOutOfScope, safetytypes.NewMsgTrip(f.operator.String(), "*"))
	f.requireFails(f.operator, safetytypes.ErrOutOfScope, safetytypes.NewMsgTrip(f.operator.String(), "/cosmos.*"))
	// segment boundary: /cosmos.bank.* does not cover /cosmos.banking.*
	f.requireFails(f.operator, safetytypes.ErrOutOfScope, safetytypes.NewMsgTrip(f.operator.String(), "/cosmos.banking.v1.MsgSend"))
	// all or nothing: the in-scope MsgSend is not tripped either
	f.requireFails(f.operator, safetytypes.ErrOutOfScope, safetytypes.NewMsgTrip(f.operator.String(), msgSendURL, msgDelegateURL))
	require.Empty(t, f.disabled())

	// in scope but matching no registered msg
	f.requireFails(f.operator, safetytypes.ErrNoMatchingMsg, safetytypes.NewMsgTrip(f.operator.String(), "/cosmos.bank.v1beta1.MsgSnd"))
	// empty and malformed lists fail before reaching state
	f.requireFails(f.operator, safetytypes.ErrEmptyPatterns, safetytypes.NewMsgTrip(f.operator.String()))
	require.Empty(t, f.disabled())

	// the operator cannot change params
	p := f.params()
	p.Scope = []string{"*"}
	f.requireFails(f.operator, sdkerrors.ErrUnauthorized, &safetytypes.MsgUpdateParams{Authority: f.operator.String(), Params: p})
	require.Equal(t, []string{"/cosmos.bank.*"}, f.params().Scope)

	// a non-operator cannot trip or reset
	f.requireFails(f.user, sdkerrors.ErrUnauthorized, safetytypes.NewMsgTrip(f.user.String(), msgSendURL))
	f.mustDeliver(f.operator, safetytypes.NewMsgTrip(f.operator.String(), msgSendURL))
	f.requireFails(f.user, sdkerrors.ErrUnauthorized, safetytypes.NewMsgReset(f.user.String(), msgSendURL))
	// reset requires the exact pattern
	f.requireFails(f.operator, safetytypes.ErrNotDisabled, safetytypes.NewMsgReset(f.operator.String(), "/cosmos.bank.*"))
	// tripping twice fails
	f.requireFails(f.operator, safetytypes.ErrAlreadyDisabled, safetytypes.NewMsgTrip(f.operator.String(), msgSendURL))
	require.Equal(t, []string{msgSendURL}, f.disabled())

	// with a "*" scope the operator still cannot trip only-protected patterns
	f.setParams(f.operator.String(), "*")
	for _, p := range []string{"/cosmos.gov.*", "/cosmos.gov.v1.MsgSubmitProposal", "/cosmos.safety.v1.MsgReset", "/cosmos.upgrade.*", "/cosmos.consensus.v1.MsgUpdateParams"} {
		f.requireFails(f.operator, safetytypes.ErrNoMatchingMsg, safetytypes.NewMsgTrip(f.operator.String(), p))
	}
}

// TestNarrowScopeAndAuthorityReset checks that narrowing the operator scope
// leaves existing disabled entries in place, the operator cannot reset them
// once they fall outside its scope, and the authority, through gov, can.
func TestNarrowScopeAndAuthorityReset(t *testing.T) {
	f := newFixture(t, nil, nil)
	f.setParams(f.operator.String(), "/cosmos.bank.*", "/cosmos.staking.*")
	f.mustDeliver(f.operator, safetytypes.NewMsgTrip(f.operator.String(), msgSendURL, msgDelegateURL))
	require.ElementsMatch(t, []string{msgSendURL, msgDelegateURL}, f.disabled())

	f.setParams(f.operator.String(), "/cosmos.bank.*")
	require.ElementsMatch(t, []string{msgSendURL, msgDelegateURL}, f.disabled(), "narrowing the scope must not reset anything")
	require.False(t, f.isAllowed(msgDelegateURL))

	f.requireFails(f.operator, safetytypes.ErrOutOfScope, safetytypes.NewMsgReset(f.operator.String(), msgDelegateURL))
	// all or nothing: the in-scope MsgSend stays disabled too
	f.requireFails(f.operator, safetytypes.ErrOutOfScope, safetytypes.NewMsgReset(f.operator.String(), msgSendURL, msgDelegateURL))
	require.ElementsMatch(t, []string{msgSendURL, msgDelegateURL}, f.disabled())

	// removing the operator entirely also keeps the list
	f.setParams("")
	require.ElementsMatch(t, []string{msgSendURL, msgDelegateURL}, f.disabled())
	f.requireFails(f.operator, sdkerrors.ErrUnauthorized, safetytypes.NewMsgReset(f.operator.String(), msgSendURL))

	// the authority resets what the operator tripped, regardless of scope
	res := f.mustPass(safetytypes.NewMsgReset(govAddr, msgDelegateURL, msgSendURL))
	resets := typedEvents[*safetytypes.EventReset](t, res.Events)
	require.Len(t, resets, 1)
	require.Equal(t, govAddr, resets[0].Sender)
	require.Empty(t, f.disabled())
	f.mustDeliver(f.user, send(f.user, f.other))
}

// TestAuthorityCannotTrip checks that the authority, acting through gov,
// cannot trip, and that a plain account cannot update params.
func TestAuthorityCannotTrip(t *testing.T) {
	f := newFixture(t, nil, nil)
	f.setParams(f.operator.String(), "*")

	reason := f.mustFailExecution(safetytypes.NewMsgTrip(govAddr, msgSendURL))
	require.Contains(t, reason, "is not the operator")
	require.Empty(t, f.disabled())

	p := f.params()
	p.Operator = f.user.String()
	f.requireFails(f.user, sdkerrors.ErrUnauthorized, &safetytypes.MsgUpdateParams{Authority: f.user.String(), Params: p})
	require.Equal(t, f.operator.String(), f.params().Operator)
}

// TestAuthorityThatIsAlsoOperatorCanTrip checks the one exception: when the
// authority address is also the operator, it trips as the operator.
func TestAuthorityThatIsAlsoOperatorCanTrip(t *testing.T) {
	f := newFixture(t, nil, nil)
	f.setParams(govAddr, msgSendURL)

	f.mustPass(safetytypes.NewMsgTrip(govAddr, msgSendURL))
	require.Equal(t, []string{msgSendURL}, f.disabled())
}

// TestWildcardTripCannotLockOutReset checks that tripping "*" blocks ordinary
// msgs but leaves gov and x/safety executable, so the authority can always
// reset through a proposal.
func TestWildcardTripCannotLockOutReset(t *testing.T) {
	f := newFixture(t, nil, nil)
	f.setParams(f.operator.String(), "*")
	f.mustDeliver(f.operator, safetytypes.NewMsgTrip(f.operator.String(), "*"))

	f.requireRejected(f.user, safetytypes.ErrMsgDisabled, send(f.user, f.other))
	valAddr := sdk.ValAddress(f.valSet.Validators[0].Address).String()
	f.requireRejected(f.user, safetytypes.ErrMsgDisabled, stakingtypes.NewMsgDelegate(f.user.String(), valAddr, sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)))
	for _, url := range []string{"/cosmos.gov.v1.MsgSubmitProposal", "/cosmos.gov.v1.MsgVote", "/cosmos.safety.v1.MsgReset", "/cosmos.consensus.v1.MsgUpdateParams", "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade"} {
		require.True(t, f.isAllowed(url), url)
	}

	// propose submits and votes through the tripped chain
	f.mustPass(safetytypes.NewMsgReset(govAddr, "*"))
	require.Empty(t, f.disabled())
	f.mustDeliver(f.user, send(f.user, f.other))
}

// TestPrefixTrip checks that a prefix trip blocks every msg under it and
// nothing outside it.
func TestPrefixTrip(t *testing.T) {
	f := newFixture(t, nil, nil)
	f.setParams(f.operator.String(), "/cosmos.bank.*")
	f.mustDeliver(f.operator, safetytypes.NewMsgTrip(f.operator.String(), "/cosmos.bank.v1beta1.*"))

	f.requireRejected(f.user, safetytypes.ErrMsgDisabled, send(f.user, f.other))
	f.requireRejected(f.user, safetytypes.ErrMsgDisabled, multiSend(f.user, f.other))
	valAddr := sdk.ValAddress(f.valSet.Validators[0].Address).String()
	f.mustDeliver(f.user, stakingtypes.NewMsgDelegate(f.user.String(), valAddr, sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)))

	// a prefix entry resets only by its exact pattern
	f.requireFails(f.operator, safetytypes.ErrNotDisabled, safetytypes.NewMsgReset(f.operator.String(), msgSendURL))
	f.mustDeliver(f.operator, safetytypes.NewMsgReset(f.operator.String(), "/cosmos.bank.v1beta1.*"))
	f.mustDeliver(f.user, send(f.user, f.other))
}

// TestProposalWithTrippedMsgRejected checks that a gov proposal with a tripped
// msg is rejected at submission in CheckTx: the ante decorator opens any msg
// exposing GetMsgs, including the protected gov MsgSubmitProposal, and checks
// its contents. In FinalizeBlock the decorator stands aside, so a proposer can
// still include the submission; the router checks the tripped msg when the
// proposal executes.
func TestProposalWithTrippedMsgRejected(t *testing.T) {
	f := newFixture(t, nil, nil)
	f.setParams(f.operator.String(), "/cosmos.bank.*")
	f.mustDeliver(f.operator, safetytypes.NewMsgTrip(f.operator.String(), msgSendURL))

	govParams, err := f.app.GovKeeper.Params.Get(f.ctx())
	require.NoError(t, err)
	inner := sendFromGov(f.other)
	submit, err := govv1.NewMsgSubmitProposal([]sdk.Msg{inner}, govParams.MinDeposit, f.voter.String(), "", "send", "send", false)
	require.NoError(t, err)
	bz := f.requireCheckTxRejected(f.voter, safetytypes.ErrMsgDisabled, submit)
	res := f.finalize(bz).TxResults[0]
	require.Zero(t, res.Code, res.Log)
}
