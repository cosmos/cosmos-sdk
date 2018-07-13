package types

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

//changing the int in NewSource will allow you to test different, deterministic, sets of operations
var r = rand.New(rand.NewSource(6595))

func TestGetInflation(t *testing.T) {
	pool := InitialPool()
	params := DefaultParams()

	// Governing Mechanism:
	//    BondedRatio = BondedTokens / TotalSupply
	//    inflationRateChangePerYear = (1- BondedRatio/ GoalBonded) * MaxInflationRateChange

	tests := []struct {
		name string
		setBondedTokens, setLooseTokens,
		setInflation, expectedChange sdk.Rat
	}{
		// with 0% bonded atom supply the inflation should increase by InflationRateChange
		{"test 1", sdk.ZeroRat(), sdk.ZeroRat(), sdk.NewRat(7, 100), params.InflationRateChange.Quo(hrsPerYrRat).Round(precision)},

		// 100% bonded, starting at 20% inflation and being reduced
		// (1 - (1/0.67))*(0.13/8667)
		{"test 2", sdk.OneRat(), sdk.ZeroRat(), sdk.NewRat(20, 100),
			sdk.OneRat().Sub(sdk.OneRat().Quo(params.GoalBonded)).Mul(params.InflationRateChange).Quo(hrsPerYrRat).Round(precision)},

		// 50% bonded, starting at 10% inflation and being increased
		{"test 3", sdk.OneRat(), sdk.OneRat(), sdk.NewRat(10, 100),
			sdk.OneRat().Sub(sdk.NewRat(1, 2).Quo(params.GoalBonded)).Mul(params.InflationRateChange).Quo(hrsPerYrRat).Round(precision)},

		// test 7% minimum stop (testing with 100% bonded)
		{"test 4", sdk.OneRat(), sdk.ZeroRat(), sdk.NewRat(7, 100), sdk.ZeroRat()},
		{"test 5", sdk.OneRat(), sdk.ZeroRat(), sdk.NewRat(70001, 1000000), sdk.NewRat(-1, 1000000).Round(precision)},

		// test 20% maximum stop (testing with 0% bonded)
		{"test 6", sdk.ZeroRat(), sdk.ZeroRat(), sdk.NewRat(20, 100), sdk.ZeroRat()},
		{"test 7", sdk.ZeroRat(), sdk.ZeroRat(), sdk.NewRat(199999, 1000000), sdk.NewRat(1, 1000000).Round(precision)},

		// perfect balance shouldn't change inflation
		{"test 8", sdk.NewRat(67), sdk.NewRat(33), sdk.NewRat(15, 100), sdk.ZeroRat()},
	}
	for _, tc := range tests {
		pool.BondedTokens, pool.LooseTokens = tc.setBondedTokens, tc.setLooseTokens
		pool.Inflation = tc.setInflation

		inflation := pool.NextInflation(params)
		diffInflation := inflation.Sub(tc.setInflation)

		require.True(t, diffInflation.Equal(tc.expectedChange),
			"Name: %v\nDiff:  %v\nExpected: %v\n", tc.name, diffInflation, tc.expectedChange)
	}
}

// Test that provisions are correctly added to the pool and validators each hour for 1 year
func TestProcessProvisions(t *testing.T) {
	pool := InitialPool()
	params := DefaultParams()

	var (
		initialTotalTokens int64 = 550000000
		cumulativeExpProvs       = sdk.ZeroRat()
	)
	pool.LooseTokens = sdk.NewRat(initialTotalTokens)

	// process the provisions for a year
	for hr := 0; hr < 100; hr++ {
		var expProvisions sdk.Rat
		_, expProvisions, pool = updateProvisions(t, pool, params, hr)
		cumulativeExpProvs = cumulativeExpProvs.Add(expProvisions)
	}

	//get the pool and do the final value checks from checkFinalPoolValues
	checkFinalPoolValues(t, pool, sdk.NewRat(initialTotalTokens), cumulativeExpProvs)
}

//_________________________________________________________________________________________
////////////////////////////////HELPER FUNCTIONS BELOW/////////////////////////////////////

// Final check on the global pool values for what the total tokens accumulated from each hour of provisions
func checkFinalPoolValues(t *testing.T, pool Pool, initialTotalTokens, cumulativeExpProvs sdk.Rat) {
	calculatedTotalTokens := initialTotalTokens.Add(cumulativeExpProvs)
	require.True(sdk.RatEq(t, calculatedTotalTokens, pool.TokenSupply()))
}

// Processes provisions are added to the pool correctly every hour
// Returns expected Provisions, expected Inflation, and pool, to help with cumulative calculations back in main Tests
func updateProvisions(t *testing.T, pool Pool, params Params, hr int) (sdk.Rat, sdk.Rat, Pool) {
	expInflation := pool.NextInflation(params)
	expProvisions := expInflation.Mul(pool.TokenSupply()).Quo(hrsPerYrRat)
	startTotalSupply := pool.TokenSupply()
	pool = pool.ProcessProvisions(params)

	//check provisions were added to pool
	require.True(sdk.RatEq(t, startTotalSupply.Add(expProvisions), pool.TokenSupply()))

	return expInflation, expProvisions, pool
}

// Checks that The inflation will correctly increase or decrease after an update to the pool
// nolint: gocyclo
func checkInflation(t *testing.T, pool Pool, previousInflation, updatedInflation sdk.Rat, msg string) {
	inflationChange := updatedInflation.Sub(previousInflation)

	switch {
	//BELOW 67% - Rate of change positive and increasing, while we are between 7% <= and < 20% inflation
	case pool.BondedRatio().LT(sdk.NewRat(67, 100)) && updatedInflation.LT(sdk.NewRat(20, 100)):
		require.Equal(t, true, inflationChange.GT(sdk.ZeroRat()), msg)

	//BELOW 67% - Rate of change should be 0 while inflation continually stays at 20% until we reach 67% bonded ratio
	case pool.BondedRatio().LT(sdk.NewRat(67, 100)) && updatedInflation.Equal(sdk.NewRat(20, 100)):
		if previousInflation.Equal(sdk.NewRat(20, 100)) {
			require.Equal(t, true, inflationChange.IsZero(), msg)

			//This else statement covers the one off case where we first hit 20%, but we still needed a positive ROC to get to 67% bonded ratio (i.e. we went from 19.99999% to 20%)
		} else {
			require.Equal(t, true, inflationChange.GT(sdk.ZeroRat()), msg)
		}

	//ABOVE 67% - Rate of change should be negative while the bond is above 67, and should stay negative until we reach inflation of 7%
	case pool.BondedRatio().GT(sdk.NewRat(67, 100)) && updatedInflation.LT(sdk.NewRat(20, 100)) && updatedInflation.GT(sdk.NewRat(7, 100)):
		require.Equal(t, true, inflationChange.LT(sdk.ZeroRat()), msg)

	//ABOVE 67% - Rate of change should be 0 while inflation continually stays at 7%.
	case pool.BondedRatio().GT(sdk.NewRat(67, 100)) && updatedInflation.Equal(sdk.NewRat(7, 100)):
		if previousInflation.Equal(sdk.NewRat(7, 100)) {
			require.Equal(t, true, inflationChange.IsZero(), msg)

			//This else statement covers the one off case where we first hit 7%, but we still needed a negative ROC to continue to get down to 67%. (i.e. we went from 7.00001% to 7%)
		} else {
			require.Equal(t, true, inflationChange.LT(sdk.ZeroRat()), msg)
		}
	}
}
