package types_test

import (
	"math/rand"
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/x/feemarket/types"
)

var OneHundred = math.LegacyMustNewDecFromStr("100")

func TestState_Update(t *testing.T) {
	t.Run("can add to window", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		err := state.Update(100, params)
		require.NoError(t, err)
		require.Equal(t, uint64(100), state.Window[0])
	})

	t.Run("can add several txs to window", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		err := state.Update(100, params)
		require.NoError(t, err)
		require.Equal(t, uint64(100), state.Window[0])

		err = state.Update(200, params)
		require.NoError(t, err)
		require.Equal(t, uint64(300), state.Window[0])
	})

	t.Run("errors when it exceeds max block utilization", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		err := state.Update(params.MaxBlockUtilization+1, params)
		require.Error(t, err)
	})

	t.Run("can update with several blocks in default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		err := state.Update(100, params)
		require.NoError(t, err)
		require.Equal(t, uint64(100), state.Window[0])

		state.IncrementHeight()

		err = state.Update(200, params)
		require.NoError(t, err)
		require.Equal(t, uint64(200), state.Window[0])

		err = state.Update(300, params)
		require.NoError(t, err)
		require.Equal(t, uint64(500), state.Window[0])
	})

	t.Run("can update with several blocks in default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		params := types.DefaultAIMDParams()

		err := state.Update(100, params)
		require.NoError(t, err)
		require.Equal(t, uint64(100), state.Window[0])

		state.IncrementHeight()

		err = state.Update(200, params)
		require.NoError(t, err)
		require.Equal(t, uint64(200), state.Window[1])

		state.IncrementHeight()

		err = state.Update(300, params)
		require.NoError(t, err)
		require.Equal(t, uint64(300), state.Window[2])

		state.IncrementHeight()

		err = state.Update(400, params)
		require.NoError(t, err)
		require.Equal(t, uint64(400), state.Window[3])
	})

	t.Run("correctly wraps around with aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		params := types.DefaultAIMDParams()
		state.Window = make([]uint64, 3)

		err := state.Update(100, params)
		require.NoError(t, err)
		require.Equal(t, uint64(100), state.Window[0])

		state.IncrementHeight()

		err = state.Update(200, params)
		require.NoError(t, err)
		require.Equal(t, uint64(200), state.Window[1])

		state.IncrementHeight()

		err = state.Update(300, params)
		require.NoError(t, err)
		require.Equal(t, uint64(300), state.Window[2])

		state.IncrementHeight()
		require.Equal(t, uint64(0), state.Window[0])

		err = state.Update(400, params)
		require.NoError(t, err)
		require.Equal(t, uint64(400), state.Window[0])
		require.Equal(t, uint64(200), state.Window[1])
		require.Equal(t, uint64(300), state.Window[2])
	})
}

