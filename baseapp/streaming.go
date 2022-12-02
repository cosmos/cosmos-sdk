package baseapp

import (
	"context"
	"fmt"
	"github.com/spf13/cast"
	"sort"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	store "github.com/cosmos/cosmos-sdk/store/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// ABCIListener is the interface that we're exposing as a streaming service.
type ABCIListener interface {
	// ListenBeginBlock updates the streaming service with the latest BeginBlock messages
	ListenBeginBlock(ctx context.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) error
	// ListenEndBlock updates the steaming service with the latest EndBlock messages
	ListenEndBlock(ctx context.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) error
	// ListenDeliverTx updates the steaming service with the latest DeliverTx messages
	ListenDeliverTx(ctx context.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) error
	// ListenCommit updates the steaming service with the latest Commit messages and state changes
	ListenCommit(ctx context.Context, res abci.ResponseCommit, changeSet []*store.StoreKVPair) error
}

const (
	StreamingTomlKey                  = "streaming"
	StreamingABCITomlKey              = "abci"
	StreamingABCIPluginTomlKey        = "plugin"
	StreamingABCIKeysTomlKey          = "keys"
	StreamingABCIStopNodeOnErrTomlKey = "stop-node-on-err"
	StreamingABCIAsync                = "async"
)

// RegisterStreamingPlugin registers streaming plugins with the App.
func RegisterStreamingPlugin(
	bApp *BaseApp,
	appOpts servertypes.AppOptions,
	keys map[string]*store.KVStoreKey,
	streamingPlugin interface{},
) error {
	switch t := streamingPlugin.(type) {
	case ABCIListener:
		registerABCIListenerPlugin(bApp, appOpts, keys, t)
	default:
		return fmt.Errorf("unexpected plugin type %T", t)
	}
	return nil
}

func registerABCIListenerPlugin(
	bApp *BaseApp,
	appOpts servertypes.AppOptions,
	keys map[string]*store.KVStoreKey,
	abciListener ABCIListener,
) {
	asyncKey := fmt.Sprintf("%s.%s.%s", StreamingTomlKey, StreamingABCITomlKey, StreamingABCIAsync)
	async := cast.ToBool(appOpts.Get(asyncKey))
	stopNodeOnErrKey := fmt.Sprintf("%s.%s.%s", StreamingTomlKey, StreamingABCITomlKey, StreamingABCIStopNodeOnErrTomlKey)
	stopNodeOnErr := cast.ToBool(appOpts.Get(stopNodeOnErrKey))
	keysKey := fmt.Sprintf("%s.%s.%s", StreamingTomlKey, StreamingABCITomlKey, StreamingABCIKeysTomlKey)
	exposeKeysStr := cast.ToStringSlice(appOpts.Get(keysKey))
	exposedKeys := exposeStoreKeysSorted(exposeKeysStr, keys)
	bApp.cms.AddListeners(exposedKeys)
	bApp.SetStreamingService(abciListener)
	bApp.stopNodeOnABCIListenerErr = stopNodeOnErr
	bApp.abciListenersAsync = async
}

func exposeAll(list []string) bool {
	for _, ele := range list {
		if ele == "*" {
			return true
		}
	}
	return false
}

func exposeStoreKeysSorted(keysStr []string, keys map[string]*store.KVStoreKey) []store.StoreKey {
	var exposeStoreKeys []store.StoreKey
	if exposeAll(keysStr) {
		exposeStoreKeys = make([]store.StoreKey, 0, len(keys))
		for _, storeKey := range keys {
			exposeStoreKeys = append(exposeStoreKeys, storeKey)
		}
	} else {
		exposeStoreKeys = make([]store.StoreKey, 0, len(keysStr))
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
