package stake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

//nolint
var (
	// Keys for store prefixes
	CandidatesAddrKey = []byte{0x01} // key for all candidates' addresses
	ParamKey          = []byte{0x02} // key for global parameters relating to staking
	GlobalStateKey    = []byte{0x03} // key for global parameters relating to staking

	// Key prefixes
	CandidateKeyPrefix        = []byte{0x04} // prefix for each key to a candidate
	ValidatorKeyPrefix        = []byte{0x05} // prefix for each key to a candidate
	ValidatorUpdatesKeyPrefix = []byte{0x06} // prefix for each key to a candidate
	DelegatorBondKeyPrefix    = []byte{0x07} // prefix for each key to a delegator's bond
	DelegatorBondsKeyPrefix   = []byte{0x08} // prefix for each key to a delegator's bond
)

// XXX remove beggining word get from all these keys
// GetCandidateKey - get the key for the candidate with address
func GetCandidateKey(addr sdk.Address) []byte {
	return append(CandidateKeyPrefix, addr.Bytes()...)
}

// GetValidatorKey - get the key for the validator used in the power-store
func GetValidatorKey(addr sdk.Address, power sdk.Rat, cdc *wire.Codec) []byte {
	b, _ := cdc.MarshalJSON(power)                                   // TODO need to handle error here?
	return append(ValidatorKeyPrefix, append(b, addr.Bytes()...)...) // TODO does this need prefix if its in its own store
}

// GetValidatorUpdatesKey - get the key for the validator used in the power-store
func GetValidatorUpdatesKey(addr sdk.Address) []byte {
	return append(ValidatorUpdatesKeyPrefix, addr.Bytes()...) // TODO does this need prefix if its in its own store
}

// GetDelegatorBondKey - get the key for delegator bond with candidate
func GetDelegatorBondKey(delegatorAddr, candidateAddr sdk.Address, cdc *wire.Codec) []byte {
	return append(GetDelegatorBondKeyPrefix(delegatorAddr, cdc), candidateAddr.Bytes()...)
}

// GetDelegatorBondKeyPrefix - get the prefix for a delegator for all candidates
func GetDelegatorBondKeyPrefix(delegatorAddr sdk.Address, cdc *wire.Codec) []byte {
	res, err := cdc.MarshalJSON(&delegatorAddr)
	if err != nil {
		panic(err)
	}
	return append(DelegatorBondKeyPrefix, res...)
}

// GetDelegatorBondsKey - get the key for list of all the delegator's bonds
func GetDelegatorBondsKey(delegatorAddr sdk.Address, cdc *wire.Codec) []byte {
	res, err := cdc.MarshalJSON(&delegatorAddr)
	if err != nil {
		panic(err)
	}
	return append(DelegatorBondsKeyPrefix, res...)
}

//___________________________________________________________________________

// keeper of the staking store
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        *wire.Codec
	coinKeeper bank.CoinKeeper

	//just caches
	gs     GlobalState
	params Params
}

func NewKeeper(ctx sdk.Context, cdc *wire.Codec, key sdk.StoreKey, ck bank.CoinKeeper) Keeper {
	keeper := Keeper{
		storeKey:   key,
		cdc:        cdc,
		coinKeeper: ck,
	}
	return keeper
}

//XXX load/save -> get/set
func (k Keeper) getCandidate(ctx sdk.Context, addr sdk.Address) (candidate Candidate, found bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(GetCandidateKey(addr))
	if b == nil {
		return candidate, false
	}
	err := k.cdc.UnmarshalJSON(b, &candidate)
	if err != nil {
		panic(err)
	}
	return candidate, true
}

func (k Keeper) setCandidate(ctx sdk.Context, candidate Candidate) {
	store := ctx.KVStore(k.storeKey)

	// XXX should only remove validator if we know candidate is a validator
	k.removeValidator(ctx, candidate.Address)
	validator := Validator{candidate.Address, candidate.VotingPower}
	k.updateValidator(ctx, validator)

	b, err := k.cdc.MarshalJSON(candidate)
	if err != nil {
		panic(err)
	}
	store.Set(GetCandidateKey(candidate.Address), b)
}

func (k Keeper) removeCandidate(ctx sdk.Context, candidateAddr sdk.Address) {
	store := ctx.KVStore(k.storeKey)

	// XXX should only remove validator if we know candidate is a validator
	k.removeValidator(ctx, candidateAddr)
	store.Delete(GetCandidateKey(candidateAddr))
}

//___________________________________________________________________________

//func loadValidator(store sdk.KVStore, address sdk.Address, votingPower sdk.Rat) *Validator {
//b := store.Get(GetValidatorKey(address, votingPower))
//if b == nil {
//return nil
//}
//validator := new(Validator)
//err := cdc.UnmarshalJSON(b, validator)
//if err != nil {
//panic(err) // This error should never occur big problem if does
//}
//return validator
//}

