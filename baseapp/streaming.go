package baseapp

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"cosmossdk.io/store/streaming"
	storetypes "cosmossdk.io/store/types"
	"github.com/spf13/cast"

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

// RegisterStreamingServices registers streaming services with the BaseApp.
func (app *BaseApp) RegisterStreamingServices(appOpts servertypes.AppOptions, keys map[string]*storetypes.KVStoreKey) {
	// register streaming services
	streamingCfg := cast.ToStringMap(appOpts.Get(StreamingTomlKey))
	for service := range streamingCfg {
		pluginKey := fmt.Sprintf("%s.%s.%s", StreamingTomlKey, service, StreamingABCIPluginTomlKey)
		pluginName := strings.TrimSpace(cast.ToString(appOpts.Get(pluginKey)))
		if len(pluginName) > 0 {
			logLevel := cast.ToString(appOpts.Get(flags.FlagLogLevel))
			plugin, err := streaming.NewStreamingPlugin(pluginName, logLevel)
			if err != nil {
				fmt.Printf("failed to load streaming plugin: %s", err)
				os.Exit(1)
			}
			if err := app.registerStreamingPlugin(appOpts, keys, plugin); err != nil {
				fmt.Printf("failed to register streaming plugin: %s", err)
				os.Exit(1)
			}
		}
	}
}

// registerStreamingPlugin registers streaming plugins with the BaseApp.
func (app *BaseApp) registerStreamingPlugin(
	appOpts servertypes.AppOptions,
	keys map[string]*storetypes.KVStoreKey,
	streamingPlugin interface{},
) error {
	switch t := streamingPlugin.(type) {
	case storetypes.ABCIListener:
		app.registerABCIListenerPlugin(appOpts, keys, t)
	default:
		return fmt.Errorf("unexpected plugin type %T", t)
	}
	return nil
}

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
			AbciListeners: []storetypes.ABCIListener{abciListener},
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
