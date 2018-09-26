package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

//_________________________________________________________________________
// Accumulated updates to the active/bonded validator set for tendermint

// get the most recently updated validators
//
// CONTRACT: Only validators with non-zero power or zero-power that were bonded
// at the previous block height or were removed from the validator set entirely
// are returned to Tendermint.
func (k Keeper) GetTendermintUpdates(ctx sdk.Context) (updates []abci.Validator) {

	// REF CODE
	//// add to accumulated changes for tendermint
	//bzABCI := k.cdc.MustMarshalBinary(validator.ABCIValidator())
	//tstore := ctx.TransientStore(k.storeTKey)
	//tstore.Set(GetTendermintUpdatesTKey(validator.OperatorAddr), bzABCI)

	// XXX code from issue
	/*
		last := fetchOldValidatorSet()
		tendermintUpdates := make(map[sdk.ValAddress]uint64)

		for _, validator := range topvalidator { //(iterate(top hundred)) {
			switch validator.State {
			case Unbonded:
				unbondedToBonded(ctx, validator.Addr)
				tendermintUpdates[validator.Addr] = validator.Power
			case Unbonding:
				unbondingToBonded(ctx, validator.Addr)
				tendermintUpdates[validator.Addr] = validator.Power
			case Bonded: // do nothing
				store.delete(last[validator.Addr])
				// jailed validators are ranked last, so if we get to a jailed validator
				// we have no more bonded validators
				if validator.Jailed {
					break
				}
			}
		}

		for _, validator := range previousValidators {
			bondedToUnbonding(ctx, validator.Addr)
			tendermintUpdates[validator.Addr] = 0
		}

		return tendermintUpdates
	*/

	return updates
}

// XXX FOR REFERENCE USE OR DELETED
func kickOutValidators(k Keeper, ctx sdk.Context, toKickOut map[string]byte) {
	for key := range toKickOut {
		ownerAddr := []byte(key)
		validator := k.mustGetValidator(ctx, ownerAddr)
		k.beginUnbondingValidator(ctx, validator)
	}
}

//___________________________________________________________________________
// State transitions

func (k Keeper) bondedToUnbonding(ctx sdk.Context, validator types.Validator) {
	if validator.Status != sdk.Bonded {
		panic(fmt.Sprintf("bad state transition bondedToUnbonded, validator: %v\n", validator))
	}
	k.beginUnbondingValidator(ctx, validator)
}

func (k Keeper) unbondingToBonded(ctx sdk.Context, validator types.Validator) {
	if validator.Status != sdk.Unbonding {
		panic(fmt.Sprintf("bad state transition unbondingToBonded, validator: %v\n", validator))
	}
	k.bondValidator(ctx, validator)
}

func (k Keeper) unbondedToBonded(ctx sdk.Context, validator types.Validator) {
	if validator.Status != sdk.Unbonded {
		panic(fmt.Sprintf("bad state transition unbondedToBonded, validator: %v\n", validator))
	}
	k.bondValidator(ctx, validator)
}

func (k Keeper) unbondingToUnbonded(ctx sdk.Context, validator types.Validator) {
	if validator.Status != sdk.Unbonded {
		panic(fmt.Sprintf("bad state transition unbondingToBonded, validator: %v\n", validator))
	}
	k.completeUnbondingValidator(ctx, validator)
}

//________________________________________________________________________________________________

// send a validator to jail
func (k Keeper) JailValidator(ctx sdk.Context, validator types.Validator) {
	if validator.Jailed {
		panic(fmt.Sprintf("cannot jail already jailed validator, validator: %v\n", validator))
	}

	pool := k.GetPool(ctx)
	k.DeleteValidatorByPowerIndex(ctx, validator, pool)
	validator.Jailed = true
	k.SetValidator(ctx, validator)
	k.SetValidatorByPowerIndex(ctx, validator, pool)
}

// remove a validator from jail
func (k Keeper) UnjailValidator(ctx sdk.Context, validator types.Validator) {
	if !validator.Jailed {
		panic(fmt.Sprintf("cannot jail already jailed validator, validator: %v\n", validator))
	}

	pool := k.GetPool(ctx)
	k.DeleteValidatorByPowerIndex(ctx, validator, pool)
	validator.Jailed = false
	k.SetValidator(ctx, validator)
	k.SetValidatorByPowerIndex(ctx, validator, pool)
}

//________________________________________________________________________________________________

