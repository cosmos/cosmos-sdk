package v043_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v042distribution "github.com/cosmos/cosmos-sdk/x/distribution/migrations/v042"
	v043distribution "github.com/cosmos/cosmos-sdk/x/distribution/migrations/v043"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func TestStoreMigration(t *testing.T) {
	distributionKey := sdk.NewKVStoreKey("distribution")
	ctx := testutil.DefaultContext(distributionKey, sdk.NewTransientStoreKey("transient_test"))
	store := ctx.KVStore(distributionKey)

	_, _, addr1 := testdata.KeyTestPubAddr()
	valAddr := sdk.ValAddress(addr1)
	_, _, addr2 := testdata.KeyTestPubAddr()
	// Use dummy value for all keys.
	value := []byte("foo")

	testCases := []struct {
		name   string
		oldKey []byte
		newKey []byte
	}{
		{
			"FeePoolKey",
			v042distribution.FeePoolKey,
			types.FeePoolKey,
		},
		{
			"ProposerKey",
			v042distribution.ProposerKey,
			types.ProposerKey,
		},
		{
			"ValidatorOutstandingRewards",
			v042distribution.GetValidatorOutstandingRewardsKey(valAddr),
			types.GetValidatorOutstandingRewardsKey(valAddr),
		},
		{
			"DelegatorWithdrawAddr",
			v042distribution.GetDelegatorWithdrawAddrKey(addr2),
			types.GetDelegatorWithdrawAddrKey(addr2),
		},
		{
			"DelegatorStartingInfo",
			v042distribution.GetDelegatorStartingInfoKey(valAddr, addr2),
			types.GetDelegatorStartingInfoKey(valAddr, addr2),
		},
		{
			"ValidatorHistoricalRewards",
			v042distribution.GetValidatorHistoricalRewardsKey(valAddr, 6),
			types.GetValidatorHistoricalRewardsKey(valAddr, 6),
		},
		{
			"ValidatorCurrentRewards",
			v042distribution.GetValidatorCurrentRewardsKey(valAddr),
			types.GetValidatorCurrentRewardsKey(valAddr),
		},
		{
			"ValidatorAccumulatedCommission",
			v042distribution.GetValidatorAccumulatedCommissionKey(valAddr),
			types.GetValidatorAccumulatedCommissionKey(valAddr),
		},
		{
			"ValidatorSlashEvent",
			v042distribution.GetValidatorSlashEventKey(valAddr, 6, 8),
			types.GetValidatorSlashEventKey(valAddr, 6, 8),
		},
	}

	// Set all the old keys to the store
	for _, tc := range testCases {
		store.Set(tc.oldKey, value)
	}

	// Run migrations.
	err := v043distribution.MigrateStore(ctx, distributionKey)
	require.NoError(t, err)

	// Make sure the new keys are set and old keys are deleted.
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if !bytes.Equal(tc.oldKey, tc.newKey) {
				require.Nil(t, store.Get(tc.oldKey))
			}
			require.Equal(t, value, store.Get(tc.newKey))
		})
	}
}
