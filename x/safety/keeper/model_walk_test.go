//go:build sims

package keeper_test

import (
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestKeeperModelRandomWalk runs the model over deterministic random inputs
// and checks that every msg kind succeeds often enough for the fuzz target to
// be meaningful rather than rejecting everything.
func TestKeeperModelRandomWalk(t *testing.T) {
	env := getModelEnv()
	rng := rand.New(rand.NewPCG(1, 2))
	total := map[string]int{}
	for range 500 {
		data := make([]byte, 160)
		for i := range data {
			data[i] = byte(rng.Uint32())
		}
		for k, v := range runKeeperModel(t, env, data) {
			total[k] += v
		}
	}
	t.Logf("successful msgs: %v", total)
	for _, kind := range []string{"UpdateParams", "Trip", "Reset"} {
		require.Greater(t, total[kind], 50, "%s rarely succeeds; the model fuzz is too weak", kind)
	}
}
