package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
)

func TestEqualProposalID(t *testing.T) {
	state1 := GenesisState{}
	state2 := GenesisState{}
	require.Equal(t, state1, state2)

	// Proposals
	state1.StartingProposalID = 1
	require.NotEqual(t, state1, state2)
	require.False(t, state1.Equal(state2))

	state2.StartingProposalID = 1
	require.Equal(t, state1, state2)
	require.True(t, state1.Equal(state2))
}

func TestEqualProposals(t *testing.T) {
	// Generate mock app and keepers
	input := getMockApp(t, 2, GenesisState{}, nil)
	SortAddresses(input.addrs)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.BaseApp.NewContext(false, abci.Header{})

	// Submit two proposals
	proposal := testProposal()
	proposal1, err := input.keeper.SubmitProposal(ctx, proposal)
	require.NoError(t, err)
	proposal2, err := input.keeper.SubmitProposal(ctx, proposal)
	require.NoError(t, err)

	// They are similar but their IDs should be different
	require.NotEqual(t, proposal1, proposal2)
	require.False(t, ProposalEqual(proposal1, proposal2))

	// Now create two genesis blocks
	state1 := GenesisState{Proposals: []Proposal{proposal1}}
	state2 := GenesisState{Proposals: []Proposal{proposal2}}
	require.NotEqual(t, state1, state2)
	require.False(t, state1.Equal(state2))

	// Now make proposals identical by setting both IDs to 55
	proposal1.ProposalID = 55
	proposal2.ProposalID = 55
	require.Equal(t, proposal1, proposal1)
	require.True(t, ProposalEqual(proposal1, proposal2))

	// Reassign proposals into state
	state1.Proposals[0] = proposal1
	state2.Proposals[0] = proposal2

	// State should be identical now..
	require.Equal(t, state1, state2)
	require.True(t, state1.Equal(state2))
}