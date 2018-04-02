package stake

import (
	"bytes"
	"encoding/json"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// keeper of the staking store
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        *wire.Codec
	coinKeeper bank.CoinKeeper

	// caches
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

// InitGenesis - store genesis parameters
func (k Keeper) InitGenesis(ctx sdk.Context, data json.RawMessage) error {
	var state GenesisState
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}
	k.setPool(ctx, state.Pool)
	k.setParams(ctx, state.Params)
	return nil
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

// Get the set of all candidates, retrieve a maxRetrieve number of records
func (k Keeper) GetCandidates(ctx sdk.Context, maxRetrieve int16) (candidates Candidates) {
	store := ctx.KVStore(k.storeKey)
	iterator := store.Iterator(subspace(CandidatesKey))

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

func (k Keeper) setCandidate(ctx sdk.Context, candidate Candidate) {
	store := ctx.KVStore(k.storeKey)
	address := candidate.Address

	// retreive the old candidate record
	oldCandidate, oldFound := k.GetCandidate(ctx, address)

	// marshal the candidate record and add to the state
	bz, err := k.cdc.MarshalBinary(candidate)
	if err != nil {
		panic(err)
	}
	store.Set(GetCandidateKey(candidate.Address), bz)

	// mashal the new validator record
	validator := Validator{address, candidate.Assets}
	bz, err = k.cdc.MarshalBinary(validator)
	if err != nil {
		panic(err)
	}

	// if the voting power is the same no need to update any of the other indexes
	if oldFound && oldCandidate.Assets.Equal(candidate.Assets) {
		return
	}

	// update the list ordered by voting power
	if oldFound {
		store.Delete(GetValidatorKey(address, oldCandidate.Assets, k.cdc))
	}
	store.Set(GetValidatorKey(address, validator.VotingPower, k.cdc), bz)

	// add to the validators to update list if is already a validator
	// or is a new validator
	setAcc := false
	if store.Get(GetRecentValidatorKey(address)) != nil {
		setAcc = true

		// want to check in the else statement because inefficient
	} else if k.isNewValidator(ctx, store, address) {
		setAcc = true
	}
	if setAcc {
		store.Set(GetAccUpdateValidatorKey(validator.Address), bz)
	}
	return
}

func (k Keeper) removeCandidate(ctx sdk.Context, address sdk.Address) {

	// first retreive the old candidate record
	oldCandidate, found := k.GetCandidate(ctx, address)
	if !found {
		return
	}

	// delete the old candidate record
	store := ctx.KVStore(k.storeKey)
	store.Delete(GetCandidateKey(address))
	store.Delete(GetValidatorKey(address, oldCandidate.Assets, k.cdc))

	// delete from recent and power weighted validator groups if the validator
	// exists and add validator with zero power to the validator updates
	if store.Get(GetRecentValidatorKey(address)) == nil {
		return
	}
	bz, err := k.cdc.MarshalBinary(Validator{address, sdk.ZeroRat})
	if err != nil {
		panic(err)
	}
	store.Set(GetAccUpdateValidatorKey(address), bz)
	store.Delete(GetRecentValidatorKey(address))
}

//___________________________________________________________________________

// get the most recent updated validator set from the Candidates. These bonds
// are already sorted by Assets from the UpdateVotingPower function which
// is the only function which is to modify the Assets
// this function also updaates the most recent validators saved in store
func (k Keeper) GetValidators(ctx sdk.Context) (validators []Validator) {
	store := ctx.KVStore(k.storeKey)

	// clear the recent validators store, add to the ToKickOut Temp store
	iterator := store.Iterator(subspace(RecentValidatorsKey))
	for ; iterator.Valid(); iterator.Next() {
		addr := AddrFromKey(iterator.Key())
		store.Set(GetToKickOutValidatorKey(addr), []byte{})
		store.Delete(iterator.Key())
	}
	iterator.Close()

	// add the actual validator power sorted store
	maxVal := k.GetParams(ctx).MaxValidators
	iterator = store.ReverseIterator(subspace(ValidatorsKey)) // largest to smallest
	validators = make([]Validator, maxVal)
	i := 0
	for ; ; i++ {
		if !iterator.Valid() || i > int(maxVal-1) {
			iterator.Close()
			break
		}
		bz := iterator.Value()
		var val Validator
		err := k.cdc.UnmarshalBinary(bz, &val)
		if err != nil {
			panic(err)
		}
		validators[i] = val

		// remove from ToKickOut group
		store.Delete(GetToKickOutValidatorKey(val.Address))

		// also add to the recent validators group
		store.Set(GetRecentValidatorKey(val.Address), bz) // XXX should store nothing

		iterator.Next()
	}

	// add any kicked out validators to the acc change
	iterator = store.Iterator(subspace(ToKickOutValidatorsKey))
	for ; iterator.Valid(); iterator.Next() {
		addr := AddrFromKey(iterator.Key())
		bz, err := k.cdc.MarshalBinary(Validator{addr, sdk.ZeroRat})
		if err != nil {
			panic(err)
		}
		store.Set(GetAccUpdateValidatorKey(addr), bz)
		store.Delete(iterator.Key())
	}
	iterator.Close()

	return validators[:i] // trim
}

// TODO this is madly inefficient because need to call every time we set a candidate
// Should use something better than an iterator maybe?
// Used to determine if something has just been added to the actual validator set
func (k Keeper) isNewValidator(ctx sdk.Context, store sdk.KVStore, address sdk.Address) bool {
	// add the actual validator power sorted store
	maxVal := k.GetParams(ctx).MaxValidators
	iterator := store.ReverseIterator(subspace(ValidatorsKey)) // largest to smallest
	for i := 0; ; i++ {
		if !iterator.Valid() || i > int(maxVal-1) {
			iterator.Close()
			break
		}
		bz := iterator.Value()
		var val Validator
		err := k.cdc.UnmarshalBinary(bz, &val)
		if err != nil {
			panic(err)
		}
		if bytes.Equal(val.Address, address) {
			return true
		}
		iterator.Next()
	}

	return false
}

// Is the address provided a part of the most recently saved validator group?
func (k Keeper) IsRecentValidator(ctx sdk.Context, address sdk.Address) bool {
	store := ctx.KVStore(k.storeKey)
	if store.Get(GetRecentValidatorKey(address)) == nil {
		return false
	}
	return true
}

//_________________________________________________________________________
// Accumulated updates to the validator set

// get the most recently updated validators
func (k Keeper) getAccUpdateValidators(ctx sdk.Context) (updates []Validator) {
	store := ctx.KVStore(k.storeKey)

	iterator := store.Iterator(subspace(AccUpdateValidatorsKey)) //smallest to largest
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
func (k Keeper) clearAccUpdateValidators(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	k.deleteSubSpace(store, AccUpdateValidatorsKey)
}

// TODO move to common functionality somewhere
func (k Keeper) deleteSubSpace(store sdk.KVStore, key []byte) {
	iterator := store.Iterator(subspace(key))
	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
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
		panic(errors.New("Stored params should not have been nil"))
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

//_______________________________________________________________________

// load/save the pool
func (k Keeper) GetPool(ctx sdk.Context) (gs Pool) {
	// check if cached before anything
	if k.gs != (Pool{}) {
		return k.gs
	}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(PoolKey)
	if b == nil {
		panic(errors.New("Stored pool should not have been nil"))
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
