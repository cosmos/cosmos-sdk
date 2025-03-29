package types_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

func TestParamsValidation(t *testing.T) {
	tests := []struct {
		name        string
		params      types.Params
		expectErr   bool
		errContains string
	}{
		{
			name:        "default params valid",
			params:      types.DefaultParams(),
			expectErr:   false,
			errContains: "",
		},
		{
			name: "custom valid params",
			params: types.Params{
				EnabledDistributionDenoms: []string{sdk.DefaultBondDenom},
				DistributionFrequency:     10,
			},
			expectErr:   false,
			errContains: "",
		},
		{
			name: "invalid denom",
			params: types.Params{
				EnabledDistributionDenoms: []string{"invalid!denom"},
				DistributionFrequency:     1,
			},
			expectErr:   true,
			errContains: "invalid denom",
		},
		{
			name: "duplicate denom",
			params: types.Params{
				EnabledDistributionDenoms: []string{sdk.DefaultBondDenom, sdk.DefaultBondDenom},
				DistributionFrequency:     1,
			},
			expectErr:   true,
			errContains: "duplicate enabled distribution denom",
		},
		{
			name: "zero distribution frequency",
			params: types.Params{
				EnabledDistributionDenoms: []string{sdk.DefaultBondDenom},
				DistributionFrequency:     0,
			},
			expectErr:   true,
			errContains: "DistributionFrequency must be greater than 0",
		},
	}

	for _, tc := range tests {
		// capture range variable
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()
			if tc.expectErr {
				require.Error(t, err)
				require.True(t, strings.Contains(err.Error(), tc.errContains),
					"expected error to contain %q, got %q", tc.errContains, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