func TestState_UpdateBaseGasPrice(t *testing.T) {
	t.Run("empty block with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		state.BaseGasPrice = math.LegacyMustNewDecFromStr("1000")
		params.MinBaseGasPrice = math.LegacyMustNewDecFromStr("125")

		newBaseGasPrice := state.UpdateBaseGasPrice(params)
		expectedBaseGasPrice := math.LegacyMustNewDecFromStr("875")
		require.True(t, expectedBaseGasPrice.Equal(newBaseGasPrice))
	})

	t.Run("target block with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		state.BaseGasPrice = math.LegacyMustNewDecFromStr("1000")
		params.MinBaseGasPrice = math.LegacyMustNewDecFromStr("125")

		state.Window[0] = params.TargetBlockUtilization()

		newBaseGasPrice := state.UpdateBaseGasPrice(params)
		expectedBaseGasPrice := math.LegacyMustNewDecFromStr("1000")
		require.True(t, expectedBaseGasPrice.Equal(newBaseGasPrice))
	})

	t.Run("full block with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		state.BaseGasPrice = math.LegacyMustNewDecFromStr("1000")
		params.MinBaseGasPrice = math.LegacyMustNewDecFromStr("125")

		state.Window[0] = params.MaxBlockUtilization

		newBaseGasPrice := state.UpdateBaseGasPrice(params)
		expectedBaseGasPrice := math.LegacyMustNewDecFromStr("1125")
		require.True(t, expectedBaseGasPrice.Equal(newBaseGasPrice))
	})

	t.Run("empty block with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		params := types.DefaultAIMDParams()

		state.BaseGasPrice = math.LegacyMustNewDecFromStr("1000")
		params.MinBaseGasPrice = math.LegacyMustNewDecFromStr("125")
		state.LearningRate = math.LegacyMustNewDecFromStr("0.125")

		state.UpdateLearningRate(params)
		newBaseGasPrice := state.UpdateBaseGasPrice(params)

		expectedBaseGasPrice := math.LegacyMustNewDecFromStr("850")
		require.True(t, expectedBaseGasPrice.Equal(newBaseGasPrice))
	})

	t.Run("target block with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		params := types.DefaultAIMDParams()

		state.BaseGasPrice = math.LegacyMustNewDecFromStr("1000")
		params.MinBaseGasPrice = math.LegacyMustNewDecFromStr("125")
		state.LearningRate = math.LegacyMustNewDecFromStr("0.125")

		for i := 0; i < len(state.Window); i++ {
			state.Window[i] = params.TargetBlockUtilization()
		}

		state.UpdateLearningRate(params)
		newBaseGasPrice := state.UpdateBaseGasPrice(params)

		expectedBaseGasPrice := math.LegacyMustNewDecFromStr("1000")
		require.True(t, expectedBaseGasPrice.Equal(newBaseGasPrice))
	})

	t.Run("full blocks with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		params := types.DefaultAIMDParams()

		state.BaseGasPrice = math.LegacyMustNewDecFromStr("1000")
		params.MinBaseGasPrice = math.LegacyMustNewDecFromStr("125")
		state.LearningRate = math.LegacyMustNewDecFromStr("0.125")

		for i := 0; i < len(state.Window); i++ {
			state.Window[i] = params.MaxBlockUtilization
		}

		state.UpdateLearningRate(params)
		newBaseGasPrice := state.UpdateBaseGasPrice(params)

		expectedBaseGasPrice := math.LegacyMustNewDecFromStr("1150")
		require.True(t, expectedBaseGasPrice.Equal(newBaseGasPrice))
	})

	t.Run("never goes below min base gas price with default eip1599", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		// Empty block
		newBaseGasPrice := state.UpdateBaseGasPrice(params)
		expectedBaseGasPrice := params.MinBaseGasPrice
		require.True(t, expectedBaseGasPrice.Equal(newBaseGasPrice))
	})

	t.Run("never goes below min base gas price with default aimd eip1599", func(t *testing.T) {
		state := types.DefaultAIMDState()
		params := types.DefaultAIMDParams()

		// Empty blocks
		newBaseGasPrice := state.UpdateBaseGasPrice(params)
		expectedBaseGasPrice := params.MinBaseGasPrice
		require.True(t, expectedBaseGasPrice.Equal(newBaseGasPrice))
	})

	t.Run("empty blocks with aimd eip1559 with a delta", func(t *testing.T) {
		// Instantiate the params with a delta.
		params := types.DefaultAIMDParams()

		paramsWithDelta := types.DefaultAIMDParams()
		delta := math.LegacyNewDec(10)
		paramsWithDelta.Delta = delta

		// Empty blocks
		state := types.DefaultAIMDState()
		state.BaseGasPrice = state.BaseGasPrice.Mul(math.LegacyMustNewDecFromStr("10"))
		lr := state.UpdateLearningRate(params)
		bgs := state.UpdateBaseGasPrice(params)

		state = types.DefaultAIMDState()
		state.BaseGasPrice = state.BaseGasPrice.Mul(math.LegacyMustNewDecFromStr("10"))
		lrWithDelta := state.UpdateLearningRate(paramsWithDelta)
		bgsWithDelta := state.UpdateBaseGasPrice(paramsWithDelta)

		// Ensure that the learning rate is the same.
		require.Equal(t, lr, lrWithDelta)

		// Ensure that the base gas price is less with the delta.
		require.True(t, bgsWithDelta.LT(bgs))
	})

	t.Run("full blocks with aimd eip1559 with a delta", func(t *testing.T) {
		// Instantiate the params with a delta.
		params := types.DefaultAIMDParams()

		paramsWithDelta := types.DefaultAIMDParams()
		delta := math.LegacyNewDec(10)
		paramsWithDelta.Delta = delta

		// Empty blocks
		state := types.DefaultAIMDState()
		state.BaseGasPrice = state.BaseGasPrice.Mul(math.LegacyMustNewDecFromStr("10"))
		for i := 0; i < len(state.Window); i++ {
			state.Window[i] = params.MaxBlockUtilization
		}

		lr := state.UpdateLearningRate(params)
		bgs := state.UpdateBaseGasPrice(params)

		state = types.DefaultAIMDState()
		state.BaseGasPrice = state.BaseGasPrice.Mul(math.LegacyMustNewDecFromStr("10"))
		for i := 0; i < len(state.Window); i++ {
			state.Window[i] = params.MaxBlockUtilization
		}

		lrWithDelta := state.UpdateLearningRate(paramsWithDelta)
		bgsWithDelta := state.UpdateBaseGasPrice(paramsWithDelta)

		// Ensure that the learning rate is the same.
		require.Equal(t, lr, lrWithDelta)

		// Ensure that the base gas price is greater with the delta.
		require.True(t, bgsWithDelta.GT(bgs))
	})

	t.Run("target blocks with aimd eip1559 with a delta", func(t *testing.T) {
		// Instantiate the params with a delta.
		params := types.DefaultAIMDParams()

		paramsWithDelta := types.DefaultAIMDParams()
		delta := math.LegacyNewDec(10)
		paramsWithDelta.Delta = delta

		// Empty blocks
		state := types.DefaultAIMDState()
		state.BaseGasPrice = state.BaseGasPrice.Mul(math.LegacyMustNewDecFromStr("10"))
		for i := 0; i < len(state.Window); i++ {
			state.Window[i] = params.TargetBlockUtilization()
		}

		lr := state.UpdateLearningRate(params)
		bgs := state.UpdateBaseGasPrice(params)

		state = types.DefaultAIMDState()
		state.BaseGasPrice = state.BaseGasPrice.Mul(math.LegacyMustNewDecFromStr("10"))
		for i := 0; i < len(state.Window); i++ {
			state.Window[i] = params.TargetBlockUtilization()
		}

		lrWithDelta := state.UpdateLearningRate(paramsWithDelta)
		bgsWithDelta := state.UpdateBaseGasPrice(paramsWithDelta)

		// Ensure that the learning rate is the same.
		require.Equal(t, lr, lrWithDelta)

		// Ensure that the base gas prices are equal.
		require.Equal(t, bgs, bgsWithDelta)
	})

	t.Run("half target block size with aimd eip1559 with a delta", func(t *testing.T) {
		state := types.DefaultAIMDState()
		state.Window = make([]uint64, 1)
		state.BaseGasPrice = state.BaseGasPrice.Mul(math.LegacyMustNewDecFromStr("10"))
		prevBF := state.BaseGasPrice

		// Instantiate the params with a delta.
		params := types.DefaultAIMDParams()
		params.Window = 1
		params.Delta = math.LegacyNewDec(10)

		// 1/4th of the window is full.
		state.Window[0] = params.TargetBlockUtilization() / 2

		prevLR := state.LearningRate
		lr := state.UpdateLearningRate(params)
		bgs := state.UpdateBaseGasPrice(params)

		expectedUtilization := math.LegacyMustNewDecFromStr("-0.5")
		expectedLR := prevLR.Add(params.Alpha)
		expectedLRAdjustment := (expectedLR.Mul(expectedUtilization)).Add(math.LegacyOneDec())

		expectedNetUtilization := math.LegacyNewDec(-1 * int64(params.TargetBlockUtilization()) / 2)
		deltaDiff := expectedNetUtilization.Mul(params.Delta)
		expectedGasPrice := prevBF.Mul(expectedLRAdjustment).Add(deltaDiff)

		require.Equal(t, expectedLR, lr)
		require.Equal(t, expectedGasPrice, bgs)
	})

	t.Run("3/4 max block size with aimd eip1559 with a delta", func(t *testing.T) {
		state := types.DefaultAIMDState()
		state.Window = make([]uint64, 1)
		state.BaseGasPrice = state.BaseGasPrice.Mul(math.LegacyMustNewDecFromStr("10"))
		prevBGS := state.BaseGasPrice

		// Instantiate the params with a delta.
		params := types.DefaultAIMDParams()
		params.Window = 1
		params.Delta = math.LegacyNewDec(10)

		// 1/4th of the window is full.
		state.Window[0] = params.MaxBlockUtilization / 4 * 3

		prevLR := state.LearningRate
		lr := state.UpdateLearningRate(params)
		bgs := state.UpdateBaseGasPrice(params)

		expectedUtilization := math.LegacyMustNewDecFromStr("0.5")
		expectedLR := prevLR.Add(params.Alpha)
		expectedLRAdjustment := (expectedLR.Mul(expectedUtilization)).Add(math.LegacyOneDec())

		expectedNetUtilization := math.LegacyNewDec(int64(params.MaxBlockUtilization) / 4)
		deltaDiff := expectedNetUtilization.Mul(params.Delta)
		expectedGasPrice := prevBGS.Mul(expectedLRAdjustment).Add(deltaDiff)
		require.Equal(t, expectedLR, lr)
		require.Equal(t, expectedGasPrice, bgs)
	})
}

