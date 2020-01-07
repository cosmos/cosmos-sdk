package types

import (
	"fmt"
)

// denomUnits contains a mapping of denomination mapped to their respective unit
// multipliers (e.g. 1atom = 10^-6uatom).
var denomUnits = map[string]Dec{}

// RegisterDenom registers a denomination with a corresponding unit. If the
// denomination is already registered, an error will be returned.
func RegisterDenom(denom string, unit Dec) error {
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
func GetDenomUnit(denom string) (Dec, bool) {
	if err := validateDenom(denom); err != nil {
		return ZeroDec(), false
	}

	unit, ok := denomUnits[denom]
	if !ok {
		return ZeroDec(), false
	}

	return unit, true
}

// ConvertCoin attempts to convert a coin to a given denomination. If the given
// denomination is invalid or if neither denomination is registered, an error
// is returned.
func ConvertCoin(coin Coin, denom string) (Coin, error) {
	if err := validateDenom(denom); err != nil {
		return Coin{}, err
	}

	srcUnit, ok := GetDenomUnit(coin.Denom)
	if !ok {
		return Coin{}, fmt.Errorf("source denom not registered: %s", coin.Denom)
	}

	dstUnit, ok := GetDenomUnit(denom)
	if !ok {
		return Coin{}, fmt.Errorf("destination denom not registered: %s", denom)
	}

	if srcUnit.Equal(dstUnit) {
		return NewCoin(denom, coin.Amount), nil
	}

	return NewCoin(denom, coin.Amount.ToDec().Mul(srcUnit.Quo(dstUnit)).TruncateInt()), nil
}
