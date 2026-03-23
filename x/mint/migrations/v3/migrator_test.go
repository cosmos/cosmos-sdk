package v3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/mint"
	v3 "github.com/cosmos/cosmos-sdk/x/mint/migrations/v3"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

type mockSubspace struct {
	ps types.Params
}

func newMockSubspace(ps types.Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func TestMigrate(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(mint.AppModuleBasic{})
	cdc := encCfg.Codec

	storeKey := storetypes.NewKVStoreKey(types.ModuleName)
	tKey := storetypes.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	kvStoreService := runtime.NewKVStoreService(storeKey)
	store := kvStoreService.OpenKVStore(ctx)

	sb := collections.NewSchemaBuilder(kvStoreService)
	params := collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc))

	dp := newMockSubspace(types.DefaultParams())
	require.NoError(t, params.Set(ctx, dp.ps))
	require.NoError(t, v3.Migrate(ctx, store, cdc, params))

	var res types.Params
	bz, err := store.Get(types.ParamsKey)
	require.NoError(t, err)
	require.NoError(t, cdc.Unmarshal(bz, &res))
	require.Equal(t, dp.ps, res)
}
