package indexer

import (
	"context"
	"fmt"
	"sync"

	"cosmossdk.io/schema/addressutil"
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

	// AddressCodec is the address codec that indexers can use to encode and decode addresses. It should always be
	// provided, but if it is omitted, the indexer manager will use a default codec which encodes and decodes addresses
	// as hex strings.
	AddressCodec addressutil.AddressCodec

	// DoneWaitGroup is a wait group that all indexer manager go routines will wait on before returning when the context
	// is done.
	// It is optional.
	DoneWaitGroup *sync.WaitGroup
}

// ManagerConfig is the configuration of the indexer manager and contains the configuration for each indexer target.
type ManagerConfig struct {
	// Target is a map of named indexer targets to their configuration.
	Target map[string]Config

	// ChannelBufferSize is the buffer size of the channels used for buffering data sent to indexer go routines.
	// It defaults to 1024.
	ChannelBufferSize *int `json:"channel_buffer_size"`
}

type ManagerResult struct {
	Listener     appdata.Listener
	ModuleFilter ModuleFilterConfig
}

// StartManager starts the indexer manager with the given options. The state machine should write all relevant app data to
// the returned listener.
func StartManager(opts ManagerOptions) (ManagerResult, error) {
	logger := opts.Logger
	if logger == nil {
		logger = logutil.NoopLogger{}
	}

	logger.Info("Starting indexer manager")

	scopeableLogger, canScopeLogger := logger.(logutil.ScopeableLogger)

	cfg, err := unmarshalConfig(opts.Config)
	if err != nil {
		return ManagerResult{}, err
	}

	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	listeners := make([]appdata.Listener, 0, len(cfg.Target))

	allModuleFilters := make([]ModuleFilterConfig, 0, len(cfg.Target))
	for targetName, targetCfg := range cfg.Target {
		init, ok := indexerRegistry[targetCfg.Type]
		if !ok {
			return ManagerResult{}, fmt.Errorf("indexer type %q not found", targetCfg.Type)
		}

		logger.Info("Starting indexer", "target", targetName, "type", targetCfg.Type)

		if err := targetCfg.Filter.Validate(); err != nil {
			return ManagerResult{}, fmt.Errorf("invalid filter for target %q: %w", targetName, err)
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
			return ManagerResult{}, err
		}

		listener := targetCfg.Filter.Apply(initRes.Listener)
		listener = addSyncAndSanityCheck(initRes.LastBlockPersisted, listener, opts, targetCfg.Filter.Modules)
		listeners = append(listeners, listener)

		allModuleFilters = append(allModuleFilters, targetCfg.Filter.Modules)
	}

	bufSize := 1024
	if cfg.ChannelBufferSize != nil {
		bufSize = *cfg.ChannelBufferSize
	}
	asyncOpts := appdata.AsyncListenerOptions{
		Context:       ctx,
		DoneWaitGroup: opts.DoneWaitGroup,
		BufferSize:    bufSize,
	}

	rootListener := appdata.AsyncListenerMux(
		asyncOpts,
		listeners...,
	)

	rootModuleFilter := combineModuleFilters(allModuleFilters)
	rootListener, err = decoding.Middleware(rootListener, opts.Resolver, decoding.MiddlewareOptions{
		ModuleFilter: rootModuleFilter.ToFunction(),
	})
	if err != nil {
		return ManagerResult{}, err
	}
	rootListener = rootModuleFilter.Apply(rootListener)
	rootListener = appdata.AsyncListener(asyncOpts, rootListener)

	return ManagerResult{
		Listener:     rootListener,
		ModuleFilter: rootModuleFilter,
	}, nil
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
