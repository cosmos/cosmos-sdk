package stake

import (
	"bytes"
	"fmt"

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

func (k Keeper) setValidator(ctx sdk.Context, validator Validator) Validator {
	fmt.Println("wackydebugoutput setValidator 0")
	store := ctx.KVStore(k.storeKey)
	pool := k.getPool(store)
	address := validator.Address

	// update the main list ordered by address before exiting
	defer func() {
		fmt.Println("wackydebugoutput setValidator 1")
		bz := k.cdc.MustMarshalBinary(validator)
		store.Set(GetValidatorKey(address), bz)
	}()
	fmt.Println("wackydebugoutput setValidator 2")

	// retreive the old validator record
	oldValidator, oldFound := k.GetValidator(ctx, address)

	powerIncreasing := false
	if oldFound {
		fmt.Println("wackydebugoutput setValidator 3")
		// if the voting power is the same no need to update any of the other indexes
		if oldValidator.Status == sdk.Bonded &&
			oldValidator.PShares.Equal(validator.PShares) {
			fmt.Println("wackydebugoutput setValidator 4")
			return validator
		} else if oldValidator.PShares.Bonded().LT(validator.PShares.Bonded()) {
			fmt.Println("wackydebugoutput setValidator 5")
			powerIncreasing = true
		}
		fmt.Println("wackydebugoutput setValidator 6")
		// delete the old record in the power ordered list
		store.Delete(GetValidatorsBondedByPowerKey(oldValidator, pool))
	}
	fmt.Println("wackydebugoutput setValidator 7")

	// if already a validator, copy the old block height and counter, else set them
	if oldFound && oldValidator.Status == sdk.Bonded {
		fmt.Println("wackydebugoutput setValidator 8")
		validator.BondHeight = oldValidator.BondHeight
		validator.BondIntraTxCounter = oldValidator.BondIntraTxCounter
	} else {
		fmt.Println("wackydebugoutput setValidator 9")
		validator.BondHeight = ctx.BlockHeight()
		counter := k.getIntraTxCounter(ctx)
		validator.BondIntraTxCounter = counter
		k.setIntraTxCounter(ctx, counter+1)
	}
	fmt.Println("wackydebugoutput setValidator 10")

	// update the list ordered by voting power
	bz := k.cdc.MustMarshalBinary(validator)
	store.Set(GetValidatorsBondedByPowerKey(validator, pool), bz)

	// add to the validators and return to update list if is already a validator and power is increasing
	if powerIncreasing && oldFound && oldValidator.Status == sdk.Bonded {
		fmt.Println("wackydebugoutput setValidator 11")

		// update the recent validator store
		store.Set(GetValidatorsBondedKey(validator.PubKey), bz)

		// and the Tendermint updates
		bz := k.cdc.MustMarshalBinary(validator.abciValidator(k.cdc))
		store.Set(GetTendermintUpdatesKey(address), bz)
		return validator
	}
	fmt.Println("wackydebugoutput setValidator 12")

	// update the validator set for this validator
	valIsNowBonded := k.updateValidators(ctx, store, validator.Address)

	if (!oldFound && valIsNowBonded) ||
		(oldFound && oldValidator.Status != sdk.Bonded && valIsNowBonded) {
		fmt.Println("wackydebugoutput setValidator 13")

		validator.Status = sdk.Bonded
		validator, pool = validator.UpdateSharesLocation(pool)
		k.setPool(ctx, pool)
	}
	fmt.Println("wackydebugoutput setValidator 14")
	return validator
}

func (k Keeper) removeValidator(ctx sdk.Context, address sdk.Address) {

	// first retreive the old validator record
	validator, found := k.GetValidator(ctx, address)
	if !found {
		return
	}

	// delete the old validator record
	store := ctx.KVStore(k.storeKey)
	pool := k.getPool(store)
	store.Delete(GetValidatorKey(address))
	store.Delete(GetValidatorsBondedByPowerKey(validator, pool))

	// delete from current and power weighted validator groups if the validator
	// exists and add validator with zero power to the validator updates
	if store.Get(GetValidatorsBondedKey(validator.PubKey)) == nil {
		return
	}
	bz := k.cdc.MustMarshalBinary(validator.abciValidatorZero(k.cdc))
	store.Set(GetTendermintUpdatesKey(address), bz)
	store.Delete(GetValidatorsBondedKey(validator.PubKey))
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
	for {
		if !iterator.Valid() || i > int(maxValidators-1) {
			iterator.Close()
			break
		}
		bz := iterator.Value()
		var validator Validator
		k.cdc.MustUnmarshalBinary(bz, &validator)
		if validator.Status == sdk.Bonded {
			validators[i] = validator
			i++
		}
		iterator.Next()
	}
	return validators[:i] // trim
}

// XXX TODO build in consideration for revoked
//
// Update the validator group and kick out any old validators. In addition this
// function adds (or doesn't add) a validator which has updated its bonded
// tokens to the validator group. -> this validator is specified through the
// updatedValidatorAddr term.
//
// The correct subset is retrieved by iterating through an index of the
// validators sorted by power, stored using the ValidatorsByPowerKey. Simultaniously
// the current validator records are updated in store with the
// ValidatorsBondedKey. This store is used to determine if a validator is a
// validator without needing to iterate over the subspace as we do in
// GetValidators.
func (k Keeper) updateValidators(ctx sdk.Context, store sdk.KVStore, updatedValidatorAddr sdk.Address) (updatedIsBonded bool) {
	updatedIsBonded = false

	// clear the current validators store, add to the ToKickOut temp store
	toKickOut := make(map[string][]byte) // map[key]value
	iterator := store.SubspaceIterator(ValidatorsBondedKey)
	for ; iterator.Valid(); iterator.Next() {

		bz := iterator.Value()
		var validator Validator
		k.cdc.MustUnmarshalBinary(bz, &validator)

		addr := validator.Address

		// iterator.Value is the validator object
		toKickOut[string(addr)] = iterator.Value()
		store.Delete(iterator.Key())
	}
	iterator.Close()

	// add the actual validator power sorted store
	maxValidators := k.GetParams(ctx).MaxValidators
	fmt.Printf("debug maxValidators: %v\n", maxValidators)
	iterator = store.ReverseSubspaceIterator(ValidatorsByPowerKey) // largest to smallest
	i := 0
	for ; ; i++ {
		fmt.Printf("debug i: %v\n", i)
		if !iterator.Valid() || i > int(maxValidators-1) {
			iterator.Close()
			break
		}
		bz := iterator.Value()
		var validator Validator
		k.cdc.MustUnmarshalBinary(bz, &validator)

		// remove from ToKickOut group
		toKickOut[string(validator.Address)] = nil

		// also add to the current validators group
		store.Set(GetValidatorsBondedKey(validator.PubKey), bz)

		// MOST IMPORTANTLY, add to the accumulated changes if this is the modified validator
		if bytes.Equal(updatedValidatorAddr, validator.Address) {
			bz = k.cdc.MustMarshalBinary(validator.abciValidator(k.cdc))
			store.Set(GetTendermintUpdatesKey(updatedValidatorAddr), bz)
			updatedIsBonded = true // the updatedValidatorAddr is for a bonded validator
		}

		iterator.Next()
	}

	for _, value := range toKickOut {
		if value == nil {
			continue
		}
		var validator Validator
		k.cdc.MustUnmarshalBinary(value, &validator)
		fmt.Printf("debug validator: %v\n", validator)
		k.unbondValidator(ctx, store, validator)
	}
	return
}

// perform all the store operations for when a validator is moved from bonded to unbonded
func (k Keeper) unbondValidator(ctx sdk.Context, store sdk.KVStore, validator Validator) {
	pool := k.GetPool(ctx)

	// set the status
	validator.Status = sdk.Unbonded
	validator, pool = validator.UpdateSharesLocation(pool)
	k.setPool(ctx, pool)

	// save the now unbonded validator record
	bz := k.cdc.MustMarshalBinary(validator)
	store.Set(GetValidatorKey(validator.Address), bz)

	// add to accumulated changes for tendermint
	bz = k.cdc.MustMarshalBinary(validator.abciValidatorZero(k.cdc))
	store.Set(GetTendermintUpdatesKey(validator.Address), bz)

	// also remove from the Bonded Validators Store
	store.Delete(GetValidatorsBondedKey(validator.PubKey))
}

//_________________________________________________________________________
// Accumulated updates to the active/bonded validator set for tendermint

// get the most recently updated validators
func (k Keeper) getTendermintUpdates(ctx sdk.Context) (updates []abci.Validator) {
	store := ctx.KVStore(k.storeKey)

	iterator := store.SubspaceIterator(TendermintUpdatesKey) //smallest to largest
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
func (k Keeper) clearTendermintUpdates(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)

	// delete subspace
	iterator := store.SubspaceIterator(TendermintUpdatesKey)
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

	// if max validator count changes, must recalculate validator set
	if k.params.MaxValidators != params.MaxValidators {
		k.updateValidators(ctx, store, sdk.Address{})
	}
	k.params = params // update the cache
}

//_______________________________________________________________________

// load/save the pool
func (k Keeper) GetPool(ctx sdk.Context) (pool Pool) {
	store := ctx.KVStore(k.storeKey)
	return k.getPool(store)
}
func (k Keeper) getPool(store sdk.KVStore) (pool Pool) {
	// check if cached before anything
	if !k.pool.equal(Pool{}) {
		return k.pool
	}
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
func (k Keeper) IterateValidatorsBonded(ctx sdk.Context, fn func(index int64, validator sdk.Validator)) {
	store := ctx.KVStore(k.storeKey)
	iterator := store.SubspaceIterator(ValidatorsBondedKey)
	i := int64(0)
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
func (k Keeper) IterateDelegators(ctx sdk.Context, delAddr sdk.Address, fn func(index int64, delegation sdk.Delegation)) {
	store := ctx.KVStore(k.storeKey)
	key := GetDelegationsKey(delAddr, k.cdc)
	iterator := store.SubspaceIterator(key)
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var delegation Delegation
		k.cdc.MustUnmarshalBinary(bz, &delegation)
		fn(i, delegation) // XXX is this safe will the fields be able to get written to?
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
