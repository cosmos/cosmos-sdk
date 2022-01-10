package v1beta2_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta2"
	"github.com/stretchr/testify/require"
)

func TestEmptyGenesis(t *testing.T) {
	state1 := v1beta2.GenesisState{}
	require.True(t, state1.Empty())

	state2 := v1beta2.DefaultGenesisState()
	require.False(t, state2.Empty())
}
