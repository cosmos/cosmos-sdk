package stake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Pool - dynamic parameters of the current state
type Pool struct {
	TotalSupply       int64   `json:"total_supply"`        // total supply of all tokens
	UnbondedShares    sdk.Rat `json:"unbonded_shares"`     // sum of all shares distributed for the Unbonded Pool
	UnbondingShares   sdk.Rat `json:"unbonding_shares"`    // shares moving from Bonded to Unbonded Pool
	BondedShares      sdk.Rat `json:"bonded_shares"`       // sum of all shares distributed for the Bonded Pool
	UnbondedTokens    int64   `json:"unbonded_pool"`       // reserve of unbonded tokens held with validators
	UnbondingTokens   int64   `json:"unbonding_pool"`      // tokens moving from bonded to unbonded pool
	BondedTokens      int64   `json:"bonded_pool"`         // reserve of bonded tokens
	InflationLastTime int64   `json:"inflation_last_time"` // block which the last inflation was processed // TODO make time
	Inflation         sdk.Rat `json:"inflation"`           // current annual inflation rate

	DateLastCommissionReset int64 `json:"date_last_commission_reset"` // unix timestamp for last commission accounting reset (daily)

	// Fee Related
	PrevBondedShares sdk.Rat `json:"prev_bonded_shares"` // last recorded bonded shares - for fee calcualtions
}

func (p Pool) equal(p2 Pool) bool {
	return p.TotalSupply == p2.TotalSupply &&
		p.BondedShares.Equal(p2.BondedShares) &&
		p.UnbondingShares.Equal(p2.UnbondingShares) &&
		p.UnbondedShares.Equal(p2.UnbondedShares) &&
		p.BondedTokens == p2.BondedTokens &&
		p.UnbondingTokens == p2.UnbondingTokens &&
		p.UnbondedTokens == p2.UnbondedTokens &&
		p.InflationLastTime == p2.InflationLastTime &&
		p.Inflation.Equal(p2.Inflation) &&
		p.DateLastCommissionReset == p2.DateLastCommissionReset &&
		p.PrevBondedShares.Equal(p2.PrevBondedShares)
}

// initial pool for testing
func initialPool() Pool {
	return Pool{
		TotalSupply:             0,
		BondedShares:            sdk.ZeroRat(),
		UnbondingShares:         sdk.ZeroRat(),
		UnbondedShares:          sdk.ZeroRat(),
		BondedTokens:            0,
		UnbondingTokens:         0,
		UnbondedTokens:          0,
		InflationLastTime:       0,
		Inflation:               sdk.NewRat(7, 100),
		DateLastCommissionReset: 0,
		PrevBondedShares:        sdk.ZeroRat(),
	}
}

//____________________________________________________________________

// get the bond ratio of the global state
func (p Pool) bondedRatio() sdk.Rat {
	if p.TotalSupply > 0 {
		return sdk.NewRat(p.BondedTokens, p.TotalSupply)
	}
	return sdk.ZeroRat()
}

// get the exchange rate of bonded token per issued share
func (p Pool) bondedShareExRate() sdk.Rat {
	if p.BondedShares.IsZero() {
		return sdk.OneRat()
	}
	return sdk.NewRat(p.BondedTokens).Quo(p.BondedShares)
}

// get the exchange rate of unbonding tokens held in validators per issued share
func (p Pool) unbondingShareExRate() sdk.Rat {
	if p.UnbondingShares.IsZero() {
		return sdk.OneRat()
	}
	return sdk.NewRat(p.UnbondingTokens).Quo(p.UnbondingShares)
}

// get the exchange rate of unbonded tokens held in validators per issued share
func (p Pool) unbondedShareExRate() sdk.Rat {
	if p.UnbondedShares.IsZero() {
		return sdk.OneRat()
	}
	return sdk.NewRat(p.UnbondedTokens).Quo(p.UnbondedShares)
}

//_______________________________________________________________________

func (p Pool) addTokensUnbonded(amount int64) (p2 Pool, issuedShares PoolShares) {
	issuedSharesAmount := sdk.NewRat(amount).Quo(p.unbondedShareExRate()) // tokens * (shares/tokens)
	p.UnbondedShares = p.UnbondedShares.Add(issuedSharesAmount)
	p.UnbondedTokens += amount
	return p, NewUnbondedShares(issuedSharesAmount)
}

func (p Pool) removeSharesUnbonded(shares sdk.Rat) (p2 Pool, removedTokens int64) {
	removedTokens = p.unbondedShareExRate().Mul(shares).Evaluate() // (tokens/shares) * shares
	p.UnbondedShares = p.UnbondedShares.Sub(shares)
	p.UnbondedTokens -= removedTokens
	return p, removedTokens
}

func (p Pool) addTokensUnbonding(amount int64) (p2 Pool, issuedShares PoolShares) {
	issuedSharesAmount := sdk.NewRat(amount).Quo(p.unbondingShareExRate()) // tokens * (shares/tokens)
	p.UnbondingShares = p.UnbondingShares.Add(issuedSharesAmount)
	p.UnbondingTokens += amount
	return p, NewUnbondingShares(issuedSharesAmount)
}

func (p Pool) removeSharesUnbonding(shares sdk.Rat) (p2 Pool, removedTokens int64) {
	removedTokens = p.unbondingShareExRate().Mul(shares).Evaluate() // (tokens/shares) * shares
	p.UnbondingShares = p.UnbondingShares.Sub(shares)
	p.UnbondingTokens -= removedTokens
	return p, removedTokens
}

func (p Pool) addTokensBonded(amount int64) (p2 Pool, issuedShares PoolShares) {
	issuedSharesAmount := sdk.NewRat(amount).Quo(p.bondedShareExRate()) // tokens * (shares/tokens)
	p.BondedShares = p.BondedShares.Add(issuedSharesAmount)
	p.BondedTokens += amount
	return p, NewBondedShares(issuedSharesAmount)
}

func (p Pool) removeSharesBonded(shares sdk.Rat) (p2 Pool, removedTokens int64) {
	removedTokens = p.bondedShareExRate().Mul(shares).Evaluate() // (tokens/shares) * shares
	p.BondedShares = p.BondedShares.Sub(shares)
	p.BondedTokens -= removedTokens
	return p, removedTokens
}
