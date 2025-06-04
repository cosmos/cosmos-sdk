package keeper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"

	abci "github.com/cometbft/cometbft/v2/abci/types"
	gogotypes "github.com/cosmos/gogoproto/types"

	"cosmossdk.io/core/address"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// BlockValidatorUpdates calculates the ValidatorUpdates for the current block
// Called in each EndBlock
func (k Keeper) BlockValidatorUpdates(ctx context.Context) ([]abci.ValidatorUpdate, error) {
	// Calculate validator set changes.
	//
	// NOTE: ApplyAndReturnValidatorSetUpdates has to come before
	// UnbondAllMatureValidatorQueue.
	// This fixes a bug when the unbonding period is instant (is the case in
	// some of the tests). The test expected the validator to be completely
	// unbonded after the Endblocker (go from Bonded -> Unbonding during
	// ApplyAndReturnValidatorSetUpdates and then Unbonding -> Unbonded during
	// UnbondAllMatureValidatorQueue).
	validatorUpdates, err := k.ApplyAndReturnValidatorSetUpdates(ctx)
	if err != nil {
		return nil, err
	}

	// unbond all mature validators from the unbonding queue
	err = k.UnbondAllMatureValidators(ctx)
	if err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Remove all mature unbonding delegations from the ubd queue.
	matureUnbonds, err := k.DequeueAllMatureUBDQueue(ctx, sdkCtx.BlockHeader().Time)
	if err != nil {
		return nil, err
	}

	for _, dvPair := range matureUnbonds {
		addr, err := k.validatorAddressCodec.StringToBytes(dvPair.ValidatorAddress)
		if err != nil {
			return nil, err
		}
		delegatorAddress, err := k.authKeeper.AddressCodec().StringToBytes(dvPair.DelegatorAddress)
		if err != nil {
			return nil, err
		}

		balances, err := k.CompleteUnbonding(ctx, delegatorAddress, addr)
		if err != nil {
			continue
		}

		sdkCtx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeCompleteUnbonding,
				sdk.NewAttribute(sdk.AttributeKeyAmount, balances.String()),
				sdk.NewAttribute(types.AttributeKeyValidator, dvPair.ValidatorAddress),
				sdk.NewAttribute(types.AttributeKeyDelegator, dvPair.DelegatorAddress),
			),
		)
	}

	// Remove all mature redelegations from the red queue.
	matureRedelegations, err := k.DequeueAllMatureRedelegationQueue(ctx, sdkCtx.BlockHeader().Time)
	if err != nil {
		return nil, err
	}

	for _, dvvTriplet := range matureRedelegations {
		valSrcAddr, err := k.validatorAddressCodec.StringToBytes(dvvTriplet.ValidatorSrcAddress)
		if err != nil {
			return nil, err
		}
		valDstAddr, err := k.validatorAddressCodec.StringToBytes(dvvTriplet.ValidatorDstAddress)
		if err != nil {
			return nil, err
		}
		delegatorAddress, err := k.authKeeper.AddressCodec().StringToBytes(dvvTriplet.DelegatorAddress)
		if err != nil {
			return nil, err
		}

		balances, err := k.CompleteRedelegation(
			ctx,
			delegatorAddress,
			valSrcAddr,
			valDstAddr,
		)
		if err != nil {
			continue
		}

		sdkCtx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeCompleteRedelegation,
				sdk.NewAttribute(sdk.AttributeKeyAmount, balances.String()),
				sdk.NewAttribute(types.AttributeKeyDelegator, dvvTriplet.DelegatorAddress),
				sdk.NewAttribute(types.AttributeKeySrcValidator, dvvTriplet.ValidatorSrcAddress),
				sdk.NewAttribute(types.AttributeKeyDstValidator, dvvTriplet.ValidatorDstAddress),
			),
		)
	}

	return validatorUpdates, nil
}

