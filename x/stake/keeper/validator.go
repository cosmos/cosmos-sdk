package keeper

import (
	"bytes"
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// get a single validator
func (k Keeper) GetValidator(ctx sdk.Context, addr sdk.Address) (validator types.Validator, found bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(GetValidatorKey(addr))
	if b == nil {
		return validator, false
	}
	k.cdc.MustUnmarshalBinary(b, &validator)
	return validator, true
}

// get a single validator by pubkey
func (k Keeper) GetValidatorByPubKey(ctx sdk.Context, pubkey crypto.PubKey) (validator types.Validator, found bool) {
	store := ctx.KVStore(k.storeKey)
	addr := store.Get(GetValidatorByPubKeyIndexKey(pubkey))
	if addr == nil {
		return validator, false
	}
	return k.GetValidator(ctx, addr)
}

// set the main record holding validator details
func (k Keeper) SetValidator(ctx sdk.Context, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	// set main store
	bz := k.cdc.MustMarshalBinary(validator)
	store.Set(GetValidatorKey(validator.Owner), bz)
}

// validator index
func (k Keeper) SetValidatorByPubKeyIndex(ctx sdk.Context, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	// set pointer by pubkey
	store.Set(GetValidatorByPubKeyIndexKey(validator.PubKey), validator.Owner)
}

// validator index
func (k Keeper) SetValidatorByPowerIndex(ctx sdk.Context, validator types.Validator, pool types.Pool) {
	store := ctx.KVStore(k.storeKey)
	store.Set(GetValidatorsByPowerIndexKey(validator, pool), validator.Owner)
}

// validator index
func (k Keeper) SetValidatorBondedIndex(ctx sdk.Context, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	store.Set(GetValidatorsBondedIndexKey(validator.Owner), validator.Owner)
}

// used in testing
func (k Keeper) validatorByPowerIndexExists(ctx sdk.Context, power []byte) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Get(power) != nil
}

// Get the set of all validators with no limits, used during genesis dump
func (k Keeper) GetAllValidators(ctx sdk.Context) (validators []types.Validator) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, ValidatorsKey)

	i := 0
	for ; ; i++ {
		if !iterator.Valid() {
			break
		}
		bz := iterator.Value()
		var validator types.Validator
		k.cdc.MustUnmarshalBinary(bz, &validator)
		validators = append(validators, validator)
		iterator.Next()
	}
	iterator.Close()
	return validators
}

// Get the set of all validators, retrieve a maxRetrieve number of records
func (k Keeper) GetValidators(ctx sdk.Context, maxRetrieve int16) (validators []types.Validator) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, ValidatorsKey)

	validators = make([]types.Validator, maxRetrieve)
	i := 0
	for ; ; i++ {
		if !iterator.Valid() || i > int(maxRetrieve-1) {
			break
		}
		bz := iterator.Value()
		var validator types.Validator
		k.cdc.MustUnmarshalBinary(bz, &validator)
		validators[i] = validator
		iterator.Next()
	}
	iterator.Close()
	return validators[:i] // trim
}

//___________________________________________________________________________

