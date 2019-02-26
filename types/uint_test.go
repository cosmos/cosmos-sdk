package types

import (
	"math"
	"math/big"
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUintPanics(t *testing.T) {
	// Max Uint = 1.15e+77
	// Min Uint = 0
	u1 := NewUint(0)
	u2 := OneUint()

	require.Equal(t, uint64(0), u1.Uint64())
	require.Equal(t, uint64(1), u2.Uint64())

	require.Panics(t, func() { NewUintFromBigInt(big.NewInt(-5)) })
	require.Panics(t, func() { NewUintFromString("-1") })
	require.NotPanics(t, func() {
		require.True(t, NewUintFromString("0").Equal(ZeroUint()))
		require.True(t, NewUintFromString("5").Equal(NewUint(5)))
	})

	// Overflow check
	require.True(t, u1.Add(u1).Equal(ZeroUint()))
	require.True(t, u1.Add(OneUint()).Equal(OneUint()))
	require.Equal(t, uint64(0), u1.Uint64())
	require.Equal(t, uint64(1), OneUint().Uint64())
	require.Panics(t, func() { u1.SubUint64(2) })
	require.True(t, u1.SubUint64(0).Equal(ZeroUint()))
	require.True(t, u2.Add(OneUint()).Sub(OneUint()).Equal(OneUint()))    // i2 == 1
	require.True(t, u2.Add(OneUint()).Mul(NewUint(5)).Equal(NewUint(10))) // i2 == 10
	require.True(t, NewUint(7).Quo(NewUint(2)).Equal(NewUint(3)))
	require.True(t, NewUint(0).Quo(NewUint(2)).Equal(ZeroUint()))
	require.True(t, NewUint(5).MulUint64(4).Equal(NewUint(20)))
	require.True(t, NewUint(5).MulUint64(0).Equal(ZeroUint()))

	uintmax := NewUintFromBigInt(new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1)))
	uintmin := ZeroUint()

	// divs by zero
	require.Panics(t, func() { OneUint().Mul(ZeroUint().SubUint64(uint64(1))) })
	require.Panics(t, func() { OneUint().QuoUint64(0) })
	require.Panics(t, func() { OneUint().Quo(ZeroUint()) })
	require.Panics(t, func() { ZeroUint().QuoUint64(0) })
	require.Panics(t, func() { OneUint().Quo(ZeroUint().Sub(OneUint())) })
	require.Panics(t, func() { uintmax.Add(OneUint()) })
	require.Panics(t, func() { uintmin.Sub(OneUint()) })

	require.Equal(t, uint64(0), MinUint(ZeroUint(), OneUint()).Uint64())
	require.Equal(t, uint64(1), MaxUint(ZeroUint(), OneUint()).Uint64())

	// comparison ops
	require.True(t,
		OneUint().GT(ZeroUint()),
	)
	require.False(t,
		OneUint().LT(ZeroUint()),
	)
	require.True(t,
		OneUint().GTE(ZeroUint()),
	)
	require.False(t,
		OneUint().LTE(ZeroUint()),
	)

	require.False(t, ZeroUint().GT(OneUint()))
	require.True(t, ZeroUint().LT(OneUint()))
	require.False(t, ZeroUint().GTE(OneUint()))
	require.True(t, ZeroUint().LTE(OneUint()))
}

func TestIdentUint(t *testing.T) {
	for d := 0; d < 1000; d++ {
		n := rand.Uint64()
		i := NewUint(n)

		ifromstr := NewUintFromString(strconv.FormatUint(n, 10))

		cases := []uint64{
			i.Uint64(),
			i.i.Uint64(),
			ifromstr.Uint64(),
			NewUintFromBigInt(new(big.Int).SetUint64(n)).Uint64(),
		}

		for tcnum, tc := range cases {
			require.Equal(t, n, tc, "Uint is modified during conversion. tc #%d", tcnum)
		}
	}
}

