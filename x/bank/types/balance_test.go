package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestBalanceValidate(t *testing.T) {
	testCases := []struct {
		name    string
		balance bank.Balance
		expErr  bool
	}{
		{
			"valid balance",
			bank.Balance{
				Address: "cosmos1yq8lgssgxlx9smjhes6ryjasmqmd3ts2559g0t",
				Coins:   sdk.Coins{sdk.NewInt64Coin("uatom", 1)},
			},
			false,
		},
		{"empty balance", bank.Balance{}, true},
		{
			"nil balance coins",
			bank.Balance{
				Address: "cosmos1yq8lgssgxlx9smjhes6ryjasmqmd3ts2559g0t",
			},
			true,
		},
		{
			"dup coins",
			bank.Balance{
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
			bank.Balance{
				Address: "cosmos1yq8lgssgxlx9smjhes6ryjasmqmd3ts2559g0t",
				Coins: sdk.Coins{
					sdk.Coin{Denom: "", Amount: sdk.OneInt()},
				},
			},
			true,
		},
		{
			"negative coin",
			bank.Balance{
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

func TestBalance_GetAddress(t *testing.T) {
	tests := []struct {
		name      string
		Address   string
		wantPanic bool
	}{
		{"empty address", "", true},
		{"malformed address", "invalid", true},
		{"valid address", "cosmos1vy0ga0klndqy92ceqehfkvgmn4t94eteq4hmqv", false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := bank.Balance{Address: tt.Address}
			if tt.wantPanic {
				require.Panics(t, func() { b.GetAddress() })
			} else {
				require.False(t, b.GetAddress().Empty())
			}
		})
	}
}
