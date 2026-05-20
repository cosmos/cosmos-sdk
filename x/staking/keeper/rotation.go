package keeper

import (
	"context"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// MaxConsKeyRotations is the maximum number of pending consensus key rotations
// a validator may have inside the unbonding window.
const MaxConsKeyRotations = 1

// HasPendingConsKeyRotation returns whether the validator has a pending
// consensus key rotation inside the unbonding window.
func (k Keeper) HasPendingConsKeyRotation(ctx context.Context, valAddr sdk.ValAddress) (bool, error) {
	return k.storeService.OpenKVStore(ctx).Has(types.GetValidatorConsKeyRotationKey(valAddr))
}

// HasRotatedConsAddr returns whether the given consensus address was previously
// rotated away from and is still inside its unbonding window.
func (k Keeper) HasRotatedConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (bool, error) {
	return k.storeService.OpenKVStore(ctx).Has(types.GetRotatedConsAddrIndexKey(consAddr))
}

// SetConsKeyRotation writes to indexes that track a pending consensus key
// rotation.
func (k Keeper) SetConsKeyRotation(ctx context.Context, valAddr sdk.ValAddress, oldPubKey cryptotypes.PubKey, fee sdk.Coin) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	unbondingTime, err := k.UnbondingTime(ctx)
	if err != nil {
		return err
	}
	maturity := sdkCtx.BlockHeader().Time.Add(unbondingTime)

	oldConsAddr := sdk.ConsAddress(oldPubKey.Address())

	store := k.storeService.OpenKVStore(ctx)
	if err := store.Set(types.GetConsKeyRotationQueueKey(maturity, valAddr), oldConsAddr); err != nil {
		return err
	}

	if err := store.Set(types.GetValidatorConsKeyRotationKey(valAddr), []byte{}); err != nil {
		return err
	}

	return store.Set(types.GetRotatedConsAddrIndexKey(oldConsAddr), valAddr)
}
