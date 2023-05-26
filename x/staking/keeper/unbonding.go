package keeper

import (
	"context"
	"encoding/binary"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// IncrementUnbondingID increments and returns a unique ID for an unbonding operation
func (k Keeper) IncrementUnbondingID(ctx context.Context) (unbondingID uint64, err error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.UnbondingIDKey)
	if err != nil {
		return 0, err
	}

	if bz != nil {
		unbondingID = binary.BigEndian.Uint64(bz)
	}

	unbondingID++

	// Convert back into bytes for storage
	bz = make([]byte, 8)
	binary.BigEndian.PutUint64(bz, unbondingID)

	if err = store.Set(types.UnbondingIDKey, bz); err != nil {
		return 0, err
	}

	return unbondingID, err
}

// DeleteUnbondingIndex removes a mapping from UnbondingId to unbonding operation
func (k Keeper) DeleteUnbondingIndex(ctx context.Context, id uint64) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Delete(types.GetUnbondingIndexKey(id))
}

func (k Keeper) GetUnbondingType(ctx context.Context, id uint64) (unbondingType types.UnbondingType, err error) {
	store := k.storeService.OpenKVStore(ctx)

	bz, err := store.Get(types.GetUnbondingTypeKey(id))
	if err != nil {
		return unbondingType, err
	}

	if bz == nil {
		return unbondingType, types.ErrNoUnbondingType
	}

	return types.UnbondingType(binary.BigEndian.Uint64(bz)), nils
}

func (k Keeper) SetUnbondingType(ctx context.Context, id uint64, unbondingType types.UnbondingType) error {
	store := k.storeService.OpenKVStore(ctx)

	// Convert into bytes for storage
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(unbondingType))

	return store.Set(types.GetUnbondingTypeKey(id), bz)
}

// GetUnbondingDelegationByUnbondingID returns a unbonding delegation that has an unbonding delegation entry with a certain ID
func (k Keeper) GetUnbondingDelegationByUnbondingID(ctx context.Context, id uint64) (ubd types.UnbondingDelegation, err error) {
	store := k.storeService.OpenKVStore(ctx)

	ubdKey, err := store.Get(types.GetUnbondingIndexKey(id))
	if err != nil {
		return types.UnbondingDelegation{}, err
	}

	if ubdKey == nil {
		return types.UnbondingDelegation{}, types.ErrNoUnbondingDelegation
	}

	value, err := store.Get(ubdKey)
	if err != nil {
		return types.UnbondingDelegation{}, err
	}

	if value == nil {
		return types.UnbondingDelegation{}, types.ErrNoUnbondingDelegation
	}

	ubd, err = types.UnmarshalUBD(k.cdc, value)
	// An error here means that what we got wasn't the right type
	if err != nil {
		return types.UnbondingDelegation{}, err
	}

	return ubd, nil
}

// GetRedelegationByUnbondingID returns a unbonding delegation that has an unbonding delegation entry with a certain ID
func (k Keeper) GetRedelegationByUnbondingID(ctx context.Context, id uint64) (red types.Redelegation, err error) {
	store := k.storeService.OpenKVStore(ctx)

	redKey, err := store.Get(types.GetUnbondingIndexKey(id))
	if err != nil {
		return types.Redelegation{}, err
	}

	if redKey == nil {
		return types.Redelegation{}, types.ErrNoRedelegation
	}

	value, err := store.Get(redKey)
	if err != nil {
		return types.Redelegation{}, err
	}

	if value == nil {
		return types.Redelegation{}, types.ErrNoRedelegation
	}

	red, err = types.UnmarshalRED(k.cdc, value)
	// An error here means that what we got wasn't the right type
	if err != nil {
		return types.Redelegation{}, err
	}

	return red, nil
}

