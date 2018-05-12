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

// get the exchange rate of unbonding tokens held in validators per issued share
func (p Pool) unbondingShareExRate() sdk.Rat {
	if p.UnbondingShares.IsZero() {
		return sdk.OneRat()
	}
	return sdk.NewRat(p.UnbondingPool).Quo(p.UnbondingShares)
}

// get the exchange rate of unbonded tokens held in validators per issued share
func (p Pool) unbondedShareExRate() sdk.Rat {
	if p.UnbondedShares.IsZero() {
		return sdk.OneRat()
	}
	return sdk.NewRat(p.UnbondedPool).Quo(p.UnbondedShares)
}

// XXX write test
// update the location of the shares within a validator if its bond status has changed
func (p Pool) UpdateSharesLocation(validator Validator) (Pool, Validator) {
	var tokens int64

	switch {
	case !validator.BondedShares.IsZero():
		if validator.Status == sdk.Bonded { // return if nothing needs switching
			return p, validator
		}
		p, tokens = p.removeSharesBonded(validator.BondedShares)
	case !validator.UnbondingShares.IsZero():
		if validator.Status == sdk.Unbonding {
			return p, validator
		}
		p, tokens = p.removeSharesUnbonding(validator.BondedShares)
	case !validator.UnbondedShares.IsZero():
		if validator.Status == sdk.Unbonding {
			return p, validator
		}
		p, tokens = p.removeSharesUnbonded(validator.BondedShares)
	}

	switch validator.Status {
	case sdk.Bonded:
		p, validator.BondedShares = p.addTokensBonded(tokens)
	case sdk.Unbonding:
		p, validator.UnbondingShares = p.addTokensUnbonding(tokens)
	case sdk.Unbonded, sdk.Revoked:
		p, validator.UnbondedShares = p.addTokensUnbonded(tokens)
	}

	return p, validator
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

func (p Pool) addTokensUnbonding(amount int64) (p2 Pool, issuedShares sdk.Rat) {
	issuedShares = sdk.NewRat(amount).Quo(p.unbondingShareExRate()) // tokens * (shares/tokens)
	p.UnbondingShares = p.UnbondingShares.Add(issuedShares)
	p.UnbondingPool += amount
	return p, issuedShares
}

func (p Pool) removeSharesUnbonding(shares sdk.Rat) (p2 Pool, removedTokens int64) {
	removedTokens = p.unbondingShareExRate().Mul(shares).Evaluate() // (tokens/shares) * shares
	p.UnbondingShares = p.UnbondingShares.Sub(shares)
	p.UnbondingPool -= removedTokens
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

// add tokens to a validator
func (p Pool) validatorAddTokens(validator Validator,
	amount int64) (p2 Pool, validator2 Validator, issuedDelegatorShares sdk.Rat) {

	exRate := validator.DelegatorShareExRate()

	var receivedGlobalShares sdk.Rat
	if validator.Status == sdk.Bonded {
		p, receivedGlobalShares = p.addTokensBonded(amount)
	} else {
		p, receivedGlobalShares = p.addTokensUnbonded(amount)
	}
	validator.BondedShares = validator.BondedShares.Add(receivedGlobalShares)

	issuedDelegatorShares = exRate.Mul(receivedGlobalShares)
	validator.DelegatorShares = validator.DelegatorShares.Add(issuedDelegatorShares)

	return p, validator, issuedDelegatorShares
}

// remove shares from a validator
func (p Pool) validatorRemoveShares(validator Validator,
	shares sdk.Rat) (p2 Pool, validator2 Validator, createdCoins int64) {

	//exRate := validator.DelegatorShareExRate() //XXX make sure not used

	globalPoolSharesToRemove := validator.DelegatorShareExRate().Mul(shares)
	if validator.Status == sdk.Bonded {
		p, createdCoins = p.removeSharesBonded(globalPoolSharesToRemove)
	} else {
		p, createdCoins = p.removeSharesUnbonded(globalPoolSharesToRemove)
	}
	validator.BondedShares = validator.BondedShares.Sub(globalPoolSharesToRemove)
	validator.DelegatorShares = validator.DelegatorShares.Sub(shares)
	return p, validator, createdCoins
}
