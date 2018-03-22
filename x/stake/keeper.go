package stake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// keeper of the staking store
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        *wire.Codec
	coinKeeper bank.CoinKeeper

	//just caches
	gs     Pool
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

//_________________________________________________________________________

// get a single candidate
func (k Keeper) GetCandidate(ctx sdk.Context, addr sdk.Address) (candidate Candidate, found bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(GetCandidateKey(addr))
	if b == nil {
		return candidate, false
	}
	err := k.cdc.UnmarshalBinary(b, &candidate)
	if err != nil {
		panic(err)
	}
	return candidate, true
}

func (k Keeper) setCandidate(ctx sdk.Context, candidate Candidate) {
	store := ctx.KVStore(k.storeKey)

	k.removeValidator(ctx, candidate.Address)
	validator := Validator{candidate.Address, candidate.VotingPower}
	k.updateValidator(ctx, validator)

	b, err := k.cdc.MarshalBinary(candidate)
	if err != nil {
		panic(err)
	}
	store.Set(GetCandidateKey(candidate.Address), b)
}

func (k Keeper) removeCandidate(ctx sdk.Context, candidateAddr sdk.Address) {
	store := ctx.KVStore(k.storeKey)
	k.removeValidator(ctx, candidateAddr)
	store.Delete(GetCandidateKey(candidateAddr))
}

// Get the set of all candidates, retrieve a maxRetrieve number of records
func (k Keeper) GetCandidates(ctx sdk.Context, maxRetrieve int16) (candidates Candidates) {
	store := ctx.KVStore(k.storeKey)
	iterator := store.Iterator(subspace(CandidateKeyPrefix))

	candidates = make([]Candidate, maxRetrieve)
	i := 0
	for ; ; i++ {
		if !iterator.Valid() || i > int(maxRetrieve-1) {
			iterator.Close()
			break
		}
		bz := iterator.Value()
		var candidate Candidate
		err := k.cdc.UnmarshalBinary(bz, &candidate)
		if err != nil {
			panic(err)
		}
		candidates[i] = candidate
		iterator.Next()
	}
	return candidates[:i] // trim
}

//___________________________________________________________________________

// updateValidator - update a validator and create accumulate any changes
// in the changed validator substore
func (k Keeper) updateValidator(ctx sdk.Context, validator Validator) {
	store := ctx.KVStore(k.storeKey)

	b, err := k.cdc.MarshalBinary(validator)
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

	// XXX ensure that this record is a validator even?

	//add validator with zero power to the validator updates
	b, err := k.cdc.MarshalBinary(Validator{address, sdk.ZeroRat})
	if err != nil {
		panic(err)
	}
	store.Set(GetValidatorUpdatesKey(address), b)

	// now actually delete from the validator set
	candidate, found := k.GetCandidate(ctx, address)
	if found {
		store.Delete(GetValidatorKey(address, candidate.VotingPower, k.cdc))
	}
}

// get the most recent updated validator set from the Candidates. These bonds
// are already sorted by VotingPower from the UpdateVotingPower function which
// is the only function which is to modify the VotingPower
func (k Keeper) GetValidators(ctx sdk.Context) (validators []Validator) {
	store := ctx.KVStore(k.storeKey)
	maxVal := k.GetParams(ctx).MaxValidators

	iterator := store.Iterator(subspace(ValidatorKeyPrefix)) //smallest to largest

	validators = make([]Validator, maxVal)
	i := 0
	for ; ; i++ {
		if !iterator.Valid() || i > int(maxVal-1) {
			iterator.Close()
			break
		}
		valBytes := iterator.Value()
		var val Validator
		err := k.cdc.UnmarshalBinary(valBytes, &val)
		if err != nil {
			panic(err)
		}
		validators[i] = val
		iterator.Next()
	}
	return validators[:i] // trim
}

//_________________________________________________________________________

// get the most updated validators
func (k Keeper) getValidatorUpdates(ctx sdk.Context) (updates []Validator) {
	store := ctx.KVStore(k.storeKey)

	iterator := store.Iterator(subspace(ValidatorUpdatesKeyPrefix)) //smallest to largest

	for ; iterator.Valid(); iterator.Next() {
		valBytes := iterator.Value()
		var val Validator
		err := k.cdc.UnmarshalBinary(valBytes, &val)
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

//_____________________________________________________________________

func (k Keeper) getDelegatorBond(ctx sdk.Context,
	delegatorAddr, candidateAddr sdk.Address) (bond DelegatorBond, found bool) {

	store := ctx.KVStore(k.storeKey)
	delegatorBytes := store.Get(GetDelegatorBondKey(delegatorAddr, candidateAddr, k.cdc))
	if delegatorBytes == nil {
		return bond, false
	}

	err := k.cdc.UnmarshalBinary(delegatorBytes, &bond)
	if err != nil {
		panic(err)
	}
	return bond, true
}

func (k Keeper) setDelegatorBond(ctx sdk.Context, bond DelegatorBond) {
	store := ctx.KVStore(k.storeKey)
	b, err := k.cdc.MarshalBinary(bond)
	if err != nil {
		panic(err)
	}
	store.Set(GetDelegatorBondKey(bond.DelegatorAddr, bond.CandidateAddr, k.cdc), b)
}

func (k Keeper) removeDelegatorBond(ctx sdk.Context, bond DelegatorBond) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(GetDelegatorBondKey(bond.DelegatorAddr, bond.CandidateAddr, k.cdc))
}

// load all bonds of a delegator
func (k Keeper) getDelegatorBonds(ctx sdk.Context, delegator sdk.Address, maxRetrieve int16) (bonds []DelegatorBond) {
	store := ctx.KVStore(k.storeKey)
	delegatorPrefixKey := GetDelegatorBondsKey(delegator, k.cdc)
	iterator := store.Iterator(subspace(delegatorPrefixKey)) //smallest to largest

	bonds = make([]DelegatorBond, maxRetrieve)
	i := 0
	for ; ; i++ {
		if !iterator.Valid() || i > int(maxRetrieve-1) {
			iterator.Close()
			break
		}
		bondBytes := iterator.Value()
		var bond DelegatorBond
		err := k.cdc.UnmarshalBinary(bondBytes, &bond)
		if err != nil {
			panic(err)
		}
		bonds[i] = bond
		iterator.Next()
	}
	return bonds[:i] // trim
}

//_______________________________________________________________________

// load/save the global staking params
func (k Keeper) GetParams(ctx sdk.Context) (params Params) {
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

	err := k.cdc.UnmarshalBinary(b, &params)
	if err != nil {
		panic(err)
	}
	return
}
func (k Keeper) setParams(ctx sdk.Context, params Params) {
	store := ctx.KVStore(k.storeKey)
	b, err := k.cdc.MarshalBinary(params)
	if err != nil {
		panic(err)
	}
	store.Set(ParamKey, b)
	k.params = Params{} // clear the cache
}
