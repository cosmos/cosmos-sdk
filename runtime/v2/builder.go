package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"cosmossdk.io/core/appmodule"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/appmanager"
	"cosmossdk.io/server/v2/stf"
	"cosmossdk.io/server/v2/stf/branch"
	rootstore "cosmossdk.io/store/v2/root"
)

// AppBuilder is a type that is injected into a container by the runtime/v2 module
// (as *AppBuilder) which can be used to create an app which is compatible with
// the existing app.go initialization conventions.
type AppBuilder[T transaction.Tx] struct {
	app          *App[T]
	storeOptions *rootstore.FactoryOptions

	// the following fields are used to overwrite the default
	branch      func(state store.ReaderMap) store.WriterMap
	txValidator func(ctx context.Context, tx T) error
	postTxExec  func(ctx context.Context, tx T, success bool) error
}

// DefaultGenesis returns a default genesis from the registered AppModule's.
func (a *AppBuilder[T]) DefaultGenesis() map[string]json.RawMessage {
	return a.app.moduleManager.DefaultGenesis()
}

// RegisterModules registers the provided modules with the module manager.
// This is the primary hook for integrating with modules which are not registered using the app config.
func (a *AppBuilder[T]) RegisterModules(modules ...appmodulev2.AppModule) error {
	for _, appModule := range modules {
		if mod, ok := appModule.(appmodule.HasName); ok {
			name := mod.Name()
			if _, ok := a.app.moduleManager.modules[name]; ok {
				return fmt.Errorf("module named %q already exists", name)
			}
			a.app.moduleManager.modules[name] = appModule

			if mod, ok := appModule.(appmodulev2.HasRegisterInterfaces); ok {
				mod.RegisterInterfaces(a.app.interfaceRegistrar)
			}

			if mod, ok := appModule.(appmodule.HasAminoCodec); ok {
				mod.RegisterLegacyAminoCodec(a.app.amino)
			}
		}
	}

	return nil
}

// RegisterStores registers the provided store keys.
// This method should only be used for registering extra stores
// wiich is necessary for modules that not registered using the app config.
// To be used in combination of RegisterModules.
func (a *AppBuilder[T]) RegisterStores(keys ...string) {
	a.app.storeKeys = append(a.app.storeKeys, keys...)
	if a.storeOptions != nil {
		a.storeOptions.StoreKeys = append(a.storeOptions.StoreKeys, keys...)
	}
}

// Build builds an *App instance.
func (a *AppBuilder[T]) Build(opts ...AppBuilderOption[T]) (*App[T], error) {
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

	// default post tx exec
	if a.postTxExec == nil {
		a.postTxExec = func(ctx context.Context, tx T, success bool) error {
			return nil
		}
	}

	if err := a.app.moduleManager.RegisterServices(a.app); err != nil {
		return nil, err
	}

	endBlocker, valUpdate := a.app.moduleManager.EndBlock()

	stf, err := stf.NewSTF[T](
		a.app.logger.With("module", "stf"),
		a.app.msgRouterBuilder,
		a.app.queryRouterBuilder,
		a.app.moduleManager.PreBlocker(),
		a.app.moduleManager.BeginBlock(),
		endBlocker,
		a.txValidator,
		valUpdate,
		a.postTxExec,
		a.branch,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create STF: %w", err)
	}

	a.app.stf = stf

	rs, err := rootstore.CreateRootStore(a.storeOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create root store: %w", err)
	}
	a.app.db = rs

	appManagerBuilder := appmanager.Builder[T]{
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

// AppBuilderOption is a function that can be passed to AppBuilder.Build to customize the resulting app.
type AppBuilderOption[T transaction.Tx] func(*AppBuilder[T])

// AppBuilderWithBranch sets a custom branch implementation for the app.
func AppBuilderWithBranch[T transaction.Tx](branch func(state store.ReaderMap) store.WriterMap) AppBuilderOption[T] {
	return func(a *AppBuilder[T]) {
		a.branch = branch
	}
}

// AppBuilderWithTxValidator sets the tx validator for the app.
// It overrides all default tx validators defined by modules.
func AppBuilderWithTxValidator[T transaction.Tx](txValidators func(ctx context.Context, tx T) error) AppBuilderOption[T] {
	return func(a *AppBuilder[T]) {
		a.txValidator = txValidators
	}
}

// AppBuilderWithPostTxExec sets logic that will be executed after each transaction.
// When not provided, a no-op function will be used.
func AppBuilderWithPostTxExec[T transaction.Tx](
	postTxExec func(
		ctx context.Context,
		tx T,
		success bool,
	) error,
) AppBuilderOption[T] {
	return func(a *AppBuilder[T]) {
		a.postTxExec = postTxExec
	}
}
