package mint

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestNextInflation(t *testing.T) {
	minter := DefaultInitialMinter()
	params := DefaultParams()

	// Governing Mechanism:
	//    inflationRateChangePerYear = (1- BondedRatio/ GoalBonded) * MaxInflationRateChange

	tests := []struct {
		bondedRatio, setInflation, expChange sdk.Dec
	}{
		// with 0% bonded atom supply the inflation should increase by InflationRateChange
		{sdk.ZeroDec(), sdk.NewDecWithPrec(7, 2), params.InflationRateChange.Quo(hrsPerYr)},

		// 100% bonded, starting at 20% inflation and being reduced
		// (1 - (1/0.67))*(0.13/8667)
		{sdk.OneDec(), sdk.NewDecWithPrec(20, 2),
			sdk.OneDec().Sub(sdk.OneDec().Quo(params.GoalBonded)).Mul(params.InflationRateChange).Quo(hrsPerYr)},

		// 50% bonded, starting at 10% inflation and being increased
		{sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(10, 2),
			sdk.OneDec().Sub(sdk.NewDecWithPrec(5, 1).Quo(params.GoalBonded)).Mul(params.InflationRateChange).Quo(hrsPerYr)},

		// test 7% minimum stop (testing with 100% bonded)
		{sdk.OneDec(), sdk.NewDecWithPrec(7, 2), sdk.ZeroDec()},
		{sdk.OneDec(), sdk.NewDecWithPrec(70001, 6), sdk.NewDecWithPrec(-1, 6)},

		// test 20% maximum stop (testing with 0% bonded)
		{sdk.ZeroDec(), sdk.NewDecWithPrec(20, 2), sdk.ZeroDec()},
		{sdk.ZeroDec(), sdk.NewDecWithPrec(199999, 6), sdk.NewDecWithPrec(1, 6)},

		// perfect balance shouldn't change inflation
		{sdk.NewDecWithPrec(67, 2), sdk.NewDecWithPrec(15, 2), sdk.ZeroDec()},
	}
	for i, tc := range tests {
		minter.Inflation = tc.setInflation

		inflation := minter.NextInflationRate(params, tc.bondedRatio)
		diffInflation := inflation.Sub(tc.setInflation)

		require.True(t, diffInflation.Equal(tc.expChange),
			"Test Index: %v\nDiff:  %v\nExpected: %v\n", i, diffInflation, tc.expChange)
	}
}

func TestNextProvision(t *testing.T) {
	minter := InitialMinter(sdk.NewDecWithPrec(1, 1))
	params := DefaultParams()

	tests := []struct {
		hourlyProvisions int64
		blockTime        time.Time
		expProvisions    sdk.Coin
	}{
		{3600, time.Unix(0, 0), sdk.NewCoin(params.MintDenom, sdk.NewInt(0))},
		{3600, time.Unix(60*60, 0), sdk.NewCoin(params.MintDenom, sdk.NewInt(3600))},
		{3600, time.Unix(60, 0), sdk.NewCoin(params.MintDenom, sdk.NewInt(60))},
		{3600, time.Unix(1, 0), sdk.NewCoin(params.MintDenom, sdk.NewInt(1))},

		// some truncate cases
		{3600, time.Unix(1, 500), sdk.NewCoin(params.MintDenom, sdk.NewInt(1))},
		{3600, time.Unix(1, 999), sdk.NewCoin(params.MintDenom, sdk.NewInt(1))},
		{3601, time.Unix(1, 0), sdk.NewCoin(params.MintDenom, sdk.NewInt(1))},
		{3601, time.Unix(1800, 0), sdk.NewCoin(params.MintDenom, sdk.NewInt(1800))},
		{3601, time.Unix(3600, 0), sdk.NewCoin(params.MintDenom, sdk.NewInt(3601))},
		{3700, time.Unix(1800, 0), sdk.NewCoin(params.MintDenom, sdk.NewInt(1850))},
	}
	for i, tc := range tests {
		minter.HourlyProvisions = sdk.NewInt(tc.hourlyProvisions)
		provisions := minter.NextProvision(params, tc.blockTime)

		require.True(t, tc.expProvisions.IsEqual(provisions),
			"test: %v\n\tExp: %v\n\tGot: %v\n",
			i, tc.expProvisions, provisions)
	}
}