// GetValidatorByUnbondingID returns the validator that is unbonding with a certain unbonding op ID
func (k Keeper) GetValidatorByUnbondingID(ctx context.Context, id uint64) (val types.Validator, err error) {
	store := k.storeService.OpenKVStore(ctx)

	valKey, err := store.Get(types.GetUnbondingIndexKey(id))
	if err != nil {
		return types.Validator{}, err
	}

	if valKey == nil {
		return types.Validator{}, types.ErrNoValidatorFound
	}

	value, err := store.Get(valKey)
	if err != nil {
		return types.Validator{}, err
	}

	if value == nil {
		return types.Validator{}, types.ErrNoValidatorFound
	}

	val, err = types.UnmarshalValidator(k.cdc, value)
	// An error here means that what we got wasn't the right type
	if err != nil {
		return types.Validator{}, err
	}

	return val, nil
}

// SetUnbondingDelegationByUnbondingID sets an index to look up an UnbondingDelegation by the unbondingID of an UnbondingDelegationEntry that it contains
// Note, it does not set the unbonding delegation itself, use SetUnbondingDelegation(ctx, ubd) for that
func (k Keeper) SetUnbondingDelegationByUnbondingID(ctx context.Context, ubd types.UnbondingDelegation, id uint64) error {
	store := k.storeService.OpenKVStore(ctx)
	delAddr, err := k.authKeeper.AddressCodec().StringToBytes(ubd.DelegatorAddress)
	if err != nil {
		return err
	}
	valAddr, err := sdk.ValAddressFromBech32(ubd.ValidatorAddress)
	if err != nil {
		return err
	}

	ubdKey := types.GetUBDKey(delAddr, valAddr)
	if err = store.Set(types.GetUnbondingIndexKey(id), ubdKey); err != nil {
		return err
	}

	// Set unbonding type so that we know how to deserialize it later
	return k.SetUnbondingType(ctx, id, types.UnbondingType_UnbondingDelegation)
}

// SetRedelegationByUnbondingID sets an index to look up an Redelegation by the unbondingID of an RedelegationEntry that it contains
// Note, it does not set the redelegation itself, use SetRedelegation(ctx, red) for that
func (k Keeper) SetRedelegationByUnbondingID(ctx context.Context, red types.Redelegation, id uint64) error {
	store := k.storeService.OpenKVStore(ctx)

	delAddr, err := k.authKeeper.AddressCodec().StringToBytes(red.DelegatorAddress)
	if err != nil {
		panic(err)
	}

	valSrcAddr, err := sdk.ValAddressFromBech32(red.ValidatorSrcAddress)
	if err != nil {
		panic(err)
	}

	valDstAddr, err := sdk.ValAddressFromBech32(red.ValidatorDstAddress)
	if err != nil {
		panic(err)
	}

	redKey := types.GetREDKey(delAddr, valSrcAddr, valDstAddr)
	store.Set(types.GetUnbondingIndexKey(id), redKey)

	// Set unbonding type so that we know how to deserialize it later
	k.SetUnbondingType(ctx, id, types.UnbondingType_Redelegation)
}

// SetValidatorByUnbondingID sets an index to look up a Validator by the unbondingID corresponding to its current unbonding
// Note, it does not set the validator itself, use SetValidator(ctx, val) for that
func (k Keeper) SetValidatorByUnbondingID(ctx sdk.Context, val types.Validator, id uint64) {
	store := ctx.KVStore(k.storeKey)

	valAddr, err := sdk.ValAddressFromBech32(val.OperatorAddress)
	if err != nil {
		panic(err)
	}

	valKey := types.GetValidatorKey(valAddr)
	store.Set(types.GetUnbondingIndexKey(id), valKey)

	// Set unbonding type so that we know how to deserialize it later
	k.SetUnbondingType(ctx, id, types.UnbondingType_ValidatorUnbonding)
}

// unbondingDelegationEntryArrayIndex and redelegationEntryArrayIndex are utilities to find
// at which position in the Entries array the entry with a given id is
func unbondingDelegationEntryArrayIndex(ubd types.UnbondingDelegation, id uint64) (index int, found bool) {
	for i, entry := range ubd.Entries {
		// we find the entry with the right ID
		if entry.UnbondingId == id {
			return i, true
		}
	}

	return 0, false
}

