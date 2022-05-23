package module_test

import (
	"context"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/module"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestEndBlockerPruning(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	addrs := simapp.AddTestAddrsIncremental(app, ctx, 3, sdk.NewInt(30000000))
	addr1 := addrs[0]
	addr2 := addrs[1]
	addr3 := addrs[2]

	// Initial group, group policy and balance setup
	members := []group.MemberRequest{
		{Address: addr1.String(), Weight: "1"}, {Address: addr2.String(), Weight: "2"},
	}

	groupRes, err := app.GroupKeeper.CreateGroup(ctx, &group.MsgCreateGroup{
		Admin:   addr1.String(),
		Members: members,
	})
	require.NoError(t, err)

	groupRes2, err := app.GroupKeeper.CreateGroup(ctx, &group.MsgCreateGroup{
		Admin:   addr2.String(),
		Members: members,
	})
	require.NoError(t, err)

	groupID := groupRes.GroupId
	groupID2 := groupRes2.GroupId

	policy := group.NewThresholdDecisionPolicy(
		"2",
		time.Second,
		0,
	)

	policyReq := &group.MsgCreateGroupPolicy{
		Admin:   addr1.String(),
		GroupId: groupID,
	}

	err = policyReq.SetDecisionPolicy(policy)
	require.NoError(t, err)
	policyRes, err := app.GroupKeeper.CreateGroupPolicy(ctx, policyReq)
	require.NoError(t, err)

	policy2 := group.NewThresholdDecisionPolicy(
		"1",
		time.Second,
		0,
	)

	policyReq2 := &group.MsgCreateGroupPolicy{
		Admin:   addr2.String(),
		GroupId: groupID2,
	}

	err = policyReq2.SetDecisionPolicy(policy2)
	require.NoError(t, err)
	policyRes2, err := app.GroupKeeper.CreateGroupPolicy(ctx, policyReq2)
	require.NoError(t, err)

	groupPolicyAddr, err := sdk.AccAddressFromBech32(policyRes.Address)
	require.NoError(t, err)
	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, groupPolicyAddr, sdk.Coins{sdk.NewInt64Coin("test", 10000)}))

	groupPolicyAddr2, err := sdk.AccAddressFromBech32(policyRes2.Address)
	require.NoError(t, err)
	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, groupPolicyAddr2, sdk.Coins{sdk.NewInt64Coin("test", 10000)}))

	votingPeriod := policy.GetVotingPeriod()

	msgSend1 := &banktypes.MsgSend{
		FromAddress: groupPolicyAddr.String(),
		ToAddress:   addr2.String(),
		Amount:      sdk.Coins{sdk.NewInt64Coin("test", 100)},
	}
	msgSend2 := &banktypes.MsgSend{
		FromAddress: groupPolicyAddr2.String(),
		ToAddress:   addr2.String(),
		Amount:      sdk.Coins{sdk.NewInt64Coin("test", 100)},
	}
	proposers := []string{addr2.String()}

	specs := map[string]struct {
		setupProposal     func(ctx sdk.Context) uint64
		expErr            bool
		expErrMsg         string
		newCtx            sdk.Context
		expExecutorResult group.ProposalExecutorResult
		expStatus         group.ProposalStatus
	}{
		"proposal pruned after executor result success": {
			setupProposal: func(ctx sdk.Context) uint64 {
				msgs := []sdk.Msg{msgSend1}
				pID, err := submitProposalAndVote(app, ctx, msgs, proposers, groupPolicyAddr, group.VOTE_OPTION_YES)
				require.NoError(t, err)
				_, err = app.GroupKeeper.Exec(ctx, &group.MsgExec{Executor: addr3.String(), ProposalId: pID})
				require.NoError(t, err)
				sdkCtx := sdk.UnwrapSDKContext(ctx)
				require.NoError(t, testutil.FundAccount(app.BankKeeper, sdkCtx, groupPolicyAddr, sdk.Coins{sdk.NewInt64Coin("test", 10002)}))

				return pID
			},
			expErrMsg:         "load proposal: not found",
			newCtx:            ctx,
			expExecutorResult: group.PROPOSAL_EXECUTOR_RESULT_SUCCESS,
		},
		"proposal with multiple messages pruned when executed with result success": {
			setupProposal: func(ctx sdk.Context) uint64 {
				msgs := []sdk.Msg{msgSend1, msgSend1}
				pID, err := submitProposalAndVote(app, ctx, msgs, proposers, groupPolicyAddr, group.VOTE_OPTION_YES)
				require.NoError(t, err)
				_, err = app.GroupKeeper.Exec(ctx, &group.MsgExec{Executor: addr3.String(), ProposalId: pID})
				require.NoError(t, err)
				sdkCtx := sdk.UnwrapSDKContext(ctx)
				require.NoError(t, testutil.FundAccount(app.BankKeeper, sdkCtx, groupPolicyAddr, sdk.Coins{sdk.NewInt64Coin("test", 10002)}))

				return pID
			},
			expErrMsg:         "load proposal: not found",
			newCtx:            ctx,
			expExecutorResult: group.PROPOSAL_EXECUTOR_RESULT_SUCCESS,
		},
		"proposal not pruned when not executed and rejected": {
			setupProposal: func(ctx sdk.Context) uint64 {
				msgs := []sdk.Msg{msgSend1}
				pID, err := submitProposalAndVote(app, ctx, msgs, proposers, groupPolicyAddr, group.VOTE_OPTION_NO)
				require.NoError(t, err)
				_, err = app.GroupKeeper.Exec(ctx, &group.MsgExec{Executor: addr3.String(), ProposalId: pID})
				require.NoError(t, err)
				sdkCtx := sdk.UnwrapSDKContext(ctx)
				require.NoError(t, testutil.FundAccount(app.BankKeeper, sdkCtx, groupPolicyAddr, sdk.Coins{sdk.NewInt64Coin("test", 10002)}))

				return pID
			},
			newCtx:            ctx,
			expExecutorResult: group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN,
			expStatus:         group.PROPOSAL_STATUS_REJECTED,
		},
		"open proposal is not pruned which must not fail ": {
			setupProposal: func(ctx sdk.Context) uint64 {
				pID, err := submitProposal(app, ctx, []sdk.Msg{msgSend1}, proposers, groupPolicyAddr)
				require.NoError(t, err)
				_, err = app.GroupKeeper.Exec(ctx, &group.MsgExec{Executor: addr3.String(), ProposalId: pID})
				require.NoError(t, err)
				sdkCtx := sdk.UnwrapSDKContext(ctx)
				require.NoError(t, testutil.FundAccount(app.BankKeeper, sdkCtx, groupPolicyAddr, sdk.Coins{sdk.NewInt64Coin("test", 10002)}))

				return pID
			},
			newCtx:            ctx,
			expExecutorResult: group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN,
			expStatus:         group.PROPOSAL_STATUS_SUBMITTED,
		},
		"proposal not pruned with group policy modified before tally": {
			setupProposal: func(ctx sdk.Context) uint64 {
				pID, err := submitProposal(app, ctx, []sdk.Msg{msgSend1}, proposers, groupPolicyAddr)
				require.NoError(t, err)
				_, err = app.GroupKeeper.UpdateGroupPolicyMetadata(ctx, &group.MsgUpdateGroupPolicyMetadata{
					Admin:              addr1.String(),
					GroupPolicyAddress: groupPolicyAddr.String(),
				})
				require.NoError(t, err)
				_, err = app.GroupKeeper.Exec(ctx, &group.MsgExec{Executor: addr3.String(), ProposalId: pID})
				require.Error(t, err) // since proposal with status Aborted cannot be executed
				sdkCtx := sdk.UnwrapSDKContext(ctx)
				require.NoError(t, testutil.FundAccount(app.BankKeeper, sdkCtx, groupPolicyAddr, sdk.Coins{sdk.NewInt64Coin("test", 10002)}))

				return pID
			},
			newCtx:            ctx,
			expExecutorResult: group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN,
			expStatus:         group.PROPOSAL_STATUS_ABORTED,
		},
		"pruned when proposal is executable when failed before": {
			setupProposal: func(ctx sdk.Context) uint64 {
				msgs := []sdk.Msg{msgSend1}
				pID, err := submitProposalAndVote(app, ctx, msgs, proposers, groupPolicyAddr, group.VOTE_OPTION_YES)
				require.NoError(t, err)
				_, err = app.GroupKeeper.Exec(ctx, &group.MsgExec{Executor: addrs[2].String(), ProposalId: pID})
				require.NoError(t, err)
				return pID
			},
			newCtx:            ctx,
			expErrMsg:         "load proposal: not found",
			expExecutorResult: group.PROPOSAL_EXECUTOR_RESULT_SUCCESS,
		},
		"proposal with status withdrawn is pruned after voting period end": {
			setupProposal: func(sdkCtx sdk.Context) uint64 {
				pId, err := submitProposal(app, sdkCtx, []sdk.Msg{msgSend1}, proposers, groupPolicyAddr)
				require.NoError(t, err)
				_, err = app.GroupKeeper.WithdrawProposal(ctx, &group.MsgWithdrawProposal{
					ProposalId: pId,
					Address:    proposers[0],
				})
				require.NoError(t, err)
				return pId
			},
			newCtx:    ctx.WithBlockTime(ctx.BlockTime().Add(votingPeriod).Add(time.Hour)),
			expErrMsg: "load proposal: not found",
			expStatus: group.PROPOSAL_STATUS_WITHDRAWN,
		},
		"proposal with status withdrawn is not pruned (before voting period)": {
			setupProposal: func(sdkCtx sdk.Context) uint64 {
				pId, err := submitProposal(app, sdkCtx, []sdk.Msg{msgSend1}, proposers, groupPolicyAddr)
				require.NoError(t, err)
				_, err = app.GroupKeeper.WithdrawProposal(ctx, &group.MsgWithdrawProposal{
					ProposalId: pId,
					Address:    proposers[0],
				})
				require.NoError(t, err)
				return pId
			},
			newCtx:            ctx,
			expErrMsg:         "",
			expExecutorResult: group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN,
			expStatus:         group.PROPOSAL_STATUS_WITHDRAWN,
		},
		"proposal with status aborted is pruned after voting period end (due to updated group policy decision policy)": {
			setupProposal: func(sdkCtx sdk.Context) uint64 {
				pId, err := submitProposal(app, sdkCtx, []sdk.Msg{msgSend2}, proposers, groupPolicyAddr2)
				require.NoError(t, err)

				policy := group.NewThresholdDecisionPolicy("3", time.Second, 0)
				msg := &group.MsgUpdateGroupPolicyDecisionPolicy{
					Admin:              addrs[1].String(),
					GroupPolicyAddress: groupPolicyAddr2.String(),
				}
				err = msg.SetDecisionPolicy(policy)
				require.NoError(t, err)
				_, err = app.GroupKeeper.UpdateGroupPolicyDecisionPolicy(ctx, msg)
				require.NoError(t, err)

				return pId
			},
			newCtx:            ctx.WithBlockTime(ctx.BlockTime().Add(votingPeriod).Add(time.Hour)),
			expErrMsg:         "load proposal: not found",
			expStatus:         group.PROPOSAL_STATUS_ABORTED,
			expExecutorResult: group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN,
		},
		"proposal with status aborted is not pruned before voting period end (due to updated group policy)": {
			setupProposal: func(sdkCtx sdk.Context) uint64 {
				pId, err := submitProposal(app, sdkCtx, []sdk.Msg{msgSend2}, proposers, groupPolicyAddr2)
				require.NoError(t, err)

				policy := group.NewThresholdDecisionPolicy("3", time.Second, 0)
				msg := &group.MsgUpdateGroupPolicyDecisionPolicy{
					Admin:              addrs[1].String(),
					GroupPolicyAddress: groupPolicyAddr2.String(),
				}
				err = msg.SetDecisionPolicy(policy)
				require.NoError(t, err)
				_, err = app.GroupKeeper.UpdateGroupPolicyDecisionPolicy(ctx, msg)
				require.NoError(t, err)

				return pId
			},
			newCtx:            ctx,
			expErrMsg:         "",
			expStatus:         group.PROPOSAL_STATUS_ABORTED,
			expExecutorResult: group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN,
		},
	}
	for msg, spec := range specs {
		spec := spec
		t.Run(msg, func(t *testing.T) {
			proposalID := spec.setupProposal(ctx)

			module.EndBlocker(spec.newCtx, app.GroupKeeper)

			if spec.expErrMsg != "" && spec.expExecutorResult != group.PROPOSAL_EXECUTOR_RESULT_SUCCESS {
				_, err = app.GroupKeeper.Proposal(spec.newCtx, &group.QueryProposalRequest{ProposalId: proposalID})
				require.Error(t, err)
				require.Contains(t, err.Error(), spec.expErrMsg)
				return
			}
			if spec.expExecutorResult == group.PROPOSAL_EXECUTOR_RESULT_SUCCESS {
				// Make sure proposal is deleted from state
				_, err = app.GroupKeeper.Proposal(spec.newCtx, &group.QueryProposalRequest{ProposalId: proposalID})
				require.Contains(t, err.Error(), spec.expErrMsg)
				res, err := app.GroupKeeper.VotesByProposal(ctx, &group.QueryVotesByProposalRequest{ProposalId: proposalID})
				require.NoError(t, err)
				require.Empty(t, res.GetVotes())
			} else {
				// Check that proposal and votes exists
				res, err := app.GroupKeeper.Proposal(spec.newCtx, &group.QueryProposalRequest{ProposalId: proposalID})
				require.NoError(t, err)
				_, err = app.GroupKeeper.VotesByProposal(ctx, &group.QueryVotesByProposalRequest{ProposalId: res.Proposal.Id})
				require.NoError(t, err)
				require.Equal(t, "", spec.expErrMsg)

				exp := group.ProposalExecutorResult_name[int32(spec.expExecutorResult)]
				got := group.ProposalExecutorResult_name[int32(res.Proposal.ExecutorResult)]
				assert.Equal(t, exp, got)

				require.Equal(t, res.GetProposal().Status, spec.expStatus)
			}
		})
	}
}

