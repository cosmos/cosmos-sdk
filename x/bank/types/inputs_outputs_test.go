package types

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func TestValidateInputOutputs(t *testing.T) {
	addr := sdk.AccAddress([]byte("multisend-validation")).String()
	coin := func(denom string, amount int64) sdk.Coins {
		return sdk.Coins{{Denom: denom, Amount: math.NewInt(amount)}}
	}
	for _, tc := range []struct {
		name    string
		input   Input
		outputs []Output
		err     error
	}{
		{"shared denomination", Input{addr, coin("atom", 3)}, []Output{{addr, coin("atom", 1)}, {addr, coin("atom", 2)}}, nil},
		{"different denominations", Input{addr, sdk.NewCoins(sdk.NewInt64Coin("atom", 1), sdk.NewInt64Coin("stake", 2))}, []Output{{addr, coin("stake", 2)}, {addr, coin("atom", 1)}}, nil},
		{"missing denomination", Input{addr, sdk.NewCoins(sdk.NewInt64Coin("atom", 1), sdk.NewInt64Coin("stake", 2))}, []Output{{addr, coin("atom", 1)}}, ErrInputOutputMismatch},
		{"extra denomination", Input{addr, coin("atom", 1)}, []Output{{addr, coin("atom", 1)}, {addr, coin("stake", 2)}}, ErrInputOutputMismatch},
		{"different denomination with same count", Input{addr, coin("atom", 1)}, []Output{{addr, coin("stake", 1)}}, ErrInputOutputMismatch},
		{"incorrect amount", Input{addr, coin("atom", 3)}, []Output{{addr, coin("atom", 2)}}, ErrInputOutputMismatch},
		{"empty outputs", Input{addr, coin("atom", 1)}, nil, ErrInputOutputMismatch},
		{"invalid input first", Input{"", coin("atom", 1)}, []Output{{addr, coin("atom", -1)}}, sdkerrors.ErrInvalidAddress},
		{"invalid output before mismatch", Input{addr, coin("atom", 1)}, []Output{{addr, coin("stake", 2)}, {"", coin("atom", 1)}}, sdkerrors.ErrInvalidAddress},
		{"invalid coins before later invalid address", Input{addr, coin("atom", 1)}, []Output{{addr, coin("atom", -1)}, {"", coin("atom", 1)}}, sdkerrors.ErrInvalidCoins},
		{"unsorted output", Input{addr, coin("atom", 1)}, []Output{{addr, sdk.Coins{sdk.NewInt64Coin("stake", 1), sdk.NewInt64Coin("atom", 1)}}}, sdkerrors.ErrInvalidCoins},
		{"duplicate denomination in one output", Input{addr, coin("atom", 2)}, []Output{{addr, sdk.Coins{sdk.NewInt64Coin("atom", 1), sdk.NewInt64Coin("atom", 1)}}}, sdkerrors.ErrInvalidCoins},
		{"empty output coins", Input{addr, coin("atom", 1)}, []Output{{addr, nil}}, sdkerrors.ErrInvalidCoins},
	} {
		t.Run(tc.name, func(t *testing.T) {
			beforeInput, err := tc.input.Marshal()
			require.NoError(t, err)
			beforeOutputs := make([][]byte, len(tc.outputs))
			for i := range tc.outputs {
				beforeOutputs[i], err = tc.outputs[i].Marshal()
				require.NoError(t, err)
			}
			err = ValidateInputOutputs(tc.input, tc.outputs)
			if tc.err == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.err)
			}
			afterInput, err := tc.input.Marshal()
			require.NoError(t, err)
			require.Equal(t, beforeInput, afterInput)
			for i := range tc.outputs {
				afterOutput, err := tc.outputs[i].Marshal()
				require.NoError(t, err)
				require.Equal(t, beforeOutputs[i], afterOutput)
			}
		})
	}
}

func TestValidateInputOutputsAmountBoundary(t *testing.T) {
	addr := sdk.AccAddress([]byte("multisend-validation")).String()
	maximum := math.NewIntFromBigInt(new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)))
	input := Input{addr, sdk.Coins{sdk.NewCoin("atom", maximum)}}
	outputs := []Output{
		{addr, sdk.Coins{sdk.NewCoin("atom", maximum.SubRaw(1))}},
		{addr, sdk.Coins{sdk.NewInt64Coin("atom", 1)}},
	}
	require.NoError(t, ValidateInputOutputs(input, outputs))
	outputs[1].Coins[0].Amount = math.NewInt(2)
	require.Panics(t, func() { _ = ValidateInputOutputs(input, outputs) })
}

func multiSendValidationInput(n int, distinct bool) (Input, []Output) {
	addr := sdk.AccAddress([]byte("multisend-validation")).String()
	outputs := make([]Output, n)
	coins := make(sdk.Coins, 0, n)
	for i := range outputs {
		denom := "token"
		if distinct {
			denom = fmt.Sprintf("token%03d", i)
			coins = append(coins, sdk.NewInt64Coin(denom, 1))
		}
		outputs[i] = Output{addr, sdk.Coins{sdk.NewInt64Coin(denom, 1)}}
	}
	if !distinct {
		coins = sdk.Coins{sdk.NewInt64Coin("token", int64(n))}
	}
	return Input{addr, coins}, outputs
}

func TestValidateInputOutputsManyDenominations(t *testing.T) {
	for _, distinct := range []bool{false, true} {
		for _, n := range []int{1, 10, 100} {
			t.Run(fmt.Sprintf("distinct=%t/N=%d", distinct, n), func(t *testing.T) {
				input, outputs := multiSendValidationInput(n, distinct)
				require.NoError(t, ValidateInputOutputs(input, outputs))
				outputs[len(outputs)-1].Coins[0].Amount = math.NewInt(2)
				require.ErrorIs(t, ValidateInputOutputs(input, outputs), ErrInputOutputMismatch)
			})
		}
	}
}

func BenchmarkValidateInputOutputs(b *testing.B) {
	for _, distinct := range []bool{false, true} {
		for _, n := range []int{1, 10, 100} {
			b.Run(fmt.Sprintf("distinct=%t/N=%d", distinct, n), func(b *testing.B) {
				input, outputs := multiSendValidationInput(n, distinct)
				b.ReportAllocs()
				for b.Loop() {
					if err := ValidateInputOutputs(input, outputs); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}
