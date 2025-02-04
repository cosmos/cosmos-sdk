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
	}{
		{
			name: "valid fee pool",
			feePool: types.FeePool{
				DecimalPool:   sdk.DecCoins{},
				CommunityPool: sdk.DecCoins{},
			},
			shouldPanic: false,
		},
		{
			name: "negative decimal pool",
			feePool: types.FeePool{
				DecimalPool: sdk.DecCoins{
					sdk.DecCoin{Denom: "stake", Amount: sdk.NewDec(-1)},
				},
				CommunityPool: sdk.DecCoins{},
			},
			shouldPanic: false,
		},
		{
			name: "non-zero community pool",
			feePool: types.FeePool{
				DecimalPool: sdk.DecCoins{},
				CommunityPool: sdk.DecCoins{
					sdk.DecCoin{Denom: "stake", Amount: sdk.NewDec(1)},
				},
			},
			shouldPanic: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				require.Panics(t, func() {
					tc.feePool.ValidateGenesis()
				}, "expected ValidateGenesis to panic with non-zero community pool")
			} else {
				err := tc.feePool.ValidateGenesis()
				if tc.feePool.DecimalPool.IsAnyNegative() {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			}
		})
	}
}
