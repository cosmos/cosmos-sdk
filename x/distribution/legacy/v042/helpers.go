package v042

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	v040auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v040"
)

// MigratePrefixAddress is a helper function that migrates all keys of format:
// prefix_bytes | address_bytes
// into format:
// prefix_bytes | address_len (1 byte) | address_bytes
func MigratePrefixAddress(store sdk.KVStore, prefixBz []byte) {
	oldStore := prefix.NewStore(store, prefixBz)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		addr := oldStoreIter.Key()
		newStoreKey := append(prefixBz, address.MustLengthPrefix(addr)...)

		// Set new key on store. Values don't change.
		store.Set(newStoreKey, oldStoreIter.Value())
		oldStore.Delete(oldStoreIter.Key())
	}
}

// MigratePrefixAddressBytes is a helper function that migrates all keys of format:
// prefix_bytes | address_bytes | arbitrary_bytes
// into format:
// prefix_bytes | address_len (1 byte) | address_bytes | arbitrary_bytes
func MigratePrefixAddressBytes(store sdk.KVStore, prefixBz []byte) {
	oldStore := prefix.NewStore(store, prefixBz)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		addr := oldStoreIter.Key()[:v040auth.AddrLen]
		endBz := oldStoreIter.Key()[v040auth.AddrLen:]
		newStoreKey := append(append(prefixBz, address.MustLengthPrefix(addr)...), endBz...)

		// Set new key on store. Values don't change.
		store.Set(newStoreKey, oldStoreIter.Value())
		oldStore.Delete(oldStoreIter.Key())
	}
}

// MigratePrefixAddressAddress is a helper function that migrates all keys of format:
// prefix_bytes | address_1_bytes | address_2_bytes
// into format:
// prefix_bytes | address_1_len (1 byte) | address_1_bytes | address_2_len (1 byte) | address_2_bytes
func MigratePrefixAddressAddress(store sdk.KVStore, prefixBz []byte) {
	oldStore := prefix.NewStore(store, prefixBz)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		addr1 := oldStoreIter.Key()[:v040auth.AddrLen]
		addr2 := oldStoreIter.Key()[v040auth.AddrLen:]
		newStoreKey := append(append(prefixBz, address.MustLengthPrefix(addr1)...), address.MustLengthPrefix(addr2)...)

		// Set new key on store. Values don't change.
		store.Set(newStoreKey, oldStoreIter.Value())
		oldStore.Delete(oldStoreIter.Key())
	}
}
