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
		{"fee is one", NewParams(defaultParams.NativeDenom, sdk.NewDec(1))},
		{"fee above one", NewParams(defaultParams.NativeDenom, sdk.NewDec(2))},
		{"fee is negative", NewParams(defaultParams.NativeDenom, sdk.NewDec(-1))},
	}

	for _, tc := range invalidTests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateParams(tc.params)
			require.NotNil(t, err)
		})
	}
}
