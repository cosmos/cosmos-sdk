package stake

import (
	"bytes"
	"fmt"

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
	return k.getValidator(store, addr)
}

// get a single validator by pubkey
func (k Keeper) GetValidatorByPubKey(ctx sdk.Context, pubkey crypto.PubKey) (validator Validator, found bool) {
	store := ctx.KVStore(k.storeKey)
	addr := store.Get(GetValidatorByPubKeyIndexKey(pubkey))
	if addr == nil {
		return validator, false
	}
	return k.getValidator(store, addr)
}

// get a single validator (reuse store)
func (k Keeper) getValidator(store sdk.KVStore, addr sdk.Address) (validator Validator, found bool) {
	b := store.Get(GetValidatorKey(addr))
	if b == nil {
		return validator, false
	}
	k.cdc.MustUnmarshalBinary(b, &validator)
	return validator, true
}

// set the main record holding validator details
func (k Keeper) setValidator(ctx sdk.Context, validator Validator) {
	store := ctx.KVStore(k.storeKey)
	// set main store
	bz := k.cdc.MustMarshalBinary(validator)
	store.Set(GetValidatorKey(validator.Owner), bz)
}

func (k Keeper) setValidatorByPubKeyIndex(ctx sdk.Context, validator Validator) {
	store := ctx.KVStore(k.storeKey)
	// set pointer by pubkey
	store.Set(GetValidatorByPubKeyIndexKey(validator.PubKey), validator.Owner)
}

func (k Keeper) setValidatorByPowerIndex(ctx sdk.Context, validator Validator, pool Pool) {
	store := ctx.KVStore(k.storeKey)
	store.Set(GetValidatorsByPowerKey(validator, pool), validator.Owner)
}

// used in testing
func (k Keeper) validatorByPowerIndexExists(ctx sdk.Context, power []byte) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Get(power) != nil
}

// Get the set of all validators with no limits, used during genesis dump
func (k Keeper) getAllValidators(ctx sdk.Context) (validators Validators) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, ValidatorsKey)

	i := 0
	for ; ; i++ {
		if !iterator.Valid() {
			iterator.Close()
			break
		}
		bz := iterator.Value()
		var validator Validator
		k.cdc.MustUnmarshalBinary(bz, &validator)
		validators = append(validators, validator)
		iterator.Next()
	}
	return validators
}

// Get the set of all validators, retrieve a maxRetrieve number of records
func (k Keeper) GetValidators(ctx sdk.Context, maxRetrieve int16) (validators Validators) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, ValidatorsKey)

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

//___________________________________________________________________________

// get the group of the bonded validators
func (k Keeper) GetValidatorsBonded(ctx sdk.Context) (validators []Validator) {
	store := ctx.KVStore(k.storeKey)

	// add the actual validator power sorted store
	maxValidators := k.GetParams(ctx).MaxValidators
	validators = make([]Validator, maxValidators)

	iterator := sdk.KVStorePrefixIterator(store, ValidatorsBondedKey)
	i := 0
	for ; iterator.Valid(); iterator.Next() {

		// sanity check
		if i > int(maxValidators-1) {
			panic("maxValidators is less than the number of records in ValidatorsBonded Store, store should have been updated")
		}
		address := iterator.Value()
		validator, found := k.getValidator(store, address)
		if !found {
			panic(fmt.Sprintf("validator record not found for address: %v\n", address))
		}

		validators[i] = validator
		i++
	}
	iterator.Close()
	return validators[:i] // trim
}

// get the group of bonded validators sorted by power-rank
func (k Keeper) GetValidatorsByPower(ctx sdk.Context) []Validator {
	store := ctx.KVStore(k.storeKey)
	maxValidators := k.GetParams(ctx).MaxValidators
	validators := make([]Validator, maxValidators)
	iterator := sdk.KVStoreReversePrefixIterator(store, ValidatorsByPowerKey) // largest to smallest
	i := 0
	for {
		if !iterator.Valid() || i > int(maxValidators-1) {
			iterator.Close()
			break
		}
		address := iterator.Value()
		validator, found := k.getValidator(store, address)
		if !found {
			panic(fmt.Sprintf("validator record not found for address: %v\n", address))
		}

		// Reached to revoked validators, stop iterating
		if validator.Revoked {
			iterator.Close()
			break
		}
		if validator.Status() == sdk.Bonded {
			validators[i] = validator
			i++
		}
		iterator.Next()
	}
	return validators[:i] // trim
}

