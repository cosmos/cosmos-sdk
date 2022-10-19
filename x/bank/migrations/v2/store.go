package v2

import (
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v042auth "github.com/cosmos/cosmos-sdk/x/auth/migrations/v042"
	v1 "github.com/cosmos/cosmos-sdk/x/bank/migrations/v1"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// migrateSupply migrates the supply to be stored by denom key instead in a
// single blob.
// ref: https://github.com/cosmos/cosmos-sdk/issues/7092
func migrateSupply(st sdk.KVStore, cdc codec.BinaryCodec) error {
	// Old supply was stored as a single blob under the SupplyKey.
	var oldSupplyI v1.SupplyI
	err := cdc.UnmarshalInterface(st.Get(v1.SupplyKey), &oldSupplyI)
	if err != nil {
		return err
	}

	newStore := store.NewStoreAPI(st)
	// We delete the single key holding the whole blob.
	newStore.Delete(v1.SupplyKey)

	if oldSupplyI == nil {
		return nil
	}

	// We add a new key for each denom
	supplyStore := prefix.NewStore(st, SupplyKey)
	newSupplyStore := store.NewStoreAPI(supplyStore)

	// We're sure that SupplyI is a Supply struct, there's no other
	// implementation.
	oldSupply := oldSupplyI.(*types.Supply)
	for i := range oldSupply.Total {
		coin := oldSupply.Total[i]
		coinBz, err := coin.Amount.Marshal()
		if err != nil {
			return err
		}

		newSupplyStore.Set([]byte(coin.Denom), coinBz)
	}

	return nil
}

// migrateBalanceKeys migrate the balances keys to cater for variable-length
// addresses.
func migrateBalanceKeys(st sdk.KVStore) {
	// old key is of format:
	// prefix ("balances") || addrBytes (20 bytes) || denomBytes
	// new key is of format
	// prefix (0x02) || addrLen (1 byte) || addrBytes || denomBytes
	newStore := store.NewStoreAPI(st)
	oldStore := prefix.NewStore(st, v1.BalancesPrefix)
	oldStore2 := store.NewStoreAPI(oldStore)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		addr := v1.AddressFromBalancesStore(oldStoreIter.Key())
		denom := oldStoreIter.Key()[v042auth.AddrLen:]
		newStoreKey := types.CreatePrefixedAccountStoreKey(addr, denom)

		// Set new key on store. Values don't change.
		newStore.Set(newStoreKey, oldStoreIter.Value())
		oldStore2.Delete(oldStoreIter.Key())
	}
}

// MigrateStore performs in-place store migrations from v0.40 to v0.43. The
// migration includes:
//
// - Change addresses to be length-prefixed.
// - Change balances prefix to 1 byte
// - Change supply to be indexed by denom
// - Prune balances & supply with zero coins (ref: https://github.com/cosmos/cosmos-sdk/pull/9229)
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	migrateBalanceKeys(store)

	if err := pruneZeroBalances(store, cdc); err != nil {
		return err
	}

	if err := migrateSupply(store, cdc); err != nil {
		return err
	}

	return pruneZeroSupply(store)
}

// pruneZeroBalances removes the zero balance addresses from balances store.
func pruneZeroBalances(st sdk.KVStore, cdc codec.BinaryCodec) error {
	balancesStore := prefix.NewStore(st, BalancesPrefix)
	newBalancesStore := store.NewStoreAPI(balancesStore)
	iterator := balancesStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var balance sdk.Coin
		if err := cdc.Unmarshal(iterator.Value(), &balance); err != nil {
			return err
		}

		if balance.IsZero() {
			newBalancesStore.Delete(iterator.Key())
		}
	}
	return nil
}

// pruneZeroSupply removes zero balance denom from supply store.
func pruneZeroSupply(st sdk.KVStore) error {
	supplyStore := prefix.NewStore(st, SupplyKey)
	newSupplyStore := store.NewStoreAPI(supplyStore)
	iterator := supplyStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var amount math.Int
		if err := amount.Unmarshal(iterator.Value()); err != nil {
			return err
		}

		if amount.IsZero() {
			newSupplyStore.Delete(iterator.Key())
		}
	}

	return nil
}
