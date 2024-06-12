package coins

import (
	"testing"

	"github.com/stretchr/testify/require"

	base "cosmossdk.io/api/cosmos/base/v1beta1"
)

func TestCoinIsZero(t *testing.T) {
	type testCase[T withAmount] struct {
		name   string
		coins  []T
		isZero bool
	}
	tests := []testCase[*base.Coin]{
		{
			name: "not zero coin",
			coins: []*base.Coin{
				{
					Denom:  "stake",
					Amount: "100",
				},
			},
			isZero: false,
		},
		{
			name: "zero coin",
			coins: []*base.Coin{
				{
					Denom:  "stake",
					Amount: "0",
				},
			},
			isZero: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsZero(tt.coins)
			require.NoError(t, err)
			require.Equal(t, got, tt.isZero)
		})
	}
}

func TestDecCoinIsZero(t *testing.T) {
	type testCase[T withAmount] struct {
		name   string
		coins  []T
		isZero bool
	}
	tests := []testCase[*base.DecCoin]{
		{
			name: "not zero coin",
			coins: []*base.DecCoin{
				{
					Denom:  "stake",
					Amount: "100",
				},
			},
			isZero: false,
		},
		{
			name: "zero coin",
			coins: []*base.DecCoin{
				{
					Denom:  "stake",
					Amount: "0",
				},
			},
			isZero: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsZero(tt.coins)
			require.NoError(t, err)
			require.Equal(t, got, tt.isZero)
		})
	}
}
