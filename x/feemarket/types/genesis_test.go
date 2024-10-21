package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/x/feemarket/types"
)

func TestGenesis(t *testing.T) {
	t.Run("can create a new default genesis state", func(t *testing.T) {
		gs := types.DefaultGenesisState()
		require.NoError(t, gs.ValidateBasic())
	})

	t.Run("can accept a valid genesis state for AIMD eip-1559", func(t *testing.T) {
		gs := types.DefaultAIMDGenesisState()
		require.NoError(t, gs.ValidateBasic())
	})
}
