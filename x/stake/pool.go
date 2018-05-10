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

// get the exchange rate of unbonded tokens held in validators per issued share
func (p Pool) unbondedShareExRate() sdk.Rat {
	if p.UnbondedShares.IsZero() {
		return sdk.OneRat()
	}
	return sdk.NewRat(p.UnbondedPool).Quo(p.UnbondedShares)
}

// move a validators asset pool from bonded to unbonded pool
func (p Pool) bondedToUnbondedPool(validator Validator) (Pool, Validator) {

	// replace bonded shares with unbonded shares
	p, tokens := p.removeSharesBonded(validator.BondedShares)
	p, validator.BondedShares = p.addTokensUnbonded(tokens)
	validator.Status = Unbonded
	return p, validator
}

// move a validators asset pool from unbonded to bonded pool
func (p Pool) unbondedToBondedPool(validator Validator) (Pool, Validator) {

	// replace unbonded shares with bonded shares
	p, tokens := p.removeSharesUnbonded(validator.BondedShares)
	p, validator.BondedShares = p.addTokensBonded(tokens)
	validator.Status = Bonded
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

	exRate := validator.delegatorShareExRate()

	var receivedGlobalShares sdk.Rat
	if validator.Status == Bonded {
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

	//exRate := validator.delegatorShareExRate() //XXX make sure not used

	globalPoolSharesToRemove := validator.delegatorShareExRate().Mul(shares)
	if validator.Status == Bonded {
		p, createdCoins = p.removeSharesBonded(globalPoolSharesToRemove)
	} else {
		p, createdCoins = p.removeSharesUnbonded(globalPoolSharesToRemove)
	}
	validator.BondedShares = validator.BondedShares.Sub(globalPoolSharesToRemove)
	validator.DelegatorShares = validator.DelegatorShares.Sub(shares)
	return p, validator, createdCoins
}
