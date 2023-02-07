package keeper_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

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
		proposal, err := suite.govKeeper.SubmitProposal(suite.ctx, tp, "", "test", "summary", sdk.AccAddress("cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r"), tc.expedited)
		suite.Require().NoError(err)
		proposalID := proposal.Id
		suite.govKeeper.SetProposal(suite.ctx, proposal)

		gotProposal, ok := suite.govKeeper.GetProposal(suite.ctx, proposalID)
		suite.Require().True(ok)
		suite.Require().Equal(proposal, gotProposal)
	}
}

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
		suite.Require().PanicsWithValue(fmt.Sprintf("couldn't find proposal with id#%d", 10),
			func() {
				suite.govKeeper.DeleteProposal(suite.ctx, 10)
			},
		)
		tp := TestProposal
		proposal, err := suite.govKeeper.SubmitProposal(suite.ctx, tp, "", "test", "summary", sdk.AccAddress("cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r"), tc.expedited)
		suite.Require().NoError(err)
		proposalID := proposal.Id
		suite.govKeeper.SetProposal(suite.ctx, proposal)
		suite.Require().NotPanics(func() {
			suite.govKeeper.DeleteProposal(suite.ctx, proposalID)
		}, "")
	}
}

func (suite *KeeperTestSuite) TestActivateVotingPeriod() {
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
		proposal, err := suite.govKeeper.SubmitProposal(suite.ctx, tp, "", "test", "summary", sdk.AccAddress("cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r"), tc.expedited)
		suite.Require().NoError(err)

		suite.Require().Nil(proposal.VotingStartTime)

		suite.govKeeper.ActivateVotingPeriod(suite.ctx, proposal)

		proposal, ok := suite.govKeeper.GetProposal(suite.ctx, proposal.Id)
		suite.Require().True(ok)
		suite.Require().True(proposal.VotingStartTime.Equal(suite.ctx.BlockHeader().Time))

		activeIterator := suite.govKeeper.ActiveProposalQueueIterator(suite.ctx, *proposal.VotingEndTime)
		suite.Require().True(activeIterator.Valid())

		proposalID := types.GetProposalIDFromBytes(activeIterator.Value())
		suite.Require().Equal(proposalID, proposal.Id)
		activeIterator.Close()
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
		_, err = suite.govKeeper.SubmitProposal(suite.ctx, []sdk.Msg{prop}, tc.metadata, "title", "", sdk.AccAddress("cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r"), tc.expedited)
		suite.Require().True(errors.Is(tc.expectedErr, err), "tc #%d; got: %v, expected: %v", i, err, tc.expectedErr)
	}
}

