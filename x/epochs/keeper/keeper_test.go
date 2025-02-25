package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	epochskeeper "github.com/cosmos/cosmos-sdk/x/epochs/keeper"
	"github.com/cosmos/cosmos-sdk/x/epochs/types"
)

type KeeperTestSuite struct {
	suite.Suite
	Ctx          sdk.Context
	EpochsKeeper epochskeeper.Keeper
	queryClient  types.QueryClient
}

func (s *KeeperTestSuite) SetupTest() {
	ctx, epochsKeeper := Setup(s.T())

	s.Ctx = ctx
	s.EpochsKeeper = epochsKeeper
	queryRouter := baseapp.NewGRPCQueryRouter()
	cfg := module.NewConfigurator(nil, nil, queryRouter)
	types.RegisterQueryServer(cfg.QueryServer(), epochskeeper.NewQuerier(s.EpochsKeeper))
	grpcQueryService := &baseapp.QueryServiceTestHelper{
		GRPCQueryRouter: queryRouter,
		Ctx:             s.Ctx,
	}
	encCfg := moduletestutil.MakeTestEncodingConfig()
	grpcQueryService.SetInterfaceRegistry(encCfg.InterfaceRegistry)
	s.queryClient = types.NewQueryClient(grpcQueryService)
}

func Setup(t *testing.T) (sdk.Context, epochskeeper.Keeper) {
	t.Helper()

	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockTime(time.Now().UTC())
	encCfg := moduletestutil.MakeTestEncodingConfig()

	epochsKeeper := epochskeeper.NewKeeper(
		storeService,
		encCfg.Codec,
	)
	epochsKeeper.SetHooks(types.NewMultiEpochHooks())
	ctx = ctx.WithBlockTime(time.Now().UTC()).WithBlockHeight(1).WithChainID("epochs")

	err := epochsKeeper.InitGenesis(ctx, *types.DefaultGenesis())
	require.NoError(t, err)
	SetEpochStartTime(ctx, epochsKeeper)

	return ctx, epochsKeeper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func SetEpochStartTime(ctx sdk.Context, epochsKeeper epochskeeper.Keeper) {
	epochs, err := epochsKeeper.AllEpochInfos(ctx)
	if err != nil {
		panic(err)
	}
	for _, epoch := range epochs {
		epoch.StartTime = ctx.BlockTime()
		err := epochsKeeper.EpochInfo.Remove(ctx, epoch.Identifier)
		if err != nil {
			panic(err)
		}
		err = epochsKeeper.AddEpochInfo(ctx, epoch)
		if err != nil {
			panic(err)
		}
	}
}
