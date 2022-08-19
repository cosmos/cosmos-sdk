package v5

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// MigrateStore performs in-place store migrations from v3 to v4.
/*
	this migrateStore will remove the ubdEntries with same creation_height
	and create a new ubdEntry with updated balance and initial_balance
*/
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.UnbondingDelegationKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		ubd := types.MustUnmarshalUBD(cdc, iterator.Value())
		entriesAtSameCreationHeight := make(map[int64][]types.UnbondingDelegationEntry)
		for _, ubdEntry := range ubd.Entries {
			entriesAtSameCreationHeight[ubdEntry.CreationHeight] = append(entriesAtSameCreationHeight[ubdEntry.CreationHeight], ubdEntry)
		}
		// clear the old ubdEntries
		ubd.Entries = nil

		for _, entries := range entriesAtSameCreationHeight {
			var ubdEntry types.UnbondingDelegationEntry
			for _, entry := range entries {
				ubdEntry.Balance = ubdEntry.Balance.Add(entry.Balance)
				ubdEntry.InitialBalance = ubdEntry.InitialBalance.Add(entry.InitialBalance)
				ubdEntry.CreationHeight = entry.CreationHeight
				ubdEntry.CompletionTime = entry.CompletionTime
			}
			ubd.Entries = append(ubd.Entries, ubdEntry)
		}

		// set the new ubd to the store
		setUBDToStore(ctx, storeKey, cdc, ubd)
	}

	return nil
}

func setUBDToStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec, ubd types.UnbondingDelegation) {
	delegatorAddress := sdk.MustAccAddressFromBech32(ubd.DelegatorAddress)

	store := ctx.KVStore(storeKey)
	bz := types.MustMarshalUBD(cdc, ubd)
	addr, err := sdk.ValAddressFromBech32(ubd.ValidatorAddress)
	if err != nil {
		panic(err)
	}
	key := types.GetUBDKey(delegatorAddress, addr)
	store.Set(key, bz)
	store.Set(types.GetUBDByValIndexKey(delegatorAddress, addr), []byte{}) // index, store empty bytes
}
