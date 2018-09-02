package keeper

import (
	"bytes"
	"container/list"
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// Cache the amino decoding of validators, as it can be the case that repeated slashing calls
// cause many calls to GetValidator, which were shown to throttle the state machine in our
// simulation. Note this is quite biased though, as the simulator does more slashes than a
// live chain should, however we require the slashing to be fast as noone pays gas for it.
type cachedValidator struct {
	val        types.Validator
	marshalled string // marshalled amino bytes for the validator object (not operator address)
}

// validatorCache-key: validator amino bytes
var validatorCache = make(map[string]cachedValidator, 500)
var validatorCacheList = list.New()

// get a single validator
func (k Keeper) GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator types.Validator, found bool) {
	store := ctx.KVStore(k.storeKey)
	value := store.Get(GetValidatorKey(addr))
	if value == nil {
		return validator, false
	}

	// If these amino encoded bytes are in the cache, return the cached validator
	strValue := string(value)
	if val, ok := validatorCache[strValue]; ok {
		valToReturn := val.val
		// Doesn't mutate the cache's value
		valToReturn.Operator = addr
		return valToReturn, true
	}

	// amino bytes weren't found in cache, so amino unmarshal and add it to the cache
	validator = types.MustUnmarshalValidator(k.cdc, addr, value)
	cachedVal := cachedValidator{validator, strValue}
	validatorCache[strValue] = cachedValidator{validator, strValue}
	validatorCacheList.PushBack(cachedVal)

	// if the cache is too big, pop off the last element from it
	if validatorCacheList.Len() > 500 {
		valToRemove := validatorCacheList.Remove(validatorCacheList.Front()).(cachedValidator)
		delete(validatorCache, valToRemove.marshalled)
	}

	validator = types.MustUnmarshalValidator(k.cdc, addr, value)
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
	bz := types.MustMarshalValidator(k.cdc, validator)
	store.Set(GetValidatorKey(validator.Operator), bz)
}

// validator index
func (k Keeper) SetValidatorByPubKeyIndex(ctx sdk.Context, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	store.Set(GetValidatorByPubKeyIndexKey(validator.PubKey), validator.Operator)
}

// validator index
func (k Keeper) SetValidatorByPowerIndex(ctx sdk.Context, validator types.Validator, pool types.Pool) {
	store := ctx.KVStore(k.storeKey)
	store.Set(GetValidatorsByPowerIndexKey(validator, pool), validator.Operator)
}

