package keeper_test

import (
	"testing"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	"gotest.tools/v3/assert"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type fixture struct {
	ctx sdk.Context

	queryClient       v1.QueryClient
	legacyQueryClient v1beta1.QueryClient

	bankKeeper    bankkeeper.Keeper
	stakingKeeper *stakingkeeper.Keeper
	govKeeper     *keeper.Keeper
}

func initFixture(tb testing.TB) *fixture {
	tb.Helper()
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, distrtypes.StoreKey, stakingtypes.StoreKey, types.StoreKey,
	)
	cdc := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{}, gov.AppModuleBasic{}).Codec

	logger := log.NewTestLogger(tb)
	cms := integration.CreateMultiStore(keys, logger)

	newCtx := sdk.NewContext(cms, cmtproto.Header{}, true, logger)

	authority := authtypes.NewModuleAddress(types.ModuleName)

	maccPerms := map[string][]string{
		distrtypes.ModuleName:          nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		types.ModuleName:               {authtypes.Burner},
	}

	accountKeeper := authkeeper.NewAccountKeeper(
		cdc,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		addresscodec.NewBech32Codec(sdk.Bech32MainPrefix),
		sdk.Bech32MainPrefix,
		authority.String(),
	)

	blockedAddresses := map[string]bool{
		accountKeeper.GetAuthority(): false,
	}
	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		accountKeeper,
		blockedAddresses,
		authority.String(),
		log.NewNopLogger(),
	)

	stakingKeeper := stakingkeeper.NewKeeper(cdc, runtime.NewKVStoreService(keys[stakingtypes.StoreKey]), accountKeeper, bankKeeper, authority.String(), addresscodec.NewBech32Codec(sdk.Bech32PrefixValAddr), addresscodec.NewBech32Codec(sdk.Bech32PrefixConsAddr))

	// set default staking params
	assert.NilError(tb, stakingKeeper.SetParams(newCtx, stakingtypes.DefaultParams()))

	distrKeeper := distrkeeper.NewKeeper(
		cdc, runtime.NewKVStoreService(keys[distrtypes.StoreKey]), accountKeeper, bankKeeper, stakingKeeper, distrtypes.ModuleName, authority.String(),
	)

	// Create MsgServiceRouter, but don't populate it before creating the gov
	// keeper.
	router := baseapp.NewMsgServiceRouter()
	router.SetInterfaceRegistry(cdc.InterfaceRegistry())

	govKeeper := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[types.StoreKey]),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		distrKeeper,
		router,
		types.DefaultConfig(),
		authority.String(),
	)
	err := govKeeper.ProposalID.Set(newCtx, 1)
	assert.NilError(tb, err)
	govRouter := v1beta1.NewRouter()
	govRouter.AddRoute(types.RouterKey, v1beta1.ProposalHandler)
	govKeeper.SetLegacyRouter(govRouter)
	err = govKeeper.Params.Set(newCtx, v1.DefaultParams())
	assert.NilError(tb, err)

	authModule := auth.NewAppModule(cdc, accountKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper, nil)
	stakingModule := staking.NewAppModule(cdc, stakingKeeper, accountKeeper, bankKeeper, nil)
	distrModule := distribution.NewAppModule(cdc, distrKeeper, accountKeeper, bankKeeper, stakingKeeper, nil)
	govModule := gov.NewAppModule(cdc, govKeeper, accountKeeper, bankKeeper, nil)

	integrationApp := integration.NewIntegrationApp(newCtx, logger, keys, cdc, map[string]appmodule.AppModule{
		authtypes.ModuleName:    authModule,
		banktypes.ModuleName:    bankModule,
		distrtypes.ModuleName:   distrModule,
		stakingtypes.ModuleName: stakingModule,
		types.ModuleName:        govModule,
	})

	sdkCtx := sdk.UnwrapSDKContext(integrationApp.Context())

	msgSrvr := keeper.NewMsgServerImpl(govKeeper)
	legacyMsgSrvr := keeper.NewLegacyMsgServerImpl(authority.String(), msgSrvr)

	// Register MsgServer and QueryServer
	v1.RegisterMsgServer(router, msgSrvr)
	v1beta1.RegisterMsgServer(router, legacyMsgSrvr)

	v1.RegisterQueryServer(integrationApp.QueryHelper(), keeper.NewQueryServer(govKeeper))
	v1beta1.RegisterQueryServer(integrationApp.QueryHelper(), keeper.NewLegacyQueryServer(govKeeper))

	queryClient := v1.NewQueryClient(integrationApp.QueryHelper())
	legacyQueryClient := v1beta1.NewQueryClient(integrationApp.QueryHelper())

	return &fixture{
		ctx:               sdkCtx,
		queryClient:       queryClient,
		legacyQueryClient: legacyQueryClient,
		bankKeeper:        bankKeeper,
		stakingKeeper:     stakingKeeper,
		govKeeper:         govKeeper,
	}
}
