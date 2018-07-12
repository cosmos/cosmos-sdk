package types

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Pool - dynamic parameters of the current state
type Pool struct {
	LooseTokens       sdk.Rat `json:"loose_tokens"`        // tokens which are not bonded in a validator
	BondedTokens      sdk.Rat `json:"bonded_tokens"`       // reserve of bonded tokens
	InflationLastTime int64   `json:"inflation_last_time"` // block which the last inflation was processed // TODO make time
	Inflation         sdk.Rat `json:"inflation"`           // current annual inflation rate

	DateLastCommissionReset int64 `json:"date_last_commission_reset"` // unix timestamp for last commission accounting reset (daily)

	// Fee Related
	PrevBondedShares sdk.Rat `json:"prev_bonded_shares"` // last recorded bonded shares - for fee calculations
}

// nolint
func (p Pool) Equal(p2 Pool) bool {
	bz1 := MsgCdc.MustMarshalBinary(&p)
	bz2 := MsgCdc.MustMarshalBinary(&p2)
	return bytes.Equal(bz1, bz2)
}

// initial pool for testing
func InitialPool() Pool {
	return Pool{
		LooseTokens:             sdk.ZeroRat(),
		BondedTokens:            sdk.ZeroRat(),
		InflationLastTime:       0,
		Inflation:               sdk.NewRat(7, 100),
		DateLastCommissionReset: 0,
		PrevBondedShares:        sdk.ZeroRat(),
	}
}

//____________________________________________________________________

// Sum total of all staking tokens in the pool
func (p Pool) TokenSupply() sdk.Rat {
	return p.LooseTokens.Add(p.BondedTokens)
}

//____________________________________________________________________

// get the bond ratio of the global state
func (p Pool) BondedRatio() sdk.Rat {
	if p.TokenSupply().GT(sdk.ZeroRat()) {
		return p.BondedTokens.Quo(p.TokenSupply())
	}
	return sdk.ZeroRat()
}

//_______________________________________________________________________

func (p Pool) addBondedTokens(bondedTokens sdk.Rat) Pool {
	p.BondedTokens = p.BondedTokens.Add(bondedTokens)
	p.LooseTokens = p.LooseTokens.Sub(bondedTokens)
	if p.LooseTokens.LT(sdk.ZeroRat()) {
		panic(fmt.Sprintf("sanity check: loose tokens negative, pool: %v", p))
	}
	return p
}

func (p Pool) removeBondedTokens(bondedTokens sdk.Rat) Pool {
	p.BondedTokens = p.BondedTokens.Sub(bondedTokens)
	p.LooseTokens = p.LooseTokens.Add(bondedTokens)
	if p.BondedTokens.LT(sdk.ZeroRat()) {
		panic(fmt.Sprintf("sanity check: bonded tokens negative, pool: %v", p))
	}
	return p
}
