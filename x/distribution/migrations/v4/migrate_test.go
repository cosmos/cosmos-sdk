package v4_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	v4 "github.com/cosmos/cosmos-sdk/x/distribution/migrations/v4"
	dstrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func TestMigrateStore(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{}).Codec
	storeKey := storetypes.NewKVStoreKey(v4.ModuleName)
	storeService := runtime.NewKVStoreService(storeKey)
	tKey := storetypes.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := ctx.KVStore(storeKey)

	sb := collections.NewSchemaBuilder(storeService)
	nakamotoBonusCollection := collections.NewItem(sb, v4.NakamotoBonusKey, "nakamoto_bonus", sdk.LegacyDecValue)

	var params dstrtypes.Params
	bz := store.Get(v4.ParamsKey)
	require.NoError(t, cdc.Unmarshal(bz, &params))
	require.NotNil(t, params)

	// Run migrations.
	err := v4.MigrateStore(ctx, storeService, cdc, nakamotoBonusCollection)
	require.NoError(t, err)

	// Check params
	bz = store.Get(v4.ParamsKey)
	require.NoError(t, cdc.Unmarshal(bz, &params))
	require.NotNil(t, params)
	require.Equal(t, dstrtypes.DefaultParams().NakamotoBonus, params.NakamotoBonus)

	// Check Nakamoto Bonus
	nbResult, err := nakamotoBonusCollection.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, v4.DefaultNakamotoBonus, nbResult)
}
