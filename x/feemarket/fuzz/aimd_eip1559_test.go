package fuzz_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"cosmossdk.io/x/feemarket/types"
)

// TestAIMDLearningRate ensures that the additive increase
// multiplicative decrease learning rate algorithm correctly
// adjusts the learning rate. In particular, if the block
// utilization is greater than theta or less than 1 - theta, then
// the learning rate is increased by the additive increase
// parameter. Otherwise, the learning rate is decreased by
// the multiplicative decrease parameter.
func TestAIMDLearningRate(t *testing.T) {
	t.Parallel()

	rapid.Check(t, func(t *rapid.T) {
		state := types.DefaultAIMDState()
		window := rapid.Int64Range(1, 50).Draw(t, "window")
		state.Window = make([]uint64, window)

		params := CreateRandomAIMDParams(t)

		// Randomly generate the block utilization.
		numBlocks := rapid.Uint64Range(0, 1000).Draw(t, "num_blocks")
		gasGen := rapid.Uint64Range(0, params.MaxBlockUtilization)

		// Update the fee market.
		for i := uint64(0); i < numBlocks; i++ {
			blockUtilization := gasGen.Draw(t, "gas")
			prevLearningRate := state.LearningRate

			// Update the fee market.
			if err := state.Update(blockUtilization, params); err != nil {
				t.Fatalf("block update errors: %v", err)
			}

			// Update the learning rate.
			lr := state.UpdateLearningRate(params)
			utilization := state.GetAverageUtilization(params)

			// Ensure that the learning rate is always bounded.
			require.True(t, lr.GTE(params.MinLearningRate))
			require.True(t, lr.LTE(params.MaxLearningRate))

			if utilization.LTE(params.Gamma) || utilization.GTE(math.LegacyOneDec().Sub(params.Gamma)) {
				require.True(t, lr.GTE(prevLearningRate))
			} else {
				require.True(t, lr.LTE(prevLearningRate))
			}

			// Update the current height.
			state.IncrementHeight()
		}
	})
}

// TestAIMDGasPrice ensures that the additive increase multiplicative
// decrease gas price adjustment algorithm correctly adjusts the base
// fee. In particular, the gas price should function the same as the
// default EIP-1559 base fee adjustment algorithm.
func TestAIMDGasPrice(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		state := types.DefaultAIMDState()
		window := rapid.Int64Range(1, 50).Draw(t, "window")
		state.Window = make([]uint64, window)

		params := CreateRandomAIMDParams(t)

		// Randomly generate the block utilization.
		numBlocks := rapid.Uint64Range(0, uint64(window)*10).Draw(t, "num_blocks")
		gasGen := rapid.Uint64Range(0, params.MaxBlockUtilization)

		// Update the fee market.
		for i := uint64(0); i < numBlocks; i++ {
			blockUtilization := gasGen.Draw(t, "gas")
			prevBaseGasPrice := state.BaseGasPrice

			if err := state.Update(blockUtilization, params); err != nil {
				t.Fatalf("block update errors: %v", err)
			}

			var total uint64
			for _, utilization := range state.Window {
				total += utilization
			}

			// Update the learning rate.
			lr := state.UpdateLearningRate(params)
			// Update the base gas price.

			var newPrice math.LegacyDec
			func() {
				defer func() {
					if rec := recover(); rec != nil {
						newPrice = params.MinBaseGasPrice
					}
				}()

				// Calculate the new base gasPrice with the learning rate adjustment.
				currentBlockSize := math.LegacyNewDecFromInt(math.NewIntFromUint64(state.Window[state.Index]))
				targetBlockSize := math.LegacyNewDecFromInt(math.NewIntFromUint64(params.TargetBlockUtilization()))
				utilization := (currentBlockSize.Sub(targetBlockSize)).Quo(targetBlockSize)

				// Truncate the learning rate adjustment to an integer.
				//
				// This is equivalent to
				// 1 + (learningRate * (currentBlockSize - targetBlockSize) / targetBlockSize)
				learningRateAdjustment := math.LegacyOneDec().Add(lr.Mul(utilization))

				// Calculate the delta adjustment.
				net := math.LegacyNewDecFromInt(state.GetNetUtilization(params)).Mul(params.Delta)

				// Update the base gasPrice.
				newPrice = prevBaseGasPrice.Mul(learningRateAdjustment).Add(net)
				// Ensure the base gasPrice is greater than the minimum base gasPrice.
				if newPrice.LT(params.MinBaseGasPrice) {
					newPrice = params.MinBaseGasPrice
				}
			}()

			state.UpdateBaseGasPrice(params)

			// Ensure that the minimum base fee is always less than the base gas price.
			require.True(t, params.MinBaseGasPrice.LTE(state.BaseGasPrice))

			require.Equal(t, newPrice, state.BaseGasPrice)

			// Update the current height.
			state.IncrementHeight()
		}
	})
}

// CreateRandomAIMDParams returns a random set of parameters for the AIMD
// EIP-1559 fee market implementation.
func CreateRandomAIMDParams(t *rapid.T) types.Params {
	a := rapid.Uint64Range(1, 1000).Draw(t, "alpha")
	alpha := math.LegacyNewDec(int64(a)).Quo(math.LegacyNewDec(1000))

	b := rapid.Uint64Range(50, 99).Draw(t, "beta")
	beta := math.LegacyNewDec(int64(b)).Quo(math.LegacyNewDec(100))

	g := rapid.Uint64Range(10, 50).Draw(t, "gamma")
	gamma := math.LegacyNewDec(int64(g)).Quo(math.LegacyNewDec(100))

	d := rapid.Uint64Range(1, 1000).Draw(t, "delta")
	delta := math.LegacyNewDec(int64(d)).Quo(math.LegacyNewDec(1000))

	targetBlockUtilization := rapid.Uint64Range(1, 30_000_000).Draw(t, "target_block_utilization")
	maxBlockUtilization := rapid.Uint64Range(targetBlockUtilization, targetBlockUtilization*5).Draw(t, "max_block_utilization")

	distributeFees := rapid.Bool().Draw(t, "distribute_fees")

	params := types.DefaultAIMDParams()
	params.Alpha = alpha
	params.Beta = beta
	params.Gamma = gamma
	params.Delta = delta
	params.MaxBlockUtilization = maxBlockUtilization
	params.DistributeFees = distributeFees

	return params
}