//_________________________________________________________________________
// Accumulated updates to the active/bonded validator set for tendermint

// get the most recently updated validators
func (k Keeper) getTendermintUpdates(ctx sdk.Context) (updates []abci.Validator) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, TendermintUpdatesKey) //smallest to largest
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
	iterator := sdk.KVStorePrefixIterator(store, TendermintUpdatesKey)
	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
	}
	iterator.Close()
}

//___________________________________________________________________________

// perfom all the nessisary steps for when a validator changes its power
// updates all validator stores as well as tendermint update store
// may kick out validators if new validator is entering the bonded validator group
func (k Keeper) updateValidator(ctx sdk.Context, validator Validator) Validator {
	store := ctx.KVStore(k.storeKey)
	pool := k.getPool(store)
	ownerAddr := validator.Owner

	// always update the main list ordered by owner address before exiting
	defer func() {
		bz := k.cdc.MustMarshalBinary(validator)
		store.Set(GetValidatorKey(ownerAddr), bz)
	}()

	// retreive the old validator record
	oldValidator, oldFound := k.GetValidator(ctx, ownerAddr)

	if validator.Revoked && oldValidator.Status() == sdk.Bonded {
		validator = k.unbondValidator(ctx, store, validator)

		// need to also clear the cliff validator spot because the revoke has
		// opened up a new spot which will be filled when
		// updateValidatorsBonded is called
		k.clearCliffValidator(ctx)
	}

	powerIncreasing := false
	if oldFound && oldValidator.PoolShares.Bonded().LT(validator.PoolShares.Bonded()) {
		powerIncreasing = true
	}

	// if already a validator, copy the old block height and counter, else set them
	if oldFound && oldValidator.Status() == sdk.Bonded {
		validator.BondHeight = oldValidator.BondHeight
		validator.BondIntraTxCounter = oldValidator.BondIntraTxCounter
	} else {
		validator.BondHeight = ctx.BlockHeight()
		counter := k.getIntraTxCounter(ctx)
		validator.BondIntraTxCounter = counter
		k.setIntraTxCounter(ctx, counter+1)
	}

	// update the list ordered by voting power
	if oldFound {
		store.Delete(GetValidatorsByPowerKey(oldValidator, pool))
	}
	valPower := GetValidatorsByPowerKey(validator, pool)
	store.Set(valPower, validator.Owner)

	// efficiency case:
	// if already bonded and power increasing only need to update tendermint
	if powerIncreasing && !validator.Revoked && oldValidator.Status() == sdk.Bonded {
		bz := k.cdc.MustMarshalBinary(validator.abciValidator(k.cdc))
		store.Set(GetTendermintUpdatesKey(ownerAddr), bz)
		return validator
	}

	// efficiency case:
	// if was unbonded/or is a new validator - and the new power is less than the cliff validator
	cliffPower := k.getCliffValidatorPower(ctx)
	if cliffPower != nil &&
		(!oldFound || (oldFound && oldValidator.Status() == sdk.Unbonded)) &&
		bytes.Compare(valPower, cliffPower) == -1 { //(valPower < cliffPower
		return validator
	}

	// update the validator set for this validator
	updatedVal := k.updateBondedValidators(ctx, store, validator)
	if updatedVal.Owner != nil { // updates to validator occured  to be updated
		validator = updatedVal
	}
	return validator
}