// validator index
func (k Keeper) SetValidatorBondedIndex(ctx sdk.Context, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	store.Set(GetValidatorsBondedIndexKey(validator.Operator), []byte{})
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
		addr := iterator.Key()[1:]
		validator := types.MustUnmarshalValidator(k.cdc, addr, iterator.Value())
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
		addr := iterator.Key()[1:]
		validator := types.MustUnmarshalValidator(k.cdc, addr, iterator.Value())
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
		address := GetAddressFromValBondedIndexKey(iterator.Key())
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
//
// TODO: Rename to GetBondedValidatorsByPower or GetValidatorsByPower(ctx, status)
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
		if validator.Status == sdk.Bonded {
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

// Perform all the necessary steps for when a validator changes its power. This
// function updates all validator stores as well as tendermint update store.
// It may kick out validators if a new validator is entering the bonded validator
// group.
//
// nolint: gocyclo
// TODO: Remove above nolint, function needs to be simplified!
func (k Keeper) UpdateValidator(ctx sdk.Context, validator types.Validator) types.Validator {
	store := ctx.KVStore(k.storeKey)
	pool := k.GetPool(ctx)
	oldValidator, oldFound := k.GetValidator(ctx, validator.Operator)

	validator = k.updateForJailing(ctx, oldFound, oldValidator, validator)
	powerIncreasing := k.getPowerIncreasing(ctx, oldFound, oldValidator, validator)
	validator.BondHeight, validator.BondIntraTxCounter = k.bondIncrement(ctx, oldFound, oldValidator, validator)
	valPower := k.updateValidatorPower(ctx, oldFound, oldValidator, validator, pool)
	cliffPower := k.GetCliffValidatorPower(ctx)

	switch {
	// if the validator is already bonded and the power is increasing, we need
	// perform the following:
	// a) update Tendermint
	// b) check if the cliff validator needs to be updated
	case powerIncreasing && !validator.Jailed &&
		(oldFound && oldValidator.Status == sdk.Bonded):

		bz := k.cdc.MustMarshalBinary(validator.ABCIValidator())
		store.Set(GetTendermintUpdatesKey(validator.Operator), bz)

		if cliffPower != nil {
			cliffAddr := sdk.ValAddress(k.GetCliffValidator(ctx))
			if bytes.Equal(cliffAddr, validator.Operator) {
				k.updateCliffValidator(ctx, validator)
			}
		}

	// if is a new validator and the new power is less than the cliff validator
	case cliffPower != nil && !oldFound &&
		bytes.Compare(valPower, cliffPower) == -1: //(valPower < cliffPower
		// skip to completion

		// if was unbonded and the new power is less than the cliff validator
	case cliffPower != nil &&
		(oldFound && oldValidator.Status == sdk.Unbonded) &&
		bytes.Compare(valPower, cliffPower) == -1: //(valPower < cliffPower
		// skip to completion

		// default case -  validator was either:
		//  a) not-bonded and now has power-rank greater than  cliff validator
		//  b) bonded and now has decreased in power
	default:
		// update the validator set for this validator
		updatedVal, updated := k.UpdateBondedValidators(ctx, validator)
		if updated {
			// the validator has changed bonding status
			validator = updatedVal
			break
		}

		// if decreased in power but still bonded, update Tendermint validator
		if oldFound && oldValidator.BondedTokens().GT(validator.BondedTokens()) {
			bz := k.cdc.MustMarshalBinary(validator.ABCIValidator())
			store.Set(GetTendermintUpdatesKey(validator.Operator), bz)
		}
	}

	k.SetValidator(ctx, validator)
	return validator
}

// updateCliffValidator determines if the current cliff validator needs to be
// updated or swapped. If the provided affected validator is the current cliff
// validator before it's power was increased, either the cliff power key will
// be updated or if it's power is greater than the next bonded validator by
// power, it'll be swapped.
func (k Keeper) updateCliffValidator(ctx sdk.Context, affectedVal types.Validator) {
	var newCliffVal types.Validator

	store := ctx.KVStore(k.storeKey)
	pool := k.GetPool(ctx)
	cliffAddr := sdk.ValAddress(k.GetCliffValidator(ctx))

	oldCliffVal, found := k.GetValidator(ctx, cliffAddr)
	if !found {
		panic(fmt.Sprintf("cliff validator record not found for address: %v\n", cliffAddr))
	}

	// Create a validator iterator ranging from smallest to largest by power
	// starting the current cliff validator's power.
	start := GetValidatorsByPowerIndexKey(oldCliffVal, pool)
	end := sdk.PrefixEndBytes(ValidatorsByPowerIndexKey)
	iterator := store.Iterator(start, end)

	if iterator.Valid() {
		ownerAddr := iterator.Value()
		currVal, found := k.GetValidator(ctx, ownerAddr)
		if !found {
			panic(fmt.Sprintf("validator record not found for address: %v\n", ownerAddr))
		}

		if currVal.Status != sdk.Bonded || currVal.Jailed {
			panic(fmt.Sprintf("unexpected jailed or unbonded validator for address: %s\n", ownerAddr))
		}

		newCliffVal = currVal
		iterator.Close()
	} else {
		panic("failed to create valid validator power iterator")
	}

	affectedValRank := GetValidatorsByPowerIndexKey(affectedVal, pool)
	newCliffValRank := GetValidatorsByPowerIndexKey(newCliffVal, pool)

	if bytes.Equal(affectedVal.Operator, newCliffVal.Operator) {
		// The affected validator remains the cliff validator, however, since
		// the store does not contain the new power, update the new power rank.
		store.Set(ValidatorPowerCliffKey, affectedValRank)
	} else if bytes.Compare(affectedValRank, newCliffValRank) > 0 {
		// The affected validator no longer remains the cliff validator as it's
		// power is greater than the new cliff validator.
		k.setCliffValidator(ctx, newCliffVal, pool)
	} else {
		panic("invariant broken: the cliff validator should change or it should remain the same")
	}
}

func (k Keeper) updateForJailing(ctx sdk.Context, oldFound bool, oldValidator, newValidator types.Validator) types.Validator {
	if newValidator.Jailed && oldFound && oldValidator.Status == sdk.Bonded {
		newValidator = k.unbondValidator(ctx, newValidator)

		// need to also clear the cliff validator spot because the jail has
		// opened up a new spot which will be filled when
		// updateValidatorsBonded is called
		k.clearCliffValidator(ctx)
	}
	return newValidator
}

func (k Keeper) getPowerIncreasing(ctx sdk.Context, oldFound bool, oldValidator, newValidator types.Validator) bool {
	if oldFound && oldValidator.BondedTokens().LT(newValidator.BondedTokens()) {
		return true
	}
	return false
}

// get the bond height and incremented intra-tx counter
func (k Keeper) bondIncrement(ctx sdk.Context, oldFound bool, oldValidator,
	newValidator types.Validator) (height int64, intraTxCounter int16) {

	// if already a validator, copy the old block height and counter, else set them
	if oldFound && oldValidator.Status == sdk.Bonded {
		height = oldValidator.BondHeight
		intraTxCounter = oldValidator.BondIntraTxCounter
		return
	}
	height = ctx.BlockHeight()
	counter := k.GetIntraTxCounter(ctx)
	intraTxCounter = counter
	k.SetIntraTxCounter(ctx, counter+1)
	return
}

func (k Keeper) updateValidatorPower(ctx sdk.Context, oldFound bool, oldValidator,
	newValidator types.Validator, pool types.Pool) (valPower []byte) {
	store := ctx.KVStore(k.storeKey)

	// update the list ordered by voting power
	if oldFound {
		store.Delete(GetValidatorsByPowerIndexKey(oldValidator, pool))
	}
	valPower = GetValidatorsByPowerIndexKey(newValidator, pool)
	store.Set(valPower, newValidator.Operator)

	return valPower
}

// Update the bonded validator group based on a change to the validator
// affectedValidator. This function potentially adds the affectedValidator to
// the bonded validator group which kicks out the cliff validator. Under this
// situation this function returns the updated affectedValidator.
//
// The correct bonded subset of validators is retrieved by iterating through an
// index of the validators sorted by power, stored using the
// ValidatorsByPowerIndexKey.  Simultaneously the current validator records are
// updated in store with the ValidatorsBondedIndexKey. This store is used to
// determine if a validator is a validator without needing to iterate over all
// validators.
//
// nolint: gocyclo
// TODO: Remove the above golint
func (k Keeper) UpdateBondedValidators(
	ctx sdk.Context, affectedValidator types.Validator) (
	updatedVal types.Validator, updated bool) {

	store := ctx.KVStore(k.storeKey)

	oldCliffValidatorAddr := k.GetCliffValidator(ctx)
	maxValidators := k.GetParams(ctx).MaxValidators
	bondedValidatorsCount := 0
	var validator, validatorToBond types.Validator
	newValidatorBonded := false

	// create a validator iterator ranging from largest to smallest by power
	iterator := sdk.KVStoreReversePrefixIterator(store, ValidatorsByPowerIndexKey)
	for {
		if !iterator.Valid() || bondedValidatorsCount > int(maxValidators-1) {
			break
		}

		// either retrieve the original validator from the store, or under the
		// situation that this is the "affected validator" just use the
		// validator provided because it has not yet been updated in the store
		ownerAddr := iterator.Value()
		if bytes.Equal(ownerAddr, affectedValidator.Operator) {
			validator = affectedValidator
		} else {
			var found bool
			validator, found = k.GetValidator(ctx, ownerAddr)
			if !found {
				panic(fmt.Sprintf("validator record not found for address: %v\n", ownerAddr))
			}
		}

		// increment bondedValidatorsCount / get the validator to bond
		if !validator.Jailed {
			if validator.Status != sdk.Bonded {
				validatorToBond = validator
				if newValidatorBonded {
					panic("already decided to bond a validator, can't bond another!")
				}
				newValidatorBonded = true
			}
		} else {
			// TODO: document why we must break here.
			break
		}

		// increment the total number of bonded validators and potentially mark
		// the validator to bond
		if validator.Status != sdk.Bonded {
			validatorToBond = validator
			newValidatorBonded = true
		}

		bondedValidatorsCount++
		iterator.Next()
	}

	iterator.Close()

	if newValidatorBonded && bytes.Equal(oldCliffValidatorAddr, validator.Operator) {
		panic("cliff validator has not been changed, yet we bonded a new validator")
	}

	// clear or set the cliff validator
	if bondedValidatorsCount == int(maxValidators) {
		k.setCliffValidator(ctx, validator, k.GetPool(ctx))
	} else if len(oldCliffValidatorAddr) > 0 {
		k.clearCliffValidator(ctx)
	}

	// swap the cliff validator for a new validator if the affected validator
	// was bonded
	if newValidatorBonded {
		if oldCliffValidatorAddr != nil {
			oldCliffVal, found := k.GetValidator(ctx, oldCliffValidatorAddr)
			if !found {
				panic(fmt.Sprintf("validator record not found for address: %v\n", oldCliffValidatorAddr))
			}

			if bytes.Equal(validatorToBond.Operator, affectedValidator.Operator) {
				// unbond the old cliff validator iff the affected validator was
				// newly bonded and has greater power
				k.unbondValidator(ctx, oldCliffVal)
			} else {
				// otherwise unbond the affected validator, which must have been
				// kicked out
				affectedValidator = k.unbondValidator(ctx, affectedValidator)
			}
		}

		validator = k.bondValidator(ctx, validatorToBond)
		if bytes.Equal(validator.Operator, affectedValidator.Operator) {
			return validator, true
		}

		return affectedValidator, true
	}

	return types.Validator{}, false
}

// full update of the bonded validator set, many can be added/kicked
func (k Keeper) UpdateBondedValidatorsFull(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)

	// clear the current validators store, add to the ToKickOut temp store
	toKickOut := make(map[string]byte)
	iterator := sdk.KVStorePrefixIterator(store, ValidatorsBondedIndexKey)
	for ; iterator.Valid(); iterator.Next() {
		ownerAddr := GetAddressFromValBondedIndexKey(iterator.Key())
		toKickOut[string(ownerAddr)] = 0
	}

	iterator.Close()

	var validator types.Validator

	oldCliffValidatorAddr := k.GetCliffValidator(ctx)
	maxValidators := k.GetParams(ctx).MaxValidators
	bondedValidatorsCount := 0

	iterator = sdk.KVStoreReversePrefixIterator(store, ValidatorsByPowerIndexKey)
	for {
		if !iterator.Valid() || bondedValidatorsCount > int(maxValidators-1) {
			break
		}

		var found bool

		ownerAddr := iterator.Value()
		validator, found = k.GetValidator(ctx, ownerAddr)
		if !found {
			panic(fmt.Sprintf("validator record not found for address: %v\n", ownerAddr))
		}

		_, found = toKickOut[string(ownerAddr)]
		if found {
			delete(toKickOut, string(ownerAddr))
		} else {
			// If the validator wasn't in the toKickOut group it means it wasn't
			// previously a validator, therefor update the validator to enter
			// the validator group.
			validator = k.bondValidator(ctx, validator)
		}

		if validator.Jailed {
			// we should no longer consider jailed validators as they are ranked
			// lower than any non-jailed/bonded validators
			if validator.Status == sdk.Bonded {
				panic(fmt.Sprintf("jailed validator cannot be bonded for address: %s\n", ownerAddr))
			}

			break
		}

		bondedValidatorsCount++
		iterator.Next()
	}

	iterator.Close()

	// clear or set the cliff validator
	if bondedValidatorsCount == int(maxValidators) {
		k.setCliffValidator(ctx, validator, k.GetPool(ctx))
	} else if len(oldCliffValidatorAddr) > 0 {
		k.clearCliffValidator(ctx)
	}

	kickOutValidators(k, ctx, toKickOut)
	return
}

func kickOutValidators(k Keeper, ctx sdk.Context, toKickOut map[string]byte) {
	for key := range toKickOut {
		ownerAddr := []byte(key)
		validator, found := k.GetValidator(ctx, ownerAddr)
		if !found {
			panic(fmt.Sprintf("validator record not found for address: %v\n", ownerAddr))
		}
		k.unbondValidator(ctx, validator)
	}
}

// perform all the store operations for when a validator status becomes unbonded
func (k Keeper) unbondValidator(ctx sdk.Context, validator types.Validator) types.Validator {

	store := ctx.KVStore(k.storeKey)
	pool := k.GetPool(ctx)

	// sanity check
	if validator.Status == sdk.Unbonded {
		panic(fmt.Sprintf("should not already be unbonded, validator: %v\n", validator))
	}

	// set the status
	validator, pool = validator.UpdateStatus(pool, sdk.Unbonded)
	k.SetPool(ctx, pool)

	// save the now unbonded validator record
	k.SetValidator(ctx, validator)

	// add to accumulated changes for tendermint
	bzABCI := k.cdc.MustMarshalBinary(validator.ABCIValidatorZero())
	store.Set(GetTendermintUpdatesKey(validator.Operator), bzABCI)

	// also remove from the Bonded types.Validators Store
	store.Delete(GetValidatorsBondedIndexKey(validator.Operator))

	// call the unbond hook if present
	if k.validatorHooks != nil {
		k.validatorHooks.OnValidatorBeginUnbonding(ctx, validator.ConsAddress())
	}

	// return updated validator
	return validator
}

// perform all the store operations for when a validator status becomes bonded
func (k Keeper) bondValidator(ctx sdk.Context, validator types.Validator) types.Validator {

	store := ctx.KVStore(k.storeKey)
	pool := k.GetPool(ctx)

	// sanity check
	if validator.Status == sdk.Bonded {
		panic(fmt.Sprintf("should not already be bonded, validator: %v\n", validator))
	}

	// set the status
	validator, pool = validator.UpdateStatus(pool, sdk.Bonded)
	k.SetPool(ctx, pool)

	// save the now bonded validator record to the three referenced stores
	k.SetValidator(ctx, validator)
	store.Set(GetValidatorsBondedIndexKey(validator.Operator), []byte{})

	// add to accumulated changes for tendermint
	bzABCI := k.cdc.MustMarshalBinary(validator.ABCIValidator())
	store.Set(GetTendermintUpdatesKey(validator.Operator), bzABCI)

	// call the bond hook if present
	if k.validatorHooks != nil {
		k.validatorHooks.OnValidatorBonded(ctx, validator.ConsAddress())
	}

	// return updated validator
	return validator
}

// remove the validator record and associated indexes
func (k Keeper) RemoveValidator(ctx sdk.Context, address sdk.ValAddress) {

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
	if store.Get(GetValidatorsBondedIndexKey(validator.Operator)) == nil {
		return
	}
	store.Delete(GetValidatorsBondedIndexKey(validator.Operator))

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
	store.Set(ValidatorCliffIndexKey, validator.Operator)
}

// clear the current validator and power of the validator on the cliff
func (k Keeper) clearCliffValidator(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(ValidatorPowerCliffKey)
	store.Delete(ValidatorCliffIndexKey)
}
