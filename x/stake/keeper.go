package stake

import (
	"bytes"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/bank"
	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
)

// keeper of the staking store
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        *wire.Codec
	coinKeeper bank.Keeper

	// caches
	pool   Pool
	params Params

	// codespace
	codespace sdk.CodespaceType
}

func NewKeeper(cdc *wire.Codec, key sdk.StoreKey, ck bank.Keeper, codespace sdk.CodespaceType) Keeper {
	keeper := Keeper{
		storeKey:   key,
		cdc:        cdc,
		coinKeeper: ck,
		codespace:  codespace,
	}
	return keeper
}

// get the current in-block validator operation counter
func (k Keeper) getIntraTxCounter(ctx sdk.Context) int16 {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(IntraTxCounterKey)
	if b == nil {
		return 0
	}
	var counter int16
	k.cdc.MustUnmarshalBinary(b, &counter)
	return counter
}

// set the current in-block validator operation counter
func (k Keeper) setIntraTxCounter(ctx sdk.Context, counter int16) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinary(counter)
	store.Set(IntraTxCounterKey, bz)
}

//_________________________________________________________________________

// get a single candidate
func (k Keeper) GetCandidate(ctx sdk.Context, addr sdk.Address) (candidate Candidate, found bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(GetCandidateKey(addr))
	if b == nil {
		return candidate, false
	}
	k.cdc.MustUnmarshalBinary(b, &candidate)
	return candidate, true
}

// Get the set of all candidates, retrieve a maxRetrieve number of records
func (k Keeper) GetCandidates(ctx sdk.Context, maxRetrieve int16) (candidates Candidates) {
	store := ctx.KVStore(k.storeKey)
	iterator := store.SubspaceIterator(CandidatesKey)

	candidates = make([]Candidate, maxRetrieve)
	i := 0
	for ; ; i++ {
		if !iterator.Valid() || i > int(maxRetrieve-1) {
			iterator.Close()
			break
		}
		bz := iterator.Value()
		var candidate Candidate
		k.cdc.MustUnmarshalBinary(bz, &candidate)
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

	// if found, copy the old block height and counter
	if oldFound {
		candidate.ValidatorBondHeight = oldCandidate.ValidatorBondHeight
		candidate.ValidatorBondCounter = oldCandidate.ValidatorBondCounter
	}

	// marshal the candidate record and add to the state
	bz := k.cdc.MustMarshalBinary(candidate)
	store.Set(GetCandidateKey(address), bz)

	if oldFound {
		// if the voting power is the same no need to update any of the other indexes
		if oldCandidate.BondedShares.Equal(candidate.BondedShares) {
			return
		}

		// if this candidate wasn't just bonded then update the height and counter
		if oldCandidate.Status != Bonded {
			candidate.ValidatorBondHeight = ctx.BlockHeight()
			counter := k.getIntraTxCounter(ctx)
			candidate.ValidatorBondCounter = counter
			k.setIntraTxCounter(ctx, counter+1)
		}

		// delete the old record in the power ordered list
		store.Delete(GetValidatorKey(oldCandidate.validator()))
	}

	// set the new candidate record
	bz = k.cdc.MustMarshalBinary(candidate)
	store.Set(GetCandidateKey(address), bz)

	// update the list ordered by voting power
	validator := candidate.validator()
	bzVal := k.cdc.MustMarshalBinary(validator)
	store.Set(GetValidatorKey(validator), bzVal)

	// add to the validators to update list if is already a validator
	if store.Get(GetRecentValidatorKey(candidate.PubKey)) != nil {
		bzAbci := k.cdc.MustMarshalBinary(validator.abciValidator(k.cdc))
		store.Set(GetAccUpdateValidatorKey(address), bzAbci)

		// also update the recent validator store
		store.Set(GetRecentValidatorKey(validator.PubKey), bzVal)
		return
	}

	// maybe add to the validator list and kick somebody off
	k.addNewValidatorOrNot(ctx, store, candidate.Address)
	return
}

func (k Keeper) removeCandidate(ctx sdk.Context, address sdk.Address) {

	// first retreive the old candidate record
	candidate, found := k.GetCandidate(ctx, address)
	if !found {
		return
	}

	// delete the old candidate record
	store := ctx.KVStore(k.storeKey)
	store.Delete(GetCandidateKey(address))
	store.Delete(GetValidatorKey(candidate.validator()))

	// delete from recent and power weighted validator groups if the validator
	// exists and add validator with zero power to the validator updates
	if store.Get(GetRecentValidatorKey(candidate.PubKey)) == nil {
		return
	}
	bz := k.cdc.MustMarshalBinary(candidate.validator().abciValidatorZero(k.cdc))
	store.Set(GetAccUpdateValidatorKey(address), bz)
	store.Delete(GetRecentValidatorKey(candidate.PubKey))
}

//___________________________________________________________________________

// get the group of the most recent validators
func (k Keeper) GetValidators(ctx sdk.Context) (validators []Validator) {
	store := ctx.KVStore(k.storeKey)

	// add the actual validator power sorted store
	maxValidators := k.GetParams(ctx).MaxValidators
	validators = make([]Validator, maxValidators)

	iterator := store.SubspaceIterator(RecentValidatorsKey)
	i := 0
	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var validator Validator
		k.cdc.MustUnmarshalBinary(bz, &validator)
		validators[i] = validator
		i++
	}
	iterator.Close()
	return validators[:i] // trim
}

