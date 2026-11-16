package v3_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	v3 "github.com/cosmos/cosmos-sdk/x/feegrant/migrations/v3"
	"github.com/cosmos/cosmos-sdk/x/feegrant/module"
)

func TestMigrateStore(t *testing.T) {
	encodingConfig := moduletestutil.MakeTestEncodingConfig(module.AppModuleBasic{})
	feegrantKey := storetypes.NewKVStoreKey(v3.ModuleName)
	storeService := runtime.NewKVStoreService(feegrantKey)
	ctx := testutil.DefaultContext(feegrantKey, storetypes.NewTransientStoreKey("transient_test"))
	k := keeper.NewKeeper(encodingConfig.Codec, storeService, nil)

	granter := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	grantee1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	grantee2 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	grantee3 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	exp := ctx.BlockTime().Add(time.Hour)

	queueKey := func(grantee sdk.AccAddress) []byte {
		bz, err := collections.EncodeKeyWithPrefix(v3.FeeAllowanceQueueKeyPrefix, k.FeeAllowanceQueue.KeyCodec(), collections.Join3(exp, grantee, granter))
		require.NoError(t, err)
		return bz
	}
	emptyKey1, emptyKey2, falseKey := queueKey(grantee1), queueKey(grantee2), queueKey(grantee3)
	// an entry outside the queue prefix that also has an empty value
	otherKey := append(append([]byte{}, v3.FeeAllowanceKeyPrefix...), granter...)

	store := ctx.KVStore(feegrantKey)
	store.Set(emptyKey1, []byte{})
	store.Set(emptyKey2, []byte{})
	store.Set(falseKey, []byte{0x00})
	store.Set(otherKey, []byte{})

	_, err := k.FeeAllowanceQueue.Get(ctx, collections.Join3(exp, grantee1, granter))
	require.ErrorIs(t, err, collections.ErrEncoding)

	require.NoError(t, v3.MigrateStore(ctx, storeService))

	// empty values are set to the encoding of true
	require.Equal(t, []byte{0x01}, store.Get(emptyKey1))
	require.Equal(t, []byte{0x01}, store.Get(emptyKey2))
	for _, grantee := range []sdk.AccAddress{grantee1, grantee2} {
		value, err := k.FeeAllowanceQueue.Get(ctx, collections.Join3(exp, grantee, granter))
		require.NoError(t, err)
		require.True(t, value)
	}

	// non-empty values are left unchanged
	require.Equal(t, []byte{0x00}, store.Get(falseKey))

	// entries under other prefixes are left unchanged
	require.True(t, store.Has(otherKey))
	require.Empty(t, store.Get(otherKey))

	// running the migration again changes nothing
	before := dumpStore(t, store)
	require.NoError(t, v3.MigrateStore(ctx, storeService))
	require.Equal(t, before, dumpStore(t, store))
}

func TestMigrateStoreEmpty(t *testing.T) {
	feegrantKey := storetypes.NewKVStoreKey(v3.ModuleName)
	ctx := testutil.DefaultContext(feegrantKey, storetypes.NewTransientStoreKey("transient_test"))

	require.NoError(t, v3.MigrateStore(ctx, runtime.NewKVStoreService(feegrantKey)))
	require.Empty(t, dumpStore(t, ctx.KVStore(feegrantKey)))
}

func dumpStore(t *testing.T, store storetypes.KVStore) map[string][]byte {
	t.Helper()
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	entries := make(map[string][]byte)
	for ; iterator.Valid(); iterator.Next() {
		entries[string(iterator.Key())] = iterator.Value()
	}
	return entries
}
