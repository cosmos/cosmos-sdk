package stake

import (
	"bytes"

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
func (k Keeper) getCounter(ctx sdk.Context) int16 {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(CounterKey)
	if b == nil {
		return 0
	}
	var counter int16
	k.cdc.MustUnmarshalBinary(b, &counter)
	return counter
}

// set the current in-block validator operation counter
func (k Keeper) setCounter(ctx sdk.Context, counter int16) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinary(counter)
	store.Set(CounterKey, bz)
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

	// if the voting power is the same no need to update any of the other indexes
	if oldFound && oldCandidate.Assets.Equal(candidate.Assets) {
		return
	}

	updateHeight := false

	// update the list ordered by voting power
	if oldFound {
		if !k.isNewValidator(ctx, store, address) {
			updateHeight = true
		}
		// else already in the validator set - retain the old validator height and counter
		store.Delete(GetValidatorKey(address, oldCandidate.Assets, oldCandidate.ValidatorBondHeight, oldCandidate.ValidatorBondCounter, k.cdc))
	} else {
		updateHeight = true
	}

	if updateHeight {
		// wasn't a candidate or wasn't in the validator set, update the validator block height and counter
		candidate.ValidatorBondHeight = ctx.BlockHeight()
		counter := k.getCounter(ctx)
		candidate.ValidatorBondCounter = counter
		k.setCounter(ctx, counter+1)
	}

	// update the candidate record
	bz = k.cdc.MustMarshalBinary(candidate)
	store.Set(GetCandidateKey(address), bz)

	// marshal the new validator record
	validator := candidate.validator()
	bz = k.cdc.MustMarshalBinary(validator)
	store.Set(GetValidatorKey(address, validator.Power, validator.Height, validator.Counter, k.cdc), bz)

	// add to the validators to update list if is already a validator
	// or is a new validator
	setAcc := false
	if store.Get(GetRecentValidatorKey(candidate.PubKey)) != nil {
		setAcc = true

		// want to check in the else statement because inefficient
	} else if k.isNewValidator(ctx, store, address) {
		setAcc = true
	}

	if setAcc {
		bz = k.cdc.MustMarshalBinary(validator.abciValidator(k.cdc))
		store.Set(GetAccUpdateValidatorKey(address), bz)

	}

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
	store.Delete(GetValidatorKey(address, candidate.Assets, candidate.ValidatorBondHeight, candidate.ValidatorBondCounter, k.cdc))

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

// Get the validator set from the candidates. The correct subset is retrieved
// by iterating through an index of the candidates sorted by power, stored
// using the ValidatorsKey. Simultaniously the most recent the validator
// records are updated in store with the RecentValidatorsKey. This store is
// used to determine if a candidate is a validator without needing to iterate
// over the subspace as we do in GetValidators
func (k Keeper) GetValidators(ctx sdk.Context) (validators []Validator) {
	store := ctx.KVStore(k.storeKey)

	// clear the recent validators store, add to the ToKickOut Temp store
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
	validators = make([]Validator, maxValidators)
	i := 0
	for ; ; i++ {
		if !iterator.Valid() || i > int(maxValidators-1) {
			iterator.Close()
			break
		}
		bz := iterator.Value()
		var validator Validator
		k.cdc.MustUnmarshalBinary(bz, &validator)
		validators[i] = validator

		// remove from ToKickOut group
		store.Delete(GetToKickOutValidatorKey(validator.Address))

		// also add to the recent validators group
		store.Set(GetRecentValidatorKey(validator.PubKey), bz)

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

	return validators[:i] // trim
}

// TODO this is madly inefficient because need to call every time we set a candidate
// Should use something better than an iterator maybe?
// Used to determine if something has just been added to the actual validator set
func (k Keeper) isNewValidator(ctx sdk.Context, store sdk.KVStore, address sdk.Address) bool {
	// add the actual validator power sorted store
	maxVal := k.GetParams(ctx).MaxValidators
	iterator := store.ReverseSubspaceIterator(ValidatorsKey) // largest to smallest
	for i := 0; ; i++ {
		if !iterator.Valid() || i > int(maxVal-1) {
			iterator.Close()
			break
		}
		bz := iterator.Value()
		var val Validator
		k.cdc.MustUnmarshalBinary(bz, &val)
		if bytes.Equal(val.Address, address) {
			return true
		}
		iterator.Next()
	}

	return false
}

// Is the address provided a part of the most recently saved validator group?
func (k Keeper) IsRecentValidator(ctx sdk.Context, pk crypto.PubKey) bool {
	store := ctx.KVStore(k.storeKey)
	if store.Get(GetRecentValidatorKey(pk)) == nil {
		return false
	}
	return true
}

// Is the power of non-absent prevotes
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
			break
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

func (k Keeper) Validators(ctx sdk.Context) []*sdk.Validator {
	vals := k.GetValidators(ctx)
	res := make([]*sdk.Validator, len(vals))

	for i, val := range vals {
		abcival := val.abciValidator(k.cdc)
		res[i] = &abcival
	}

	return res
}

func (k Keeper) Size(ctx sdk.Context) int {
	return len(k.GetValidators(ctx))
}

func (k Keeper) IsValidator(ctx sdk.Context, addr sdk.Address) bool {
	for _, v := range k.GetValidators(ctx) {
		if bytes.Equal(v.Address, addr) {
			return true
		}
	}
	return false
}

func (k Keeper) GetByAddress(ctx sdk.Context, addr sdk.Address) (int, *sdk.Validator) {
	for i, v := range k.GetValidators(ctx) {
		if bytes.Equal(v.Address, addr) {
			val := v.abciValidator(k.cdc)
			return i, &val
		}
	}
	return -1, nil
}

func (k Keeper) GetByIndex(ctx sdk.Context, index int) *sdk.Validator {
	valset := k.GetValidators(ctx)

	if index < 0 || index >= len(valset) {
		return nil
	}

	val := valset[index].abciValidator(k.cdc)
	return &val
}

func (k Keeper) TotalPower(ctx sdk.Context) sdk.Rat {
	pool := k.GetPool(ctx)
	return pool.BondedShares
}
