package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"math"
	"math/big"
	"testing"
)

func mustNewDecFromStr(t *testing.T, str string) (d sdk.Dec) {
	d, err := sdk.NewDecFromStr(str)
	require.NoError(t, err)

	return d
}

func TestDecExp(t *testing.T) {
	tests := []struct {
		sdkDec sdk.Dec
		want   sdk.Dec
	}{
		{mustNewDecFromStr(t, "12.34567"), mustNewDecFromStr(t, "229962.147896858174618390")},
		{mustNewDecFromStr(t, "0.5"), mustNewDecFromStr(t, "1.648721270700128146")},
		{mustNewDecFromStr(t, "-0.5"), mustNewDecFromStr(t, "0.606530659712633423")},
		{mustNewDecFromStr(t, "-7.654321"), mustNewDecFromStr(t, "0.000473991580066384")},
		{mustNewDecFromStr(t, "-41"), mustNewDecFromStr(t, "0.000000000000000001")},
		{mustNewDecFromStr(t, "-42"), mustNewDecFromStr(t, "0.000000000000000000")},
		{mustNewDecFromStr(t, "88"), mustNewDecFromStr(t, "165163625499400185552832979626485876677.500000000000000000")},
	}
	for i, tc := range tests {
		got := globalFastExp.DecExp(tc.sdkDec)
		require.True(t, tc.want.Equal(*got), "Incorrect result on test case %d", i)
	}
	// test overflow case
	require.True(t, globalFastExp.DecExp(mustNewDecFromStr(t, "89")) == nil)
}

func TestInflationMonotonicity(t *testing.T) {
	var max = new(big.Int).Mul(big.NewInt(300_000_000), new(big.Int).Exp(big.NewInt(10), big.NewInt(sdk.Precision), nil))
	var peak = new(big.Int).Rsh(max, 1)
	// use a large prime number in case monotonicity only happens for high trailing
	// zero numbers or some other regular condition
	var numSteps = int64(104729)
	var step = new(big.Int).Div(max, big.NewInt(numSteps))
	// a small amount large enough that neighboring inflation rates are not equal,
	// check that the curve is monotonic and smooth with increments of this
	var small = new(big.Int).Exp(big.NewInt(10), big.NewInt(18-7), nil)
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
	var max = new(big.Int).Mul(big.NewInt(300_000_000), new(big.Int).Exp(big.NewInt(10), big.NewInt(sdk.Precision), nil))
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
		var resultF = math.Exp(math.Pow(anomF-150_000_000.0e18, 2) / (-float64(2.0) * math.Pow(float64(50_000_000.0e18), 2)))

		// conversion to 64 bit fracint
		resultF *= math.Pow(float64(2.0), 52)
		var outputTmp = new(big.Int).SetUint64(uint64(resultF))
		outputTmp.Lsh(outputTmp, 12)

		var result = globalInflationCurve.calculateInflationBinary(anom)

		// note that the limiting factor here is float precision
		var diff = new(big.Int).Sub(outputTmp, result)
		diff.Abs(diff)
		require.True(t, diff.Cmp(maxDiff) == -1, "%v not under maxDiff", diff)

		anom.Add(anom, step)
	}
}

