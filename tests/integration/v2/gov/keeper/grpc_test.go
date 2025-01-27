package keeper

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/math"
	v1 "cosmossdk.io/x/gov/types/v1"
	"cosmossdk.io/x/gov/types/v1beta1"
)

func TestLegacyGRPCQueryTally(t *testing.T) {
	t.Parallel()

	f := initFixture(t)
	ctx, queryServer := f.ctx, f.legacyQueryServer
	addrs, _ := createValidators(t, f, []int64{5, 5, 5})

	var (
		req    *v1beta1.QueryTallyResultRequest
		expRes *v1beta1.QueryTallyResultResponse
	)

	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		expErrMsg string
	}{
		{
			"request tally after few votes",
			func() {
				proposal, err := f.govKeeper.SubmitProposal(ctx, TestProposal, "", "test", "description", addrs[0], v1.ProposalType_PROPOSAL_TYPE_STANDARD)
				assert.NilError(t, err)
				proposal.Status = v1.StatusVotingPeriod
				err = f.govKeeper.Proposals.Set(ctx, proposal.Id, proposal)
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
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("Case %s", testCase.msg), func(t *testing.T) {
			testCase.malleate()

			tally, err := queryServer.TallyResult(f.ctx, req)

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
