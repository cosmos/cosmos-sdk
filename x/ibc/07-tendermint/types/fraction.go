package types

import (
	tmmath "github.com/tendermint/tendermint/libs/math"
)

// NewFractionFromTm returns a new Fraction instance from a tmmath.Fraction
func NewFractionFromTm(f tmmath.Fraction) Fraction {
	return Fraction{
		Numerator:   f.Numerator,
		Denominator: f.Denominator,
	}
}

// ToTendermint converts Fraction to tmmath.Fraction
func (f Fraction) ToTendermint() tmmath.Fraction {
	return tmmath.Fraction{
		Numerator:   f.Numerator,
		Denominator: f.Denominator,
	}
}
