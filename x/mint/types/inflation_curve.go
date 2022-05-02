package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"math/big"
)

var (
	precisionReuse       = new(big.Int).Exp(big.NewInt(10), big.NewInt(sdk.Precision), nil)
	globalFastExp        = newFastExp()
	globalInflationCurve = newInflationCurve()
)

// FastExp contains the constants needed for fast computation of i128f64 fixed point exponential functions.
// Use the instance `globalFastExp` so that only one copy of the constants is needed.
type FastExp struct {
	// Constants `C_n = ln(2)^n / n!` (ordering of the list is reversed and starts with the highest
	// `n` that doesn't result in zero) as 128 bit fracints with truncation rounding
	exp2m1Constants [30]*big.Int
	// for base change in `exp`
	lbE *big.Int
	// for large 1.0 value with 128 fixed point
	exp2FpOne *big.Int
}

func newFastExp() *FastExp {
	constants := [30]*big.Int{
		newBigIntWithTenBase("21"),
		newBigIntWithTenBase("931"),
		newBigIntWithTenBase("38977"),
		newBigIntWithTenBase("1574505"),
		newBigIntWithTenBase("61331333"),
		newBigIntWithTenBase("2300542666"),
		newBigIntWithTenBase("82974537419"),
		newBigIntWithTenBase("2872966887729"),
		newBigIntWithTenBase("95330746876023"),
		newBigIntWithTenBase("3025730306770147"),
		newBigIntWithTenBase("91669328281539416"),
		newBigIntWithTenBase("2645017706267986357"),
		newBigIntWithTenBase("72503124630030170854"),
		newBigIntWithTenBase("1882798170348581770631"),
		newBigIntWithTenBase("46177160917064115362943"),
		newBigIntWithTenBase("1065912976918080910514347"),
		newBigIntWithTenBase("23066810487283611717419874"),
		newBigIntWithTenBase("465897223387814402039923456"),
		newBigIntWithTenBase("8737918978691986426796998167"),
		newBigIntWithTenBase("151273828538981816810644004423"),
		newBigIntWithTenBase("2400662024744240530078289187559"),
		newBigIntWithTenBase("34634231979489737747471345012729"),
		newBigIntWithTenBase("449699712496270121258630534506364"),
		newBigIntWithTenBase("5190236360860492089195900933424948"),
		newBigIntWithTenBase("52415497811985085885772577715591695"),
		newBigIntWithTenBase("453717472554463172968722258900220684"),
		newBigIntWithTenBase("3272879738094991899426931849626998944"),
		newBigIntWithTenBase("18887069470302456743998381224473040631"),
		newBigIntWithTenBase("81744844385192085827234082247051493269"),
		newBigIntWithTenBase("235865763225513294137944142764154484399"),
	}
	return &FastExp{
		exp2m1Constants: constants,
		lbE:             newBigIntWithTenBase("490923683258796565746369346286093237521"),
		exp2FpOne:       newBigIntWithTenBase("340282366920938463463374607431768211456"),
	}
}

// InflationCurve is the struct used for the inflation curve calculation. Use the instance
// `globalInflationCurve` so that only one copy of the constants is needed.
type InflationCurve struct {
	fastExp *FastExp
	// for curve peak position = 150_000_000 NOM
	peakOffset *big.Int
	// adjusts peak height, `-1/(2*(stdDev^2))` with 384 fixed point, stdDev = 50_000_000 NOM
	peakScale *big.Int
}

// Fast calculation of a bell curve for the hyperinflation regime.
func newInflationCurve() *InflationCurve {
	// see TestInflationConstants for calculation
	return &InflationCurve{
		fastExp:    globalFastExp,
		peakOffset: newBigIntWithTenBase("-150000000000000000000000000"),
		peakScale:  newBigIntWithTenBase("-7880401239278895842455808020028722761015947854093089333589658680"),
	}
}

// DecExp quickly calculates `e^x` as a `Dec`. Returns `nil` if overflow occurs.
// This should be called on `globalInflationCurve` so that constants can be reused.
func (fe *FastExp) DecExp(x sdk.Dec) *sdk.Dec {
	// convert to i128f64 binary fixed point
	var tmp0 = x.BigInt()
	tmp0.Lsh(tmp0, 64)
	tmp0.Div(tmp0, precisionReuse)
	var tmp1 = fe.exp(tmp0)
	if tmp1 == nil {
		return nil
	}
	// multiply by 10^18 to get to maximum precision allowed by `Dec`
	tmp1.Mul(tmp1, precisionReuse)
	// remove binary fixed point component
	tmp1.Rsh(tmp1, 64)
	var res = sdk.NewDecFromBigIntWithPrec(tmp1, sdk.Precision)
	return &res
}

