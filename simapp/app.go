package simapp

import (
	"fmt"
	"io"

	dbm "github.com/cosmos/cosmos-db"

	clienthelpers "cosmossdk.io/client/v2/helpers"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/runtime"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/x/counter"
	counterkeeper "github.com/cosmos/cosmos-sdk/testutil/x/counter/keeper"
	countertypes "github.com/cosmos/cosmos-sdk/testutil/x/counter/types"
)

const appName = "SimApp"

// DefaultNodeHome default home directories for the application daemon
var DefaultNodeHome string

var (
	_ runtime.AppI            = (*SimApp)(nil)
	_ servertypes.Application = (*SimApp)(nil)
)

// SimApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type SimApp struct {
	*app.SDKApp
}

func init() {
	var err error
	DefaultNodeHome, err = clienthelpers.GetNodeHomeDirectory(".simapp")
	if err != nil {
		panic(err)
	}
}

// NewSimApp returns a reference to an initialized SimApp.
func NewSimApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *SimApp {
	sdkAppConfig := app.DefaultSDKAppConfig(appName, appOpts, baseAppOptions...)
	sdkAppConfig.WithEpochs = true

	sdkApp := app.NewSDKApp(logger, db, traceStore, sdkAppConfig)

	simApp := &SimApp{
		SDKApp: sdkApp,
	}

	// set up keeper ...
	// app.AddModule()
	// add keeper
	// add module to module manager
	// update keys
	//

	key := storetypes.NewKVStoreKey(countertypes.ModuleName)
	counterKeeper := counterkeeper.NewKeeper(runtime.NewKVStoreService(key))
	counterModule := counter.NewAppModule(counterKeeper)
	wrappedModule := app.Module{
		AppModule: counterModule,
		StoreKeys: map[string]*storetypes.KVStoreKey{
			countertypes.ModuleName: key,
		},
		Name:      countertypes.ModuleName,
		MaccPerms: nil,
	}

	err := simApp.AddModule(wrappedModule)
	if err != nil {
		panic(err)
	}

	simApp.LoadModules()

	// RegisterUpgradeHandlers is used for registering any on-chain upgrades.
	simApp.RegisterUpgradeHandlers()

	if loadLatest {
		if err := simApp.LoadLatestVersion(); err != nil {
			panic(fmt.Errorf("error loading last version: %w", err))
		}
	}

	return simApp
}
