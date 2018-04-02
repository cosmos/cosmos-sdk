package stake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// load/save the global staking state
func (k Keeper) GetPool(ctx sdk.Context) (gs Pool) {
	// check if cached before anything
	if k.gs != (Pool{}) {
		return k.gs
	}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(PoolKey)
	if b == nil {
		return initialPool()
	}
	err := k.cdc.UnmarshalBinary(b, &gs)
	if err != nil {
		panic(err) // This error should never occur big problem if does
	}
	return
}

func (k Keeper) setPool(ctx sdk.Context, p Pool) {
	store := ctx.KVStore(k.storeKey)
	b, err := k.cdc.MarshalBinary(p)
	if err != nil {
		panic(err)
	}
	store.Set(PoolKey, b)
	k.gs = Pool{} // clear the cache
}

//_______________________________________________________________________

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
	p := k.GetPool(ctx)
	issuedShares = p.bondedShareExRate().Inv().Mul(sdk.NewRat(amount)) // (tokens/shares)^-1 * tokens
	p.BondedPool += amount
	p.BondedShares = p.BondedShares.Add(issuedShares)
	k.setPool(ctx, p)
	return
}

func (k Keeper) removeSharesBonded(ctx sdk.Context, shares sdk.Rat) (removedTokens int64) {
	p := k.GetPool(ctx)
	removedTokens = p.bondedShareExRate().Mul(shares).Evaluate() // (tokens/shares) * shares
	p.BondedShares = p.BondedShares.Sub(shares)
	p.BondedPool -= removedTokens
	k.setPool(ctx, p)
	return
}

func (k Keeper) addTokensUnbonded(ctx sdk.Context, amount int64) (issuedShares sdk.Rat) {
	p := k.GetPool(ctx)
	issuedShares = p.unbondedShareExRate().Inv().Mul(sdk.NewRat(amount)) // (tokens/shares)^-1 * tokens
	p.UnbondedShares = p.UnbondedShares.Add(issuedShares)
	p.UnbondedPool += amount
	k.setPool(ctx, p)
	return
}

func (k Keeper) removeSharesUnbonded(ctx sdk.Context, shares sdk.Rat) (removedTokens int64) {
	p := k.GetPool(ctx)
	removedTokens = p.unbondedShareExRate().Mul(shares).Evaluate() // (tokens/shares) * shares
	p.UnbondedShares = p.UnbondedShares.Sub(shares)
	p.UnbondedPool -= removedTokens
	k.setPool(ctx, p)
	return
}

//_______________________________________________________________________

// add tokens to a candidate
func (k Keeper) candidateAddTokens(ctx sdk.Context, candidate Candidate, amount int64) (issuedDelegatorShares sdk.Rat) {

	p := k.GetPool(ctx)
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
	k.setPool(ctx, p) // TODO cache Pool?
	return
}

// remove shares from a candidate
func (k Keeper) candidateRemoveShares(ctx sdk.Context, candidate Candidate, shares sdk.Rat) (createdCoins int64) {

	p := k.GetPool(ctx)
	//exRate := candidate.delegatorShareExRate() //XXX make sure not used

	globalPoolSharesToRemove := candidate.delegatorShareExRate().Mul(shares)
	if candidate.Status == Bonded {
		createdCoins = k.removeSharesBonded(ctx, globalPoolSharesToRemove)
	} else {
		createdCoins = k.removeSharesUnbonded(ctx, globalPoolSharesToRemove)
	}
	candidate.Assets = candidate.Assets.Sub(globalPoolSharesToRemove)
	candidate.Liabilities = candidate.Liabilities.Sub(shares)
	k.setPool(ctx, p) // TODO cache Pool?
	return
}
