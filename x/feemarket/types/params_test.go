package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/x/feemarket/types"
)

func TestParams(t *testing.T) {
	testCases := []struct {
		name        string
		p           types.Params
		expectedErr bool
	}{
		{
			name:        "valid base eip-1559 params",
			p:           types.DefaultParams(),
			expectedErr: false,
		},
		{
			name:        "valid aimd eip-1559 params",
			p:           types.DefaultAIMDParams(),
			expectedErr: false,
		},
		{
			name:        "invalid window",
			p:           types.Params{},
			expectedErr: true,
		},
		{
			name: "nil alpha",
			p: types.Params{
				Window:   1,
				FeeDenom: types.DefaultFeeDenom,
			},
			expectedErr: true,
		},
		{
			name: "negative alpha",
			p: types.Params{
				Window:   1,
				Alpha:    math.LegacyMustNewDecFromStr("-0.1"),
				FeeDenom: types.DefaultFeeDenom,
			},
			expectedErr: true,
		},
		{
			name: "beta is nil",
			p: types.Params{
				Window:   1,
				Alpha:    math.LegacyMustNewDecFromStr("0.1"),
				FeeDenom: types.DefaultFeeDenom,
			},
			expectedErr: true,
		},
		{
			name: "beta is negative",
			p: types.Params{
				Window:   1,
				Alpha:    math.LegacyMustNewDecFromStr("0.1"),
				Beta:     math.LegacyMustNewDecFromStr("-0.1"),
				FeeDenom: types.DefaultFeeDenom,
			},
			expectedErr: true,
		},
		{
			name: "beta is greater than 1",
			p: types.Params{
				Window:   1,
				Alpha:    math.LegacyMustNewDecFromStr("0.1"),
				Beta:     math.LegacyMustNewDecFromStr("1.1"),
				FeeDenom: types.DefaultFeeDenom,
			},
			expectedErr: true,
		},
		{
			name: "theta is nil",
			p: types.Params{
				Window:   1,
				Alpha:    math.LegacyMustNewDecFromStr("0.1"),
				Beta:     math.LegacyMustNewDecFromStr("0.1"),
				FeeDenom: types.DefaultFeeDenom,
			},
			expectedErr: true,
		},
		{
			name: "theta is negative",
			p: types.Params{
				Window:   1,
				Alpha:    math.LegacyMustNewDecFromStr("0.1"),
				Beta:     math.LegacyMustNewDecFromStr("0.1"),
				Gamma:    math.LegacyMustNewDecFromStr("-0.1"),
				FeeDenom: types.DefaultFeeDenom,
			},
			expectedErr: true,
		},
		{
			name: "theta is greater than 1",
			p: types.Params{
				Window:   1,
				Alpha:    math.LegacyMustNewDecFromStr("0.1"),
				Beta:     math.LegacyMustNewDecFromStr("0.1"),
				Gamma:    math.LegacyMustNewDecFromStr("1.1"),
				FeeDenom: types.DefaultFeeDenom,
			},
			expectedErr: true,
		},
		{
			name: "delta is nil",
			p: types.Params{
				Window:   1,
				Alpha:    math.LegacyMustNewDecFromStr("0.1"),
				Beta:     math.LegacyMustNewDecFromStr("0.1"),
				Gamma:    math.LegacyMustNewDecFromStr("0.1"),
				FeeDenom: types.DefaultFeeDenom,
			},
			expectedErr: true,
		},
		{
			name: "delta is negative",
			p: types.Params{
				Window:   1,
				Alpha:    math.LegacyMustNewDecFromStr("0.1"),
				Beta:     math.LegacyMustNewDecFromStr("0.1"),
				Gamma:    math.LegacyMustNewDecFromStr("0.1"),
				Delta:    math.LegacyMustNewDecFromStr("-0.1"),
				FeeDenom: types.DefaultFeeDenom,
			},
			expectedErr: true,
		},
		{
			name: "target block size is zero",
			p: types.Params{
				Window:   1,
				Alpha:    math.LegacyMustNewDecFromStr("0.1"),
				Beta:     math.LegacyMustNewDecFromStr("0.1"),
				Gamma:    math.LegacyMustNewDecFromStr("0.1"),
				Delta:    math.LegacyMustNewDecFromStr("0.1"),
				FeeDenom: types.DefaultFeeDenom,
			},
			expectedErr: true,
		},
		{
			name: "max block size is zero",
			p: types.Params{
				Window:   1,
				Alpha:    math.LegacyMustNewDecFromStr("0.1"),
				Beta:     math.LegacyMustNewDecFromStr("0.1"),
				Gamma:    math.LegacyMustNewDecFromStr("0.1"),
				Delta:    math.LegacyMustNewDecFromStr("0.1"),
				FeeDenom: types.DefaultFeeDenom,
			},
			expectedErr: true,
		},
		{
			name: "min base gas price is nil",
			p: types.Params{
				Window:              1,
				Alpha:               math.LegacyMustNewDecFromStr("0.1"),
				Beta:                math.LegacyMustNewDecFromStr("0.1"),
				Gamma:               math.LegacyMustNewDecFromStr("0.1"),
				Delta:               math.LegacyMustNewDecFromStr("0.1"),
				MaxBlockUtilization: 3,
				FeeDenom:            types.DefaultFeeDenom,
			},
			expectedErr: true,
		},
		{
			name: "min base has price is negative",
			p: types.Params{
				Window:              1,
				Alpha:               math.LegacyMustNewDecFromStr("0.1"),
				Beta:                math.LegacyMustNewDecFromStr("0.1"),
				Gamma:               math.LegacyMustNewDecFromStr("0.1"),
				Delta:               math.LegacyMustNewDecFromStr("0.1"),
				MaxBlockUtilization: 3,
				MinBaseGasPrice:     math.LegacyMustNewDecFromStr("-1.0"),
				FeeDenom:            types.DefaultFeeDenom,
			},
			expectedErr: true,
		},
		{
			name: "min learning rate is nil",
			p: types.Params{
				Window:              1,
				Alpha:               math.LegacyMustNewDecFromStr("0.1"),
				Beta:                math.LegacyMustNewDecFromStr("0.1"),
				Gamma:               math.LegacyMustNewDecFromStr("0.1"),
				Delta:               math.LegacyMustNewDecFromStr("0.1"),
				MaxBlockUtilization: 3,
				MinBaseGasPrice:     math.LegacyMustNewDecFromStr("1.0"),
				FeeDenom:            types.DefaultFeeDenom,
			},
			expectedErr: true,
		},
		{
			name: "min learning rate is negative",
			p: types.Params{
				Window:              1,
				Alpha:               math.LegacyMustNewDecFromStr("0.1"),
				Beta:                math.LegacyMustNewDecFromStr("0.1"),
				Gamma:               math.LegacyMustNewDecFromStr("0.1"),
				Delta:               math.LegacyMustNewDecFromStr("0.1"),
				MaxBlockUtilization: 3,
				MinBaseGasPrice:     math.LegacyMustNewDecFromStr("1.0"),
				MinLearningRate:     math.LegacyMustNewDecFromStr("-0.1"),
				FeeDenom:            types.DefaultFeeDenom,
			},
			expectedErr: true,
		},
		{
			name: "max learning rate is nil",
			p: types.Params{
				Window:              1,
				Alpha:               math.LegacyMustNewDecFromStr("0.1"),
				Beta:                math.LegacyMustNewDecFromStr("0.1"),
				Gamma:               math.LegacyMustNewDecFromStr("0.1"),
				Delta:               math.LegacyMustNewDecFromStr("0.1"),
				MaxBlockUtilization: 3,
				MinBaseGasPrice:     math.LegacyMustNewDecFromStr("1.0"),
				MinLearningRate:     math.LegacyMustNewDecFromStr("0.1"),
				FeeDenom:            types.DefaultFeeDenom,
			},
			expectedErr: true,
		},
		{
			name: "max learning rate is negative",
			p: types.Params{
				Window:              1,
				Alpha:               math.LegacyMustNewDecFromStr("0.1"),
				Beta:                math.LegacyMustNewDecFromStr("0.1"),
				Gamma:               math.LegacyMustNewDecFromStr("0.1"),
				Delta:               math.LegacyMustNewDecFromStr("0.1"),
				MaxBlockUtilization: 3,
				MinBaseGasPrice:     math.LegacyMustNewDecFromStr("1.0"),
				MinLearningRate:     math.LegacyMustNewDecFromStr("0.1"),
				MaxLearningRate:     math.LegacyMustNewDecFromStr("-0.1"),
				FeeDenom:            types.DefaultFeeDenom,
			},
			expectedErr: true,
		},
		{
			name: "min learning rate is greater than max learning rate",
			p: types.Params{
				Window:              1,
				Alpha:               math.LegacyMustNewDecFromStr("0.1"),
				Beta:                math.LegacyMustNewDecFromStr("0.1"),
				Gamma:               math.LegacyMustNewDecFromStr("0.1"),
				Delta:               math.LegacyMustNewDecFromStr("0.1"),
				MaxBlockUtilization: 3,
				MinBaseGasPrice:     math.LegacyMustNewDecFromStr("1.0"),
				MinLearningRate:     math.LegacyMustNewDecFromStr("0.1"),
				MaxLearningRate:     math.LegacyMustNewDecFromStr("0.05"),
				FeeDenom:            types.DefaultFeeDenom,
			},
			expectedErr: true,
		},
		{
			name: "fee denom is empty",
			p: types.Params{
				Window:              1,
				Alpha:               math.LegacyMustNewDecFromStr("0.1"),
				Beta:                math.LegacyMustNewDecFromStr("0.1"),
				Gamma:               math.LegacyMustNewDecFromStr("0.1"),
				Delta:               math.LegacyMustNewDecFromStr("0.1"),
				MaxBlockUtilization: 3,
				MinBaseGasPrice:     math.LegacyMustNewDecFromStr("1.0"),
				MinLearningRate:     math.LegacyMustNewDecFromStr("0.01"),
				MaxLearningRate:     math.LegacyMustNewDecFromStr("0.05"),
			},
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.p.ValidateBasic()
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
