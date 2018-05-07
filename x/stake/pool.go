package stake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// get the bond ratio of the global state
func (p Pool) bondedRatio() sdk.Rat {
	if p.TotalSupply > 0 {
		return sdk.NewRat(p.BondedPool, p.TotalSupply)
	}
	return sdk.ZeroRat()
}

// get the exchange rate of bonded token per issued share
func (p Pool) bondedShareExRate() sdk.Rat {
	if p.BondedShares.IsZero() {
		return sdk.OneRat()
	}
	return sdk.NewRat(p.BondedPool).Quo(p.BondedShares)
}

// get the exchange rate of unbonded tokens held in candidates per issued share
func (p Pool) unbondedShareExRate() sdk.Rat {
	if p.UnbondedShares.IsZero() {
		return sdk.OneRat()
	}
	return sdk.NewRat(p.UnbondedPool).Quo(p.UnbondedShares)
}

// move a candidates asset pool from bonded to unbonded pool
func (p Pool) bondedToUnbondedPool(candidate Candidate) (Pool, Candidate) {

	// replace bonded shares with unbonded shares
	p, tokens := p.removeSharesBonded(candidate.Assets)
	p, candidate.Assets = p.addTokensUnbonded(tokens)
	candidate.Status = Unbonded
	return p, candidate
}

// move a candidates asset pool from unbonded to bonded pool
func (p Pool) unbondedToBondedPool(candidate Candidate) (Pool, Candidate) {

	// replace unbonded shares with bonded shares
	p, tokens := p.removeSharesUnbonded(candidate.Assets)
	p, candidate.Assets = p.addTokensBonded(tokens)
	candidate.Status = Bonded
	return p, candidate
}

//_______________________________________________________________________

func (p Pool) addTokensBonded(amount int64) (p2 Pool, issuedShares sdk.Rat) {
	issuedShares = sdk.NewRat(amount).Quo(p.bondedShareExRate()) // tokens * (shares/tokens)
	p.BondedPool += amount
	p.BondedShares = p.BondedShares.Add(issuedShares)
	return p, issuedShares
}

func (p Pool) removeSharesBonded(shares sdk.Rat) (p2 Pool, removedTokens int64) {
	removedTokens = p.bondedShareExRate().Mul(shares).Evaluate() // (tokens/shares) * shares
	p.BondedShares = p.BondedShares.Sub(shares)
	p.BondedPool = p.BondedPool - removedTokens
	return p, removedTokens
}

func (p Pool) addTokensUnbonded(amount int64) (p2 Pool, issuedShares sdk.Rat) {
	issuedShares = sdk.NewRat(amount).Quo(p.unbondedShareExRate()) // tokens * (shares/tokens)
	p.UnbondedShares = p.UnbondedShares.Add(issuedShares)
	p.UnbondedPool += amount
	return p, issuedShares
}

func (p Pool) removeSharesUnbonded(shares sdk.Rat) (p2 Pool, removedTokens int64) {
	removedTokens = p.unbondedShareExRate().Mul(shares).Evaluate() // (tokens/shares) * shares
	p.UnbondedShares = p.UnbondedShares.Sub(shares)
	p.UnbondedPool -= removedTokens
	return p, removedTokens
}

//_______________________________________________________________________

// add tokens to a candidate
func (p Pool) candidateAddTokens(candidate Candidate,
	amount int64) (p2 Pool, candidate2 Candidate, issuedDelegatorShares sdk.Rat) {

	exRate := candidate.delegatorShareExRate()

	var receivedGlobalShares sdk.Rat
	if candidate.Status == Bonded {
		p, receivedGlobalShares = p.addTokensBonded(amount)
	} else {
		p, receivedGlobalShares = p.addTokensUnbonded(amount)
	}
	candidate.Assets = candidate.Assets.Add(receivedGlobalShares)

	issuedDelegatorShares = exRate.Mul(receivedGlobalShares)
	candidate.Liabilities = candidate.Liabilities.Add(issuedDelegatorShares)

	return p, candidate, issuedDelegatorShares
}

// remove shares from a candidate
func (p Pool) candidateRemoveShares(candidate Candidate,
	shares sdk.Rat) (p2 Pool, candidate2 Candidate, createdCoins int64) {

	//exRate := candidate.delegatorShareExRate() //XXX make sure not used

	globalPoolSharesToRemove := candidate.delegatorShareExRate().Mul(shares)
	if candidate.Status == Bonded {
		p, createdCoins = p.removeSharesBonded(globalPoolSharesToRemove)
	} else {
		p, createdCoins = p.removeSharesUnbonded(globalPoolSharesToRemove)
	}
	candidate.Assets = candidate.Assets.Sub(globalPoolSharesToRemove)
	candidate.Liabilities = candidate.Liabilities.Sub(shares)
	return p, candidate, createdCoins
}
