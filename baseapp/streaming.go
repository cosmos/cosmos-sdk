package baseapp

import (
	"fmt"
	"sort"

	"github.com/spf13/cast"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

const (
	StreamingTomlKey                  = "streaming"
	StreamingABCITomlKey              = "abci"
	StreamingABCIPluginTomlKey        = "plugin"
	StreamingABCIKeysTomlKey          = "keys"
	StreamingABCIStopNodeOnErrTomlKey = "stop-node-on-err"
)

// RegisterStreamingPlugin registers streaming plugins with the App.
func RegisterStreamingPlugin(
	bApp *BaseApp,
	appOpts servertypes.AppOptions,
	keys map[string]*storetypes.KVStoreKey,
	streamingPlugin interface{},
) error {
	switch t := streamingPlugin.(type) {
	case storetypes.ABCIListener:
		registerABCIListenerPlugin(bApp, appOpts, keys, t)
	default:
		return fmt.Errorf("unexpected plugin type %T", t)
	}
	return nil
}

func registerABCIListenerPlugin(
	bApp *BaseApp,
	appOpts servertypes.AppOptions,
	keys map[string]*storetypes.KVStoreKey,
	abciListener storetypes.ABCIListener,
) {
	stopNodeOnErrKey := fmt.Sprintf("%s.%s.%s", StreamingTomlKey, StreamingABCITomlKey, StreamingABCIStopNodeOnErrTomlKey)
	stopNodeOnErr := cast.ToBool(appOpts.Get(stopNodeOnErrKey))
	keysKey := fmt.Sprintf("%s.%s.%s", StreamingTomlKey, StreamingABCITomlKey, StreamingABCIKeysTomlKey)
	exposeKeysStr := cast.ToStringSlice(appOpts.Get(keysKey))
	exposedKeys := exposeStoreKeysSorted(exposeKeysStr, keys)
	bApp.cms.AddListeners(exposedKeys)
	bApp.streamingManager = storetypes.StreamingManager{
		AbciListeners: []storetypes.ABCIListener{abciListener},
		StopNodeOnErr: stopNodeOnErr,
	}
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
