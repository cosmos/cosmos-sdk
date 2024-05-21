package keeper

import (
	"fmt"
	"sort"
	"time"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (k Keeper) GetLastTokenizeShareRecordID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.LastTokenizeShareRecordIDKey)
	if bytes == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bytes)
}

func (k Keeper) SetLastTokenizeShareRecordID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastTokenizeShareRecordIDKey, sdk.Uint64ToBigEndian(id))
}

func (k Keeper) GetTokenizeShareRecord(ctx sdk.Context, id uint64) (tokenizeShareRecord types.TokenizeShareRecord, err error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetTokenizeShareRecordByIndexKey(id))
	if bz == nil {
		return tokenizeShareRecord, sdkerrors.Wrap(types.ErrTokenizeShareRecordNotExists, fmt.Sprintf("tokenizeShareRecord %d does not exist", id))
	}

	k.cdc.MustUnmarshal(bz, &tokenizeShareRecord)
	return tokenizeShareRecord, nil
}

func (k Keeper) GetTokenizeShareRecordsByOwner(ctx sdk.Context, owner sdk.AccAddress) (tokenizeShareRecords []types.TokenizeShareRecord) {
	store := ctx.KVStore(k.storeKey)

	it := sdk.KVStorePrefixIterator(store, types.GetTokenizeShareRecordIdsByOwnerPrefix(owner))
	defer it.Close()

	for ; it.Valid(); it.Next() {
		var id gogotypes.UInt64Value
		k.cdc.MustUnmarshal(it.Value(), &id)

		tokenizeShareRecord, err := k.GetTokenizeShareRecord(ctx, id.Value)
		if err != nil {
			continue
		}
		tokenizeShareRecords = append(tokenizeShareRecords, tokenizeShareRecord)
	}
	return
}

func (k Keeper) GetTokenizeShareRecordByDenom(ctx sdk.Context, denom string) (types.TokenizeShareRecord, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetTokenizeShareRecordIDByDenomKey(denom))
	if bz == nil {
		return types.TokenizeShareRecord{}, fmt.Errorf("tokenize share record not found from denom: %s", denom)
	}

	var id gogotypes.UInt64Value
	k.cdc.MustUnmarshal(bz, &id)

	return k.GetTokenizeShareRecord(ctx, id.Value)
}

func (k Keeper) GetAllTokenizeShareRecords(ctx sdk.Context) (tokenizeShareRecords []types.TokenizeShareRecord) {
	store := ctx.KVStore(k.storeKey)

	it := sdk.KVStorePrefixIterator(store, types.TokenizeShareRecordPrefix)
	defer it.Close()

	for ; it.Valid(); it.Next() {
		var tokenizeShareRecord types.TokenizeShareRecord
		k.cdc.MustUnmarshal(it.Value(), &tokenizeShareRecord)

		tokenizeShareRecords = append(tokenizeShareRecords, tokenizeShareRecord)
	}
	return
}

func (k Keeper) AddTokenizeShareRecord(ctx sdk.Context, tokenizeShareRecord types.TokenizeShareRecord) error {
	if k.hasTokenizeShareRecord(ctx, tokenizeShareRecord.Id) {
		return sdkerrors.Wrapf(types.ErrTokenizeShareRecordAlreadyExists, "TokenizeShareRecord already exists: %d", tokenizeShareRecord.Id)
	}

	k.setTokenizeShareRecord(ctx, tokenizeShareRecord)

	owner, err := sdk.AccAddressFromBech32(tokenizeShareRecord.Owner)
	if err != nil {
		return err
	}

	k.setTokenizeShareRecordWithOwner(ctx, owner, tokenizeShareRecord.Id)
	k.setTokenizeShareRecordWithDenom(ctx, tokenizeShareRecord.GetShareTokenDenom(), tokenizeShareRecord.Id)

	return nil
}

func (k Keeper) DeleteTokenizeShareRecord(ctx sdk.Context, recordID uint64) error {
	record, err := k.GetTokenizeShareRecord(ctx, recordID)
	if err != nil {
		return err
	}
	owner, err := sdk.AccAddressFromBech32(record.Owner)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetTokenizeShareRecordByIndexKey(recordID))
	store.Delete(types.GetTokenizeShareRecordIDByOwnerAndIDKey(owner, recordID))
	store.Delete(types.GetTokenizeShareRecordIDByDenomKey(record.GetShareTokenDenom()))
	return nil
}

func (k Keeper) hasTokenizeShareRecord(ctx sdk.Context, id uint64) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetTokenizeShareRecordByIndexKey(id))
}

