package keeper_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/x/gov/types"
	v1 "cosmossdk.io/x/gov/types/v1"
	"cosmossdk.io/x/gov/types/v1beta1"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO(tip): remove this
func (suite *KeeperTestSuite) TestGetSetProposal() {
	testCases := map[string]struct {
		proposalType v1.ProposalType
	}{
		"unspecified proposal type": {},
		"regular proposal": {
			proposalType: v1.ProposalType_PROPOSAL_TYPE_STANDARD,
		},
		"expedited proposal": {
			proposalType: v1.ProposalType_PROPOSAL_TYPE_EXPEDITED,
		},
	}

	for _, tc := range testCases {
		tp := TestProposal
		proposal, err := suite.govKeeper.SubmitProposal(suite.ctx, tp, "", "test", "summary", suite.addrs[0], tc.proposalType)
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
		proposalType v1.ProposalType
	}{
		"unspecified proposal type": {},
		"regular proposal": {
			proposalType: v1.ProposalType_PROPOSAL_TYPE_STANDARD,
		},
		"expedited proposal": {
			proposalType: v1.ProposalType_PROPOSAL_TYPE_EXPEDITED,
		},
	}

	for _, tc := range testCases {
		// delete non-existing proposal
		suite.Require().ErrorIs(suite.govKeeper.DeleteProposal(suite.ctx, 10), collections.ErrNotFound)

		tp := TestProposal
		proposal, err := suite.govKeeper.SubmitProposal(suite.ctx, tp, "", "test", "summary", suite.addrs[0], tc.proposalType)
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
		name         string
		proposalType v1.ProposalType
	}{
		{name: "unspecified proposal type"},
		{name: "regular proposal", proposalType: v1.ProposalType_PROPOSAL_TYPE_STANDARD},
		{name: "expedited proposal", proposalType: v1.ProposalType_PROPOSAL_TYPE_EXPEDITED},
	}

	for _, tc := range testCases {
		tp := TestProposal
		proposal, err := suite.govKeeper.SubmitProposal(suite.ctx, tp, "", "test", "summary", suite.addrs[0], tc.proposalType)
		suite.Require().NoError(err)

		suite.Require().Nil(proposal.VotingStartTime)

		err = suite.govKeeper.ActivateVotingPeriod(suite.ctx, proposal)
		suite.Require().NoError(err)

		proposal, err = suite.govKeeper.Proposals.Get(suite.ctx, proposal.Id)
		suite.Require().Nil(err)
		suite.Require().True(proposal.VotingStartTime.Equal(suite.ctx.HeaderInfo().Time))

		has, err := suite.govKeeper.ActiveProposalsQueue.Has(suite.ctx, collections.Join(*proposal.VotingEndTime, proposal.Id))
		suite.Require().NoError(err)
		suite.Require().True(has)
		suite.Require().NoError(suite.govKeeper.DeleteProposal(suite.ctx, proposal.Id))
	}
}

func (suite *KeeperTestSuite) TestDeleteProposalInVotingPeriod() {
	testCases := []struct {
		name         string
		proposalType v1.ProposalType
	}{
		{name: "unspecified proposal type"},
		{name: "regular proposal", proposalType: v1.ProposalType_PROPOSAL_TYPE_STANDARD},
		{name: "expedited proposal", proposalType: v1.ProposalType_PROPOSAL_TYPE_EXPEDITED},
	}

	for _, tc := range testCases {
		suite.reset()
		tp := TestProposal
		proposal, err := suite.govKeeper.SubmitProposal(suite.ctx, tp, "", "test", "summary", suite.addrs[0], tc.proposalType)
		suite.Require().NoError(err)
		suite.Require().Nil(proposal.VotingStartTime)

		suite.Require().NoError(suite.govKeeper.ActivateVotingPeriod(suite.ctx, proposal))

		proposal, err = suite.govKeeper.Proposals.Get(suite.ctx, proposal.Id)
		suite.Require().Nil(err)
		suite.Require().True(proposal.VotingStartTime.Equal(suite.ctx.HeaderInfo().Time))

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
		content      v1beta1.Content
		authority    string
		metadata     string
		proposalType v1.ProposalType
		expectedErr  error
	}{
		{&tp, govAcct, "", v1.ProposalType_PROPOSAL_TYPE_STANDARD, nil},
		{&tp, govAcct, "", v1.ProposalType_PROPOSAL_TYPE_EXPEDITED, nil},
		// Keeper does not check the validity of title and description, no error
		{&v1beta1.TextProposal{Title: "", Description: "description"}, govAcct, "", v1.ProposalType_PROPOSAL_TYPE_STANDARD, nil},
		{&v1beta1.TextProposal{Title: strings.Repeat("1234567890", 100), Description: "description"}, govAcct, "", v1.ProposalType_PROPOSAL_TYPE_STANDARD, nil},
		{&v1beta1.TextProposal{Title: "title", Description: ""}, govAcct, "", v1.ProposalType_PROPOSAL_TYPE_STANDARD, nil},
		{&v1beta1.TextProposal{Title: "title", Description: strings.Repeat("1234567890", 1000)}, govAcct, "", v1.ProposalType_PROPOSAL_TYPE_EXPEDITED, nil},
		// error when signer is not gov acct
		{&tp, randomAddr.String(), "", v1.ProposalType_PROPOSAL_TYPE_STANDARD, types.ErrInvalidSigner},
		// error only when invalid route
		{&invalidProposalRoute{}, govAcct, "", v1.ProposalType_PROPOSAL_TYPE_STANDARD, types.ErrNoProposalHandlerExists},
	}

	for i, tc := range testCases {
		prop, err := v1.NewLegacyContent(tc.content, tc.authority)
		suite.Require().NoError(err)
		_, err = suite.govKeeper.SubmitProposal(suite.ctx, []sdk.Msg{prop}, tc.metadata, "title", "", suite.addrs[0], tc.proposalType)
		suite.Require().True(errors.Is(tc.expectedErr, err), "tc #%d; got: %v, expected: %v", i, err, tc.expectedErr)
	}
}

