package types

import (
	"math/big"
)

// staking constants
const (

	// default bond denomination
	DefaultBondDenom = "stake"
)

// PowerReduction is the amount of staking tokens required for 1 unit of Tendermint power
var PowerReduction = NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil))

// TokensToTendermintPower - convert input tokens to potential tendermint power
func TokensToTendermintPower(tokens Int) int64 {
	return (tokens.Quo(PowerReduction)).Int64()
}

// TokensFromTendermintPower - convert input power to tokens
func TokensFromTendermintPower(power int64) Int {
	return NewInt(power).Mul(PowerReduction)
}
