package epochs_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/epochs"
	"github.com/cosmos/cosmos-sdk/x/epochs/keeper"
	"github.com/cosmos/cosmos-sdk/x/epochs/types"
)

type testEpochHooks struct{}

func (h testEpochHooks) AfterEpochEnd(ctx context.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}

func (h testEpochHooks) BeforeEpochStart(ctx context.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}

func TestInvokeSetHooks(t *testing.T) {
	// Create a mock keeper
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	encCfg := testutil.MakeTestEncodingConfig()
	mockKeeper := keeper.NewKeeper(storeService, encCfg.Codec)

	// Create mock hooks
	hook1 := types.EpochHooksWrapper{
		EpochHooks: testEpochHooks{},
	}
	hook2 := types.EpochHooksWrapper{
		EpochHooks: testEpochHooks{},
	}
	hooks := map[string]types.EpochHooksWrapper{
		"moduleA": hook1,
		"moduleB": hook2,
	}

	// Call InvokeSetHooks
	err := epochs.InvokeSetHooks(&mockKeeper, hooks)
	require.NoError(t, err)

	// Verify that hooks were set correctly
	require.NotNil(t, mockKeeper.Hooks())
	require.IsType(t, types.MultiEpochHooks{}, mockKeeper.Hooks())

	// Verify the order of hooks (lexical order by module name)
	multiHooks := mockKeeper.Hooks().(types.MultiEpochHooks)
	require.Equal(t, 2, len(multiHooks))
	require.Equal(t, hook1, multiHooks[0])
	require.Equal(t, hook2, multiHooks[1])
}
