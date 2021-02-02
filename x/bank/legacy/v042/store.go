package v042

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	v040auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v040"
	v040bank "github.com/cosmos/cosmos-sdk/x/bank/legacy/v040"
)

// KVStore keys
var (
	// BalancesPrefix is the for the account balances store. We use a byte
	// (instead of say `[]]byte("balances")` to save some disk space).
	BalancesPrefix = []byte{0x02}
)

// AddressFromBalancesStore returns an account address from a balances prefix
// store. The key must not contain the perfix BalancesPrefix as the prefix store
// iterator discards the actual prefix.
func AddressFromBalancesStore(key []byte) sdk.AccAddress {
	addrLen := key[0]
	addr := key[1 : addrLen+1]

	return sdk.AccAddress(addr)
}

// CreateAccountBalancesPrefix creates the prefix for an account's balances.
func CreateAccountBalancesPrefix(addr []byte) []byte {
	return append(BalancesPrefix, address.MustLengthPrefix(addr)...)
}

// StoreMigration performs in-place store migrations from v0.40 to v0.42. The
// migration includes:
//
// - Change addresses to be length-prefixed.
// - Change balances prefix to 1 byte
func StoreMigration(store sdk.KVStore) error {
	// old key is of format:
	// prefix ("balances") || addrBytes (20 bytes) || denomBytes
	// new key is of format
	// prefix (0x02) || addrLen (1 byte) || addrBytes || denomBytes
	oldStore := prefix.NewStore(store, v040bank.BalancesPrefix)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		addr := v040bank.AddressFromBalancesStore(oldStoreIter.Key())
		denom := oldStoreIter.Key()[v040auth.AddrLen:]
		newStoreKey := append(CreateAccountBalancesPrefix(addr), denom...)

		// Set new key on store. Values don't change.
		store.Set(newStoreKey, oldStoreIter.Value())
		oldStore.Delete(oldStoreIter.Key())
	}

	return nil
}
