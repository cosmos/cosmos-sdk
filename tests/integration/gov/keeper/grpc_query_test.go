package keeper_test

import (
	gocontext "context"
	"fmt"
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/math"

	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func TestGRPCQueryTally(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx, queryClient := f.ctx, f.queryClient

	addrs, _ := createValidators(t, f, []int64{5, 5, 5})

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
				proposal, err = f.govKeeper.SubmitProposal(ctx, TestProposal, "", "test", "description", addrs[0], false)
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
				err := f.govKeeper.SetProposal(ctx, proposal)
				assert.NilError(t, err)
				assert.NilError(t, f.govKeeper.AddVote(ctx, proposal.Id, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
				assert.NilError(t, f.govKeeper.AddVote(ctx, proposal.Id, addrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
				assert.NilError(t, f.govKeeper.AddVote(ctx, proposal.Id, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

				req = &v1.QueryTallyResultRequest{ProposalId: proposal.Id}

				expRes = &v1.QueryTallyResultResponse{
					Tally: &v1.TallyResult{
						YesCount:        math.NewInt(3 * 5 * 1000000).String(),
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
				err := f.govKeeper.SetProposal(ctx, proposal)
				assert.NilError(t, err)
				proposal, _ = f.govKeeper.Proposals.Get(ctx, proposal.Id)

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

	ctx, queryClient := f.ctx, f.legacyQueryClient

	addrs, _ := createValidators(t, f, []int64{5, 5, 5})

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
				proposal, err = f.govKeeper.SubmitProposal(ctx, TestProposal, "", "test", "description", addrs[0], false)
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
				err := f.govKeeper.SetProposal(ctx, proposal)
				assert.NilError(t, err)
				assert.NilError(t, f.govKeeper.AddVote(ctx, proposal.Id, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
				assert.NilError(t, f.govKeeper.AddVote(ctx, proposal.Id, addrs[1], v1.NewNonSplitVoteOption(v1.OptionYes), ""))
				assert.NilError(t, f.govKeeper.AddVote(ctx, proposal.Id, addrs[2], v1.NewNonSplitVoteOption(v1.OptionYes), ""))

				req = &v1beta1.QueryTallyResultRequest{ProposalId: proposal.Id}

				expRes = &v1beta1.QueryTallyResultResponse{
					Tally: v1beta1.TallyResult{
						Yes:        math.NewInt(3 * 5 * 1000000),
						No:         math.NewInt(0),
						Abstain:    math.NewInt(0),
						NoWithVeto: math.NewInt(0),
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
				err := f.govKeeper.SetProposal(ctx, proposal)
				assert.NilError(t, err)
				proposal, _ = f.govKeeper.Proposals.Get(ctx, proposal.Id)

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
	yes, _ := math.NewIntFromString(t.YesCount)
	no, _ := math.NewIntFromString(t.NoCount)
	noWithVeto, _ := math.NewIntFromString(t.NoWithVetoCount)
	abstain, _ := math.NewIntFromString(t.AbstainCount)
	return v1beta1.TallyResult{
		Yes:        yes,
		No:         no,
		NoWithVeto: noWithVeto,
		Abstain:    abstain,
	}
}