// Update the validator group and kick out any old validators. In addition this
// function adds (or doesn't add) a validator which has updated its bonded
// tokens to the validator group. -> this validator is specified through the
// updatedValidatorAddr term.
//
// The correct subset is retrieved by iterating through an index of the
// validators sorted by power, stored using the ValidatorsByPowerKey.
// Simultaneously the current validator records are updated in store with the
// ValidatorsBondedKey. This store is used to determine if a validator is a
// validator without needing to iterate over the subspace as we do in
// GetValidators.
//
// Optionally also return the validator from a retrieve address if the validator has been bonded
func (k Keeper) updateBondedValidators(ctx sdk.Context, store sdk.KVStore,
	newValidator Validator) (updatedVal Validator) {

	kickCliffValidator := false
	oldCliffValidatorAddr := k.getCliffValidator(ctx)

	// add the actual validator power sorted store
	maxValidators := k.GetParams(ctx).MaxValidators
	iterator := sdk.KVStoreReversePrefixIterator(store, ValidatorsByPowerKey) // largest to smallest
	bondedValidatorsCount := 0
	var validator Validator
	for {
		if !iterator.Valid() || bondedValidatorsCount > int(maxValidators-1) {

			// TODO benchmark if we should read the current power and not write if it's the same
			if bondedValidatorsCount == int(maxValidators) { // is cliff validator
				k.setCliffValidator(ctx, validator, k.GetPool(ctx))
			}
			iterator.Close()
			break
		}

		// either retrieve the original validator from the store, or under the
		// situation that this is the "new validator" just use the validator
		// provided because it has not yet been updated in the main validator
		// store
		ownerAddr := iterator.Value()
		if bytes.Equal(ownerAddr, newValidator.Owner) {
			validator = newValidator
		} else {
			var found bool
			validator, found = k.getValidator(store, ownerAddr)
			if !found {
				panic(fmt.Sprintf("validator record not found for address: %v\n", ownerAddr))
			}
		}

		// if not previously a validator (and unrevoked),
		// kick the cliff validator / bond this new validator
		if validator.Status() != sdk.Bonded && !validator.Revoked {
			kickCliffValidator = true

			validator = k.bondValidator(ctx, store, validator)
			if bytes.Equal(ownerAddr, newValidator.Owner) {
				updatedVal = validator
			}
		}

		if validator.Revoked && validator.Status() == sdk.Bonded {
			panic(fmt.Sprintf("revoked validator cannot be bonded, address: %v\n", ownerAddr))
		} else {
			bondedValidatorsCount++
		}

		iterator.Next()
	}

	// perform the actual kicks
	if oldCliffValidatorAddr != nil && kickCliffValidator {
		validator, found := k.getValidator(store, oldCliffValidatorAddr)
		if !found {
			panic(fmt.Sprintf("validator record not found for address: %v\n", oldCliffValidatorAddr))
		}
		k.unbondValidator(ctx, store, validator)
	}

	return
}