// Test some specific values
func TestInflationExact(t *testing.T) {
	var aNomConversion = sdk.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(sdk.Precision), nil))
	var input = sdk.NewInt(50_000_000).Mul(aNomConversion)
	var actual = globalInflationCurve.CalculateInflationDec(input)
	// an independent high precision calculator says
	// the (10^18 adjusted) value is  135335283236612691.8939...
	var expected = sdk.NewDecWithPrec(135335283236612691, sdk.Precision)
	require.True(t, actual.BigInt().Cmp(expected.BigInt()) == 0)

	input = sdk.NewInt(123_456_789).Mul(aNomConversion)
	actual = globalInflationCurve.CalculateInflationDec(input)
	// perfect truncation again
	expected = sdk.NewDecWithPrec(868568860243501720, sdk.Precision)
	require.True(t, actual.BigInt().Cmp(expected.BigInt()) == 0)

	// test exactly at peak in case there is some edge case
	input = sdk.NewInt(150_000_000).Mul(aNomConversion)
	actual = globalInflationCurve.CalculateInflationDec(input)
	expected = sdk.NewDecWithPrec(1000000000000000000, sdk.Precision)
	require.True(t, actual.BigInt().Cmp(expected.BigInt()) == 0)

	input = sdk.NewInt(150_000_000).Mul(aNomConversion).Sub(sdk.NewInt(1))
	actual = globalInflationCurve.CalculateInflationDec(input)
	// truncation works as expected
	expected = sdk.NewDecWithPrec(999999999999999999, sdk.Precision)
	require.True(t, actual.BigInt().Cmp(expected.BigInt()) == 0)
}

// Fixed point integer combining a `big.Int` with a fixed point position. The numerical value is
// interpreted as `bits * 2.0^(-fp)`
type FP struct {
	bits *big.Int
	fp   int
}

func NewFP(bits *big.Int, fp int) FP {
	return FP{
		bits,
		fp,
	}
}

func CopyFP(x *FP) FP {
	return FP{
		bits: new(big.Int).Set(x.bits),
		fp:   x.fp,
	}
}

func NewZeroFP(fp int) FP {
	return NewFP(new(big.Int), fp)
}

// Returns an `FP` with value 1.0. Returns `nil` if `fp < 0`, because a representation of 1 is not possible.
func NewOneFP(fp int) *FP {
	if fp < 0 {
		return nil
	} else {
		var x = NewFP(new(big.Int).SetUint64(1), fp)
		x.bits.Lsh(x.bits, uint(fp))
		return &x
	}
}

func (x FP) String() string {
	return fmt.Sprintf("FP(bits: %v, fp: %v)", *x.bits, x.fp)
}

// If `x < 0.0`
func (x FP) IsNeg() bool {
	return x.bits.Sign() == -1
}

// If `x == 0.0`
func (x FP) IsZero() bool {
	return x.bits.Cmp(new(big.Int)) == 0
}

func (x *FP) Cmp(y *FP) int {
	return x.bits.Cmp(y.bits)
}

func (x *FP) Lt(y *FP) bool {
	var cmp = x.bits.Cmp(y.bits)
	return cmp == -1
}

func (x *FP) Le(y *FP) bool {
	var cmp = x.bits.Cmp(y.bits)
	return (cmp == -1) || (cmp == 0)
}

// negate
func (x *FP) Neg() {
	x.bits.Neg(x.bits)
}

// Increments `lhs` at the ULP level
func (lhs *FP) IncAssign() {
	lhs.bits.Add(lhs.bits, new(big.Int).SetUint64(1))
}

// Decrements `lhs` at the ULP level
func (lhs *FP) DecAssign() {
	lhs.bits.Sub(lhs.bits, new(big.Int).SetUint64(1))
}

// Add `rhs` to `lhs`. Assumes the fixed points are equal
func (lhs *FP) SameFpAddAssign(rhs *FP, o *bool) {
	if lhs.fp == rhs.fp {
		lhs.bits.Add(lhs.bits, rhs.bits)
	} else {
		*o = true
	}
}

// Subtract `rhs` from `lhs`. Sets `o` to `true` if the fixed points are not equal
func (lhs *FP) SameFpSubAssign(rhs *FP, o *bool) {
	if lhs.fp == rhs.fp {
		lhs.bits.Sub(lhs.bits, rhs.bits)
	} else {
		*o = true
	}
}

// Adds `rhs` to `lhs` if `b == false`, else Subtracts `rhs` from `lhs`
func (lhs *FP) SameFpNegAddAssign(b bool, rhs *FP, o *bool) {
	if lhs.fp != rhs.fp {
		*o = true
	} else if b {
		lhs.bits.Sub(lhs.bits, rhs.bits)
	} else {
		lhs.bits.Add(lhs.bits, rhs.bits)
	}
}

