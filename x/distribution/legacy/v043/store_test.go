package v043_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v043distribution "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v043"
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
			v043distribution.FeePoolKey,
			types.FeePoolKey,
		},
		{
			"ProposerKey",
			v043distribution.ProposerKey,
			types.ProposerKey,
		},
		{
			"ValidatorOutstandingRewards",
			v043distribution.GetValidatorOutstandingRewardsKey(valAddr),
			types.GetValidatorOutstandingRewardsKey(valAddr),
		},
		{
			"DelegatorWithdrawAddr",
			v043distribution.GetDelegatorWithdrawAddrKey(addr2),
			types.GetDelegatorWithdrawAddrKey(addr2),
		},
		{
			"DelegatorStartingInfo",
			v043distribution.GetDelegatorStartingInfoKey(valAddr, addr2),
			types.GetDelegatorStartingInfoKey(valAddr, addr2),
		},
		{
			"ValidatorHistoricalRewards",
			v043distribution.GetValidatorHistoricalRewardsKey(valAddr, 6),
			types.GetValidatorHistoricalRewardsKey(valAddr, 6),
		},
		{
			"ValidatorCurrentRewards",
			v043distribution.GetValidatorCurrentRewardsKey(valAddr),
			types.GetValidatorCurrentRewardsKey(valAddr),
		},
		{
			"ValidatorAccumulatedCommission",
			v043distribution.GetValidatorAccumulatedCommissionKey(valAddr),
			types.GetValidatorAccumulatedCommissionKey(valAddr),
		},
		{
			"ValidatorSlashEvent",
			v043distribution.GetValidatorSlashEventKey(valAddr, 6, 8),
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
			if bytes.Compare(tc.oldKey, tc.newKey) != 0 {
				require.Nil(t, store.Get(tc.oldKey))
			}
			require.Equal(t, value, store.Get(tc.newKey))
		})
	}
}
