package keeper

import (
	"context"
	"errors"
	"fmt"
	"time"

	gogotypes "github.com/cosmos/gogoproto/types"

	"cosmossdk.io/collections"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/testutil/x/staking/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var timeBzKeySize = uint64(29) // time bytes key size is 29 by default

// GetValidator gets a single validator
func (k Keeper) GetValidator(ctx context.Context, addr sdk.ValAddress) (validator types.Validator, err error) {
	validator, err = k.Validators.Get(ctx, addr)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.Validator{}, types.ErrNoValidatorFound
		}
		return validator, err
	}
	return validator, nil
}

// GetValidatorByConsAddr gets a single validator by consensus address
func (k Keeper) GetValidatorByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (validator types.Validator, err error) {
	opAddr, err := k.ValidatorByConsensusAddress.Get(ctx, consAddr)
	if err != nil {
		return types.Validator{}, types.ErrNoValidatorFound
	}

	return k.GetValidator(ctx, opAddr)
}

// SetValidator sets the main record holding validator details
func (k Keeper) SetValidator(ctx context.Context, validator types.Validator) error {
	valBz, err := k.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
	if err != nil {
		return err
	}
	return k.Validators.Set(ctx, sdk.ValAddress(valBz), validator)
}

// SetValidatorByConsAddr sets a validator by conesensus address
func (k Keeper) SetValidatorByConsAddr(ctx context.Context, validator types.Validator) error {
	consPk, err := validator.GetConsAddr()
	if err != nil {
		return err
	}

	bz, err := k.validatorAddressCodec.StringToBytes(validator.GetOperator())
	if err != nil {
		return err
	}

	return k.ValidatorByConsensusAddress.Set(ctx, consPk, bz)
}

// SetValidatorByPowerIndex sets a validator by power index
func (k Keeper) SetValidatorByPowerIndex(ctx context.Context, validator types.Validator) error {
	// jailed validators are not kept in the power index
	if validator.Jailed {
		return nil
	}

	store := k.KVStoreService.OpenKVStore(ctx)
	str, err := k.validatorAddressCodec.StringToBytes(validator.GetOperator())
	if err != nil {
		return err
	}
	return store.Set(types.GetValidatorsByPowerIndexKey(validator, k.PowerReduction(ctx), k.validatorAddressCodec), str)
}

// DeleteValidatorByPowerIndex deletes a record by power index
func (k Keeper) DeleteValidatorByPowerIndex(ctx context.Context, validator types.Validator) error {
	store := k.KVStoreService.OpenKVStore(ctx)
	return store.Delete(types.GetValidatorsByPowerIndexKey(validator, k.PowerReduction(ctx), k.validatorAddressCodec))
}

// SetNewValidatorByPowerIndex adds new entry by power index
func (k Keeper) SetNewValidatorByPowerIndex(ctx context.Context, validator types.Validator) error {
	store := k.KVStoreService.OpenKVStore(ctx)
	str, err := k.validatorAddressCodec.StringToBytes(validator.GetOperator())
	if err != nil {
		return err
	}
	return store.Set(types.GetValidatorsByPowerIndexKey(validator, k.PowerReduction(ctx), k.validatorAddressCodec), str)
}

// AddValidatorTokensAndShares updates the tokens of an existing validator, updates the validators power index key
func (k Keeper) AddValidatorTokensAndShares(ctx context.Context, validator types.Validator,
	tokensToAdd math.Int,
) (valOut types.Validator, addedShares math.LegacyDec, err error) {
	err = k.DeleteValidatorByPowerIndex(ctx, validator)
	if err != nil {
		return valOut, addedShares, err
	}

	validator, addedShares = validator.AddTokensFromDel(tokensToAdd)
	err = k.SetValidator(ctx, validator)
	if err != nil {
		return validator, addedShares, err
	}

	err = k.SetValidatorByPowerIndex(ctx, validator)
	return validator, addedShares, err
}

// get groups of validators

// GetAllValidators gets the set of all validators with no limits, used during genesis dump
func (k Keeper) GetAllValidators(ctx context.Context) (validators []types.Validator, err error) {
	store := k.KVStoreService.OpenKVStore(ctx)

	iterator, err := store.Iterator(types.ValidatorsKey, storetypes.PrefixEndBytes(types.ValidatorsKey))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		validator, err := types.UnmarshalValidator(k.cdc, iterator.Value())
		if err != nil {
			return nil, err
		}
		validators = append(validators, validator)
	}

	return validators, nil
}

// GetValidators returns a given amount of all the validators
func (k Keeper) GetValidators(ctx context.Context, maxRetrieve uint32) (validators []types.Validator, err error) {
	store := k.KVStoreService.OpenKVStore(ctx)
	validators = make([]types.Validator, maxRetrieve)

	iterator, err := store.Iterator(types.ValidatorsKey, storetypes.PrefixEndBytes(types.ValidatorsKey))
	if err != nil {
		return nil, err
	}

	i := 0
	for ; iterator.Valid() && i < int(maxRetrieve); iterator.Next() {
		validator, err := types.UnmarshalValidator(k.cdc, iterator.Value())
		if err != nil {
			return nil, err
		}
		validators[i] = validator
		i++
	}

	return validators[:i], nil // trim if the array length < maxRetrieve
}

