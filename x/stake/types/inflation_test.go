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
		setInflation, expectedChange sdk.Dec
	}{
		// with 0% bonded atom supply the inflation should increase by InflationRateChange
		{"test 1", sdk.ZeroDec(), sdk.ZeroDec(), sdk.NewDecWithPrec(7, 2), params.InflationRateChange.Quo(hrsPerYrDec)},

		// 100% bonded, starting at 20% inflation and being reduced
		// (1 - (1/0.67))*(0.13/8667)
		{"test 2", sdk.OneDec(), sdk.ZeroDec(), sdk.NewDecWithPrec(20, 2),
			sdk.OneDec().Sub(sdk.OneDec().Quo(params.GoalBonded)).Mul(params.InflationRateChange).Quo(hrsPerYrDec)},

		// 50% bonded, starting at 10% inflation and being increased
		{"test 3", sdk.OneDec(), sdk.OneDec(), sdk.NewDecWithPrec(10, 2),
			sdk.OneDec().Sub(sdk.NewDecWithPrec(5, 1).Quo(params.GoalBonded)).Mul(params.InflationRateChange).Quo(hrsPerYrDec)},

		// test 7% minimum stop (testing with 100% bonded)
		{"test 4", sdk.OneDec(), sdk.ZeroDec(), sdk.NewDecWithPrec(7, 2), sdk.ZeroDec()},
		{"test 5", sdk.OneDec(), sdk.ZeroDec(), sdk.NewDecWithPrec(70001, 6), sdk.NewDecWithPrec(-1, 6)},

		// test 20% maximum stop (testing with 0% bonded)
		{"test 6", sdk.ZeroDec(), sdk.ZeroDec(), sdk.NewDecWithPrec(20, 2), sdk.ZeroDec()},
		{"test 7", sdk.ZeroDec(), sdk.ZeroDec(), sdk.NewDecWithPrec(199999, 6), sdk.NewDecWithPrec(1, 6)},

		// perfect balance shouldn't change inflation
		{"test 8", sdk.NewDec(67), sdk.NewDec(33), sdk.NewDecWithPrec(15, 2), sdk.ZeroDec()},
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
		cumulativeExpProvs       = sdk.ZeroDec()
	)
	pool.LooseTokens = sdk.NewDec(initialTotalTokens)

	// process the provisions for a year
	for hr := 0; hr < 100; hr++ {
		var expProvisions sdk.Dec
		_, expProvisions, pool = updateProvisions(t, pool, params, hr)
		cumulativeExpProvs = cumulativeExpProvs.Add(expProvisions)
	}

	//get the pool and do the final value checks from checkFinalPoolValues
	checkFinalPoolValues(t, pool, sdk.NewDec(initialTotalTokens), cumulativeExpProvs)
}

//_________________________________________________________________________________________
////////////////////////////////HELPER FUNCTIONS BELOW/////////////////////////////////////

// Final check on the global pool values for what the total tokens accumulated from each hour of provisions
func checkFinalPoolValues(t *testing.T, pool Pool, initialTotalTokens, cumulativeExpProvs sdk.Dec) {
	calculatedTotalTokens := initialTotalTokens.Add(cumulativeExpProvs)
	require.True(sdk.DecEq(t, calculatedTotalTokens, pool.TokenSupply()))
}

// Processes provisions are added to the pool correctly every hour
// Returns expected Provisions, expected Inflation, and pool, to help with cumulative calculations back in main Tests
func updateProvisions(t *testing.T, pool Pool, params Params, hr int) (sdk.Dec, sdk.Dec, Pool) {

	expInflation := pool.NextInflation(params)
	expProvisions := expInflation.Mul(pool.TokenSupply()).Quo(hrsPerYrDec)
	startTotalSupply := pool.TokenSupply()
	pool = pool.ProcessProvisions(params)

	//check provisions were added to pool
	require.True(sdk.DecEq(t, startTotalSupply.Add(expProvisions), pool.TokenSupply()))

	return expInflation, expProvisions, pool
}

// Checks that The inflation will correctly increase or decrease after an update to the pool
func checkInflation(t *testing.T, pool Pool, previousInflation, updatedInflation sdk.Dec, msg string) {
	inflationChange := updatedInflation.Sub(previousInflation)

	switch {
	//BELOW 67% - Rate of change positive and increasing, while we are between 7% <= and < 20% inflation
	case pool.BondedRatio().LT(sdk.NewDecWithPrec(67, 2)) && updatedInflation.LT(sdk.NewDecWithPrec(20, 2)):
		require.Equal(t, true, inflationChange.GT(sdk.ZeroDec()), msg)

	//BELOW 67% - Rate of change should be 0 while inflation continually stays at 20% until we reach 67% bonded ratio
	case pool.BondedRatio().LT(sdk.NewDecWithPrec(67, 2)) && updatedInflation.Equal(sdk.NewDecWithPrec(20, 2)):
		if previousInflation.Equal(sdk.NewDecWithPrec(20, 2)) {
			require.Equal(t, true, inflationChange.IsZero(), msg)

			//This else statement covers the one off case where we first hit 20%, but we still needed a positive ROC to get to 67% bonded ratio (i.e. we went from 19.99999% to 20%)
		} else {
			require.Equal(t, true, inflationChange.GT(sdk.ZeroDec()), msg)
		}

	//ABOVE 67% - Rate of change should be negative while the bond is above 67, and should stay negative until we reach inflation of 7%
	case pool.BondedRatio().GT(sdk.NewDecWithPrec(67, 2)) &&
		updatedInflation.LT(sdk.NewDecWithPrec(20, 2)) && updatedInflation.GT(sdk.NewDecWithPrec(7, 2)):
		require.Equal(t, true, inflationChange.LT(sdk.ZeroDec()), msg)

	//ABOVE 67% - Rate of change should be 0 while inflation continually stays at 7%.
	case pool.BondedRatio().GT(sdk.NewDecWithPrec(67, 2)) &&
		updatedInflation.Equal(sdk.NewDecWithPrec(7, 2)):

		if previousInflation.Equal(sdk.NewDecWithPrec(7, 2)) {
			require.Equal(t, true, inflationChange.IsZero(), msg)

			//This else statement covers the one off case where we first hit 7%, but we still needed a negative ROC to continue to get down to 67%. (i.e. we went from 7.00001% to 7%)
		} else {
			require.Equal(t, true, inflationChange.LT(sdk.ZeroDec()), msg)
		}
	}
}
