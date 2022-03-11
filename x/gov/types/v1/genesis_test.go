package v1_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/require"
)

func TestEmptyGenesis(t *testing.T) {
	state1 := v1.GenesisState{}
	require.True(t, state1.Empty())

	state2 := v1.DefaultGenesisState()
	require.False(t, state2.Empty())
}