func TestState_UpdateLearningRate(t *testing.T) {
	t.Run("empty block with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		state.UpdateLearningRate(params)
		expectedLearningRate := math.LegacyMustNewDecFromStr("0.125")
		require.True(t, expectedLearningRate.Equal(state.LearningRate))
	})

	t.Run("target block with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		state.Window[0] = params.TargetBlockUtilization()

		state.UpdateLearningRate(params)
		expectedLearningRate := math.LegacyMustNewDecFromStr("0.125")
		require.True(t, expectedLearningRate.Equal(state.LearningRate))
	})

	t.Run("full block with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		state.Window[0] = params.MaxBlockUtilization

		state.UpdateLearningRate(params)
		expectedLearningRate := math.LegacyMustNewDecFromStr("0.125")
		require.True(t, expectedLearningRate.Equal(state.LearningRate))
	})

	t.Run("between 0 and target with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		state.Window[0] = 50000

		state.UpdateLearningRate(params)
		expectedLearningRate := math.LegacyMustNewDecFromStr("0.125")
		require.True(t, expectedLearningRate.Equal(state.LearningRate))
	})

	t.Run("between target and max with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		state.Window[0] = params.TargetBlockUtilization() + 50000

		state.UpdateLearningRate(params)
		expectedLearningRate := math.LegacyMustNewDecFromStr("0.125")
		require.True(t, expectedLearningRate.Equal(state.LearningRate))
	})

	t.Run("random value with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		randomValue := rand.Int63n(1000000000)
		state.Window[0] = uint64(randomValue)

		state.UpdateLearningRate(params)
		expectedLearningRate := math.LegacyMustNewDecFromStr("0.125")
		require.True(t, expectedLearningRate.Equal(state.LearningRate))
	})

	t.Run("empty block with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		params := types.DefaultAIMDParams()

		state.UpdateLearningRate(params)
		expectedLearningRate := params.MinLearningRate.Add(params.Alpha)
		require.True(t, expectedLearningRate.Equal(state.LearningRate))
	})

	t.Run("target block with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		defaultLR := math.LegacyMustNewDecFromStr("0.125")
		state.LearningRate = defaultLR

		params := types.DefaultAIMDParams()

		for i := 0; i < len(state.Window); i++ {
			state.Window[i] = params.TargetBlockUtilization()
		}

		state.UpdateLearningRate(params)
		expectedLearningRate := defaultLR.Mul(params.Beta)
		require.True(t, expectedLearningRate.Equal(state.LearningRate))
	})

	t.Run("full blocks with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		defaultLR := math.LegacyMustNewDecFromStr("0.125")
		state.LearningRate = defaultLR

		params := types.DefaultAIMDParams()

		for i := 0; i < len(state.Window); i++ {
			state.Window[i] = params.MaxBlockUtilization
		}

		state.UpdateLearningRate(params)
		expectedLearningRate := defaultLR.Add(params.Alpha)
		require.True(t, expectedLearningRate.Equal(state.LearningRate))
	})

	t.Run("varying blocks with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		defaultLR := math.LegacyMustNewDecFromStr("0.125")
		state.LearningRate = defaultLR

		params := types.DefaultAIMDParams()

		for i := 0; i < len(state.Window); i++ {
			if i%2 == 0 {
				state.Window[i] = params.MaxBlockUtilization
			} else {
				state.Window[i] = 0
			}
		}

		state.UpdateLearningRate(params)
		expectedLearningRate := defaultLR.Mul(params.Beta)
		require.True(t, expectedLearningRate.Equal(state.LearningRate))
	})

	t.Run("exceeds threshold with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		defaultLR := math.LegacyMustNewDecFromStr("0.125")
		state.LearningRate = defaultLR

		params := types.DefaultAIMDParams()

		for i := 0; i < len(state.Window); i++ {
			if i%2 == 0 {
				state.Window[i] = params.MaxBlockUtilization
			} else {
				state.Window[i] = params.TargetBlockUtilization() + 1
			}
		}

		state.UpdateLearningRate(params)
		expectedLearningRate := defaultLR.Add(params.Alpha)
		require.True(t, expectedLearningRate.Equal(state.LearningRate))
	})
}