func (k Keeper) setTokenizeShareRecord(ctx sdk.Context, tokenizeShareRecord types.TokenizeShareRecord) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&tokenizeShareRecord)

	store.Set(types.GetTokenizeShareRecordByIndexKey(tokenizeShareRecord.Id), bz)
}

func (k Keeper) setTokenizeShareRecordWithOwner(ctx sdk.Context, owner sdk.AccAddress, id uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: id})

	store.Set(types.GetTokenizeShareRecordIDByOwnerAndIDKey(owner, id), bz)
}

func (k Keeper) deleteTokenizeShareRecordWithOwner(ctx sdk.Context, owner sdk.AccAddress, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetTokenizeShareRecordIDByOwnerAndIDKey(owner, id))
}

func (k Keeper) setTokenizeShareRecordWithDenom(ctx sdk.Context, denom string, id uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: id})

	store.Set(types.GetTokenizeShareRecordIDByDenomKey(denom), bz)
}

// TransferRedelegationsOfTokenizedShares transfers
// redelegations whose underlying shares are involved in a given tokenization/redemption
// by changing the delegator address to the destination address.
// In the case of tokenization, the redelegations should be transferred from a delegator to the share record module account.
// In the case of redemption, the redelegations should be transferred from the share record module account to the delegator.
func (k Keeper) TransferRedelegationsOfTokenizedShares(ctx sdk.Context, delegation types.Delegation, amount sdk.Dec, srcAddr, dstAddr sdk.AccAddress) error {
	// Iterate over the delegator's redelegations and store the one
	// for which the shares are part of tokenization
	reds := []types.Redelegation{}
	k.IterateDelegatorRedelegations(ctx, srcAddr, func(red types.Redelegation) (stop bool) {
		// check only redelegations to the validator of the delegation
		if red.ValidatorDstAddress == delegation.ValidatorAddress {
			reds = append(reds, red)
		}

		return false
	})
	return k.updateRedelegationsWithTokenizedShares(
		ctx,
		delegation,
		amount,
		reds,
		dstAddr.String(),
	)
}

// updateTokenizedSharesRedelegation defines the amount of shares from the given
// redelegations are involved in the process of tokenizing or redeeming a given amount shares of a delegation.
// If the result is positive, it transfers redelegations for the same amount of shares
// to the given destination delegator address.
func (k Keeper) updateRedelegationsWithTokenizedShares(
	ctx sdk.Context,
	delegation types.Delegation,
	amount sdk.Dec,
	redelegations []types.Redelegation,
	dstDelegatorAddress string,
) error {
	// compute the amount of shares from redelegations that are still bonded
	redsShares, err := k.ComputeRemainingRedelegatedSharesAfterUnbondings(
		ctx,
		delegation.GetDelegatorAddr(),
		redelegations,
		delegation.GetValidatorAddr(),
	)
	if err != nil {
		return err
	}

	// compute the shares left, e.g. not coming from redelegations
	sharesLeft := delegation.Shares.Sub(redsShares)

	// if the amount of shares left is negative, this means there are some redelegations
	// tracking shares that do not longer exist
	if sharesLeft.IsNegative() {
		return fmt.Errorf("delegator address %s has more redelegated shares %s than delegation shares %s in validator %s",
			delegation.DelegatorAddress, redsShares, sharesLeft, delegation.ValidatorAddress)
	}

	// if the shares left is GTE to the amount,
	// no redelegations need to be transferred
	if sharesLeft.GTE(amount) {
		return nil
	}

	// compute how many redelegated shares are required
	// to be transferred
	amountToTransfer := amount.Sub(sharesLeft)
	// get the minimum subset of the redelegations for which the total shares
	// GTE to the amount redelegated shares to transfer
	redelegationsToTransfer, err := GetMinimumRedelegationsSubsetByShares(amountToTransfer, redelegations)
	if err != nil {
		return err
	}

	// transfer redelegations
	transferredReds, remainingReds, ok := TransferRedelegations(
		amountToTransfer,
		dstDelegatorAddress,
		redelegationsToTransfer,
	)
	if !ok {
		return fmt.Errorf("fail to transfer %s shares from redelegations due to insufficient delegation shares %s",
			amount, redsShares)
	}

	// check that we get the expected returned length of redelegations
	// note that it should never happen
	if len(redelegationsToTransfer) != len(transferredReds) || len(remainingReds) > 1 {
		return fmt.Errorf("fail to tokenize redelegation shares: length of redelegations to transfer is not ok")
	}

	// update redelegations in store
	for i := 0; i < len(redelegationsToTransfer); i++ {
		k.SetRedelegation(ctx, transferredReds[i])
		k.RemoveRedelegation(ctx, redelegationsToTransfer[i])
		// insert the redelegation into the queue
		// its ok to not update the old queue entry because erroring ones are ignored
		for _, entry := range transferredReds[i].Entries {
			k.InsertRedelegationQueue(ctx, transferredReds[i], entry.CompletionTime)
		}
	}

	// this is the case where a redelegation is split into two because
	// the amount of shares to transfer does not match exactly with the shares of the redelegations
	if len(remainingReds) == 1 {
		k.SetRedelegation(ctx, remainingReds[0])
		// insert the redelegation into the queue
		// its ok to not update the old queue entry because erroring ones are ignored
		for _, entry := range remainingReds[0].Entries {
			k.InsertRedelegationQueue(ctx, remainingReds[0], entry.CompletionTime)
		}
	}

	return nil
}