// GetBondedValidatorsByPower gets the current group of bonded validators sorted by power-rank
func (k Keeper) GetBondedValidatorsByPower(ctx context.Context) ([]types.Validator, error) {
	maxValidators, err := k.MaxValidators(ctx)
	if err != nil {
		return nil, err
	}
	validators := make([]types.Validator, maxValidators)

	iterator, err := k.ValidatorsPowerStoreIterator(ctx)
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	i := 0
	for ; iterator.Valid() && i < int(maxValidators); iterator.Next() {
		address := iterator.Value()
		validator, err := k.GetValidator(ctx, address)
		if err != nil {
			return nil, fmt.Errorf("validator record not found for address: %X", address)
		}

		if validator.IsBonded() {
			validators[i] = validator
			i++
		}
	}

	return validators[:i], nil // trim
}

// ValidatorsPowerStoreIterator returns an iterator for the current validator power store
func (k Keeper) ValidatorsPowerStoreIterator(ctx context.Context) (corestore.Iterator, error) {
	store := k.KVStoreService.OpenKVStore(ctx)
	return store.ReverseIterator(types.ValidatorsByPowerIndexKey, storetypes.PrefixEndBytes(types.ValidatorsByPowerIndexKey))
}

// Last Validator Index

// GetLastValidatorPower loads the last validator power.
// Returns zero if the operator was not a validator last block.
func (k Keeper) GetLastValidatorPower(ctx context.Context, operator sdk.ValAddress) (power int64, err error) {
	intV, err := k.LastValidatorPower.Get(ctx, operator)
	return intV.GetValue(), err
}

// SetLastValidatorPower sets the last validator power.
func (k Keeper) SetLastValidatorPower(ctx context.Context, operator sdk.ValAddress, power int64) error {
	return k.LastValidatorPower.Set(ctx, operator, gogotypes.Int64Value{Value: power})
}

// DeleteLastValidatorPower deletes the last validator power.
func (k Keeper) DeleteLastValidatorPower(ctx context.Context, operator sdk.ValAddress) error {
	return k.LastValidatorPower.Remove(ctx, operator)
}

// IterateLastValidatorPowers iterates over last validator powers.
func (k Keeper) IterateLastValidatorPowers(ctx context.Context, handler func(operator sdk.ValAddress, power int64) (stop bool)) error {
	err := k.LastValidatorPower.Walk(ctx, nil, func(key []byte, value gogotypes.Int64Value) (bool, error) {
		addr := sdk.ValAddress(key)

		if handler(addr, value.GetValue()) {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// GetLastValidators gets the group of the bonded validators
func (k Keeper) GetLastValidators(ctx context.Context) (validators []types.Validator, err error) {
	// add the actual validator power sorted store
	maxValidators, err := k.MaxValidators(ctx)
	if err != nil {
		return nil, err
	}

	i := 0
	validators = make([]types.Validator, maxValidators)

	err = k.LastValidatorPower.Walk(ctx, nil, func(key []byte, _ gogotypes.Int64Value) (bool, error) {
		// Note, we do NOT error here as the MaxValidators param may change via on-chain
		// governance. In cases where the param is increased, this case should never
		// be hit. In cases where the param is decreased, we will simply not return
		// the remainder of the validator set, as the ApplyAndReturnValidatorSetUpdates
		// call should ensure the validators past the cliff will be moved to the
		// unbonding set.
		if i >= int(maxValidators) {
			return true, nil
		}

		validator, err := k.GetValidator(ctx, key)
		if err != nil {
			return true, err
		}

		validators[i] = validator
		i++

		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return validators[:i], nil // trim
}

// DeleteValidatorQueueTimeSlice deletes all entries in the queue indexed by a
// given height and time.
func (k Keeper) DeleteValidatorQueueTimeSlice(ctx context.Context, endTime time.Time, endHeight int64) error {
	return k.ValidatorQueue.Remove(ctx, collections.Join3(timeBzKeySize, endTime, uint64(endHeight)))
}

// IsValidatorJailed checks and returns boolean of a validator status jailed or not.
func (k Keeper) IsValidatorJailed(ctx context.Context, addr sdk.ConsAddress) (bool, error) {
	v, err := k.GetValidatorByConsAddr(ctx, addr)
	if err != nil {
		return false, err
	}

	return v.Jailed, nil
}

// GetPubKeyByConsAddr returns the consensus public key by consensus address.
// Caller receives a Cosmos SDK Pubkey type and must cast it to a comet type
func (k Keeper) GetPubKeyByConsAddr(ctx context.Context, addr sdk.ConsAddress) (cryptotypes.PubKey, error) {
	v, err := k.GetValidatorByConsAddr(ctx, addr)
	if err != nil {
		return nil, err
	}

	pubkey, err := v.ConsPubKey()
	if err != nil {
		return nil, err
	}

	return pubkey, nil
}
