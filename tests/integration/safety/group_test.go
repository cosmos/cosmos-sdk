package safety_test

import (
	"context"
	"math/rand"
	"strings"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/enterprise/group/x/group"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	safetyante "github.com/cosmos/cosmos-sdk/x/safety/ante"
	safetytypes "github.com/cosmos/cosmos-sdk/x/safety/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// enterprise/group is its own Go module, so x/safety cannot import it. These
// tests pin the group type URLs x/safety hardcodes in its wrapper set, check
// that the ante decorator opens group proposals, and run the group recovery
// paths on a chain that wires the group module.

// The group proposal, vote and exec msgs are wrappers: only an exact disabled
// entry blocks them, they stay out of the protected set, and the other group
// msgs are neither.
func TestGroupWrapperTypeURLs(t *testing.T) {
	for _, msg := range []sdk.Msg{&group.MsgSubmitProposal{}, &group.MsgVote{}, &group.MsgExec{}} {
		url := sdk.MsgTypeURL(msg)
		require.True(t, safetytypes.IsWrapper(url), url)
		require.False(t, safetytypes.IsProtected(url), url)
		require.True(t, safetytypes.Covers("/cosmos.group.*", url), url)
		require.False(t, safetytypes.Blocks("/cosmos.group.*", url), url)
		require.False(t, safetytypes.Blocks(safetytypes.Wildcard, url), url)
		require.True(t, safetytypes.Blocks(url, url), url)
	}
	for _, msg := range []sdk.Msg{&group.MsgCreateGroup{}, &group.MsgUpdateGroupMetadata{}, &group.MsgWithdrawProposal{}} {
		url := sdk.MsgTypeURL(msg)
		require.False(t, safetytypes.IsWrapper(url), url)
		require.True(t, safetytypes.Blocks("/cosmos.group.*", url), url)
	}

	// the decorator rejects an exactly disabled group proposal before
	// opening it
	breaker := disabledSet{sdk.MsgTypeURL(&group.MsgSubmitProposal{}): true}
	policy := sdk.AccAddress("group_policy________").String()
	m, err := group.NewMsgSubmitProposal(policy, []string{policy}, nil, "", group.Exec_EXEC_TRY, "t", "s")
	require.NoError(t, err)
	next := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) { return ctx, nil }
	_, err = safetyante.NewSafetyDecorator(breaker).AnteHandle(sdk.Context{}, groupTx{msgs: []sdk.Msg{m}}, false, next)
	require.ErrorIs(t, err, safetytypes.ErrMsgDisabled)
}

type disabledSet map[string]bool

func (d disabledSet) IsAllowed(_ context.Context, typeURL string) (bool, error) {
	return !d[typeURL], nil
}

type groupTx struct{ msgs []sdk.Msg }

func (t groupTx) GetMsgs() []sdk.Msg                    { return t.msgs }
func (t groupTx) GetMsgsV2() ([]protov2.Message, error) { return nil, nil }
func (t groupTx) ValidateBasic() error                  { return nil }
func (t groupTx) Hash() [32]byte                        { return [32]byte{} }
func (t groupTx) GetSigners() ([][]byte, error)         { return nil, nil }

// group MsgSubmitProposal exposes GetMsgs on a value receiver, unlike gov's
// pointer receiver; the decorator must open it either way.
func TestAnteOpensGroupProposal(t *testing.T) {
	send := &banktypes.MsgSend{}
	deleg := &stakingtypes.MsgDelegate{}
	policy := sdk.AccAddress("group_policy________").String()
	breaker := disabledSet{sdk.MsgTypeURL(send): true}

	run := func(msgs ...sdk.Msg) error {
		t.Helper()
		m, err := group.NewMsgSubmitProposal(policy, []string{policy}, msgs, "", group.Exec_EXEC_TRY, "t", "s")
		require.NoError(t, err)
		next := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) { return ctx, nil }
		_, err = safetyante.NewSafetyDecorator(breaker).AnteHandle(sdk.Context{}, groupTx{msgs: []sdk.Msg{m}}, false, next)
		return err
	}

	require.NoError(t, run(deleg))
	require.ErrorIs(t, run(send), safetytypes.ErrMsgDisabled)
	require.ErrorIs(t, run(deleg, send), safetytypes.ErrMsgDisabled)
}

