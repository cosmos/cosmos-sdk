package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/appmanager"
	"cosmossdk.io/server/v2/stf"
	"cosmossdk.io/server/v2/stf/branch"
	rootstore "cosmossdk.io/store/v2/root"

	sdkmodule "github.com/cosmos/cosmos-sdk/types/module"
)

type branchFunc func(state store.ReaderMap) store.WriterMap

// AppBuilder is a type that is injected into a container by the runtime module
// (as *AppBuilder) which can be used to create an app which is compatible with
// the existing app.go initialization conventions.
type AppBuilder struct {
	app          *App
	storeOptions *rootstore.FactoryOptions

	// the following fields are used to overwrite the default
	branch      branchFunc
	txValidator func(ctx context.Context, tx transaction.Tx) error
}

// DefaultGenesis returns a default genesis from the registered AppModule's.
func (a *AppBuilder) DefaultGenesis() map[string]json.RawMessage {
	return a.app.moduleManager.DefaultGenesis()
}

// RegisterModules registers the provided modules with the module manager.
// This is the primary hook for integrating with modules which are not registered using the app config.
func (a *AppBuilder) RegisterModules(modules ...appmodulev2.AppModule) error {
	for _, appModule := range modules {
		if mod, ok := appModule.(sdkmodule.HasName); ok {
			name := mod.Name()
			if _, ok := a.app.moduleManager.modules[name]; ok {
				return fmt.Errorf("module named %q already exists", name)
			}
			a.app.moduleManager.modules[name] = appModule

			if mod, ok := appModule.(appmodulev2.HasRegisterInterfaces); ok {
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
func (a *AppBuilder) Build(opts ...AppBuilderOption) (*App, error) {
	for _, opt := range opts {
		opt(a)
	}

	// default branch
	if a.branch == nil {
		a.branch = branch.DefaultNewWriterMap
	}

	// default tx validator
	if a.txValidator == nil {
		a.txValidator = a.app.moduleManager.TxValidators()
	}

	if err := a.app.moduleManager.RegisterServices(a.app); err != nil {
		return nil, err
	}

	stfMsgHandler, err := a.app.msgRouterBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build STF message handler: %w", err)
	}
	stfQueryHandler, err := a.app.queryRouterBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build query handler: %w", err)
	}

	endBlocker, valUpdate := a.app.moduleManager.EndBlock()

	// TODO how to set?
	// no-op postTxExec
	postTxExec := func(ctx context.Context, tx transaction.Tx, success bool) error {
		return nil
	}

	a.app.stf = stf.NewSTF[transaction.Tx](
		stfMsgHandler,
		stfQueryHandler,
		a.app.moduleManager.PreBlocker(),
		a.app.moduleManager.BeginBlock(),
		endBlocker,
		a.txValidator,
		valUpdate,
		postTxExec,
		a.branch,
	)

	rs, err := rootstore.CreateRootStore(a.storeOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create root store: %w", err)
	}
	a.app.db = rs

	appManagerBuilder := appmanager.Builder[transaction.Tx]{
		STF:                a.app.stf,
		DB:                 a.app.db,
		ValidateTxGasLimit: a.app.config.GasConfig.ValidateTxGasLimit,
		QueryGasLimit:      a.app.config.GasConfig.QueryGasLimit,
		SimulationGasLimit: a.app.config.GasConfig.SimulationGasLimit,
		InitGenesis: func(ctx context.Context, src io.Reader, txHandler func(json.RawMessage) error) error {
			// this implementation assumes that the state is a JSON object
			bz, err := io.ReadAll(src)
			if err != nil {
				return fmt.Errorf("failed to read import state: %w", err)
			}
			var genesisState map[string]json.RawMessage
			if err = json.Unmarshal(bz, &genesisState); err != nil {
				return err
			}
			if err = a.app.moduleManager.InitGenesisJSON(ctx, genesisState, txHandler); err != nil {
				return fmt.Errorf("failed to init genesis: %w", err)
			}
			return nil
		},
	}

	appManager, err := appManagerBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build app manager: %w", err)
	}
	a.app.AppManager = appManager

	return a.app, nil
}

// AppBuilderOption is a function that can be passed to AppBuilder.Build to
// customize the resulting app.
type AppBuilderOption func(*AppBuilder)

// AppBuilderWithBranch sets a custom branch implementation for the app.
func AppBuilderWithBranch(branch branchFunc) AppBuilderOption {
	return func(a *AppBuilder) {
		a.branch = branch
	}
}

// AppBuilderWithTxValidator sets the tx validator for the app.
// It overrides all default tx validators defined by modules.
func AppBuilderWithTxValidator(validator func(ctx context.Context, tx transaction.Tx) error) AppBuilderOption {
	return func(a *AppBuilder) {
		a.txValidator = validator
	}
}