// Only used for testing
// get the group of the most recent validators
func (k Keeper) getValidatorsOrdered(ctx sdk.Context) []Validator {
	vals := k.GetValidators(ctx)
	sort.Sort(sort.Reverse(validators(vals)))
	return vals
}

// Is the address provided a part of the most recently saved validator group?
func (k Keeper) IsValidator(ctx sdk.Context, pk crypto.PubKey) bool {
	store := ctx.KVStore(k.storeKey)
	if store.Get(GetRecentValidatorKey(pk)) == nil {
		return false
	}
	return true
}

// This function add's (or doesn't add) a candidate record to the validator group
// simultaniously it kicks any old validators out
//
// The correct subset is retrieved by iterating through an index of the
// candidates sorted by power, stored using the ValidatorsKey. Simultaniously
// the most recent the validator records are updated in store with the
// RecentValidatorsKey. This store is used to determine if a candidate is a
// validator without needing to iterate over the subspace as we do in
// GetValidators
func (k Keeper) addNewValidatorOrNot(ctx sdk.Context, store sdk.KVStore, address sdk.Address) {

	// clear the recent validators store, add to the ToKickOut temp store
	iterator := store.SubspaceIterator(RecentValidatorsKey)
	for ; iterator.Valid(); iterator.Next() {

		bz := iterator.Value()
		var validator Validator
		k.cdc.MustUnmarshalBinary(bz, &validator)

		addr := validator.Address

		// iterator.Value is the validator object
		store.Set(GetToKickOutValidatorKey(addr), iterator.Value())
		store.Delete(iterator.Key())
	}
	iterator.Close()

	// add the actual validator power sorted store
	maxValidators := k.GetParams(ctx).MaxValidators
	iterator = store.ReverseSubspaceIterator(ValidatorsKey) // largest to smallest
	i := 0
	for ; ; i++ {
		if !iterator.Valid() || i > int(maxValidators-1) {
			iterator.Close()
			break
		}
		bz := iterator.Value()
		var validator Validator
		k.cdc.MustUnmarshalBinary(bz, &validator)

		// remove from ToKickOut group
		store.Delete(GetToKickOutValidatorKey(validator.Address))

		// also add to the recent validators group
		store.Set(GetRecentValidatorKey(validator.PubKey), bz)

		// MOST IMPORTANTLY, add to the accumulated changes if this is the modified candidate
		if bytes.Equal(address, validator.Address) {
			bz = k.cdc.MustMarshalBinary(validator.abciValidator(k.cdc))
			store.Set(GetAccUpdateValidatorKey(address), bz)
		}

		iterator.Next()
	}

	// add any kicked out validators to the acc change
	iterator = store.SubspaceIterator(ToKickOutValidatorsKey)
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		addr := AddrFromKey(key)

		// get the zero abci validator from the ToKickOut iterator value
		bz := iterator.Value()
		var validator Validator
		k.cdc.MustUnmarshalBinary(bz, &validator)
		bz = k.cdc.MustMarshalBinary(validator.abciValidatorZero(k.cdc))

		store.Set(GetAccUpdateValidatorKey(addr), bz)
		store.Delete(key)
	}
	iterator.Close()
}

