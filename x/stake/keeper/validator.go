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
		valToReturn.OperatorAddr = addr
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

// get a single validator by consensus address
func (k Keeper) GetValidatorByConsAddr(ctx sdk.Context, consAddr sdk.ConsAddress) (validator types.Validator, found bool) {
	store := ctx.KVStore(k.storeKey)
	opAddr := store.Get(GetValidatorByConsAddrKey(consAddr))
	if opAddr == nil {
		return validator, false
	}
	return k.GetValidator(ctx, opAddr)
}

// get a single validator by pubkey
func (k Keeper) GetValidatorByConsPubKey(ctx sdk.Context, consPubKey crypto.PubKey) (validator types.Validator, found bool) {
	store := ctx.KVStore(k.storeKey)
	consAddr := sdk.ConsAddress(consPubKey.Address())
	opAddr := store.Get(GetValidatorByConsAddrKey(consAddr))
	if opAddr == nil {
		return validator, false
	}
	return k.GetValidator(ctx, opAddr)
}

// set the main record holding validator details
func (k Keeper) SetValidator(ctx sdk.Context, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshalValidator(k.cdc, validator)
	store.Set(GetValidatorKey(validator.OperatorAddr), bz)
}

// validator index
// TODO change to SetValidatorByConsAddr? used for retrieving from ConsPubkey as well- kinda confusing
func (k Keeper) SetValidatorByConsAddr(ctx sdk.Context, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	consAddr := sdk.ConsAddress(validator.OperatorAddr.Bytes())
	store.Set(GetValidatorByConsAddrKey(consAddr), validator.OperatorAddr)
}

// validator index
func (k Keeper) SetValidatorByPowerIndex(ctx sdk.Context, validator types.Validator, pool types.Pool) {
	store := ctx.KVStore(k.storeKey)
	store.Set(GetValidatorsByPowerIndexKey(validator, pool), validator.OperatorAddr)
}

// validator index
func (k Keeper) SetValidatorBondedIndex(ctx sdk.Context, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	store.Set(GetValidatorsBondedIndexKey(validator.OperatorAddr), []byte{})
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
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		addr := iterator.Key()[1:]
		validator := types.MustUnmarshalValidator(k.cdc, addr, iterator.Value())
		validators = append(validators, validator)
	}
	return validators
}

// return a given amount of all the validators
func (k Keeper) GetValidators(ctx sdk.Context, maxRetrieve uint16) (validators []types.Validator) {
	store := ctx.KVStore(k.storeKey)
	validators = make([]types.Validator, maxRetrieve)

	iterator := sdk.KVStorePrefixIterator(store, ValidatorsKey)
	defer iterator.Close()

	i := 0
	for ; iterator.Valid() && i < int(maxRetrieve); iterator.Next() {
		addr := iterator.Key()[1:]
		validator := types.MustUnmarshalValidator(k.cdc, addr, iterator.Value())
		validators[i] = validator
		i++
	}
	return validators[:i] // trim if the array length < maxRetrieve
}

//___________________________________________________________________________

// get the group of the bonded validators
func (k Keeper) GetValidatorsBonded(ctx sdk.Context) (validators []types.Validator) {
	store := ctx.KVStore(k.storeKey)

	// add the actual validator power sorted store
	maxValidators := k.GetParams(ctx).MaxValidators
	validators = make([]types.Validator, maxValidators)

	iterator := sdk.KVStorePrefixIterator(store, ValidatorsBondedIndexKey)
	defer iterator.Close()

	i := 0
	for ; iterator.Valid(); iterator.Next() {

		// sanity check
		if i > int(maxValidators-1) {
			panic("maxValidators is less than the number of records in ValidatorsBonded Store, store should have been updated")
		}
		address := GetAddressFromValBondedIndexKey(iterator.Key())
		validator, found := k.GetValidator(ctx, address)
		ensureValidatorFound(found, address)

		validators[i] = validator
		i++
	}
	return validators[:i] // trim
}

// get the group of bonded validators sorted by power-rank
//
// TODO: Rename to GetBondedValidatorsByPower or GetValidatorsByPower(ctx, status)
func (k Keeper) GetValidatorsByPower(ctx sdk.Context) []types.Validator {
	store := ctx.KVStore(k.storeKey)
	maxValidators := k.GetParams(ctx).MaxValidators
	validators := make([]types.Validator, maxValidators)

	iterator := sdk.KVStoreReversePrefixIterator(store, ValidatorsByPowerIndexKey)
	defer iterator.Close()

	i := 0
	for ; iterator.Valid() && i < int(maxValidators); iterator.Next() {
		address := iterator.Value()
		validator, found := k.GetValidator(ctx, address)
		ensureValidatorFound(found, address)

		if validator.Status == sdk.Bonded {
			validators[i] = validator
			i++
		}
	}
	return validators[:i] // trim
}

