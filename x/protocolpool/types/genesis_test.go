package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

func TestValidateGenesis(t *testing.T) {
	tests := []struct {
		name         string
		genesisState *types.GenesisState
		expectedErr  string
	}{
		{
			name:         "default genesis state valid",
			genesisState: types.DefaultGenesisState(),
			expectedErr:  "",
		},
		{
			name: "valid genesis state with a continuous fund",
			genesisState: &types.GenesisState{
				ContinuousFunds: []types.ContinuousFund{
					{
						Recipient:  "cosmos1validaddress",
						Percentage: math.LegacyMustNewDecFromStr("0.5"),
						Expiry:     nil,
					},
				},
				Params: types.DefaultParams(),
			},
			expectedErr: "",
		},
		{
			name: "invalid genesis state with continuous fund (empty recipient)",
			genesisState: &types.GenesisState{
				ContinuousFunds: []types.ContinuousFund{
					{
						Recipient:  "",
						Percentage: math.LegacyMustNewDecFromStr("0.5"),
						Expiry:     nil,
					},
				},
				Params: types.DefaultParams(),
			},
			expectedErr: "recipient cannot be empty",
		},
		{
			name: "invalid genesis state with params (zero distribution frequency)",
			genesisState: &types.GenesisState{
				ContinuousFunds: []types.ContinuousFund{},
				Params: types.Params{
					EnabledDistributionDenoms: []string{"stake"},
					DistributionFrequency:     0,
				},
			},
			expectedErr: "DistributionFrequency must be greater than 0",
		},
		{
			name: "invalid genesis state with continuous fund (percentage > 1)",
			genesisState: &types.GenesisState{
				ContinuousFunds: []types.ContinuousFund{
					{
						Recipient:  "cosmos1validaddress",
						Percentage: math.LegacyMustNewDecFromStr("1.1"),
						Expiry:     nil,
					},
				},
				Params: types.DefaultParams(),
			},
			expectedErr: "percentage cannot be greater than one",
		},
		{
			name: "invalid genesis state with continuous funds (percentage sum > 1)",
			genesisState: &types.GenesisState{
				ContinuousFunds: []types.ContinuousFund{
					{
						Recipient:  "cosmos1validaddress",
						Percentage: math.LegacyMustNewDecFromStr("0.5"),
						Expiry:     nil,
					},
					{
						Recipient:  "cosmos2validaddress",
						Percentage: math.LegacyMustNewDecFromStr("0.6"),
						Expiry:     nil,
					},
				},
				Params: types.DefaultParams(),
			},
			expectedErr: "total percentage cannot be greater than 100",
		},
		{
			name: "invalid genesis state with invalid continuous fund (percentage sum > 1)",
			genesisState: &types.GenesisState{
				ContinuousFunds: []types.ContinuousFund{
					{
						Recipient:  "cosmos1dupaddress",
						Percentage: math.LegacyMustNewDecFromStr("1.1"),
						Expiry:     nil,
					},
				},
				Params: types.DefaultParams(),
			},
			expectedErr: "percentage cannot be greater than one",
		},
		{
			name: "invalid genesis state with duplicate recipient addresses",
			genesisState: &types.GenesisState{
				ContinuousFunds: []types.ContinuousFund{
					{
						Recipient:  "cosmos1dupaddress",
						Percentage: math.LegacyMustNewDecFromStr("0.4"),
						Expiry:     nil,
					},
					{
						Recipient:  "cosmos1dupaddress", // duplicate
						Percentage: math.LegacyMustNewDecFromStr("0.3"),
						Expiry:     nil,
					},
				},
				Params: types.DefaultParams(),
			},
			expectedErr: "duplicated continuous fund recipient address",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.genesisState.Validate()
			if tc.expectedErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
