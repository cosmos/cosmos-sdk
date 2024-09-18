package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/x/feemarket/types"
)

func FuzzDefaultFeeMarket(f *testing.F) {
	testCases := []uint64{
		0,
		1_000,
		10_000,
		100_000,
		1_000_000,
		10_000_000,
		100_000_000,
	}

	for _, tc := range testCases {
		f.Add(tc)
	}

	defaultLR := math.LegacyMustNewDecFromStr("0.125")

	// Default fee market.
	f.Fuzz(func(t *testing.T, blockGasUsed uint64) {
		state := types.DefaultState()
		params := types.DefaultParams()

		params.MinBaseGasPrice = math.LegacyMustNewDecFromStr("100")
		state.BaseGasPrice = math.LegacyMustNewDecFromStr("200")
		err := state.Update(blockGasUsed, params)

		if blockGasUsed > params.MaxBlockUtilization {
			require.Error(t, err)
			return
		}

		require.NoError(t, err)
		require.Equal(t, blockGasUsed, state.Window[state.Index])

		// Ensure the learning rate is always the default learning rate.
		lr := state.UpdateLearningRate(
			params,
		)
		require.Equal(t, defaultLR, lr)

		oldFee := state.BaseGasPrice
		newFee := state.UpdateBaseGasPrice(params)

		if blockGasUsed > params.TargetBlockUtilization() {
			require.True(t, newFee.GT(oldFee))
		} else {
			require.True(t, newFee.LT(oldFee))
		}
	})
}

func FuzzAIMDFeeMarket(f *testing.F) {
	testCases := []uint64{
		0,
		1_000,
		10_000,
		100_000,
		1_000_000,
		10_000_000,
		100_000_000,
	}

	for _, tc := range testCases {
		f.Add(tc)
	}

	// Fee market with adjustable learning rate.
	f.Fuzz(func(t *testing.T, blockGasUsed uint64) {
		state := types.DefaultAIMDState()
		params := types.DefaultAIMDParams()
		params.MinBaseGasPrice = math.LegacyMustNewDecFromStr("100")
		state.BaseGasPrice = math.LegacyMustNewDecFromStr("200")
		state.Window = make([]uint64, 1)
		err := state.Update(blockGasUsed, params)

		if blockGasUsed > params.MaxBlockUtilization {
			require.Error(t, err)
			return
		}

		require.NoError(t, err)
		require.Equal(t, blockGasUsed, state.Window[state.Index])

		oldFee := state.BaseGasPrice
		newFee := state.UpdateBaseGasPrice(params)

		if blockGasUsed > params.TargetBlockUtilization() {
			require.True(t, newFee.GT(oldFee))
		} else {
			require.True(t, newFee.LT(oldFee))
		}
	})
}