func TestArithUint(t *testing.T) {
	for d := 0; d < 1000; d++ {
		n1 := uint64(rand.Uint32())
		u1 := NewUint(n1)
		n2 := uint64(rand.Uint32())
		u2 := NewUint(n2)

		cases := []struct {
			ures Uint
			nres uint64
		}{
			{u1.Add(u2), n1 + n2},
			{u1.Mul(u2), n1 * n2},
			{u1.Quo(u2), n1 / n2},
			{u1.AddUint64(n2), n1 + n2},
			{u1.MulUint64(n2), n1 * n2},
			{u1.QuoUint64(n2), n1 / n2},
			{MinUint(u1, u2), minuint(n1, n2)},
			{MaxUint(u1, u2), maxuint(n1, n2)},
		}

		for tcnum, tc := range cases {
			require.Equal(t, tc.nres, tc.ures.Uint64(), "Uint arithmetic operation does not match with uint64 operation. tc #%d", tcnum)
		}

		if n2 > n1 {
			n1, n2 = n2, n1
			u1, u2 = NewUint(n1), NewUint(n2)
		}

		subs := []struct {
			ures Uint
			nres uint64
		}{
			{u1.Sub(u2), n1 - n2},
			{u1.SubUint64(n2), n1 - n2},
		}

		for tcnum, tc := range subs {
			require.Equal(t, tc.nres, tc.ures.Uint64(), "Uint subtraction does not match with uint64 operation. tc #%d", tcnum)
		}
	}
}

func TestCompUint(t *testing.T) {
	for d := 0; d < 1000; d++ {
		n1 := rand.Uint64()
		i1 := NewUint(n1)
		n2 := rand.Uint64()
		i2 := NewUint(n2)

		cases := []struct {
			ires bool
			nres bool
		}{
			{i1.Equal(i2), n1 == n2},
			{i1.GT(i2), n1 > n2},
			{i1.LT(i2), n1 < n2},
			{i1.GTE(i2), !i1.LT(i2)},
			{!i1.GTE(i2), i1.LT(i2)},
		}

		for tcnum, tc := range cases {
			require.Equal(t, tc.nres, tc.ires, "Uint comparison operation does not match with uint64 operation. tc #%d", tcnum)
		}
	}
}

func TestImmutabilityAllUint(t *testing.T) {
	ops := []func(*Uint){
		func(i *Uint) { _ = i.Add(NewUint(rand.Uint64())) },
		func(i *Uint) { _ = i.Sub(NewUint(rand.Uint64() % i.Uint64())) },
		func(i *Uint) { _ = i.Mul(randuint()) },
		func(i *Uint) { _ = i.Quo(randuint()) },
		func(i *Uint) { _ = i.AddUint64(rand.Uint64()) },
		func(i *Uint) { _ = i.SubUint64(rand.Uint64() % i.Uint64()) },
		func(i *Uint) { _ = i.MulUint64(rand.Uint64()) },
		func(i *Uint) { _ = i.QuoUint64(rand.Uint64()) },
		func(i *Uint) { _ = i.IsZero() },
		func(i *Uint) { _ = i.Equal(randuint()) },
		func(i *Uint) { _ = i.GT(randuint()) },
		func(i *Uint) { _ = i.GTE(randuint()) },
		func(i *Uint) { _ = i.LT(randuint()) },
		func(i *Uint) { _ = i.LTE(randuint()) },
		func(i *Uint) { _ = i.String() },
	}

	for i := 0; i < 1000; i++ {
		n := rand.Uint64()
		ni := NewUint(n)

		for opnum, op := range ops {
			op(&ni)

			require.Equal(t, n, ni.Uint64(), "Uint is modified by operation. #%d", opnum)
			require.Equal(t, NewUint(n), ni, "Uint is modified by operation. #%d", opnum)
		}
	}
}

func TestSafeSub(t *testing.T) {
	testCases := []struct {
		x, y     Uint
		expected uint64
		panic    bool
	}{
		{NewUint(0), NewUint(0), 0, false},
		{NewUint(10), NewUint(5), 5, false},
		{NewUint(5), NewUint(10), 5, true},
		{NewUint(math.MaxUint64), NewUint(0), math.MaxUint64, false},
	}

	for i, tc := range testCases {
		if tc.panic {
			require.Panics(t, func() { tc.x.Sub(tc.y) })
			continue
		}
		require.Equal(
			t, tc.expected, tc.x.Sub(tc.y).Uint64(),
			"invalid subtraction result; x: %s, y: %s, tc: #%d", tc.x, tc.y, i,
		)
	}
}

func TestParseUint(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    Uint
		wantErr bool
	}{
		{"malformed", args{"malformed"}, Uint{}, true},
		{"empty", args{""}, Uint{}, true},
		{"positive", args{"50"}, NewUint(uint64(50)), false},
		{"negative", args{"-1"}, Uint{}, true},
		{"zero", args{"0"}, ZeroUint(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseUint(tt.args.s)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.True(t, got.Equal(tt.want))
		})
	}
}

func randuint() Uint {
	return NewUint(rand.Uint64())
}
