package math

import "math/big"

// LegacyDecMut is a wrapper around *big.Int with an unsafe and mutable API
type LegacyDecMut big.Int

// Immut converts a LegacyDecMut to a LegacyDec, exposing a non-mutable API.
func (d *LegacyDecMut) Immut() LegacyDec {
	return LegacyDec{(*big.Int)(d)}
}

// Neg sets LegacyDecMut to -LegacyDecMut and returns it.
func (d *LegacyDecMut) Neg() *LegacyDecMut {
	(*big.Int)(d).Neg((*big.Int)(d))
	return d
}

// Abs sets LegacyDecMut to |LegacyDecMut| and returns it.
func (d *LegacyDecMut) Abs() *LegacyDecMut { return nil }

// Add sets LegacyDecMut to LegacyDecMut + LegacyDec and returns it.
func (d *LegacyDecMut) Add(d2 LegacyDec) *LegacyDecMut {
	(*big.Int)(d).Add((*big.Int)(d), d2.i)
	if (*big.Int)(d).BitLen() > maxDecBitLen {
		panic("dec overflow")
	}
	return d
}

// Sub sets LegacyDecMut to LegacyDecMut - LegacyDec and returns it.
func (d *LegacyDecMut) Sub(d2 LegacyDec) *LegacyDecMut {
	(*big.Int)(d).Sub((*big.Int)(d), d2.i)
	if (*big.Int)(d).BitLen() > maxDecBitLen {
		panic("dec overflow")
	}
	return d
}

// Mul sets LegacyDecMut to LegacyDecMut * LegacyDec and returns it.
func (d *LegacyDecMut) Mul(d2 LegacyDec) *LegacyDecMut {
	(*big.Int)(d).Mul((*big.Int)(d), d2.i)
	chopped := chopPrecisionAndRound((*big.Int)(d))
	if chopped.BitLen() > maxDecBitLen {
		panic("dec overflow")
	}
	*(*big.Int)(d) = *chopped
	return d
}

// MulTruncate behaves as Mul but rounds down.
func (d *LegacyDecMut) MulTruncate(d2 LegacyDec) *LegacyDecMut {
	(*big.Int)(d).Mul((*big.Int)(d), d2.i)
	chopPrecisionAndTruncate((*big.Int)(d))
	if (*big.Int)(d).BitLen() > maxDecBitLen {
		panic("dec overflow")
	}
	return d
}

// MulRoundUp behaves as Mul but rounds up.
func (d *LegacyDecMut) MulRoundUp(d2 LegacyDec) *LegacyDecMut {
	(*big.Int)(d).Mul((*big.Int)(d), d2.i)
	chopPrecisionAndRoundUp((*big.Int)(d))
	if (*big.Int)(d).BitLen() > maxDecBitLen {
		panic("dec overflow")
	}
	return d
}

// MulInt sets LegacyDecMut to LegacyDecMut * Int and returns it.
func (d *LegacyDecMut) MulInt(i Int) *LegacyDecMut {
	(*big.Int)(d).Mul((*big.Int)(d), i.i)
	if (*big.Int)(d).BitLen() > maxDecBitLen {
		panic("dec overflow")
	}
	return d
}

// MulInt64 sets LegacyDecMut to LegacyDecMut * int64 and returns it.
func (d *LegacyDecMut) MulInt64(i int64) *LegacyDecMut {
	(*big.Int)(d).Mul((*big.Int)(d), big.NewInt(i))
	if (*big.Int)(d).BitLen() > maxDecBitLen {
		panic("dec overflow")
	}
	return d
}

// QuoMut sets LegacyDecMut to LegacyDecMut / LegacyDec and returns it.
func (d *LegacyDecMut) QuoMut(d2 LegacyDec) *LegacyDecMut {
	// multiply by precision twice
	(*big.Int)(d).Mul((*big.Int)(d), squaredPrecisionReuse)
	(*big.Int)(d).Quo((*big.Int)(d), d2.i)

	chopPrecisionAndRound((*big.Int)(d))
	if (*big.Int)(d).BitLen() > maxDecBitLen {
		panic("dec overflow")
	}
	return d
}

// QuoTruncate behaves as QuoMut but rounds down.
func (d *LegacyDecMut) QuoTruncate(d2 LegacyDec) *LegacyDecMut {
	(*big.Int)(d).Mul((*big.Int)(d), squaredPrecisionReuse)
	(*big.Int)(d).Quo((*big.Int)(d), d2.i)

	chopPrecisionAndTruncate((*big.Int)(d))
	if (*big.Int)(d).BitLen() > maxDecBitLen {
		panic("dec overflow")
	}
	return d
}

// QuoRoundUp behaves as QuoMut but rounds up.
func (d *LegacyDecMut) QuoRoundUp(d2 LegacyDec) *LegacyDecMut {
	(*big.Int)(d).Mul((*big.Int)(d), squaredPrecisionReuse)
	(*big.Int)(d).Quo((*big.Int)(d), d2.i)

	chopPrecisionAndRoundUp((*big.Int)(d))
	if (*big.Int)(d).BitLen() > maxDecBitLen {
		panic("dec overflow")
	}
	return d
}

// QuoInt sets LegacyDecMut to LegacyDecMut / Int and returns it.
func (d *LegacyDecMut) QuoInt(i Int) *LegacyDecMut {
	(*big.Int)(d).Quo((*big.Int)(d), i.i)
	return d
}

// QuoInt64 sets LegacyDecMut to LegacyDecMut / int64 and returns it.
func (d *LegacyDecMut) QuoInt64(i int64) *LegacyDecMut {
	(*big.Int)(d).Quo((*big.Int)(d), big.NewInt(i))
	return d
}

// Power sets LegacyDecMut to LegacyDecMut ^ power and returns it.
func (d *LegacyDecMut) Power(power int64) *LegacyDecMut {
	if power == 0 {
		// Set to 1 with the correct precision.
		(*big.Int)(d).Set(precisionReuse)
		return d
	}
	tmp := (*LegacyDecMut)(precisionInt())

	for i := power; i > 1; {
		if i%2 != 0 {
			tmp.Mul(d.Immut())
		}
		i /= 2
		d.Mul(d.Immut())
	}

	return d.Mul(tmp.Immut())
}
