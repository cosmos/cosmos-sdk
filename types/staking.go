package types

import (
	"math/big"
)

// staking constants
const (

	// default bond denomination
	DefaultBondDenom = "stake"

	// Delay, in blocks, between when validator updates are returned to Tendermint and when they are applied.
	// For example, if this is 0, the validator set at the end of a block will sign the next block, or
	// if this is 1, the validator set at the end of a block will sign the block after the next.
	// Constant as this should not change without a hard fork.
	// TODO: Link to some Tendermint docs, this is very unobvious.
	ValidatorUpdateDelay int64 = 1

	BondStatusUnbonded  = "Unbonded"
	BondStatusUnbonding = "Unbonding"
	BondStatusBonded    = "Bonded"
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
