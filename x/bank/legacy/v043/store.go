package v043

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v040auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v040"
	v040bank "github.com/cosmos/cosmos-sdk/x/bank/legacy/v040"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// migrateSupply migrates the supply to be stored by denom key instead in a
// single blob.
// ref: https://github.com/cosmos/cosmos-sdk/issues/7092
func migrateSupply(store sdk.KVStore, cdc codec.BinaryMarshaler) error {
	// Old supply was stored as a single blob under the SupplyKey.
	var oldSupply types.SupplyI // nolint:staticcheck
	err := cdc.UnmarshalInterface(store.Get(v040bank.SupplyKey), &oldSupply)
	if err != nil {
		return err
	}

	// We delete the single key holding the whole blob.
	store.Delete(v040bank.SupplyKey)

	// We add a new key for each denom
	supplyStore := prefix.NewStore(store, types.SupplyKey)
	for i := range oldSupply.GetTotal() {
		coin := oldSupply.GetTotal()[i]
		coinBz, err := cdc.MarshalBinaryBare(&coin)
		if err != nil {
			return err
		}

		supplyStore.Set([]byte(coin.Denom), coinBz)
	}

	return nil
}

// migrateBalanceKeys migrate the balances keys to cater for variable-length
// addresses.
func migrateBalanceKeys(store sdk.KVStore) {
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
		newStoreKey := append(types.CreateAccountBalancesPrefix(addr), denom...)

		// Set new key on store. Values don't change.
		store.Set(newStoreKey, oldStoreIter.Value())
		oldStore.Delete(oldStoreIter.Key())
	}
}

// MigrateStore performs in-place store migrations from v0.40 to v0.42. The
// migration includes:
//
// - Change addresses to be length-prefixed.
// - Change balances prefix to 1 byte
// - Change supply to be indexed by denom
func MigrateStore(ctx sdk.Context, storeKey sdk.StoreKey, cdc codec.BinaryMarshaler) error {
	store := ctx.KVStore(storeKey)

	migrateBalanceKeys(store)
	return migrateSupply(store, cdc)
}
