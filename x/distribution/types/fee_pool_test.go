package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	"cosmossdk.io/x/distribution/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestValidateGenesis(t *testing.T) {
	testCases := []struct {
		name        string
		feePool     types.FeePool
		shouldPanic bool
		expectError bool
	}{
		{
			name:        "valid fee pool",
			feePool:     types.InitialFeePool(),
			shouldPanic: false,
			expectError: false,
		},
		{
			name: "negative decimal pool",
			feePool: types.FeePool{
				DecimalPool:   sdk.DecCoins{{Denom: "stake", Amount: math.LegacyNewDec(-1)}},
				CommunityPool: sdk.DecCoins{},
			},
			shouldPanic: false,
			expectError: true,
		},
		{
			name: "negative community pool",
			feePool: types.FeePool{
				DecimalPool:   sdk.DecCoins{},
				CommunityPool: sdk.DecCoins{{Denom: "stake", Amount: math.LegacyNewDec(-1)}},
			},
			shouldPanic: true,
			expectError: false,
		},
		{
			name: "multiple coins in community pool with one negative value",
			feePool: types.FeePool{
				DecimalPool: sdk.DecCoins{},
				CommunityPool: sdk.DecCoins{
					{Denom: "stake", Amount: math.LegacyNewDec(1)},
					{Denom: "atom", Amount: math.LegacyNewDec(-2)},
				},
			},
			shouldPanic: true,
			expectError: false,
		},
		{
			name: "multiple coins in community pool with all positive values",
			feePool: types.FeePool{
				DecimalPool: sdk.DecCoins{},
				CommunityPool: sdk.DecCoins{
					{Denom: "stake", Amount: math.LegacyNewDec(1)},
					{Denom: "atom", Amount: math.LegacyNewDec(2)},
				},
			},
			shouldPanic: false,
			expectError: false,
		},
		{
			name: "nil decimal pool",
			feePool: types.FeePool{
				CommunityPool: sdk.DecCoins{},
			},
			shouldPanic: false,
			expectError: false,
		},
		{
			name: "nil community pool",
			feePool: types.FeePool{
				DecimalPool: sdk.DecCoins{},
			},
			shouldPanic: false,
			expectError: false,
		},
		{
			name: "zero value community pool",
			feePool: types.FeePool{
				DecimalPool:   sdk.DecCoins{},
				CommunityPool: sdk.DecCoins{{Denom: "stake", Amount: math.LegacyNewDec(0)}},
			},
			shouldPanic: false,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				require.Panics(t, func() {
					tc.feePool.ValidateGenesis()
				}, "ValidateGenesis should panic for test case: %s", tc.name)
			} else {
				err := tc.feePool.ValidateGenesis()
				if tc.expectError {
					require.Error(t, err, "ValidateGenesis should return error for test case: %s", tc.name)
				} else {
					require.NoError(t, err, "ValidateGenesis should not return error for valid FeePool in test case: %s", tc.name)
				}
			}
		})
	}
}
