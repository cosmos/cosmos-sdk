package poa_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/poa"
)

func TestInitDefaultGenesis(t *testing.T) {
	genesisState := poa.DefaultGenesisState()
	require.Equal(t, 0, len(genesisState.Validators))
	require.Equal(t, 0, len(genesisState.LastValidatorPowers))
}

// func TestInitGenesis(t *testing.T) {
//
// }
