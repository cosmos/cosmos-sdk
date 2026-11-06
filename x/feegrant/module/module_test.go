package module_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/collections"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdkmodule "github.com/cosmos/cosmos-sdk/types/module"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	"github.com/cosmos/cosmos-sdk/x/feegrant/module"
	feegranttestutil "github.com/cosmos/cosmos-sdk/x/feegrant/testutil"
)

// TestRunMigrationsEmptyQueueValue runs the registered migrations from
// consensus version 2, as an upgrade handler does, and checks that they
// rewrite a fee allowance queue entry that has an empty value.
func TestRunMigrationsEmptyQueueValue(t *testing.T) {
	key := storetypes.NewKVStoreKey(feegrant.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModuleBasic{})
	ctx := testCtx.Ctx

	ctrl := gomock.NewController(t)
	accountKeeper := feegranttestutil.NewMockAccountKeeper(ctrl)
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()
	bankKeeper := feegranttestutil.NewMockBankKeeper(ctrl)

	k := keeper.NewKeeper(encCfg.Codec, runtime.NewKVStoreService(key), accountKeeper)
	appModule := module.NewAppModule(encCfg.Codec, accountKeeper, bankKeeper, k, encCfg.InterfaceRegistry)
	require.Equal(t, uint64(3), appModule.ConsensusVersion())

	msgRouter := baseapp.NewMsgServiceRouter()
	msgRouter.SetInterfaceRegistry(encCfg.InterfaceRegistry)
	queryRouter := baseapp.NewGRPCQueryRouter()
	queryRouter.SetInterfaceRegistry(encCfg.InterfaceRegistry)
	cfg := sdkmodule.NewConfigurator(encCfg.Codec, msgRouter, queryRouter)
	mm := sdkmodule.NewManager(appModule)
	require.NoError(t, mm.RegisterServices(cfg))

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

	vm, err := mm.RunMigrations(ctx, cfg, sdkmodule.VersionMap{feegrant.ModuleName: 2})
	require.NoError(t, err)
	require.Equal(t, uint64(3), vm[feegrant.ModuleName])
	require.Equal(t, []byte{0x01}, store.Get(queueKeyBz))

	ctx = ctx.WithBlockTime(exp.Add(time.Second))
	require.NoError(t, k.RemoveExpiredAllowances(ctx, 200))

	has, err := k.FeeAllowance.Has(ctx, collections.Join(grantee, granter))
	require.NoError(t, err)
	require.False(t, has)
	require.False(t, store.Has(queueKeyBz))
}
