package simapp

import (
	"fmt"
	"io"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/x/counter"
)

const appName = "SimApp"

// DefaultNodeHome default home directories for the application daemon
var DefaultNodeHome string

var _ app.AppI = (*SimApp)(nil)

// SimApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type SimApp struct {
	*app.SDKApp

	// TODO add other structures or configurations if you want
}

func init() {
	var err error
	DefaultNodeHome, err = app.GetNodeHomeDirectory(".simapp")
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
	sdkAppConfig.ExtendVoteHandler = NewVoteExtensionHandler().ExtendVote()
	sdkAppConfig.VerifyVoteExtensionHandler = NewVoteExtensionHandler().VerifyVoteExtension()

	sdkApp := app.NewSDKApp(logger, db, traceStore, sdkAppConfig)

	simApp := &SimApp{
		SDKApp: sdkApp,
	}

	err := simApp.AddModules(counter.NewExtendedAppModule())
	if err != nil {
		panic(err)
	}

	simApp.LoadModules()

	// RegisterUpgradeHandlers is used for registering any on-chain upgrades.
	app.RegisterUpgradeHandlers(simApp, MyUpgrade)

	if loadLatest {
		if err := simApp.LoadLatestVersion(); err != nil {
			panic(fmt.Errorf("error loading last version: %w", err))
		}
	}

	return simApp
}
