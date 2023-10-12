package math

// NegMut reverses the decimal sign, mutable
// Deprecated: use Mut().Neg().Immut() instead
func (d LegacyDec) NegMut() LegacyDec { return d.Mut().Neg().Immut() }

// AbsMut sets the decimal to its absolute value, mutable. Returns itself.
// Deprecated: use Mut().Abs().Immut() instead
func (d LegacyDec) AbsMut() LegacyDec { return d.Mut().Abs().Immut() } // absolute value, mutable

// AddMut adds d2 to d, mutable. Returns itself.
// Deprecated: use Mut().Add(d2).Immut() instead
func (d LegacyDec) AddMut(d2 LegacyDec) LegacyDec { return d.Mut().Add(d2).Immut() }

// SubMut subtracts d2 from d, mutable. Returns itself.
// Deprecated: use Mut().Sub(d2).Immut() instead
func (d LegacyDec) SubMut(d2 LegacyDec) LegacyDec { return d.Mut().Sub(d2).Immut() }

// MulMut multiplies d by d2, mutable. Returns itself.
// Deprecated: use Mut().Mul(d2).Immut() instead
func (d LegacyDec) MulMut(d2 LegacyDec) LegacyDec { return d.Mut().Mul(d2).Immut() }

// MulTruncateMut multiplies d by d2, mutable, rounds down. Returns itself.
// Deprecated: use Mut().MulTruncate(d2).Immut() instead
func (d LegacyDec) MulTruncateMut(d2 LegacyDec) LegacyDec { return d.Mut().MulTruncate(d2).Immut() }

// QuoMut divides a decimal by another decimal and returns the result as a decimal.
// Mutates LegacyDec.
// Deprecated: Use Mut().Quot().Immut() instead.
func (d LegacyDec) QuoMut(d2 LegacyDec) LegacyDec { return d.Mut().Quo(d2).Immut() }

// QuoTruncateMut divides a decimal by another decimal and returns the result as a decimal, rounded-down.
// Mutates LegacyDec.
// Deprecated: Use Mut().QuotTruncate().Immut() instead.
func (d LegacyDec) QuoTruncateMut(d2 LegacyDec) LegacyDec { return d.Mut().QuoTruncate(d2).Immut() }

// QuoRoundupMut divides a decimal by another decimal and returns the result as a decimal, rounded-up.
// Mutates LegacyDec.
// Deprecated: Use Mut().QuotRoundUp().Immut() instead.
func (d LegacyDec) QuoRoundupMut(d2 LegacyDec) LegacyDec { return d.Mut().QuoRoundUp(d2).Immut() }

// QuoIntMut divides a decimal by an Int and returns the result as a decimal.
// Mutates LegacyDec.
// Deprecated: Use Mut().QuotInt().Immut() instead.
func (d LegacyDec) QuoIntMut(i Int) LegacyDec { return d.Mut().QuoInt(i).Immut() }

// QuoInt64Mut divides a decimal by an int64 and returns the result as a decimal.
// Mutates LegacyDec.
// Deprecated: Use Mut().QuotInt64().Immut() instead.
func (d LegacyDec) QuoInt64Mut(i int64) LegacyDec { return d.Mut().QuoInt64(i).Immut() }

// PowerMut returns the result of raising to a positive integer power
// mutates the LegacyDec.
// Deprecated: Use Mut().Power().Immut() instead.
func (d LegacyDec) PowerMut(power uint64) LegacyDec {
	return d.Mut().Power(power).Immut()
}

// MulRoundUpMut multiplies d by d2, mutable, rounds up. Returns itself.
// Deprecated: use Mut().MulRoundUp(d2).Immut() instead
func (d LegacyDec) MulRoundUpMut(d2 LegacyDec) LegacyDec { return d.Mut().MulRoundUp(d2).Immut() }

// MulIntMut multiplies d by Int, mutable. Returns itself.
// Deprecated: use Mut().MulInt(i).Immut() instead
func (d LegacyDec) MulIntMut(i Int) LegacyDec { return d.Mut().MulInt(i).Immut() }

// MulInt64Mut multiplies d by int64, mutable. Returns itself.
// Deprecated: use Mut().MulInt64(i).Immut() instead
func (d LegacyDec) MulInt64Mut(i int64) LegacyDec { return d.Mut().MulInt64(i).Immut() }