// full update of the bonded validator set, many can be added/kicked
func (k Keeper) updateBondedValidatorsFull(ctx sdk.Context, store sdk.KVStore) {
	// clear the current validators store, add to the ToKickOut temp store
	toKickOut := make(map[string]byte)
	iterator := sdk.KVStorePrefixIterator(store, ValidatorsBondedKey)
	for ; iterator.Valid(); iterator.Next() {
		ownerAddr := iterator.Value()
		toKickOut[string(ownerAddr)] = 0 // set anything
	}
	iterator.Close()

	// add the actual validator power sorted store
	maxValidators := k.GetParams(ctx).MaxValidators
	iterator = sdk.KVStoreReversePrefixIterator(store, ValidatorsByPowerKey) // largest to smallest
	bondedValidatorsCount := 0
	var validator Validator
	for {
		if !iterator.Valid() || bondedValidatorsCount > int(maxValidators-1) {

			if bondedValidatorsCount == int(maxValidators) { // is cliff validator
				k.setCliffValidator(ctx, validator, k.GetPool(ctx))
			}
			iterator.Close()
			break
		}

		// either retrieve the original validator from the store,
		// or under the situation that this is the "new validator" just
		// use the validator provided because it has not yet been updated
		// in the main validator store
		ownerAddr := iterator.Value()
		var found bool
		validator, found = k.getValidator(store, ownerAddr)
		if !found {
			panic(fmt.Sprintf("validator record not found for address: %v\n", ownerAddr))
		}

		_, found = toKickOut[string(ownerAddr)]
		if found {
			delete(toKickOut, string(ownerAddr))
		} else {

			// if it wasn't in the toKickOut group it means
			// this wasn't a previously a validator, therefor
			// update the validator to enter the validator group
			validator = k.bondValidator(ctx, store, validator)
		}

		if validator.Revoked && validator.Status() == sdk.Bonded {
			panic(fmt.Sprintf("revoked validator cannot be bonded, address: %v\n", ownerAddr))
		} else {
			bondedValidatorsCount++
		}

		iterator.Next()
	}

	// perform the actual kicks
	for key := range toKickOut {
		ownerAddr := []byte(key)
		validator, found := k.getValidator(store, ownerAddr)
		if !found {
			panic(fmt.Sprintf("validator record not found for address: %v\n", ownerAddr))
		}
		k.unbondValidator(ctx, store, validator)
	}
	return
}

// perform all the store operations for when a validator status becomes unbonded
func (k Keeper) unbondValidator(ctx sdk.Context, store sdk.KVStore, validator Validator) Validator {
	pool := k.GetPool(ctx)

	// sanity check
	if validator.Status() == sdk.Unbonded {
		panic(fmt.Sprintf("should not already be be unbonded,  validator: %v\n", validator))
	}

	// set the status
	validator, pool = validator.UpdateStatus(pool, sdk.Unbonded)
	k.setPool(ctx, pool)

	// save the now unbonded validator record
	bzVal := k.cdc.MustMarshalBinary(validator)
	store.Set(GetValidatorKey(validator.Owner), bzVal)

	// add to accumulated changes for tendermint
	bzABCI := k.cdc.MustMarshalBinary(validator.abciValidatorZero(k.cdc))
	store.Set(GetTendermintUpdatesKey(validator.Owner), bzABCI)

	// also remove from the Bonded Validators Store
	store.Delete(GetValidatorsBondedKey(validator.PubKey))
	return validator
}

// perform all the store operations for when a validator status becomes bonded
func (k Keeper) bondValidator(ctx sdk.Context, store sdk.KVStore, validator Validator) Validator {
	pool := k.GetPool(ctx)

	// sanity check
	if validator.Status() == sdk.Bonded {
		panic(fmt.Sprintf("should not already be be bonded, validator: %v\n", validator))
	}

	// set the status
	validator, pool = validator.UpdateStatus(pool, sdk.Bonded)
	k.setPool(ctx, pool)

	// save the now bonded validator record to the three referenced stores
	bzVal := k.cdc.MustMarshalBinary(validator)
	store.Set(GetValidatorKey(validator.Owner), bzVal)
	store.Set(GetValidatorsBondedKey(validator.PubKey), validator.Owner)

	// add to accumulated changes for tendermint
	bzABCI := k.cdc.MustMarshalBinary(validator.abciValidator(k.cdc))
	store.Set(GetTendermintUpdatesKey(validator.Owner), bzABCI)

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
	store.Delete(GetValidatorByPubKeyIndexKey(validator.PubKey))
	store.Delete(GetValidatorsByPowerKey(validator, pool))

	// delete from the current and power weighted validator groups if the validator
	// is bonded - and add validator with zero power to the validator updates
	if store.Get(GetValidatorsBondedKey(validator.PubKey)) == nil {
		return
	}
	store.Delete(GetValidatorsBondedKey(validator.PubKey))

	bz := k.cdc.MustMarshalBinary(validator.abciValidatorZero(k.cdc))
	store.Set(GetTendermintUpdatesKey(address), bz)
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

// load all delegations used during genesis dump
func (k Keeper) getAllDelegations(ctx sdk.Context) (delegations []Delegation) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, DelegationKey)

	i := 0
	for ; ; i++ {
		if !iterator.Valid() {
			iterator.Close()
			break
		}
		bondBytes := iterator.Value()
		var delegation Delegation
		k.cdc.MustUnmarshalBinary(bondBytes, &delegation)
		delegations = append(delegations, delegation)
		iterator.Next()
	}
	return delegations[:i] // trim
}