//_________________________________________________________________________
// Accumulated updates to the active/bonded validator set for tendermint

// get the most recently updated validators
//
// CONTRACT: Only validators with non-zero power or zero-power that were bonded
// at the previous block height or were removed from the validator set entirely
// are returned to Tendermint.
func (k Keeper) GetValidTendermintUpdates(ctx sdk.Context) (updates []abci.Validator) {
	tstore := ctx.TransientStore(k.storeTKey)

	iterator := sdk.KVStorePrefixIterator(tstore, TendermintUpdatesTKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var abciVal abci.Validator

		abciValBytes := iterator.Value()
		k.cdc.MustUnmarshalBinary(abciValBytes, &abciVal)

		val, found := k.GetValidator(ctx, abciVal.GetAddress())
		if found {
			// The validator is new or already exists in the store and must adhere to
			// Tendermint invariants.
			prevBonded := val.BondHeight < ctx.BlockHeight() && val.BondHeight > val.UnbondingHeight
			zeroPower := val.GetPower().Equal(sdk.ZeroDec())

			if !zeroPower || zeroPower && prevBonded {
				updates = append(updates, abciVal)
			}
		} else {
			// Add the ABCI validator in such a case where the validator was removed
			// from the store as it must have existed before.
			updates = append(updates, abciVal)
		}
	}
	return
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
	tstore := ctx.TransientStore(k.storeTKey)
	pool := k.GetPool(ctx)
	oldValidator, oldFound := k.GetValidator(ctx, validator.OperatorAddr)

	validator = k.updateForJailing(ctx, oldFound, oldValidator, validator)
	powerIncreasing := k.getPowerIncreasing(ctx, oldFound, oldValidator, validator)
	validator.BondHeight, validator.BondIntraTxCounter = k.bondIncrement(ctx, oldFound, oldValidator)
	valPower := k.updateValidatorPower(ctx, oldFound, oldValidator, validator, pool)
	cliffPower := k.GetCliffValidatorPower(ctx)
	cliffValExists := (cliffPower != nil)
	var valPowerLTcliffPower bool
	if cliffValExists {
		valPowerLTcliffPower = (bytes.Compare(valPower, cliffPower) == -1)
	}

	switch {

	// if the validator is already bonded and the power is increasing, we need
	// perform the following:
	// a) update Tendermint
	// b) check if the cliff validator needs to be updated
	case powerIncreasing && !validator.Jailed &&
		(oldFound && oldValidator.Status == sdk.Bonded):

		bz := k.cdc.MustMarshalBinary(validator.ABCIValidator())
		tstore.Set(GetTendermintUpdatesTKey(validator.OperatorAddr), bz)

		if cliffValExists {
			cliffAddr := sdk.ValAddress(k.GetCliffValidator(ctx))
			if bytes.Equal(cliffAddr, validator.OperatorAddr) {
				k.updateCliffValidator(ctx, validator)
			}
		}

	// if is a new validator and the new power is less than the cliff validator
	case cliffValExists && !oldFound && valPowerLTcliffPower:
		// skip to completion

		// if was unbonded and the new power is less than the cliff validator
	case cliffValExists &&
		(oldFound && oldValidator.Status == sdk.Unbonded) &&
		valPowerLTcliffPower: //(valPower < cliffPower
		// skip to completion

	default:
		// default case - validator was either:
		//  a) not-bonded and now has power-rank greater than  cliff validator
		//  b) bonded and now has decreased in power

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
			tstore.Set(GetTendermintUpdatesTKey(validator.OperatorAddr), bz)
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
		panic(fmt.Sprintf("cliff validator record not found for address: %X\n", cliffAddr))
	}

	// Create a validator iterator ranging from smallest to largest by power
	// starting the current cliff validator's power.
	start := GetValidatorsByPowerIndexKey(oldCliffVal, pool)
	end := sdk.PrefixEndBytes(ValidatorsByPowerIndexKey)
	iterator := store.Iterator(start, end)

	if iterator.Valid() {
		ownerAddr := iterator.Value()
		currVal, found := k.GetValidator(ctx, ownerAddr)
		ensureValidatorFound(found, ownerAddr)

		if currVal.Status != sdk.Bonded || currVal.Jailed {
			panic(fmt.Sprintf("unexpected jailed or unbonded validator for address: %X\n", ownerAddr))
		}

		newCliffVal = currVal
		iterator.Close()
	} else {
		panic("failed to create valid validator power iterator")
	}

	affectedValRank := GetValidatorsByPowerIndexKey(affectedVal, pool)
	newCliffValRank := GetValidatorsByPowerIndexKey(newCliffVal, pool)

	if bytes.Equal(affectedVal.OperatorAddr, newCliffVal.OperatorAddr) {
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
		newValidator = k.beginUnbondingValidator(ctx, newValidator)

		// need to also clear the cliff validator spot because the jail has
		// opened up a new spot which will be filled when
		// updateValidatorsBonded is called
		k.clearCliffValidator(ctx)
	}
	return newValidator
}

