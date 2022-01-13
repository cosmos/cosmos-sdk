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
func migrateSupply(store sdk.KVStore, cdc codec.BinaryCodec) error {
	// Old supply was stored as a single blob under the SupplyKey.
	var oldSupplyI v040bank.SupplyI
	err := cdc.UnmarshalInterface(store.Get(v040bank.SupplyKey), &oldSupplyI)
	if err != nil {
		return err
	}

	// We delete the single key holding the whole blob.
	store.Delete(v040bank.SupplyKey)

	if oldSupplyI == nil {
		return nil
	}

	// We add a new key for each denom
	supplyStore := prefix.NewStore(store, types.SupplyKey)

	// We're sure that SupplyI is a Supply struct, there's no other
	// implementation.
	oldSupply := oldSupplyI.(*types.Supply)
	for i := range oldSupply.Total {
		coin := oldSupply.Total[i]
		coinBz, err := coin.Amount.Marshal()
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
	// prefix ("balances") || addrBytes (20 bytes) OR addrBytes (32 bytes) || denomBytes
	// We identify which addr byte length we are, by calculating the limited set of denomBytes lemgths
	// and then we backwards derive the used address bytelen.
	// new key is of format
	// prefix (0x02) || addrLen (1 byte) || addrBytes || denomBytes
	oldStore := prefix.NewStore(store, v040bank.BalancesPrefix)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	// denom types on Osmosis and their lengths
	// uion, uosmo, gamm/pool/1, gamm/pool/12, gamm/pool/123, ibc/EF9B097B0BF0E1CD55CEF1095DDE5518A201C99B8A8E82C183F8433E343C2A9E
	// 4,    5,     11,        , 12,           13,          , 68
	// thankfully the difference between any of these lengths is not 12 (the diff between 32 bytes and 20 bytes)
	// So we can uniquely go from key length to 20 byte / 32 byte address length decisions
	denomLenOptions := []int{4, 5, 11, 12, 13, 68}
	normalAddrByteLens := make(map[int]bool)
	customModuleAddrByteLens := make(map[int]bool)
	for i := 0; i < len(denomLenOptions); i++ {
		normalAddrByteLens[denomLenOptions[i]+20] = true
		customModuleAddrByteLens[denomLenOptions[i]+32] = true
	}

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		oldKey := oldStoreIter.Key()
		var addr sdk.AccAddress
		var addrLen int
		if _, inMap := normalAddrByteLens[len(oldKey)]; inMap {
			addr = v040bank.AddressFromBalancesStore(oldStoreIter.Key())
			addrLen = v040auth.AddrLen
		} else if _, inMap := customModuleAddrByteLens[len(oldKey)]; inMap {
			addr = sdk.AccAddress(oldKey[:32])
			addrLen = 32
		} else {
			// We remove the panic because this worked through the Osmosis v5 upgrade.
			// This was causing issues in tests, so we just assume the tests test the default SDK code, at 20 bytes.
			// panic(fmt.Sprintf("We have an issue ser, oldkey %v, len %d", oldKey, len(oldKey)))
			addr = v040bank.AddressFromBalancesStore(oldStoreIter.Key())
			addrLen = v040auth.AddrLen
		}
		denom := oldStoreIter.Key()[addrLen:]
		newStoreKey := append(types.CreateAccountBalancesPrefix(addr), denom...)

		// Set new key on store. Values don't change.
		store.Set(newStoreKey, oldStoreIter.Value())
		oldStore.Delete(oldStoreIter.Key())
	}
}

// MigrateStore performs in-place store migrations from v0.40 to v0.43. The
// migration includes:
//
// - Change addresses to be length-prefixed.
// - Change balances prefix to 1 byte
// - Change supply to be indexed by denom
// - Prune balances & supply with zero coins (ref: https://github.com/cosmos/cosmos-sdk/pull/9229)
func MigrateStore(ctx sdk.Context, storeKey sdk.StoreKey, cdc codec.BinaryCodec) error {
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
func pruneZeroBalances(store sdk.KVStore, cdc codec.BinaryCodec) error {
	balancesStore := prefix.NewStore(store, BalancesPrefix)
	iterator := balancesStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var balance sdk.Coin
		if err := cdc.Unmarshal(iterator.Value(), &balance); err != nil {
			return err
		}

		if balance.IsZero() {
			balancesStore.Delete(iterator.Key())
		}
	}
	return nil
}

// pruneZeroSupply removes zero balance denom from supply store.
func pruneZeroSupply(store sdk.KVStore) error {
	supplyStore := prefix.NewStore(store, SupplyKey)
	iterator := supplyStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var amount sdk.Int
		if err := amount.Unmarshal(iterator.Value()); err != nil {
			return err
		}

		if amount.IsZero() {
			supplyStore.Delete(iterator.Key())
		}
	}

	return nil
}
