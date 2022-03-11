package module_test

import (
	"context"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/module"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestEndBlocker(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	addrs := simapp.AddTestAddrsIncremental(app, ctx, 4, types.NewInt(30000000))

	// Initial group, group policy and balance setup
	members := []group.Member{
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

	groupPolicyAddr, err := types.AccAddressFromBech32(policyRes.Address)
	require.NoError(t, err)

	votingPeriod := policy.GetVotingPeriod()
	now := time.Now()

	msgSend := &banktypes.MsgSend{
		FromAddress: groupPolicyAddr.String(),
		ToAddress:   addrs[3].String(),
		Amount:      types.Coins{types.NewInt64Coin("test", 100)},
	}

	proposers := []string{addrs[2].String()}

	specs := map[string]struct {
		preRun            func(sdkCtx types.Context) uint64
		proposalId        uint64
		admin             string
		expErrMsg         string
		newCtx            types.Context
		tallyRes          group.TallyResult
		expStatus         group.ProposalStatus
		expExecutorResult group.ProposalResult
	}{
		"tally updated after voting power end": {
			preRun: func(sdkCtx types.Context) uint64 {
				pId, err := submitProposal(app, sdkCtx, []types.Msg{msgSend}, proposers, groupPolicyAddr)
				require.NoError(t, err)
				return pId
			},
			admin:             proposers[0],
			newCtx:            ctx.WithBlockTime(now.Add(votingPeriod).Add(time.Hour)),
			tallyRes:          group.DefaultTallyResult(),
			expStatus:         group.PROPOSAL_STATUS_SUBMITTED,
			expExecutorResult: group.PROPOSAL_RESULT_UNFINALIZED,
		},
		"tally within voting period": {
			preRun: func(sdkCtx types.Context) uint64 {
				pId, err := submitProposal(app, sdkCtx, []types.Msg{msgSend}, proposers, groupPolicyAddr)
				require.NoError(t, err)

				return pId
			},
			admin:             proposers[0],
			newCtx:            ctx,
			tallyRes:          group.DefaultTallyResult(),
			expStatus:         group.PROPOSAL_STATUS_SUBMITTED,
			expExecutorResult: group.PROPOSAL_RESULT_UNFINALIZED,
		},
		"tally within voting period(with votes)": {
			preRun: func(sdkCtx types.Context) uint64 {
				pId, err := submitProposalAndVote(app, ctx, []types.Msg{msgSend}, proposers, groupPolicyAddr, group.VOTE_OPTION_YES)
				require.NoError(t, err)

				return pId
			},
			admin:             proposers[0],
			newCtx:            ctx,
			tallyRes:          group.DefaultTallyResult(),
			expStatus:         group.PROPOSAL_STATUS_SUBMITTED,
			expExecutorResult: group.PROPOSAL_RESULT_UNFINALIZED,
		},
		"tally after voting period(with votes)": {
			preRun: func(sdkCtx types.Context) uint64 {
				pId, err := submitProposalAndVote(app, ctx, []types.Msg{msgSend}, proposers, groupPolicyAddr, group.VOTE_OPTION_YES)
				require.NoError(t, err)

				return pId
			},
			admin:  proposers[0],
			newCtx: ctx.WithBlockTime(now.Add(votingPeriod).Add(time.Hour)),
			tallyRes: group.TallyResult{
				YesCount:        "2",
				NoCount:         "0",
				NoWithVetoCount: "0",
				AbstainCount:    "0",
			},
			expStatus:         group.PROPOSAL_STATUS_CLOSED,
			expExecutorResult: group.PROPOSAL_RESULT_ACCEPTED,
		},
		"tally of closed proposal": {
			preRun: func(sdkCtx types.Context) uint64 {
				pId, err := submitProposal(app, sdkCtx, []types.Msg{msgSend}, proposers, groupPolicyAddr)
				require.NoError(t, err)

				_, err = app.GroupKeeper.WithdrawProposal(ctx, &group.MsgWithdrawProposal{
					ProposalId: pId,
					Address:    groupPolicyAddr.String(),
				})

				require.NoError(t, err)
				return pId
			},
			admin:             proposers[0],
			newCtx:            ctx,
			tallyRes:          group.DefaultTallyResult(),
			expStatus:         group.PROPOSAL_STATUS_WITHDRAWN,
			expExecutorResult: group.PROPOSAL_RESULT_UNFINALIZED,
		},
		"tally of closed proposal (with votes)": {
			preRun: func(sdkCtx types.Context) uint64 {
				pId, err := submitProposalAndVote(app, ctx, []types.Msg{msgSend}, proposers, groupPolicyAddr, group.VOTE_OPTION_YES)
				require.NoError(t, err)

				_, err = app.GroupKeeper.WithdrawProposal(ctx, &group.MsgWithdrawProposal{
					ProposalId: pId,
					Address:    groupPolicyAddr.String(),
				})

				require.NoError(t, err)
				return pId
			},
			admin:             proposers[0],
			newCtx:            ctx,
			tallyRes:          group.DefaultTallyResult(),
			expStatus:         group.PROPOSAL_STATUS_WITHDRAWN,
			expExecutorResult: group.PROPOSAL_RESULT_UNFINALIZED,
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
			require.Equal(t, resp.GetProposal().Result, spec.expExecutorResult)
		})
	}
}

func submitProposal(
	app *simapp.SimApp, ctx context.Context, msgs []types.Msg,
	proposers []string, groupPolicyAddr types.AccAddress) (uint64, error) {
	proposalReq := &group.MsgSubmitProposal{
		Address:   groupPolicyAddr.String(),
		Proposers: proposers,
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
	app *simapp.SimApp, ctx context.Context, msgs []types.Msg,
	proposers []string, groupPolicyAddr types.AccAddress, voteOption group.VoteOption) (uint64, error) {
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
