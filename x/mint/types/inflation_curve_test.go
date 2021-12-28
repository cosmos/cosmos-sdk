package types

import (
	"math"
	"math/big"
	"testing"
	"github.com/stretchr/testify/require"
)

func TestInflationMonotonicity(t *testing.T) {
	var max = new(big.Int).Mul(big.NewInt(300_000_000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	var peak = new(big.Int).Rsh(max, 1)
	// use a large prime number in case monotonicity only happens for high trailing
    // zero numbers or some other regular condition
	var numSteps = int64(104729)
	var step = new(big.Int).Div(max, big.NewInt(numSteps))
	// a small amount large enough that neighboring inflation rates are not equal,
    // check that the curve is monotonic and smooth with increments of this
	var small = new(big.Int).Exp(big.NewInt(10), big.NewInt(18 - 7), nil)
	var anom = new(big.Int)
	var maxDiff = big.NewInt(100000)
    // second order smoothness should be very smooth
	var maxSecondDiff = big.NewInt(10)
	for i := int64(0); i < numSteps; i++ {
		var beforePeak0 = anom.Cmp(peak) == -1
		var result0 = globalInflationCurve.calculateInflationBinary(anom)
		anom.Add(anom, small)
		var result1 = globalInflationCurve.calculateInflationBinary(anom)
		anom.Add(anom, small)
		var result2 = globalInflationCurve.calculateInflationBinary(anom)
		var beforePeak2 = anom.Cmp(peak) == -1
		anom.Add(anom, step)

		// ignore the peak
		if beforePeak0 == beforePeak2 {
			if beforePeak0 {
				// monotonically increasing
				require.True(t, result0.Cmp(result1) == -1, "%v not less than %v", result0, result1)
				require.True(t, result1.Cmp(result2) == -1, "%v not less than %v", result1, result2)
			} else {
				// monotonically decreasing
				require.True(t, result0.Cmp(result1) == 1, "%v not greater than %v", result0, result1)
				require.True(t, result1.Cmp(result2) == 1, "%v not greater than %v", result1, result2)
			}
			var finiteDiff0 = new(big.Int).Sub(result0, result1)
			finiteDiff0.Abs(finiteDiff0)
			var finiteDiff1 = new(big.Int).Sub(result1, result2)
			finiteDiff1.Abs(finiteDiff1)
			require.True(t, finiteDiff0.Cmp(maxDiff) == -1, "%v not under maxDiff", finiteDiff0)
			require.True(t, finiteDiff1.Cmp(maxDiff) == -1, "%v not under maxDiff", finiteDiff1)
			var secondDiff = new(big.Int).Sub(finiteDiff0, finiteDiff1)
			secondDiff.Abs(secondDiff)
			require.True(t, secondDiff.Cmp(maxSecondDiff) == -1, "%v not under maxSecondDiff", secondDiff)
		}
	}
}

// Test using floating point that the curve is approximately what it should be
func TestInflationApproximate(t *testing.T) {
	var max = new(big.Int).Mul(big.NewInt(300_000_000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	var numSteps = int64(104729)
	var step = new(big.Int).Div(max, big.NewInt(numSteps))
	var anom = new(big.Int)
	var maxDiff = new(big.Int).Exp(big.NewInt(2), big.NewInt(64), nil)
	maxDiff.Div(maxDiff, new(big.Int).Exp(big.NewInt(10), big.NewInt(15), nil))
	for i := int64(0); i < numSteps; i++ {
		var inputTmp = new(big.Int).Set(anom)
		// 3e26 has 88 significant bits, shift by 36 to get 52 bits which will be
        // suitable for getting the precision into 64 bit IEEE floats
		inputTmp.Rsh(inputTmp, 36)
		var anomF = float64(inputTmp.Uint64())
        // undo the shift on the float side
		anomF *= math.Pow(float64(2.0), 36)
		var resultF = math.Exp(math.Pow(anomF - 150_000_000.0e18, 2) / (-float64(2.0) * math.Pow(float64(50_000_000.0e18), 2)))

		// conversion to 64 bit fracint
		resultF *= math.Pow(float64(2.0), 52);
		var outputTmp = new(big.Int).SetUint64(uint64(resultF))
		outputTmp.Lsh(outputTmp, 12)

		var result = globalInflationCurve.calculateInflationBinary(anom)

		var diff = new(big.Int).Sub(outputTmp, result)
		diff.Abs(diff)
		require.True(t, diff.Cmp(maxDiff) == -1, "%v not under maxDiff", diff)

		anom.Add(anom, step)
	}
}