func TestState_GetNetUtilization(t *testing.T) {
	t.Run("empty block with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		netUtilization := state.GetNetUtilization(params)
		expectedUtilization := math.NewInt(0).Sub(math.NewIntFromUint64(params.TargetBlockUtilization()))
		require.True(t, expectedUtilization.Equal(netUtilization))
	})

	t.Run("target block with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		state.Window[0] = params.TargetBlockUtilization()

		netUtilization := state.GetNetUtilization(params)
		expectedUtilization := math.NewInt(0)
		require.True(t, expectedUtilization.Equal(netUtilization))
	})

	t.Run("full block with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		state.Window[0] = params.MaxBlockUtilization

		netUtilization := state.GetNetUtilization(params)
		expectedUtilization := math.NewIntFromUint64(params.MaxBlockUtilization - params.TargetBlockUtilization())
		require.True(t, expectedUtilization.Equal(netUtilization))
	})

	t.Run("empty block with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		params := types.DefaultParams()

		netUtilization := state.GetNetUtilization(params)

		multiple := math.NewIntFromUint64(uint64(len(state.Window)))
		expectedUtilization := math.NewInt(0).Sub(math.NewIntFromUint64(params.TargetBlockUtilization())).Mul(multiple)
		require.True(t, expectedUtilization.Equal(netUtilization))
	})

	t.Run("full blocks with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		params := types.DefaultAIMDParams()

		for i := 0; i < len(state.Window); i++ {
			state.Window[i] = params.MaxBlockUtilization
		}

		netUtilization := state.GetNetUtilization(params)

		multiple := math.NewIntFromUint64(uint64(len(state.Window)))
		expectedUtilization := math.NewIntFromUint64(params.MaxBlockUtilization).Sub(math.NewIntFromUint64(params.TargetBlockUtilization())).Mul(multiple)
		require.True(t, expectedUtilization.Equal(netUtilization))
	})

	t.Run("varying blocks with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		params := types.DefaultAIMDParams()

		for i := 0; i < len(state.Window); i++ {
			if i%2 == 0 {
				state.Window[i] = params.MaxBlockUtilization
			} else {
				state.Window[i] = 0
			}
		}

		netUtilization := state.GetNetUtilization(params)
		expectedUtilization := math.ZeroInt()
		require.True(t, expectedUtilization.Equal(netUtilization))
	})

	t.Run("exceeds target rate with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		params := types.DefaultAIMDParams()

		for i := 0; i < len(state.Window); i++ {
			if i%2 == 0 {
				state.Window[i] = params.MaxBlockUtilization
			} else {
				state.Window[i] = params.TargetBlockUtilization()
			}
		}

		netUtilization := state.GetNetUtilization(params)
		first := math.NewIntFromUint64(params.MaxBlockUtilization).Mul(math.NewIntFromUint64(params.Window / 2))
		second := math.NewIntFromUint64(params.TargetBlockUtilization()).Mul(math.NewIntFromUint64(params.Window / 2))
		expectedUtilization := first.Add(second).Sub(math.NewIntFromUint64(params.TargetBlockUtilization()).Mul(math.NewIntFromUint64(params.Window)))
		require.True(t, expectedUtilization.Equal(netUtilization))
	})

	t.Run("state with 4 entries in window with different updates", func(t *testing.T) {
		state := types.DefaultAIMDState()
		state.Window = make([]uint64, 4)

		params := types.DefaultAIMDParams()
		params.MaxBlockUtilization = 200

		state.Window[0] = 100
		state.Window[1] = 200
		state.Window[2] = 0
		state.Window[3] = 50

		netUtilization := state.GetNetUtilization(params)
		expectedUtilization := math.NewIntFromUint64(50).Mul(math.NewInt(-1))
		require.True(t, expectedUtilization.Equal(netUtilization))
	})

	t.Run("state with 4 entries in window with monotonically increasing updates", func(t *testing.T) {
		state := types.DefaultAIMDState()
		state.Window = make([]uint64, 4)

		params := types.DefaultAIMDParams()
		params.MaxBlockUtilization = 200

		state.Window[0] = 0
		state.Window[1] = 25
		state.Window[2] = 50
		state.Window[3] = 75

		netUtilization := state.GetNetUtilization(params)
		expectedUtilization := math.NewIntFromUint64(250).Mul(math.NewInt(-1))
		require.True(t, expectedUtilization.Equal(netUtilization))
	})
}

