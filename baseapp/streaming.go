package baseapp

import (
	"context"
	"fmt"
	"sort"
	"strings"

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	"github.com/spf13/cast"

	"cosmossdk.io/log"
	"cosmossdk.io/schema"
	"cosmossdk.io/schema/appdata"
	"cosmossdk.io/schema/decoding"
	"cosmossdk.io/schema/indexer"
	"cosmossdk.io/store/streaming"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/client/flags"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

const (
	StreamingTomlKey                  = "streaming"
	StreamingABCITomlKey              = "abci"
	StreamingABCIPluginTomlKey        = "plugin"
	StreamingABCIKeysTomlKey          = "keys"
	StreamingABCIStopNodeOnErrTomlKey = "stop-node-on-err"
)

// EnableIndexer enables the built-in indexer with the provided options (usually from the app.toml indexer key),
// kv-store keys, and app modules. Using the built-in indexer framework is mutually exclusive from using other
// types of streaming listeners.
func (app *BaseApp) EnableIndexer(indexerOpts interface{}, keys map[string]*storetypes.KVStoreKey, appModules map[string]any) error {
	listener, err := indexer.StartManager(indexer.ManagerOptions{
		Config:     indexerOpts,
		Resolver:   decoding.ModuleSetDecoderResolver(appModules),
		SyncSource: nil,
		Logger:     app.logger.With(log.ModuleKey, "indexer"),
	})
	if err != nil {
		return err
	}

	exposedKeys := exposeStoreKeysSorted([]string{"*"}, keys)
	app.cms.AddListeners(exposedKeys)

	app.streamingManager = storetypes.StreamingManager{
		ABCIListeners: []storetypes.ABCIListener{listenerWrapper{listener}},
		StopNodeOnErr: true,
	}

	return nil
}

// RegisterStreamingServices registers streaming services with the BaseApp.
func (app *BaseApp) RegisterStreamingServices(appOpts servertypes.AppOptions, keys map[string]*storetypes.KVStoreKey) error {
	// register streaming services
	streamingCfg := cast.ToStringMap(appOpts.Get(StreamingTomlKey))
	for service := range streamingCfg {
		pluginKey := fmt.Sprintf("%s.%s.%s", StreamingTomlKey, service, StreamingABCIPluginTomlKey)
		pluginName := strings.TrimSpace(cast.ToString(appOpts.Get(pluginKey)))
		if len(pluginName) > 0 {
			logLevel := cast.ToString(appOpts.Get(flags.FlagLogLevel))
			plugin, err := streaming.NewStreamingPlugin(pluginName, logLevel)
			if err != nil {
				return fmt.Errorf("failed to load streaming plugin: %w", err)
			}
			if err := app.registerStreamingPlugin(appOpts, keys, plugin); err != nil {
				return fmt.Errorf("failed to register streaming plugin %w", err)
			}
		}
	}

	return nil
}

// registerStreamingPlugin registers streaming plugins with the BaseApp.
func (app *BaseApp) registerStreamingPlugin(
	appOpts servertypes.AppOptions,
	keys map[string]*storetypes.KVStoreKey,
	streamingPlugin interface{},
) error {
	v, ok := streamingPlugin.(storetypes.ABCIListener)
	if !ok {
		return fmt.Errorf("unexpected plugin type %T", v)
	}

	app.registerABCIListenerPlugin(appOpts, keys, v)
	return nil
}

// registerABCIListenerPlugin registers plugins that implement the ABCIListener interface.
func (app *BaseApp) registerABCIListenerPlugin(
	appOpts servertypes.AppOptions,
	keys map[string]*storetypes.KVStoreKey,
	abciListener storetypes.ABCIListener,
) {
	stopNodeOnErrKey := fmt.Sprintf("%s.%s.%s", StreamingTomlKey, StreamingABCITomlKey, StreamingABCIStopNodeOnErrTomlKey)
	stopNodeOnErr := cast.ToBool(appOpts.Get(stopNodeOnErrKey))
	keysKey := fmt.Sprintf("%s.%s.%s", StreamingTomlKey, StreamingABCITomlKey, StreamingABCIKeysTomlKey)
	exposeKeysStr := cast.ToStringSlice(appOpts.Get(keysKey))
	exposedKeys := exposeStoreKeysSorted(exposeKeysStr, keys)
	app.cms.AddListeners(exposedKeys)
	app.SetStreamingManager(
		storetypes.StreamingManager{
			ABCIListeners: []storetypes.ABCIListener{abciListener},
			StopNodeOnErr: stopNodeOnErr,
		},
	)
}

func exposeAll(list []string) bool {
	for _, ele := range list {
		if ele == "*" {
			return true
		}
	}
	return false
}

func exposeStoreKeysSorted(keysStr []string, keys map[string]*storetypes.KVStoreKey) []storetypes.StoreKey {
	var exposeStoreKeys []storetypes.StoreKey
	if exposeAll(keysStr) {
		exposeStoreKeys = make([]storetypes.StoreKey, 0, len(keys))
		for key := range keys {
			exposeStoreKeys = append(exposeStoreKeys, keys[key])
		}
	} else {
		exposeStoreKeys = make([]storetypes.StoreKey, 0, len(keysStr))
		for _, keyStr := range keysStr {
			if storeKey, ok := keys[keyStr]; ok {
				exposeStoreKeys = append(exposeStoreKeys, storeKey)
			}
		}
	}
	// sort storeKeys for deterministic output
	sort.SliceStable(exposeStoreKeys, func(i, j int) bool {
		return exposeStoreKeys[i].Name() < exposeStoreKeys[j].Name()
	})

	return exposeStoreKeys
}

type listenerWrapper struct {
	listener appdata.Listener
}

func (p listenerWrapper) ListenFinalizeBlock(_ context.Context, req abci.FinalizeBlockRequest, res abci.FinalizeBlockResponse) error {
	if p.listener.StartBlock != nil {
		err := p.listener.StartBlock(appdata.StartBlockData{
			Height: uint64(req.Height),
		})
		if err != nil {
			return err
		}
	}

	//// TODO txs, events

	return nil
}

func (p listenerWrapper) ListenCommit(ctx context.Context, res abci.CommitResponse, changeSet []*storetypes.StoreKVPair) error {
	if cb := p.listener.OnKVPair; cb != nil {
		updates := make([]appdata.ActorKVPairUpdate, len(changeSet))
		for i, pair := range changeSet {
			updates[i] = appdata.ActorKVPairUpdate{
				Actor: []byte(pair.StoreKey),
				StateChanges: []schema.KVPairUpdate{
					{
						Key:    pair.Key,
						Value:  pair.Value,
						Remove: pair.Delete,
					},
				},
			}
		}
		err := cb(appdata.KVPairData{Updates: updates})
		if err != nil {
			return err
		}
	}

	if p.listener.Commit != nil {
		commitCb, err := p.listener.Commit(appdata.CommitData{})
		if err != nil {
			return err
		}
		if commitCb != nil {
			err := commitCb()
			if err != nil {
				return err
			}
		}
	}

	return nil
}
