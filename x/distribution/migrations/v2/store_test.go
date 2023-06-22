package v2_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	v1 "github.com/cosmos/cosmos-sdk/x/distribution/migrations/v1"
	v2 "github.com/cosmos/cosmos-sdk/x/distribution/migrations/v2"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func TestStoreMigration(t *testing.T) {
	distributionKey := storetypes.NewKVStoreKey("distribution")
	storeService := runtime.NewKVStoreService(distributionKey)
	ctx := testutil.DefaultContext(distributionKey, storetypes.NewTransientStoreKey("transient_test"))
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
			v1.FeePoolKey,
			types.FeePoolKey,
		},
		{
			"ProposerKey",
			v1.ProposerKey,
			types.ProposerKey,
		},
		{
			"ValidatorOutstandingRewards",
			v1.GetValidatorOutstandingRewardsKey(valAddr),
			append(types.ValidatorOutstandingRewardsPrefix, address.MustLengthPrefix(valAddr.Bytes())...),
		},
		{
			"DelegatorWithdrawAddr",
			v1.GetDelegatorWithdrawAddrKey(addr2),
			append(types.DelegatorWithdrawAddrPrefix, address.MustLengthPrefix(addr2.Bytes())...),
		},
		{
			"DelegatorStartingInfo",
			v1.GetDelegatorStartingInfoKey(valAddr, addr2),
			append(append(types.DelegatorStartingInfoPrefix, address.MustLengthPrefix(valAddr.Bytes())...), address.MustLengthPrefix(addr2.Bytes())...),
		},
		{
			"ValidatorHistoricalRewards",
			v1.GetValidatorHistoricalRewardsKey(valAddr, 6),
			getValidatorHistoricalRewardsKey(valAddr, 6),
		},
		{
			"ValidatorCurrentRewards",
			v1.GetValidatorCurrentRewardsKey(valAddr),
			append(types.ValidatorCurrentRewardsPrefix, address.MustLengthPrefix(valAddr.Bytes())...),
		},
		{
			"ValidatorAccumulatedCommission",
			v1.GetValidatorAccumulatedCommissionKey(valAddr),
			append(types.ValidatorAccumulatedCommissionPrefix, address.MustLengthPrefix(valAddr.Bytes())...),
		},
		{
			"ValidatorSlashEvent",
			v1.GetValidatorSlashEventKey(valAddr, 6, 8),
			types.GetValidatorSlashEventKey(valAddr, 6, 8),
		},
	}

	// Set all the old keys to the store
	for _, tc := range testCases {
		store.Set(tc.oldKey, value)
	}

	// Run migrations.
	err := v2.MigrateStore(ctx, storeService)
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

// getValidatorHistoricalRewardsKey creates the key for a validator's historical rewards.
// TODO: remove me
func getValidatorHistoricalRewardsKey(v sdk.ValAddress, k uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, k)
	return append(append(types.ValidatorHistoricalRewardsPrefix, address.MustLengthPrefix(v.Bytes())...), b...)
}