// ApplyAndReturnValidatorSetUpdates applies and return accumulated updates to the bonded validator set. Also,
// * Updates the active valset as keyed by LastValidatorPowerKey.
// * Updates the total power as keyed by LastTotalPowerKey.
// * Updates validator status' according to updated powers.
// * Updates the fee pool bonded vs not-bonded tokens.
// * Updates relevant indices.
// It gets called once after genesis, another time maybe after genesis transactions,
// then once at every EndBlock.
//
// CONTRACT: Only validators with non-zero power or zero-power that were bonded
// at the previous block height or were removed from the validator set entirely
// are returned to CometBFT.
func (k Keeper) ApplyAndReturnValidatorSetUpdates(ctx context.Context) (updates []abci.ValidatorUpdate, err error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	maxValidators := params.MaxValidators
	powerReduction := k.PowerReduction(ctx)
	totalPower := math.ZeroInt()
	amtFromBondedToNotBonded, amtFromNotBondedToBonded := math.ZeroInt(), math.ZeroInt()

	// Retrieve the last validator set.
	// The persistent set is updated later in this function.
	// (see LastValidatorPowerKey).
	last, err := k.getLastValidatorsByAddr(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get last validator set: %w", err)
	}

	// Iterate over validators, highest power to lowest.
	iterator, err := k.ValidatorsPowerStoreIterator(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get validators power store iterator: %w", err)
	}
	defer iterator.Close()

	for count := 0; iterator.Valid() && count < int(maxValidators); iterator.Next() {
		// everything that is iterated in this loop is becoming or already a
		// part of the bonded validator set
		valAddr := sdk.ValAddress(iterator.Value())
		validator, err := k.GetValidator(ctx, valAddr)
		if err != nil {
			return nil, fmt.Errorf("validator record not found for address: %X", valAddr)
		}

		if validator.Jailed {
			return nil, errors.New("should never retrieve a jailed validator from the power store")
		}

		// if we get to a zero-power validator (which we don't bond),
		// there are no more possible bonded validators
		if validator.PotentialConsensusPower(k.PowerReduction(ctx)) == 0 {
			break
		}

		// apply the appropriate state change if necessary
		switch {
		case validator.IsUnbonded():
			validator, err = k.unbondedToBonded(ctx, validator)
			if err != nil {
				return nil, err
			}
			amtFromNotBondedToBonded = amtFromNotBondedToBonded.Add(validator.GetTokens())
		case validator.IsUnbonding():
			validator, err = k.unbondingToBonded(ctx, validator)
			if err != nil {
				return nil, err
			}
			amtFromNotBondedToBonded = amtFromNotBondedToBonded.Add(validator.GetTokens())
		case validator.IsBonded():
			// no state change
		default:
			return nil, errors.New("unexpected validator status")
		}

		valAddrStr := string(valAddr)
		// fetch the old power bytes
		oldPower, found := last[valAddrStr]
		newPower := validator.ConsensusPower(powerReduction)

		// update the validator set if power has changed
		if !found || oldPower != newPower {
			updates = append(updates, validator.ABCIValidatorUpdate(powerReduction))

			if err = k.SetLastValidatorPower(ctx, valAddr, newPower); err != nil {
				return nil, err
			}
		}

		delete(last, valAddrStr)
		count++

		totalPower = totalPower.AddRaw(newPower)
	}

	noLongerBonded, err := sortNoLongerBonded(last, k.validatorAddressCodec)
	if err != nil {
		return nil, err
	}

	for _, valAddrBytes := range noLongerBonded {
		validator, err := k.GetValidator(ctx, sdk.ValAddress(valAddrBytes))
		if err != nil {
			return nil, fmt.Errorf("validator record not found for address: %X", sdk.ValAddress(valAddrBytes))
		}
		validator, err = k.bondedToUnbonding(ctx, validator)
		if err != nil {
			return nil, err
		}
		str, err := k.validatorAddressCodec.StringToBytes(validator.GetOperator())
		if err != nil {
			return nil, fmt.Errorf("failed to get validator operator address: %w", err)
		}
		amtFromBondedToNotBonded = amtFromBondedToNotBonded.Add(validator.GetTokens())
		if err = k.DeleteLastValidatorPower(ctx, str); err != nil {
			return nil, err
		}

		updates = append(updates, validator.ABCIValidatorUpdateZero())
	}

	// Update the pools based on the recent updates in the validator set:
	// - The tokens from the non-bonded candidates that enter the new validator set need to be transferred
	// to the Bonded pool.
	// - The tokens from the bonded validators that are being kicked out from the validator set
	// need to be transferred to the NotBonded pool.
	switch {
	// Compare and subtract the respective amounts to only perform one transfer.
	// This is done in order to avoid doing multiple updates inside each iterator/loop.
	case amtFromNotBondedToBonded.GT(amtFromBondedToNotBonded):
		if err = k.notBondedTokensToBonded(ctx, amtFromNotBondedToBonded.Sub(amtFromBondedToNotBonded)); err != nil {
			return nil, err
		}
	case amtFromNotBondedToBonded.LT(amtFromBondedToNotBonded):
		if err = k.bondedTokensToNotBonded(ctx, amtFromBondedToNotBonded.Sub(amtFromNotBondedToBonded)); err != nil {
			return nil, err
		}
	default: // equal amounts of tokens; no update required
	}

	// set total power on lookup index if there are any updates
	if len(updates) > 0 {
		if err = k.SetLastTotalPower(ctx, totalPower); err != nil {
			return nil, err
		}
	}

	// set the list of validator updates
	if err = k.SetValidatorUpdates(ctx, updates); err != nil {
		return nil, err
	}

	return updates, err
}

