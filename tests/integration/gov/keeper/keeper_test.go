package keeper_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/auth"
	authkeeper "cosmossdk.io/x/auth/keeper"
	authsims "cosmossdk.io/x/auth/simulation"
	authtypes "cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/bank"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	"cosmossdk.io/x/gov"
	"cosmossdk.io/x/gov/keeper"
	"cosmossdk.io/x/gov/types"
	v1 "cosmossdk.io/x/gov/types/v1"
	"cosmossdk.io/x/gov/types/v1beta1"
	minttypes "cosmossdk.io/x/mint/types"
	poolkeeper "cosmossdk.io/x/protocolpool/keeper"
	pooltypes "cosmossdk.io/x/protocolpool/types"
	"cosmossdk.io/x/staking"
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	stakingtypes "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

type fixture struct {
	ctx sdk.Context

	queryClient       v1.QueryClient
	legacyQueryClient v1beta1.QueryClient

	accountKeeper authkeeper.AccountKeeper
	bankKeeper    bankkeeper.Keeper
	stakingKeeper *stakingkeeper.Keeper
	govKeeper     *keeper.Keeper
}

func initFixture(tb testing.TB) *fixture {
	tb.Helper()
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey, pooltypes.StoreKey, types.StoreKey,
	)
	encodingCfg := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, auth.AppModule{}, bank.AppModule{}, gov.AppModule{})
	cdc := encodingCfg.Codec

	logger := log.NewTestLogger(tb)
	cms := integration.CreateMultiStore(keys, logger)

	newCtx := sdk.NewContext(cms, true, logger)

	authority := authtypes.NewModuleAddress(types.ModuleName)

	maccPerms := map[string][]string{
		pooltypes.ModuleName:           {},
		pooltypes.StreamAccount:        {},
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		types.ModuleName:               {authtypes.Burner},
	}

	accountKeeper := authkeeper.NewAccountKeeper(
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[authtypes.StoreKey]), log.NewNopLogger()),
		cdc,
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
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[banktypes.StoreKey]), log.NewNopLogger()),
		cdc,
		accountKeeper,
		blockedAddresses,
		authority.String(),
	)

	stakingKeeper := stakingkeeper.NewKeeper(cdc, runtime.NewEnvironment(runtime.NewKVStoreService(keys[stakingtypes.StoreKey]), log.NewNopLogger()), accountKeeper, bankKeeper, authority.String(), addresscodec.NewBech32Codec(sdk.Bech32PrefixValAddr), addresscodec.NewBech32Codec(sdk.Bech32PrefixConsAddr))

	poolKeeper := poolkeeper.NewKeeper(cdc, runtime.NewEnvironment(runtime.NewKVStoreService(keys[pooltypes.StoreKey]), log.NewNopLogger()), accountKeeper, bankKeeper, stakingKeeper, authority.String())

	// set default staking params
	err := stakingKeeper.Params.Set(newCtx, stakingtypes.DefaultParams())
	assert.NilError(tb, err)

	// Create MsgServiceRouter, but don't populate it before creating the gov
	// keeper.
	router := baseapp.NewMsgServiceRouter()
	router.SetInterfaceRegistry(cdc.InterfaceRegistry())
	queryRouter := baseapp.NewGRPCQueryRouter()
	queryRouter.SetInterfaceRegistry(cdc.InterfaceRegistry())

	govKeeper := keeper.NewKeeper(
		cdc,
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[types.StoreKey]), log.NewNopLogger(), runtime.EnvWithRouterService(queryRouter, router)),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		poolKeeper,
		keeper.DefaultConfig(),
		authority.String(),
	)
	assert.NilError(tb, govKeeper.ProposalID.Set(newCtx, 1))
	govRouter := v1beta1.NewRouter()
	govRouter.AddRoute(types.RouterKey, v1beta1.ProposalHandler)
	govKeeper.SetLegacyRouter(govRouter)
	err = govKeeper.Params.Set(newCtx, v1.DefaultParams())
	assert.NilError(tb, err)

	authModule := auth.NewAppModule(cdc, accountKeeper, authsims.RandomGenesisAccounts)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper)
	stakingModule := staking.NewAppModule(cdc, stakingKeeper, accountKeeper, bankKeeper)
	govModule := gov.NewAppModule(cdc, govKeeper, accountKeeper, bankKeeper, poolKeeper)

	integrationApp := integration.NewIntegrationApp(newCtx, logger, keys, cdc,
		encodingCfg.InterfaceRegistry.SigningContext().AddressCodec(),
		encodingCfg.InterfaceRegistry.SigningContext().ValidatorAddressCodec(),
		map[string]appmodule.AppModule{
			authtypes.ModuleName:    authModule,
			banktypes.ModuleName:    bankModule,
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
		accountKeeper:     accountKeeper,
		bankKeeper:        bankKeeper,
		stakingKeeper:     stakingKeeper,
		govKeeper:         govKeeper,
	}
}
