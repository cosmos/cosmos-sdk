package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmptyGenesis(t *testing.T) {
	state1 := GenesisState{}
	require.True(t, state1.Empty())

	state2 := DefaultGenesisState()
	require.False(t, state2.Empty())
}