// Validator state transitions

func (k Keeper) bondedToUnbonding(ctx context.Context, validator types.Validator) (types.Validator, error) {
	if !validator.IsBonded() {
		return types.Validator{}, fmt.Errorf("bad state transition bondedToUnbonding, validator: %v", validator)
	}

	return k.BeginUnbondingValidator(ctx, validator)
}

func (k Keeper) unbondingToBonded(ctx context.Context, validator types.Validator) (types.Validator, error) {
	if !validator.IsUnbonding() {
		return types.Validator{}, fmt.Errorf("bad state transition unbondingToBonded, validator: %v", validator)
	}

	return k.bondValidator(ctx, validator)
}

func (k Keeper) unbondedToBonded(ctx context.Context, validator types.Validator) (types.Validator, error) {
	if !validator.IsUnbonded() {
		return types.Validator{}, fmt.Errorf("bad state transition unbondedToBonded, validator: %v", validator)
	}

	return k.bondValidator(ctx, validator)
}

// UnbondingToUnbonded switches a validator from unbonding state to unbonded state
func (k Keeper) UnbondingToUnbonded(ctx context.Context, validator types.Validator) (types.Validator, error) {
	if !validator.IsUnbonding() {
		return types.Validator{}, fmt.Errorf("bad state transition unbondingToUnbonded, validator: %v", validator)
	}

	return k.completeUnbondingValidator(ctx, validator)
}

// send a validator to jail
func (k Keeper) jailValidator(ctx context.Context, validator types.Validator) error {
	if validator.Jailed {
		return types.ErrValidatorJailed.Wrapf("cannot jail already jailed validator, validator: %v", validator)
	}

	validator.Jailed = true
	if err := k.SetValidator(ctx, validator); err != nil {
		return err
	}

	return k.DeleteValidatorByPowerIndex(ctx, validator)
}

// remove a validator from jail
func (k Keeper) unjailValidator(ctx context.Context, validator types.Validator) error {
	if !validator.Jailed {
		return fmt.Errorf("cannot unjail already unjailed validator, validator: %v", validator)
	}

	validator.Jailed = false
	if err := k.SetValidator(ctx, validator); err != nil {
		return err
	}

	return k.SetValidatorByPowerIndex(ctx, validator)
}

// perform all the store operations for when a validator status becomes bonded
func (k Keeper) bondValidator(ctx context.Context, validator types.Validator) (types.Validator, error) {
	// delete the validator by power index, as the key will change
	if err := k.DeleteValidatorByPowerIndex(ctx, validator); err != nil {
		return types.Validator{}, err
	}

	validator = validator.UpdateStatus(types.Bonded)

	// save the now bonded validator record to the two referenced stores
	if err := k.SetValidator(ctx, validator); err != nil {
		return types.Validator{}, err
	}

	if err := k.SetValidatorByPowerIndex(ctx, validator); err != nil {
		return types.Validator{}, err
	}

	// delete from queue if present
	if err := k.DeleteValidatorQueue(ctx, validator); err != nil {
		return types.Validator{}, err
	}

	// trigger hook
	consAddr, err := validator.GetConsAddr()
	if err != nil {
		return types.Validator{}, err
	}

	str, err := k.validatorAddressCodec.StringToBytes(validator.GetOperator())
	if err != nil {
		return types.Validator{}, fmt.Errorf("failed to get validator operator address: %w", err)
	}

	if err := k.Hooks().AfterValidatorBonded(ctx, consAddr, str); err != nil {
		return types.Validator{}, err
	}

	return validator, nil
}

