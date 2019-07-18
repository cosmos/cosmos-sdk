package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// names used as root for pool module accounts:
//
// - NotBondedPool -> "not_bonded_tokens_pool"
//
// - BondedPool -> "bonded_tokens_pool"
const (
	NotBondedPoolName = "not_bonded_tokens_pool"
	BondedPoolName    = "bonded_tokens_pool"
)

// Pool - tracking bonded and not-bonded token supply of the bond denomination
type Pool struct {
	NotBondedTokens sdk.Int `json:"not_bonded_tokens" yaml:"not_bonded_tokens"` // tokens which are not bonded to a validator (unbonded or unbonding)
	BondedTokens    sdk.Int `json:"bonded_tokens" yaml:"bonded_tokens"`         // tokens which are currently bonded to a validator
}

// NewPool creates a new Pool instance used for queries
func NewPool(notBonded, bonded sdk.Int) Pool {
	return Pool{
		NotBondedTokens: notBonded,
		BondedTokens:    bonded,
	}
}

// String returns a human readable string representation of a pool.
func (p Pool) String() string {
	return fmt.Sprintf(`Pool:	
  Not Bonded Tokens:  %s	
  Bonded Tokens:      %s`, p.NotBondedTokens,
		p.BondedTokens)
}