// TransferRedelegations iterates through the redelegations and updates their delegator address to the dstDelegatorAddr
// until redelegations with *amount* shares have had their delegator address changed.
// Note that the last redelegation that is transferred may have too many shares. In that case,
// we split that redelegation into two, one with the correct amount of shares and one with the remaining shares.
// The function returns the transferred redelegations, the remaining redelegations, and a boolean indicating
// whether the transfer was successful or not.
func TransferRedelegations(
	amount sdk.Dec,
	dstDelegatorAddr string,
	redelegations []types.Redelegation,
) (transferredReds []types.Redelegation, remainingReds []types.Redelegation, ok bool) {
	// iterate over all the redelegations until
	// their combined shares is equal to the given amount
	aLeft := amount
	for idx, red := range redelegations {
		var entriesLeft, entriesTransferred []types.RedelegationEntry
		aLeft, entriesLeft, entriesTransferred = updateRedelegationEntriesByAmount(aLeft, red.Entries)

		// append transferred redelegations
		transferredReds = append(transferredReds, types.Redelegation{
			DelegatorAddress:    dstDelegatorAddr,
			ValidatorSrcAddress: red.ValidatorSrcAddress,
			ValidatorDstAddress: red.ValidatorDstAddress,
			Entries:             entriesTransferred,
		})

		// check if enough shares are collected
		if aLeft.IsZero() {
			if len(entriesLeft) > 0 {
				// update remaining redelegations
				remainingReds = append(remainingReds, types.Redelegation{
					DelegatorAddress:    red.DelegatorAddress,
					ValidatorSrcAddress: red.ValidatorSrcAddress,
					ValidatorDstAddress: red.ValidatorDstAddress,
					Entries:             entriesLeft,
				})

				remainingReds = append(remainingReds, redelegations[idx+1:]...)
			}

			break
		}
	}

	// check that the total amount is transferred
	if aLeft.IsPositive() {
		return nil, nil, false
	}

	return transferredReds, remainingReds, true
}

// updateRedelegationEntriesByAmount splits the given redelegation entries into two slices:
// the second slice contains entries with up to the given amount of shares, and the first slice contains the remaining entries.
func updateRedelegationEntriesByAmount(amount sdk.Dec, entries []types.RedelegationEntry) (sdk.Dec, []types.RedelegationEntry, []types.RedelegationEntry) {
	// res1 will be the slice that contains the remaining relegedations
	res1 := make([]types.RedelegationEntry, 0)
	// res2 will be the slice that contains the transferred entries
	res2 := make([]types.RedelegationEntry, 0)

	// iterate over all the entries and add them to the first slice
	// until their combined shares is equal to the given amount
	aLeft := amount
	for _, entry := range entries {
		if aLeft.IsZero() {
			res1 = append(res1, entry)
		} else {
			// check if the entry shares are less than the remaining shares
			if entry.SharesDst.LTE(aLeft) {
				res2 = append(res2, entry)
				aLeft = aLeft.Sub(entry.SharesDst)
			} else {
				// split the entry into two entries
				// one with the given amount of shares and one with the remaining shares
				entry1 := entry
				entry1.SharesDst = entry.SharesDst.Sub(aLeft) // collect the remaining shares
				res1 = append(res1, entry1)

				entry2 := entry
				entry2.SharesDst = aLeft // finish filling the given amount of shares
				res2 = append(res2, entry2)

				aLeft = sdk.ZeroDec()
			}
		}
	}

	return aLeft, res1, res2
}

