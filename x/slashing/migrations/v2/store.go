package v2

import (
	"context"

	storetypes "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/runtime"
	v1 "github.com/cosmos/cosmos-sdk/x/slashing/migrations/v1"
	v2staking "github.com/cosmos/cosmos-sdk/x/staking/migrations/v2"
)

// MigrateStore performs in-place store migrations from v0.40 to v0.43. The
// migration includes:
//
// - Change addresses to be length-prefixed.
func MigrateStore(ctx context.Context, storeService storetypes.KVStoreService) error {
	store := runtime.KVStoreAdapter(storeService.OpenKVStore(ctx))
	v2staking.MigratePrefixAddress(store, v1.ValidatorSigningInfoKeyPrefix)
	v2staking.MigratePrefixAddressBytes(store, v1.ValidatorMissedBlockBitArrayKeyPrefix)
	v2staking.MigratePrefixAddress(store, v1.AddrPubkeyRelationKeyPrefix)

	return nil
}