// perform all the store operations for when a validator status becomes bonded
func (k Keeper) bondValidator(ctx sdk.Context, validator types.Validator) {

	store := ctx.KVStore(k.storeKey)
	pool := k.GetPool(ctx)

	k.DeleteValidatorByPowerIndex(ctx, validator, pool)

	// XXX WHAT DO WE DO FOR BondIntraTxCounter Height Now??????????????????????????

	validator.BondHeight = ctx.BlockHeight()

	// set the status
	validator, pool = validator.UpdateStatus(pool, sdk.Bonded)
	k.SetPool(ctx, pool)

	// save the now bonded validator record to the three referenced stores
	k.SetValidator(ctx, validator)
	store.Set(GetValidatorsBondedIndexKey(validator.OperatorAddr), []byte{})

	k.SetValidatorByPowerIndex(ctx, validator, pool)

	// call the bond hook if present
	if k.hooks != nil {
		k.hooks.OnValidatorBonded(ctx, validator.ConsAddress())
	}
}

// perform all the store operations for when a validator status begins unbonding
func (k Keeper) beginUnbondingValidator(ctx sdk.Context, validator types.Validator) types.Validator {

	store := ctx.KVStore(k.storeKey)
	pool := k.GetPool(ctx)
	params := k.GetParams(ctx)

	k.DeleteValidatorByPowerIndex(ctx, validator, pool)

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

	// also remove from the Bonded types.Validators Store
	store.Delete(GetValidatorsBondedIndexKey(validator.OperatorAddr))

	k.SetValidatorByPowerIndex(ctx, validator, pool)

	// call the unbond hook if present
	if k.hooks != nil {
		k.hooks.OnValidatorBeginUnbonding(ctx, validator.ConsAddress())
	}

	return validator
}

// perform all the store operations for when a validator status becomes unbonded
func (k Keeper) completeUnbondingValidator(ctx sdk.Context, validator types.Validator) {
	pool := k.GetPool(ctx)
	validator, pool = validator.UpdateStatus(pool, sdk.Unbonded)
	k.SetPool(ctx, pool)
	k.SetValidator(ctx, validator)
}

//______________________________________________________________________________________________________

// XXX need to figure out how to set a validator's BondIntraTxCounter - probably during delegation bonding?
//     or keep track of tx for final bonding and set during endblock??? wish we could reduce this complexity

// get the bond height and incremented intra-tx counter
// nolint: unparam
func (k Keeper) bondIncrement(ctx sdk.Context, isNewValidator bool, validator types.Validator) (
	height int64, intraTxCounter int16) {

	// if already a validator, copy the old block height and counter
	if !isNewValidator && validator.Status == sdk.Bonded {
		height = validator.BondHeight
		intraTxCounter = validator.BondIntraTxCounter
		return
	}

	height = ctx.BlockHeight()
	counter := k.GetIntraTxCounter(ctx)
	intraTxCounter = counter

	k.SetIntraTxCounter(ctx, counter+1)
	return
}

//______________________________________________________________________________________________________
/*
// XXX Delete this reference function before merge
// Perform all the necessary steps for when a validator changes its power. This
// function updates all validator stores as well as tendermint update store.
// It may kick out validators if a new validator is entering the bonded validator
// group.
//
// TODO: Remove above nolint, function needs to be simplified!
func (k Keeper) REFERENCEXXXDELETEUpdateValidator(ctx sdk.Context, validator types.Validator) types.Validator {
	tstore := ctx.TransientStore(k.storeTKey)
	pool := k.GetPool(ctx)
	oldValidator, oldFound := k.GetValidator(ctx, validator.OperatorAddr)

	validator = k.updateForJailing(ctx, oldFound, oldValidator, validator)
	powerIncreasing := k.getPowerIncreasing(ctx, oldFound, oldValidator, validator)
	validator.BondHeight, validator.BondIntraTxCounter = k.bondIncrement(ctx, oldFound, oldValidator)
	valPower := k.updateValidatorPower(ctx, oldFound, oldValidator, validator, pool)

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

// XXX Delete this reference function before merge
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
func (k Keeper) XXXREFERENCEUpdateBondedValidators(
	ctx sdk.Context, affectedValidator types.Validator) (
	updatedVal types.Validator, updated bool) {

	store := ctx.KVStore(k.storeKey)

	oldCliffValidatorAddr := k.GetCliffValidator(ctx)
	maxValidators := k.GetParams(ctx).MaxValidators
	bondedValidatorsCount := 0
	var validator, validatorToBond types.Validator
	//newValidatorBonded := false

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
			validator = k.mustGetValidator(ctx, ownerAddr)
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

	// swap the cliff validator for a new validator if the affected validator
	// was bonded
	if newValidatorBonded {
		if oldCliffValidatorAddr != nil {
			oldCliffVal := k.mustGetValidator(ctx, oldCliffValidatorAddr)

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
*/
