package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProxyGasMeter(t *testing.T) {
	baseGas := uint64(1000)
	limit := uint64(300)

	bgm := NewGasMeter(baseGas)
	pgm := NewProxyGasMeter(bgm, limit)

	require.Equal(t, Gas(0), pgm.GasConsumed())
	require.Equal(t, limit, pgm.Limit())
	require.Equal(t, limit, pgm.GasRemaining())

	pgm.ConsumeGas(100, "test")
	require.Equal(t, Gas(100), pgm.GasConsumed())
	require.Equal(t, Gas(100), bgm.GasConsumed())
	require.Equal(t, limit-100, pgm.GasRemaining())
	require.False(t, pgm.IsOutOfGas())
	require.False(t, pgm.IsPastLimit())

	pgm.ConsumeGas(200, "test")
	require.Equal(t, Gas(300), pgm.GasConsumed())
	require.Equal(t, Gas(300), bgm.GasConsumed())
	require.Equal(t, Gas(0), pgm.GasRemaining())
	require.Equal(t, Gas(700), bgm.GasRemaining())
	require.True(t, pgm.IsOutOfGas())
	require.False(t, pgm.IsPastLimit())

	require.Panics(t, func() {
		pgm.ConsumeGas(1, "test")
	})
	require.Equal(t, Gas(699), bgm.GasRemaining())

	pgm.RefundGas(1, "test")
	require.Equal(t, Gas(300), pgm.GasConsumed())
	require.Equal(t, Gas(0), pgm.GasRemaining())
	require.True(t, pgm.IsOutOfGas())
	require.False(t, pgm.IsPastLimit())
}