// TestGroupPolicyAuthorityUnderBroadTrips runs a group policy as the x/safety
// authority. A "/cosmos.group.*" or "*" trip leaves the group proposal, vote
// and exec msgs running, while the router and the ante decorator still refuse
// the unprotected msgs in a proposal, so the policy resets "*" through a
// proposal.
func TestGroupPolicyAuthorityUnderBroadTrips(t *testing.T) {
	f := newGroupFixture(t)
	require.Equal(t, f.policy.String(), f.params().Authority)
	urlMetadata := sdk.MsgTypeURL(&group.MsgUpdateGroupMetadata{})

	// two proposals pass their vote before any trip: one sets group
	// metadata, one sends from the policy
	metadata := &group.MsgUpdateGroupMetadata{Admin: f.policy.String(), GroupId: f.groupID, Metadata: "tripped"}
	sendFromPolicy := banktypes.NewMsgSend(f.policy, f.member.addr, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)))
	metaID := f.submit(metadata)
	f.requireTx(f.member, "", f.vote(metaID))
	sendID := f.submit(sendFromPolicy)
	f.requireTx(f.member, "", f.vote(sendID))

	// the group's own package: its wrappers run, its other msgs do not
	f.trip("/cosmos.group.*")
	f.requireTx(f.member, "", f.exec(metaID))
	require.Equal(t, group.PROPOSAL_EXECUTOR_RESULT_FAILURE, f.executorResult(metaID))
	require.Empty(t, f.groupMetadata())
	f.requireCheckTxRejected(f.member, urlMetadata, f.proposal(metadata))
	f.reset("/cosmos.group.*")

	// "*": the same, and the policy still reaches MsgReset
	f.trip(safetytypes.Wildcard)
	balance := f.bankKeeper.GetBalance(f.ctx(), f.policy, sdk.DefaultBondDenom)
	f.requireTx(f.member, "", f.exec(sendID))
	require.Equal(t, group.PROPOSAL_EXECUTOR_RESULT_FAILURE, f.executorResult(sendID))
	require.Equal(t, balance, f.bankKeeper.GetBalance(f.ctx(), f.policy, sdk.DefaultBondDenom))
	f.requireCheckTxRejected(f.member, msgSendURL, f.proposal(sendFromPolicy))

	resetID := f.submit(safetytypes.NewMsgReset(f.policy.String(), safetytypes.Wildcard))
	f.requireTx(f.member, "", f.vote(resetID))
	f.requireTx(f.member, "", f.exec(resetID))
	require.Empty(t, f.disabled())
}

// TestGroupWrappersBlockedByExactTrip trips each group wrapper by its exact
// URL, under "*", and shows CheckTx and the router refuse it.
func TestGroupWrappersBlockedByExactTrip(t *testing.T) {
	f := newGroupFixture(t)
	reset := safetytypes.NewMsgReset(f.policy.String(), safetytypes.Wildcard)
	urlSubmit := sdk.MsgTypeURL(&group.MsgSubmitProposal{})
	urlVote := sdk.MsgTypeURL(&group.MsgVote{})
	urlExec := sdk.MsgTypeURL(&group.MsgExec{})

	f.trip(safetytypes.Wildcard, urlSubmit)
	f.requireTx(f.member, urlSubmit, f.proposal(reset))
	f.reset(urlSubmit)

	id := f.submit(reset)
	f.trip(urlVote)
	f.requireTx(f.member, urlVote, f.vote(id))
	f.reset(urlVote)

	f.requireTx(f.member, "", f.vote(id))
	f.trip(urlExec)
	f.requireTx(f.member, urlExec, f.exec(id))
	require.ElementsMatch(t, []string{safetytypes.Wildcard, urlExec}, f.disabled())
	f.reset(urlExec)

	f.requireTx(f.member, "", f.exec(id))
	require.Empty(t, f.disabled())
}

// An exact trip of group MsgSubmitProposal alone locks out a group policy
// authority with no proposal pending: it cannot submit a reset, plain or with
// Exec=TRY, and the open MsgVote and MsgExec have no proposal to act on. Once
// the operator resets the trip, the policy submits proposals again.
func TestGroupPolicyLockedOutByExactSubmitTrip(t *testing.T) {
	f := newGroupFixture(t)
	urlSubmit := sdk.MsgTypeURL(&group.MsgSubmitProposal{})
	reset := safetytypes.NewMsgReset(f.policy.String(), urlSubmit)
	f.trip(urlSubmit)

	f.requireTx(f.member, urlSubmit, f.proposal(reset))
	try := f.proposal(reset)
	try.Exec = group.Exec_EXEC_TRY
	f.requireTx(f.member, urlSubmit, try)

	res, err := f.groupKeeper.ProposalsByGroupPolicy(f.ctx(), &group.QueryProposalsByGroupPolicyRequest{Address: f.policy.String()})
	require.NoError(t, err)
	require.Empty(t, res.Proposals)
	require.Equal(t, []string{urlSubmit}, f.disabled())

	f.reset(urlSubmit)
	require.Empty(t, f.disabled())
	f.submit(&group.MsgUpdateGroupMetadata{Admin: f.policy.String(), GroupId: f.groupID, Metadata: "after reset"})
}

