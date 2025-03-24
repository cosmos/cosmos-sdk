package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestRegisterInterfaces(t *testing.T) {
	interfaceRegistry := codectestutil.CodecOptions{}.NewInterfaceRegistry()
	RegisterInterfaces(interfaceRegistry)
	require.NoError(t, interfaceRegistry.EnsureRegistered(&MsgFundCommunityPool{}))
	require.NoError(t, interfaceRegistry.EnsureRegistered(&MsgCommunityPoolSpend{}))
	require.NoError(t, interfaceRegistry.EnsureRegistered(&MsgCreateContinuousFund{}))
	require.NoError(t, interfaceRegistry.EnsureRegistered(&MsgCancelContinuousFund{}))
}

func TestNewMsgFundCommunityPool(t *testing.T) {
	amount := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100)))
	depositor := "cosmos1qypq2q2l8z4wz2z2l8z4wz2z2l8z4wz2z2l8z4"
	msg := NewMsgFundCommunityPool(amount, depositor)
	require.Equal(t, amount, msg.Amount)
	require.Equal(t, depositor, msg.Depositor)
}

func TestNewMsgCommunityPoolSpend(t *testing.T) {
	amount := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100)))
	authority := "cosmos1qypq2q2l8z4wz2z2l8z4wz2z2l8z4wz2z2l8z4"
	recipient := "cosmos1qypq2q2l8z4wz2z2l8z4wz2z2l8z4wz2z2l8z5"
	msg := NewMsgCommunityPoolSpend(amount, authority, recipient)
	require.Equal(t, amount, msg.Amount)
	require.Equal(t, authority, msg.Authority)
	require.Equal(t, recipient, msg.Recipient)
}

func TestValidateGenesis(t *testing.T) {
	defaultGenesis := DefaultGenesisState()
	err := ValidateGenesis(defaultGenesis)
	require.NoError(t, err)

	gs := NewGenesisState(
		[]ContinuousFund{
			{
				Recipient:  "cosmos1qypq2q2l8z4wz2z2l8z4wz2z2l8z4wz2z2l8z4",
				Percentage: math.LegacyMustNewDecFromStr("0.1"),
				Expiry:     nil,
			},
		},
	)

	err = ValidateGenesis(gs)
	require.NoError(t, err)

	gs.ContinuousFunds[0].Recipient = ""
	err = ValidateGenesis(gs)
	require.EqualError(t, err, "recipient cannot be empty")
}

func TestValidateContinuousFund(t *testing.T) {
	testCases := []struct {
		name      string
		cf        ContinuousFund
		expErrMsg string
	}{
		{
			"valid continuous fund",
			ContinuousFund{
				Recipient:  "cosmos1qypq2q2l8z4wz2z2l8z4wz2z2l8z4wz2z2l8z4",
				Percentage: math.LegacyMustNewDecFromStr("0.1"),
				Expiry:     nil,
			},
			"",
		},
		{
			"empty recipient",
			ContinuousFund{
				Recipient: "",
			},
			"recipient cannot be empty",
		},
		{
			"zero percentage",
			ContinuousFund{
				Recipient:  "cosmos1qypq2q2l8z4wz2z2l8z4wz2z2l8z4wz2z2l8z4",
				Percentage: math.LegacyZeroDec(),
			},
			"percentage cannot be zero or empty",
		},
		{
			"nil percentage",
			ContinuousFund{
				Recipient:  "cosmos1qypq2q2l8z4wz2z2l8z4wz2z2l8z4wz2z2l8z4",
				Percentage: math.LegacyDec{},
			},
			"percentage cannot be zero or empty",
		},
		{
			"negative percentage",
			ContinuousFund{
				Recipient:  "cosmos1qypq2q2l8z4wz2z2l8z4wz2z2l8z4wz2z2l8z4",
				Percentage: math.LegacyMustNewDecFromStr("-0.1"),
			},
			"percentage cannot be negative",
		},
		{
			"percentage exceeds 100%",
			ContinuousFund{
				Recipient:  "cosmos1qypq2q2l8z4wz2z2l8z4wz2z2l8z4wz2z2l8z4",
				Percentage: math.LegacyMustNewDecFromStr("1.1"),
			},
			"percentage cannot be greater than one",
		},
	}

	for _, tc := range testCases {
		err := validateContinuousFund(tc.cf)
		if tc.expErrMsg == "" {
			require.NoError(t, err)
		} else {
			require.EqualError(t, err, tc.expErrMsg)
		}
	}
}

// TestPercentageCoinMul tests the PercentageCoinMul function.
func TestPercentageCoinMul(t *testing.T) {
	tests := []struct {
		name       string
		percentage math.LegacyDec
		coins      sdk.Coins
		expected   sdk.Coins
	}{
		{
			name:       "zero percentage",
			percentage: math.LegacyMustNewDecFromStr("0.0"),
			coins:      sdk.NewCoins(sdk.NewCoin("atom", math.NewInt(100))),
			expected:   sdk.NewCoins(sdk.NewCoin("atom", math.NewInt(0))),
		},
		{
			name:       "100 percent",
			percentage: math.LegacyMustNewDecFromStr("1.0"),
			coins:      sdk.NewCoins(sdk.NewCoin("atom", math.NewInt(100))),
			expected:   sdk.NewCoins(sdk.NewCoin("atom", math.NewInt(100))),
		},
		{
			name:       "50 percent",
			percentage: math.LegacyMustNewDecFromStr("0.5"),
			coins:      sdk.NewCoins(sdk.NewCoin("atom", math.NewInt(100))),
			expected:   sdk.NewCoins(sdk.NewCoin("atom", math.NewInt(50))),
		},
		{
			name:       "fraction with truncation",
			percentage: math.LegacyMustNewDecFromStr("0.333333333333333333"), // Approx. 1/3.
			coins:      sdk.NewCoins(sdk.NewCoin("atom", math.NewInt(100))),
			// 100 * 1/3 = 33.333... which truncates to 33.
			expected: sdk.NewCoins(sdk.NewCoin("atom", math.NewInt(33))),
		},
		{
			name:       "multiple denominations",
			percentage: math.LegacyMustNewDecFromStr("0.5"),
			coins: sdk.NewCoins(
				sdk.NewCoin("atom", math.NewInt(100)),
				sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(200)),
			),
			expected: sdk.NewCoins(
				sdk.NewCoin("atom", math.NewInt(50)),
				sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100)),
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Call the function under test.
			result := PercentageCoinMul(tc.percentage, tc.coins)

			// Compare the resulting coins with the expected coins.
			if !result.Equal(tc.expected) {
				t.Errorf("unexpected result for %s:\nexpected: %s\ngot:      %s", tc.name, tc.expected.String(), result.String())
			}
		})
	}
}
