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

// Validator state transitions

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