func redelegationEntryArrayIndex(red types.Redelegation, id uint64) (index int, found bool) {
	for i, entry := range red.Entries {
		// we find the entry with the right ID
		if entry.UnbondingId == id {
			return i, true
		}
	}

	return 0, false
}

// UnbondingCanComplete allows a stopped unbonding operation, such as an
// unbonding delegation, a redelegation, or a validator unbonding to complete.
// In order for the unbonding operation with `id` to eventually complete, every call
// to PutUnbondingOnHold(id) must be matched by a call to UnbondingCanComplete(id).
func (k Keeper) UnbondingCanComplete(ctx sdk.Context, id uint64) error {
	unbondingType, found := k.GetUnbondingType(ctx, id)
	if !found {
		return types.ErrUnbondingNotFound
	}

	switch unbondingType {
	case types.UnbondingType_UnbondingDelegation:
		if err := k.unbondingDelegationEntryCanComplete(ctx, id); err != nil {
			return err
		}
	case types.UnbondingType_Redelegation:
		if err := k.redelegationEntryCanComplete(ctx, id); err != nil {
			return err
		}
	case types.UnbondingType_ValidatorUnbonding:
		if err := k.validatorUnbondingCanComplete(ctx, id); err != nil {
			return err
		}
	default:
		return types.ErrUnbondingNotFound
	}

	return nil
}

func (k Keeper) unbondingDelegationEntryCanComplete(ctx sdk.Context, id uint64) error {
	ubd, found := k.GetUnbondingDelegationByUnbondingID(ctx, id)
	if !found {
		return types.ErrUnbondingNotFound
	}

	i, found := unbondingDelegationEntryArrayIndex(ubd, id)
	if !found {
		return types.ErrUnbondingNotFound
	}

	// The entry must be on hold
	if !ubd.Entries[i].OnHold() {
		return errorsmod.Wrapf(
			types.ErrUnbondingOnHoldRefCountNegative,
			"undelegation unbondingID(%d), expecting UnbondingOnHoldRefCount > 0, got %T",
			id, ubd.Entries[i].UnbondingOnHoldRefCount,
		)
	}
	ubd.Entries[i].UnbondingOnHoldRefCount--

	// Check if entry is matured.
	if !ubd.Entries[i].OnHold() && ubd.Entries[i].IsMature(ctx.BlockHeader().Time) {
		// If matured, complete it.
		delegatorAddress, err := k.authKeeper.AddressCodec().StringToBytes(ubd.DelegatorAddress)
		if err != nil {
			return err
		}

		bondDenom := k.GetParams(ctx).BondDenom

		// track undelegation only when remaining or truncated shares are non-zero
		if !ubd.Entries[i].Balance.IsZero() {
			amt := sdk.NewCoin(bondDenom, ubd.Entries[i].Balance)
			if err := k.bankKeeper.UndelegateCoinsFromModuleToAccount(
				ctx, types.NotBondedPoolName, delegatorAddress, sdk.NewCoins(amt),
			); err != nil {
				return err
			}
		}

		// Remove entry
		ubd.RemoveEntry(int64(i))
		// Remove from the UnbondingIndex
		k.DeleteUnbondingIndex(ctx, id)
	}

	// set the unbonding delegation or remove it if there are no more entries
	if len(ubd.Entries) == 0 {
		k.RemoveUnbondingDelegation(ctx, ubd)
	} else {
		k.SetUnbondingDelegation(ctx, ubd)
	}

	// Successfully completed unbonding
	return nil
}

