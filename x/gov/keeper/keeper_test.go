package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

func TestIncrementProposalNumber(t *testing.T) {
	ctx, _, keeper, _, _ := createTestInput(t, false, 100)

	tp := TestProposal
	keeper.SubmitProposal(ctx, tp)
	keeper.SubmitProposal(ctx, tp)
	keeper.SubmitProposal(ctx, tp)
	keeper.SubmitProposal(ctx, tp)
	keeper.SubmitProposal(ctx, tp)
	proposal6, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)

	require.Equal(t, uint64(6), proposal6.ProposalID)
}

func TestProposalQueues(t *testing.T) {
	ctx, _, keeper, _, _ := createTestInput(t, false, 100)

	// create test proposals
	tp := TestProposal
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)

	inactiveIterator := keeper.InactiveProposalQueueIterator(ctx, proposal.DepositEndTime)
	require.True(t, inactiveIterator.Valid())

	proposalID := types.GetProposalIDFromBytes(inactiveIterator.Value())
	require.Equal(t, proposalID, proposal.ProposalID)
	inactiveIterator.Close()

	keeper.activateVotingPeriod(ctx, proposal)

	proposal, ok := keeper.GetProposal(ctx, proposal.ProposalID)
	require.True(t, ok)

	activeIterator := keeper.ActiveProposalQueueIterator(ctx, proposal.VotingEndTime)
	require.True(t, activeIterator.Valid())
	keeper.cdc.UnmarshalBinaryLengthPrefixed(activeIterator.Value(), &proposalID)
	require.Equal(t, proposalID, proposal.ProposalID)
	activeIterator.Close()
}
