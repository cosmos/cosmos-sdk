package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestBalanceValidate(t *testing.T) {

	testCases := []struct {
		name    string
		balance Balance
		expErr  bool
	}{
		{
			"valid balance",
			Balance{
				Address: "cosmos1yq8lgssgxlx9smjhes6ryjasmqmd3ts2559g0t",
				Coins:   sdk.Coins{sdk.NewInt64Coin("uatom", 1)},
			},
			false,
		},
		{"empty balance", Balance{}, true},
		{
			"nil balance coins",
			Balance{
				Address: "cosmos1yq8lgssgxlx9smjhes6ryjasmqmd3ts2559g0t",
			},
			true,
		},
		{
			"dup coins",
			Balance{
				Address: "cosmos1yq8lgssgxlx9smjhes6ryjasmqmd3ts2559g0t",
				Coins: sdk.Coins{
					sdk.NewInt64Coin("uatom", 1),
					sdk.NewInt64Coin("uatom", 1),
				},
			},
			true,
		},
		{
			"invalid coin denom",
			Balance{
				Address: "cosmos1yq8lgssgxlx9smjhes6ryjasmqmd3ts2559g0t",
				Coins: sdk.Coins{
					sdk.Coin{Denom: "", Amount: sdk.OneInt()},
				},
			},
			true,
		},
		{
			"negative coin",
			Balance{
				Address: "cosmos1yq8lgssgxlx9smjhes6ryjasmqmd3ts2559g0t",
				Coins: sdk.Coins{
					sdk.Coin{Denom: "uatom", Amount: sdk.NewInt(-1)},
				},
			},
			true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			err := tc.balance.Validate()

			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
