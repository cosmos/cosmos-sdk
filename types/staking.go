package types

import (
	"math/big"
)

// staking constants
const (

	// default bond denomination
	DefaultBondDenom = "stake"

	// Delay, in blocks, between when validator updates are returned to the
	// consensus-engine and when they are applied. For example, if
	// ValidatorUpdateDelay is set to X, and if a validator set update is
	// returned with new validators at the end of block 10, then the new
	// validators are expected to sign blocks beginning at block 11+X.
	//
	// This value is constant as this should not change without a hard fork.
	// For Tendermint this should be set to 1 block, for more details see:
	// https://tendermint.com/docs/spec/abci/apps.html#endblock
	ValidatorUpdateDelay int64 = 1
)

// PowerReduction is the amount of staking tokens required for 1 unit of consensus-engine power
var PowerReduction = NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil))

// TokensToConsensusPower - convert input tokens to potential consensus-engine power
func TokensToConsensusPower(tokens Int) int64 {
	return (tokens.Quo(PowerReduction)).Int64()
}

// TokensFromConsensusPower - convert input power to tokens
func TokensFromConsensusPower(power int64) Int {
	return NewInt(power).Mul(PowerReduction)
}
