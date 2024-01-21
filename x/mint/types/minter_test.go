package types

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestInflationRate(t *testing.T) {
	t.Parallel()

	type testCase struct {
		year     int
		expected math.LegacyDec
	}

	genesis := time.Date(2024, 01, 20, 21, 51, 30, 00, time.UTC)

	tests := []testCase{
		{year: 0, expected: math.LegacyNewDecWithPrec(8, 2)},                  // 8.00%
		{year: 1, expected: math.LegacyNewDecWithPrec(72, 3)},                 // 7.20%
		{year: 2, expected: math.LegacyNewDecWithPrec(648, 4)},                // 6.48%
		{year: 3, expected: math.LegacyNewDecWithPrec(5832, 5)},               // 5.832%
		{year: 4, expected: math.LegacyNewDecWithPrec(52488, 6)},              // 5.2488%
		{year: 5, expected: math.LegacyNewDecWithPrec(472392, 7)},             // 4.72392%
		{year: 6, expected: math.LegacyNewDecWithPrec(4251528, 8)},            // 4.251528%
		{year: 7, expected: math.LegacyNewDecWithPrec(38263752, 9)},           // 3.8263752%
		{year: 8, expected: math.LegacyNewDecWithPrec(344373768, 10)},         // 3.44373768%
		{year: 9, expected: math.LegacyNewDecWithPrec(3099363912, 11)},        // 3.099363912%
		{year: 10, expected: math.LegacyNewDecWithPrec(27894275208, 12)},      // 2.7894275208%
		{year: 11, expected: math.LegacyNewDecWithPrec(251048476872, 13)},     // 2.51048476872%
		{year: 12, expected: math.LegacyNewDecWithPrec(2259436291848, 14)},    // 2.259436291848%
		{year: 13, expected: math.LegacyNewDecWithPrec(20334926626632, 15)},   // 2.0334926626632%
		{year: 14, expected: math.LegacyNewDecWithPrec(183014339639688, 16)},  // 1.83014339639688%
		{year: 15, expected: math.LegacyNewDecWithPrec(1647129056757192, 17)}, // 1.647129056757192%
		{year: 16, expected: math.LegacyNewDecWithPrec(15, 3)},                // 1.50%
		{year: 17, expected: math.LegacyNewDecWithPrec(15, 3)},                // 1.50%
	}

	for i, tc := range tests {
		tc := tc
		name := fmt.Sprintf("Year %d", tc.year)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			currentBlockTime := genesis.AddDate(tc.year, 0, 0)
			inflation := InflationRate(genesis, currentBlockTime)

			diff := inflation.Sub(tc.expected)

			require.True(t, inflation.Equal(tc.expected), "Test Index: %v\nDiff:  %v\nExpected: %v\n", i, diff, tc.expected)
		})
	}
}

//func TestBlockProvision(t *testing.T) {
//	minter := InitialMinter(math.LegacyNewDecWithPrec(1, 1))
//	params := DefaultParams()
//
//	secondsPerYear := int64(60 * 60 * 8766)
//
//	tests := []struct {
//		annualProvisions int64
//		expProvisions    int64
//	}{
//		{secondsPerYear / 5, 1},
//		{secondsPerYear/5 + 1, 1},
//		{(secondsPerYear / 5) * 2, 2},
//		{(secondsPerYear / 5) / 2, 0},
//	}
//	for i, tc := range tests {
//		minter.AnnualProvisions = math.LegacyNewDec(tc.annualProvisions)
//		provisions := minter.BlockProvision(params)
//
//		expProvisions := sdk.NewCoin(params.MintDenom,
//			math.NewInt(tc.expProvisions))
//
//		require.True(t, expProvisions.IsEqual(provisions),
//			"test: %v\n\tExp: %v\n\tGot: %v\n",
//			i, tc.expProvisions, provisions)
//	}
//}
