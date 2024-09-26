package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"cosmossdk.io/core/appmodule"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/server"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/runtime/v2/services"
	"cosmossdk.io/server/v2/appmanager"
	"cosmossdk.io/server/v2/stf"
	"cosmossdk.io/server/v2/stf/branch"
	"cosmossdk.io/store/v2/db"
	rootstore "cosmossdk.io/store/v2/root"
)

// AppBuilder is a type that is injected into a container by the runtime/v2 module
// (as *AppBuilder) which can be used to create an app which is compatible with
// the existing app.go initialization conventions.
type AppBuilder[T transaction.Tx] struct {
	app          *App[T]
	config       server.DynamicConfig
	storeOptions *rootstore.Options

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
func (a *AppBuilder[T]) RegisterModules(modules map[string]appmodulev2.AppModule) error {
	for name, appModule := range modules {
		// if a (legacy) module implements the HasName interface, check that the name matches
		if mod, ok := appModule.(interface{ Name() string }); ok {
			if name != mod.Name() {
				a.app.logger.Warn(fmt.Sprintf("module name %q does not match name returned by HasName: %q", name, mod.Name()))
			}
		}

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

	return nil
}

// RegisterStores registers the provided store keys.
// This method should only be used for registering extra stores
// which is necessary for modules that not registered using the app config.
// To be used in combination of RegisterModules.
func (a *AppBuilder[T]) RegisterStores(keys ...string) {
	a.app.storeKeys = append(a.app.storeKeys, keys...)
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

	home := a.config.GetString(FlagHome)
	scRawDb, err := db.NewDB(
		db.DBType(a.config.GetString("store.app-db-backend")),
		"application",
		filepath.Join(home, "data"),
		nil,
	)
	if err != nil {
		panic(err)
	}

	var storeOptions rootstore.Options
	if a.storeOptions != nil {
		storeOptions = *a.storeOptions
	} else {
		storeOptions = rootstore.DefaultStoreOptions()
	}
	factoryOptions := &rootstore.FactoryOptions{
		Logger:    a.app.logger,
		RootDir:   home,
		Options:   storeOptions,
		StoreKeys: append(a.app.storeKeys, "stf"),
		SCRawDB:   scRawDb,
	}

	rs, err := rootstore.CreateRootStore(factoryOptions)
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
		InitGenesis: func(
			ctx context.Context,
			src io.Reader,
			txHandler func(json.RawMessage) error,
		) (store.WriterMap, error) {
			// this implementation assumes that the state is a JSON object
			bz, err := io.ReadAll(src)
			if err != nil {
				return nil, fmt.Errorf("failed to read import state: %w", err)
			}
			var genesisJSON map[string]json.RawMessage
			if err = json.Unmarshal(bz, &genesisJSON); err != nil {
				return nil, err
			}

			v, zeroState, err := a.app.db.StateLatest()
			if err != nil {
				return nil, fmt.Errorf("unable to get latest state: %w", err)
			}
			if v != 0 { // TODO: genesis state may be > 0, we need to set version on store
				return nil, errors.New("cannot init genesis on non-zero state")
			}
			genesisCtx := services.NewGenesisContext(a.branch(zeroState))
			genesisState, err := genesisCtx.Run(ctx, func(ctx context.Context) error {
				err = a.app.moduleManager.InitGenesisJSON(ctx, genesisJSON, txHandler)
				if err != nil {
					return fmt.Errorf("failed to init genesis: %w", err)
				}
				return nil
			})

			return genesisState, err
		},
		ExportGenesis: func(ctx context.Context, version uint64) ([]byte, error) {
			state, err := a.app.db.StateAt(version)
			if err != nil {
				return nil, fmt.Errorf("unable to get state at given version: %w", err)
			}

			genesisJson, err := a.app.moduleManager.ExportGenesisForModules(
				ctx,
				func() store.WriterMap {
					return a.branch(state)
				},
			)
			if err != nil {
				return nil, fmt.Errorf("failed to export genesis: %w", err)
			}

			bz, err := json.Marshal(genesisJson)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal genesis: %w", err)
			}

			return bz, nil
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
func AppBuilderWithPostTxExec[T transaction.Tx](postTxExec func(ctx context.Context, tx T, success bool) error) AppBuilderOption[T] {
	return func(a *AppBuilder[T]) {
		a.postTxExec = postTxExec
	}
}

func AppBuilderWithStoreOptions[T transaction.Tx](opts *rootstore.Options) AppBuilderOption[T] {
	return func(a *AppBuilder[T]) {
		a.storeOptions = opts
	}
}
