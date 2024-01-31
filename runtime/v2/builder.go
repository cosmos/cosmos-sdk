package runtime

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/transaction"
	coreappmanager "cosmossdk.io/server/v2/core/appmanager"
	"cosmossdk.io/server/v2/core/mempool"
	"cosmossdk.io/server/v2/core/store"
	"cosmossdk.io/server/v2/stf"
	"cosmossdk.io/server/v2/stf/branch"
	"github.com/cosmos/cosmos-sdk/types/module"
)

type branchFunc func(state store.ReaderMap) store.WriterMap

// AppBuilder is a type that is injected into a container by the runtime module
// (as *AppBuilder) which can be used to create an app which is compatible with
// the existing app.go initialization conventions.
type AppBuilder struct {
	app *App

	// options for building the app
	branch      branchFunc
	txValidator func(ctx context.Context, tx transaction.Tx) error
}

// DefaultGenesis returns a default genesis from the registered AppModuleBasic's.
func (a *AppBuilder) DefaultGenesis() map[string]json.RawMessage {
	panic("not implemented")
}

// RegisterModules registers the provided modules with the module manager and
// the basic module manager. This is the primary hook for integrating with
// modules which are not registered using the app config.
func (a *AppBuilder) RegisterModules(modules ...module.AppModule) error {
	for _, appModule := range modules {
		name := appModule.Name()
		if _, ok := a.app.moduleManager.modules[name]; ok {
			return fmt.Errorf("module named %q already exists", name)
		}

		a.app.moduleManager.modules[name] = appModule
		appModule.RegisterInterfaces(a.app.interfaceRegistry)
		appModule.RegisterLegacyAminoCodec(a.app.amino)
	}

	return nil
}

// Build builds an *App instance.
func (a *AppBuilder) Build(db store.Store, opts ...AppBuilderOption) (*App, error) {
	for _, opt := range opts {
		opt(a)
	}

	// default branch
	if a.branch == nil {
		a.branch = branch.DefaultNewWriterMap
	}

	// default tx validator
	if a.txValidator == nil {
		a.txValidator = a.app.moduleManager.TxValidation()
	}

	if err := a.app.moduleManager.RegisterServices(a.app); err != nil {
		return nil, err
	}

	stfMsgHandler, err := a.app.msgRouterBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build STF message handler: %w", err)
	}

	endBlocker, valUpdate := a.app.moduleManager.EndBlock()

	a.app.stf = stf.NewSTF[transaction.Tx](
		stfMsgHandler,
		stfMsgHandler,
		a.app.moduleManager.PreBlocker(),
		a.app.moduleManager.BeginBlock(),
		endBlocker,
		a.txValidator,
		valUpdate,
		a.branch,
	)
	a.app.db = db

	return a.app, nil
}

// AppBuilderOption is a function that can be passed to AppBuilder.Build to
// customize the resulting app.
type AppBuilderOption func(*AppBuilder)

func AppBuilderWithMempool(mempool mempool.Mempool[transaction.Tx]) AppBuilderOption {
	return func(a *AppBuilder) {
		a.app.mempool = mempool
	}
}

func AppBuilderWithPrepareBlockHandler(handler coreappmanager.PrepareHandler[transaction.Tx]) AppBuilderOption {
	return func(a *AppBuilder) {
		a.app.prepareBlockHandler = handler
	}
}

func AppBuilderWithVerifyBlockHandler(handler coreappmanager.ProcessHandler[transaction.Tx]) AppBuilderOption {
	return func(a *AppBuilder) {
		a.app.verifyBlockHandler = handler
	}
}

func AppBuilderWithBranch(branch branchFunc) AppBuilderOption {
	return func(a *AppBuilder) {
		a.branch = branch
	}
}

// AppBuilderWithTxValidator sets the tx validator for the app.
// It overrides the default tx validator from all modules.
func AppBuilderWithTxValidator(validator func(ctx context.Context, tx transaction.Tx) error) AppBuilderOption {
	return func(a *AppBuilder) {
		a.txValidator = validator
	}
}
