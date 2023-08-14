package keeper_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// TODO(tip): remove this
func (suite *KeeperTestSuite) TestGetSetProposal() {
	testCases := map[string]struct {
		expedited bool
	}{
		"regular proposal": {},
		"expedited proposal": {
			expedited: true,
		},
	}

	for _, tc := range testCases {
		tp := TestProposal
		proposal, err := suite.govKeeper.SubmitProposal(suite.ctx, tp, "", "test", "summary", suite.addrs[0], tc.expedited)
		suite.Require().NoError(err)
		proposalID := proposal.Id
		err = suite.govKeeper.SetProposal(suite.ctx, proposal)
		suite.Require().NoError(err)

		gotProposal, err := suite.govKeeper.Proposals.Get(suite.ctx, proposalID)
		suite.Require().Nil(err)
		suite.Require().Equal(proposal, gotProposal)
	}
}

// TODO(tip): remove this
func (suite *KeeperTestSuite) TestDeleteProposal() {
	testCases := map[string]struct {
		expedited bool
	}{
		"regular proposal": {},
		"expedited proposal": {
			expedited: true,
		},
	}

	for _, tc := range testCases {
		// delete non-existing proposal
		suite.Require().ErrorIs(suite.govKeeper.DeleteProposal(suite.ctx, 10), collections.ErrNotFound)

		tp := TestProposal
		proposal, err := suite.govKeeper.SubmitProposal(suite.ctx, tp, "", "test", "summary", suite.addrs[0], tc.expedited)
		suite.Require().NoError(err)
		proposalID := proposal.Id
		err = suite.govKeeper.SetProposal(suite.ctx, proposal)
		suite.Require().NoError(err)

		suite.Require().NotPanics(func() {
			err := suite.govKeeper.DeleteProposal(suite.ctx, proposalID)
			suite.Require().NoError(err)
		}, "")
	}
}

func (suite *KeeperTestSuite) TestActivateVotingPeriod() {
	testCases := []struct {
		name      string
		expedited bool
	}{
		{name: "regular proposal"},
		{name: "expedited proposal", expedited: true},
	}

	for _, tc := range testCases {
		tp := TestProposal
		proposal, err := suite.govKeeper.SubmitProposal(suite.ctx, tp, "", "test", "summary", suite.addrs[0], tc.expedited)
		suite.Require().NoError(err)

		suite.Require().Nil(proposal.VotingStartTime)

		err = suite.govKeeper.ActivateVotingPeriod(suite.ctx, proposal)
		suite.Require().NoError(err)

		proposal, err = suite.govKeeper.Proposals.Get(suite.ctx, proposal.Id)
		suite.Require().Nil(err)
		suite.Require().True(proposal.VotingStartTime.Equal(suite.ctx.BlockHeader().Time))

		has, err := suite.govKeeper.ActiveProposalsQueue.Has(suite.ctx, collections.Join(*proposal.VotingEndTime, proposal.Id))
		suite.Require().NoError(err)
		suite.Require().True(has)
		suite.Require().NoError(suite.govKeeper.DeleteProposal(suite.ctx, proposal.Id))
	}
}

func (suite *KeeperTestSuite) TestDeleteProposalInVotingPeriod() {
	testCases := []struct {
		name      string
		expedited bool
	}{
		{name: "regular proposal"},
		{name: "expedited proposal", expedited: true},
	}

	for _, tc := range testCases {
		suite.reset()
		tp := TestProposal
		proposal, err := suite.govKeeper.SubmitProposal(suite.ctx, tp, "", "test", "summary", suite.addrs[0], tc.expedited)
		suite.Require().NoError(err)
		suite.Require().Nil(proposal.VotingStartTime)

		suite.Require().NoError(suite.govKeeper.ActivateVotingPeriod(suite.ctx, proposal))

		proposal, err = suite.govKeeper.Proposals.Get(suite.ctx, proposal.Id)
		suite.Require().Nil(err)
		suite.Require().True(proposal.VotingStartTime.Equal(suite.ctx.BlockHeader().Time))

		has, err := suite.govKeeper.ActiveProposalsQueue.Has(suite.ctx, collections.Join(*proposal.VotingEndTime, proposal.Id))
		suite.Require().NoError(err)
		suite.Require().True(has)

		// add vote
		voteOptions := []*v1.WeightedVoteOption{{Option: v1.OptionYes, Weight: "1.0"}}
		err = suite.govKeeper.AddVote(suite.ctx, proposal.Id, suite.addrs[0], voteOptions, "")
		suite.Require().NoError(err)

		suite.Require().NoError(suite.govKeeper.DeleteProposal(suite.ctx, proposal.Id))

		// add vote but proposal is deleted along with its VotingPeriodProposalKey
		err = suite.govKeeper.AddVote(suite.ctx, proposal.Id, suite.addrs[0], voteOptions, "")
		suite.Require().ErrorContains(err, ": inactive proposal")
	}
}

