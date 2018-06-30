package types

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFromInt64(t *testing.T) {
	for n := 0; n < 20; n++ {
		r := rand.Int63()
		require.Equal(t, r, NewInt(r).Int64())
	}
}

func TestInt(t *testing.T) {
	// Max Int = 2^255-1 = 5.789e+76
	// Min Int = -(2^255-1) = -5.789e+76
	require.NotPanics(t, func() { NewIntWithDecimal(1, 76) })
	i1 := NewIntWithDecimal(1, 76)
	require.NotPanics(t, func() { NewIntWithDecimal(2, 76) })
	i2 := NewIntWithDecimal(2, 76)
	require.NotPanics(t, func() { NewIntWithDecimal(3, 76) })
	i3 := NewIntWithDecimal(3, 76)

	require.Panics(t, func() { NewIntWithDecimal(6, 76) })
	require.Panics(t, func() { NewIntWithDecimal(9, 80) })

	// Overflow check
	require.NotPanics(t, func() { i1.Add(i1) })
	require.NotPanics(t, func() { i2.Add(i2) })
	require.Panics(t, func() { i3.Add(i3) })

	require.NotPanics(t, func() { i1.Sub(i1.Neg()) })
	require.NotPanics(t, func() { i2.Sub(i2.Neg()) })
	require.Panics(t, func() { i3.Sub(i3.Neg()) })

	require.Panics(t, func() { i1.Mul(i1) })
	require.Panics(t, func() { i2.Mul(i2) })
	require.Panics(t, func() { i3.Mul(i3) })

	require.Panics(t, func() { i1.Neg().Mul(i1.Neg()) })
	require.Panics(t, func() { i2.Neg().Mul(i2.Neg()) })
	require.Panics(t, func() { i3.Neg().Mul(i3.Neg()) })

	// Underflow check
	i3n := i3.Neg()
	require.NotPanics(t, func() { i3n.Sub(i1) })
	require.NotPanics(t, func() { i3n.Sub(i2) })
	require.Panics(t, func() { i3n.Sub(i3) })

	require.NotPanics(t, func() { i3n.Add(i1.Neg()) })
	require.NotPanics(t, func() { i3n.Add(i2.Neg()) })
	require.Panics(t, func() { i3n.Add(i3.Neg()) })

	require.Panics(t, func() { i1.Mul(i1.Neg()) })
	require.Panics(t, func() { i2.Mul(i2.Neg()) })
	require.Panics(t, func() { i3.Mul(i3.Neg()) })

	// Bound check
	intmax := NewIntFromBigInt(new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(255), nil), big.NewInt(1)))
	intmin := intmax.Neg()
	require.NotPanics(t, func() { intmax.Add(ZeroInt()) })
	require.NotPanics(t, func() { intmin.Sub(ZeroInt()) })
	require.Panics(t, func() { intmax.Add(OneInt()) })
	require.Panics(t, func() { intmin.Sub(OneInt()) })

	// Division-by-zero check
	require.Panics(t, func() { i1.Div(NewInt(0)) })
}

func TestUint(t *testing.T) {
	// Max Uint = 1.15e+77
	// Min Uint = 0
	require.NotPanics(t, func() { NewUintWithDecimal(5, 76) })
	i1 := NewUintWithDecimal(5, 76)
	require.NotPanics(t, func() { NewUintWithDecimal(10, 76) })
	i2 := NewUintWithDecimal(10, 76)
	require.NotPanics(t, func() { NewUintWithDecimal(11, 76) })
	i3 := NewUintWithDecimal(11, 76)

	require.Panics(t, func() { NewUintWithDecimal(12, 76) })
	require.Panics(t, func() { NewUintWithDecimal(1, 80) })

	// Overflow check
	require.NotPanics(t, func() { i1.Add(i1) })
	require.Panics(t, func() { i2.Add(i2) })
	require.Panics(t, func() { i3.Add(i3) })

	require.Panics(t, func() { i1.Mul(i1) })
	require.Panics(t, func() { i2.Mul(i2) })
	require.Panics(t, func() { i3.Mul(i3) })

	// Underflow check
	require.NotPanics(t, func() { i2.Sub(i1) })
	require.NotPanics(t, func() { i2.Sub(i2) })
	require.Panics(t, func() { i2.Sub(i3) })

	// Bound check
	uintmax := NewUintFromBigInt(new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1)))
	uintmin := NewUint(0)
	require.NotPanics(t, func() { uintmax.Add(ZeroUint()) })
	require.NotPanics(t, func() { uintmin.Sub(ZeroUint()) })
	require.Panics(t, func() { uintmax.Add(OneUint()) })
	require.Panics(t, func() { uintmin.Sub(OneUint()) })

	// Division-by-zero check
	require.Panics(t, func() { i1.Div(uintmin) })
}
