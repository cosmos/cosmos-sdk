package v046

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	v043 "github.com/cosmos/cosmos-sdk/x/bank/migrations/v043"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// MigrateStore performs in-place store migrations from v0.43 to v0.45. The
// migration includes:
//
// - Migrate coin storage to save only amount.
// - Add an additional reverse index from denomination to address.
// - Remove duplicate denom from denom metadata store key.
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	err := addDenomReverseIndex(store, cdc)
	if err != nil {
		return err
	}

	return migrateDenomMetadata(store)
}

func addDenomReverseIndex(store sdk.KVStore, cdc codec.BinaryCodec) error {
	oldBalancesStore := prefix.NewStore(store, v043.BalancesPrefix)

	oldBalancesIter := oldBalancesStore.Iterator(nil, nil)
	defer oldBalancesIter.Close()

	denomPrefixStores := make(map[string]prefix.Store) // memoize prefix stores

	for ; oldBalancesIter.Valid(); oldBalancesIter.Next() {
		var balance sdk.Coin
		if err := cdc.Unmarshal(oldBalancesIter.Value(), &balance); err != nil {
			return err
		}

		addr, err := v043.AddressFromBalancesStore(oldBalancesIter.Key())
		if err != nil {
			return err
		}

		var coin sdk.DecCoin
		if err := cdc.Unmarshal(oldBalancesIter.Value(), &coin); err != nil {
			return err
		}

		bz, err := coin.Amount.Marshal()
		if err != nil {
			return err
		}

		newStore := prefix.NewStore(store, types.CreateAccountBalancesPrefix(addr))
		newStore.Set([]byte(coin.Denom), bz)

		denomPrefixStore, ok := denomPrefixStores[balance.Denom]
		if !ok {
			denomPrefixStore = prefix.NewStore(store, CreateDenomAddressPrefix(balance.Denom))
			denomPrefixStores[balance.Denom] = denomPrefixStore
		}

		// Store a reverse index from denomination to account address with a
		// sentinel value.
		denomPrefixStore.Set(address.MustLengthPrefix(addr), []byte{0})
	}

	return nil
}

func migrateDenomMetadata(store sdk.KVStore) error {
	oldDenomMetaDataStore := prefix.NewStore(store, v043.DenomMetadataPrefix)

	oldDenomMetaDataIter := oldDenomMetaDataStore.Iterator(nil, nil)
	defer oldDenomMetaDataIter.Close()

	for ; oldDenomMetaDataIter.Valid(); oldDenomMetaDataIter.Next() {
		oldKey := oldDenomMetaDataIter.Key()
		l := len(oldKey) / 2

		newKey := make([]byte, len(types.DenomMetadataPrefix)+l)
		// old key: prefix_bytes | denom_bytes | denom_bytes
		copy(newKey, types.DenomMetadataPrefix)
		copy(newKey[len(types.DenomMetadataPrefix):], oldKey[:l])
		store.Set(newKey, oldDenomMetaDataIter.Value())
		oldDenomMetaDataStore.Delete(oldKey)
	}

	return nil
}

// Migrate_V046_4_To_V046_5 is a helper function to migrate chains from <=v0.46.4
// to v0.46.5 ONLY.
//
// IMPORTANT: Please do not use this function if you are upgrading to v0.46
// from <=v0.45.
//
// This function migrates the store in-place by fixing the bank denom bug
// discovered in https://github.com/cosmos/cosmos-sdk/pull/13821. It has been
// fixed in v0.46.5, but if your chain had already migrated to v0.46, then you
// can apply this patch (in a coordinated upgrade, e.g. in the upgrade handler)
// to fix the bank denom state.
//
// The store is expected to be the bank store, and not any prefixed substore.
func Migrate_V046_4_To_V046_5(store sdk.KVStore) error {
	denomStore := prefix.NewStore(store, DenomMetadataPrefix)

	iter := denomStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		oldKey := iter.Key()

		// In the previous bugged version, we took one character too long,
		// see this line diff:
		// https://github.com/cosmos/cosmos-sdk/commit/62443b8c28a23efe43df2158aa2833c02c42af16#diff-d4d8a522eca0bd1fd052a756b80d0a50bff7bd8e487105221475eb78e232b46aR83
		//
		// Therefore we trim the last byte.
		newKey := oldKey[:len(oldKey)-1]

		denomStore.Set(newKey, iter.Value())
		denomStore.Delete(oldKey)
	}

	return nil
}
