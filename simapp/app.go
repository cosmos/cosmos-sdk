package simapp

import (
	"fmt"
	"io"

	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/testutil/x/counter"

	clienthelpers "cosmossdk.io/client/v2/helpers"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/runtime"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
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
	*SDKApp
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
	sdkAppConfig := DefaultSDKAppConfig(appName, appOpts, baseAppOptions...)
	sdkAppConfig.WithEpochs = true

	sdkApp := NewSDKApp(logger, db, traceStore, sdkAppConfig)

	app := &SimApp{
		SDKApp: sdkApp,
	}

	// set up keeper ...
	// app.AddModule()
	// add keeper
	// add module to module manager
	// update keys
	//

	counterKeeper := counterkeeper.NewKeeper(runtime.NewKVStoreService(storetypes.NewKVStoreKey(countertypes.ModuleName)))
	_ = counter.NewAppModule(counterKeeper)

	app.LoadModules()

	// RegisterUpgradeHandlers is used for registering any on-chain upgrades.
	app.RegisterUpgradeHandlers()

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			panic(fmt.Errorf("error loading last version: %w", err))
		}
	}

	return app
}
