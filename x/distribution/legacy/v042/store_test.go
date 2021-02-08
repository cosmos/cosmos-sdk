package v042_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v040distribution "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v040"
	v042distribution "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v042"
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
			v040distribution.FeePoolKey,
			v042distribution.FeePoolKey,
		},
		{
			"ProposerKey",
			v040distribution.ProposerKey,
			v042distribution.ProposerKey,
		},
		{
			"ValidatorOutstandingRewards",
			v040distribution.GetValidatorOutstandingRewardsKey(valAddr),
			v042distribution.GetValidatorOutstandingRewardsKey(valAddr),
		},
		{
			"DelegatorWithdrawAddr",
			v040distribution.GetDelegatorWithdrawAddrKey(addr2),
			v042distribution.GetDelegatorWithdrawAddrKey(addr2),
		},
		{
			"DelegatorStartingInfo",
			v040distribution.GetDelegatorStartingInfoKey(valAddr, addr2),
			v042distribution.GetDelegatorStartingInfoKey(valAddr, addr2),
		},
		{
			"ValidatorHistoricalRewards",
			v040distribution.GetValidatorHistoricalRewardsKey(valAddr, 6),
			v042distribution.GetValidatorHistoricalRewardsKey(valAddr, 6),
		},
		{
			"ValidatorCurrentRewards",
			v040distribution.GetValidatorCurrentRewardsKey(valAddr),
			v042distribution.GetValidatorCurrentRewardsKey(valAddr),
		},
		{
			"ValidatorAccumulatedCommission",
			v040distribution.GetValidatorAccumulatedCommissionKey(valAddr),
			v042distribution.GetValidatorAccumulatedCommissionKey(valAddr),
		},
		{
			"ValidatorSlashEvent",
			v040distribution.GetValidatorSlashEventKey(valAddr, 6, 8),
			v042distribution.GetValidatorSlashEventKey(valAddr, 6, 8),
		},
	}

	// Set all the old keys to the store
	for _, tc := range testCases {
		store.Set(tc.oldKey, value)
	}

	// Run migrations.
	err := v042distribution.MigrateStore(ctx, distributionKey, nil)
	require.NoError(t, err)

	// Make sure the new keys are set and old keys are deleted.
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if bytes.Compare(tc.oldKey, tc.newKey) != 0 {
				require.Nil(t, store.Get(tc.oldKey))
			}
			require.Equal(t, value, store.Get(tc.newKey))
		})
	}
}