// get the group of the bonded validators
func (k Keeper) GetValidatorsBonded(ctx sdk.Context) (validators []types.Validator) {
	store := ctx.KVStore(k.storeKey)

	// add the actual validator power sorted store
	maxValidators := k.GetParams(ctx).MaxValidators
	validators = make([]types.Validator, maxValidators)

	iterator := sdk.KVStorePrefixIterator(store, ValidatorsBondedIndexKey)
	i := 0
	for ; iterator.Valid(); iterator.Next() {

		// sanity check
		if i > int(maxValidators-1) {
			panic("maxValidators is less than the number of records in ValidatorsBonded Store, store should have been updated")
		}
		address := iterator.Value()
		validator, found := k.GetValidator(ctx, address)
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
func (k Keeper) GetValidatorsByPower(ctx sdk.Context) []types.Validator {
	store := ctx.KVStore(k.storeKey)
	maxValidators := k.GetParams(ctx).MaxValidators
	validators := make([]types.Validator, maxValidators)
	iterator := sdk.KVStoreReversePrefixIterator(store, ValidatorsByPowerIndexKey) // largest to smallest
	i := 0
	for {
		if !iterator.Valid() || i > int(maxValidators-1) {
			break
		}
		address := iterator.Value()
		validator, found := k.GetValidator(ctx, address)
		if !found {
			panic(fmt.Sprintf("validator record not found for address: %v\n", address))
		}
		if validator.Status() == sdk.Bonded {
			validators[i] = validator
			i++
		}
		iterator.Next()
	}
	iterator.Close()
	return validators[:i] // trim
}

//_________________________________________________________________________
// Accumulated updates to the active/bonded validator set for tendermint

// get the most recently updated validators
func (k Keeper) GetTendermintUpdates(ctx sdk.Context) (updates []abci.Validator) {
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
func (k Keeper) ClearTendermintUpdates(ctx sdk.Context) {
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
func (k Keeper) UpdateValidator(ctx sdk.Context, validator types.Validator) types.Validator {
	store := ctx.KVStore(k.storeKey)
	pool := k.GetPool(ctx)
	ownerAddr := validator.Owner

	// always update the main list ordered by owner address before exiting
	defer func() {
		bz := k.cdc.MustMarshalBinary(validator)
		store.Set(GetValidatorKey(ownerAddr), bz)
	}()

	// retrieve the old validator record
	oldValidator, oldFound := k.GetValidator(ctx, ownerAddr)

	if validator.Revoked && oldValidator.Status() == sdk.Bonded {
		validator = k.unbondValidator(ctx, validator)

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
		counter := k.GetIntraTxCounter(ctx)
		validator.BondIntraTxCounter = counter
		k.SetIntraTxCounter(ctx, counter+1)
	}

	// update the list ordered by voting power
	if oldFound {
		store.Delete(GetValidatorsByPowerIndexKey(oldValidator, pool))
	}
	valPower := GetValidatorsByPowerIndexKey(validator, pool)
	store.Set(valPower, validator.Owner)

	// efficiency case:
	// if already bonded and power increasing only need to update tendermint
	if powerIncreasing && !validator.Revoked && oldValidator.Status() == sdk.Bonded {
		bz := k.cdc.MustMarshalBinary(validator.ABCIValidator())
		store.Set(GetTendermintUpdatesKey(ownerAddr), bz)
		return validator
	}

	// efficiency case:
	// if was unbonded/or is a new validator - and the new power is less than the cliff validator
	cliffPower := k.GetCliffValidatorPower(ctx)
	if cliffPower != nil &&
		(!oldFound || (oldFound && oldValidator.Status() == sdk.Unbonded)) &&
		bytes.Compare(valPower, cliffPower) == -1 { //(valPower < cliffPower
		return validator
	}

	// update the validator set for this validator
	updatedVal := k.UpdateBondedValidators(ctx, validator)
	if updatedVal.Owner != nil { // updates to validator occurred  to be updated
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
// validators sorted by power, stored using the ValidatorsByPowerIndexKey.
// Simultaneously the current validator records are updated in store with the
// ValidatorsBondedIndexKey. This store is used to determine if a validator is a
// validator without needing to iterate over the subspace as we do in
// GetValidators.
//
// Optionally also return the validator from a retrieve address if the validator has been bonded
func (k Keeper) UpdateBondedValidators(ctx sdk.Context,
	newValidator types.Validator) (updatedVal types.Validator) {

	store := ctx.KVStore(k.storeKey)

	kickCliffValidator := false
	oldCliffValidatorAddr := k.GetCliffValidator(ctx)

	// add the actual validator power sorted store
	maxValidators := k.GetParams(ctx).MaxValidators
	iterator := sdk.KVStoreReversePrefixIterator(store, ValidatorsByPowerIndexKey) // largest to smallest
	bondedValidatorsCount := 0
	var validator types.Validator
	for {
		if !iterator.Valid() || bondedValidatorsCount > int(maxValidators-1) {

			// TODO benchmark if we should read the current power and not write if it's the same
			if bondedValidatorsCount == int(maxValidators) { // is cliff validator
				k.setCliffValidator(ctx, validator, k.GetPool(ctx))
			}
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
			validator, found = k.GetValidator(ctx, ownerAddr)
			if !found {
				panic(fmt.Sprintf("validator record not found for address: %v\n", ownerAddr))
			}
		}

		// if not previously a validator (and unrevoked),
		// kick the cliff validator / bond this new validator
		if validator.Status() != sdk.Bonded && !validator.Revoked {
			kickCliffValidator = true

			validator = k.bondValidator(ctx, validator)
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
	iterator.Close()

	// perform the actual kicks
	if oldCliffValidatorAddr != nil && kickCliffValidator {
		validator, found := k.GetValidator(ctx, oldCliffValidatorAddr)
		if !found {
			panic(fmt.Sprintf("validator record not found for address: %v\n", oldCliffValidatorAddr))
		}
		k.unbondValidator(ctx, validator)
	}

	return
}

// full update of the bonded validator set, many can be added/kicked
func (k Keeper) UpdateBondedValidatorsFull(ctx sdk.Context) {

	store := ctx.KVStore(k.storeKey)

	// clear the current validators store, add to the ToKickOut temp store
	toKickOut := make(map[string]byte)
	iterator := sdk.KVStorePrefixIterator(store, ValidatorsBondedIndexKey)
	for ; iterator.Valid(); iterator.Next() {
		ownerAddr := iterator.Value()
		toKickOut[string(ownerAddr)] = 0 // set anything
	}
	iterator.Close()

	// add the actual validator power sorted store
	maxValidators := k.GetParams(ctx).MaxValidators
	iterator = sdk.KVStoreReversePrefixIterator(store, ValidatorsByPowerIndexKey) // largest to smallest
	bondedValidatorsCount := 0
	var validator types.Validator
	for {
		if !iterator.Valid() || bondedValidatorsCount > int(maxValidators-1) {

			if bondedValidatorsCount == int(maxValidators) { // is cliff validator
				k.setCliffValidator(ctx, validator, k.GetPool(ctx))
			}
			break
		}

		// either retrieve the original validator from the store,
		// or under the situation that this is the "new validator" just
		// use the validator provided because it has not yet been updated
		// in the main validator store
		ownerAddr := iterator.Value()
		var found bool
		validator, found = k.GetValidator(ctx, ownerAddr)
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
			validator = k.bondValidator(ctx, validator)
		}

		if validator.Revoked && validator.Status() == sdk.Bonded {
			panic(fmt.Sprintf("revoked validator cannot be bonded, address: %v\n", ownerAddr))
		} else {
			bondedValidatorsCount++
		}

		iterator.Next()
	}
	iterator.Close()

	// perform the actual kicks
	for key := range toKickOut {
		ownerAddr := []byte(key)
		validator, found := k.GetValidator(ctx, ownerAddr)
		if !found {
			panic(fmt.Sprintf("validator record not found for address: %v\n", ownerAddr))
		}
		k.unbondValidator(ctx, validator)
	}
	return
}

// perform all the store operations for when a validator status becomes unbonded
func (k Keeper) unbondValidator(ctx sdk.Context, validator types.Validator) types.Validator {

	store := ctx.KVStore(k.storeKey)
	pool := k.GetPool(ctx)

	// sanity check
	if validator.Status() == sdk.Unbonded {
		panic(fmt.Sprintf("should not already be unbonded,  validator: %v\n", validator))
	}

	// set the status
	validator, pool = validator.UpdateStatus(pool, sdk.Unbonded)
	k.SetPool(ctx, pool)

	// save the now unbonded validator record
	bzVal := k.cdc.MustMarshalBinary(validator)
	store.Set(GetValidatorKey(validator.Owner), bzVal)

	// add to accumulated changes for tendermint
	bzABCI := k.cdc.MustMarshalBinary(validator.ABCIValidatorZero())
	store.Set(GetTendermintUpdatesKey(validator.Owner), bzABCI)

	// also remove from the Bonded types.Validators Store
	store.Delete(GetValidatorsBondedIndexKey(validator.Owner))
	return validator
}

// perform all the store operations for when a validator status becomes bonded
func (k Keeper) bondValidator(ctx sdk.Context, validator types.Validator) types.Validator {

	store := ctx.KVStore(k.storeKey)
	pool := k.GetPool(ctx)

	// sanity check
	if validator.Status() == sdk.Bonded {
		panic(fmt.Sprintf("should not already be bonded, validator: %v\n", validator))
	}

	// set the status
	validator, pool = validator.UpdateStatus(pool, sdk.Bonded)
	k.SetPool(ctx, pool)

	// save the now bonded validator record to the three referenced stores
	bzVal := k.cdc.MustMarshalBinary(validator)
	store.Set(GetValidatorKey(validator.Owner), bzVal)
	store.Set(GetValidatorsBondedIndexKey(validator.Owner), validator.Owner)

	// add to accumulated changes for tendermint
	bzABCI := k.cdc.MustMarshalBinary(validator.ABCIValidator())
	store.Set(GetTendermintUpdatesKey(validator.Owner), bzABCI)

	return validator
}

// remove the validator record and associated indexes
func (k Keeper) RemoveValidator(ctx sdk.Context, address sdk.Address) {

	// first retrieve the old validator record
	validator, found := k.GetValidator(ctx, address)
	if !found {
		return
	}

	// delete the old validator record
	store := ctx.KVStore(k.storeKey)
	pool := k.GetPool(ctx)
	store.Delete(GetValidatorKey(address))
	store.Delete(GetValidatorByPubKeyIndexKey(validator.PubKey))
	store.Delete(GetValidatorsByPowerIndexKey(validator, pool))

	// delete from the current and power weighted validator groups if the validator
	// is bonded - and add validator with zero power to the validator updates
	if store.Get(GetValidatorsBondedIndexKey(validator.Owner)) == nil {
		return
	}
	store.Delete(GetValidatorsBondedIndexKey(validator.Owner))

	bz := k.cdc.MustMarshalBinary(validator.ABCIValidatorZero())
	store.Set(GetTendermintUpdatesKey(address), bz)
}

//__________________________________________________________________________

// get the current validator on the cliff
func (k Keeper) GetCliffValidator(ctx sdk.Context) []byte {
	store := ctx.KVStore(k.storeKey)
	return store.Get(ValidatorCliffIndexKey)
}

// get the current power of the validator on the cliff
func (k Keeper) GetCliffValidatorPower(ctx sdk.Context) []byte {
	store := ctx.KVStore(k.storeKey)
	return store.Get(ValidatorPowerCliffKey)
}

// set the current validator and power of the validator on the cliff
func (k Keeper) setCliffValidator(ctx sdk.Context, validator types.Validator, pool types.Pool) {
	store := ctx.KVStore(k.storeKey)
	bz := GetValidatorsByPowerIndexKey(validator, pool)
	store.Set(ValidatorPowerCliffKey, bz)
	store.Set(ValidatorCliffIndexKey, validator.Owner)
}

// clear the current validator and power of the validator on the cliff
func (k Keeper) clearCliffValidator(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(ValidatorPowerCliffKey)
	store.Delete(ValidatorCliffIndexKey)
}
