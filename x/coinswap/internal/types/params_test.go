package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestValidateParams(t *testing.T) {
	// check that valid case work
	defaultParams := DefaultParams()
	err := ValidateParams(defaultParams)
	require.Nil(t, err)

	// all cases should return an error
	invalidTests := []struct {
		name   string
		params Params
	}{
		{"empty native denom", NewParams("     ", defaultParams.Fee)},
		{"native denom with caps", NewParams("aTom", defaultParams.Fee)},
		{"native denom too short", NewParams("a", defaultParams.Fee)},
		{"native denom too long", NewParams("a very long coin denomination", defaultParams.Fee)},
		{"fee numerator == denominator", NewParams(defaultParams.NativeDenom, NewFeeParam(sdk.NewInt(1000), sdk.NewInt(1000)))},
		{"fee numerator > denominator", NewParams(defaultParams.NativeDenom, NewFeeParam(sdk.NewInt(10000), sdk.NewInt(10)))},
		{"fee numerator negative", NewParams(defaultParams.NativeDenom, NewFeeParam(sdk.NewInt(-1), sdk.NewInt(10)))},
		{"fee denominator negative", NewParams(defaultParams.NativeDenom, NewFeeParam(sdk.NewInt(10), sdk.NewInt(-1)))},
	}

	for _, tc := range invalidTests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateParams(tc.params)
			require.NotNil(t, err)
		})
	}
}