type invalidProposalRoute struct{ v1beta1.TextProposal }

func (invalidProposalRoute) ProposalRoute() string { return "nonexistingroute" }

func (suite *KeeperTestSuite) TestSubmitProposal() {
	govAcct := suite.govKeeper.GetGovernanceAccount(suite.ctx).GetAddress().String()
	_, _, randomAddr := testdata.KeyTestPubAddr()
	tp := v1beta1.TextProposal{Title: "title", Description: "description"}

	testCases := []struct {
		content     v1beta1.Content
		authority   string
		metadata    string
		expedited   bool
		expectedErr error
	}{
		{&tp, govAcct, "", false, nil},
		{&tp, govAcct, "", true, nil},
		// Keeper does not check the validity of title and description, no error
		{&v1beta1.TextProposal{Title: "", Description: "description"}, govAcct, "", false, nil},
		{&v1beta1.TextProposal{Title: strings.Repeat("1234567890", 100), Description: "description"}, govAcct, "", false, nil},
		{&v1beta1.TextProposal{Title: "title", Description: ""}, govAcct, "", false, nil},
		{&v1beta1.TextProposal{Title: "title", Description: strings.Repeat("1234567890", 1000)}, govAcct, "", true, nil},
		// error when metadata is too long (>10000)
		{&tp, govAcct, strings.Repeat("a", 100001), true, types.ErrMetadataTooLong},
		// error when signer is not gov acct
		{&tp, randomAddr.String(), "", false, types.ErrInvalidSigner},
		// error only when invalid route
		{&invalidProposalRoute{}, govAcct, "", false, types.ErrNoProposalHandlerExists},
	}

	for i, tc := range testCases {
		prop, err := v1.NewLegacyContent(tc.content, tc.authority)
		suite.Require().NoError(err)
		_, err = suite.govKeeper.SubmitProposal(suite.ctx, []sdk.Msg{prop}, tc.metadata, "title", "", suite.addrs[0], tc.expedited)
		suite.Require().True(errors.Is(tc.expectedErr, err), "tc #%d; got: %v, expected: %v", i, err, tc.expectedErr)
	}
}

