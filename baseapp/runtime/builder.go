package runtime

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
)

type AppBuilder struct {
	app *App
}

func (a *AppBuilder) RegisterModules(modules ...module.AppModule) error {
	for _, appModule := range modules {
		if _, ok := a.app.mm.Modules[appModule.Name()]; ok {
			return fmt.Errorf("module named %q already exists", appModule.Name())
		}
		a.app.mm.Modules[appModule.Name()] = appModule
	}
	return nil
}

func (a *AppBuilder) DefaultGenesis() map[string]json.RawMessage {
	genesis := make(map[string]json.RawMessage)
	for name, wrapper := range a.app.privateState.moduleBasics {
		genesis[name] = wrapper.DefaultGenesis(a.app.privateState.cdc)
	}
	return genesis
}

func (a *AppBuilder) Create(logger log.Logger, db dbm.DB, traceStore io.Writer, baseAppOptions ...func(*baseapp.BaseApp)) *App {
	for _, option := range a.app.baseAppOptions {
		baseAppOptions = append(baseAppOptions, option)
	}
	bApp := baseapp.NewBaseApp(a.app.config.AppName, logger, db, baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(a.app.privateState.interfaceRegistry)
	bApp.MountStores(a.app.privateState.storeKeys...)
	bApp.SetTxHandler(a.app.txHandler)

	a.app.BaseApp = bApp
	return a.app
}

func (a *AppBuilder) Finish(loadLatest bool) error {
	if a.app == nil {
		return fmt.Errorf("app not created yet, can't finish")
	}

	configurator := module.NewConfigurator(a.app.privateState.cdc, a.app.msgServiceRegistrar, a.app.GRPCQueryRouter())
	a.app.mm.RegisterServices(configurator)
	a.app.mm.SetOrderInitGenesis(a.app.config.InitGenesis...)
	a.app.mm.SetOrderBeginBlockers(a.app.config.BeginBlockers...)
	a.app.mm.SetOrderEndBlockers(a.app.config.EndBlockers...)
	a.app.SetBeginBlocker(a.app.mm.BeginBlock)
	a.app.SetEndBlocker(a.app.mm.EndBlock)
	a.app.SetInitChainer(a.app.InitChainer)

	if loadLatest {
		if err := a.app.LoadLatestVersion(); err != nil {
			return err
		}
	}

	return nil
}