// Fixed point multiply-assign `lhs` by `rhs`
func (lhs *FP) FpMulAssign(rhs *FP) {
	lhs.bits.Mul(lhs.bits, rhs.bits)
	// the numerical fixed point is now at `lhs.fp + rhs.fp`, shift so that we get back to `lhs.fp`
	var abs = unsignedAbs(rhs.fp)
	if rhs.fp <= 0 {
		lhs.bits.Lsh(lhs.bits, abs)
	} else {
		lhs.bits.Rsh(lhs.bits, abs)
	}
}

// Divides `lhs` by an integer (no fixed point shifts applied)
func (lhs *FP) ShortUDivideAssign(rhs uint64) {
	lhs.bits.Div(lhs.bits, new(big.Int).SetUint64(rhs))
}

// Divides `duo` by `div` and sets `quo` to the quotient and `rem` to the remainder.
// Sets `o` to `true` if `quo` and `rem` do not have same fp types, or `div` is zero.
//
// This function works numerically by calculating
// `(duo * 2^duo.fp * 2^(-div.fp) * 2^quo.fp) / div`
func (quo *FP) FpDivideAssign(
	rem *FP,
	duo *FP,
	div *FP,
	o *bool,
) {
	if div.IsZero() || (quo.fp != rem.fp) {
		*o = true
		return
	}
	var num_fp = duo.fp - div.fp + quo.fp
	var duoC = CopyFP(duo)
	var divC = CopyFP(div)
	if num_fp < 0 {
		duoC.bits.Rsh(duoC.bits, unsignedAbs(num_fp))
	} else {
		duoC.bits.Lsh(duoC.bits, unsignedAbs(num_fp))
	}
	quo.bits.DivMod(duoC.bits, divC.bits, rem.bits)
}

// Truncate-assigns `rhs` to `x` (trying to copy the numerical value while keeping `x.fp` the same).
// Sets `o` to `true` if the most significant bit of `rhs` would be cut off.
func (x *FP) OTruncateAssign(rhs *FP, o *bool) {
	var tmp = new(big.Int)
	if x.fp <= rhs.fp {
		var s = rhs.fp - x.fp
		if s >= rhs.bits.BitLen() {
			*o = true
		}
		tmp.Rsh(rhs.bits, uint(rhs.fp-x.fp))
	} else {
		var s = x.fp - rhs.fp
		tmp.Lsh(rhs.bits, uint(s))
	}
	x.bits.Set(tmp)
}

// Bounds for some numerical value, with the lower bound being inclusive and the upper bound being exclusive
type FPBounds struct {
	lo FP
	hi FP
}

func NewFPBounds(lo FP, hi FP) FPBounds {
	return FPBounds{
		lo,
		hi,
	}
}

func (bounds FPBounds) String() string {
	return fmt.Sprintf("FPBounds(lo: %v, hi: %v)", bounds.lo, bounds.hi)
}

// Faster exponential calculations requires `2^x` and different kinds
// logarithms, so we have to bootstrap our way with these slow functions that
// depend on nothing but basic arithmetic. `e^x` allows us to use bisection to
// calculate `ln(x)` which allows us to calculate `ln(2)`, then we can change
// base to calculate `2^x`, calculate `lb(x)`, and finally `lb(e)` which
// completes the bootstrapping constants we need.

// Note: the following functions are intended to be rigorous but not fast, see the `exp` function
// for `InflationCurve`