func (suite *KeeperTestSuite) TestGetProposalsFiltered() {
	proposalID := uint64(1)
	status := []v1.ProposalStatus{v1.StatusDepositPeriod, v1.StatusVotingPeriod}

	addr1 := sdk.AccAddress("foo_________________")

	for _, s := range status {
		for i := 0; i < 50; i++ {
			p, err := v1.NewProposal(TestProposal, proposalID, time.Now(), time.Now(), "metadata", "title", "summary", sdk.AccAddress("cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r"), false)
			suite.Require().NoError(err)

			p.Status = s

			if i%2 == 0 {
				d := v1.NewDeposit(proposalID, addr1, nil)
				v := v1.NewVote(proposalID, addr1, v1.NewNonSplitVoteOption(v1.OptionYes), "")
				suite.govKeeper.SetDeposit(suite.ctx, d)
				suite.govKeeper.SetVote(suite.ctx, v)
			}

			suite.govKeeper.SetProposal(suite.ctx, p)
			proposalID++
		}
	}

	testCases := []struct {
		params             v1.QueryProposalsParams
		expectedNumResults int
	}{
		{v1.NewQueryProposalsParams(1, 50, v1.StatusNil, nil, nil), 50},
		{v1.NewQueryProposalsParams(1, 50, v1.StatusDepositPeriod, nil, nil), 50},
		{v1.NewQueryProposalsParams(1, 50, v1.StatusVotingPeriod, nil, nil), 50},
		{v1.NewQueryProposalsParams(1, 25, v1.StatusNil, nil, nil), 25},
		{v1.NewQueryProposalsParams(2, 25, v1.StatusNil, nil, nil), 25},
		{v1.NewQueryProposalsParams(1, 50, v1.StatusRejected, nil, nil), 0},
		{v1.NewQueryProposalsParams(1, 50, v1.StatusNil, addr1, nil), 50},
		{v1.NewQueryProposalsParams(1, 50, v1.StatusNil, nil, addr1), 50},
		{v1.NewQueryProposalsParams(1, 50, v1.StatusNil, addr1, addr1), 50},
		{v1.NewQueryProposalsParams(1, 50, v1.StatusDepositPeriod, addr1, addr1), 25},
		{v1.NewQueryProposalsParams(1, 50, v1.StatusDepositPeriod, nil, nil), 50},
		{v1.NewQueryProposalsParams(1, 50, v1.StatusVotingPeriod, nil, nil), 50},
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Test Case %d", i), func() {
			proposals := suite.govKeeper.GetProposalsFiltered(suite.ctx, tc.params)
			suite.Require().Len(proposals, tc.expectedNumResults)

			for _, p := range proposals {
				if v1.ValidProposalStatus(tc.params.ProposalStatus) {
					suite.Require().Equal(tc.params.ProposalStatus, p.Status)
				}
			}
		})
	}
}

func (suite *KeeperTestSuite) TestCancelProposal() {
	govAcct := suite.govKeeper.GetGovernanceAccount(suite.ctx).GetAddress().String()
	tp := v1beta1.TextProposal{Title: "title", Description: "description"}
	prop, err := v1.NewLegacyContent(&tp, govAcct)
	suite.Require().NoError(err)
	proposalResp, err := suite.govKeeper.SubmitProposal(suite.ctx, []sdk.Msg{prop}, "", "title", "summary", suite.addrs[0], false)
	suite.Require().NoError(err)
	proposalID := proposalResp.Id

	proposal2Resp, err := suite.govKeeper.SubmitProposal(suite.ctx, []sdk.Msg{prop}, "", "title", "summary", suite.addrs[1], true)
	proposal2ID := proposal2Resp.Id
	makeProposalPass := func() {
		proposal2, ok := suite.govKeeper.GetProposal(suite.ctx, proposal2ID)
		suite.Require().True(ok)

		proposal2.Status = v1.ProposalStatus_PROPOSAL_STATUS_PASSED
		suite.govKeeper.SetProposal(suite.ctx, proposal2)
	}

	testCases := []struct {
		name        string
		proposalID  uint64
		proposer    string
		expectedErr bool
	}{
		{
			name:        "without proposer",
			proposalID:  1,
			proposer:    "",
			expectedErr: true,
		},
		{
			name:        "invalid proposal id",
			proposalID:  1,
			proposer:    string(suite.addrs[0]),
			expectedErr: true,
		},
		{
			name:        "valid proposalID but invalid proposer",
			proposalID:  proposalID,
			proposer:    suite.addrs[1].String(),
			expectedErr: true,
		},
		{
			name:        "valid proposalID but invalid proposal which has already passed",
			proposalID:  proposal2ID,
			proposer:    suite.addrs[1].String(),
			expectedErr: true,
		},
		{
			name:        "valid proposer and proposal id",
			proposalID:  proposalID,
			proposer:    suite.addrs[0].String(),
			expectedErr: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			if tc.proposalID == proposal2ID {
				// making proposal status pass
				makeProposalPass()
			}
			err = suite.govKeeper.CancelProposal(suite.ctx, tc.proposalID, tc.proposer)
			if tc.expectedErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
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
