package types

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	"github.com/stretchr/testify/require"
)

func TestValidateParams(t *testing.T) {
	testCases := []struct {
		name    string
		params  Params
		expPass bool
	}{
		{"default params", DefaultParams(), true},
		{"custom params", NewParams(exported.Tendermint), true},
		{"blank client", NewParams(" "), false},
	}

	for _, tc := range testCases {
		err := tc.params.Validate()
		if tc.expPass {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}
	}
}
