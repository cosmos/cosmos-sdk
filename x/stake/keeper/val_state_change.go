package keeper

import (
	"bytes"
	"fmt"
	"sort"

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
func (k Keeper) GetTendermintUpdates(ctx sdk.Context) (updates []abci.ValidatorUpdate) {

	store := ctx.KVStore(k.storeKey)
	maxValidators := k.GetParams(ctx).MaxValidators

	// copy last validator set
	last := make(map[[sdk.AddrLen]byte]interface{})
	iterator := sdk.KVStorePrefixIterator(store, BondedValidatorsIndexKey)
	for ; iterator.Valid(); iterator.Next() {
		var operator [sdk.AddrLen]byte
		copy(operator[:], iterator.Key()[1:])
		last[operator] = nil
	}

	iterator = sdk.KVStoreReversePrefixIterator(store, ValidatorsByPowerIndexKey)
	count := 0
	for ; iterator.Valid() && count < int(maxValidators); iterator.Next() {

		// fetch the validator
		operator := sdk.ValAddress(iterator.Value())
		validator := k.mustGetValidator(ctx, operator)

		// jailed validators are ranked last, so if we get to a jailed validator
		// we have no more bonded validators
		// TODO we can remove this if we remove jailed validators from the power store
		if validator.Jailed {
			break
		}

		// apply the appropriate state change if necessary
		switch validator.Status {
		case sdk.Unbonded:
			validator = k.unbondedToBonded(ctx, validator)
		case sdk.Unbonding:
			validator = k.unbondingToBonded(ctx, validator)
		case sdk.Bonded:
			// no state change
		}

		// update the validator (power might have changed)
		updates = append(updates, validator.ABCIValidatorUpdate())

		// validator still in the validator set
		var opbytes [sdk.AddrLen]byte
		copy(opbytes[:], operator[:])
		delete(last, opbytes)

		// set the bonded validator index
		// TODO move me
		store.Set(GetBondedValidatorIndexKey(operator), []byte{})

		// keep count
		count++

	}

	// sort the map keys for determinism
	noLongerBonded := make([][]byte, len(last))
	index := 0
	for oper := range last {
		operator := make([]byte, sdk.AddrLen)
		copy(operator[:], oper[:])
		noLongerBonded[index] = operator
		index++
	}
	sort.SliceStable(noLongerBonded, func(i, j int) bool {
		return bytes.Compare(noLongerBonded[i], noLongerBonded[j]) == -1
	})

	// iterate through the sorted no-longer-bonded validators
	// any validators left in `last` are no longer bonded
	for _, operator := range noLongerBonded {
		// fetch the validator
		// TODO might it have been deleted in RemoveValidator?
		validator := k.mustGetValidator(ctx, sdk.ValAddress(operator))

		// bonded to unbonding
		k.bondedToUnbonding(ctx, validator)

		// delete from the bonded validator index
		// TODO move me
		store.Delete(GetBondedValidatorIndexKey(operator))

		// update the validator
		updates = append(updates, validator.ABCIValidatorUpdateZero())
	}

	return updates
}

// Validator state transitions

func (k Keeper) bondedToUnbonding(ctx sdk.Context, validator types.Validator) types.Validator {
	if validator.Status != sdk.Bonded {
		panic(fmt.Sprintf("bad state transition bondedToUnbonded, validator: %v\n", validator))
	}
	return k.beginUnbondingValidator(ctx, validator)
}

func (k Keeper) unbondingToBonded(ctx sdk.Context, validator types.Validator) types.Validator {
	if validator.Status != sdk.Unbonding {
		panic(fmt.Sprintf("bad state transition unbondingToBonded, validator: %v\n", validator))
	}
	return k.bondValidator(ctx, validator)
}

func (k Keeper) unbondedToBonded(ctx sdk.Context, validator types.Validator) types.Validator {
	if validator.Status != sdk.Unbonded {
		panic(fmt.Sprintf("bad state transition unbondedToBonded, validator: %v\n", validator))
	}
	return k.bondValidator(ctx, validator)
}

func (k Keeper) unbondingToUnbonded(ctx sdk.Context, validator types.Validator) types.Validator {
	if validator.Status != sdk.Unbonded {
		panic(fmt.Sprintf("bad state transition unbondingToBonded, validator: %v\n", validator))
	}
	return k.completeUnbondingValidator(ctx, validator)
}

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
	// TODO we should be able to just delete the index, and only set it again once unjailed
}

// remove a validator from jail
func (k Keeper) UnjailValidator(ctx sdk.Context, validator types.Validator) {
	if !validator.Jailed {
		panic(fmt.Sprintf("cannot unjail already unjailed validator, validator: %v\n", validator))
	}

	pool := k.GetPool(ctx)
	k.DeleteValidatorByPowerIndex(ctx, validator, pool)
	validator.Jailed = false
	k.SetValidator(ctx, validator)
	k.SetValidatorByPowerIndex(ctx, validator, pool)
}

//________________________________________________________________________________________________

// perform all the store operations for when a validator status becomes bonded
func (k Keeper) bondValidator(ctx sdk.Context, validator types.Validator) types.Validator {

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

	return validator
}

// perform all the store operations for when a validator status begins unbonding
func (k Keeper) beginUnbondingValidator(ctx sdk.Context, validator types.Validator) types.Validator {

	store := ctx.KVStore(k.storeKey)
	pool := k.GetPool(ctx)
	params := k.GetParams(ctx)

	k.DeleteValidatorByPowerIndex(ctx, validator, pool)

	// sanity check
	if validator.Status != sdk.Bonded {
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
func (k Keeper) completeUnbondingValidator(ctx sdk.Context, validator types.Validator) types.Validator {
	pool := k.GetPool(ctx)
	validator, pool = validator.UpdateStatus(pool, sdk.Unbonded)
	k.SetPool(ctx, pool)
	k.SetValidator(ctx, validator)
	return validator
}

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
