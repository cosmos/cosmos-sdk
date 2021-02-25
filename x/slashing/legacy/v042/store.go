package v042

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v042distribution "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v042"
	v040slashing "github.com/cosmos/cosmos-sdk/x/slashing/legacy/v040"
)

// MigrateStore performs in-place store migrations from v0.40 to v0.42. The
// migration includes:
//
// - Change addresses to be length-prefixed.
func MigrateStore(ctx sdk.Context, storeKey sdk.StoreKey) error {
	store := ctx.KVStore(storeKey)
	v042distribution.MigratePrefixAddress(store, v040slashing.ValidatorSigningInfoKeyPrefix)
	v042distribution.MigratePrefixAddressBytes(store, v040slashing.ValidatorMissedBlockBitArrayKeyPrefix)
	v042distribution.MigratePrefixAddress(store, v040slashing.AddrPubkeyRelationKeyPrefix)

	return nil
}