// updateValidator - update a validator and create accumulate any changes
// in the changed validator substore
func (k Keeper) updateValidator(ctx sdk.Context, validator Validator) {
	store := ctx.KVStore(k.storeKey)

	b, err := k.cdc.MarshalJSON(validator)
	if err != nil {
		panic(err)
	}

	// add to the validators to update list if necessary
	store.Set(GetValidatorUpdatesKey(validator.Address), b)

	// update the list ordered by voting power
	store.Set(GetValidatorKey(validator.Address, validator.VotingPower, k.cdc), b)
}

func (k Keeper) removeValidator(ctx sdk.Context, address sdk.Address) {
	store := ctx.KVStore(k.storeKey)

	//add validator with zero power to the validator updates
	b, err := k.cdc.MarshalJSON(Validator{address, sdk.ZeroRat})
	if err != nil {
		panic(err)
	}
	store.Set(GetValidatorUpdatesKey(address), b)

	// now actually delete from the validator set
	candidate, found := k.getCandidate(ctx, address)
	if found {
		store.Delete(GetValidatorKey(address, candidate.VotingPower, k.cdc))
	}
}

// get the most recent updated validator set from the Candidates. These bonds
// are already sorted by VotingPower from the UpdateVotingPower function which
// is the only function which is to modify the VotingPower
func (k Keeper) getValidators(ctx sdk.Context, maxVal uint16) (validators []Validator) {
	store := ctx.KVStore(k.storeKey)

	iterator := store.Iterator(subspace(ValidatorKeyPrefix)) //smallest to largest

	validators = make([]Validator, maxVal)
	for i := 0; ; i++ {
		if !iterator.Valid() || i > int(maxVal) {
			iterator.Close()
			break
		}
		valBytes := iterator.Value()
		var val Validator
		err := k.cdc.UnmarshalJSON(valBytes, &val)
		if err != nil {
			panic(err)
		}
		validators[i] = val
		iterator.Next()
	}

	return
}

//_________________________________________________________________________

// get the most updated validators
func (k Keeper) getValidatorUpdates(ctx sdk.Context) (updates []Validator) {
	store := ctx.KVStore(k.storeKey)

	iterator := store.Iterator(subspace(ValidatorUpdatesKeyPrefix)) //smallest to largest

	for ; iterator.Valid(); iterator.Next() {
		valBytes := iterator.Value()
		var val Validator
		err := k.cdc.UnmarshalJSON(valBytes, &val)
		if err != nil {
			panic(err)
		}
		updates = append(updates, val)
	}

	iterator.Close()
	return
}

// remove all validator update entries
func (k Keeper) clearValidatorUpdates(ctx sdk.Context, maxVal int) {
	store := ctx.KVStore(k.storeKey)
	iterator := store.Iterator(subspace(ValidatorUpdatesKeyPrefix))
	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key()) // XXX write test for this, may need to be in a second loop
	}
	iterator.Close()
}

//---------------------------------------------------------------------

// getCandidates - get the active list of all candidates
func (k Keeper) getCandidates(ctx sdk.Context) (candidates Candidates) {
	store := ctx.KVStore(k.storeKey)
	iterator := store.Iterator(subspace(CandidateKeyPrefix))

	for ; iterator.Valid(); iterator.Next() {
		candidateBytes := iterator.Value()
		var candidate Candidate
		err := k.cdc.UnmarshalJSON(candidateBytes, &candidate)
		if err != nil {
			panic(err)
		}
		candidates = append(candidates, candidate)
	}
	iterator.Close()
	return candidates
}

//_____________________________________________________________________

// XXX use a store iterator here instead
//// load the pubkeys of all candidates a delegator is delegated too
//func (k Keeper) getDelegatorCandidates(ctx sdk.Context, delegator sdk.Address) (candidateAddrs []sdk.Address) {
//store := ctx.KVStore(k.storeKey)

//candidateBytes := store.Get(GetDelegatorBondsKey(delegator, k.cdc))
//if candidateBytes == nil {
//return nil
//}

//err := k.cdc.UnmarshalJSON(candidateBytes, &candidateAddrs)
//if err != nil {
//panic(err)
//}
//return
//}

//_____________________________________________________________________

func (k Keeper) getDelegatorBond(ctx sdk.Context,
	delegatorAddr, candidateAddr sdk.Address) (bond DelegatorBond, found bool) {

	store := ctx.KVStore(k.storeKey)
	delegatorBytes := store.Get(GetDelegatorBondKey(delegatorAddr, candidateAddr, k.cdc))
	if delegatorBytes == nil {
		return bond, false
	}

	err := k.cdc.UnmarshalJSON(delegatorBytes, &bond)
	if err != nil {
		panic(err)
	}
	return bond, true
}

func (k Keeper) setDelegatorBond(ctx sdk.Context, bond DelegatorBond) {
	store := ctx.KVStore(k.storeKey)

	// XXX use store iterator
	// if a new bond add to the list of bonds
	//if k.getDelegatorBond(delegator, bond.Address) == nil {
	//pks := k.getDelegatorCandidates(delegator)
	//pks = append(pks, bond.Address)
	//b, err := k.cdc.MarshalJSON(pks)
	//if err != nil {
	//panic(err)
	//}
	//store.Set(GetDelegatorBondsKey(delegator, k.cdc), b)
	//}

	// now actually save the bond
	b, err := k.cdc.MarshalJSON(bond)
	if err != nil {
		panic(err)
	}
	store.Set(GetDelegatorBondKey(bond.DelegatorAddr, bond.CandidateAddr, k.cdc), b)
}

