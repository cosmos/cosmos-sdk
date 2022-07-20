package runtime

import (
	"encoding/json"
	"io"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/version"
)

// AppBuilder is a type that is injected into a container by the runtime module
// (as *AppBuilder) which can be used to create an app which is compatible with
// the existing app.go initialization conventions.
type AppBuilder struct {
	app *App
}

// DefaultGenesis returns a default genesis from the registered
// AppModuleBasic's.
func (a *AppBuilder) DefaultGenesis() map[string]json.RawMessage {
	return a.app.basicManager.DefaultGenesis(a.app.cdc)
}

// Build builds an *App instance.
func (a *AppBuilder) Build(
	logger log.Logger,
	db dbm.Connection,
	traceStore io.Writer,
	baseAppOptions ...baseapp.AppOption,
) *App {
	for _, option := range a.app.baseAppOptions {
		baseAppOptions = append(baseAppOptions, baseapp.AppOptionFunc(option))
	}
	baseAppOptions = append(baseAppOptions, baseapp.SetSubstores(a.app.storeKeys...))

	bApp := baseapp.NewBaseApp(a.app.config.AppName, logger, db, nil, baseAppOptions...)
	bApp.SetMsgServiceRouter(a.app.msgServiceRouter)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(a.app.interfaceRegistry)

	a.app.BaseApp = bApp
	return a.app
}
