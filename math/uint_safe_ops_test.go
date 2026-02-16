package math_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
)

func TestUintSafeAdd(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		a, b    sdkmath.Uint
		want    sdkmath.Uint
		wantErr bool
	}{
		{
			name: "zero plus zero",
			a:    sdkmath.ZeroUint(),
			b:    sdkmath.ZeroUint(),
			want: sdkmath.ZeroUint(),
		},
		{
			name: "zero plus one",
			a:    sdkmath.ZeroUint(),
			b:    sdkmath.OneUint(),
			want: sdkmath.OneUint(),
		},
		{
			name: "normal addition",
			a:    sdkmath.NewUint(100),
			b:    sdkmath.NewUint(200),
			want: sdkmath.NewUint(300),
		},
		{
			name: "large values",
			a:    sdkmath.NewUint(1<<63 - 1),
			b:    sdkmath.NewUint(1<<63 - 1),
			want: sdkmath.NewUintFromBigInt(new(big.Int).Add(new(big.Int).SetUint64(1<<63-1), new(big.Int).SetUint64(1<<63-1))),
		},
		{
			name:    "overflow beyond 256 bits",
			a:       sdkmath.NewUintFromBigInt(new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1))),
			b:       sdkmath.OneUint(),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			res, err := tc.a.SafeAdd(tc.b)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "overflow")
			} else {
				require.NoError(t, err)
				require.True(t, res.Equal(tc.want), "expected %s, got %s", tc.want, res)
			}
		})
	}
}

func TestUintSafeSub(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		a, b    sdkmath.Uint
		want    sdkmath.Uint
		wantErr bool
	}{
		{
			name: "zero minus zero",
			a:    sdkmath.ZeroUint(),
			b:    sdkmath.ZeroUint(),
			want: sdkmath.ZeroUint(),
		},
		{
			name: "normal subtraction",
			a:    sdkmath.NewUint(500),
			b:    sdkmath.NewUint(200),
			want: sdkmath.NewUint(300),
		},
		{
			name:    "underflow",
			a:       sdkmath.NewUint(5),
			b:       sdkmath.NewUint(10),
			wantErr: true,
		},
		{
			name:    "zero minus one",
			a:       sdkmath.ZeroUint(),
			b:       sdkmath.OneUint(),
			wantErr: true,
		},
		{
			name: "equal values",
			a:    sdkmath.NewUint(42),
			b:    sdkmath.NewUint(42),
			want: sdkmath.ZeroUint(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			res, err := tc.a.SafeSub(tc.b)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "underflow")
			} else {
				require.NoError(t, err)
				require.True(t, res.Equal(tc.want), "expected %s, got %s", tc.want, res)
			}
		})
	}
}

func TestUintSafeMul(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		a, b    sdkmath.Uint
		want    sdkmath.Uint
		wantErr bool
	}{
		{
			name: "zero times anything",
			a:    sdkmath.ZeroUint(),
			b:    sdkmath.NewUint(999),
			want: sdkmath.ZeroUint(),
		},
		{
			name: "one times value",
			a:    sdkmath.OneUint(),
			b:    sdkmath.NewUint(12345),
			want: sdkmath.NewUint(12345),
		},
		{
			name: "normal multiplication",
			a:    sdkmath.NewUint(100),
			b:    sdkmath.NewUint(200),
			want: sdkmath.NewUint(20000),
		},
		{
			name:    "overflow beyond 256 bits",
			a:       sdkmath.NewUintFromBigInt(new(big.Int).Exp(big.NewInt(2), big.NewInt(200), nil)),
			b:       sdkmath.NewUintFromBigInt(new(big.Int).Exp(big.NewInt(2), big.NewInt(200), nil)),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			res, err := tc.a.SafeMul(tc.b)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "overflow")
			} else {
				require.NoError(t, err)
				require.True(t, res.Equal(tc.want), "expected %s, got %s", tc.want, res)
			}
		})
	}
}

