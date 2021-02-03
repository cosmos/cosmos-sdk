package v042

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v040distribution "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v040"
)

// MigrateStore performs in-place store migrations from v0.40 to v0.42. The
// migration includes:
//
// - Change addresses to be length-prefixed.
func MigrateStore(store sdk.KVStore) error {
	MigratePrefixAddress(store, v040distribution.ValidatorOutstandingRewardsPrefix)
	MigratePrefixAddress(store, v040distribution.DelegatorWithdrawAddrPrefix)
	MigratePrefixAddressAddress(store, v040distribution.DelegatorStartingInfoPrefix)
	MigratePrefixAddressBytes(store, v040distribution.ValidatorHistoricalRewardsPrefix)
	MigratePrefixAddress(store, v040distribution.ValidatorCurrentRewardsPrefix)
	MigratePrefixAddress(store, v040distribution.ValidatorAccumulatedCommissionPrefix)
	MigratePrefixAddressBytes(store, v040distribution.ValidatorSlashEventPrefix)

	return nil
}