// load all bonds of a delegator
func (k Keeper) GetDelegations(ctx sdk.Context, delegator sdk.Address, maxRetrieve int16) (bonds []Delegation) {
	store := ctx.KVStore(k.storeKey)
	delegatorPrefixKey := GetDelegationsKey(delegator, k.cdc)
	iterator := sdk.KVStorePrefixIterator(store, delegatorPrefixKey) //smallest to largest

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
func (k Keeper) GetParams(ctx sdk.Context) Params {
	store := ctx.KVStore(k.storeKey)
	return k.getParams(store)
}
func (k Keeper) getParams(store sdk.KVStore) (params Params) {
	b := store.Get(ParamKey)
	if b == nil {
		panic("Stored params should not have been nil")
	}

	k.cdc.MustUnmarshalBinary(b, &params)
	return
}

// Need a distinct function because setParams depends on an existing previous
// record of params to exist (to check if maxValidators has changed) - and we
// panic on retrieval if it doesn't exist - hence if we use setParams for the very
// first params set it will panic.
func (k Keeper) setNewParams(ctx sdk.Context, params Params) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinary(params)
	store.Set(ParamKey, b)
}

// Public version of setNewParams
func (k Keeper) SetNewParams(ctx sdk.Context, params Params) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinary(params)
	store.Set(ParamKey, b)
}

func (k Keeper) setParams(ctx sdk.Context, params Params) {
	store := ctx.KVStore(k.storeKey)
	exParams := k.getParams(store)

	// if max validator count changes, must recalculate validator set
	if exParams.MaxValidators != params.MaxValidators {
		k.updateBondedValidatorsFull(ctx, store)
	}
	b := k.cdc.MustMarshalBinary(params)
	store.Set(ParamKey, b)
}

//_______________________________________________________________________

// load/save the pool
func (k Keeper) GetPool(ctx sdk.Context) (pool Pool) {
	store := ctx.KVStore(k.storeKey)
	return k.getPool(store)
}
func (k Keeper) getPool(store sdk.KVStore) (pool Pool) {
	b := store.Get(PoolKey)
	if b == nil {
		panic("Stored pool should not have been nil")
	}
	k.cdc.MustUnmarshalBinary(b, &pool)
	return
}

func (k Keeper) setPool(ctx sdk.Context, pool Pool) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinary(pool)
	store.Set(PoolKey, b)
}

// Public version of setpool
func (k Keeper) SetPool(ctx sdk.Context, pool Pool) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinary(pool)
	store.Set(PoolKey, b)
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

//__________________________________________________________________________

// get the current validator on the cliff
func (k Keeper) getCliffValidator(ctx sdk.Context) []byte {
	store := ctx.KVStore(k.storeKey)
	return store.Get(ValidatorCliffKey)
}

// get the current power of the validator on the cliff
func (k Keeper) getCliffValidatorPower(ctx sdk.Context) []byte {
	store := ctx.KVStore(k.storeKey)
	return store.Get(ValidatorPowerCliffKey)
}

// set the current validator and power of the validator on the cliff
func (k Keeper) setCliffValidator(ctx sdk.Context, validator Validator, pool Pool) {
	store := ctx.KVStore(k.storeKey)
	bz := GetValidatorsByPowerKey(validator, pool)
	store.Set(ValidatorPowerCliffKey, bz)
	store.Set(ValidatorCliffKey, validator.Owner)
}

// clear the current validator and power of the validator on the cliff
func (k Keeper) clearCliffValidator(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(ValidatorPowerCliffKey)
	store.Delete(ValidatorCliffKey)
}

