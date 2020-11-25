package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/slashing/simulation"
)

func TestParamChanges(t *testing.T) {
	s := rand.NewSource(1)
	r := rand.New(s)

	expected := []struct {
		composedKey string
		key         string
		simValue    string
		subspace    string
	}{
		{"slashing/SignedBlocksWindow", "SignedBlocksWindow", "\"231\"", "slashing"},
		{"slashing/MinSignedPerWindow", "MinSignedPerWindow", "\"0.700000000000000000\"", "slashing"},
		{"slashing/SlashFractionDowntime", "SlashFractionDowntime", "\"0.020833333333333333\"", "slashing"},
	}

	paramChanges := simulation.ParamChanges(r)

	require.Len(t, paramChanges, 3)

	for i, p := range paramChanges {
		require.Equal(t, expected[i].composedKey, p.ComposedKey())
		require.Equal(t, expected[i].key, p.Key())
		require.Equal(t, expected[i].simValue, p.SimValue()(r))
		require.Equal(t, expected[i].subspace, p.Subspace())
	}
}
