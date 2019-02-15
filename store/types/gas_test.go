package types

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGasMeter(t *testing.T) {
	cases := []struct {
		limit Gas
		usage []Gas
	}{
		{10, []Gas{1, 2, 3, 4}},
		{1000, []Gas{40, 30, 20, 10, 900}},
		{100000, []Gas{99999, 1}},
		{100000000, []Gas{50000000, 40000000, 10000000}},
		{65535, []Gas{32768, 32767}},
		{65536, []Gas{32768, 32767, 1}},
	}

	for tcnum, tc := range cases {
		meter := NewGasMeter(tc.limit)
		used := uint64(0)

		for unum, usage := range tc.usage {
			used += usage
			require.NotPanics(t, func() { meter.ConsumeGas(usage, "") }, "Not exceeded limit but panicked. tc #%d, usage #%d", tcnum, unum)
			require.Equal(t, used, meter.GasConsumed(), "Gas consumption not match. tc #%d, usage #%d", tcnum, unum)
			require.Equal(t, used, meter.GasConsumedToLimit(), "Gas consumption (to limit) not match. tc #%d, usage #%d", tcnum, unum)
			require.False(t, meter.IsPastLimit(), "Not exceeded limit but got IsPastLimit() true")
			if unum < len(tc.usage)-1 {
				require.False(t, meter.IsOutOfGas(), "Not yet at limit but got IsOutOfGas() true")
			} else {
				require.True(t, meter.IsOutOfGas(), "At limit but got IsOutOfGas() false")
			}
		}

		require.Panics(t, func() { meter.ConsumeGas(1, "") }, "Exceeded but not panicked. tc #%d", tcnum)
		require.Equal(t, meter.GasConsumedToLimit(), meter.Limit(), "Gas consumption (to limit) not match limit")
		require.Equal(t, meter.GasConsumed(), meter.Limit()+1, "Gas consumption not match limit+1")
		break

	}
}

func TestAddUint64Overflow(t *testing.T) {
	testCases := []struct {
		a, b     uint64
		result   uint64
		overflow bool
	}{
		{0, 0, 0, false},
		{100, 100, 200, false},
		{math.MaxUint64 / 2, math.MaxUint64/2 + 1, math.MaxUint64, false},
		{math.MaxUint64 / 2, math.MaxUint64/2 + 2, 0, true},
	}

	for i, tc := range testCases {
		res, overflow := addUint64Overflow(tc.a, tc.b)
		require.Equal(
			t, tc.overflow, overflow,
			"invalid overflow result; tc: #%d, a: %d, b: %d", i, tc.a, tc.b,
		)
		require.Equal(
			t, tc.result, res,
			"invalid uint64 result; tc: #%d, a: %d, b: %d", i, tc.a, tc.b,
		)
	}
}