// nolint: unparam
func (k Keeper) getPowerIncreasing(ctx sdk.Context, oldFound bool, oldValidator, newValidator types.Validator) bool {
	if oldFound && oldValidator.BondedTokens().LT(newValidator.BondedTokens()) {
		return true
	}
	return false
}

// get the bond height and incremented intra-tx counter
// nolint: unparam
func (k Keeper) bondIncrement(
	ctx sdk.Context, found bool, oldValidator types.Validator) (height int64, intraTxCounter int16) {

	// if already a validator, copy the old block height and counter
	if found && oldValidator.Status == sdk.Bonded {
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
	store.Set(valPower, newValidator.OperatorAddr)

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
	for ; iterator.Valid() && bondedValidatorsCount < int(maxValidators); iterator.Next() {

		// either retrieve the original validator from the store, or under the
		// situation that this is the "affected validator" just use the
		// validator provided because it has not yet been updated in the store
		ownerAddr := iterator.Value()
		if bytes.Equal(ownerAddr, affectedValidator.OperatorAddr) {
			validator = affectedValidator
		} else {
			var found bool
			validator, found = k.GetValidator(ctx, ownerAddr)
			ensureValidatorFound(found, ownerAddr)
		}

		// if we've reached jailed validators no further bonded validators exist
		if validator.Jailed {
			if validator.Status == sdk.Bonded {
				panic(fmt.Sprintf("jailed validator cannot be bonded, address: %X\n", ownerAddr))
			}

			break
		}

		// increment the total number of bonded validators and potentially mark
		// the validator to bond
		if validator.Status != sdk.Bonded {
			validatorToBond = validator
			if newValidatorBonded {
				panic("already decided to bond a validator, can't bond another!")
			}
			newValidatorBonded = true
		}

		bondedValidatorsCount++
	}

	iterator.Close()

	if newValidatorBonded && bytes.Equal(oldCliffValidatorAddr, validator.OperatorAddr) {
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
			ensureValidatorFound(found, oldCliffValidatorAddr)

			if bytes.Equal(validatorToBond.OperatorAddr, affectedValidator.OperatorAddr) {

				// begin unbonding the old cliff validator iff the affected
				// validator was newly bonded and has greater power
				k.beginUnbondingValidator(ctx, oldCliffVal)
			} else {
				// otherwise begin unbonding the affected validator, which must
				// have been kicked out
				affectedValidator = k.beginUnbondingValidator(ctx, affectedValidator)
			}
		}

		validator = k.bondValidator(ctx, validatorToBond)
		if bytes.Equal(validator.OperatorAddr, affectedValidator.OperatorAddr) {
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
	for ; iterator.Valid() && bondedValidatorsCount < int(maxValidators); iterator.Next() {
		var found bool

		ownerAddr := iterator.Value()
		validator, found = k.GetValidator(ctx, ownerAddr)
		ensureValidatorFound(found, ownerAddr)

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
		ensureValidatorFound(found, ownerAddr)
		k.beginUnbondingValidator(ctx, validator)
	}
}

// perform all the store operations for when a validator status becomes unbonded
func (k Keeper) beginUnbondingValidator(ctx sdk.Context, validator types.Validator) types.Validator {

	store := ctx.KVStore(k.storeKey)
	pool := k.GetPool(ctx)
	params := k.GetParams(ctx)

	// sanity check
	if validator.Status == sdk.Unbonded ||
		validator.Status == sdk.Unbonding {
		panic(fmt.Sprintf("should not already be unbonded or unbonding, validator: %v\n", validator))
	}

	// set the status
	validator, pool = validator.UpdateStatus(pool, sdk.Unbonding)
	k.SetPool(ctx, pool)

	validator.UnbondingMinTime = ctx.BlockHeader().Time.Add(params.UnbondingTime)
	validator.UnbondingHeight = ctx.BlockHeader().Height

	// save the now unbonded validator record
	k.SetValidator(ctx, validator)

	// add to accumulated changes for tendermint
	bzABCI := k.cdc.MustMarshalBinary(validator.ABCIValidatorZero())
	tstore := ctx.TransientStore(k.storeTKey)
	tstore.Set(GetTendermintUpdatesTKey(validator.OperatorAddr), bzABCI)

	// also remove from the Bonded types.Validators Store
	store.Delete(GetValidatorsBondedIndexKey(validator.OperatorAddr))

	// call the unbond hook if present
	if k.hooks != nil {
		k.hooks.OnValidatorBeginUnbonding(ctx, validator.ConsAddress())
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

	validator.BondHeight = ctx.BlockHeight()

	// set the status
	validator, pool = validator.UpdateStatus(pool, sdk.Bonded)
	k.SetPool(ctx, pool)

	// save the now bonded validator record to the three referenced stores
	k.SetValidator(ctx, validator)
	store.Set(GetValidatorsBondedIndexKey(validator.OperatorAddr), []byte{})

	// add to accumulated changes for tendermint
	bzABCI := k.cdc.MustMarshalBinary(validator.ABCIValidator())
	tstore := ctx.TransientStore(k.storeTKey)
	tstore.Set(GetTendermintUpdatesTKey(validator.OperatorAddr), bzABCI)

	// call the bond hook if present
	if k.hooks != nil {
		k.hooks.OnValidatorBonded(ctx, validator.ConsAddress())
	}

	// return updated validator
	return validator
}

// remove the validator record and associated indexes
func (k Keeper) RemoveValidator(ctx sdk.Context, address sdk.ValAddress) {

	// call the hook if present
	if k.hooks != nil {
		k.hooks.OnValidatorRemoved(ctx, address)
	}

	// first retrieve the old validator record
	validator, found := k.GetValidator(ctx, address)
	if !found {
		return
	}

	// delete the old validator record
	store := ctx.KVStore(k.storeKey)
	pool := k.GetPool(ctx)
	store.Delete(GetValidatorKey(address))
	store.Delete(GetValidatorByConsAddrKey(sdk.ConsAddress(validator.ConsPubKey.Address())))
	store.Delete(GetValidatorsByPowerIndexKey(validator, pool))

	// delete from the current and power weighted validator groups if the validator
	// is bonded - and add validator with zero power to the validator updates
	if store.Get(GetValidatorsBondedIndexKey(validator.OperatorAddr)) == nil {
		return
	}
	store.Delete(GetValidatorsBondedIndexKey(validator.OperatorAddr))

	bz := k.cdc.MustMarshalBinary(validator.ABCIValidatorZero())
	tstore := ctx.TransientStore(k.storeTKey)
	tstore.Set(GetTendermintUpdatesTKey(address), bz)
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
	store.Set(ValidatorCliffIndexKey, validator.OperatorAddr)
}

// clear the current validator and power of the validator on the cliff
func (k Keeper) clearCliffValidator(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(ValidatorPowerCliffKey)
	store.Delete(ValidatorCliffIndexKey)
}

func ensureValidatorFound(found bool, ownerAddr []byte) {
	if !found {
		panic(fmt.Sprintf("validator record not found for address: %X\n", ownerAddr))
	}
}

//__________________________________________________________________________

// XXX remove this code - this is should be superceded by commission work that bez is doing
// get a single validator
func (k Keeper) UpdateValidatorCommission(ctx sdk.Context, addr sdk.ValAddress, newCommission sdk.Dec) sdk.Error {

	// call the hook if present
	if k.hooks != nil {
		k.hooks.OnValidatorCommissionChange(ctx, addr)
	}

	validator, found := k.GetValidator(ctx, addr)

	// check for errors
	switch {
	case !found:
		return types.ErrNoValidatorFound(k.Codespace())
	case newCommission.LT(sdk.ZeroDec()):
		return types.ErrCommissionNegative(k.Codespace())
	case newCommission.GT(validator.CommissionMax):
		return types.ErrCommissionBeyondMax(k.Codespace())
		//case rateChange(Commission) > CommissionMaxChange:    // XXX XXX XXX TODO implementation
		//return types.ErrCommissionPastRate(k.Codespace())
	}

	// TODO adjust all the commission terms appropriately

	validator.Commission = newCommission

	k.SetValidator(ctx, validator)
	return nil
}