func TestEndBlockerTallying(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addrs := simapp.AddTestAddrsIncremental(app, ctx, 4, sdk.NewInt(30000000))

	// Initial group, group policy and balance setup
	members := []group.MemberRequest{
		{Address: addrs[1].String(), Weight: "1"}, {Address: addrs[2].String(), Weight: "2"},
	}

	groupRes, err := app.GroupKeeper.CreateGroup(ctx, &group.MsgCreateGroup{
		Admin:   addrs[0].String(),
		Members: members,
	})

	require.NoError(t, err)
	groupID := groupRes.GroupId

	policy := group.NewThresholdDecisionPolicy(
		"2",
		time.Second,
		0,
	)

	policyReq := &group.MsgCreateGroupPolicy{
		Admin:   addrs[0].String(),
		GroupId: groupID,
	}

	err = policyReq.SetDecisionPolicy(policy)
	require.NoError(t, err)
	policyRes, err := app.GroupKeeper.CreateGroupPolicy(ctx, policyReq)
	require.NoError(t, err)

	groupPolicyAddr, err := sdk.AccAddressFromBech32(policyRes.Address)
	require.NoError(t, err)

	votingPeriod := policy.GetVotingPeriod()

	msgSend := &banktypes.MsgSend{
		FromAddress: groupPolicyAddr.String(),
		ToAddress:   addrs[3].String(),
		Amount:      sdk.Coins{sdk.NewInt64Coin("test", 100)},
	}

	proposers := []string{addrs[2].String()}

	specs := map[string]struct {
		preRun    func(sdkCtx sdk.Context) uint64
		admin     string
		expErrMsg string
		newCtx    sdk.Context
		tallyRes  group.TallyResult
		expStatus group.ProposalStatus
	}{
		"tally updated after voting period end": {
			preRun: func(sdkCtx sdk.Context) uint64 {
				pId, err := submitProposal(app, sdkCtx, []sdk.Msg{msgSend}, proposers, groupPolicyAddr)
				require.NoError(t, err)
				return pId
			},
			admin:     proposers[0],
			newCtx:    ctx.WithBlockTime(ctx.BlockTime().Add(votingPeriod).Add(time.Hour)),
			tallyRes:  group.DefaultTallyResult(),
			expStatus: group.PROPOSAL_STATUS_REJECTED,
		},
		"tally within voting period": {
			preRun: func(sdkCtx sdk.Context) uint64 {
				pId, err := submitProposal(app, sdkCtx, []sdk.Msg{msgSend}, proposers, groupPolicyAddr)
				require.NoError(t, err)

				return pId
			},
			admin:     proposers[0],
			newCtx:    ctx,
			tallyRes:  group.DefaultTallyResult(),
			expStatus: group.PROPOSAL_STATUS_SUBMITTED,
		},
		"tally within voting period(with votes)": {
			preRun: func(sdkCtx sdk.Context) uint64 {
				pId, err := submitProposalAndVote(app, ctx, []sdk.Msg{msgSend}, proposers, groupPolicyAddr, group.VOTE_OPTION_YES)
				require.NoError(t, err)

				return pId
			},
			admin:     proposers[0],
			newCtx:    ctx,
			tallyRes:  group.DefaultTallyResult(),
			expStatus: group.PROPOSAL_STATUS_SUBMITTED,
		},
		"tally after voting period (not passing)": {
			preRun: func(sdkCtx sdk.Context) uint64 {
				// `addrs[1]` has weight 1
				pId, err := submitProposalAndVote(app, ctx, []sdk.Msg{msgSend}, []string{addrs[1].String()}, groupPolicyAddr, group.VOTE_OPTION_YES)
				require.NoError(t, err)

				return pId
			},
			admin:  proposers[0],
			newCtx: ctx.WithBlockTime(ctx.BlockTime().Add(votingPeriod).Add(time.Hour)),
			tallyRes: group.TallyResult{
				YesCount:        "1",
				NoCount:         "0",
				NoWithVetoCount: "0",
				AbstainCount:    "0",
			},
			expStatus: group.PROPOSAL_STATUS_REJECTED,
		},
		"tally after voting period(with votes)": {
			preRun: func(sdkCtx sdk.Context) uint64 {
				pId, err := submitProposalAndVote(app, ctx, []sdk.Msg{msgSend}, proposers, groupPolicyAddr, group.VOTE_OPTION_YES)
				require.NoError(t, err)

				return pId
			},
			admin:  proposers[0],
			newCtx: ctx.WithBlockTime(ctx.BlockTime().Add(votingPeriod).Add(time.Hour)),
			tallyRes: group.TallyResult{
				YesCount:        "2",
				NoCount:         "0",
				NoWithVetoCount: "0",
				AbstainCount:    "0",
			},
			expStatus: group.PROPOSAL_STATUS_ACCEPTED,
		},
		"tally of withdrawn proposal": {
			preRun: func(sdkCtx sdk.Context) uint64 {
				pId, err := submitProposal(app, sdkCtx, []sdk.Msg{msgSend}, proposers, groupPolicyAddr)
				require.NoError(t, err)

				_, err = app.GroupKeeper.WithdrawProposal(ctx, &group.MsgWithdrawProposal{
					ProposalId: pId,
					Address:    proposers[0],
				})

				require.NoError(t, err)
				return pId
			},
			admin:     proposers[0],
			newCtx:    ctx,
			tallyRes:  group.DefaultTallyResult(),
			expStatus: group.PROPOSAL_STATUS_WITHDRAWN,
		},
		"tally of withdrawn proposal (with votes)": {
			preRun: func(sdkCtx sdk.Context) uint64 {
				pId, err := submitProposalAndVote(app, ctx, []sdk.Msg{msgSend}, proposers, groupPolicyAddr, group.VOTE_OPTION_YES)
				require.NoError(t, err)

				_, err = app.GroupKeeper.WithdrawProposal(ctx, &group.MsgWithdrawProposal{
					ProposalId: pId,
					Address:    proposers[0],
				})

				require.NoError(t, err)
				return pId
			},
			admin:     proposers[0],
			newCtx:    ctx,
			tallyRes:  group.DefaultTallyResult(),
			expStatus: group.PROPOSAL_STATUS_WITHDRAWN,
		},
	}

	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			spec := spec
			pId := spec.preRun(ctx)

			module.EndBlocker(spec.newCtx, app.GroupKeeper)
			resp, err := app.GroupKeeper.Proposal(spec.newCtx, &group.QueryProposalRequest{
				ProposalId: pId,
			})

			if spec.expErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), spec.expErrMsg)
				return
			}

			require.NoError(t, err)
			require.Equal(t, resp.GetProposal().FinalTallyResult, spec.tallyRes)
			require.Equal(t, resp.GetProposal().Status, spec.expStatus)
		})
	}
}

func submitProposal(app *simapp.SimApp, ctx context.Context, msgs []sdk.Msg, proposers []string, groupPolicyAddr sdk.AccAddress) (uint64, error) {
	proposalReq := &group.MsgSubmitProposal{
		GroupPolicyAddress: groupPolicyAddr.String(),
		Proposers:          proposers,
	}
	err := proposalReq.SetMsgs(msgs)
	if err != nil {
		return 0, err
	}

	proposalRes, err := app.GroupKeeper.SubmitProposal(ctx, proposalReq)
	if err != nil {
		return 0, err
	}

	return proposalRes.ProposalId, nil
}

func submitProposalAndVote(
	app *simapp.SimApp, ctx context.Context, msgs []sdk.Msg,
	proposers []string, groupPolicyAddr sdk.AccAddress, voteOption group.VoteOption,
) (uint64, error) {
	myProposalID, err := submitProposal(app, ctx, msgs, proposers, groupPolicyAddr)
	if err != nil {
		return 0, err
	}
	_, err = app.GroupKeeper.Vote(ctx, &group.MsgVote{
		ProposalId: myProposalID,
		Voter:      proposers[0],
		Option:     voteOption,
	})
	if err != nil {
		return 0, err
	}
	return myProposalID, nil
}