func (k Keeper) redelegationEntryCanComplete(ctx sdk.Context, id uint64) error {
	red, found := k.GetRedelegationByUnbondingID(ctx, id)
	if !found {
		return types.ErrUnbondingNotFound
	}

	i, found := redelegationEntryArrayIndex(red, id)
	if !found {
		return types.ErrUnbondingNotFound
	}

	// The entry must be on hold
	if !red.Entries[i].OnHold() {
		return errorsmod.Wrapf(
			types.ErrUnbondingOnHoldRefCountNegative,
			"redelegation unbondingID(%d), expecting UnbondingOnHoldRefCount > 0, got %T",
			id, red.Entries[i].UnbondingOnHoldRefCount,
		)
	}
	red.Entries[i].UnbondingOnHoldRefCount--

	if !red.Entries[i].OnHold() && red.Entries[i].IsMature(ctx.BlockHeader().Time) {
		// If matured, complete it.
		// Remove entry
		red.RemoveEntry(int64(i))
		// Remove from the Unbonding index
		k.DeleteUnbondingIndex(ctx, id)
	}

	// set the redelegation or remove it if there are no more entries
	if len(red.Entries) == 0 {
		k.RemoveRedelegation(ctx, red)
	} else {
		k.SetRedelegation(ctx, red)
	}

	// Successfully completed unbonding
	return nil
}

func (k Keeper) validatorUnbondingCanComplete(ctx sdk.Context, id uint64) error {
	val, found := k.GetValidatorByUnbondingID(ctx, id)
	if !found {
		return types.ErrUnbondingNotFound
	}

	if val.UnbondingOnHoldRefCount <= 0 {
		return errorsmod.Wrapf(
			types.ErrUnbondingOnHoldRefCountNegative,
			"val(%s), expecting UnbondingOnHoldRefCount > 0, got %T",
			val.OperatorAddress, val.UnbondingOnHoldRefCount,
		)
	}
	val.UnbondingOnHoldRefCount--
	k.SetValidator(ctx, val)

	return nil
}

// PutUnbondingOnHold allows an external module to stop an unbonding operation,
// such as an unbonding delegation, a redelegation, or a validator unbonding.
// In order for the unbonding operation with `id` to eventually complete, every call
// to PutUnbondingOnHold(id) must be matched by a call to UnbondingCanComplete(id).
func (k Keeper) PutUnbondingOnHold(ctx sdk.Context, id uint64) error {
	unbondingType, found := k.GetUnbondingType(ctx, id)
	if !found {
		return types.ErrUnbondingNotFound
	}
	switch unbondingType {
	case types.UnbondingType_UnbondingDelegation:
		if err := k.putUnbondingDelegationEntryOnHold(ctx, id); err != nil {
			return err
		}
	case types.UnbondingType_Redelegation:
		if err := k.putRedelegationEntryOnHold(ctx, id); err != nil {
			return err
		}
	case types.UnbondingType_ValidatorUnbonding:
		if err := k.putValidatorOnHold(ctx, id); err != nil {
			return err
		}
	default:
		return types.ErrUnbondingNotFound
	}

	return nil
}

func (k Keeper) putUnbondingDelegationEntryOnHold(ctx sdk.Context, id uint64) error {
	ubd, found := k.GetUnbondingDelegationByUnbondingID(ctx, id)
	if !found {
		return types.ErrUnbondingNotFound
	}

	i, found := unbondingDelegationEntryArrayIndex(ubd, id)
	if !found {
		return types.ErrUnbondingNotFound
	}

	ubd.Entries[i].UnbondingOnHoldRefCount++
	k.SetUnbondingDelegation(ctx, ubd)

	return nil
}

func (k Keeper) putRedelegationEntryOnHold(ctx sdk.Context, id uint64) error {
	red, found := k.GetRedelegationByUnbondingID(ctx, id)
	if !found {
		return types.ErrUnbondingNotFound
	}

	i, found := redelegationEntryArrayIndex(red, id)
	if !found {
		return types.ErrUnbondingNotFound
	}

	red.Entries[i].UnbondingOnHoldRefCount++
	k.SetRedelegation(ctx, red)

	return nil
}

func (k Keeper) putValidatorOnHold(ctx sdk.Context, id uint64) error {
	val, found := k.GetValidatorByUnbondingID(ctx, id)
	if !found {
		return types.ErrUnbondingNotFound
	}

	val.UnbondingOnHoldRefCount++
	k.SetValidator(ctx, val)

	return nil
}
