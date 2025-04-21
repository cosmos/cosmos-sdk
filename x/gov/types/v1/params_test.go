package v1_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func TestCalculateThresholdAndMinDeposit(t *testing.T) {
	// Period = 2 day
	// ExpeditedPeriod = 1 day
	// MinDeposit = 10000000stake
	// ExpeditedMinDeposit = 50000000stake
	// Threshold = 0.5
	// ExpeditedThreshold = 0.667
	params := v1.DefaultParams()

	tests := []struct {
		name          string
		duration      time.Duration
		expThreshold  string
		expMinDeposit []sdk.Coin
	}{
		{
			name:          "default",
			duration:      time.Hour * 48,
			expThreshold:  "0.500000000000000000",
			expMinDeposit: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10000000))),
		},
		{
			name:          "expedited",
			duration:      time.Hour * 24,
			expThreshold:  "0.667000000000000000",
			expMinDeposit: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(50000000))),
		},
		{
			name:          "less than expedited -> expedited",
			duration:      time.Hour * 23,
			expThreshold:  "0.667000000000000000",
			expMinDeposit: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(50000000))),
		},
		{
			name:          "greater than default -> default",
			duration:      time.Hour * 50,
			expThreshold:  "0.500000000000000000",
			expMinDeposit: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10000000))),
		},
		{
			name:          "(default + expedited)/2",
			duration:      time.Hour * 36,
			expThreshold:  "0.583500000000000000",
			expMinDeposit: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(30000000))),
		},
		{
			name:          "40 hours",
			duration:      time.Hour * 40,
			expThreshold:  "0.555666666666666673",
			expMinDeposit: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(23333334))),
		},
		{
			name:          "30 hours",
			duration:      time.Hour * 30,
			expThreshold:  "0.625250000000000000",
			expMinDeposit: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(40000000))),
		},
		{
			name:          "25 hours",
			duration:      time.Hour * 25,
			expThreshold:  "0.660041666666666667",
			expMinDeposit: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(48333334))),
		},
	}
	for _, tt := range tests {
		actualThreshold, actualMinDeposit := params.CalculateThresholdAndMinDeposit(tt.duration)
		require.Equal(t, tt.expThreshold, actualThreshold)
		require.Equal(t, tt.expMinDeposit, actualMinDeposit)
	}
}