func TestUintSafeQuo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		a, b    sdkmath.Uint
		want    sdkmath.Uint
		wantErr bool
	}{
		{
			name:    "division by zero",
			a:       sdkmath.NewUint(100),
			b:       sdkmath.ZeroUint(),
			wantErr: true,
		},
		{
			name: "normal division",
			a:    sdkmath.NewUint(100),
			b:    sdkmath.NewUint(3),
			want: sdkmath.NewUint(33),
		},
		{
			name: "exact division",
			a:    sdkmath.NewUint(100),
			b:    sdkmath.NewUint(10),
			want: sdkmath.NewUint(10),
		},
		{
			name: "zero divided by non-zero",
			a:    sdkmath.ZeroUint(),
			b:    sdkmath.NewUint(5),
			want: sdkmath.ZeroUint(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			res, err := tc.a.SafeQuo(tc.b)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "division by zero")
			} else {
				require.NoError(t, err)
				require.True(t, res.Equal(tc.want), "expected %s, got %s", tc.want, res)
			}
		})
	}
}

func TestUintSafeMod(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		a, b    sdkmath.Uint
		want    sdkmath.Uint
		wantErr bool
	}{
		{
			name:    "mod by zero",
			a:       sdkmath.NewUint(100),
			b:       sdkmath.ZeroUint(),
			wantErr: true,
		},
		{
			name: "normal mod",
			a:    sdkmath.NewUint(10),
			b:    sdkmath.NewUint(3),
			want: sdkmath.OneUint(),
		},
		{
			name: "exact divisor",
			a:    sdkmath.NewUint(10),
			b:    sdkmath.NewUint(5),
			want: sdkmath.ZeroUint(),
		},
		{
			name: "mod of zero",
			a:    sdkmath.ZeroUint(),
			b:    sdkmath.NewUint(7),
			want: sdkmath.ZeroUint(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			res, err := tc.a.SafeMod(tc.b)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "division by zero")
			} else {
				require.NoError(t, err)
				require.True(t, res.Equal(tc.want), "expected %s, got %s", tc.want, res)
			}
		})
	}
}

// TestUintSafeMethodsDoNotMutate verifies that safe methods do not mutate the receiver.
func TestUintSafeMethodsDoNotMutate(t *testing.T) {
	t.Parallel()

	original := sdkmath.NewUint(100)
	other := sdkmath.NewUint(50)

	// Store original value
	origVal := original.Uint64()

	_, _ = original.SafeAdd(other)
	require.Equal(t, origVal, original.Uint64(), "SafeAdd mutated receiver")

	_, _ = original.SafeSub(other)
	require.Equal(t, origVal, original.Uint64(), "SafeSub mutated receiver")

	_, _ = original.SafeMul(other)
	require.Equal(t, origVal, original.Uint64(), "SafeMul mutated receiver")

	_, _ = original.SafeQuo(other)
	require.Equal(t, origVal, original.Uint64(), "SafeQuo mutated receiver")

	_, _ = original.SafeMod(other)
	require.Equal(t, origVal, original.Uint64(), "SafeMod mutated receiver")
}

// TestUintAddPanicsOnOverflow ensures Add still panics on overflow (backwards compat).
func TestUintAddPanicsOnOverflow(t *testing.T) {
	t.Parallel()
	uintMax := sdkmath.NewUintFromBigInt(new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1)))
	require.Panics(t, func() { uintMax.Add(sdkmath.OneUint()) })
}

// TestUintSubPanicsOnUnderflow ensures Sub still panics on underflow (backwards compat).
func TestUintSubPanicsOnUnderflow(t *testing.T) {
	t.Parallel()
	require.Panics(t, func() { sdkmath.ZeroUint().Sub(sdkmath.OneUint()) })
}

// TestUintMulPanicsOnOverflow ensures Mul still panics on overflow (backwards compat).
func TestUintMulPanicsOnOverflow(t *testing.T) {
	t.Parallel()
	large := sdkmath.NewUintFromBigInt(new(big.Int).Exp(big.NewInt(2), big.NewInt(200), nil))
	require.Panics(t, func() { large.Mul(large) })
}

// TestUintQuoPanicsOnDivByZero ensures Quo still panics on div-by-zero (backwards compat).
func TestUintQuoPanicsOnDivByZero(t *testing.T) {
	t.Parallel()
	require.Panics(t, func() { sdkmath.OneUint().Quo(sdkmath.ZeroUint()) })
}

// TestUintModPanicsOnDivByZero ensures Mod still panics on div-by-zero (backwards compat).
func TestUintModPanicsOnDivByZero(t *testing.T) {
	t.Parallel()
	require.Panics(t, func() { sdkmath.OneUint().Mod(sdkmath.ZeroUint()) })
}