func (suite *KeeperTestSuite) TestCancelProposal() {
	govAcct := suite.govKeeper.GetGovernanceAccount(suite.ctx).GetAddress().String()
	tp := v1beta1.TextProposal{Title: "title", Description: "description"}
	prop, err := v1.NewLegacyContent(&tp, govAcct)
	suite.Require().NoError(err)
	proposal, err := suite.govKeeper.SubmitProposal(suite.ctx, []sdk.Msg{prop}, "", "title", "summary", suite.addrs[0], false)
	suite.Require().NoError(err)
	proposalID := proposal.Id

	proposal2, err := suite.govKeeper.SubmitProposal(suite.ctx, []sdk.Msg{prop}, "", "title", "summary", suite.addrs[1], true)
	suite.Require().NoError(err)
	proposal2ID := proposal2.Id

	// proposal3 is only used to check the votes for proposals which doesn't go through `CancelProposal` are still present in state
	proposal3, err := suite.govKeeper.SubmitProposal(suite.ctx, []sdk.Msg{prop}, "", "title", "summary", suite.addrs[2], false)
	suite.Require().NoError(err)
	proposal3ID := proposal3.Id

	// add votes for proposal 3
	suite.Require().NoError(suite.govKeeper.ActivateVotingPeriod(suite.ctx, proposal3))

	proposal3, err = suite.govKeeper.Proposals.Get(suite.ctx, proposal3ID)
	suite.Require().Nil(err)
	suite.Require().True(proposal3.VotingStartTime.Equal(suite.ctx.BlockHeader().Time))
	// add vote
	voteOptions := []*v1.WeightedVoteOption{{Option: v1.OptionYes, Weight: "1.0"}}
	err = suite.govKeeper.AddVote(suite.ctx, proposal3ID, suite.addrs[0], voteOptions, "")
	suite.Require().NoError(err)

	testCases := []struct {
		name        string
		malleate    func() (proposalID uint64, proposer string)
		proposalID  uint64
		proposer    string
		expectedErr bool
	}{
		{
			name: "without proposer",
			malleate: func() (uint64, string) {
				return 1, ""
			},
			expectedErr: true,
		},
		{
			name: "invalid proposal id",
			malleate: func() (uint64, string) {
				return 1, suite.addrs[1].String()
			},
			expectedErr: true,
		},
		{
			name: "valid proposalID but invalid proposer",
			malleate: func() (uint64, string) {
				return proposalID, suite.addrs[1].String()
			},
			expectedErr: true,
		},
		{
			name: "valid proposalID but invalid proposal which has already passed",
			malleate: func() (uint64, string) {
				// making proposal status pass
				proposal2, err := suite.govKeeper.Proposals.Get(suite.ctx, proposal2ID)
				suite.Require().Nil(err)

				proposal2.Status = v1.ProposalStatus_PROPOSAL_STATUS_PASSED
				err = suite.govKeeper.SetProposal(suite.ctx, proposal2)
				suite.Require().NoError(err)
				return proposal2ID, suite.addrs[1].String()
			},
			expectedErr: true,
		},
		{
			name: "valid proposer and proposal id",
			malleate: func() (uint64, string) {
				return proposalID, suite.addrs[0].String()
			},
			expectedErr: false,
		},
		{
			name: "valid case with deletion of votes",
			malleate: func() (uint64, string) {
				suite.Require().NoError(suite.govKeeper.ActivateVotingPeriod(suite.ctx, proposal))

				proposal, err = suite.govKeeper.Proposals.Get(suite.ctx, proposal.Id)
				suite.Require().Nil(err)
				suite.Require().True(proposal.VotingStartTime.Equal(suite.ctx.BlockHeader().Time))

				// add vote
				voteOptions := []*v1.WeightedVoteOption{{Option: v1.OptionYes, Weight: "1.0"}}
				err = suite.govKeeper.AddVote(suite.ctx, proposalID, suite.addrs[0], voteOptions, "")
				suite.Require().NoError(err)
				vote, err := suite.govKeeper.Votes.Get(suite.ctx, collections.Join(proposalID, suite.addrs[0]))
				suite.Require().NoError(err)
				suite.Require().NotNil(vote)

				return proposalID, suite.addrs[0].String()
			},
			expectedErr: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			pID, proposer := tc.malleate()
			err = suite.govKeeper.CancelProposal(suite.ctx, pID, proposer)
			if tc.expectedErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
	_, err = suite.govKeeper.Votes.Get(suite.ctx, collections.Join(proposalID, suite.addrs[0]))
	suite.Require().ErrorContains(err, collections.ErrNotFound.Error())

	// check that proposal 3 votes are still present in the state
	votes, err := suite.govKeeper.Votes.Get(suite.ctx, collections.Join(proposal3ID, suite.addrs[0]))
	suite.Require().NoError(err)
	suite.Require().NotNil(votes)
}

func TestMigrateProposalMessages(t *testing.T) {
	content := v1beta1.NewTextProposal("Test", "description")
	contentMsg, err := v1.NewLegacyContent(content, sdk.AccAddress("test1").String())
	require.NoError(t, err)
	content, err = v1.LegacyContentFromMessage(contentMsg)
	require.NoError(t, err)
	require.Equal(t, "Test", content.GetTitle())
	require.Equal(t, "description", content.GetDescription())
}
