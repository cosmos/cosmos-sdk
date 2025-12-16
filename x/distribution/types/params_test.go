package types_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func TestParams_ValidateBasic(t *testing.T) {
	toDec := sdkmath.LegacyMustNewDecFromStr
	tests := []struct {
		name   string
		params types.Params
		err    error
	}{
		{
			name:   "success",
			params: types.DefaultParams(),
		},
		{
			name: "negative community tax",
			params: types.Params{
				CommunityTax:        toDec("-0.1"),
				WithdrawAddrEnabled: false,
				NakamotoBonus: types.NakamotoBonus{
					Enabled:            true,
					Period:             types.DefaultNakamotoBonusPeriod,
					Step:               types.DefaultNakamotoBonusStep,
					MinimumCoefficient: types.DefaultNakamotoBonusMinimumCoefficient,
					MaximumCoefficient: types.DefaultNakamotoBonusMaximumCoefficient,
				},
			},
			err: fmt.Errorf("community tax must be positive: -0.100000000000000000"),
		},
		{
			name: "negative base proposer reward (must not matter)",
			params: types.Params{
				CommunityTax:        toDec("0.1"),
				WithdrawAddrEnabled: false,
				NakamotoBonus: types.NakamotoBonus{
					Enabled:            true,
					Period:             types.DefaultNakamotoBonusPeriod,
					Step:               types.DefaultNakamotoBonusStep,
					MinimumCoefficient: types.DefaultNakamotoBonusMinimumCoefficient,
					MaximumCoefficient: types.DefaultNakamotoBonusMaximumCoefficient,
				},
			},
		},
		{
			name: "negative bonus proposer reward (must not matter)",
			params: types.Params{
				CommunityTax:        toDec("0.1"),
				WithdrawAddrEnabled: false,
				NakamotoBonus: types.NakamotoBonus{
					Enabled:            true,
					Period:             types.DefaultNakamotoBonusPeriod,
					Step:               types.DefaultNakamotoBonusStep,
					MinimumCoefficient: types.DefaultNakamotoBonusMinimumCoefficient,
					MaximumCoefficient: types.DefaultNakamotoBonusMaximumCoefficient,
				},
			},
		},
		{
			name: "total sum greater than 1 (must not matter)",
			params: types.Params{
				CommunityTax:        toDec("0.2"),
				WithdrawAddrEnabled: false,
				NakamotoBonus: types.NakamotoBonus{
					Enabled:            true,
					Period:             types.DefaultNakamotoBonusPeriod,
					Step:               types.DefaultNakamotoBonusStep,
					MinimumCoefficient: types.DefaultNakamotoBonusMinimumCoefficient,
					MaximumCoefficient: types.DefaultNakamotoBonusMaximumCoefficient,
				},
			},
		},
		{
			name: "community tax greater than 1",
			params: types.Params{
				CommunityTax:        toDec("1.1"),
				WithdrawAddrEnabled: false,
				NakamotoBonus: types.NakamotoBonus{
					Enabled:            true,
					Period:             types.DefaultNakamotoBonusPeriod,
					Step:               types.DefaultNakamotoBonusStep,
					MinimumCoefficient: types.DefaultNakamotoBonusMinimumCoefficient,
					MaximumCoefficient: types.DefaultNakamotoBonusMaximumCoefficient,
				},
			},
			err: fmt.Errorf("community tax too large: 1.100000000000000000"),
		},
		{
			name: "community tax nil",
			params: types.Params{
				CommunityTax:        sdkmath.LegacyDec{},
				WithdrawAddrEnabled: false,
				NakamotoBonus: types.NakamotoBonus{
					Enabled:            true,
					Period:             types.DefaultNakamotoBonusPeriod,
					Step:               types.DefaultNakamotoBonusStep,
					MinimumCoefficient: types.DefaultNakamotoBonusMinimumCoefficient,
					MaximumCoefficient: types.DefaultNakamotoBonusMaximumCoefficient,
				},
			},
			err: fmt.Errorf("community tax must be not nil"),
		},
		{
			name: "community tax too large",
			params: types.Params{
				CommunityTax:        sdkmath.LegacyOneDec().Add(sdkmath.LegacyNewDec(1)),
				WithdrawAddrEnabled: true,
				NakamotoBonus: types.NakamotoBonus{
					Enabled:            true,
					Period:             types.DefaultNakamotoBonusPeriod,
					Step:               types.DefaultNakamotoBonusStep,
					MinimumCoefficient: types.DefaultNakamotoBonusMinimumCoefficient,
					MaximumCoefficient: types.DefaultNakamotoBonusMaximumCoefficient,
				},
			},
			err: fmt.Errorf("community tax too large: 2.000000000000000000"),
		},
		{
			name: "success with nakamoto bonus enabled",
			params: types.Params{
				CommunityTax:        toDec("0.1"),
				WithdrawAddrEnabled: true,
				NakamotoBonus: types.NakamotoBonus{
					Enabled:            true,
					Period:             types.DefaultNakamotoBonusPeriod,
					Step:               types.DefaultNakamotoBonusStep,
					MinimumCoefficient: types.DefaultNakamotoBonusMinimumCoefficient,
					MaximumCoefficient: types.DefaultNakamotoBonusMaximumCoefficient,
				},
			},
		},
		{
			name: "negative nakamoto bonus coefficient",
			params: types.Params{
				CommunityTax:        toDec("0.1"),
				WithdrawAddrEnabled: true,
				NakamotoBonus: types.NakamotoBonus{
					Enabled:            true,
					Period:             types.DefaultNakamotoBonusPeriod,
					Step:               types.DefaultNakamotoBonusStep,
					MinimumCoefficient: types.DefaultNakamotoBonusMinimumCoefficient,
					MaximumCoefficient: types.DefaultNakamotoBonusMaximumCoefficient,
				},
			},
		},
		{
			name: "negative nakamoto bonus step",
			params: types.Params{
				CommunityTax:        toDec("0.1"),
				WithdrawAddrEnabled: true,
				NakamotoBonus: types.NakamotoBonus{
					Enabled:            true,
					Period:             types.DefaultNakamotoBonusPeriod,
					Step:               toDec("-0.3"),
					MinimumCoefficient: types.DefaultNakamotoBonusMinimumCoefficient,
					MaximumCoefficient: types.DefaultNakamotoBonusMaximumCoefficient,
				},
			},
			err: fmt.Errorf("nakamoto bonus step must be positive: -0.300000000000000000"),
		},
		{
			name: "nakamoto bonus step nil",
			params: types.Params{
				CommunityTax:        toDec("0.1"),
				WithdrawAddrEnabled: true,
				NakamotoBonus: types.NakamotoBonus{
					Enabled:            true,
					Period:             types.DefaultNakamotoBonusPeriod,
					MinimumCoefficient: types.DefaultNakamotoBonusMinimumCoefficient,
					MaximumCoefficient: types.DefaultNakamotoBonusMaximumCoefficient,
				},
			},
			err: fmt.Errorf("nakamoto bonus step must be not nil"),
		},
		{
			name: "nakamoto bonus step too large",
			params: types.Params{
				CommunityTax:        toDec("0.1"),
				WithdrawAddrEnabled: true,
				NakamotoBonus: types.NakamotoBonus{
					Enabled:            true,
					Period:             types.DefaultNakamotoBonusPeriod,
					Step:               sdkmath.LegacyOneDec().Add(sdkmath.LegacyNewDec(1)),
					MinimumCoefficient: types.DefaultNakamotoBonusMinimumCoefficient,
					MaximumCoefficient: types.DefaultNakamotoBonusMaximumCoefficient,
				},
			},
			err: fmt.Errorf("nakamoto bonus step too large: 2.000000000000000000"),
		},
		{
			name: "nakamoto bonus period zero",
			params: types.Params{
				CommunityTax:        toDec("0.1"),
				WithdrawAddrEnabled: true,
				NakamotoBonus: types.NakamotoBonus{
					Enabled:            true,
					Period:             0,
					Step:               types.DefaultNakamotoBonusStep,
					MinimumCoefficient: types.DefaultNakamotoBonusMinimumCoefficient,
					MaximumCoefficient: types.DefaultNakamotoBonusMaximumCoefficient,
				},
			},
			err: fmt.Errorf("nakamoto bonus period must be greater than zero: 0"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.ValidateBasic()
			if tt.err != nil {
				require.Error(t, err, "expected error: %v, got: %v", tt.err, err)
				require.Equal(t, tt.err, err, "expected error: %v, got: %v", tt.err, err)
				return
			}
			require.NoError(t, err, "expected no error, got: %v", err)
		})
	}
}

func TestDefaultParams(t *testing.T) {
	require.NoError(t, types.DefaultParams().ValidateBasic())
}