// Group MsgSubmitProposal with Exec=TRY votes and executes through the group
// keeper, so exact trips of group MsgVote and MsgExec do not stop a proposal
// from executing. Tripping MsgSubmitProposal as well does.
func TestGroupTryExecRunsPastExactVoteAndExecTrips(t *testing.T) {
	f := newGroupFixture(t)
	urlSubmit := sdk.MsgTypeURL(&group.MsgSubmitProposal{})
	urlVote := sdk.MsgTypeURL(&group.MsgVote{})
	urlExec := sdk.MsgTypeURL(&group.MsgExec{})
	tryProposal := func(metadata string) *group.MsgSubmitProposal {
		m := f.proposal(&group.MsgUpdateGroupMetadata{Admin: f.policy.String(), GroupId: f.groupID, Metadata: metadata})
		m.Exec = group.Exec_EXEC_TRY
		return m
	}

	f.trip(urlVote, urlExec)
	f.requireTx(f.member, "", tryProposal("executed through try"))
	f.proposals++
	require.Equal(t, "executed through try", f.groupMetadata())

	f.trip(urlSubmit)
	f.requireTx(f.member, urlSubmit, tryProposal("blocked"))
	require.Equal(t, "executed through try", f.groupMetadata())
}

// Group MsgVote with Exec=TRY executes the proposal through the group keeper,
// so under "*" and an exact trip of group MsgExec the group policy authority
// still resets "*".
func TestGroupAuthorityResetsThroughVoteTryUnderExactExecTrip(t *testing.T) {
	f := newGroupFixture(t)
	urlExec := sdk.MsgTypeURL(&group.MsgExec{})
	id := f.submit(safetytypes.NewMsgReset(f.policy.String(), safetytypes.Wildcard))
	f.trip(safetytypes.Wildcard, urlExec)

	vote := f.vote(id)
	vote.Exec = group.Exec_EXEC_TRY
	f.requireTx(f.member, "", vote)
	require.Equal(t, []string{urlExec}, f.disabled())
}

// An exact trip of group MsgVote alone locks out a group policy authority
// whose threshold needs a member who is not a proposer. A proposal with
// Exec=TRY counts only its proposers' votes, the other member cannot vote, and
// the open MsgExec has no accepted proposal to execute.
func TestGroupPolicyLockedOutByExactVoteTrip(t *testing.T) {
	f := newGroupFixture(t)
	urlVote := sdk.MsgTypeURL(&group.MsgVote{})

	// a second group: member and operator, weight 1 each, threshold 2
	create, err := group.NewMsgCreateGroupWithPolicy(f.member.String(),
		[]group.MemberRequest{{Address: f.member.String(), Weight: "1"}, {Address: f.operator.String(), Weight: "1"}},
		"", "", true, group.NewThresholdDecisionPolicy("2", time.Hour, 0))
	require.NoError(t, err)
	f.requireTx(f.member, "", create)
	res, err := f.groupKeeper.GroupPoliciesByGroup(f.ctx(), &group.QueryGroupPoliciesByGroupRequest{GroupId: f.groupID + 1})
	require.NoError(t, err)
	require.Len(t, res.GroupPolicies, 1)
	policy2 := res.GroupPolicies[0].Address

	id := f.submit(&safetytypes.MsgUpdateParams{
		Authority: f.policy.String(),
		Params:    safetytypes.Params{Authority: policy2, Operator: f.operator.String(), Scope: []string{safetytypes.Wildcard}},
	})
	f.requireTx(f.member, "", f.vote(id))
	f.requireTx(f.member, "", f.exec(id))
	require.Equal(t, policy2, f.params().Authority)

	f.trip(urlVote)
	try, err := group.NewMsgSubmitProposal(policy2, []string{f.member.String()},
		[]sdk.Msg{safetytypes.NewMsgReset(policy2, urlVote)}, "", group.Exec_EXEC_TRY, "", "")
	require.NoError(t, err)
	f.requireTx(f.member, "", try)
	props, err := f.groupKeeper.ProposalsByGroupPolicy(f.ctx(), &group.QueryProposalsByGroupPolicyRequest{Address: policy2})
	require.NoError(t, err)
	require.Len(t, props.Proposals, 1)
	pid := props.Proposals[0].Id
	require.Equal(t, group.PROPOSAL_STATUS_SUBMITTED, props.Proposals[0].Status)

	f.requireTx(f.operator, urlVote, &group.MsgVote{ProposalId: pid, Voter: f.operator.String(), Option: group.VOTE_OPTION_YES})
	f.requireTx(f.member, "", &group.MsgExec{ProposalId: pid, Executor: f.member.String()})
	require.Equal(t, []string{urlVote}, f.disabled())
}