func TestState_GetAverageUtilization(t *testing.T) {
	t.Run("empty block with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		avgUtilization := state.GetAverageUtilization(params)
		expectedUtilization := math.LegacyZeroDec()
		require.True(t, expectedUtilization.Equal(avgUtilization))
	})

	t.Run("target block with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		state.Window[0] = params.TargetBlockUtilization()

		avgUtilization := state.GetAverageUtilization(params)
		expectedUtilization := math.LegacyMustNewDecFromStr("0.5")
		require.True(t, expectedUtilization.Equal(avgUtilization))
	})

	t.Run("full block with default eip-1559", func(t *testing.T) {
		state := types.DefaultState()
		params := types.DefaultParams()

		state.Window[0] = params.MaxBlockUtilization

		avgUtilization := state.GetAverageUtilization(params)
		expectedUtilization := math.LegacyMustNewDecFromStr("1.0")
		require.True(t, expectedUtilization.Equal(avgUtilization))
	})

	t.Run("empty block with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		params := types.DefaultAIMDParams()

		avgUtilization := state.GetAverageUtilization(params)
		expectedUtilization := math.LegacyZeroDec()
		require.True(t, expectedUtilization.Equal(avgUtilization))
	})

	t.Run("target block with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		params := types.DefaultAIMDParams()

		for i := 0; i < len(state.Window); i++ {
			state.Window[i] = params.TargetBlockUtilization()
		}

		avgUtilization := state.GetAverageUtilization(params)
		expectedUtilization := math.LegacyMustNewDecFromStr("0.5")
		require.True(t, expectedUtilization.Equal(avgUtilization))
	})

	t.Run("full blocks with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		params := types.DefaultAIMDParams()

		for i := 0; i < len(state.Window); i++ {
			state.Window[i] = params.MaxBlockUtilization
		}

		avgUtilization := state.GetAverageUtilization(params)
		expectedUtilization := math.LegacyMustNewDecFromStr("1.0")
		require.True(t, expectedUtilization.Equal(avgUtilization))
	})

	t.Run("varying blocks with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		params := types.DefaultAIMDParams()

		for i := 0; i < len(state.Window); i++ {
			if i%2 == 0 {
				state.Window[i] = params.MaxBlockUtilization
			} else {
				state.Window[i] = 0
			}
		}

		avgUtilization := state.GetAverageUtilization(params)
		expectedUtilization := math.LegacyMustNewDecFromStr("0.5")
		require.True(t, expectedUtilization.Equal(avgUtilization))
	})

	t.Run("exceeds target rate with default aimd eip-1559", func(t *testing.T) {
		state := types.DefaultAIMDState()
		params := types.DefaultAIMDParams()

		for i := 0; i < len(state.Window); i++ {
			if i%2 == 0 {
				state.Window[i] = params.MaxBlockUtilization
			} else {
				state.Window[i] = params.TargetBlockUtilization()
			}
		}

		avgUtilization := state.GetAverageUtilization(params)
		expectedUtilization := math.LegacyMustNewDecFromStr("0.75")
		require.True(t, expectedUtilization.Equal(avgUtilization))
	})

	t.Run("state with 4 entries in window with different updates", func(t *testing.T) {
		state := types.DefaultAIMDState()
		state.Window = make([]uint64, 4)

		params := types.DefaultAIMDParams()
		params.MaxBlockUtilization = 200

		state.Window[0] = 100
		state.Window[1] = 200
		state.Window[2] = 0
		state.Window[3] = 50

		avgUtilization := state.GetAverageUtilization(params)
		expectedUtilization := math.LegacyMustNewDecFromStr("0.4375")
		require.True(t, expectedUtilization.Equal(avgUtilization))
	})

	t.Run("state with 4 entries in window with monotonically increasing updates", func(t *testing.T) {
		state := types.DefaultAIMDState()
		state.Window = make([]uint64, 4)

		params := types.DefaultAIMDParams()
		params.Window = 4
		params.MaxBlockUtilization = 200

		state.Window[0] = 0
		state.Window[1] = 25
		state.Window[2] = 50
		state.Window[3] = 75

		avgUtilization := state.GetAverageUtilization(params)
		expectedUtilization := math.LegacyMustNewDecFromStr("0.1875")
		require.True(t, expectedUtilization.Equal(avgUtilization))
	})
}

