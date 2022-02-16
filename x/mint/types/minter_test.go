package types

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestNextInflation(t *testing.T) {
	minter := DefaultInitialMinter()
	params := DefaultParams()
	blocksPerYr := sdk.NewDec(int64(params.BlocksPerYear))

	// Governing Mechanism:
	//    inflationRateChangePerYear = (1- BondedRatio/ GoalBonded) * MaxInflationRateChange

	var stableSupply = EndHyperInflation
	var dec50m, _ = sdk.NewDecFromStr("0.135335283236612691")
	var dec150m, _ = sdk.NewDecFromStr("1.000000000000000000")
	var dec200m, _ = sdk.NewDecFromStr("0.606530659712633423")
	tests := []struct {
		bondedRatio, setInflation, expChange sdk.Dec
		totalSupply                          sdk.Int
	}{
		// with 0% bonded atom supply the inflation should increase by InflationRateChange
		{sdk.ZeroDec(), sdk.NewDecWithPrec(7, 2), params.InflationRateChange.Quo(blocksPerYr), stableSupply},

		// 100% bonded, starting at 20% inflation and being reduced
		// (1 - (1/0.67))*(0.13/8667)
		{sdk.OneDec(), sdk.NewDecWithPrec(20, 2),
			sdk.OneDec().Sub(sdk.OneDec().Quo(params.GoalBonded)).Mul(params.InflationRateChange).Quo(blocksPerYr), stableSupply},

		// 50% bonded, starting at 10% inflation and being increased
		{sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(10, 2),
			sdk.OneDec().Sub(sdk.NewDecWithPrec(5, 1).Quo(params.GoalBonded)).Mul(params.InflationRateChange).Quo(blocksPerYr), stableSupply},

		// test 7% minimum stop (testing with 100% bonded)
		{sdk.OneDec(), sdk.NewDecWithPrec(7, 2), sdk.ZeroDec(), stableSupply},
		{sdk.OneDec(), sdk.NewDecWithPrec(700000001, 10), sdk.NewDecWithPrec(-1, 10), stableSupply},

		// test 20% maximum stop (testing with 0% bonded)
		{sdk.ZeroDec(), sdk.NewDecWithPrec(20, 2), sdk.ZeroDec(), stableSupply},
		{sdk.ZeroDec(), sdk.NewDecWithPrec(1999999999, 10), sdk.NewDecWithPrec(1, 10), stableSupply},

		// perfect balance shouldn't change inflation
		{sdk.NewDecWithPrec(67, 2), sdk.NewDecWithPrec(15, 2), sdk.ZeroDec(), stableSupply},

		// test hyperinflationary regime
		{sdk.ZeroDec(), dec50m, sdk.ZeroDec(), sdk.NewIntWithDecimal(50_000_000, 18)},
		{sdk.ZeroDec(), dec150m, sdk.ZeroDec(), sdk.NewIntWithDecimal(150_000_000, 18)},
		{sdk.ZeroDec(), dec200m, sdk.ZeroDec(), sdk.NewIntWithDecimal(200_000_000, 18)},
	}
	for i, tc := range tests {
		minter.Inflation = tc.setInflation

		inflation := minter.NextInflationRate(params, tc.bondedRatio, tc.totalSupply)
		diffInflation := inflation.Sub(tc.setInflation)

		require.True(t, diffInflation.Equal(tc.expChange),
			"Test Index: %v\nDiff:  %v\nExpected: %v\n", i, diffInflation, tc.expChange)
	}
}

func TestBlockProvision(t *testing.T) {
	minter := InitialMinter(sdk.NewDecWithPrec(1, 1))
	params := DefaultParams()

	secondsPerYear := int64(60 * 60 * 8766)

	tests := []struct {
		annualProvisions int64
		expProvisions    int64
	}{
		{secondsPerYear / 5, 1},
		{secondsPerYear/5 + 1, 1},
		{(secondsPerYear / 5) * 2, 2},
		{(secondsPerYear / 5) / 2, 0},
	}
	for i, tc := range tests {
		minter.AnnualProvisions = sdk.NewDec(tc.annualProvisions)
		provisions := minter.BlockProvision(params)

		expProvisions := sdk.NewCoin(params.MintDenom,
			sdk.NewInt(tc.expProvisions))

		require.True(t, expProvisions.IsEqual(provisions),
			"test: %v\n\tExp: %v\n\tGot: %v\n",
			i, tc.expProvisions, provisions)
	}
}

// Benchmarking :)
// previously using sdk.Int operations:
// BenchmarkBlockProvision-4 5000000 220 ns/op
//
// using sdk.Dec operations: (current implementation)
// BenchmarkBlockProvision-4 3000000 429 ns/op
func BenchmarkBlockProvision(b *testing.B) {
	b.ReportAllocs()
	minter := InitialMinter(sdk.NewDecWithPrec(1, 1))
	params := DefaultParams()

	s1 := rand.NewSource(100)
	r1 := rand.New(s1)
	minter.AnnualProvisions = sdk.NewDec(r1.Int63n(1000000))

	// run the BlockProvision function b.N times
	for n := 0; n < b.N; n++ {
		minter.BlockProvision(params)
	}
}

// Next inflation benchmarking
// BenchmarkNextInflation-4 1000000 1828 ns/op
func BenchmarkNextInflation(b *testing.B) {
	b.ReportAllocs()
	minter := InitialMinter(sdk.NewDecWithPrec(1, 1))
	params := DefaultParams()
	bondedRatio := sdk.NewDecWithPrec(1, 1)

	// run the NextInflationRate function b.N times
	for n := 0; n < b.N; n++ {
		// the hyperinflationary regime is by far the most expensive to calculate
		minter.NextInflationRate(params, bondedRatio, sdk.NewInt(100_000_000))
	}
}

// Next annual provisions benchmarking
// BenchmarkNextAnnualProvisions-4 5000000 251 ns/op
func BenchmarkNextAnnualProvisions(b *testing.B) {
	b.ReportAllocs()
	minter := InitialMinter(sdk.NewDecWithPrec(1, 1))
	params := DefaultParams()
	totalSupply := sdk.NewInt(100000000000000)

	// run the NextAnnualProvisions function b.N times
	for n := 0; n < b.N; n++ {
		minter.NextAnnualProvisions(params, totalSupply)
	}
}
