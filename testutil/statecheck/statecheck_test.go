package statecheck_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"

	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/testutil/statecheck"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	"github.com/cosmos/cosmos-sdk/x/feegrant/module"
)

// TestCheckFeegrantQueueEmptyValue writes a fee allowance queue entry with an
// empty value, the encoding older feegrant code wrote, and requires Check to
// report it. collections.BoolValue cannot decode an empty value.
func TestCheckFeegrantQueueEmptyValue(t *testing.T) {
	key := storetypes.NewKVStoreKey(feegrant.StoreKey)
	ctx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test")).Ctx
	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModuleBasic{})
	k := keeper.NewKeeper(encCfg.Codec, runtime.NewKVStoreService(key), nil)
	modules := []statecheck.Module{{Name: feegrant.ModuleName, StoreKey: key, Schema: k.Schema}}

	addrs := simtestutil.CreateIncrementalAccounts(2)
	granter, grantee := addrs[0], addrs[1]
	exp := ctx.BlockTime().Add(time.Hour)
	grant, err := feegrant.NewGrant(granter, grantee, &feegrant.BasicAllowance{Expiration: &exp})
	require.NoError(t, err)
	require.NoError(t, k.FeeAllowance.Set(ctx, collections.Join(grantee, granter), grant))
	queueKey := collections.Join3(exp, grantee, granter)
	require.NoError(t, k.FeeAllowanceQueue.Set(ctx, queueKey, true))

	_, err = statecheck.Check(ctx, modules)
	require.NoError(t, err)

	queueKeyBz, err := collections.EncodeKeyWithPrefix(feegrant.FeeAllowanceQueueKeyPrefix, k.FeeAllowanceQueue.KeyCodec(), queueKey)
	require.NoError(t, err)
	ctx.KVStore(key).Set(queueKeyBz, []byte{})

	_, err = statecheck.Check(ctx, modules)
	require.ErrorIs(t, err, collections.ErrEncoding)
	require.ErrorContains(t, err, "feegrant/allowances_queue")
	require.ErrorContains(t, err, "wanted size to be at least 1")
}

// TestCheckKeyTrailingBytes requires Check to report a key that its key codec
// decodes without consuming every byte.
func TestCheckKeyTrailingBytes(t *testing.T) {
	key := storetypes.NewKVStoreKey(feegrant.StoreKey)
	ctx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test")).Ctx
	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModuleBasic{})
	k := keeper.NewKeeper(encCfg.Codec, runtime.NewKVStoreService(key), nil)

	addrs := simtestutil.CreateIncrementalAccounts(2)
	queueKey := collections.Join3(ctx.BlockTime(), addrs[1], addrs[0])
	queueKeyBz, err := collections.EncodeKeyWithPrefix(feegrant.FeeAllowanceQueueKeyPrefix, k.FeeAllowanceQueue.KeyCodec(), queueKey)
	require.NoError(t, err)
	ctx.KVStore(key).Set(append(queueKeyBz, 0xff), []byte{0x01})

	_, err = statecheck.Check(ctx, []statecheck.Module{{Name: feegrant.ModuleName, StoreKey: key, Schema: k.Schema}})
	require.ErrorIs(t, err, collections.ErrEncoding)
	require.ErrorContains(t, err, "feegrant/allowances_queue")
}

// TestCheckEmptySchema requires Check to fail for a module whose schema lists
// no collections, since none of its entries could be checked.
func TestCheckEmptySchema(t *testing.T) {
	key := storetypes.NewKVStoreKey(feegrant.StoreKey)
	ctx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test")).Ctx

	_, err := statecheck.Check(ctx, []statecheck.Module{{Name: feegrant.ModuleName, StoreKey: key}})
	require.ErrorContains(t, err, "feegrant: schema lists no collections")
}

// TestCheckRawPrefixes requires Check to fail on an entry outside every
// collection prefix, to count it instead when it falls under a raw prefix of
// the module, and to still fail when the module lists only other raw prefixes.
func TestCheckRawPrefixes(t *testing.T) {
	key := storetypes.NewKVStoreKey(feegrant.StoreKey)
	ctx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test")).Ctx
	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModuleBasic{})
	k := keeper.NewKeeper(encCfg.Codec, runtime.NewKVStoreService(key), nil)
	ctx.KVStore(key).Set([]byte{0x7f, 0x01}, []byte{0x01})

	m := statecheck.Module{Name: feegrant.ModuleName, StoreKey: key, Schema: k.Schema}
	_, err := statecheck.Check(ctx, []statecheck.Module{m})
	require.ErrorContains(t, err, "feegrant: key 7F01 is outside every collection prefix and every raw prefix")

	m.RawPrefixes = [][]byte{{0x7e}}
	_, err = statecheck.Check(ctx, []statecheck.Module{m})
	require.ErrorContains(t, err, "feegrant: key 7F01 is outside every collection prefix and every raw prefix")

	m.RawPrefixes = [][]byte{{0x7e}, {0x7f}}
	res, err := statecheck.Check(ctx, []statecheck.Module{m})
	require.NoError(t, err)
	require.Equal(t, []statecheck.Result{{Name: feegrant.ModuleName, Checked: 0, Raw: 1}}, res)
}

// TestCodecTypes requires KeyType and CodecTypes to read the codecs of a map,
// a key set and an item out of a built schema.
func TestCodecTypes(t *testing.T) {
	sb := collections.NewSchemaBuilder(runtime.NewKVStoreService(storetypes.NewKVStoreKey("codectypes")))
	collections.NewMap(sb, collections.NewPrefix(1), "map", collections.StringKey, collections.Uint64Value)
	collections.NewKeySet(sb, collections.NewPrefix(2), "set", collections.PairKeyCodec(collections.StringKey, collections.Uint64Key))
	collections.NewItem(sb, collections.NewPrefix(3), "item", collections.BoolValue)
	schema, err := sb.Build()
	require.NoError(t, err)

	type codecs struct{ keyType, key, value string }
	got := map[string]codecs{}
	for _, coll := range schema.ListCollections() {
		key, value := statecheck.CodecTypes(coll)
		got[coll.GetName()] = codecs{statecheck.KeyType(coll), key, value}
	}
	require.Equal(t, map[string]codecs{
		"map":  {"string", "codec.stringKey[string]", "codec.keyToValueCodec[uint64]"},
		"set":  {"Pair[string, uint64]", "collections.pairKeyCodec[string,uint64]", "collections.NoValue"},
		"item": {"no_key", "collections.noKey", "codec.keyToValueCodec[bool]"},
	}, got)
}
