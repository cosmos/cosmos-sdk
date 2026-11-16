package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"

	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	"github.com/cosmos/cosmos-sdk/x/feegrant/module"
)

// TestMigrate2to3EmptyQueueValue covers a fee allowance whose expiry queue
// entry has an empty value, as written by the v1 -> v2 migration and by the
// standalone cosmossdk.io/x/feegrant module. RemoveExpiredAllowances cannot
// decode that entry once it expires, which fails the EndBlocker, until
// Migrate2to3 rewrites the value.
func TestMigrate2to3EmptyQueueValue(t *testing.T) {
	key := storetypes.NewKVStoreKey(feegrant.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModuleBasic{})
	k := keeper.NewKeeper(encCfg.Codec, runtime.NewKVStoreService(key), nil)
	ctx := testCtx.Ctx

	addrs := simtestutil.CreateIncrementalAccounts(2)
	granter, grantee := addrs[0], addrs[1]
	exp := ctx.BlockTime().Add(time.Hour)

	grant, err := feegrant.NewGrant(granter, grantee, &feegrant.BasicAllowance{Expiration: &exp})
	require.NoError(t, err)
	require.NoError(t, k.FeeAllowance.Set(ctx, collections.Join(grantee, granter), grant))

	queueKey := collections.Join3(exp, grantee, granter)
	queueKeyBz, err := collections.EncodeKeyWithPrefix(feegrant.FeeAllowanceQueueKeyPrefix, k.FeeAllowanceQueue.KeyCodec(), queueKey)
	require.NoError(t, err)
	store := ctx.KVStore(key)
	store.Set(queueKeyBz, []byte{})

	ctx = ctx.WithBlockTime(exp.Add(time.Second))

	err = k.RemoveExpiredAllowances(ctx, 200)
	require.ErrorIs(t, err, collections.ErrEncoding)
	require.ErrorContains(t, err, "wanted size to be at least 1")

	require.NoError(t, keeper.NewMigrator(k).Migrate2to3(ctx))
	require.Equal(t, []byte{0x01}, store.Get(queueKeyBz))

	require.NoError(t, k.RemoveExpiredAllowances(ctx, 200))

	has, err := k.FeeAllowance.Has(ctx, collections.Join(grantee, granter))
	require.NoError(t, err)
	require.False(t, has)

	has, err = k.FeeAllowanceQueue.Has(ctx, queueKey)
	require.NoError(t, err)
	require.False(t, has)
	require.False(t, store.Has(queueKeyBz))
}
