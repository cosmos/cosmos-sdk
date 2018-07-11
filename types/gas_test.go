package types

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGasMeter(t *testing.T) {
	limits := []Gas{
		10,
		1000,
		100000,
		100000000,
		65535,
		65536,
	}

	for _, limit := range limits {
		meter := NewGasMeter(limit)
		used := int64(0)

		for {
			gas := Gas(rand.Int63n(limit))
			used += gas
			if used > limit {
				require.Panics(t, func() { meter.ConsumeGas(gas, "") })
				break
			}
			require.NotPanics(t, func() { meter.ConsumeGas(gas, "") })
			require.Equal(t, used, meter.GasConsumed())
		}
	}
}