// Returns the lower and upper bounds for `e^x` using a series approximation in `x`'s fp type.
// Returns `None` if one of the bounds overflows or some internal error condition occurs.
//
// Note: `x` is assumed to be perfect numerically, if the user's `x` is a truncated version of some
// numerical value, then the lower bound of `x.ExpBounds` and upper bound of
// `xIncremented.ExpBounds` are the real bounds.
func (input *FP) ExpBounds(maxBitLen int) *FPBounds {
	var x = NewFP(new(big.Int).Set(input.bits), input.fp)
	var sign = x.IsNeg()
	x.bits.Abs(x.bits)
	// these are set to the initial +1 in the taylor series
	var loMul = NewOneFP(x.fp)
	var hiMul = NewOneFP(x.fp)
	// partial sums of the taylor series
	var loSum = NewZeroFP(x.fp)
	var hiSum = NewZeroFP(x.fp)
	for i := uint64(1); ; i++ {
		var o = false
		if sign && ((i & 1) == 0) {
			// note what we do here: `hiMul` is subtracted from `loSum` and `loMul` is
			// subtracted from `hiSum` in order to capture the maximally pessimistic bounds.
			loSum.SameFpSubAssign(hiMul, &o)
			hiSum.SameFpSubAssign(loMul, &o)
		} else {
			loSum.SameFpAddAssign(loMul, &o)
			hiSum.SameFpAddAssign(hiMul, &o)
		}
		// increase the power of x in `loMul` and `hiMul`
		loMul.FpMulAssign(&x)
		hiMul.FpMulAssign(&x)
		// (exclusive) upper bound of what the truncation in the fixed point multiplication above
		// removed
		hiMul.IncAssign()
		// increase the factorial in the divisor
		loMul.ShortUDivideAssign(i)
		hiMul.ShortUDivideAssign(i)
		if o == true {
			return nil
		}
		if hiMul.IsZero() {
			break
		}
		hiMul.IncAssign()
		if hiMul.bits.BitLen() > maxBitLen {
			return nil
		}
	}
	// if `sign` and the result should be in the range 0.0..1.0, the `hiMul` 1.0 value will be
	// subtracted first from `loSum` before the `break`, so there aren't any weird low fp cases
	// where `loMul` would need to be decremented

	// be really conservative and add 4 because of the last increment for `hiMul` after the `break`,
	// plus 3 for low fp cases where the remaining partials could add up to `e` and/or be enough to
	// roll over the ULP.
	hiSum.bits.Add(hiSum.bits, new(big.Int).SetUint64(4))

	// be extra conservative for cases
	var bounds = NewFPBounds(loSum, hiSum)
	return &bounds
}