// cummulative power of the non-absent prevotes
func (k Keeper) GetTotalPrecommitVotingPower(ctx sdk.Context) sdk.Rat {
	store := ctx.KVStore(k.storeKey)

	// get absent prevote indexes
	absents := ctx.AbsentValidators()

	TotalPower := sdk.ZeroRat()
	i := int32(0)
	iterator := store.SubspaceIterator(RecentValidatorsKey)
	for ; iterator.Valid(); iterator.Next() {

		skip := false
		for j, absentIndex := range absents {
			if absentIndex > i {
				break
			}

			// if non-voting validator found, skip adding its power
			if absentIndex == i {
				absents = append(absents[:j], absents[j+1:]...) // won't need again
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		bz := iterator.Value()
		var validator Validator
		k.cdc.MustUnmarshalBinary(bz, &validator)
		TotalPower = TotalPower.Add(validator.Power)
		i++
	}
	iterator.Close()
	return TotalPower
}

//_________________________________________________________________________
// Accumulated updates to the validator set

// get the most recently updated validators
func (k Keeper) getAccUpdateValidators(ctx sdk.Context) (updates []abci.Validator) {
	store := ctx.KVStore(k.storeKey)

	iterator := store.SubspaceIterator(AccUpdateValidatorsKey) //smallest to largest
	for ; iterator.Valid(); iterator.Next() {
		valBytes := iterator.Value()
		var val abci.Validator
		k.cdc.MustUnmarshalBinary(valBytes, &val)
		updates = append(updates, val)
	}
	iterator.Close()
	return
}

// remove all validator update entries
func (k Keeper) clearAccUpdateValidators(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)

	// delete subspace
	iterator := store.SubspaceIterator(AccUpdateValidatorsKey)
	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
	}
	iterator.Close()
}

//_____________________________________________________________________

// load a delegator bond
func (k Keeper) GetDelegatorBond(ctx sdk.Context,
	delegatorAddr, candidateAddr sdk.Address) (bond DelegatorBond, found bool) {

	store := ctx.KVStore(k.storeKey)
	delegatorBytes := store.Get(GetDelegatorBondKey(delegatorAddr, candidateAddr, k.cdc))
	if delegatorBytes == nil {
		return bond, false
	}

	k.cdc.MustUnmarshalBinary(delegatorBytes, &bond)
	return bond, true
}

// load all bonds
func (k Keeper) getBonds(ctx sdk.Context, maxRetrieve int16) (bonds []DelegatorBond) {
	store := ctx.KVStore(k.storeKey)
	iterator := store.SubspaceIterator(DelegatorBondKeyPrefix)

	bonds = make([]DelegatorBond, maxRetrieve)
	i := 0
	for ; ; i++ {
		if !iterator.Valid() || i > int(maxRetrieve-1) {
			iterator.Close()
			break
		}
		bondBytes := iterator.Value()
		var bond DelegatorBond
		k.cdc.MustUnmarshalBinary(bondBytes, &bond)
		bonds[i] = bond
		iterator.Next()
	}
	return bonds[:i] // trim
}

