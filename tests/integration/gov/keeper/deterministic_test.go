package keeper_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"gotest.tools/v3/assert"
	"pgregory.net/rapid"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/auth"
	authkeeper "cosmossdk.io/x/auth/keeper"
	authsims "cosmossdk.io/x/auth/simulation"
	authtestutil "cosmossdk.io/x/auth/testutil"
	_ "cosmossdk.io/x/auth/tx/config"
	authtypes "cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/bank"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktestutil "cosmossdk.io/x/bank/testutil"
	banktypes "cosmossdk.io/x/bank/types"
	"cosmossdk.io/x/gov"
	"cosmossdk.io/x/gov/keeper"
	"cosmossdk.io/x/gov/types"
	govtypes "cosmossdk.io/x/gov/types/v1"
	v1 "cosmossdk.io/x/gov/types/v1"
	"cosmossdk.io/x/gov/types/v1beta1"
	minttypes "cosmossdk.io/x/mint/types"
	poolkeeper "cosmossdk.io/x/protocolpool/keeper"
	pooltypes "cosmossdk.io/x/protocolpool/types"
	"cosmossdk.io/x/staking"
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	stakingtypes "cosmossdk.io/x/staking/types"

	_ "cosmossdk.io/x/staking"

	"github.com/cosmos/cosmos-sdk/baseapp"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	_ "github.com/cosmos/cosmos-sdk/x/consensus"
)

var (
	denomRegex   = `[a-zA-Z][a-zA-Z0-9/:._-]{2,127}`
	addr1        = sdk.MustAccAddressFromBech32("cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5")
	coin1        = sdk.NewCoin("denom", math.NewInt(10))
	metadataAtom = banktypes.Metadata{
		Description: "The native staking token of the Cosmos Hub.",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    "uatom",
				Exponent: 0,
				Aliases:  []string{"microatom"},
			},
			{
				Denom:    "atom",
				Exponent: 6,
				Aliases:  []string{"ATOM"},
			},
		},
		Base:    "uatom",
		Display: "atom",
	}
)

type deterministicFixture struct {
	ctx               sdk.Context
	queryClient       v1.QueryClient
	legacyQueryClient v1beta1.QueryClient

	accountKeeper authkeeper.AccountKeeper
	bankKeeper    bankkeeper.Keeper
	stakingKeeper *stakingkeeper.Keeper
	govKeeper     *keeper.Keeper
}

func initDeterministicFixture(t *testing.T) *deterministicFixture {
	t.Helper()
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey, pooltypes.StoreKey, types.StoreKey,
	)
	encodingCfg := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, auth.AppModule{}, bank.AppModule{}, gov.AppModule{})
	cdc := encodingCfg.Codec

	logger := log.NewTestLogger(t)
	cms := integration.CreateMultiStore(keys, logger)

	newCtx := sdk.NewContext(cms, true, logger)

	authority := authtypes.NewModuleAddress("gov")

	maccPerms := map[string][]string{
		pooltypes.ModuleName:           {},
		pooltypes.StreamAccount:        {},
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		types.ModuleName:               {authtypes.Burner},
	}

	// gomock initializations
	ctrl := gomock.NewController(t)
	acctsModKeeper := authtestutil.NewMockAccountsModKeeper(ctrl)

	accountKeeper := authkeeper.NewAccountKeeper(
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[authtypes.StoreKey]), log.NewNopLogger()),
		cdc,
		authtypes.ProtoBaseAccount,
		acctsModKeeper,
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
	assert.NilError(t, err)

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

	assert.NilError(t, govKeeper.ProposalID.Set(newCtx, 1))
	govRouter := v1beta1.NewRouter()
	govRouter.AddRoute(types.RouterKey, v1beta1.ProposalHandler)
	govKeeper.SetLegacyRouter(govRouter)
	err = govKeeper.Params.Set(newCtx, v1.DefaultParams())
	assert.NilError(t, err)

	authModule := auth.NewAppModule(cdc, accountKeeper, acctsModKeeper, authsims.RandomGenesisAccounts)
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
		},
		baseapp.NewMsgServiceRouter(),
		baseapp.NewGRPCQueryRouter(),
	)

	sdkCtx := sdk.UnwrapSDKContext(integrationApp.Context())

	// Register MsgServer and QueryServer
	msgSrvr := keeper.NewMsgServerImpl(govKeeper)
	legacyMsgSrvr := keeper.NewLegacyMsgServerImpl(authority.String(), msgSrvr)

	// Register MsgServer and QueryServer
	v1.RegisterMsgServer(router, msgSrvr)
	v1beta1.RegisterMsgServer(router, legacyMsgSrvr)

	v1.RegisterQueryServer(integrationApp.QueryHelper(), keeper.NewQueryServer(govKeeper))
	v1beta1.RegisterQueryServer(integrationApp.QueryHelper(), keeper.NewLegacyQueryServer(govKeeper))

	queryClient := v1.NewQueryClient(integrationApp.QueryHelper())
	legacyQueryClient := v1beta1.NewQueryClient(integrationApp.QueryHelper())

	f := deterministicFixture{
		ctx:               sdkCtx,
		queryClient:       queryClient,
		legacyQueryClient: legacyQueryClient,
		accountKeeper:     accountKeeper,
		bankKeeper:        bankKeeper,
		stakingKeeper:     stakingKeeper,
		govKeeper:         govKeeper,
	}

	return &f
}

func fundAccount(f *deterministicFixture, addr sdk.AccAddress, coin ...sdk.Coin) {
	err := banktestutil.FundAccount(f.ctx, f.bankKeeper, addr, sdk.NewCoins(coin...))
	assert.NilError(&testing.T{}, err)
}

func getCoin(rt *rapid.T) sdk.Coin {
	return sdk.NewCoin(
		rapid.StringMatching(denomRegex).Draw(rt, "denom"),
		math.NewInt(rapid.Int64Min(1).Draw(rt, "amount")),
	)
}

func TestGRPCQueryConstitution(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)

	rapid.Check(t, func(rt *rapid.T) {
		addr := testdata.AddressGenerator(rt).Draw(rt, "address")
		coin := getCoin(rt)
		fundAccount(f, addr, coin)

		// addrStr, err := codectestutil.CodecOptions{}.GetAddressCodec().BytesToString(addr)
		// assert.NilError(t, err)

		req := govtypes.QueryConstitutionRequest{}

		testdata.DeterministicIterations(t, f.ctx, &req, f.queryClient.Constitution, 0, true)
	})

	// addr1Str, err := codectestutil.CodecOptions{}.GetAddressCodec().BytesToString(addr1)
	// assert.NilError(t, err)

	// fundAccount(f, addr1, coin1)
	// req := banktypes.NewQueryBalanceRequest(addr1Str, coin1.GetDenom())
	// testdata.DeterministicIterations(t, f.ctx, req, f.queryClient.Balance, 1087, false)
}
