package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// names used as root for pool module accounts:
//
// - NotBondedPool -> "NotBondedTokensPool"
//
// - BondedPool -> "BondedTokensPool"
const (
	NotBondedPoolName = "NotBondedTokensPool"
	BondedPoolName    = "BondedTokensPool"
)

// Pool - tracking bonded and not-bonded token supply of the bond denomination
type Pool struct {
	NotBondedTokens sdk.Int `json:"not_bonded_tokens"` // tokens which are not bonded to a validator (unbonded or unbonding)
	BondedTokens    sdk.Int `json:"bonded_tokens"`     // tokens which are currently bonded to a validator
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
