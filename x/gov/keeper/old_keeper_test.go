package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

func TestProposalQueues(t *testing.T) {
	ctx, _, _, keeper, _, _ := createTestInput(t, false, 100) // nolint: dogsled

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
