package baseapp

import (
	"fmt"
	"github.com/spf13/cast"
	"os"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	store "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/types"
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
}

// StoreListener is the interface that we're exposing as a streaming service.
type StoreListener interface {
	// ListenStoreKVPair updates the streaming service with the latest state change message
	ListenStoreKVPair(ctx types.Context, pair store.StoreKVPair) error
}

// StreamingService for registering WriteListeners with the BaseApp and updating the service with the ABCI messages using the hooks
type StreamingService struct {
	// Listeners returns the streaming service's listeners for the BaseApp to register
	Listeners map[store.StoreKey][]store.WriteListener
	// ABCIListener interface for hooking into the ABCI messages from inside the BaseApp
	ABCIListener ABCIListener
	// StopNodeOnErr stops the node when true
	StopNodeOnErr bool
}

var (
	_ store.WriteListener = (*WriteListener)(nil)
)

// WriteListener writes state changes out to listening service
type WriteListener struct {
	bApp          *BaseApp
	listener      StoreListener
	stopNodeOnErr bool
}

// NewWriteListener create an instance of an NewWriteListener that sends StoreKVPair data to listening service
func NewWriteListener(
	bApp *BaseApp,
	listener StoreListener,
	stopNodeOnErr bool,
) *WriteListener {
	return &WriteListener{
		bApp:          bApp,
		listener:      listener,
		stopNodeOnErr: stopNodeOnErr,
	}
}

// OnWrite satisfies WriteListener.Listen
func (iw *WriteListener) OnWrite(storeKey store.StoreKey, key []byte, value []byte, delete bool) {
	ctx := iw.bApp.deliverState.ctx
	logger := iw.bApp.logger
	kvPair := new(store.StoreKVPair)
	kvPair.StoreKey = storeKey.Name()
	kvPair.Delete = delete
	kvPair.Key = key
	kvPair.Value = value
	if iw.stopNodeOnErr {
		if err := iw.listener.ListenStoreKVPair(ctx, *kvPair); err != nil {
			logger.Error(err.Error(), "storeKey", storeKey)
			os.Exit(1)
		}
	} else {
		go func() {
			if err := iw.listener.ListenStoreKVPair(ctx, *kvPair); err != nil {
				logger.Error(err.Error(), "storeKey", storeKey)
			}
		}()
	}
}

const (
	StreamingTomlKey              = "streaming"
	StreamingEnableTomlKey        = "enable"
	StreamingPluginTomlKey        = "plugin"
	StreamingKeysTomlKey          = "keys"
	StreamingStopNodeOnErrTomlKey = "stop-node-on-err"
)

// RegisterStreamingService registers the ABCI streaming service provided by the streaming plugin.
func RegisterStreamingService(
	bApp *BaseApp,
	appOpts servertypes.AppOptions,
	keys map[string]*store.KVStoreKey,
	streamingService interface{},
) error {
	// type checking
	abciListener, ok := streamingService.(ABCIListener)
	if !ok {
		return fmt.Errorf("failed to register streaming service: failed type check %v", streamingService)
	}

	// streaming service config
	stopNodeOnErrKey := fmt.Sprintf("%s.%s", StreamingTomlKey, StreamingStopNodeOnErrTomlKey)
	stopNodeOnErr := cast.ToBool(appOpts.Get(stopNodeOnErrKey))
	var listeners map[store.StoreKey][]store.WriteListener

	// streaming services can choose not to implement store listening
	storeListener, ok := streamingService.(StoreListener)
	if ok {
		keysKey := fmt.Sprintf("%s.%s", StreamingTomlKey, StreamingKeysTomlKey)
		exposeKeysStr := cast.ToStringSlice(appOpts.Get(keysKey))
		exposeStoreKeys := exposeStoreKeys(exposeKeysStr, keys)
		listeners = make(map[store.StoreKey][]store.WriteListener, len(exposeStoreKeys))
		writeListener := NewWriteListener(bApp, storeListener, stopNodeOnErr)
		for _, key := range exposeStoreKeys {
			listeners[key] = append(listeners[key], writeListener)
		}
	}

	// register service with the App
	bApp.SetStreamingService(StreamingService{
		Listeners:     listeners,
		ABCIListener:  abciListener,
		StopNodeOnErr: stopNodeOnErr,
	})

	return nil
}

func exposeAll(list []string) bool {
	for _, ele := range list {
		if ele == "*" {
			return true
		}
	}
	return false
}

func exposeStoreKeys(keysStr []string, keys map[string]*store.KVStoreKey) []store.StoreKey {
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

	return exposeStoreKeys
}
