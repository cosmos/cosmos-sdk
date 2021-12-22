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
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addrs := simapp.AddTestAddrsIncremental(app, ctx, 5, sdk.NewInt(30000000))

	tp := []sdk.Msg{}
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalId

	var invalidOption types.VoteOption = 0x10

	require.Error(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[0], types.NewNonSplitVoteOption(types.OptionYes)), "proposal not on voting period")
	require.Error(t, app.GovKeeper.AddVote(ctx, 10, addrs[0], types.NewNonSplitVoteOption(types.OptionYes)), "invalid proposal ID")

	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	require.Error(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[0], types.NewNonSplitVoteOption(invalidOption)), "invalid option")

	// Test first vote
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[0], types.NewNonSplitVoteOption(types.OptionAbstain)))
	vote, found := app.GovKeeper.GetVote(ctx, proposalID, addrs[0])
	require.True(t, found)
	require.Equal(t, addrs[0].String(), vote.Voter)
	require.Equal(t, proposalID, vote.ProposalId)
	require.True(t, len(vote.Options) == 1)
	require.Equal(t, types.OptionAbstain, vote.Options[0].Option)
	require.Equal(t, types.OptionAbstain, vote.Option)

	// Test change of vote
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[0], types.NewNonSplitVoteOption(types.OptionYes)))
	vote, found = app.GovKeeper.GetVote(ctx, proposalID, addrs[0])
	require.True(t, found)
	require.Equal(t, addrs[0].String(), vote.Voter)
	require.Equal(t, proposalID, vote.ProposalId)
	require.True(t, len(vote.Options) == 1)
	require.Equal(t, types.OptionYes, vote.Options[0].Option)
	require.Equal(t, types.OptionYes, vote.Option)

	// Test second vote
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, addrs[1], types.WeightedVoteOptions{
		types.NewWeightedVoteOption(types.OptionYes, sdk.NewDecWithPrec(60, 2)),
		types.NewWeightedVoteOption(types.OptionNo, sdk.NewDecWithPrec(30, 2)),
		types.NewWeightedVoteOption(types.OptionAbstain, sdk.NewDecWithPrec(5, 2)),
		types.NewWeightedVoteOption(types.OptionNoWithVeto, sdk.NewDecWithPrec(5, 2)),
	}))
	vote, found = app.GovKeeper.GetVote(ctx, proposalID, addrs[1])
	require.True(t, found)
	require.Equal(t, addrs[1].String(), vote.Voter)
	require.Equal(t, proposalID, vote.ProposalId)
	require.True(t, len(vote.Options) == 4)
	require.Equal(t, types.OptionYes, vote.Options[0].Option)
	require.Equal(t, types.OptionNo, vote.Options[1].Option)
	require.Equal(t, types.OptionAbstain, vote.Options[2].Option)
	require.Equal(t, types.OptionNoWithVeto, vote.Options[3].Option)
	require.Equal(t, vote.Options[0].Weight, sdk.NewDecWithPrec(60, 2).String())
	require.Equal(t, vote.Options[1].Weight, sdk.NewDecWithPrec(30, 2).String())
	require.Equal(t, vote.Options[2].Weight, sdk.NewDecWithPrec(5, 2).String())
	require.Equal(t, vote.Options[3].Weight, sdk.NewDecWithPrec(5, 2).String())
	require.Equal(t, types.OptionEmpty, vote.Option)

	// Test vote iterator
	// NOTE order of deposits is determined by the addresses
	votes := app.GovKeeper.GetAllVotes(ctx)
	require.Len(t, votes, 2)
	require.Equal(t, votes, app.GovKeeper.GetVotes(ctx, proposalID))
	require.Equal(t, addrs[0].String(), votes[0].Voter)
	require.Equal(t, proposalID, votes[0].ProposalId)
	require.True(t, len(votes[0].Options) == 1)
	require.Equal(t, types.OptionYes, votes[0].Options[0].Option)
	require.Equal(t, addrs[1].String(), votes[1].Voter)
	require.Equal(t, proposalID, votes[1].ProposalId)
	require.True(t, len(votes[1].Options) == 4)
	require.Equal(t, votes[1].Options[0].Weight, sdk.NewDecWithPrec(60, 2).String())
	require.Equal(t, votes[1].Options[1].Weight, sdk.NewDecWithPrec(30, 2).String())
	require.Equal(t, votes[1].Options[2].Weight, sdk.NewDecWithPrec(5, 2).String())
	require.Equal(t, votes[1].Options[3].Weight, sdk.NewDecWithPrec(5, 2).String())
	require.Equal(t, types.OptionEmpty, vote.Option)
}
