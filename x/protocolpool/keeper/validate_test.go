package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

// TestValidateAmount tests the validateAmount function.
func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name   string
		amount sdk.Coins
		expErr bool
		errMsg string
	}{
		{
			name:   "nil amount",
			amount: nil,
			expErr: true,
			errMsg: "amount cannot be nil",
		},
		{
			name: "negative coin amount",
			amount: sdk.Coins{
				{
					Denom:  "stake",
					Amount: math.NewInt(-100),
				},
			},
			expErr: true,
			errMsg: "-100",
		},
		{
			name:   "valid single coin",
			amount: sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
			expErr: false,
		},
		{
			name: "multiple valid coins",
			amount: sdk.NewCoins(
				sdk.NewCoin("stake", math.NewInt(100)),
				sdk.NewCoin("token", math.NewInt(200)),
			),
			expErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateAmount(tc.amount)
			if tc.expErr {
				require.Error(t, err, "expected an error but got none")
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateContinuousFund(t *testing.T) {
	now := time.Now()
	future := now.Add(1 * time.Hour)
	past := now.Add(-1 * time.Hour)

	// Create a context with the current block time.
	ctx := sdk.Context{}.WithBlockTime(now)

	tests := []struct {
		name   string
		msg    types.MsgCreateContinuousFund
		expErr bool
		errMsg string
	}{
		{
			name: "zero percentage",
			msg: types.MsgCreateContinuousFund{
				Authority:  "authority",
				Recipient:  "recipient",
				Percentage: math.LegacyZeroDec(),
				Expiry:     &future,
			},
			expErr: true,
			errMsg: "percentage cannot be zero or empty",
		},
		{
			name: "negative percentage",
			msg: types.MsgCreateContinuousFund{
				Authority:  "authority",
				Recipient:  "recipient",
				Percentage: math.LegacyNewDecFromInt(math.NewInt(-1)),
				Expiry:     &future,
			},
			expErr: true,
			errMsg: "percentage cannot be negative",
		},
		{
			name: "percentage greater than one",
			msg: types.MsgCreateContinuousFund{
				Authority:  "authority",
				Recipient:  "recipient",
				Percentage: math.LegacyMustNewDecFromStr("1.1"),
				Expiry:     &future,
			},
			expErr: true,
			errMsg: "percentage cannot be greater than one",
		},
		{
			name: "valid percentage with nil expiry",
			msg: types.MsgCreateContinuousFund{
				Authority:  "authority",
				Recipient:  "recipient",
				Percentage: math.LegacyNewDecWithPrec(5, 1), // 0.5
				Expiry:     nil,
			},
			expErr: false,
		},
		{
			name: "valid percentage with future expiry",
			msg: types.MsgCreateContinuousFund{
				Authority:  "authority",
				Recipient:  "recipient",
				Percentage: math.LegacyNewDecWithPrec(5, 1), // 0.5
				Expiry:     &future,
			},
			expErr: false,
		},
		{
			name: "expiry in past",
			msg: types.MsgCreateContinuousFund{
				Authority:  "authority",
				Recipient:  "recipient",
				Percentage: math.LegacyNewDecWithPrec(5, 1), // 0.5
				Expiry:     &past,
			},
			expErr: true,
			errMsg: "cannot be less than the current block time",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateContinuousFund(ctx, tc.msg)
			if tc.expErr {
				require.Error(t, err, "expected an error but got none")
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