// Under "*", authz MsgExec dispatches group MsgVote, which is a wrapper and
// not protected, and the vote changes state: the proposal reaches its
// threshold. The router still refuses the unprotected msg the proposal
// executes.
func TestWrapperDispatchesWrapperUnderWildcard(t *testing.T) {
	f := newGroupFixture(t)
	id := f.submit(&group.MsgUpdateGroupMetadata{Admin: f.policy.String(), GroupId: f.groupID, Metadata: "via authz"})
	grant, err := authz.NewMsgGrant(f.member.addr, f.operator.addr, authz.NewGenericAuthorization(sdk.MsgTypeURL(&group.MsgVote{})), nil)
	require.NoError(t, err)
	f.requireTx(f.member, "", grant)

	f.trip(safetytypes.Wildcard)
	vote := f.vote(id)
	vote.Exec = group.Exec_EXEC_TRY
	exec := authz.NewMsgExec(f.operator.addr, []sdk.Msg{vote})
	f.requireTx(f.operator, "", &exec)

	res, err := f.groupKeeper.Proposal(f.ctx(), &group.QueryProposalRequest{ProposalId: id})
	require.NoError(t, err)
	require.Equal(t, group.PROPOSAL_STATUS_ACCEPTED, res.Proposal.Status)
	require.Equal(t, "1", res.Proposal.FinalTallyResult.YesCount)
	require.Equal(t, group.PROPOSAL_EXECUTOR_RESULT_FAILURE, f.executorResult(id))
	require.Empty(t, f.groupMetadata())
}

// The group fixture runs the SimApp ante chain, so a tx past its timeout
// height fails CheckTx on the timeout check rather than on its signature.
func TestGroupFixtureRunsSimAppAnteChain(t *testing.T) {
	f := newGroupFixture(t)
	acc := f.accountKeeper.GetAccount(f.ctx(), f.member.addr)
	tx, err := simtestutil.GenSignedMockTx(rand.New(rand.NewSource(1)), f.txConfig,
		[]sdk.Msg{banktypes.NewMsgSend(f.member.addr, f.policy, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)))},
		nil, simtestutil.DefaultGenTxGas, "", []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, f.member.priv)
	require.NoError(t, err)
	builder, err := f.txConfig.WrapTxBuilder(tx)
	require.NoError(t, err)
	builder.SetTimeoutHeight(1)
	bz, err := f.txConfig.TxEncoder()(builder.GetTx())
	require.NoError(t, err)

	res, err := f.app.CheckTx(&abci.RequestCheckTx{Tx: bz, Type: abci.CheckTxType_New})
	require.NoError(t, err)
	require.Equal(t, sdkerrors.ErrTxTimeoutHeight.ABCICode(), res.Code, res.Log)
	require.Equal(t, sdkerrors.ErrTxTimeoutHeight.Codespace(), res.Codespace, res.Log)
}

// Exactly group MsgSubmitProposal, MsgVote and MsgExec are wrappers. Under
// "*", IsAllowed blocks every other registered group msg and lets those
// three through.
func TestGroupWrapperSetIsExact(t *testing.T) {
	wrappers := map[string]bool{
		"/cosmos.group.v1.MsgSubmitProposal": true,
		"/cosmos.group.v1.MsgVote":           true,
		"/cosmos.group.v1.MsgExec":           true,
	}
	registry := codectypes.NewInterfaceRegistry()
	sdk.RegisterInterfaces(registry)
	group.RegisterInterfaces(registry)
	f := newGroupFixture(t)
	f.trip(safetytypes.Wildcard)

	var seen, others int
	for _, url := range registry.ListImplementations(sdk.MsgInterfaceProtoName) {
		if !strings.HasPrefix(url, "/cosmos.group.") {
			continue
		}
		require.Equal(t, wrappers[url], safetytypes.IsWrapper(url), url)
		require.Equal(t, !wrappers[url], safetytypes.Blocks("/cosmos.group.*", url), url)
		allowed, err := f.safetyKeeper.IsAllowed(f.ctx(), url)
		if wrappers[url] {
			seen++
			require.NoError(t, err, url)
			require.True(t, allowed, url)
		} else {
			others++
			require.ErrorIs(t, err, safetytypes.ErrMsgDisabled, url)
			require.False(t, allowed, url)
		}
	}
	require.Equal(t, len(wrappers), seen)
	require.Positive(t, others)
}
