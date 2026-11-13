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

	require.NoError(t, statecheck.Check(ctx, modules))

	queueKeyBz, err := collections.EncodeKeyWithPrefix(feegrant.FeeAllowanceQueueKeyPrefix, k.FeeAllowanceQueue.KeyCodec(), queueKey)
	require.NoError(t, err)
	ctx.KVStore(key).Set(queueKeyBz, []byte{})

	err = statecheck.Check(ctx, modules)
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

	err = statecheck.Check(ctx, []statecheck.Module{{Name: feegrant.ModuleName, StoreKey: key, Schema: k.Schema}})
	require.ErrorIs(t, err, collections.ErrEncoding)
	require.ErrorContains(t, err, "feegrant/allowances_queue")
}