// Computes a bound on the natural logarithm of `x`. Uses bisection of `ExpBounds` starting with a
// value `2^startBitLen` (in ULPs).
func (x *FP) LnBounds(maxBitLen int) *FPBounds {
	// Uses a two stage bisection separately for both bounds, starting with
	// increasing steps until overshoot happens, then decreasing steps to narrow
	// down the tightest bounds. We need two stages because if we start with large
	// steps, `exp_bounds` can produce valid but incredibly wide bounds that
	// act in a metastable way to confuse the decreasing stage.
	if x.IsNeg() {
		return nil
	}
	var ltOne = x.Lt(NewOneFP(x.fp))

	// stage 1, separately for both bounds. Steps exponentially negative if
	// `ltOne`, else steps exponentially positive until passing `x`.
	var bisectLo = NewZeroFP(x.fp)
	var stepLo = NewFP(new(big.Int).SetUint64(1), x.fp)
	var o = false
	for {
		var tmp0 = bisectLo.ExpBounds(maxBitLen)
		if tmp0 == nil {
			return nil
		}
		var hi = tmp0.hi
		bisectLo.SameFpNegAddAssign(ltOne, &stepLo, &o)
		if o {
			return nil
		}
		if ltOne {
			if hi.Le(x) {
				break
			}
		} else if x.Lt(&hi) {
			break
		}
		stepLo.bits.Lsh(stepLo.bits, 1)
		if stepLo.IsZero() {
			// prevent infinite loops in case `step_lo` zeroes before `exp_bounds` returns
			// `None`
			return nil
		}
	}
	var bisectHi = NewZeroFP(x.fp)
	var stepHi = NewFP(new(big.Int).SetUint64(1), x.fp)
	for {
		var tmp0 = bisectHi.ExpBounds(maxBitLen)
		if tmp0 == nil {
			return nil
		}
		var lo = tmp0.lo
		bisectHi.SameFpNegAddAssign(ltOne, &stepHi, &o)
		if o {
			return nil
		}
		if ltOne {
			if lo.Lt(x) {
				break
			}
		} else if x.Le(&lo) {
			break
		}
		stepHi.bits.Lsh(stepHi.bits, 1)
		if stepHi.IsZero() {
			// prevent infinite loops in case `step_lo` zeroes before `exp_bounds` returns
			// `None`
			return nil
		}
	}

	// stage 2, hone in. Finds bounds such that `expBounds(bisectLo).upperBound
	// <= self < expBounds(bisectHi).lowerBound`.
	for {
		var tmp0 = bisectLo.ExpBounds(maxBitLen)
		if tmp0 == nil {
			return nil
		}
		var hi = tmp0.hi
		bisectLo.SameFpNegAddAssign(x.Le(&hi), &stepLo, &o)
		if o {
			return nil
		}
		stepLo.bits.Rsh(stepLo.bits, 1)
		if stepLo.IsZero() {
			break
		}
	}
	for {
		var tmp0 = bisectHi.ExpBounds(maxBitLen)
		if tmp0 == nil {
			return nil
		}
		var lo = tmp0.lo
		bisectHi.SameFpNegAddAssign(x.Lt(&lo), &stepHi, &o)
		if o {
			return nil
		}
		stepHi.bits.Rsh(stepHi.bits, 1)
		if stepHi.IsZero() {
			break
		}
	}

	// If bisection hits `self` perfectly in one step, the following steps will
	// bring it 1 ULP away. We do one extra increment to insure that the bisection
	// isn't one off.
	bisectLo.DecAssign()
	bisectHi.IncAssign()
	// final check
	var tmp0 = bisectLo.ExpBounds(maxBitLen)
	if tmp0 == nil {
		return nil
	}
	var hi = tmp0.hi
	var tmp1 = bisectHi.ExpBounds(maxBitLen)
	if tmp1 == nil {
		return nil
	}
	var lo = tmp1.lo
	if hi.Le(x) && x.Lt(&lo) {
		var bounds = NewFPBounds(bisectLo, bisectHi)
		return &bounds
	} else {
		return nil
	}
}

// Returns the lower and upper bounds for `e^(x * rhsLo)` and `e^(x *
// rhsHi)`. Returns `None` if `self` and `rhs` do not have the same fixed
// point type. Useful for general exponentiation.
func (x *FP) MulExpBounds(rhs *FPBounds, maxBitLen int) *FPBounds {
	var tmp0 = CopyFP(x)
	tmp0.FpMulAssign(&rhs.lo)
	var tmp1 = CopyFP(x)
	tmp1.FpMulAssign(&rhs.hi)
	// a negative `x` complicates things, just swap so `tmp0 < tmp1` and widen both bounds
	if tmp1.Lt(&tmp0) {
		var swap = tmp0
		tmp0 = tmp1
		tmp1 = swap
	}
	tmp0.DecAssign()
	tmp1.IncAssign()
	var res0 = tmp0.ExpBounds(maxBitLen)
	var res1 = tmp1.ExpBounds(maxBitLen)
	if (res0 == nil) || (res1 == nil) {
		return nil
	}
	var res = NewFPBounds(res0.lo, res1.hi)
	return &res
}

// Returns the lower and upper bounds for `2^x`
func (x *FP) Exp2Bounds(maxBitLen int) *FPBounds {
	var two = NewOneFP(x.fp)
	if two == nil {
		return nil
	}
	two.bits.Lsh(two.bits, 1)
	var ln2 = two.LnBounds(maxBitLen)
	if ln2 == nil {
		return nil
	}
	// use `2^x = e^(x*ln(2))``
	var res = x.MulExpBounds(ln2, maxBitLen)
	return res
}

