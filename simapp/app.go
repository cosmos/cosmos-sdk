package simapp

import (
	"fmt"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
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
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *SimApp {
	sdkAppConfig := app.DefaultSDKAppConfig(appName, appOpts, baseAppOptions...)
	sdkAppConfig.WithEpochs = true
	sdkAppConfig.ExtendVoteHandler = NewVoteExtensionHandler().ExtendVote()
	sdkAppConfig.VerifyVoteExtensionHandler = NewVoteExtensionHandler().VerifyVoteExtension()

	sdkApp := app.NewSDKApp(logger, db, nil, sdkAppConfig)

	simApp := &SimApp{
		SDKApp: sdkApp,
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
