package v043

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v042distribution "github.com/cosmos/cosmos-sdk/x/distribution/migrations/v042"
)

// MigrateStore performs in-place store migrations from v0.40 to v0.43. The
// migration includes:
//
// - Change addresses to be length-prefixed.
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	store := ctx.KVStore(storeKey)
	MigratePrefixAddress(store, v042distribution.ValidatorOutstandingRewardsPrefix)
	MigratePrefixAddress(store, v042distribution.DelegatorWithdrawAddrPrefix)
	MigratePrefixAddressAddress(store, v042distribution.DelegatorStartingInfoPrefix)
	MigratePrefixAddressBytes(store, v042distribution.ValidatorHistoricalRewardsPrefix)
	MigratePrefixAddress(store, v042distribution.ValidatorCurrentRewardsPrefix)
	MigratePrefixAddress(store, v042distribution.ValidatorAccumulatedCommissionPrefix)
	MigratePrefixAddressBytes(store, v042distribution.ValidatorSlashEventPrefix)

	return nil
}
