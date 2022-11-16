package keeper

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Increments and returns a unique ID for an UnbondingDelegationEntry
func (k Keeper) IncrementUnbondingId(ctx sdk.Context) (unbondingDelegationEntryId uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.UnbondingDelegationEntryIdKey)

	if bz == nil {
		unbondingDelegationEntryId = 0
	} else {
		unbondingDelegationEntryId = binary.BigEndian.Uint64(bz)
	}

	unbondingDelegationEntryId = unbondingDelegationEntryId + 1

	// Convert back into bytes for storage
	bz = make([]byte, 8)
	binary.BigEndian.PutUint64(bz, unbondingDelegationEntryId)

	store.Set(types.UnbondingDelegationEntryIdKey, bz)

	return unbondingDelegationEntryId
}

// Remove a ValidatorByUnbondingIndex
func (k Keeper) DeleteUnbondingIndex(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)

	indexKey := types.GetUnbondingIndexKey(id)

	store.Delete(indexKey)
}

// return a unbonding delegation that has an unbonding delegation entry with a certain ID
func (k Keeper) GetUnbondingDelegationByUnbondingId(
	ctx sdk.Context, id uint64,
) (ubd types.UnbondingDelegation, found bool) {
	store := ctx.KVStore(k.storeKey)
	indexKey := types.GetUnbondingIndexKey(id)
	ubdeKey := store.Get(indexKey)

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
	indexKey := types.GetUnbondingIndexKey(id)
	redKey := store.Get(indexKey)

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
	indexKey := types.GetUnbondingIndexKey(id)
	valKey := store.Get(indexKey)

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
func (k Keeper) SetUnbondingDelegationByUnbondingIndex(ctx sdk.Context, ubd types.UnbondingDelegation, id uint64) {
	store := ctx.KVStore(k.storeKey)

	delAddr, err := sdk.AccAddressFromBech32(ubd.DelegatorAddress)
	if err != nil {
		panic(err)
	}

	valAddr, err := sdk.ValAddressFromBech32(ubd.ValidatorAddress)
	if err != nil {
		panic(err)
	}

	indexKey := types.GetUnbondingIndexKey(id)
	ubdKey := types.GetUBDKey(delAddr, valAddr)

	store.Set(indexKey, ubdKey)
}

// Set an index to look up an Redelegation by the unbondingId of an RedelegationEntry that it contains
func (k Keeper) SetRedelegationByUnbondingIndex(ctx sdk.Context, red types.Redelegation, id uint64) {
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

	indexKey := types.GetUnbondingIndexKey(id)
	redKey := types.GetREDKey(delAddr, valSrcAddr, valDstAddr)

	store.Set(indexKey, redKey)
}

// Set an index to look up a Validator by the unbondingId corresponding to its current unbonding
func (k Keeper) SetValidatorByUnbondingIndex(ctx sdk.Context, val types.Validator, id uint64) {
	store := ctx.KVStore(k.storeKey)

	valAddr, err := sdk.ValAddressFromBech32(val.OperatorAddress)
	if err != nil {
		panic(err)
	}

	indexKey := types.GetUnbondingIndexKey(id)
	valKey := types.GetValidatorKey(valAddr)

	store.Set(indexKey, valKey)
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

// UnbondingCanComplete allows a stopped unbonding operation such as an
// unbonding delegation, a redelegation, or a validator unbonding to complete
// ----------------------------------------------------------------------------------------

func (k Keeper) UnbondingCanComplete(ctx sdk.Context, id uint64) error {
	found, err := k.unbondingDelegationEntryCanComplete(ctx, id)
	if err != nil {
		return err
	}
	if found {
		return nil
	}

	found, err = k.redelegationEntryCanComplete(ctx, id)
	if err != nil {
		return err
	}
	if found {
		return nil
	}

	found, err = k.validatorUnbondingCanComplete(ctx, id)
	if err != nil {
		return err
	}
	if found {
		return nil
	}

	// If an entry was not found
	return types.ErrUnbondingNotFound
}

func (k Keeper) unbondingDelegationEntryCanComplete(ctx sdk.Context, id uint64) (found bool, err error) {
	ubd, found := k.GetUnbondingDelegationByUnbondingId(ctx, id)
	if !found {
		return false, nil
	}

	i, found := unbondingDelegationEntryArrayIndex(ubd, id)

	if !found {
		return false, nil
	}

	// Check if entry is matured.
	if !ubd.Entries[i].IsMature(ctx.BlockHeader().Time) {
		// If not matured, set onHold to false
		ubd.Entries[i].UnbondingOnHold = false
	} else {
		// If matured, complete it.
		delegatorAddress, err := sdk.AccAddressFromBech32(ubd.DelegatorAddress)
		if err != nil {
			return false, err
		}

		bondDenom := k.GetParams(ctx).BondDenom

		// track undelegation only when remaining or truncated shares are non-zero
		if !ubd.Entries[i].Balance.IsZero() {
			amt := sdk.NewCoin(bondDenom, ubd.Entries[i].Balance)
			if err := k.bankKeeper.UndelegateCoinsFromModuleToAccount(
				ctx, types.NotBondedPoolName, delegatorAddress, sdk.NewCoins(amt),
			); err != nil {
				return false, err
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
	return true, nil
}

func (k Keeper) redelegationEntryCanComplete(ctx sdk.Context, id uint64) (found bool, err error) {
	red, found := k.GetRedelegationByUnbondingId(ctx, id)
	if !found {
		return false, nil
	}

	i, found := redelegationEntryArrayIndex(red, id)
	if !found {
		return false, nil
	}

	if !red.Entries[i].IsMature(ctx.BlockHeader().Time) {
		// If not matured, set onHold to false
		red.Entries[i].UnbondingOnHold = false
	} else {
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
	return true, nil
}

func (k Keeper) validatorUnbondingCanComplete(ctx sdk.Context, id uint64) (found bool, err error) {
	val, found := k.GetValidatorByUnbondingId(ctx, id)
	if !found {
		return false, nil
	}

	if !val.IsMature(ctx.BlockTime(), ctx.BlockHeight()) {
		val.UnbondingOnHold = false
		k.SetValidator(ctx, val)
	} else {
		// If unbonding is mature complete it
		val = k.UnbondingToUnbonded(ctx, val)
		if val.GetDelegatorShares().IsZero() {
			k.RemoveValidator(ctx, val.GetOperator())
		}

		k.DeleteUnbondingIndex(ctx, id)
	}

	return true, nil
}

// PutUnbondingOnHold allows an external module to stop an unbonding operation such as an
// unbonding delegation, a redelegation, or a validator unbonding
// ----------------------------------------------------------------------------------------
func (k Keeper) PutUnbondingOnHold(ctx sdk.Context, id uint64) error {
	found := k.putUnbondingDelegationEntryOnHold(ctx, id)
	if found {
		return nil
	}

	found = k.putRedelegationEntryOnHold(ctx, id)
	if found {
		return nil
	}

	found = k.putValidatorOnHold(ctx, id)
	if found {
		return nil
	}

	// If an entry was not found
	return types.ErrUnbondingNotFound
}

func (k Keeper) putUnbondingDelegationEntryOnHold(ctx sdk.Context, id uint64) (found bool) {
	ubd, found := k.GetUnbondingDelegationByUnbondingId(ctx, id)
	if !found {
		return false
	}

	i, found := unbondingDelegationEntryArrayIndex(ubd, id)
	if !found {
		return false
	}

	ubd.Entries[i].UnbondingOnHold = true

	k.SetUnbondingDelegation(ctx, ubd)

	return true
}

func (k Keeper) putRedelegationEntryOnHold(ctx sdk.Context, id uint64) (found bool) {
	red, found := k.GetRedelegationByUnbondingId(ctx, id)
	if !found {
		return false
	}

	i, found := redelegationEntryArrayIndex(red, id)
	if !found {
		return false
	}

	red.Entries[i].UnbondingOnHold = true

	k.SetRedelegation(ctx, red)

	return true
}

func (k Keeper) putValidatorOnHold(ctx sdk.Context, id uint64) (found bool) {
	val, found := k.GetValidatorByUnbondingId(ctx, id)
	if !found {
		return false
	}

	val.UnbondingOnHold = true

	k.SetValidator(ctx, val)

	return true
}
