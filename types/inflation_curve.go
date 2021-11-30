package types

import (
	"math/big"
)

// Stores constants needed for inflation curve calculation.
// The original Rust implementation this is generated from can be found in
// https://github.com/AaronKutch/mint_test
//
// Note about performance: the original implementation was designed so that no allocations happen
// after the constants are initialized. Go's bigint library might be allocating for every
// multiplication, we need to port over some things from `awint` if we want to speed this up.
type InflationCurve struct {
	// for exp2m1
	constants [30]*big.Int
	// for base change
	lbE *big.Int
	// for large 1.0 value
	exp2FwOne *big.Int
	// for curve peak position
	peakOffset *big.Int
	// adjusts peak height
	peakScale *big.Int
}

// Use `GlobalInflationCurve` instead of needing to call this function more than once
func newInflationCurve() *InflationCurve {
	ic := InflationCurve{}
	var i = new(big.Int)
	i, _ = new(big.Int).SetString("21", 10)
	ic.constants[0] = i
	i, _ = new(big.Int).SetString("931", 10)
	ic.constants[1] = i
	i, _ = new(big.Int).SetString("38977", 10)
	ic.constants[2] = i
	i, _ = new(big.Int).SetString("1574505", 10)
	ic.constants[3] = i
	i, _ = new(big.Int).SetString("61331333", 10)
	ic.constants[4] = i
	i, _ = new(big.Int).SetString("2300542666", 10)
	ic.constants[5] = i
	i, _ = new(big.Int).SetString("82974537419", 10)
	ic.constants[6] = i
	i, _ = new(big.Int).SetString("2872966887729", 10)
	ic.constants[7] = i
	i, _ = new(big.Int).SetString("95330746876023", 10)
	ic.constants[8] = i
	i, _ = new(big.Int).SetString("3025730306770147", 10)
	ic.constants[9] = i
	i, _ = new(big.Int).SetString("91669328281539416", 10)
	ic.constants[10] = i
	i, _ = new(big.Int).SetString("2645017706267986357", 10)
	ic.constants[11] = i
	i, _ = new(big.Int).SetString("72503124630030170854", 10)
	ic.constants[12] = i
	i, _ = new(big.Int).SetString("1882798170348581770631", 10)
	ic.constants[13] = i
	i, _ = new(big.Int).SetString("46177160917064115362943", 10)
	ic.constants[14] = i
	i, _ = new(big.Int).SetString("1065912976918080910514347", 10)
	ic.constants[15] = i
	i, _ = new(big.Int).SetString("23066810487283611717419874", 10)
	ic.constants[16] = i
	i, _ = new(big.Int).SetString("465897223387814402039923456", 10)
	ic.constants[17] = i
	i, _ = new(big.Int).SetString("8737918978691986426796998167", 10)
	ic.constants[18] = i
	i, _ = new(big.Int).SetString("151273828538981816810644004423", 10)
	ic.constants[19] = i
	i, _ = new(big.Int).SetString("2400662024744240530078289187559", 10)
	ic.constants[20] = i
	i, _ = new(big.Int).SetString("34634231979489737747471345012729", 10)
	ic.constants[21] = i
	i, _ = new(big.Int).SetString("449699712496270121258630534506364", 10)
	ic.constants[22] = i
	i, _ = new(big.Int).SetString("5190236360860492089195900933424948", 10)
	ic.constants[23] = i
	i, _ = new(big.Int).SetString("52415497811985085885772577715591695", 10)
	ic.constants[24] = i
	i, _ = new(big.Int).SetString("453717472554463172968722258900220684", 10)
	ic.constants[25] = i
	i, _ = new(big.Int).SetString("3272879738094991899426931849626998944", 10)
	ic.constants[26] = i
	i, _ = new(big.Int).SetString("18887069470302456743998381224473040631", 10)
	ic.constants[27] = i
	i, _ = new(big.Int).SetString("81744844385192085827234082247051493269", 10)
	ic.constants[28] = i
	i, _ = new(big.Int).SetString("235865763225513294137944142764154484399", 10)
	ic.constants[29] = i
	i, _ = new(big.Int).SetString("490923683258796565746369346286093237521", 10)
	ic.lbE = i
	i, _ = new(big.Int).SetString("340282366920938463463374607431768211456", 10)
	ic.exp2FwOne = i
	i, _ = new(big.Int).SetString("-150000000000000000000000000", 10)
	ic.peakOffset = i
	i, _ = new(big.Int).SetString("-7880401239278895842455808020028722761015947854093089333589658680", 10)
	ic.peakScale = i
	return &ic
}