func (suite *KeeperTestSuite) TestCancelProposal() {
	govAcct := suite.govKeeper.GetGovernanceAccount(suite.ctx).GetAddress().String()
	tp := v1beta1.TextProposal{Title: "title", Description: "description"}
	prop, err := v1.NewLegacyContent(&tp, govAcct)
	suite.Require().NoError(err)
	proposal, err := suite.govKeeper.SubmitProposal(suite.ctx, []sdk.Msg{prop}, "", "title", "summary", suite.addrs[0], v1.ProposalType_PROPOSAL_TYPE_STANDARD)
	suite.Require().NoError(err)
	proposalID := proposal.Id

	proposal2, err := suite.govKeeper.SubmitProposal(suite.ctx, []sdk.Msg{prop}, "", "title", "summary", suite.addrs[1], v1.ProposalType_PROPOSAL_TYPE_EXPEDITED)
	suite.Require().NoError(err)
	proposal2ID := proposal2.Id

	// proposal3 is only used to check the votes for proposals which doesn't go through `CancelProposal` are still present in state
	proposal3, err := suite.govKeeper.SubmitProposal(suite.ctx, []sdk.Msg{prop}, "", "title", "summary", suite.addrs[2], v1.ProposalType_PROPOSAL_TYPE_STANDARD)
	suite.Require().NoError(err)
	proposal3ID := proposal3.Id

	// add votes for proposal 3
	suite.Require().NoError(suite.govKeeper.ActivateVotingPeriod(suite.ctx, proposal3))

	proposal3, err = suite.govKeeper.Proposals.Get(suite.ctx, proposal3ID)
	suite.Require().Nil(err)
	fmt.Println(suite.ctx.HeaderInfo().Time, proposal3.VotingStartTime)
	suite.Require().True(proposal3.VotingStartTime.Equal(suite.ctx.HeaderInfo().Time))
	// add vote
	voteOptions := []*v1.WeightedVoteOption{{Option: v1.OptionYes, Weight: "1.0"}}
	err = suite.govKeeper.AddVote(suite.ctx, proposal3ID, suite.addrs[0], voteOptions, "")
	suite.Require().NoError(err)

	testCases := []struct {
		name           string
		malleate       func() (proposalID uint64, proposer string)
		proposalID     uint64
		proposer       string
		expectedErrMsg string
	}{
		{
			name: "without proposer",
			malleate: func() (uint64, string) {
				return proposalID, ""
			},
			expectedErrMsg: "invalid proposer",
		},
		{
			name: "invalid proposal id",
			malleate: func() (uint64, string) {
				return 10, suite.addrs[1].String()
			},
			expectedErrMsg: "proposal 10 doesn't exist",
		},
		{
			name: "valid proposalID but invalid proposer",
			malleate: func() (uint64, string) {
				return proposalID, suite.addrs[1].String()
			},
			expectedErrMsg: "invalid proposer",
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
			expectedErrMsg: "proposal should be in the deposit or voting period",
		},
		{
			name: "proposal canceled too late",
			malleate: func() (uint64, string) {
				suite.Require().NoError(suite.govKeeper.ActivateVotingPeriod(suite.ctx, proposal2))

				proposal2, err = suite.govKeeper.Proposals.Get(suite.ctx, proposal2.Id)
				suite.Require().Nil(err)

				headerInfo := suite.ctx.HeaderInfo()
				// try to cancel 1min before the end of the voting period
				// this should fail, as we allow to cancel proposal (by default) up to 1/2 of the voting period
				headerInfo.Time = proposal2.VotingEndTime.Add(-1 * time.Minute)
				suite.ctx = suite.ctx.WithHeaderInfo(headerInfo)

				suite.Require().NoError(err)
				return proposal2ID, suite.addrs[1].String()
			},
			expectedErrMsg: "too late",
		},
		{
			name: "valid proposer and proposal id",
			malleate: func() (uint64, string) {
				return proposalID, suite.addrs[0].String()
			},
		},
		{
			name: "valid case with deletion of votes",
			malleate: func() (uint64, string) {
				suite.Require().NoError(suite.govKeeper.ActivateVotingPeriod(suite.ctx, proposal))

				proposal, err = suite.govKeeper.Proposals.Get(suite.ctx, proposal.Id)
				suite.Require().Nil(err)
				suite.Require().True(proposal.VotingStartTime.Equal(suite.ctx.HeaderInfo().Time))

				// add vote
				voteOptions := []*v1.WeightedVoteOption{{Option: v1.OptionYes, Weight: "1.0"}}
				err = suite.govKeeper.AddVote(suite.ctx, proposalID, suite.addrs[0], voteOptions, "")
				suite.Require().NoError(err)
				vote, err := suite.govKeeper.Votes.Get(suite.ctx, collections.Join(proposalID, suite.addrs[0]))
				suite.Require().NoError(err)
				suite.Require().NotNil(vote)

				return proposalID, suite.addrs[0].String()
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			pID, proposer := tc.malleate()
			err = suite.govKeeper.CancelProposal(suite.ctx, pID, proposer)
			if tc.expectedErrMsg != "" {
				suite.Require().ErrorContains(err, tc.expectedErrMsg)
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