func TestState_ValidateBasic(t *testing.T) {
	testCases := []struct {
		name      string
		state     types.State
		expectErr bool
	}{
		{
			name:      "default base EIP-1559 state",
			state:     types.DefaultState(),
			expectErr: false,
		},
		{
			name:      "default AIMD EIP-1559 state",
			state:     types.DefaultAIMDState(),
			expectErr: false,
		},
		{
			name:      "invalid window",
			state:     types.State{},
			expectErr: true,
		},
		{
			name: "invalid negative base gas price",
			state: types.State{
				Window:       make([]uint64, 1),
				BaseGasPrice: math.LegacyMustNewDecFromStr("-1"),
			},
			expectErr: true,
		},
		{
			name: "invalid learning rate",
			state: types.State{
				Window:       make([]uint64, 1),
				BaseGasPrice: math.LegacyMustNewDecFromStr("1"),
				LearningRate: math.LegacyMustNewDecFromStr("-1.0"),
			},
			expectErr: true,
		},
		{
			name: "valid other state",
			state: types.State{
				Window:       make([]uint64, 1),
				BaseGasPrice: math.LegacyMustNewDecFromStr("1"),
				LearningRate: math.LegacyMustNewDecFromStr("0.5"),
			},
			expectErr: false,
		},
		{
			name: "invalid zero base gas price",
			state: types.State{
				Window:       make([]uint64, 1),
				BaseGasPrice: math.LegacyZeroDec(),
				LearningRate: math.LegacyMustNewDecFromStr("0.5"),
			},
			expectErr: true,
		},
		{
			name: "invalid zero learning rate",
			state: types.State{
				Window:       make([]uint64, 1),
				BaseGasPrice: math.LegacyMustNewDecFromStr("1"),
				LearningRate: math.LegacyZeroDec(),
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.state.ValidateBasic()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
