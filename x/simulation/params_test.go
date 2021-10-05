package simulation

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParamChange(t *testing.T) {
	subspace, key := "theSubspace", "key"
	f := func(r *rand.Rand) string {
		return "theResult"
	}

	pChange := NewSimParamChange(subspace, key, f)

	require.Equal(t, subspace, pChange.Subspace())
	require.Equal(t, key, pChange.Key())
	require.Equal(t, f(nil), pChange.SimValue()(nil))
	require.Equal(t, fmt.Sprintf("%s/%s", subspace, key), pChange.ComposedKey())
}