// load all bonds of a delegator
func (k Keeper) GetDelegatorBonds(ctx sdk.Context, delegator sdk.Address, maxRetrieve int16) (bonds []DelegatorBond) {
	store := ctx.KVStore(k.storeKey)
	delegatorPrefixKey := GetDelegatorBondsKey(delegator, k.cdc)
	iterator := store.SubspaceIterator(delegatorPrefixKey) //smallest to largest

	bonds = make([]DelegatorBond, maxRetrieve)
	i := 0
	for ; ; i++ {
		if !iterator.Valid() || i > int(maxRetrieve-1) {
			iterator.Close()
			break
		}
		bondBytes := iterator.Value()
		var bond DelegatorBond
		k.cdc.MustUnmarshalBinary(bondBytes, &bond)
		bonds[i] = bond
		iterator.Next()
	}
	return bonds[:i] // trim
}

func (k Keeper) setDelegatorBond(ctx sdk.Context, bond DelegatorBond) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinary(bond)
	store.Set(GetDelegatorBondKey(bond.DelegatorAddr, bond.CandidateAddr, k.cdc), b)
}

func (k Keeper) removeDelegatorBond(ctx sdk.Context, bond DelegatorBond) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(GetDelegatorBondKey(bond.DelegatorAddr, bond.CandidateAddr, k.cdc))
}

//_______________________________________________________________________

// XXX TODO trim functionality

// retrieve all the power changes which occur after a height
func (k Keeper) GetPowerChangesAfterHeight(ctx sdk.Context, earliestHeight int64) (pcs []PowerChange) {
	store := ctx.KVStore(k.storeKey)

	iterator := store.SubspaceIterator(PowerChangeKey) //smallest to largest
	for ; iterator.Valid(); iterator.Next() {
		pcBytes := iterator.Value()
		var pc PowerChange
		k.cdc.MustUnmarshalBinary(pcBytes, &pc)
		if pc.Height < earliestHeight {
			break
		}
		pcs = append(pcs, pc)
	}
	iterator.Close()
	return
}

// set a power change
func (k Keeper) setPowerChange(ctx sdk.Context, pc PowerChange) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinary(pc)
	store.Set(GetPowerChangeKey(pc.Height), b)
}

//_______________________________________________________________________

// load/save the global staking params
func (k Keeper) GetParams(ctx sdk.Context) (params Params) {
	// check if cached before anything
	if !k.params.equal(Params{}) {
		return k.params
	}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(ParamKey)
	if b == nil {
		panic("Stored params should not have been nil")
	}

	k.cdc.MustUnmarshalBinary(b, &params)
	return
}
func (k Keeper) setParams(ctx sdk.Context, params Params) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinary(params)
	store.Set(ParamKey, b)
	k.params = Params{} // clear the cache
}

//_______________________________________________________________________

// load/save the pool
func (k Keeper) GetPool(ctx sdk.Context) (pool Pool) {
	// check if cached before anything
	if !k.pool.equal(Pool{}) {
		return k.pool
	}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(PoolKey)
	if b == nil {
		panic("Stored pool should not have been nil")
	}
	k.cdc.MustUnmarshalBinary(b, &pool)
	return
}

func (k Keeper) setPool(ctx sdk.Context, p Pool) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinary(p)
	store.Set(PoolKey, b)
	k.pool = Pool{} //clear the cache
}

//__________________________________________________________________________

// Implements ValidatorSetKeeper

var _ sdk.ValidatorSetKeeper = Keeper{}

func (k Keeper) ValidatorSet(ctx sdk.Context) sdk.ValidatorSet {
	vals := k.GetValidators(ctx)
	return ValidatorSet(vals)
}

func (k Keeper) GetByAddress(ctx sdk.Context, addr sdk.Address) sdk.Validator {
	can, ok := k.GetCandidate(ctx, addr)
	if !ok {
		return nil
	}
	if can.Status != Bonded {
		return nil
	}
	return can.validator()
}

func (k Keeper) TotalPower(ctx sdk.Context) sdk.Rat {
	pool := k.GetPool(ctx)
	return pool.BondedShares
}
