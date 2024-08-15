package runtime

import (
	"encoding/json"
	"io"
	"path/filepath"

	"cosmossdk.io/x/auth/ante/unorderedtx"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/spf13/cast"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
)

// AppBuilder is a type that is injected into a container by the runtime module
// (as *AppBuilder) which can be used to create an app which is compatible with
// the existing app.go initialization conventions.
type AppBuilder struct {
	app *App

	viper *viper.Viper
}

// DefaultGenesis returns a default genesis from the registered modules.
func (a *AppBuilder) DefaultGenesis() map[string]json.RawMessage {
	return a.app.DefaultGenesis()
}

// Build builds an *App instance.
func (a *AppBuilder) Build(db dbm.DB, traceStore io.Writer, baseAppOptions ...func(*baseapp.BaseApp)) *App {
	for _, option := range a.app.baseAppOptions {
		baseAppOptions = append(baseAppOptions, option)
	}

	bApp := baseapp.NewBaseApp(a.app.config.AppName, a.app.logger, db, nil, baseAppOptions...)
	bApp.SetMsgServiceRouter(a.app.msgServiceRouter)
	bApp.SetGRPCQueryRouter(a.app.grpcQueryRouter)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(a.app.interfaceRegistry)
	bApp.MountStores(a.app.storeKeys...)

	a.app.BaseApp = bApp
	a.app.configurator = module.NewConfigurator(a.app.cdc, a.app.MsgServiceRouter(), a.app.GRPCQueryRouter())

	// register unordered tx manager
	// create, start, and load the unordered tx manager
	utxDataDir := filepath.Join(cast.ToString(a.viper.Get(flags.FlagHome)), "data")
	a.app.UnorderedTxManager = unorderedtx.NewManager(utxDataDir)
	a.app.UnorderedTxManager.Start()

	if err := a.app.ModuleManager.RegisterServices(a.app.configurator); err != nil {
		panic(err)
	}

	return a.app
}