// ComputeRemainingRedelegatedSharesAfterUnbondings takes a delegator address, validator address, and a list of redelegations
// that should all be BY the delegator TO the given validator (we do not care about the source validators here).
// It computes the shares of redelegations that are *not* matched by a subsequent unbonding.
// For example, consider this scenario for delegator D and validator V, assuming that no redelegations or unbondings
// complete during this time:
// - First, D redelegates 10 shares to V - D has 10 shares from redelegations to V
// - Then, D unbonds 5 shares from V - we assume that these unbonding shares are the ones that were just redelegated, so D has 5 shares from redelegations to V left
// - Finally, D unbonds 10 more shares to V - now D has 0 shares from redelegations to V left (and the other unbonding shares must come from a native delegation)
// This function returns the amount of shares that are still in the redelegations after the unbondings are taken into account in this manner,
// so in this example the outcome would be 0.
// See docs/architecture/adr-061-liquid-staking.md for more information.
func (k Keeper) ComputeRemainingRedelegatedSharesAfterUnbondings(
	ctx sdk.Context,
	delAddr sdk.AccAddress,
	reds []types.Redelegation,
	valAddr sdk.ValAddress,
) (sdk.Dec, error) {
	// delegationEntry defines an general entry representing either
	// the addition or withdrawal of delegation shares at completion time.
	type delegationEntry struct {
		completionTime time.Time
		shares         sdk.Dec
	}

	delegationEntries := []delegationEntry{}
	validator, found := k.GetValidator(ctx, valAddr)
	if !found {
		return sdk.ZeroDec(), types.ErrNoValidatorFound
	}

	for _, red := range reds {
		// sanity check
		// check that the redelegation has the given validator destination address
		if valAddr.String() != red.ValidatorDstAddress {
			return sdk.ZeroDec(), types.ErrBadRedelegationDst
		}
		// sanity check
		// check that the redelegation has the given delegator address
		if delAddr.String() != red.DelegatorAddress {
			return sdk.ZeroDec(), types.ErrBadDelegatorAddr
		}

		// store each redelegation entry as a delegation entry
		// adding shares at completion time
		for _, redEntry := range red.Entries {
			delegationEntries = append(delegationEntries, delegationEntry{
				redEntry.CompletionTime,
				// care about the destination shares because that is the current
				// amount of shares represented by this entry, and
				// this is how many shares this currently represents
				redEntry.SharesDst,
			})
		}
	}

	// go through all unbonding delegations
	ubd, found := k.GetUnbondingDelegation(ctx, delAddr, valAddr)
	if found {
		for _, ubdEntry := range ubd.Entries {
			// get the tokens this unbonding delegation entry represents right now
			// we care about the *current balance* because that is ultimately the
			// shares that will be removed at the completion time
			ubdEntryShares, err := validator.SharesFromTokens(ubdEntry.Balance)
			if err != nil {
				return sdk.ZeroDec(), err
			}
			// store each unbonding delegation entry as a delegation entry
			// withdrawing shares at completion time, by using it's negative amount of shares
			delegationEntries = append(delegationEntries,
				delegationEntry{
					ubdEntry.CompletionTime,
					ubdEntryShares.Neg(),
				})
		}
	}

	// sort delegation entries by completion time in ascending order
	// This is because we need to go through the delegation entries in chronological order
	// to match redelegations with unbondings
	sort.Slice(delegationEntries, func(i, j int) bool {
		return delegationEntries[i].completionTime.Before(delegationEntries[j].completionTime)
	})

	// Sum the shares of delegation entries, flooring negative values to zero.
	// This assumes that negative shares must have been taken from the initial delegation shares initially,
	// otherwise the withdrawing operation should have failed.
	remainingShares := sdk.ZeroDec()
	for _, entry := range delegationEntries {
		if remainingShares.Add(entry.shares).IsNegative() {
			remainingShares = sdk.ZeroDec()
			continue
		}

		remainingShares = remainingShares.Add(entry.shares)
	}

	return remainingShares, nil
}

// GetMinimumRedelegationsSubsetByShares takes a list of redelegations and
// returns a subset of it where the combined shares are greater than or equal to a given amount.
// It returns an error if the given amount is greater than the total shares in the given redelegations.
func GetMinimumRedelegationsSubsetByShares(amount sdk.Dec, redelegations []types.Redelegation) (out []types.Redelegation, err error) {
	redsShares := sdk.ZeroDec()
	for _, red := range redelegations {
		for _, entry := range red.Entries {
			redsShares = redsShares.Add(entry.SharesDst)
		}
		out = append(out, red)
		if redsShares.GTE(amount) {
			break
		}
	}

	if redsShares.LT(amount) {
		return nil, fmt.Errorf("shares from redelegations is less than the given amount: %s, %s", redsShares, amount)
	}

	return out, err
}
