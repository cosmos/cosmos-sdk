package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestCommissionValidate(t *testing.T) {
	testCases := []struct {
		input     Commission
		expectErr bool
	}{
		{NewCommission(sdk.ZeroDec(), sdk.MustNewDecFromStr("-1.00"), sdk.ZeroDec()), true},
		{NewCommission(sdk.ZeroDec(), sdk.MustNewDecFromStr("2.00"), sdk.ZeroDec()), true},
		{NewCommission(sdk.MustNewDecFromStr("-1.00"), sdk.ZeroDec(), sdk.ZeroDec()), true},
		{NewCommission(sdk.MustNewDecFromStr("0.75"), sdk.MustNewDecFromStr("0.50"), sdk.ZeroDec()), true},
		{NewCommission(sdk.OneDec(), sdk.OneDec(), sdk.MustNewDecFromStr("-1.00")), true},
		{NewCommission(sdk.OneDec(), sdk.MustNewDecFromStr("0.75"), sdk.MustNewDecFromStr("0.90")), true},
		{NewCommission(sdk.MustNewDecFromStr("0.20"), sdk.OneDec(), sdk.MustNewDecFromStr("0.10")), false},
	}

	for i, tc := range testCases {
		err := tc.input.Validate()
		require.Equal(t, tc.expectErr, err != nil, "unexpected result; tc #%d, input: %v", i, tc.input)
	}
}
