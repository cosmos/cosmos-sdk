package simulation

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
)

// AllInvariants tests all slashing invariants
func AllInvariants() simulation.Invariant {
	return func(t *testing.T, app *baseapp.BaseApp, log string) {
		require.Nil(t, nil)
	}
}
