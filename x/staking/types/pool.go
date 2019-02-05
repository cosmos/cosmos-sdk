package types

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Pool - dynamic parameters of the current state
type Pool struct {
	NotBondedTokens sdk.Int `json:"not_bonded_tokens"` // tokens which are not bonded in a validator
	BondedTokens    sdk.Int `json:"bonded_tokens"`     // reserve of bonded tokens
}

// nolint
func (p Pool) Equal(p2 Pool) bool {
	bz1 := MsgCdc.MustMarshalBinaryLengthPrefixed(&p)
	bz2 := MsgCdc.MustMarshalBinaryLengthPrefixed(&p2)
	return bytes.Equal(bz1, bz2)
}

// initial pool for testing
func InitialPool() Pool {
	return Pool{
		NotBondedTokens: sdk.ZeroInt(),
		BondedTokens:    sdk.ZeroInt(),
	}
}

//____________________________________________________________________

// Sum total of all staking tokens in the pool
func (p Pool) TokenSupply() sdk.Int {
	return p.NotBondedTokens.Add(p.BondedTokens)
}

//____________________________________________________________________

// get the bond ratio of the global state
func (p Pool) BondedRatio() sdk.Dec {
	supply := p.TokenSupply()
	if supply.IsPositive() {
		return sdk.NewDecFromInt(p.BondedTokens).
			QuoInt(supply)
	}
	return sdk.ZeroDec()
}

//_______________________________________________________________________

func (p Pool) notBondedTokensToBonded(bondedTokens sdk.Int) Pool {
	p.BondedTokens = p.BondedTokens.Add(bondedTokens)
	p.NotBondedTokens = p.NotBondedTokens.Sub(bondedTokens)
	if p.NotBondedTokens.IsNegative() {
		panic(fmt.Sprintf("sanity check: not-bonded tokens negative, pool: %v", p))
	}
	return p
}

func (p Pool) bondedTokensToNotBonded(bondedTokens sdk.Int) Pool {
	p.BondedTokens = p.BondedTokens.Sub(bondedTokens)
	p.NotBondedTokens = p.NotBondedTokens.Add(bondedTokens)
	if p.BondedTokens.IsNegative() {
		panic(fmt.Sprintf("sanity check: bonded tokens negative, pool: %v", p))
	}
	return p
}

// String returns a human readable string representation of a pool.
func (p Pool) String() string {
	return fmt.Sprintf(`Pool:
  Loose Tokens:  %s
  Bonded Tokens: %s
  Token Supply:  %s
  Bonded Ratio:  %v`, p.NotBondedTokens,
		p.BondedTokens, p.TokenSupply(),
		p.BondedRatio())
}

// unmarshal the current pool value from store key or panics
func MustUnmarshalPool(cdc *codec.Codec, value []byte) Pool {
	pool, err := UnmarshalPool(cdc, value)
	if err != nil {
		panic(err)
	}
	return pool
}

// unmarshal the current pool value from store key
func UnmarshalPool(cdc *codec.Codec, value []byte) (pool Pool, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(value, &pool)
	if err != nil {
		return
	}
	return
}
