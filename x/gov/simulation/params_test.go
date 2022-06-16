package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/gov/simulation"
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
		{"gov/votingparams", "votingparams", "{\"voting_period\": \"251681000000000\"}", "gov"},
		{"gov/votingparams", "votingparams", "{\"expedited_voting_period\": \"53176000000000\"}", "gov"},
		{"gov/depositparams", "depositparams", "{\"max_deposit_period\": \"153577000000000\"}", "gov"},
		{"gov/tallyparams", "tallyparams", "{\"quorum\":\"0.429000000000000000\",\"veto\":\"0.323000000000000000\"}", "gov"},
	}

	paramChanges := simulation.ParamChanges(r)
	require.Len(t, paramChanges, 4)

	for i, p := range paramChanges {

		require.Equal(t, expected[i].composedKey, p.ComposedKey())
		require.Equal(t, expected[i].key, p.Key())
		require.Equal(t, expected[i].simValue, p.SimValue()(r))
		require.Equal(t, expected[i].subspace, p.Subspace())
	}
}
