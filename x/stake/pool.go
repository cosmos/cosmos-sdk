package stake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

//TODO make these next two functions more efficient should be reading and writting to state ye know

// move a candidates asset pool from bonded to unbonded pool
func (p Pool) bondedToUnbondedPool(candidate Candidate) (Pool, Candidate) {

	// replace bonded shares with unbonded shares
	tokens := k.removeSharesBonded(ctx, candidate.Assets)
	candidate.Assets = k.addTokensUnbonded(ctx, tokens)
	candidate.Status = Unbonded
	return p, candidate
}

// move a candidates asset pool from unbonded to bonded pool
func (p Pool) unbondedToBondedPool(candidate Candidate) (Pool, Candidate) {

	// replace unbonded shares with bonded shares
	tokens := k.removeSharesUnbonded(ctx, candidate.Assets)
	candidate.Assets = k.addTokensBonded(ctx, tokens)
	candidate.Status = Bonded
	return p, candidate
}

//_______________________________________________________________________

func (p Pool) addTokensBonded(ctx sdk.Context, amount int64) (p Pool, issuedShares sdk.Rat) {
	issuedShares = p.bondedShareExRate().Inv().Mul(sdk.NewRat(amount)) // (tokens/shares)^-1 * tokens
	p.BondedPool += amount
	p.BondedShares = p.BondedShares.Add(issuedShares)
	return p, issuedShares
}

func (p Pool) removeSharesBonded(ctx sdk.Context, shares sdk.Rat) (p Pool, removedTokens int64) {
	removedTokens = p.bondedShareExRate().Mul(shares).Evaluate() // (tokens/shares) * shares
	p.BondedShares = p.BondedShares.Sub(shares)
	p.BondedPool -= removedTokens
	return p, removedTokens
}

func (p Pool) addTokensUnbonded(ctx sdk.Context, amount int64) (p Pool, issuedShares sdk.Rat) {
	issuedShares = p.unbondedShareExRate().Inv().Mul(sdk.NewRat(amount)) // (tokens/shares)^-1 * tokens
	p.UnbondedShares = p.UnbondedShares.Add(issuedShares)
	p.UnbondedPool += amount
	return p, issuedShares
}

func (p Pool) removeSharesUnbonded(ctx sdk.Context, shares sdk.Rat) (p Pool, removedTokens int64) {
	removedTokens = p.unbondedShareExRate().Mul(shares).Evaluate() // (tokens/shares) * shares
	p.UnbondedShares = p.UnbondedShares.Sub(shares)
	p.UnbondedPool -= removedTokens
	return p, removedTokens
}

//_______________________________________________________________________

// add tokens to a candidate
func (p Pool) candidateAddTokens(ctx sdk.Context, candidate Candidate,
	amount int64) (p Pool, candidate Candidate, issuedDelegatorShares sdk.Rat) {

	exRate := candidate.delegatorShareExRate()

	var receivedGlobalShares sdk.Rat
	if candidate.Status == Bonded {
		receivedGlobalShares = k.addTokensBonded(ctx, amount)
	} else {
		receivedGlobalShares = k.addTokensUnbonded(ctx, amount)
	}
	candidate.Assets = candidate.Assets.Add(receivedGlobalShares)

	issuedDelegatorShares = exRate.Mul(receivedGlobalShares)
	candidate.Liabilities = candidate.Liabilities.Add(issuedDelegatorShares)
	return p, candidate, issuedDelegatorShares
}

// remove shares from a candidate
func (p Pool) candidateRemoveShares(ctx sdk.Context, candidate Candidate,
	shares sdk.Rat) (p Pool, candidate Candidate, createdCoins int64) {

	//exRate := candidate.delegatorShareExRate() //XXX make sure not used

	globalPoolSharesToRemove := candidate.delegatorShareExRate().Mul(shares)
	if candidate.Status == Bonded {
		createdCoins = k.removeSharesBonded(ctx, globalPoolSharesToRemove)
	} else {
		createdCoins = k.removeSharesUnbonded(ctx, globalPoolSharesToRemove)
	}
	candidate.Assets = candidate.Assets.Sub(globalPoolSharesToRemove)
	candidate.Liabilities = candidate.Liabilities.Sub(shares)
	return p, candidate, createdCoins
}
