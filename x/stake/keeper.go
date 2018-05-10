package stake

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/bank"
	abci "github.com/tendermint/abci/types"
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

//_________________________________________________________________________

// get a single validator
func (k Keeper) GetValidator(ctx sdk.Context, addr sdk.Address) (validator Validator, found bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(GetValidatorKey(addr))
	if b == nil {
		return validator, false
	}
	k.cdc.MustUnmarshalBinary(b, &validator)
	return validator, true
}

// Get the set of all validators, retrieve a maxRetrieve number of records
func (k Keeper) GetValidators(ctx sdk.Context, maxRetrieve int16) (validators Validators) {
	store := ctx.KVStore(k.storeKey)
	iterator := store.SubspaceIterator(ValidatorsKey)

	validators = make([]Validator, maxRetrieve)
	i := 0
	for ; ; i++ {
		if !iterator.Valid() || i > int(maxRetrieve-1) {
			iterator.Close()
			break
		}
		bz := iterator.Value()
		var validator Validator
		k.cdc.MustUnmarshalBinary(bz, &validator)
		validators[i] = validator
		iterator.Next()
	}
	return validators[:i] // trim
}

func (k Keeper) setValidator(ctx sdk.Context, validator Validator) {
	store := ctx.KVStore(k.storeKey)
	address := validator.Address

	// retreive the old validator record
	oldValidator, oldFound := k.GetValidator(ctx, address)

	// if found, copy the old block height and counter
	if oldFound {
		validator.ValidatorBondHeight = oldValidator.ValidatorBondHeight
		validator.ValidatorBondCounter = oldValidator.ValidatorBondCounter
	}

	// marshal the validator record and add to the state
	bz := k.cdc.MustMarshalBinary(validator)
	store.Set(GetValidatorKey(address), bz)

	if oldFound {
		// if the voting power is the same no need to update any of the other indexes
		if oldValidator.BondedShares.Equal(validator.BondedShares) {
			return
		}

		// if this validator wasn't just bonded then update the height and counter
		if oldValidator.Status != sdk.Bonded {
			validator.ValidatorBondHeight = ctx.BlockHeight()
			counter := k.getIntraTxCounter(ctx)
			validator.ValidatorBondCounter = counter
			k.setIntraTxCounter(ctx, counter+1)
		}

		// delete the old record in the power ordered list
		store.Delete(GetValidatorsBondedByPowerKey(oldValidator))
	}

	// set the new validator record
	bz = k.cdc.MustMarshalBinary(validator)
	store.Set(GetValidatorKey(address), bz)

	// update the list ordered by voting power
	bzVal := k.cdc.MustMarshalBinary(validator)
	store.Set(GetValidatorsBondedByPowerKey(validator), bzVal)

	// add to the validators to update list if is already a validator
	if store.Get(GetValidatorsBondedBondedKey(validator.PubKey)) != nil {
		bzAbci := k.cdc.MustMarshalBinary(validator.abciValidator(k.cdc))
		store.Set(GetValidatorsTendermintUpdatesKey(address), bzAbci)

		// also update the current validator store
		store.Set(GetValidatorsBondedBondedKey(validator.PubKey), bzVal)
		return
	}

	// maybe add to the validator list and kick somebody off
	k.addNewValidatorOrNot(ctx, store, validator.Address)
	return
}

func (k Keeper) removeValidator(ctx sdk.Context, address sdk.Address) {

	// first retreive the old validator record
	validator, found := k.GetValidator(ctx, address)
	if !found {
		return
	}

	// delete the old validator record
	store := ctx.KVStore(k.storeKey)
	store.Delete(GetValidatorKey(address))
	store.Delete(GetValidatorsBondedByPowerKey(validator))

	// delete from current and power weighted validator groups if the validator
	// exists and add validator with zero power to the validator updates
	if store.Get(GetValidatorsBondedBondedKey(validator.PubKey)) == nil {
		return
	}
	bz := k.cdc.MustMarshalBinary(validator.abciValidatorZero(k.cdc))
	store.Set(GetValidatorsTendermintUpdatesKey(address), bz)
	store.Delete(GetValidatorsBondedBondedKey(validator.PubKey))
}

//___________________________________________________________________________

