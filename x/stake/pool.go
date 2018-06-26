package stake

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Pool - dynamic parameters of the current state
type Pool struct {
	LooseUnbondedTokens sdk.Int `json:"loose_unbonded_tokens"` // tokens not associated with any validator
	UnbondedTokens      sdk.Int `json:"unbonded_tokens"`       // reserve of unbonded tokens held with validators
	UnbondingTokens     sdk.Int `json:"unbonding_tokens"`      // tokens moving from bonded to unbonded pool
	BondedTokens        sdk.Int `json:"bonded_tokens"`         // reserve of bonded tokens
	UnbondedShares      sdk.Rat `json:"unbonded_shares"`       // sum of all shares distributed for the Unbonded Pool
	UnbondingShares     sdk.Rat `json:"unbonding_shares"`      // shares moving from Bonded to Unbonded Pool
	BondedShares        sdk.Rat `json:"bonded_shares"`         // sum of all shares distributed for the Bonded Pool
	InflationLastTime   int64   `json:"inflation_last_time"`   // block which the last inflation was processed // TODO make time
	Inflation           sdk.Rat `json:"inflation"`             // current annual inflation rate

	DateLastCommissionReset int64 `json:"date_last_commission_reset"` // unix timestamp for last commission accounting reset (daily)

	// Fee Related
	PrevBondedShares sdk.Rat `json:"prev_bonded_shares"` // last recorded bonded shares - for fee calcualtions
}

func (p Pool) equal(p2 Pool) bool {
	bz1 := msgCdc.MustMarshalBinary(&p)
	bz2 := msgCdc.MustMarshalBinary(&p2)
	return bytes.Equal(bz1, bz2)
}

// initial pool for testing
func InitialPool() Pool {
	return Pool{
		LooseUnbondedTokens:     sdk.ZeroInt(),
		BondedTokens:            sdk.ZeroInt(),
		UnbondingTokens:         sdk.ZeroInt(),
		UnbondedTokens:          sdk.ZeroInt(),
		BondedShares:            sdk.ZeroRat(),
		UnbondingShares:         sdk.ZeroRat(),
		UnbondedShares:          sdk.ZeroRat(),
		InflationLastTime:       0,
		Inflation:               sdk.NewRat(7, 100),
		DateLastCommissionReset: 0,
		PrevBondedShares:        sdk.ZeroRat(),
	}
}

//____________________________________________________________________

// Sum total of all staking tokens in the pool
func (p Pool) TokenSupply() sdk.Int {
	return p.LooseUnbondedTokens.Add(p.UnbondedTokens).Add(p.UnbondingTokens).Add(p.BondedTokens)
}

//____________________________________________________________________

// get the bond ratio of the global state
func (p Pool) bondedRatio() sdk.Rat {
	if p.TokenSupply().Sign() == 1 {
		return sdk.NewRatFromInt(p.BondedTokens, p.TokenSupply())
	}
	return sdk.ZeroRat()
}

// get the exchange rate of bonded token per issued share
func (p Pool) bondedShareExRate() sdk.Rat {
	if p.BondedShares.IsZero() {
		return sdk.OneRat()
	}
	return sdk.NewRatFromInt(p.BondedTokens).Quo(p.BondedShares)
}

// get the exchange rate of unbonding tokens held in validators per issued share
func (p Pool) unbondingShareExRate() sdk.Rat {
	if p.UnbondingShares.IsZero() {
		return sdk.OneRat()
	}
	return sdk.NewRatFromInt(p.UnbondingTokens).Quo(p.UnbondingShares)
}

// get the exchange rate of unbonded tokens held in validators per issued share
func (p Pool) unbondedShareExRate() sdk.Rat {
	if p.UnbondedShares.IsZero() {
		return sdk.OneRat()
	}
	return sdk.NewRatFromInt(p.UnbondedTokens).Quo(p.UnbondedShares)
}

//_______________________________________________________________________

func (p Pool) addTokensUnbonded(amount sdk.Int) (p2 Pool, issuedShares PoolShares) {
	issuedSharesAmount := sdk.NewRatFromInt(amount).Quo(p.unbondedShareExRate()) // tokens * (shares/tokens)
	p.UnbondedShares = p.UnbondedShares.Add(issuedSharesAmount)
	p.UnbondedTokens = p.UnbondedTokens.Add(amount)
	return p, NewUnbondedShares(issuedSharesAmount)
}

func (p Pool) removeSharesUnbonded(shares sdk.Rat) (p2 Pool, removedTokens sdk.Int) {
	removedTokens = sdk.NewIntFromBigInt(p.unbondedShareExRate().Mul(shares).EvaluateBig()) // (tokens/shares) * shares
	p.UnbondedShares = p.UnbondedShares.Sub(shares)
	p.UnbondedTokens = p.UnbondedTokens.Sub(removedTokens)
	return p, removedTokens
}

func (p Pool) addTokensUnbonding(amount sdk.Int) (p2 Pool, issuedShares PoolShares) {
	issuedSharesAmount := sdk.NewRatFromInt(amount).Quo(p.unbondingShareExRate()) // tokens * (shares/tokens)
	p.UnbondingShares = p.UnbondingShares.Add(issuedSharesAmount)
	p.UnbondingTokens = p.UnbondingTokens.Add(amount)
	return p, NewUnbondingShares(issuedSharesAmount)
}

func (p Pool) removeSharesUnbonding(shares sdk.Rat) (p2 Pool, removedTokens sdk.Int) {
	removedTokens = sdk.NewIntFromBigInt(p.unbondingShareExRate().Mul(shares).EvaluateBig()) // (tokens/shares) * shares
	p.UnbondingShares = p.UnbondingShares.Sub(shares)
	p.UnbondingTokens = p.UnbondingTokens.Sub(removedTokens)
	return p, removedTokens
}

func (p Pool) addTokensBonded(amount sdk.Int) (p2 Pool, issuedShares PoolShares) {
	issuedSharesAmount := sdk.NewRatFromInt(amount).Quo(p.bondedShareExRate()) // tokens * (shares/tokens)
	p.BondedShares = p.BondedShares.Add(issuedSharesAmount)
	p.BondedTokens = p.BondedTokens.Add(amount)
	return p, NewBondedShares(issuedSharesAmount)
}

func (p Pool) removeSharesBonded(shares sdk.Rat) (p2 Pool, removedTokens sdk.Int) {
	removedTokens = sdk.NewIntFromBigInt(p.bondedShareExRate().Mul(shares).EvaluateBig()) // (tokens/shares) * shares
	p.BondedShares = p.BondedShares.Sub(shares)
	p.BondedTokens = p.BondedTokens.Sub(removedTokens)
	return p, removedTokens
}
