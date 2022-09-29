package keeper

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Increments and returns a unique ID for an unbonding operation
func (k Keeper) IncrementUnbondingId(ctx sdk.Context) (unbondingId uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.UnbondingIdKey)

	if bz == nil {
		unbondingId = 0
	} else {
		unbondingId = binary.BigEndian.Uint64(bz)
	}

	unbondingId = unbondingId + 1

	// Convert back into bytes for storage
	bz = make([]byte, 8)
	binary.BigEndian.PutUint64(bz, unbondingId)

	store.Set(types.UnbondingIdKey, bz)

	return unbondingId
}

// Remove a mapping from UnbondingId to unbonding operation
func (k Keeper) DeleteUnbondingIndex(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetUnbondingIndexKey(id))
}

func (k Keeper) GetUnbondingType(ctx sdk.Context, id uint64) (unbondingType types.UnbondingType, found bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetUnbondingTypeKey(id))
	if bz == nil {
		return unbondingType, false
	}

	return types.UnbondingType(binary.BigEndian.Uint64(bz)), true
}

func (k Keeper) SetUnbondingType(ctx sdk.Context, id uint64, unbondingType types.UnbondingType) {
	store := ctx.KVStore(k.storeKey)

	// Convert into bytes for storage
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(unbondingType))

	store.Set(types.GetUnbondingTypeKey(id), bz)
}

// return a unbonding delegation that has an unbonding delegation entry with a certain ID
func (k Keeper) GetUnbondingDelegationByUnbondingId(
	ctx sdk.Context, id uint64,
) (ubd types.UnbondingDelegation, found bool) {
	store := ctx.KVStore(k.storeKey)

	ubdeKey := store.Get(types.GetUnbondingIndexKey(id))
	if ubdeKey == nil {
		return types.UnbondingDelegation{}, false
	}

	value := store.Get(ubdeKey)
	if value == nil {
		return types.UnbondingDelegation{}, false
	}

	ubd, err := types.UnmarshalUBD(k.cdc, value)
	// An error here means that what we got wasn't the right type
	if err != nil {
		return types.UnbondingDelegation{}, false
	}

	return ubd, true
}

// return a unbonding delegation that has an unbonding delegation entry with a certain ID
func (k Keeper) GetRedelegationByUnbondingId(
	ctx sdk.Context, id uint64,
) (red types.Redelegation, found bool) {
	store := ctx.KVStore(k.storeKey)

	redKey := store.Get(types.GetUnbondingIndexKey(id))
	if redKey == nil {
		return types.Redelegation{}, false
	}

	value := store.Get(redKey)
	if value == nil {
		return types.Redelegation{}, false
	}

	red, err := types.UnmarshalRED(k.cdc, value)
	// An error here means that what we got wasn't the right type
	if err != nil {
		return types.Redelegation{}, false
	}

	return red, true
}

// return the validator that is unbonding with a certain unbonding op ID
func (k Keeper) GetValidatorByUnbondingId(
	ctx sdk.Context, id uint64,
) (val types.Validator, found bool) {
	store := ctx.KVStore(k.storeKey)

	valKey := store.Get(types.GetUnbondingIndexKey(id))
	if valKey == nil {
		return types.Validator{}, false
	}

	value := store.Get(valKey)
	if value == nil {
		return types.Validator{}, false
	}

	val, err := types.UnmarshalValidator(k.cdc, value)
	// An error here means that what we got wasn't the right type
	if err != nil {
		return types.Validator{}, false
	}

	return val, true
}

// Set an index to look up an UnbondingDelegation by the unbondingId of an UnbondingDelegationEntry that it contains
func (k Keeper) SetUnbondingDelegationByUnbondingId(
	ctx sdk.Context, ubd types.UnbondingDelegation, id uint64,
) {
	store := ctx.KVStore(k.storeKey)

	delAddr, err := sdk.AccAddressFromBech32(ubd.DelegatorAddress)
	if err != nil {
		panic(err)
	}

	valAddr, err := sdk.ValAddressFromBech32(ubd.ValidatorAddress)
	if err != nil {
		panic(err)
	}

	ubdKey := types.GetUBDKey(delAddr, valAddr)
	store.Set(types.GetUnbondingIndexKey(id), ubdKey)

	// Set unbonding type so that we know how to deserialize it later
	k.SetUnbondingType(ctx, id, types.UnbondingType_UnbondingDelegation)
}

