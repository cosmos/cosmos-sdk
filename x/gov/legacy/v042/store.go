package v042

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	v040gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v040"
)

const proposalIDLen = 8

// migratePrefixProposalAddress is a helper function that migrates all keys of format:
// <prefix_bytes><proposal_id (8 bytes)><address_bytes>
// into format:
// <prefix_bytes><proposal_id (8 bytes)><address_len (1 byte)><address_bytes>
func migratePrefixProposalAddress(store sdk.KVStore, prefixBz []byte) {
	oldStore := prefix.NewStore(store, prefixBz)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		proposalID := oldStoreIter.Key()[:proposalIDLen]
		addr := oldStoreIter.Key()[proposalIDLen:]
		newStoreKey := append(append(prefixBz, proposalID...), address.MustLengthPrefix(addr)...)

		// Set new key on store. Values don't change.
		store.Set(newStoreKey, oldStoreIter.Value())
		oldStore.Delete(oldStoreIter.Key())
	}
}

// MigrateStore performs in-place store migrations from v0.40 to v0.42. The
// migration includes:
//
// - Change addresses to be length-prefixed.
func MigrateStore(ctx sdk.Context, storeKey sdk.StoreKey, _ codec.Marshaler) error {
	store := ctx.KVStore(storeKey)
	migratePrefixProposalAddress(store, v040gov.DepositsKeyPrefix)
	migratePrefixProposalAddress(store, v040gov.VotesKeyPrefix)

	return nil
}