// Returns the lower and upper bounds for `lb(x)` (`lb` is the binary logarithm or base 2 logarithm)
func (x *FP) LbBounds(maxBitLen int) *FPBounds {
	var two = NewOneFP(x.fp)
	if two == nil {
		return nil
	}
	two.bits.Lsh(two.bits, 1)
	var ln2 = two.LnBounds(maxBitLen)
	// use `lb(x) = ln(x)/ln(2)`
	var lnX = x.LnBounds(maxBitLen)
	var rem = NewZeroFP(x.fp)
	var quoLo = NewZeroFP(x.fp)
	var quoHi = NewZeroFP(x.fp)
	var o = false
	// smallest dividend and largest divisor
	quoLo.FpDivideAssign(&rem, &lnX.lo, &ln2.hi, &o)
	// largest dividend and smallest divisor
	quoHi.FpDivideAssign(&rem, &lnX.hi, &ln2.lo, &o)
	if o {
		return nil
	}
	var res = NewFPBounds(quoLo, quoHi)
	return &res
}

func TestTranscendentals(t *testing.T) {
	var x0 = NewOneFP(32)
	var expBounds = x0.ExpBounds(128)
	// This bounds the value of Euler's number which is best truncated to `11674931554 * 2^(-32)`
	require.True(t, expBounds.lo.bits.Cmp(new(big.Int).SetUint64(11674931549)) == 0)
	require.True(t, expBounds.hi.bits.Cmp(new(big.Int).SetUint64(11674931571)) == 0)
	var x1 = NewOneFP(32)
	x1.Neg()
	var expNegBounds = x1.ExpBounds(128)
	// 1/e
	require.True(t, expNegBounds.lo.bits.Cmp(new(big.Int).SetUint64(1580030160)) == 0)
	require.True(t, expNegBounds.hi.bits.Cmp(new(big.Int).SetUint64(1580030182)) == 0)

	// ln(3.123)
	var x2 = NewFP(new(big.Int).SetUint64(13413182865), 32)
	var lnBounds0 = x2.LnBounds(128)
	require.True(t, lnBounds0.lo.bits.Cmp(new(big.Int).SetUint64(4891083317)) == 0)
	require.True(t, lnBounds0.hi.bits.Cmp(new(big.Int).SetUint64(4891083327)) == 0)
	// ln(0.5)
	var x3 = NewOneFP(32)
	x3.bits.Rsh(x3.bits, 1)
	var lnBounds1 = x3.LnBounds(128)
	var tmp2 = new(big.Int).SetUint64(2977044497)
	tmp2.Neg(tmp2)
	require.True(t, lnBounds1.lo.bits.Cmp(tmp2) == 0)
	var tmp3 = new(big.Int).SetUint64(2977044449)
	tmp3.Neg(tmp3)
	require.True(t, lnBounds1.hi.bits.Cmp(tmp3) == 0)

	// (1.234).MulExpBounds((-4.321..1.337))
	var x4 = NewFP(new(big.Int).SetUint64(5299989643), 32)
	var rhsLo = NewFP(new(big.Int).SetUint64(18558553686), 32)
	rhsLo.Neg()
	var rhsHi = NewFP(new(big.Int).SetUint64(5742371275), 32)
	var bounds = NewFPBounds(rhsLo, rhsHi)
	var res = x4.MulExpBounds(&bounds, 128)
	require.True(t, res.lo.bits.Cmp(new(big.Int).SetUint64(20761109)) == 0)
	require.True(t, res.hi.bits.Cmp(new(big.Int).SetUint64(22360632655)) == 0)
	// (-1.234).MulExpBounds((-4.321..1.337))
	x4.Neg()
	res = x4.MulExpBounds(&bounds, 128)
	require.True(t, res.lo.bits.Cmp(new(big.Int).SetUint64(824965199)) == 0)
	require.True(t, res.hi.bits.Cmp(new(big.Int).SetUint64(888520696425)) == 0)

	// `2^(-0.42)`
	var x5 = NewFP(new(big.Int).SetUint64(1803886264), 32)
	x5.Neg()
	res = x5.Exp2Bounds(128)
	require.True(t, res.lo.bits.Cmp(new(big.Int).SetUint64(3210164309)) == 0)
	// accurate truncated value is 3210164317
	require.True(t, res.hi.bits.Cmp(new(big.Int).SetUint64(3210164328)) == 0)

	// `lb(1234)`
	var x6 = NewFP(new(big.Int).SetUint64(5299989643264), 32)
	res = x6.LbBounds(128)
	require.True(t, res.lo.bits.Cmp(new(big.Int).SetUint64(44105563196)) == 0)
	// accurate truncated value is 44105563245
	require.True(t, res.hi.bits.Cmp(new(big.Int).SetUint64(44105563376)) == 0)
}

