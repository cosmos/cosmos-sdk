package sanction_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/testutil"
)

func TestDefaultParams(t *testing.T) {
	// cz is just a short way to create coins to use in these tests.
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}

	tests := []struct {
		name                string
		dfltImSanctMinDep   sdk.Coins
		dfltImUnsanctMinDep sdk.Coins
		exp                 *sanction.Params
	}{
		{
			name:                "nil nil",
			dfltImSanctMinDep:   nil,
			dfltImUnsanctMinDep: nil,
			exp: &sanction.Params{
				ImmediateSanctionMinDeposit:   nil,
				ImmediateUnsanctionMinDeposit: nil,
			},
		},
		{
			name:                "empty empty",
			dfltImSanctMinDep:   sdk.Coins{},
			dfltImUnsanctMinDep: sdk.Coins{},
			exp: &sanction.Params{
				ImmediateSanctionMinDeposit:   sdk.Coins{},
				ImmediateUnsanctionMinDeposit: sdk.Coins{},
			},
		},
		{
			name:                "with values",
			dfltImSanctMinDep:   cz("100000scoin"),
			dfltImUnsanctMinDep: cz("1000000ucoin"),
			exp: &sanction.Params{
				ImmediateSanctionMinDeposit:   cz("100000scoin"),
				ImmediateUnsanctionMinDeposit: cz("1000000ucoin"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			origDefaultImmediateSanctionMinDeposit := sanction.DefaultImmediateSanctionMinDeposit
			origDefaultImmediateUnsanctionMinDeposit := sanction.DefaultImmediateUnsanctionMinDeposit
			defer func() {
				sanction.DefaultImmediateSanctionMinDeposit = origDefaultImmediateSanctionMinDeposit
				sanction.DefaultImmediateUnsanctionMinDeposit = origDefaultImmediateUnsanctionMinDeposit
			}()
			sanction.DefaultImmediateSanctionMinDeposit = tc.dfltImSanctMinDep
			sanction.DefaultImmediateUnsanctionMinDeposit = tc.dfltImUnsanctMinDep

			var actual *sanction.Params
			testFunc := func() {
				actual = sanction.DefaultParams()
			}
			require.NotPanics(t, testFunc, "DefaultParams")
			if !assert.Equal(t, tc.exp, actual, "DefaultParams result") && actual != nil {
				// The test failed, but the output of coins isn't handy, so assert each individually for better info.
				assert.Equal(t, tc.exp.ImmediateSanctionMinDeposit.String(),
					actual.ImmediateSanctionMinDeposit.String(), "ImmediateSanctionMinDeposit")
				assert.Equal(t, tc.exp.ImmediateUnsanctionMinDeposit.String(),
					actual.ImmediateUnsanctionMinDeposit.String(), "ImmediateUnsanctionMinDeposit")
			}
		})
	}
}

func TestParams_ValidateBasic(t *testing.T) {
	tests := []struct {
		name   string
		params *sanction.Params
		exp    []string
	}{
		{
			name: "control",
			params: &sanction.Params{
				ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("goodcoin", 55)),
				ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("bettercoin", 88)),
			},
			exp: nil,
		},
		{
			name: "nil min deps",
			params: &sanction.Params{
				ImmediateSanctionMinDeposit:   nil,
				ImmediateUnsanctionMinDeposit: nil,
			},
			exp: nil,
		},
		{
			name: "empty min deps",
			params: &sanction.Params{
				ImmediateSanctionMinDeposit:   sdk.Coins{},
				ImmediateUnsanctionMinDeposit: sdk.Coins{},
			},
			exp: nil,
		},
		{
			name: "bad sanction min dep",
			params: &sanction.Params{
				ImmediateSanctionMinDeposit:   sdk.Coins{sdk.NewInt64Coin("dupcoin", 55), sdk.NewInt64Coin("dupcoin", 44)},
				ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("bettercoin", 88)),
			},
			exp: []string{"invalid immediate sanction min deposit", "duplicate denomination dupcoin", "invalid coins"},
		},
		{
			name: "bad unsanction min dep",
			params: &sanction.Params{
				ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("goodcoin", 55)),
				ImmediateUnsanctionMinDeposit: sdk.Coins{sdk.NewInt64Coin("twocoin", 88), sdk.NewInt64Coin("twocoin", 66)},
			},
			exp: []string{"invalid immediate unsanction min deposit", "duplicate denomination twocoin", "invalid coins"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.params.ValidateBasic()
			}
			require.NotPanics(t, testFunc, "ValidateBasic")
			testutil.AssertErrorContents(t, err, tc.exp, "ValidateBasic result")
		})
	}
}
