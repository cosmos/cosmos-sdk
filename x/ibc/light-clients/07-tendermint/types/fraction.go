package types

import (
	tmmath "github.com/tendermint/tendermint/libs/math"
	"github.com/tendermint/tendermint/light"
)

// DefaultTrustLevel is the tendermint light client default trust level
var DefaultTrustLevel = NewFractionFromTm(light.DefaultTrustLevel)

// NewFractionFromTm returns a new Fraction instance from a tmmath.Fraction
func NewFractionFromTm(f tmmath.Fraction) Fraction {
	return Fraction{
		Numerator:   uint64(f.Numerator),
		Denominator: uint64(f.Denominator),
	}
}

// ToTendermint converts Fraction to tmmath.Fraction
func (f Fraction) ToTendermint() tmmath.Fraction {
	return tmmath.Fraction{
		Numerator:   int64(f.Numerator),
		Denominator: int64(f.Denominator),
	}
}