// Contains the constants needed for inflation curve calculation
var GlobalInflationCurve = newInflationCurve()

// Calculates `2^x - 1` using a 128 bit fracint
func (ic *InflationCurve) exp2m1Fracint(x *big.Int) *big.Int {
	// upstream invariants should prevent this, but check
	// just in case because a width explosion happens otherwise
	if x.BitLen() > 128 {
		return nil
	}
	var h = new(big.Int)
	for _, c := range ic.constants {
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
	} else {
		return uint(x)
	}
}

// Calculates `e^x`. Input and output are {i,u}128f64 bit fixed point integers.
// Returns `nil` if the result overflows 128 bits.
func (ic *InflationCurve) exp(x *big.Int) *big.Int {
	if x.BitLen() > 128 {
		return nil
	}
	var msb = x.Sign() == -1
	// make unsigned
	x.Abs(x)
	// convert bases
	x.Mul(x, ic.lbE)
	// lbE.fp + x.fp = 128 + 64 = 192
	var intPart = new(big.Int).Rsh(x, 192)
	// extract the fractional part
	var fracPart = new(big.Int).Lsh(intPart, 192)
	fracPart.Sub(x, fracPart)
	// get to fp = 128
	fracPart.Rsh(fracPart, 64)

	if !intPart.IsInt64() {
		// certain overflow
		return nil
	} else {
		// note: we assume `int` is 64 bits
		var shift = int(intPart.Int64())
		if msb {
			shift = -shift - 1
			// two's complement without changing the sign
			fracPart.Sub(ic.exp2FwOne, fracPart)
		}
		// include shift from fp=128 to fp=64
		shift -= 64

		// calculate exp2
		var res = ic.exp2m1Fracint(fracPart)
		res.Or(res, ic.exp2FwOne)

		switch {
		case shift < 0:
			var shift = unsignedAbs(shift)
			if shift >= 128 {
				// shift is large enough to guarantee zero
				return new(big.Int)
			} else {
				res.Rsh(res, shift)
			}
		case shift <= (128 - res.BitLen()):
			res.Lsh(res, unsignedAbs(shift))
			if res.BitLen() > 128 {
				return nil
			}
		default:
			// guaranteed overflow from large positive shift
			return nil
		}
		return res
	}
}

// Calculates `e^(-((x-peak_position)^2 / (2*(std_dev^2)))))`.
// `anomSupply` is in units of aNom. Returns a fixed point integer capped at 1.0u65f64.
func (ic *InflationCurve) calculateInflationBinary(anomSupply *big.Int) *big.Int {
	if anomSupply.BitLen() >= 90 {
		// guaranteed < 2^-64
		return new(big.Int)
	}
	// apply offset
	var tmp = new(big.Int).Add(anomSupply, ic.peakOffset)
	// square
	tmp.Mul(tmp, tmp)
	// scale
	tmp.Mul(tmp, ic.peakScale)
	// keep at fp = 64
	tmp.Rsh(tmp, 320)
	var res = ic.exp(tmp)
	if res == nil {
		// return zero
		return new(big.Int)
	}
	return res
}

// The same as `calculateInflationBinary` but with an `Int` input and `Dec` output
func (ic *InflationCurve) CalculateInflationDec(anomSupply Int) Dec {
	// People keep committing the same horrible mistake of using base 10 fixed point numbers at the
	// computational layer instead of converting between binary fixed point at the human-machine
	// interface by using banker's rounding which is already inevitable. Above, we absolutely
	// needed binary fixed point to do fast divisions by powers of two, but the output here needs
	// to be a `Dec`.
	var tmp = ic.calculateInflationBinary(anomSupply.BigInt())
	// multiply by 10^18 to get to maximum precision allowed by `Dec`
	tmp.Mul(tmp, precisionReuse)
	// remove binary fixed point component
	tmp.Rsh(tmp, 64)
	return NewDecFromBigIntWithPrec(tmp, 18)
}
