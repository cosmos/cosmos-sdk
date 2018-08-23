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
	supply := p.TokenSupply()
	if supply.GT(sdk.ZeroRat()) {
		return p.BondedTokens.Quo(supply)
	}
	return sdk.ZeroRat()
}

//_______________________________________________________________________

func (p Pool) looseTokensToBonded(bondedTokens sdk.Rat) Pool {
	p.BondedTokens = p.BondedTokens.Add(bondedTokens)
	p.LooseTokens = p.LooseTokens.Sub(bondedTokens)
	if p.LooseTokens.LT(sdk.ZeroRat()) {
		panic(fmt.Sprintf("sanity check: loose tokens negative, pool: %v", p))
	}
	return p
}

func (p Pool) bondedTokensToLoose(bondedTokens sdk.Rat) Pool {
	p.BondedTokens = p.BondedTokens.Sub(bondedTokens)
	p.LooseTokens = p.LooseTokens.Add(bondedTokens)
	if p.BondedTokens.LT(sdk.ZeroRat()) {
		panic(fmt.Sprintf("sanity check: bonded tokens negative, pool: %v", p))
	}
	return p
}

//_______________________________________________________________________
// Inflation

const precision = 100000000000     // increased to this precision for accuracy
var hrsPerYrRat = sdk.NewRat(8766) // as defined by a julian year of 365.25 days

// process provisions for an hour period
func (p Pool) ProcessProvisions(params Params) Pool {
	p.Inflation = p.NextInflation(params)
	provisions := p.Inflation.Mul(p.TokenSupply()).Quo(hrsPerYrRat)

	// TODO add to the fees provisions
	p.LooseTokens = p.LooseTokens.Add(provisions)
	return p
}

// get the next inflation rate for the hour
func (p Pool) NextInflation(params Params) (inflation sdk.Rat) {

	// The target annual inflation rate is recalculated for each previsions cycle. The
	// inflation is also subject to a rate change (positive or negative) depending on
	// the distance from the desired ratio (67%). The maximum rate change possible is
	// defined to be 13% per year, however the annual inflation is capped as between
	// 7% and 20%.

	// (1 - bondedRatio/GoalBonded) * InflationRateChange
	inflationRateChangePerYear := sdk.OneRat().Sub(p.BondedRatio().Quo(params.GoalBonded)).Mul(params.InflationRateChange)
	inflationRateChange := inflationRateChangePerYear.Quo(hrsPerYrRat)

	// increase the new annual inflation for this next cycle
	inflation = p.Inflation.Add(inflationRateChange)
	if inflation.GT(params.InflationMax) {
		inflation = params.InflationMax
	}
	if inflation.LT(params.InflationMin) {
		inflation = params.InflationMin
	}

	return inflation.Round(precision)
}
