package mint

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Minter - dynamic parameters of the current state
type Params struct {
	MintDenom           String  `json:"mint_denom"`            // type of coin to mint
	InflationRateChange sdk.Dec `json:"inflation_rate_change"` // maximum annual change in inflation rate
	InflationMax        sdk.Dec `json:"inflation_max"`         // maximum inflation rate
	InflationMin        sdk.Dec `json:"inflation_min"`         // minimum inflation rate
	GoalBonded          sdk.Dec `json:"goal_bonded"`           // goal of percent bonded atoms
}

func DefaultParams() Params {
	return Params{
		MintDenom:           "stake",
		InflationRateChange: sdk.NewDecWithPrec(13, 2),
		InflationMax:        sdk.NewDecWithPrec(20, 2),
		InflationMin:        sdk.NewDecWithPrec(7, 2),
		GoalBonded:          sdk.NewDecWithPrec(67, 2),
	}
}

//______________________________________________________________

// Minter - dynamic parameters of the current state
type Minter struct {
	InflationLastTime time.Time `json:"inflation_last_time"` // block time which the last inflation was processed
	Inflation         sdk.Dec   `json:"inflation"`           // current annual inflation rate
}

func (m Minter) InitialMinter() Minter {
	return Minter{
		InflationLastTime: time.Unix(0, 0),
		Inflation:         sdk.NewDecWithPrec(13, 2),
	}
}

var hrsPerYr = sdk.NewDec(8766) // as defined by a julian year of 365.25 days

// process provisions for an hour period
func (m Minter) ProcessProvisions(params Params) Pool {
	m.Inflation = m.NextInflation(params)
	provisions := m.Inflation.
		Mul(m.TokenSupply()).
		Quo(hrsPerYr)

	m.LooseTokens = m.LooseTokens.Add(provisions)
	return m
}

// get the next inflation rate for the hour
func (m Minter) NextInflation(params Params) (inflation sdk.Dec) {

	// The target annual inflation rate is recalculated for each previsions cycle. The
	// inflation is also subject to a rate change (positive or negative) depending on
	// the distance from the desired ratio (67%). The maximum rate change possible is
	// defined to be 13% per year, however the annual inflation is capped as between
	// 7% and 20%.

	// (1 - bondedRatio/GoalBonded) * InflationRateChange
	inflationRateChangePerYear := sdk.OneDec().
		Sub(m.BondedRatio().Quo(params.GoalBonded)).
		Mul(params.InflationRateChange)
	inflationRateChange := inflationRateChangePerYear.Quo(hrsPerYr)

	// increase the new annual inflation for this next cycle
	inflation = m.Inflation.Add(inflationRateChange)
	if inflation.GT(params.InflationMax) {
		inflation = params.InflationMax
	}
	if inflation.LT(params.InflationMin) {
		inflation = params.InflationMin
	}

	return inflation
}
