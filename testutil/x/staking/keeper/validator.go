package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/testutil/x/staking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

// RemoveValidatorTokens updates the tokens of an existing validator, updates the validators power index key
func (k Keeper) RemoveValidatorTokens(ctx context.Context,
	validator types.Validator, tokensToRemove math.Int,
) (types.Validator, error) {
	if err := k.DeleteValidatorByPowerIndex(ctx, validator); err != nil {
		return validator, err
	}

	validator = validator.RemoveTokens(tokensToRemove)
	if err := k.SetValidator(ctx, validator); err != nil {
		return validator, err
	}

	if err := k.SetValidatorByPowerIndex(ctx, validator); err != nil {
		return validator, err
	}

	return validator, nil
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