// BeginUnbondingValidator performs all the store operations for when a validator begins unbonding
func (k Keeper) BeginUnbondingValidator(ctx context.Context, validator types.Validator) (types.Validator, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return types.Validator{}, err
	}

	// delete the validator by power index, as the key will change
	if err = k.DeleteValidatorByPowerIndex(ctx, validator); err != nil {
		return types.Validator{}, err
	}

	// sanity check
	if validator.Status != types.Bonded {
		return types.Validator{}, fmt.Errorf("should not already be unbonded or unbonding, validator: %v", validator)
	}

	id, err := k.IncrementUnbondingID(ctx)
	if err != nil {
		return types.Validator{}, err
	}

	validator = validator.UpdateStatus(types.Unbonding)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// set the unbonding completion time and completion height appropriately
	validator.UnbondingTime = sdkCtx.BlockHeader().Time.Add(params.UnbondingTime)
	validator.UnbondingHeight = sdkCtx.BlockHeader().Height

	validator.UnbondingIds = append(validator.UnbondingIds, id)

	// save the now unbonded validator record and power index
	if err = k.SetValidator(ctx, validator); err != nil {
		return types.Validator{}, err
	}

	if err = k.SetValidatorByPowerIndex(ctx, validator); err != nil {
		return types.Validator{}, err
	}

	// Adds to unbonding validator queue
	if err = k.InsertUnbondingValidatorQueue(ctx, validator); err != nil {
		return types.Validator{}, err
	}

	// trigger hook
	consAddr, err := validator.GetConsAddr()
	if err != nil {
		return types.Validator{}, err
	}

	str, err := k.validatorAddressCodec.StringToBytes(validator.GetOperator())
	if err != nil {
		return types.Validator{}, fmt.Errorf("failed to get validator operator address: %w", err)
	}

	if err := k.Hooks().AfterValidatorBeginUnbonding(ctx, consAddr, str); err != nil {
		return types.Validator{}, err
	}

	if err := k.SetValidatorByUnbondingID(ctx, validator, id); err != nil {
		return types.Validator{}, err
	}

	if err := k.Hooks().AfterUnbondingInitiated(ctx, id); err != nil {
		return types.Validator{}, err
	}

	return validator, nil
}

// perform all the store operations for when a validator status becomes unbonded
func (k Keeper) completeUnbondingValidator(ctx context.Context, validator types.Validator) (types.Validator, error) {
	validator = validator.UpdateStatus(types.Unbonded)
	if err := k.SetValidator(ctx, validator); err != nil {
		return types.Validator{}, err
	}

	return validator, nil
}

// map of operator addresses to power
// We use (non bech32) strings here, because we can't have slices as keys: map[[]byte][]byte
type validatorsByAddr map[string]int64

// get the last validator set
func (k Keeper) getLastValidatorsByAddr(ctx context.Context) (validatorsByAddr, error) {
	last := make(validatorsByAddr)

	iterator, err := k.LastValidatorsIterator(ctx)
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var intVal gogotypes.Int64Value
	for ; iterator.Valid(); iterator.Next() {
		// extract the validator address from the key (prefix is 1-byte, addrLen is 1-byte)
		valAddrStr := string(types.AddressFromLastValidatorPowerKey(iterator.Key()))
		k.cdc.MustUnmarshal(iterator.Value(), &intVal)
		last[valAddrStr] = intVal.GetValue()
	}

	return last, nil
}

// given a map of remaining validators to previous bonded power
// returns the list of validators to be unbonded, sorted by operator address
func sortNoLongerBonded(last validatorsByAddr, ac address.Codec) ([][]byte, error) {
	// sort the map keys for determinism
	noLongerBonded := make([][]byte, len(last))
	index := 0

	for valAddrStr := range last {
		valAddrBytes := []byte(valAddrStr)
		noLongerBonded[index] = valAddrBytes
		index++
	}
	// sorted by address - order doesn't matter
	sort.SliceStable(noLongerBonded, func(i, j int) bool {
		// -1 means strictly less than
		return bytes.Compare(noLongerBonded[i], noLongerBonded[j]) == -1
	})

	return noLongerBonded, nil
}
