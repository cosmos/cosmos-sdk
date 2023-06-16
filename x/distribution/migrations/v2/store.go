package v2

import (
	"cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/distribution/migrations/v1"
	v2staking "github.com/cosmos/cosmos-sdk/x/staking/migrations/v2"
)

// MigrateStore performs in-place store migrations from v0.40 to v0.43. The
// migration includes:
//
// - Change addresses to be length-prefixed.
func MigrateStore(ctx sdk.Context, storeService store.KVStoreService) error {
	store := runtime.KVStoreAdapter(storeService.OpenKVStore(ctx))
	v2staking.MigratePrefixAddress(store, v1.ValidatorOutstandingRewardsPrefix)
	v2staking.MigratePrefixAddress(store, v1.DelegatorWithdrawAddrPrefix)
	v2staking.MigratePrefixAddressAddress(store, v1.DelegatorStartingInfoPrefix)
	v2staking.MigratePrefixAddressBytes(store, v1.ValidatorHistoricalRewardsPrefix)
	v2staking.MigratePrefixAddress(store, v1.ValidatorCurrentRewardsPrefix)
	v2staking.MigratePrefixAddress(store, v1.ValidatorAccumulatedCommissionPrefix)
	v2staking.MigratePrefixAddressBytes(store, v1.ValidatorSlashEventPrefix)

	return nil
}