// There is a problem when calculating transcendental functions like `e^x`. We
// know that its Taylor series is `x^0/0! + x^1/1! + x^2/2! + x^3/3! + ...`.
// Besides error incurred in calculating the individual terms, we must
// eventually stop adding terms. If all are terms are positive (which can be
// enforced with identities like `e^(-abs(x)) = 1 / e(abs(x))`), then cutting
// off some terms will result in a lower bound on the function. If the
// individual terms are also biased to have truncation errors (note: we choose
// use truncation rounding and no other kind of rounding, because the horner
// evaluation in exp2m1 must not result in a partial greater than 1.0, or
// else we could get exponential growth problems) that lower bound the
// function, then the whole thing will be a lower bound of the true value
// of `e^x`. It would be great if we could always get a perfectly rounded
// bitstring for a given fixed point numerical function, for the purposes of
// determinisim. What I mean by "perfect" is that no larger intermediate fixed
// point computations with more terms would change the truncated bitstring
// further. Unfortunately, as far as I currently know, there is no property of
// `e^x` that prevents one of the many inputs from forming an arbitrarily long
// string of binary ones in the fraction that will roll over from a very low
// term, and change the target truncation (thus messing up the perfect
// determinism). I have decided to use an intermediate fixed point type double
// the bitwidth and fixed point of the desired type we are operating on. Any
// precomputed constants will be computed with a high enough precompute fixed
// point such that they have perfect trunctions, and thus we can have a higher
// level determinism. Note that if the intermediate fixed point type is
// changed (say, we move from `exp2Fp = 128` to `exp2Fp = 256`), there is a
// slim but nonzero chance that one of the inputs in the `exp2Fp = 128` domain
// did not have a perfect truncated output, that would be incremented in the
// new `exp2Fp = 256` output. The new functions _must_ be considered
// deterministically incompatible with the old functions even if it appears
// from extensive fuzzing that all the outputs are equal.

