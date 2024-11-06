package gen_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	. "cosmossdk.io/x/benchmark/generator"
)

func TestGenerator_GenerateKeys(t *testing.T) {
	gen := NewGenerator(Options{
		Seed:        34,
		KeyMean:     64,
		KeyStdDev:   12,
		ValueMean:   1024,
		ValueStdDev: 256,
	})

	for i := 0; i < 1_000_000; i++ {
		key := gen.Key()
		require.NotNil(t, key)
		require.Greater(t, len(key), 0)
		val := gen.Value()
		require.NotNil(t, val)
		require.Greater(t, len(val), 0)
	}
}
