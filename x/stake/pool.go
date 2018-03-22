package stake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

//TODO make these next two functions more efficient should be reading and writting to state ye know

// move a candidates asset pool from bonded to unbonded pool
func (k Keeper) bondedToUnbondedPool(ctx sdk.Context, candidate Candidate) {

	// replace bonded shares with unbonded shares
	tokens := k.removeSharesBonded(ctx, candidate.Assets)
	candidate.Assets = k.addTokensUnbonded(ctx, tokens)
	candidate.Status = Unbonded
	k.setCandidate(ctx, candidate)
}

// move a candidates asset pool from unbonded to bonded pool
func (k Keeper) unbondedToBondedPool(ctx sdk.Context, candidate Candidate) {

	// replace unbonded shares with bonded shares
	tokens := k.removeSharesUnbonded(ctx, candidate.Assets)
	candidate.Assets = k.addTokensBonded(ctx, tokens)
	candidate.Status = Bonded
	k.setCandidate(ctx, candidate)
}

//_______________________________________________________________________

func (k Keeper) addTokensBonded(ctx sdk.Context, amount int64) (issuedShares sdk.Rat) {
	gs := k.getGlobalState(ctx)
	issuedShares = gs.bondedShareExRate().Inv().Mul(sdk.NewRat(amount)) // (tokens/shares)^-1 * tokens
	gs.BondedPool += amount
	gs.BondedShares = gs.BondedShares.Add(issuedShares)
	k.setGlobalState(ctx, gs)
	return
}

func (k Keeper) removeSharesBonded(ctx sdk.Context, shares sdk.Rat) (removedTokens int64) {
	gs := k.getGlobalState(ctx)
	removedTokens = gs.bondedShareExRate().Mul(shares).Evaluate() // (tokens/shares) * shares
	gs.BondedShares = gs.BondedShares.Sub(shares)
	gs.BondedPool -= removedTokens
	k.setGlobalState(ctx, gs)
	return
}

func (k Keeper) addTokensUnbonded(ctx sdk.Context, amount int64) (issuedShares sdk.Rat) {
	gs := k.getGlobalState(ctx)
	issuedShares = gs.unbondedShareExRate().Inv().Mul(sdk.NewRat(amount)) // (tokens/shares)^-1 * tokens
	gs.UnbondedShares = gs.UnbondedShares.Add(issuedShares)
	gs.UnbondedPool += amount
	k.setGlobalState(ctx, gs)
	return
}

func (k Keeper) removeSharesUnbonded(ctx sdk.Context, shares sdk.Rat) (removedTokens int64) {
	gs := k.getGlobalState(ctx)
	removedTokens = gs.unbondedShareExRate().Mul(shares).Evaluate() // (tokens/shares) * shares
	gs.UnbondedShares = gs.UnbondedShares.Sub(shares)
	gs.UnbondedPool -= removedTokens
	k.setGlobalState(ctx, gs)
	return
}

//_______________________________________________________________________

// add tokens to a candidate
func (k Keeper) candidateAddTokens(ctx sdk.Context, candidate Candidate, amount int64) (issuedDelegatorShares sdk.Rat) {

	gs := k.getGlobalState(ctx)
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
	k.setGlobalState(ctx, gs) // TODO cache GlobalState?
	return
}

// remove shares from a candidate
func (k Keeper) candidateRemoveShares(ctx sdk.Context, candidate Candidate, shares sdk.Rat) (createdCoins int64) {

	gs := k.getGlobalState(ctx)
	//exRate := candidate.delegatorShareExRate() //XXX make sure not used

	globalPoolSharesToRemove := candidate.delegatorShareExRate().Mul(shares)
	if candidate.Status == Bonded {
		createdCoins = k.removeSharesBonded(ctx, globalPoolSharesToRemove)
	} else {
		createdCoins = k.removeSharesUnbonded(ctx, globalPoolSharesToRemove)
	}
	candidate.Assets = candidate.Assets.Sub(globalPoolSharesToRemove)
	candidate.Liabilities = candidate.Liabilities.Sub(shares)
	k.setGlobalState(ctx, gs) // TODO cache GlobalState?
	return
}
