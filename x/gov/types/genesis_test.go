package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEqualProposalID(t *testing.T) {
	state1 := GenesisState{}
	state2 := GenesisState{}
	require.Equal(t, state1, state2)

	// Proposals
	state1.StartingProposalId = 1
	require.NotEqual(t, state1, state2)
	require.False(t, state1.Equal(state2))

	state2.StartingProposalId = 1
	require.Equal(t, state1, state2)
	require.True(t, state1.Equal(state2))
}