func TestFastExp(t *testing.T) {
	//var targetBw = 128
	//var targetFp = 64
	// intermediate bitwidth
	var exp2Fp = 128
	// precompute for perfect constants
	var precomputeBw = 256
	var precomputeFp = 192
	var one = NewOneFP(precomputeFp)
	var e = one.ExpBounds(precomputeBw)
	var lbE = NewFPBounds(e.lo.LbBounds(precomputeBw).lo, e.hi.LbBounds(precomputeBw).hi)
	var target0 = NewZeroFP(exp2Fp)
	var target1 = NewZeroFP(exp2Fp)
	// We will now truncate the bounds on `lb(e)` to the target fixed point.
	// If they are equal and the msb was not cut off, then the truncation is
	// perfect.
	var o = false
	target0.OTruncateAssign(&lbE.lo, &o)
	target1.OTruncateAssign(&lbE.hi, &o)
	require.True(t, !o)
	require.True(t, target0.bits.Cmp(target1.bits) == 0)
	require.True(t, target0.bits.Cmp(globalFastExp.lbE) == 0)

	var two = NewOneFP(precomputeFp)
	two.bits.Lsh(two.bits, 1)
	var ln2 = two.LnBounds(precomputeBw)

	// skip to `C_1`
	var product = NewFPBounds(CopyFP(&ln2.lo), CopyFP(&ln2.hi))

	var constants = []FP{}

	// check the perfection of the `C_n = ln(2)^n / n!` constants
	for i := uint64(0); ; i++ {
		target0.OTruncateAssign(&product.lo, &o)
		target1.OTruncateAssign(&product.hi, &o)
		require.True(t, target0.bits.Cmp(target1.bits) == 0)
		// The constants become zero at the same time the truncation starts to overflow
		// note: this must go after the above require
		if o {
			break
		}
		constants = append(constants, CopyFP(&target0))

		// multiply by both bounds of `ln2`
		product.lo.FpMulAssign(&ln2.lo)
		product.hi.FpMulAssign(&ln2.hi)
		product.hi.IncAssign()
		// divide to increase factorial
		product.lo.ShortUDivideAssign(i + 2)
		product.hi.ShortUDivideAssign(i + 2)
		product.hi.IncAssign()
	}

	var actualLen = len(globalFastExp.exp2m1Constants)
	// check that the number of constants is correct
	require.True(t, len(constants) == actualLen)

	for i := range constants {
		// the order is reversed
		require.True(t, constants[actualLen-1-i].bits.Cmp(globalFastExp.exp2m1Constants[i]) == 0)
		// print out the constants for usage in `newFastExp()`
		//fmt.Printf("newBigIntWithTenBase(\"%s\"),\n", constants[actualLen - 1 - i].bits.String());
	}

	var bigOne = NewOneFP(exp2Fp)
	require.True(t, bigOne.bits.Cmp(globalFastExp.exp2FpOne) == 0)
}

func TestInflationConstants(t *testing.T) {
	var aNomConversion = new(big.Int).Exp(big.NewInt(10), big.NewInt(sdk.Precision), nil)

	var peakOffset, _ = new(big.Int).SetString("150000000", 10)
	// we move the subtraction in the peak offset to the constant as a negative sign for technical reasons
	peakOffset.Neg(peakOffset)
	peakOffset.Mul(peakOffset, aNomConversion)

	// for usage in `newInflationCurve()`
	//fmt.Printf("newBigIntWithTenBase(\"%s\"),\n", peakOffset);
	require.True(t, peakOffset.Cmp(globalInflationCurve.peakOffset) == 0)

	// we calculate `peakScale` as `-1/(2*(stdDev^2))` in 384 fixed point. The fixed point needs to
	// be very high because this in aNom is a very small number (log_2(peakScale) ~= -171 zero bits
	// below the fixed point, so large that our fixed point will be more than the storage bits we
	// need). We choose to be conservative and want 192 significant bits to avoid losing precision
	// before getting to the exponential function in the bell curve calculation, so when rounding to
	// 64 bit multiples we get a i256f384 representation.
	// If this value is changed, it is safe to do without needing to change the representation so
	// long as it does not exceed about 1e29 aNom = 1e11 Nom. Note that if we use a 256 fixed
	// bitwidth then the value must also be larger than about 3e19 aNom = 30 Nom, or else the
	// number of significant digits will be excessive.
	var stdDev, _ = new(big.Int).SetString("50000000", 10)
	stdDev.Mul(stdDev, aNomConversion)

	// square
	stdDev.Mul(stdDev, stdDev)
	// multiply by 2
	stdDev.Lsh(stdDev, 1)
	// invert last
	var peakScale = NewZeroFP(384)
	var rem = NewZeroFP(384)
	var one = NewOneFP(0)
	var o = false
	var div = NewFP(stdDev, 0)
	// just use the natural truncation rounding from this divide
	peakScale.FpDivideAssign(&rem, one, &div, &o)
	peakScale.Neg()
	require.False(t, o)
	// for usage in `newInflationCurve()`
	//fmt.Printf("newBigIntWithTenBase(\"%s\"),\n", peakScale.bits);
	require.True(t, peakScale.bits.Cmp(globalInflationCurve.peakScale) == 0)
}
