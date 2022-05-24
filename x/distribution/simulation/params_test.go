package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/distribution/simulation"
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
		{"distribution/communitytax", "communitytax", "\"0.120000000000000000\"", "distribution"},
		{"distribution/baseproposerreward", "baseproposerreward", "\"0.280000000000000000\"", "distribution"},
		{"distribution/bonusproposerreward", "bonusproposerreward", "\"0.180000000000000000\"", "distribution"},
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
