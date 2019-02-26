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

// DenomUnits contains a mapping of denomination mapped to their respective unit
// multipliers.
var DenomUnits = map[string]Int{
	Atom:  OneInt(),
	Uatom: NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil)),
	Matom: NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(5), nil)),
	// further units to be supported...
}

// ToUatoms converts a coin to uatoms. It panics if the given coin does not have
// a denom unit multiplier.
func ToUatoms(coin Coin) Coin {
	if coin.Denom == Uatom {
		return coin
	}

	unit, ok := DenomUnits[coin.Denom]
	if !ok {
		panic(fmt.Sprintf("unsupported coin type: %s", coin.Denom))
	}

	return NewCoin(Uatom, coin.Amount.Mul(DenomUnits["uatom"].Quo(unit)))
}

// FromUatoms converts a uatom coin to another coin specified by the given denom.
// It panics if the denom does not have a unit multiplier.
func FromUatoms(coin Coin, denom string) Coin {
	if coin.Denom != Uatom {
		panic(fmt.Sprintf("invalid coin type: %s", coin.Denom))
	}

	if denom == Uatom {
		return coin
	}

	unit, ok := DenomUnits[denom]
	if !ok {
		panic(fmt.Sprintf("unsupported coin type: %s", denom))
	}

	return NewCoin(denom, coin.Amount.Quo(DenomUnits[Uatom].Quo(unit)))
}