// get the group of the bonded validators
func (k Keeper) GetValidatorsBonded(ctx sdk.Context) (validators []Validator) {
	store := ctx.KVStore(k.storeKey)

	// add the actual validator power sorted store
	maxValidators := k.GetParams(ctx).MaxValidators
	validators = make([]Validator, maxValidators)

	iterator := store.SubspaceIterator(ValidatorsBondedKey)
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

// get the group of bonded validators sorted by power-rank
func (k Keeper) GetValidatorsBondedByPower(ctx sdk.Context) []Validator {
	store := ctx.KVStore(k.storeKey)
	maxValidators := k.GetParams(ctx).MaxValidators
	validators := make([]Validator, maxValidators)
	iterator := store.ReverseSubspaceIterator(ValidatorsByPowerKey) // largest to smallest
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
		iterator.Next()
	}
	return validators
}

// This function add's (or doesn't add) a validator record to the validator group
// simultaniously it kicks any old validators out
//
// The correct subset is retrieved by iterating through an index of the
// validators sorted by power, stored using the ValidatorsByPowerKey. Simultaniously
// the current validator records are updated in store with the
// ValidatorsBondedKey. This store is used to determine if a validator is a
// validator without needing to iterate over the subspace as we do in
// GetValidatorsBonded
func (k Keeper) addNewValidatorOrNot(ctx sdk.Context, store sdk.KVStore, address sdk.Address) {

	// clear the current validators store, add to the ToKickOut temp store
	toKickOut := make(map[[]byte][]byte) // map[key]value
	iterator := store.SubspaceIterator(ValidatorsBondedKey)
	for ; iterator.Valid(); iterator.Next() {

		bz := iterator.Value()
		var validator Validator
		k.cdc.MustUnmarshalBinary(bz, &validator)

		addr := validator.Address

		// iterator.Value is the validator object
		toKickOut[addr] = iterator.Value()
		store.Delete(iterator.Key())
	}
	iterator.Close()

	// add the actual validator power sorted store
	maxValidators := k.GetParams(ctx).MaxValidators
	iterator = store.ReverseSubspaceIterator(ValidatorsByPowerKey) // largest to smallest
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
		toKickOut[validator.Address] = nil

		// also add to the current validators group
		store.Set(GetValidatorsBondedBondedKey(validator.PubKey), bz)

		// MOST IMPORTANTLY, add to the accumulated changes if this is the modified validator
		if bytes.Equal(address, validator.Address) {
			bz = k.cdc.MustMarshalBinary(validator.abciValidator(k.cdc))
			store.Set(GetValidatorsTendermintUpdatesKey(address), bz)
		}

		iterator.Next()
	}

	// add any kicked out validators to the accumulated changes for tendermint
	for key, value := range toKickOut {
		addr := AddrFromKey(key)

		var validator Validator
		k.cdc.MustUnmarshalBinary(value, &validator)
		bz := k.cdc.MustMarshalBinary(validator.abciValidatorZero(k.cdc))
		store.Set(GetValidatorsTendermintUpdatesKey(addr), bz)
	}
}

// cummulative power of the non-absent prevotes
//func (k Keeper) GetTotalPrecommitVotingPower(ctx sdk.Context) sdk.Rat {
//store := ctx.KVStore(k.storeKey)

//// get absent prevote indexes
//absents := ctx.AbsentValidators()

//TotalPower := sdk.ZeroRat()
//i := int32(0)
//iterator := store.SubspaceIterator(ValidatorsBondedKey)
//for ; iterator.Valid(); iterator.Next() {

//skip := false
//for j, absentIndex := range absents {
//if absentIndex > i {
//break
//}

//// if non-voting validator found, skip adding its power
//if absentIndex == i {
//absents = append(absents[:j], absents[j+1:]...) // won't need again
//skip = true
//break
//}
//}
//if skip {
//continue
//}

//bz := iterator.Value()
//var validator Validator
//k.cdc.MustUnmarshalBinary(bz, &validator)
//TotalPower = TotalPower.Add(validator.Power)
//i++
//}
//iterator.Close()
//return TotalPower
//}

//_________________________________________________________________________
// Accumulated updates to the active/bonded validator set for tendermint

// get the most recently updated validators
func (k Keeper) getValidatorsTendermintUpdates(ctx sdk.Context) (updates []abci.Validator) {
	store := ctx.KVStore(k.storeKey)

	iterator := store.SubspaceIterator(ValidatorsTendermintUpdatesKey) //smallest to largest
	for ; iterator.Valid(); iterator.Next() {
		valBytes := iterator.Value()
		var val abci.Validator
		k.cdc.MustUnmarshalBinary(valBytes, &val)
		updates = append(updates, val)
	}
	iterator.Close()
	return
}

// remove all validator update entries after applied to Tendermint
func (k Keeper) clearValidatorsTendermintUpdates(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)

	// delete subspace
	iterator := store.SubspaceIterator(ValidatorsTendermintUpdatesKey)
	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
	}
	iterator.Close()
}

