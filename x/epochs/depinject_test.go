package epochs_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	bankmodulev1 "cosmossdk.io/api/cosmos/bank/module/v1"
	modulev1 "cosmossdk.io/api/cosmos/epochs/module/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/epochs"
	"github.com/cosmos/cosmos-sdk/x/epochs/keeper"
	"github.com/cosmos/cosmos-sdk/x/epochs/types"
)

var _ types.EpochHooks = testEpochHooks{}

type testEpochHooks struct{}

func (h testEpochHooks) AfterEpochEnd(
	ctx context.Context,
	epochIdentifier string,
	epochNumber int64,
) error {
	return nil
}

func (h testEpochHooks) BeforeEpochStart(
	ctx context.Context,
	epochIdentifier string,
	epochNumber int64,
) error {
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

type TestInputs struct {
	depinject.In
}

type TestOutputs struct {
	depinject.Out

	Hooks types.EpochHooksWrapper
}

func DummyProvider(in TestInputs) TestOutputs {
	return TestOutputs{
		Hooks: types.EpochHooksWrapper{
			EpochHooks: testEpochHooks{},
		},
	}
}

func ProvideDeps(depinject.In) struct {
	depinject.Out
	Cdc          codec.Codec
	StoreService store.KVStoreService
} {
	encCfg := testutil.MakeTestEncodingConfig()

	key := storetypes.NewKVStoreKey(types.StoreKey)
	return struct {
		depinject.Out
		Cdc          codec.Codec
		StoreService store.KVStoreService
	}{
		Cdc:          encCfg.Codec,
		StoreService: runtime.NewKVStoreService(key),
	}
}

func TestDepinject(t *testing.T) {
	/// we just need any module's proto to register the provider here, no specific reason to use bank
	appconfig.RegisterModule(&bankmodulev1.Module{}, appconfig.Provide(DummyProvider))
	var appModules map[string]appmodule.AppModule
	keeper := new(keeper.Keeper)
	require.NoError(t,
		depinject.Inject(
			depinject.Configs(
				appconfig.Compose(
					&appv1alpha1.Config{
						Modules: []*appv1alpha1.ModuleConfig{
							{
								Name:   banktypes.ModuleName,
								Config: appconfig.WrapAny(&bankmodulev1.Module{}),
							},
							{
								Name:   types.ModuleName,
								Config: appconfig.WrapAny(&modulev1.Module{}),
							},
						},
					},
				),
				depinject.Provide(ProvideDeps),
			),
			&appModules,
			&keeper,
		),
	)

	require.NotNil(t, keeper, "expected keeper to not be nil after depinject")
	multihooks, ok := keeper.Hooks().(types.MultiEpochHooks)
	require.True(t, ok, "expected keeper to have MultiEpochHooks after depinject")
	require.Len(t, multihooks, 1, "expected MultiEpochHooks to have 1 element after depinject")
	require.Equal(
		t,
		types.EpochHooksWrapper{EpochHooks: testEpochHooks{}},
		multihooks[0],
		"expected the only hook in MultiEpochHooks to be the test hook",
	)
	module, ok := appModules[types.ModuleName].(epochs.AppModule)
	require.True(t, ok, "expected depinject to fill map with the epochs AppModule")
	require.Equal(
		t,
		types.MultiEpochHooks{types.EpochHooksWrapper{EpochHooks: testEpochHooks{}}},
		module.Keeper().Hooks(),
	)
	require.Same(t, keeper, module.Keeper()) // pointers pointing to the same instance
}
