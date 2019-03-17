package types

import (
	"testing"
	"time"

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

func TestCommissionValidateNewRate(t *testing.T) {
	now := time.Now().UTC()
	c1 := NewCommission(sdk.MustNewDecFromStr("0.40"), sdk.MustNewDecFromStr("0.80"), sdk.MustNewDecFromStr("0.10"))
	c1.UpdateTime = now

	testCases := []struct {
		input     Commission
		newRate   sdk.Dec
		blockTime time.Time
		expectErr bool
	}{
		{c1, sdk.MustNewDecFromStr("0.50"), now, true},
		{c1, sdk.MustNewDecFromStr("-1.00"), now.Add(48 * time.Hour), true},
		{c1, sdk.MustNewDecFromStr("0.90"), now.Add(48 * time.Hour), true},
		{c1, sdk.MustNewDecFromStr("0.60"), now.Add(48 * time.Hour), true},
		{c1, sdk.MustNewDecFromStr("0.50"), now.Add(48 * time.Hour), false},
		{c1, sdk.MustNewDecFromStr("0.10"), now.Add(48 * time.Hour), false},
	}

	for i, tc := range testCases {
		err := tc.input.ValidateNewRate(tc.newRate, tc.blockTime)
		require.Equal(
			t, tc.expectErr, err != nil,
			"unexpected result; tc #%d, input: %v, newRate: %s, blockTime: %s",
			i, tc.input, tc.newRate, tc.blockTime,
		)
	}
}