//_____________________________________________________________________

// load a delegator bond
func (k Keeper) GetDelegation(ctx sdk.Context,
	delegatorAddr, validatorAddr sdk.Address) (bond Delegation, found bool) {

	store := ctx.KVStore(k.storeKey)
	delegatorBytes := store.Get(GetDelegationKey(delegatorAddr, validatorAddr, k.cdc))
	if delegatorBytes == nil {
		return bond, false
	}

	k.cdc.MustUnmarshalBinary(delegatorBytes, &bond)
	return bond, true
}

// load all bonds
func (k Keeper) getBonds(ctx sdk.Context, maxRetrieve int16) (bonds []Delegation) {
	store := ctx.KVStore(k.storeKey)
	iterator := store.SubspaceIterator(DelegationKey)

	bonds = make([]Delegation, maxRetrieve)
	i := 0
	for ; ; i++ {
		if !iterator.Valid() || i > int(maxRetrieve-1) {
			iterator.Close()
			break
		}
		bondBytes := iterator.Value()
		var bond Delegation
		k.cdc.MustUnmarshalBinary(bondBytes, &bond)
		bonds[i] = bond
		iterator.Next()
	}
	return bonds[:i] // trim
}

// load all bonds of a delegator
func (k Keeper) GetDelegations(ctx sdk.Context, delegator sdk.Address, maxRetrieve int16) (bonds []Delegation) {
	store := ctx.KVStore(k.storeKey)
	delegatorPrefixKey := GetDelegationsKey(delegator, k.cdc)
	iterator := store.SubspaceIterator(delegatorPrefixKey) //smallest to largest

	bonds = make([]Delegation, maxRetrieve)
	i := 0
	for ; ; i++ {
		if !iterator.Valid() || i > int(maxRetrieve-1) {
			iterator.Close()
			break
		}
		bondBytes := iterator.Value()
		var bond Delegation
		k.cdc.MustUnmarshalBinary(bondBytes, &bond)
		bonds[i] = bond
		iterator.Next()
	}
	return bonds[:i] // trim
}

func (k Keeper) setDelegation(ctx sdk.Context, bond Delegation) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinary(bond)
	store.Set(GetDelegationKey(bond.DelegatorAddr, bond.ValidatorAddr, k.cdc), b)
}

func (k Keeper) removeDelegation(ctx sdk.Context, bond Delegation) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(GetDelegationKey(bond.DelegatorAddr, bond.ValidatorAddr, k.cdc))
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

// Implements ValidatorSet

var _ sdk.ValidatorSet = Keeper{}

// iterate through the active validator set and perform the provided function
func (k Keeper) IterateValidatorsBonded(fn func(index int64, validator sdk.Validator)) {
	store := ctx.KVStore(k.storeKey)
	iterator := store.SubspaceIterator(ValidatorsBondedKey)
	i := 0
	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var validator Validator
		k.cdc.MustUnmarshalBinary(bz, &validator)
		fn(i, validator) // XXX is this safe will the validator unexposed fields be able to get written to?
		i++
	}
	iterator.Close()
}

// get the sdk.validator for a particular address
func (k Keeper) Validator(ctx sdk.Context, addr sdk.Address) sdk.Validator {
	val, found := k.GetValidator(ctx, addr)
	if !found {
		return nil
	}
	return val
}

// total power from the bond
func (k Keeper) TotalPower(ctx sdk.Context) sdk.Rat {
	pool := k.GetPool(ctx)
	return pool.BondedShares
}

//__________________________________________________________________________

// Implements DelegationSet

var _ sdk.ValidatorSet = Keeper{}

// get the delegation for a particular set of delegator and validator addresses
func (k Keeper) Delegation(ctx sdk.Context, addrDel sdk.Address, addrVal sdk.Address) sdk.Delegation {
	bond, ok := k.GetDelegation(ctx, addrDel, addrVal)
	if !ok {
		return nil
	}
	return bond
}

// iterate through the active validator set and perform the provided function
func (k Keeper) IterateDelegators(delAddr sdk.Address, fn func(index int64, delegator sdk.Delegator)) {
	key := GetDelegationsKey(delAddr, k.cdc)
	iterator := store.SubspaceIterator(ValidatorsBondedKey)
	i := 0
	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var delegation Delegation
		k.cdc.MustUnmarshalBinary(bz, &delegation)
		fn(i, delegator) // XXX is this safe will the fields be able to get written to?
		i++
	}
	iterator.Close()
}

//__________________________________________________________________________

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
