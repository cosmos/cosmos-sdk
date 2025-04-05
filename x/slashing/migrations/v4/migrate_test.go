package v4_test

import (
	"testing"

	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	v4 "github.com/cosmos/cosmos-sdk/x/slashing/migrations/v4"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

func TestMigrate(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(slashing.AppModuleBasic{}).Codec
	storeKey := storetypes.NewKVStoreKey(slashingtypes.ModuleName)
	tKey := storetypes.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := ctx.KVStore(storeKey)
	params := slashingtypes.Params{SignedBlocksWindow: 100}

	// use a total of 30 validators
	accounts := testutil.CreateKeyringAccounts(t, keyring.NewInMemory(cdc), 30)
	for _, acc := range accounts {
		consAddr := sdk.ConsAddress(acc.Address)

		// store old signing info and bitmap entries
		bz := cdc.MustMarshal(&slashingtypes.ValidatorSigningInfo{Address: consAddr.String()})
		store.Set(v4.ValidatorSigningInfoKey(consAddr), bz)

		for i := int64(0); i < params.SignedBlocksWindow; i++ {
			// all even blocks are missed
			missed := &gogotypes.BoolValue{Value: i%2 == 0}
			bz := cdc.MustMarshal(missed)
			store.Set(v4.ValidatorMissedBlockBitArrayKey(consAddr, i), bz)
		}
	}

	err := v4.Migrate(ctx, store, nil, 10)
	require.NoError(t, err)

	nextIndex := store.Get(v4.NextMigrateValidatorMissedBlocksKey)
	require.NotNil(t, nextIndex)
	require.NoError(t, v4.Migrate(ctx, store, nextIndex, 10), "v4.MigrateStore failed")

	nextIndex = store.Get(v4.NextMigrateValidatorMissedBlocksKey)
	require.NotNil(t, nextIndex)
	require.NoErrorf(t, v4.Migrate(ctx, store, nextIndex, 10), "v4.MigrateStore failed")

	// assert the next migration index is cleared from the store
	require.Nil(t, store.Get(v4.NextMigrateValidatorMissedBlocksKey))
}
