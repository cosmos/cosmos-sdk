package baseapp

import (
	"fmt"
	"os"
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
	// OnStoreCommit updates the steaming service with the latest state changes
	// It is called once per commit cycle
	OnStoreCommit(ctx types.Context, changeSet [][]byte) error
}

var (
	_ store.WriteListener = (*MemoryListener)(nil)
)

// MemoryListener listens to the state writes and accumulate the records in memory.
type MemoryListener struct {
	bApp          *BaseApp
	listener      ABCIListener
	stopNodeOnErr bool
	stateCache    [][]byte
}

// NewMemoryListener creates a listener that accumulate the state writes in memory.
func NewMemoryListener(
	bApp *BaseApp,
	listener ABCIListener,
	stopNodeOnErr bool,
) *MemoryListener {
	return &MemoryListener{
		bApp:          bApp,
		listener:      listener,
		stopNodeOnErr: stopNodeOnErr,
	}
}

// OnCommit receives the latest batch of state changes from the store upon Committer.Commit().
func (fl *MemoryListener) OnCommit() {
	ctx := fl.bApp.deliverState.ctx
	blockHeight := ctx.BlockHeight()
	logger := fl.bApp.logger
	cache := fl.PopStateCache()
	if fl.stopNodeOnErr {
		if err := fl.listener.OnStoreCommit(ctx, cache); err != nil {
			logger.Error("OnStoreCommit listening hook failed", "blockHeight", blockHeight, "err", err)
			os.Exit(1)
		}
	} else {
		go func() {
			if err := fl.listener.OnStoreCommit(ctx, cache); err != nil {
				logger.Error("OnStoreCommit listening hook failed", "blockHeight", blockHeight, "err", err)
			}
		}()
	}
}

// OnWrite implements WriteListener interface
func (fl *MemoryListener) OnWrite(storeKey store.StoreKey, key []byte, value []byte, delete bool) {
	logger := fl.bApp.logger
	kvPair := store.StoreKVPair{
		StoreKey: storeKey.Name(),
		Delete:   delete,
		Key:      key,
		Value:    value,
	}
	bz, err := kvPair.Marshal()
	if err != nil {
		logger.Error(err.Error(), "storeKey", kvPair.StoreKey)
		if fl.stopNodeOnErr {
			os.Exit(1)
		}
	}
	fl.stateCache = append(fl.stateCache, bz)
}

// PopStateCache returns the current state caches and set to nil
func (fl *MemoryListener) PopStateCache() [][]byte {
	res := fl.stateCache
	fl.stateCache = nil
	return res
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
	exposeStoreKeys := exposeStoreKeysSorted(exposeKeysStr, keys)
	listener := NewMemoryListener(bApp, abciListener, stopNodeOnErr)
	listeners := make(map[store.StoreKey][]store.WriteListener, len(exposeStoreKeys))
	for _, key := range exposeStoreKeys {
		listeners[key] = []store.WriteListener{listener}
	}
	// register listeners
	for key, lis := range listeners {
		bApp.cms.AddListeners(key, lis)
	}
	// register the plugin within the BaseApp
	// BaseApp will pass BeginBlock, DeliverTx, and EndBlock requests and responses
	// to the streaming services to update their ABCI context
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
