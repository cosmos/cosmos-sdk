package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

func TestVotes(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addrs := simapp.AddTestAddrsIncremental(app, ctx, 5, sdk.NewInt(30000000))

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalId

	var invalidOption types.VoteOption = 0x10

	require.Error(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[0], types.SubVotes{types.NewSubVote(types.OptionYes, 1)}), "proposal not on voting period")
	require.Error(t, app.GovKeeper.AddVote(ctx, 10, addrs[0], types.SubVotes{types.NewSubVote(types.OptionYes, 1)}), "invalid proposal ID")

	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	require.Error(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[0], types.SubVotes{types.NewSubVote(invalidOption, 1)}), "invalid option")

	// Test first vote
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[0], types.SubVotes{types.NewSubVote(types.OptionAbstain, 1)}))
	vote, found := app.GovKeeper.GetVote(ctx, proposalID, addrs[0])
	require.True(t, found)
	require.Equal(t, addrs[0].String(), vote.Voter)
	require.Equal(t, proposalID, vote.ProposalId)
	require.True(t, len(vote.SubVotes) == 1)
	require.Equal(t, types.OptionAbstain, vote.SubVotes[0].Option)

	// Test change of vote
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[0], types.SubVotes{types.NewSubVote(types.OptionYes, 1)}))
	vote, found = app.GovKeeper.GetVote(ctx, proposalID, addrs[0])
	require.True(t, found)
	require.Equal(t, addrs[0].String(), vote.Voter)
	require.Equal(t, proposalID, vote.ProposalId)
	require.True(t, len(vote.SubVotes) == 1)
	require.Equal(t, types.OptionYes, vote.SubVotes[0].Option)

	// Test second vote
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[1], types.SubVotes{
		types.SubVote{Option: types.OptionYes, Rate: sdk.NewDecWithPrec(60, 2)},
		types.SubVote{Option: types.OptionNo, Rate: sdk.NewDecWithPrec(30, 2)},
		types.SubVote{Option: types.OptionAbstain, Rate: sdk.NewDecWithPrec(5, 2)},
		types.SubVote{Option: types.OptionNoWithVeto, Rate: sdk.NewDecWithPrec(5, 2)},
	}))
	vote, found = app.GovKeeper.GetVote(ctx, proposalID, addrs[1])
	require.True(t, found)
	require.Equal(t, addrs[1].String(), vote.Voter)
	require.Equal(t, proposalID, vote.ProposalId)
	require.True(t, len(vote.SubVotes) == 4)
	require.Equal(t, types.OptionYes, vote.SubVotes[0].Option)
	require.Equal(t, types.OptionNo, vote.SubVotes[1].Option)
	require.Equal(t, types.OptionAbstain, vote.SubVotes[2].Option)
	require.Equal(t, types.OptionNoWithVeto, vote.SubVotes[3].Option)
	require.True(t, vote.SubVotes[0].Rate.Equal(sdk.NewDecWithPrec(60, 2)))
	require.True(t, vote.SubVotes[1].Rate.Equal(sdk.NewDecWithPrec(30, 2)))
	require.True(t, vote.SubVotes[2].Rate.Equal(sdk.NewDecWithPrec(5, 2)))
	require.True(t, vote.SubVotes[3].Rate.Equal(sdk.NewDecWithPrec(5, 2)))

	// Test vote iterator
	// NOTE order of deposits is determined by the addresses
	votes := app.GovKeeper.GetAllVotes(ctx)
	require.Len(t, votes, 2)
	require.Equal(t, votes, app.GovKeeper.GetVotes(ctx, proposalID))
	require.Equal(t, addrs[0].String(), votes[0].Voter)
	require.Equal(t, proposalID, votes[0].ProposalId)
	require.True(t, len(votes[0].SubVotes) == 1)
	require.Equal(t, types.OptionYes, votes[0].SubVotes[0].Option)
	require.Equal(t, addrs[1].String(), votes[1].Voter)
	require.Equal(t, proposalID, votes[1].ProposalId)
	require.True(t, len(votes[1].SubVotes) == 4)
	require.True(t, votes[1].SubVotes[0].Rate.Equal(sdk.NewDecWithPrec(60, 2)))
	require.True(t, votes[1].SubVotes[1].Rate.Equal(sdk.NewDecWithPrec(30, 2)))
	require.True(t, votes[1].SubVotes[2].Rate.Equal(sdk.NewDecWithPrec(5, 2)))
	require.True(t, votes[1].SubVotes[3].Rate.Equal(sdk.NewDecWithPrec(5, 2)))
}