// CalculateInflationDec is the same as `calculateInflationBinary` but with an `Int` input and `Dec` output.
// This should be called on `globalInflationCurve` so that constants can be reused.
func (ic *InflationCurve) CalculateInflationDec(tokenSupply sdk.Int) sdk.Dec {
	// People keep committing the same horrible mistake of using base 10 fixed point numbers at the
	// computational layer instead of converting between binary fixed point at the human-machine
	// interface by using banker's rounding which is already inevitable. Above, we absolutely
	// needed binary fixed point to do fast divisions by powers of two, but the output here needs
	// to be a `Dec`.
	var tmp = ic.calculateInflationBinary(tokenSupply.BigInt())
	// multiply by 10^18 to get to maximum precision allowed by `Dec`
	tmp.Mul(tmp, precisionReuse)
	// remove binary fixed point component
	tmp.Rsh(tmp, 64)
	return sdk.NewDecFromBigIntWithPrec(tmp, sdk.Precision)
}

// newBigIntWithTenBase allocates and returns a new Int set to s with base decimals.
func newBigIntWithTenBase(s string) *big.Int {
	i, _ := new(big.Int).SetString(s, 10)
	return i
}

// exp2m1Fracint calculates `2^x - 1` using a 128 bit fractint.
func (fe *FastExp) exp2m1Fracint(x *big.Int) *big.Int {
	// upstream invariants should prevent this, but check
	// just in case because a width explosion happens otherwise
	if x.BitLen() > 128 {
		return nil
	}
	// Note: there is likely a method involving Pade approximants that uses fewer constants
	// and is faster. I have chosen this method for its simplicity. The input and constants
	// are both purely fractional and have only truncation biases applied which guarantees
	// everything stays in the [0, 1) range and cannot exponentially explode, which is bad
	// with dynamic bigints.
	var h = new(big.Int)
	// `2^x = 1 + (ln(2)/1!)*x + (ln(2)^2/2!)*x^2 + (ln(2)^3/3!)*x^3`
	// `2^x - 1 = x*(C_1 + x*(C_2 + x*(C_3 + ...)))` where `C_n = ln(2)^n / n!`
	for _, c := range fe.exp2m1Constants {
		h.Add(h, c)
		h.Mul(h, x)
		// shift from fp = 256 to fp = 128
		h.Rsh(h, 128)
	}
	return h
}

func unsignedAbs(x int) uint {
	if x < 0 {
		// arithmetic is wrapping in Go so this handles imin correctly
		return uint(-x)
	}

	return uint(x)
}

// Calculates `e^x`. Input and output are i128f64 bit fixed point integers.
// Returns `nil` if the result overflows 128 bits.
func (fe *FastExp) exp(x *big.Int) *big.Int {
	// `e^x = 2^(x*lb(e))`
	// `2^x = (2^floor(x)) * 2^(x - floor(x))`
	// let `y = x*lb(e)`
	// `e^x = (2^floor(y)) * 2^(y - floor(y))`

	var msb = x.Sign() == -1
	// make unsigned
	x.Abs(x)
	// convert bases
	x.Mul(x, fe.lbE)
	// lbE.fp + x.fp = 128 + 64 = 192
	var intPart = new(big.Int).Rsh(x, 192)
	// extract the fractional part
	var fracPart = new(big.Int).Lsh(intPart, 192)
	fracPart.Sub(x, fracPart)
	// get to fp = 128
	fracPart.Rsh(fracPart, 64)

	if !intPart.IsInt64() {
		if msb {
			// certain zero
			return new(big.Int)
		} else {
			// certain overflow
			return nil
		}
	}
	var shift = int(intPart.Int64())
	if msb {
		if shift >= 64 {
			// certain zero
			return new(big.Int)
		}
	} else if shift >= 128 {
		// certain overflow
		return nil
	}

	if msb {
		shift = -shift - 1
		// two's complement without changing the sign
		fracPart.Sub(fe.exp2FpOne, fracPart)
	}
	// include shift from fp=128 to fp=64
	shift -= 64

	// calculate exp2
	var res = fe.exp2m1Fracint(fracPart)
	// faster OR because we know the one's place cannot be set
	res.Or(res, fe.exp2FpOne)

	if shift < 0 {
		res.Rsh(res, unsignedAbs(shift))
	} else {
		res.Lsh(res, unsignedAbs(shift))
	}
	return res
}

// Calculates `e^(-((x-peakOffset)^2 / (2*(std_dev^2)))))`
// `tokenSupply` is in units of token. Returns a fixed point integer capped at 1.0u65f64.
func (ic *InflationCurve) calculateInflationBinary(tokenSupply *big.Int) *big.Int {
	// apply offset
	var tmp = new(big.Int).Add(tokenSupply, ic.peakOffset)
	// square
	tmp.Mul(tmp, tmp)
	// scale
	tmp.Mul(tmp, ic.peakScale)
	// keep at fp = 64
	tmp.Rsh(tmp, 320)
	var res = ic.fastExp.exp(tmp)
	if res == nil {
		// return zero
		return new(big.Int)
	}
	return res
}
