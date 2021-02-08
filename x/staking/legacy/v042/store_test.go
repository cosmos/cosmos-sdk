package v042_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v040staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v040"
	v042staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v042"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
)

func TestStoreMigration(t *testing.T) {
	stakingKey := sdk.NewKVStoreKey("staking")
	ctx := testutil.DefaultContext(stakingKey, sdk.NewTransientStoreKey("transient_test"))
	store := ctx.KVStore(stakingKey)

	_, pk1, addr1 := testdata.KeyTestPubAddr()
	valAddr1 := sdk.ValAddress(addr1)
	val := teststaking.NewValidator(t, valAddr1, pk1)
	_, pk1, addr2 := testdata.KeyTestPubAddr()
	valAddr2 := sdk.ValAddress(addr2)
	_, _, addr3 := testdata.KeyTestPubAddr()
	consAddr := sdk.ConsAddress(addr3.String())
	_, _, addr4 := testdata.KeyTestPubAddr()
	now := time.Now()
	// Use dummy value for all keys.
	value := []byte("foo")

	testCases := []struct {
		name   string
		oldKey []byte
		newKey []byte
	}{
		{
			"LastValidatorPowerKey",
			v040staking.GetLastValidatorPowerKey(valAddr1),
			v042staking.GetLastValidatorPowerKey(valAddr1),
		},
		{
			"LastTotalPowerKey",
			v040staking.LastTotalPowerKey,
			v042staking.LastTotalPowerKey,
		},
		{
			"ValidatorsKey",
			v040staking.GetValidatorKey(valAddr1),
			v042staking.GetValidatorKey(valAddr1),
		},
		{
			"ValidatorsByConsAddrKey",
			v040staking.GetValidatorByConsAddrKey(consAddr),
			v042staking.GetValidatorByConsAddrKey(consAddr),
		},
		{
			"ValidatorsByPowerIndexKey",
			v040staking.GetValidatorsByPowerIndexKey(val),
			v042staking.GetValidatorsByPowerIndexKey(val),
		},
		{
			"DelegationKey",
			v040staking.GetDelegationKey(addr4, valAddr1),
			v042staking.GetDelegationKey(addr4, valAddr1),
		},
		{
			"UnbondingDelegationKey",
			v040staking.GetUBDKey(addr4, valAddr1),
			v042staking.GetUBDKey(addr4, valAddr1),
		},
		{
			"UnbondingDelegationByValIndexKey",
			v040staking.GetUBDByValIndexKey(addr4, valAddr1),
			v042staking.GetUBDByValIndexKey(addr4, valAddr1),
		},
		{
			"RedelegationKey",
			v040staking.GetREDKey(addr4, valAddr1, valAddr2),
			v042staking.GetREDKey(addr4, valAddr1, valAddr2),
		},
		{
			"RedelegationByValSrcIndexKey",
			v040staking.GetREDByValSrcIndexKey(addr4, valAddr1, valAddr2),
			v042staking.GetREDByValSrcIndexKey(addr4, valAddr1, valAddr2),
		},
		{
			"RedelegationByValDstIndexKey",
			v040staking.GetREDByValDstIndexKey(addr4, valAddr1, valAddr2),
			v042staking.GetREDByValDstIndexKey(addr4, valAddr1, valAddr2),
		},
		{
			"UnbondingQueueKey",
			v040staking.GetUnbondingDelegationTimeKey(now),
			v042staking.GetUnbondingDelegationTimeKey(now),
		},
		{
			"RedelegationQueueKey",
			v040staking.GetRedelegationTimeKey(now),
			v042staking.GetRedelegationTimeKey(now),
		},
		{
			"ValidatorQueueKey",
			v040staking.GetValidatorQueueKey(now, 4),
			v042staking.GetValidatorQueueKey(now, 4),
		},
		{
			"HistoricalInfoKey",
			v040staking.GetHistoricalInfoKey(4),
			v042staking.GetHistoricalInfoKey(4),
		},
	}

	// Set all the old keys to the store
	for _, tc := range testCases {
		store.Set(tc.oldKey, value)
	}

	// Run migrations.
	err := v042staking.MigrateStore(ctx, stakingKey, nil)
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
