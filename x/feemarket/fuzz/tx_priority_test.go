package fuzz_test

import (
	"math"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/x/feemarket/ante"
)

type input struct {
	payFee          sdk.Coin
	gasLimit        int64
	currentGasPrice sdk.DecCoin
}

// TestGetTxPriority ensures that tx priority is properly bounded
func TestGetTxPriority(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		inputs := createRandomInput(t)

		priority := ante.GetTxPriority(inputs.payFee, inputs.gasLimit, inputs.currentGasPrice)
		require.GreaterOrEqual(t, priority, int64(0))
		require.LessOrEqual(t, priority, int64(math.MaxInt64))
	})
}

// CreateRandomInput returns a random inputs to the priority function.
func createRandomInput(t *rapid.T) input {
	denom := "skip"

	price := rapid.Int64Range(1, 1_000_000_000).Draw(t, "gas price")
	priceDec := sdkmath.LegacyNewDecWithPrec(price, 6)

	gasLimit := rapid.Int64Range(1_000_000, 1_000_000_000_000).Draw(t, "gas limit")

	if priceDec.MulInt64(gasLimit).GTE(sdkmath.LegacyNewDec(math.MaxInt64)) {
		t.Fatalf("not supposed to happen")
	}

	payFeeAmt := rapid.Int64Range(priceDec.MulInt64(gasLimit).TruncateInt64(), math.MaxInt64).Draw(t, "fee amount")

	return input{
		payFee:          sdk.NewCoin(denom, sdkmath.NewInt(payFeeAmt)),
		gasLimit:        gasLimit,
		currentGasPrice: sdk.NewDecCoinFromDec(denom, priceDec),
	}
}
