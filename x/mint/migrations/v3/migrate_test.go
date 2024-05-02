package v3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"
	v3 "cosmossdk.io/x/mint/migrations/v3"
	"cosmossdk.io/x/mint/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
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
	require.Equal(t, newLastReductionEpoch, int64(1))
}