//__________________________________________________________________________

// Implements ValidatorSet

var _ sdk.ValidatorSet = Keeper{}

// iterate through the active validator set and perform the provided function
func (k Keeper) IterateValidators(ctx sdk.Context, fn func(index int64, validator sdk.Validator) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, ValidatorsKey)
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var validator Validator
		k.cdc.MustUnmarshalBinary(bz, &validator)
		stop := fn(i, validator) // XXX is this safe will the validator unexposed fields be able to get written to?
		if stop {
			break
		}
		i++
	}
	iterator.Close()
}

// iterate through the active validator set and perform the provided function
func (k Keeper) IterateValidatorsBonded(ctx sdk.Context, fn func(index int64, validator sdk.Validator) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, ValidatorsBondedKey)
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		address := iterator.Value()
		validator, found := k.getValidator(store, address)
		if !found {
			panic(fmt.Sprintf("validator record not found for address: %v\n", address))
		}

		stop := fn(i, validator) // XXX is this safe will the validator unexposed fields be able to get written to?
		if stop {
			break
		}
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

// Returns self as it is both a validatorset and delegationset
func (k Keeper) GetValidatorSet() sdk.ValidatorSet {
	return k
}

// iterate through the active validator set and perform the provided function
func (k Keeper) IterateDelegations(ctx sdk.Context, delAddr sdk.Address, fn func(index int64, delegation sdk.Delegation) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	key := GetDelegationsKey(delAddr, k.cdc)
	iterator := sdk.KVStorePrefixIterator(store, key)
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var delegation Delegation
		k.cdc.MustUnmarshalBinary(bz, &delegation)
		stop := fn(i, delegation) // XXX is this safe will the fields be able to get written to?
		if stop {
			break
		}
		i++
	}
	iterator.Close()
}

// slash a validator
func (k Keeper) Slash(ctx sdk.Context, pubkey crypto.PubKey, height int64, fraction sdk.Rat) {
	// TODO height ignored for now, see https://github.com/cosmos/cosmos-sdk/pull/1011#issuecomment-390253957
	logger := ctx.Logger().With("module", "x/stake")
	val, found := k.GetValidatorByPubKey(ctx, pubkey)
	if !found {
		panic(fmt.Errorf("attempted to slash a nonexistent validator with address %s", pubkey.Address()))
	}
	sharesToRemove := val.PoolShares.Amount.Mul(fraction)
	pool := k.GetPool(ctx)
	val, pool, burned := val.removePoolShares(pool, sharesToRemove)
	k.setPool(ctx, pool)        // update the pool
	k.updateValidator(ctx, val) // update the validator, possibly kicking it out
	logger.Info(fmt.Sprintf("Validator %s slashed by fraction %v, removed %v shares and burned %v tokens", pubkey.Address(), fraction, sharesToRemove, burned))
	return
}

// revoke a validator
func (k Keeper) Revoke(ctx sdk.Context, pubkey crypto.PubKey) {
	logger := ctx.Logger().With("module", "x/stake")
	val, found := k.GetValidatorByPubKey(ctx, pubkey)
	if !found {
		panic(fmt.Errorf("validator with pubkey %s not found, cannot revoke", pubkey))
	}
	val.Revoked = true
	k.updateValidator(ctx, val) // update the validator, now revoked
	logger.Info(fmt.Sprintf("Validator %s revoked", pubkey.Address()))
	return
}

// unrevoke a validator
func (k Keeper) Unrevoke(ctx sdk.Context, pubkey crypto.PubKey) {
	logger := ctx.Logger().With("module", "x/stake")
	val, found := k.GetValidatorByPubKey(ctx, pubkey)
	if !found {
		panic(fmt.Errorf("validator with pubkey %s not found, cannot unrevoke", pubkey))
	}
	val.Revoked = false
	k.updateValidator(ctx, val) // update the validator, now unrevoked
	logger.Info(fmt.Sprintf("Validator %s unrevoked", pubkey.Address()))
	return
}
