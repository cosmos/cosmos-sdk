package v2_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec/address"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkaddress "github.com/cosmos/cosmos-sdk/types/address"
	v1 "github.com/cosmos/cosmos-sdk/x/staking/migrations/v1"
	v2 "github.com/cosmos/cosmos-sdk/x/staking/migrations/v2"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestStoreMigration(t *testing.T) {
	stakingKey := storetypes.NewKVStoreKey("staking")
	tStakingKey := storetypes.NewTransientStoreKey("transient_test")
	ctx := sdktestutil.DefaultContext(stakingKey, tStakingKey)
	store := ctx.KVStore(stakingKey)

	_, pk1, addr1 := testdata.KeyTestPubAddr()
	valAddr1 := sdk.ValAddress(addr1)
	val := testutil.NewValidator(t, valAddr1, pk1)
	_, _, addr2 := testdata.KeyTestPubAddr()
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
			v1.GetLastValidatorPowerKey(valAddr1),
			getLastValidatorPowerKey(valAddr1),
		},
		{
			"LastTotalPowerKey",
			v1.LastTotalPowerKey,
			types.LastTotalPowerKey,
		},
		{
			"ValidatorsKey",
			v1.GetValidatorKey(valAddr1),
			getValidatorKey(valAddr1),
		},
		{
			"ValidatorsByConsAddrKey",
			v1.GetValidatorByConsAddrKey(consAddr),
			v2.GetValidatorByConsAddrKey(consAddr),
		},
		{
			"ValidatorsByPowerIndexKey",
			v1.GetValidatorsByPowerIndexKey(val),
			types.GetValidatorsByPowerIndexKey(val, sdk.DefaultPowerReduction, address.NewBech32Codec("cosmosvaloper")),
		},
		{
			"DelegationKey",
			v1.GetDelegationKey(addr4, valAddr1),
			v2.GetDelegationKey(addr4, valAddr1),
		},
		{
			"UnbondingDelegationKey",
			v1.GetUBDKey(addr4, valAddr1),
			unbondingKey(addr4, valAddr1),
		},
		{
			"UnbondingDelegationByValIndexKey",
			v1.GetUBDByValIndexKey(addr4, valAddr1),
			getUBDByValIndexKey(addr4, valAddr1),
		},
		{
			"RedelegationKey",
			v1.GetREDKey(addr4, valAddr1, valAddr2),
			v2.GetREDKey(addr4, valAddr1, valAddr2),
		},
		{
			"RedelegationByValSrcIndexKey",
			v1.GetREDByValSrcIndexKey(addr4, valAddr1, valAddr2),
			v2.GetREDByValSrcIndexKey(addr4, valAddr1, valAddr2),
		},
		{
			"RedelegationByValDstIndexKey",
			v1.GetREDByValDstIndexKey(addr4, valAddr1, valAddr2),
			v2.GetREDByValDstIndexKey(addr4, valAddr1, valAddr2),
		},
		{
			"UnbondingQueueKey",
			v1.GetUnbondingDelegationTimeKey(now),
			getUnbondingDelegationTimeKey(now),
		},
		{
			"RedelegationQueueKey",
			v1.GetRedelegationTimeKey(now),
			types.GetRedelegationTimeKey(now),
		},
		{
			"ValidatorQueueKey",
			v1.GetValidatorQueueKey(now, 4),
			types.GetValidatorQueueKey(now, 4),
		},
		{
			"HistoricalInfoKey",
			v1.GetHistoricalInfoKey(4),
			v2.GetHistoricalInfoKey(4),
		},
	}

	// Set all the old keys to the store
	for _, tc := range testCases {
		store.Set(tc.oldKey, value)
	}

	// Run migrations.
	err := v2.MigrateStore(ctx, store)
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

func getLastValidatorPowerKey(operator sdk.ValAddress) []byte {
	return append(types.LastValidatorPowerKey, sdkaddress.MustLengthPrefix(operator)...)
}

func getUBDByValIndexKey(delAddr sdk.AccAddress, valAddr sdk.ValAddress) []byte {
	return append(append(types.UnbondingDelegationByValIndexKey, sdkaddress.MustLengthPrefix(valAddr)...), sdkaddress.MustLengthPrefix(delAddr)...)
}

func getUnbondingDelegationTimeKey(timestamp time.Time) []byte {
	bz := sdk.FormatTimeBytes(timestamp)
	return append(types.UnbondingQueueKey, bz...)
}

func getValidatorKey(operatorAddr sdk.ValAddress) []byte {
	return append(types.ValidatorsKey, sdkaddress.MustLengthPrefix(operatorAddr)...)
}

func unbondingKey(delAddr sdk.AccAddress, valAddr sdk.ValAddress) []byte {
	return append(append(types.UnbondingDelegationKey, sdkaddress.MustLengthPrefix(delAddr)...), sdkaddress.MustLengthPrefix(valAddr)...)
}
