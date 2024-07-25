package indexer

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/schema/appdata"
	"cosmossdk.io/schema/decoding"
	"cosmossdk.io/schema/logutil"
)

// ManagerOptions are the options for starting the indexer manager.
type ManagerOptions struct {
	// Config is the user configuration for all indexing. It should generally be an instance map[string]interface{}
	// or json.RawMessage and match the json structure of ManagerConfig. The manager will attempt to convert it to ManagerConfig.
	Config interface{}

	// Resolver is the decoder resolver that will be used to decode the data. It is required.
	Resolver decoding.DecoderResolver

	// SyncSource is a representation of the current state of key-value data to be used in a catch-up sync.
	// Catch-up syncs will be performed at initialization when necessary. SyncSource is optional but if
	// it is omitted, indexers will only be able to start indexing state from genesis.
	SyncSource decoding.SyncSource

	// Logger is the logger that indexers can use to write logs. It is optional.
	Logger logutil.Logger

	// Context is the context that indexers should use for shutdown signals via Context.Done(). It can also
	// be used to pass down other parameters to indexers if necessary. If it is omitted, context.Background
	// will be used.
	Context context.Context
}

// ManagerConfig is the configuration of the indexer manager and contains the configuration for each indexer target.
type ManagerConfig struct {
	// Target is a map of named indexer targets to their configuration.
	Target map[string]Config
}

// TODO add global include & exclude module filters
type ManagerResult struct{}

// StartManager starts the indexer manager with the given options. The state machine should write all relevant app data to
// the returned listener.
func StartManager(opts ManagerOptions) (appdata.Listener, error) {
	logger := opts.Logger
	if logger == nil {
		logger = logutil.NoopLogger{}
	}

	logger.Info("Starting indexer manager")

	scopeableLogger, canScopeLogger := logger.(logutil.ScopeableLogger)

	cfg, err := unmarshalConfig(opts.Config)
	if err != nil {
		return appdata.Listener{}, err
	}

	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	listeners := make([]appdata.Listener, 0, len(cfg.Target))

	includeModuleFilter := map[string]bool{}
	var excludeModuleFilter map[string]bool

	for targetName, targetCfg := range cfg.Target {
		init, ok := indexerRegistry[targetCfg.Type]
		if !ok {
			return appdata.Listener{}, fmt.Errorf("indexer type %q not found", targetCfg.Type)
		}

		logger.Info("Starting indexer", "target", targetName, "type", targetCfg.Type)

		if len(targetCfg.ExcludeModules) != 0 && len(targetCfg.IncludeModules) != 0 {
			return appdata.Listener{}, fmt.Errorf("only one of exclude_modules or include_modules can be set for indexer %s", targetName)
		}

		childLogger := logger
		if canScopeLogger {
			childLogger = scopeableLogger.WithAsAny("indexer", targetName).(logutil.Logger)
		}

		initRes, err := init(InitParams{
			Config:  targetCfg,
			Context: ctx,
			Logger:  childLogger,
		})
		if err != nil {
			return appdata.Listener{}, err
		}

		listener := initRes.Listener

		var excludedModuleFilter map[string]bool
		allIncludedModules := map[string]bool{}
		if len(targetCfg.ExcludeModules) != 0 {
			excluded := map[string]bool{}

			for _, moduleName := range targetCfg.ExcludeModules {
				excluded[moduleName] = true
			}

			// for excluded modules we must do an intersection
			excludedModuleFilter = filterIntersection(excludedModuleFilter, excluded)

			listener = appdata.ModuleFilter(listener, func(moduleName string) bool {
				return !excluded[moduleName]
			})

		} else if len(targetCfg.IncludeModules) != 0 {
			included := map[string]bool{}
			for _, moduleName := range targetCfg.IncludeModules {
				included[moduleName] = true
				// for included modules we do a union
				allIncludedModules[moduleName] = true
			}
			listener = appdata.ModuleFilter(listener, func(moduleName string) bool {
				return included[moduleName]
			})
		}

		if targetCfg.ExcludeBlockHeaders && listener.StartBlock != nil {
			cb := listener.StartBlock
			listener.StartBlock = func(data appdata.StartBlockData) error {
				data.HeaderBytes = nil
				data.HeaderJSON = nil
				return cb(data)
			}
		}

		if targetCfg.ExcludeTxs {
			listener.OnTx = nil
		}

		if targetCfg.ExcludeEvents {
			listener.OnEvent = nil
		}

		listeners = append(listeners, listener)

		// TODO check last block persisted
	}

	rootListener := appdata.AsyncListenerMux(
		appdata.AsyncListenerOptions{Context: ctx},
		listeners...,
	)

	rootModuleFilter := combineIncludeExcludeFilters(includeModuleFilter, excludeModuleFilter)
	if rootModuleFilter != nil {
		rootListener = appdata.ModuleFilter(rootListener, rootModuleFilter)
	}

	return rootListener, nil
}

func unmarshalConfig(cfg interface{}) (*ManagerConfig, error) {
	var jsonBz []byte
	var err error

	switch cfg := cfg.(type) {
	case map[string]interface{}:
		jsonBz, err = json.Marshal(cfg)
		if err != nil {
			return nil, err
		}
	case json.RawMessage:
		jsonBz = cfg
	default:
		return nil, fmt.Errorf("can't convert %T to %T", cfg, ManagerConfig{})
	}

	var res ManagerConfig
	err = json.Unmarshal(jsonBz, &res)
	return &res, err
}
