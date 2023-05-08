package keeper_test

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/simapp"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
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

// fixture only tests gov's keeper logic around tallying, since it
// relies on complex interactions with x/staking.
//
// It also uses simapp (and not a depinjected app) because we manually set a
// new app.StakingKeeper in `createValidators`.
type fixture struct {
	app               *simapp.SimApp
	ctx               sdk.Context
	queryClient       v1.QueryClient
	legacyQueryClient v1beta1.QueryClient
	addrs             []sdk.AccAddress
	msgSrvr           v1.MsgServer
	legacyMsgSrvr     v1beta1.MsgServer
}

type newFixture struct {
	app *integration.App

	ctx  sdk.Context
	cdc  codec.Codec
	keys map[string]*storetypes.KVStoreKey

	queryClient       v1.QueryClient
	legacyQueryClient v1beta1.QueryClient
	addrs             []sdk.AccAddress

	msgSrvr       v1.MsgServer
	legacyMsgSrvr v1beta1.MsgServer

	bankKeeper    bankkeeper.Keeper
	stakingKeeper *stakingkeeper.Keeper
	govKeeper     *keeper.Keeper
}

func initNewFixture(t testing.TB) *newFixture {
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, distrtypes.StoreKey, stakingtypes.StoreKey, types.StoreKey,
	)
	cdc := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{}, gov.AppModuleBasic{}).Codec

	logger := log.NewTestLogger(t)
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

	// Populate the gov account with some coins, as the TestProposal we have
	// is a MsgSend from the gov account.
	coins := sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100000)))
	err := bankKeeper.MintCoins(newCtx, minttypes.ModuleName, coins)
	assert.NilError(t, err)
	err = bankKeeper.SendCoinsFromModuleToModule(newCtx, minttypes.ModuleName, types.ModuleName, coins)
	assert.NilError(t, err)

	stakingKeeper := stakingkeeper.NewKeeper(cdc, keys[stakingtypes.StoreKey], accountKeeper, bankKeeper, authority.String())

	// set default staking params
	stakingKeeper.SetParams(newCtx, stakingtypes.DefaultParams())

	distrKeeper := distrkeeper.NewKeeper(
		cdc, runtime.NewKVStoreService(keys[distrtypes.StoreKey]), accountKeeper, bankKeeper, stakingKeeper, distrtypes.ModuleName, authority.String(),
	)

	govKeeper := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[types.StoreKey]),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		distrKeeper,
		&baseapp.MsgServiceRouter{},
		types.DefaultConfig(),
		authority.String(),
	)

	// // Create MsgServiceRouter, but don't populate it before creating the gov
	// // keeper.
	// msr := baseapp.NewMsgServiceRouter()
	// msr.SetInterfaceRegistry(cdc.InterfaceRegistry())

	authModule := auth.NewAppModule(cdc, accountKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper, nil)
	stakingModule := staking.NewAppModule(cdc, stakingKeeper, accountKeeper, bankKeeper, nil)
	distrModule := distribution.NewAppModule(cdc, distrKeeper, accountKeeper, bankKeeper, stakingKeeper, nil)
	govModule := gov.NewAppModule(cdc, govKeeper, accountKeeper, bankKeeper, nil)

	integrationApp := integration.NewIntegrationApp(newCtx, logger, keys, cdc, authModule, bankModule, stakingModule, distrModule, govModule)

	sdkCtx := sdk.UnwrapSDKContext(integrationApp.Context())

	msgSrvr := keeper.NewMsgServerImpl(govKeeper)
	legacyMsgSrvr := keeper.NewLegacyMsgServerImpl(authority.String(), msgSrvr)

	msr := integrationApp.MsgServiceRouter()

	// Register MsgServer and QueryServer
	v1.RegisterMsgServer(msr, msgSrvr)
	v1beta1.RegisterMsgServer(msr, legacyMsgSrvr)

	v1.RegisterQueryServer(integrationApp.QueryHelper(), keeper.NewQuerier(govKeeper))
	v1beta1.RegisterQueryServer(integrationApp.QueryHelper(), keeper.NewLegacyQueryServer(govKeeper))

	queryClient := v1.NewQueryClient(integrationApp.QueryHelper())
	legacyQueryClient := v1beta1.NewQueryClient(integrationApp.QueryHelper())

	addrs := simtestutil.AddTestAddrsIncremental(bankKeeper, stakingKeeper, sdkCtx, 2, sdk.NewInt(30000000))

	return &newFixture{
		app:               integrationApp,
		ctx:               sdkCtx,
		cdc:               cdc,
		keys:              keys,
		queryClient:       queryClient,
		legacyQueryClient: legacyQueryClient,
		addrs:             addrs,
		msgSrvr:           msgSrvr,
		legacyMsgSrvr:     legacyMsgSrvr,
		bankKeeper:        bankKeeper,
		stakingKeeper:     stakingKeeper,
		govKeeper:         govKeeper,
	}
}

// initFixture uses simapp (and not a depinjected app) because we manually set a
// new app.StakingKeeper in `createValidators` which is used in most of the
// gov keeper tests.
func initFixture(t *testing.T) *fixture {
	f := &fixture{}

	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, cmtproto.Header{})

	// Populate the gov account with some coins, as the TestProposal we have
	// is a MsgSend from the gov account.
	coins := sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100000)))
	err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins)
	assert.NilError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, coins)
	assert.NilError(t, err)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	v1.RegisterQueryServer(queryHelper, app.GovKeeper)
	legacyQueryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	v1beta1.RegisterQueryServer(legacyQueryHelper, keeper.NewLegacyQueryServer(app.GovKeeper))
	queryClient := v1.NewQueryClient(queryHelper)
	legacyQueryClient := v1beta1.NewQueryClient(legacyQueryHelper)

	f.app = app
	f.ctx = ctx
	f.queryClient = queryClient
	f.legacyQueryClient = legacyQueryClient
	f.msgSrvr = keeper.NewMsgServerImpl(f.app.GovKeeper)

	govAcct := f.app.GovKeeper.GetGovernanceAccount(f.ctx).GetAddress()
	f.legacyMsgSrvr = keeper.NewLegacyMsgServerImpl(govAcct.String(), f.msgSrvr)
	f.addrs = simtestutil.AddTestAddrsIncremental(app.BankKeeper, app.StakingKeeper, ctx, 2, sdk.NewInt(30000000))

	return f
}
