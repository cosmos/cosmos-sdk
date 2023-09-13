package v2_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	v1 "github.com/cosmos/cosmos-sdk/x/slashing/migrations/v1"
	v2 "github.com/cosmos/cosmos-sdk/x/slashing/migrations/v2"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

func TestStoreMigration(t *testing.T) {
	slashingKey := storetypes.NewKVStoreKey("slashing")
	ctx := testutil.DefaultContext(slashingKey, storetypes.NewTransientStoreKey("transient_test"))
	storeService := runtime.NewKVStoreService(slashingKey)
	store := storeService.OpenKVStore(ctx)

	_, _, addr1 := testdata.KeyTestPubAddr()
	consAddr := sdk.ConsAddress(addr1)
	// Use dummy value for all keys.
	value := []byte("foo")

	testCases := []struct {
		name   string
		oldKey []byte
		newKey []byte
	}{
		{
			"ValidatorSigningInfoKey",
			v1.ValidatorSigningInfoKey(consAddr),
			types.ValidatorSigningInfoKey(consAddr),
		},
		{
			"ValidatorMissedBlockBitArrayKey",
			v1.ValidatorMissedBlockBitArrayKey(consAddr, 2),
			v2.ValidatorMissedBlockBitArrayKey(consAddr, 2),
		},
		{
			"AddrPubkeyRelationKey",
			v1.AddrPubkeyRelationKey(consAddr),
			addrPubkeyRelationKey(consAddr),
		},
	}

	// Set all the old keys to the store
	for _, tc := range testCases {
		err := store.Set(tc.oldKey, value)
		require.NoError(t, err)
	}

	// Run migrations.
	err := v2.MigrateStore(ctx, storeService)
	require.NoError(t, err)

	// Make sure the new keys are set and old keys are deleted.
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if !bytes.Equal(tc.oldKey, tc.newKey) {
				v, err := store.Get(tc.oldKey)
				require.NoError(t, err)
				require.Nil(t, v)
			}
			v, err := store.Get(tc.newKey)
			require.NoError(t, err)
			require.Equal(t, value, v)
		})
	}
}

func addrPubkeyRelationKey(addr []byte) []byte {
	return append(types.AddrPubkeyRelationKeyPrefix, address.MustLengthPrefix(addr)...)
}
