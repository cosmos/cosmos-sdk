package baseapp

import (
	"fmt"
	"sort"
	"strings"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	store "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cast"
	abci "github.com/tendermint/tendermint/abci/types"
)

// ABCIListener is the interface that we're exposing as a streaming service.
type ABCIListener interface {
	// ListenBeginBlock updates the streaming service with the latest BeginBlock messages
	ListenBeginBlock(ctx types.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) error
	// ListenEndBlock updates the steaming service with the latest EndBlock messages
	ListenEndBlock(ctx types.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) error
	// ListenDeliverTx updates the steaming service with the latest DeliverTx messages
	ListenDeliverTx(ctx types.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) error
	// ListenCommit updates the steaming service with the latest Commit messages and state changes
	ListenCommit(ctx types.Context, res abci.ResponseCommit, changeSet []store.StoreKVPair) error
}

const (
	StreamingTomlKey              = "streaming"
	StreamingEnableTomlKey        = "enable"
	StreamingPluginTomlKey        = "plugin"
	StreamingKeysTomlKey          = "keys"
	StreamingStopNodeOnErrTomlKey = "stop-node-on-err"
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
	stopNodeOnErrKey := fmt.Sprintf("%s.%s", StreamingTomlKey, StreamingStopNodeOnErrTomlKey)
	stopNodeOnErr := cast.ToBool(appOpts.Get(stopNodeOnErrKey))
	keysKey := fmt.Sprintf("%s.%s", StreamingTomlKey, StreamingKeysTomlKey)
	exposeKeysStr := cast.ToStringSlice(appOpts.Get(keysKey))
	bApp.cms.AddListeners(exposeStoreKeysSorted(exposeKeysStr, keys))
	bApp.abciListener = abciListener
	bApp.stopNodeOnStreamingErr = stopNodeOnErr
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
		return strings.Compare(exposeStoreKeys[i].Name(), exposeStoreKeys[j].Name()) < 0
	})

	return exposeStoreKeys
}
