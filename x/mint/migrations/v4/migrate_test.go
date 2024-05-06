package v4_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/mint/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	v3 "github.com/cosmos/cosmos-sdk/x/gov/migrations/v3"
)

func TestMigrateStore(t *testing.T) {
	mintKey := storetypes.NewKVStoreKey("mint")
	ctx := testutil.DefaultContext(mintKey, storetypes.NewTransientStoreKey("transient_test"))
	storeService := runtime.NewKVStoreService(mintKey)
	sb := collections.NewSchemaBuilder(storeService)
	lastReductionEpoch := collections.NewItem(sb, types.LastReductionEpochKey, "last_reduction_epoch", collections.Int64Value)

	err := lastReductionEpoch.Set(ctx, 1)
	require.NoError(t, err)

	// Run migrations.
	err = v3.MigrateStore(ctx, lastReductionEpoch)
	require.NoError(t, err)

	newLastReductionEpoch, err := lastReductionEpoch.Get(ctx)
	require.NoError(t, err)
	// check that new LastReductionEpoch equals default value `0` after migration.
	require.Equal(t, newLastReductionEpoch, int64(0))
}
