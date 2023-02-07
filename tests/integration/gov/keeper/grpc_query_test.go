package keeper_test

import (
	gocontext "context"
	"fmt"
	"testing"

	"gotest.tools/v3/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func TestGRPCQueryTally(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	app, ctx, queryClient := f.app, f.ctx, f.queryClient

	addrs, _ := createValidators(t, ctx, app, []int64{5, 5, 5})

	var (
		req      *v1.QueryTallyResultRequest
		expRes   *v1.QueryTallyResultResponse
		proposal v1.Proposal
	)

	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		expErrMsg string
	}{
		{
			"empty request",
			func() {
				req = &v1.QueryTallyResultRequest{}
			},
			false,
			"proposal id can not be 0",
		},
		{
			"zero proposal id request",
			func() {
				req = &v1.QueryTallyResultRequest{ProposalId: 0}
			},
			false,
			"proposal id can not be 0",
		},
		{
			"query non existed proposal",
			func() {
				req = &v1.QueryTallyResultRequest{ProposalId: 1}
			},
			false,
			"proposal 1 doesn't exist",
		},
		{
			"create a proposal and get tally",
			func() {
				var err error
				proposal, err = app.GovKeeper.SubmitProposal(ctx, TestProposal, "", "test", "description", addrs[0], false)
				assert.NilError(t, err)
				assert.Assert(t, proposal.String() != "")

				req = &v1.QueryTallyResultRequest{ProposalId: proposal.Id}

				tallyResult := v1.EmptyTallyResult()
				expRes = &v1.QueryTallyResultResponse{
					Tally: &tallyResult,
				}
			},
			true,
			"",
		},
		{
			"request tally after few votes",
			func() {
				proposal.Status = v1.StatusVotingPeriod
				app.GovKeeper.SetProposal(ctx, proposal)

				assert.NilError(t, app.GovKeeper.AddVote(ctx, proposal.Id, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
				assert.NilError(t, app.GovKeeper.AddVote(ctx, proposal.Id, addrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
				assert.NilError(t, app.GovKeeper.AddVote(ctx, proposal.Id, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

				req = &v1.QueryTallyResultRequest{ProposalId: proposal.Id}

				expRes = &v1.QueryTallyResultResponse{
					Tally: &v1.TallyResult{
						YesCount:        sdk.NewInt(3 * 5 * 1000000).String(),
						NoCount:         "0",
						AbstainCount:    "0",
						NoWithVetoCount: "0",
					},
				}
			},
			true,
			"",
		},
		{
			"request final tally after status changed",
			func() {
				proposal.Status = v1.StatusPassed
				app.GovKeeper.SetProposal(ctx, proposal)
				proposal, _ = app.GovKeeper.GetProposal(ctx, proposal.Id)

				req = &v1.QueryTallyResultRequest{ProposalId: proposal.Id}

				expRes = &v1.QueryTallyResultResponse{
					Tally: proposal.FinalTallyResult,
				}
			},
			true,
			"",
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("Case %s", testCase.msg), func(t *testing.T) {
			testCase.malleate()

			tally, err := queryClient.TallyResult(gocontext.Background(), req)

			if testCase.expPass {
				assert.NilError(t, err)
				assert.Equal(t, expRes.String(), tally.String())
			} else {
				assert.ErrorContains(t, err, testCase.expErrMsg)
				assert.Assert(t, tally == nil)
			}
		})
	}
}

func TestLegacyGRPCQueryTally(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	app, ctx, queryClient := f.app, f.ctx, f.legacyQueryClient

	addrs, _ := createValidators(t, ctx, app, []int64{5, 5, 5})

	var (
		req      *v1beta1.QueryTallyResultRequest
		expRes   *v1beta1.QueryTallyResultResponse
		proposal v1.Proposal
	)

	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		expErrMsg string
	}{
		{
			"empty request",
			func() {
				req = &v1beta1.QueryTallyResultRequest{}
			},
			false,
			"proposal id can not be 0",
		},
		{
			"zero proposal id request",
			func() {
				req = &v1beta1.QueryTallyResultRequest{ProposalId: 0}
			},
			false,
			"proposal id can not be 0",
		},
		{
			"query non existed proposal",
			func() {
				req = &v1beta1.QueryTallyResultRequest{ProposalId: 1}
			},
			false,
			"proposal 1 doesn't exist",
		},
		{
			"create a proposal and get tally",
			func() {
				var err error
				proposal, err = app.GovKeeper.SubmitProposal(ctx, TestProposal, "", "test", "description", addrs[0], false)
				assert.NilError(t, err)
				assert.Assert(t, proposal.String() != "")

				req = &v1beta1.QueryTallyResultRequest{ProposalId: proposal.Id}

				tallyResult := v1beta1.EmptyTallyResult()
				expRes = &v1beta1.QueryTallyResultResponse{
					Tally: tallyResult,
				}
			},
			true,
			"",
		},
		{
			"request tally after few votes",
			func() {
				proposal.Status = v1.StatusVotingPeriod
				app.GovKeeper.SetProposal(ctx, proposal)

				assert.NilError(t, app.GovKeeper.AddVote(ctx, proposal.Id, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
				assert.NilError(t, app.GovKeeper.AddVote(ctx, proposal.Id, addrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
				assert.NilError(t, app.GovKeeper.AddVote(ctx, proposal.Id, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

				req = &v1beta1.QueryTallyResultRequest{ProposalId: proposal.Id}

				expRes = &v1beta1.QueryTallyResultResponse{
					Tally: v1beta1.TallyResult{
						Yes:        sdk.NewInt(3 * 5 * 1000000),
						No:         sdk.NewInt(0),
						Abstain:    sdk.NewInt(0),
						NoWithVeto: sdk.NewInt(0),
					},
				}
			},
			true,
			"",
		},
		{
			"request final tally after status changed",
			func() {
				proposal.Status = v1.StatusPassed
				app.GovKeeper.SetProposal(ctx, proposal)
				proposal, _ = app.GovKeeper.GetProposal(ctx, proposal.Id)

				req = &v1beta1.QueryTallyResultRequest{ProposalId: proposal.Id}

				expRes = &v1beta1.QueryTallyResultResponse{
					Tally: v1TallyToV1Beta1Tally(*proposal.FinalTallyResult),
				}
			},
			true,
			"",
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("Case %s", testCase.msg), func(t *testing.T) {
			testCase.malleate()

			tally, err := queryClient.TallyResult(gocontext.Background(), req)

			if testCase.expPass {
				assert.NilError(t, err)
				assert.Equal(t, expRes.String(), tally.String())
			} else {
				assert.ErrorContains(t, err, testCase.expErrMsg)
				assert.Assert(t, tally == nil)
			}
		})
	}
}

func v1TallyToV1Beta1Tally(t v1.TallyResult) v1beta1.TallyResult {
	yes, _ := sdk.NewIntFromString(t.YesCount)
	no, _ := sdk.NewIntFromString(t.NoCount)
	noWithVeto, _ := sdk.NewIntFromString(t.NoWithVetoCount)
	abstain, _ := sdk.NewIntFromString(t.AbstainCount)
	return v1beta1.TallyResult{
		Yes:        yes,
		No:         no,
		NoWithVeto: noWithVeto,
		Abstain:    abstain,
	}
}
