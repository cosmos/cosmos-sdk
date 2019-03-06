package types

import (
	"fmt"
	"math/big"
)

// Denomination constants
const (
	Uatom = "uatom"
	Matom = "matom"
	Atom  = "atom"
)

// denomUnits contains a mapping of denomination mapped to their respective unit
// multipliers.
var denomUnits = map[string]Int{
	Atom:  OneInt(),
	Uatom: NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil)),
	Matom: NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(5), nil)),
}

// RegisterDenom registers a denomination with a corresponding unit. If the
// denomination is already registered, an error will be returned.
func RegisterDenom(denom string, unit Int) error {
	if err := validateDenom(denom); err != nil {
		return err
	}

	if _, ok := denomUnits[denom]; ok {
		return fmt.Errorf("denom %s already registered", denom)
	}

	denomUnits[denom] = unit
	return nil
}

// GetDenomUnit returns a unit for a given denomination if it exists. A boolean
// is returned if the denomination is registered.
func GetDenomUnit(denom string) (Int, bool) {
	if err := validateDenom(denom); err != nil {
		return ZeroInt(), false
	}

	unit, ok := denomUnits[denom]
	return unit, ok
}

// ConvertCoins attempts to convert a coin to a given denomination. If the given
// denomination is invalid or if neither denomination is registered, an error
// is returned.
func ConvertCoins(coin Coin, denom string) (Coin, error) {
	if err := validateDenom(denom); err != nil {
		return Coin{}, err
	}

	srcUnit, ok := denomUnits[coin.Denom]
	if !ok {
		return Coin{}, fmt.Errorf("source denom not registered: %s", coin.Denom)
	}

	dstUnit, ok := denomUnits[denom]
	if !ok {
		return Coin{}, fmt.Errorf("destination denom not registered: %s", coin.Denom)
	}

	if srcUnit.Equal(dstUnit) {
		return NewCoin(denom, coin.Amount), nil
	}

	if srcUnit.LT(dstUnit) {
		return NewCoin(denom, coin.Amount.Quo(srcUnit.Quo(dstUnit))), nil
	}

	return NewCoin(Uatom, coin.Amount.Mul(dstUnit.Quo(srcUnit))), nil
}
