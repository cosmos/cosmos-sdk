package runtime

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/core/store"
	"cosmossdk.io/server/v2/stf"
	"cosmossdk.io/server/v2/stf/branch"

	sdkmodule "github.com/cosmos/cosmos-sdk/types/module"
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

// DefaultGenesis returns a default genesis from the registered AppModule's.
func (a *AppBuilder) DefaultGenesis() map[string]json.RawMessage {
	return a.app.moduleManager.DefaultGenesis(a.app.cdc)
}

// RegisterModules registers the provided modules with the module manager.
// This is the primary hook for integrating with modules which are not registered using the app config.
func (a *AppBuilder) RegisterModules(modules ...appmodule.AppModule) error {
	for _, appModule := range modules {
		if mod, ok := appModule.(sdkmodule.HasName); ok {
			name := mod.Name()
			if _, ok := a.app.moduleManager.modules[name]; ok {
				return fmt.Errorf("module named %q already exists", name)
			}
			a.app.moduleManager.modules[name] = appModule

			if mod, ok := appModule.(sdkmodule.HasRegisterInterfaces); ok {
				mod.RegisterInterfaces(a.app.interfaceRegistry)
			}

			if mod, ok := appModule.(sdkmodule.HasAminoCodec); ok {
				mod.RegisterLegacyAminoCodec(a.app.amino)
			}
		}
	}

	return nil
}

// Build builds an *App instance.
func (a *AppBuilder) Build(db Store, opts ...AppBuilderOption) (*App, error) {
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
