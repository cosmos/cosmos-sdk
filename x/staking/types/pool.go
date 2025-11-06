// Package types defines the core data structures for the staking module.
// This file provides pool types that track bonded and unbonded token balances
// in the staking module's token pools.
package types

import (
	"cosmossdk.io/math"
)

// Pool account names used for tracking token balances:
//
// - NotBondedPoolName -> "not_bonded_tokens_pool" - tracks tokens that are not bonded to validators
//
// - BondedPoolName -> "bonded_tokens_pool" - tracks tokens that are bonded to validators
const (
	NotBondedPoolName = "not_bonded_tokens_pool"
	BondedPoolName    = "bonded_tokens_pool"
)

// NewPool creates a new Pool instance with the given bonded and unbonded token amounts.
// The pool is used for querying the current state of bonded and unbonded tokens in the staking module.
func NewPool(notBonded, bonded math.Int) Pool {
	return Pool{
		NotBondedTokens: notBonded,
		BondedTokens:    bonded,
	}
}
