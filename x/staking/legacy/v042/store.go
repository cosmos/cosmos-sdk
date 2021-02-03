package v042

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	v040auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v040"
	v042distribution "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v042"
	v040staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v040"
)

// migratePrefixAddressAddressAddress is a helper function that migrates all keys of format:
// prefix_bytes | address_1_bytes | address_2_bytes | address_3_bytes
// into format:
// prefix_bytes | address_1_len (1 byte) | address_1_bytes | address_2_len (1 byte) | address_2_bytes | address_3_len (1 byte) | address_3_bytes
func migratePrefixAddressAddressAddress(store sdk.KVStore, prefixBz []byte) {
	oldStore := prefix.NewStore(store, prefixBz)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		addr1 := oldStoreIter.Key()[:v040auth.AddrLen]
		addr2 := oldStoreIter.Key()[v040auth.AddrLen : 2*v040auth.AddrLen]
		addr3 := oldStoreIter.Key()[2*v040auth.AddrLen:]
		newStoreKey := append(append(append(
			prefixBz,
			address.MustLengthPrefix(addr1)...), address.MustLengthPrefix(addr2)...), address.MustLengthPrefix(addr3)...,
		)

		// Set new key on store. Values don't change.
		store.Set(newStoreKey, oldStoreIter.Value())
		oldStore.Delete(oldStoreIter.Key())
	}
}

const powerBytesLen = 8

func migrateValidatorsByPowerIndexKey(store sdk.KVStore) {
	oldStore := prefix.NewStore(store, v040staking.ValidatorsByPowerIndexKey)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		powerBytes := oldStoreIter.Key()[:powerBytesLen]
		valAddr := oldStoreIter.Key()[powerBytesLen:]
		newStoreKey := append(append(ValidatorsByPowerIndexKey, powerBytes...), address.MustLengthPrefix(valAddr)...)

		// Set new key on store. Values don't change.
		store.Set(newStoreKey, oldStoreIter.Value())
		oldStore.Delete(oldStoreIter.Key())
	}
}

// MigrateStore performs in-place store migrations from v0.40 to v0.42. The
// migration includes:
//
// - Change addresses to be length-prefixed.
func MigrateStore(store sdk.KVStore) error {
	v042distribution.MigratePrefixAddress(store, v040staking.LastValidatorPowerKey)

	v042distribution.MigratePrefixAddress(store, v040staking.ValidatorsKey)
	v042distribution.MigratePrefixAddress(store, v040staking.ValidatorsByConsAddrKey)
	migrateValidatorsByPowerIndexKey(store)

	v042distribution.MigratePrefixAddressAddress(store, v040staking.DelegationKey)
	v042distribution.MigratePrefixAddressAddress(store, v040staking.UnbondingDelegationKey)
	v042distribution.MigratePrefixAddressAddress(store, v040staking.UnbondingDelegationByValIndexKey)
	migratePrefixAddressAddressAddress(store, v040staking.RedelegationKey)
	migratePrefixAddressAddressAddress(store, v040staking.RedelegationByValSrcIndexKey)
	migratePrefixAddressAddressAddress(store, v040staking.RedelegationByValDstIndexKey)

	return nil
}