// Set an index to look up an Redelegation by the unbondingId of an RedelegationEntry that it contains
func (k Keeper) SetRedelegationByUnbondingId(ctx sdk.Context, red types.Redelegation, id uint64) {
	store := ctx.KVStore(k.storeKey)

	delAddr, err := sdk.AccAddressFromBech32(red.DelegatorAddress)
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

// Set an index to look up a Validator by the unbondingId corresponding to its current unbonding
func (k Keeper) SetValidatorByUnbondingId(ctx sdk.Context, val types.Validator, id uint64) {
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
// ----------------------------------------------------------------------------------------

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
// ----------------------------------------------------------------------------------------

func (k Keeper) UnbondingCanComplete(ctx sdk.Context, id uint64) error {
	unbondingType, found := k.GetUnbondingType(ctx, id)
	if !found {
		return types.ErrUnbondingNotFound
	}

	switch unbondingType {
	case types.UnbondingType_UnbondingDelegation:
		err := k.unbondingDelegationEntryCanComplete(ctx, id)
		if err != nil {
			return err
		}
	case types.UnbondingType_Redelegation:
		err := k.redelegationEntryCanComplete(ctx, id)
		if err != nil {
			return err
		}
	case types.UnbondingType_ValidatorUnbonding:
		err := k.validatorUnbondingCanComplete(ctx, id)
		if err != nil {
			return err
		}
	}

	// If an entry was not found
	return types.ErrUnbondingNotFound
}

func (k Keeper) unbondingDelegationEntryCanComplete(ctx sdk.Context, id uint64) error {
	ubd, found := k.GetUnbondingDelegationByUnbondingId(ctx, id)
	if !found {
		return types.ErrUnbondingNotFound
	}

	i, found := unbondingDelegationEntryArrayIndex(ubd, id)

	if !found {
		return types.ErrUnbondingNotFound
	}

	// The entry must be on hold
	if !ubd.Entries[i].OnHold() {
		return sdkerrors.Wrapf(
			types.ErrUnbondingOnHoldRefCountNegative,
			"undelegation unbondingId(%d), expecting UnbondingOnHoldRefCount > 0, got %T",
			id, ubd.Entries[i].UnbondingOnHoldRefCount,
		)
	}
	ubd.Entries[i].UnbondingOnHoldRefCount--

	// Check if entry is matured.
	if !ubd.Entries[i].OnHold() && ubd.Entries[i].IsMature(ctx.BlockHeader().Time) {
		// If matured, complete it.
		delegatorAddress, err := sdk.AccAddressFromBech32(ubd.DelegatorAddress)
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
	red, found := k.GetRedelegationByUnbondingId(ctx, id)
	if !found {
		return types.ErrUnbondingNotFound
	}

	i, found := redelegationEntryArrayIndex(red, id)
	if !found {
		return types.ErrUnbondingNotFound
	}

	// The entry must be on hold
	if !red.Entries[i].OnHold() {
		return sdkerrors.Wrapf(
			types.ErrUnbondingOnHoldRefCountNegative,
			"redelegation unbondingId(%d), expecting UnbondingOnHoldRefCount > 0, got %T",
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
	val, found := k.GetValidatorByUnbondingId(ctx, id)
	if !found {
		return types.ErrUnbondingNotFound
	}

	if val.UnbondingOnHoldRefCount <= 0 {
		return sdkerrors.Wrapf(
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
// ----------------------------------------------------------------------------------------
func (k Keeper) PutUnbondingOnHold(ctx sdk.Context, id uint64) error {
	unbondingType, found := k.GetUnbondingType(ctx, id)
	if !found {
		return types.ErrUnbondingNotFound
	}
	switch unbondingType {
	case types.UnbondingType_UnbondingDelegation:
		err := k.putUnbondingDelegationEntryOnHold(ctx, id)
		if err != nil {
			return err
		}
	case types.UnbondingType_Redelegation:
		err := k.putRedelegationEntryOnHold(ctx, id)
		if err != nil {
			return err
		}
	case types.UnbondingType_ValidatorUnbonding:
		err := k.putValidatorOnHold(ctx, id)
		if err != nil {
			return err
		}
	}

	// If an entry was not found
	return types.ErrUnbondingNotFound
}

func (k Keeper) putUnbondingDelegationEntryOnHold(ctx sdk.Context, id uint64) error {
	ubd, found := k.GetUnbondingDelegationByUnbondingId(ctx, id)
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
	red, found := k.GetRedelegationByUnbondingId(ctx, id)
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
	val, found := k.GetValidatorByUnbondingId(ctx, id)
	if !found {
		return types.ErrUnbondingNotFound
	}

	val.UnbondingOnHoldRefCount++
	k.SetValidator(ctx, val)

	return nil
}