func (k Keeper) removeDelegatorBond(ctx sdk.Context, bond DelegatorBond) {
	store := ctx.KVStore(k.storeKey)

	// XXX use store iterator
	// TODO use list queries on multistore to remove iterations here!
	// first remove from the list of bonds
	//addrs := k.getDelegatorCandidates(delegator)
	//for i, addr := range addrs {
	//if bytes.Equal(candidateAddr, addr) {
	//addrs = append(addrs[:i], addrs[i+1:]...)
	//}
	//}
	//b, err := k.cdc.MarshalJSON(addrs)
	//if err != nil {
	//panic(err)
	//}
	//store.Set(GetDelegatorBondsKey(delegator, k.cdc), b)

	// now remove the actual bond
	store.Delete(GetDelegatorBondKey(bond.DelegatorAddr, bond.CandidateAddr, k.cdc))
	//updateDelegatorBonds(store, delegator) //XXX remove?
}

//_______________________________________________________________________

// load/save the global staking params
func (k Keeper) getParams(ctx sdk.Context) (params Params) {
	// check if cached before anything
	if k.params != (Params{}) {
		return k.params
	}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(ParamKey)
	if b == nil {
		k.params = defaultParams()
		return k.params
	}

	err := k.cdc.UnmarshalJSON(b, &params)
	if err != nil {
		panic(err) // This error should never occur big problem if does
	}
	return
}
func (k Keeper) setParams(ctx sdk.Context, params Params) {
	store := ctx.KVStore(k.storeKey)
	b, err := k.cdc.MarshalJSON(params)
	if err != nil {
		panic(err)
	}
	store.Set(ParamKey, b)
	k.params = Params{} // clear the cache
}

//_______________________________________________________________________

// XXX nothing is this Keeper should return a pointer...!!!!!!
// load/save the global staking state
func (k Keeper) getGlobalState(ctx sdk.Context) (gs GlobalState) {
	// check if cached before anything
	if k.gs != (GlobalState{}) {
		return k.gs
	}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(GlobalStateKey)
	if b == nil {
		return initialGlobalState()
	}
	err := k.cdc.UnmarshalJSON(b, &gs)
	if err != nil {
		panic(err) // This error should never occur big problem if does
	}
	return
}

func (k Keeper) setGlobalState(ctx sdk.Context, gs GlobalState) {
	store := ctx.KVStore(k.storeKey)
	b, err := k.cdc.MarshalJSON(gs)
	if err != nil {
		panic(err)
	}
	store.Set(GlobalStateKey, b)
	k.gs = GlobalState{} // clear the cache
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

// XXX expand to include the function of actually transfering the tokens

//XXX CONFIRM that use of the exRate is correct with Zarko Spec!
func (k Keeper) addTokensBonded(ctx sdk.Context, amount int64) (issuedShares sdk.Rat) {
	gs := k.getGlobalState(ctx)
	issuedShares = gs.bondedShareExRate().Inv().Mul(sdk.NewRat(amount)) // (tokens/shares)^-1 * tokens
	gs.BondedPool += amount
	gs.BondedShares = gs.BondedShares.Add(issuedShares)
	k.setGlobalState(ctx, gs)
	return
}

//XXX CONFIRM that use of the exRate is correct with Zarko Spec!
func (k Keeper) removeSharesBonded(ctx sdk.Context, shares sdk.Rat) (removedTokens int64) {
	gs := k.getGlobalState(ctx)
	removedTokens = gs.bondedShareExRate().Mul(shares).Evaluate() // (tokens/shares) * shares
	gs.BondedShares = gs.BondedShares.Sub(shares)
	gs.BondedPool -= removedTokens
	k.setGlobalState(ctx, gs)
	return
}

//XXX CONFIRM that use of the exRate is correct with Zarko Spec!
func (k Keeper) addTokensUnbonded(ctx sdk.Context, amount int64) (issuedShares sdk.Rat) {
	gs := k.getGlobalState(ctx)
	issuedShares = gs.unbondedShareExRate().Inv().Mul(sdk.NewRat(amount)) // (tokens/shares)^-1 * tokens
	gs.UnbondedShares = gs.UnbondedShares.Add(issuedShares)
	gs.UnbondedPool += amount
	k.setGlobalState(ctx, gs)
	return
}

//XXX CONFIRM that use of the exRate is correct with Zarko Spec!
func (k Keeper) removeSharesUnbonded(ctx sdk.Context, shares sdk.Rat) (removedTokens int64) {
	gs := k.getGlobalState(ctx)
	removedTokens = gs.unbondedShareExRate().Mul(shares).Evaluate() // (tokens/shares) * shares
	gs.UnbondedShares = gs.UnbondedShares.Sub(shares)
	gs.UnbondedPool -= removedTokens
	k.setGlobalState(ctx, gs)
	return
}

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
