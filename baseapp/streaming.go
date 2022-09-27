package baseapp

import (
	"fmt"

	"github.com/spf13/cast"

	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/store/types"
)

// ABCIListener is the interface that we're exposing as a streaming.
type ABCIListener interface {
	// ListenBeginBlock updates the streaming service with the latest BeginBlock messages
	ListenBeginBlock(blockHeight int64, req []byte, res []byte) error
	// ListenEndBlock updates the steaming service with the latest EndBlock messages
	ListenEndBlock(blockHeight int64, req []byte, res []byte) error
	// ListenDeliverTx updates the steaming service with the latest DeliverTx messages
	ListenDeliverTx(blockHeight int64, req []byte, res []byte) error
	// ListenStoreKVPair updates the steaming service with the latest StoreKVPair messages
	ListenStoreKVPair(blockHeight int64, data []byte) error
}

// StreamingService interface for registering WriteListeners with the BaseApp and updating the service with the ABCI messages using the hooks
type StreamingService struct {
	// Listeners returns the streaming service's listeners for the BaseApp to register
	Listeners map[types.StoreKey][]types.WriteListener
	// ABCIListener interface for hooking into the ABCI messages from inside the BaseApp
	ABCIListener ABCIListener
}

// KVStoreListener is used so that we do not need to update the underlying
// io.Writer inside the StoreKVPairWriteListener everytime we begin writing
type KVStoreListener struct {
	BlockHeight func() int64
	listener    ABCIListener
}

// NewKVStoreListener create an instance of an NewKVStoreListener that sends StoreKVPair data to listening service
func NewKVStoreListener(listener ABCIListener, blockHeight func() int64) *KVStoreListener {
	return &KVStoreListener{listener: listener, BlockHeight: blockHeight}
}

// Write satisfies io.Writer
func (iw *KVStoreListener) Write(b []byte) (int, error) {
	blockHeight := iw.BlockHeight()
	if err := iw.listener.ListenStoreKVPair(blockHeight, b); err != nil {
		return 0, err
	}
	return len(b), nil
}

const (
	StreamingTomlKey       = "streaming"
	StreamingEnableTomlKey = "enable"
	StreamingPluginTomlKey = "plugin"
	StreamingKeysTomlKey   = "keys"
)

// RegisterStreamingService registers the ABCI streaming service provided by the streaming plugin.
func RegisterStreamingService(
	bApp *BaseApp,
	appOpts servertypes.AppOptions,
	kodec codec.BinaryCodec,
	keys map[string]*types.KVStoreKey,
	streamingService interface{},
) error {
	// type checking
	abciListener, ok := streamingService.(ABCIListener)
	if !ok {
		return fmt.Errorf("failed to register streaming service: failed type check %v", streamingService)
	}

	// expose keys
	keysKey := fmt.Sprintf("%s.%s", StreamingTomlKey, StreamingKeysTomlKey)
	exposeKeysStr := cast.ToStringSlice(appOpts.Get(keysKey))
	exposeStoreKeys := exposeStoreKeys(exposeKeysStr, keys)
	writer := NewKVStoreListener(abciListener, func() int64 { return bApp.deliverState.ctx.BlockHeight() })
	listener := types.NewStoreKVPairWriteListener(writer, kodec)
	listeners := make(map[types.StoreKey][]types.WriteListener, len(exposeStoreKeys))
	// in this case, we are using the same listener for each Store
	for _, key := range exposeStoreKeys {
		listeners[key] = append(listeners[key], listener)
	}

	bApp.SetStreamingService(StreamingService{
		Listeners:    listeners,
		ABCIListener: abciListener,
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

func exposeStoreKeys(keysStr []string, keys map[string]*types.KVStoreKey) []types.StoreKey {
	var exposeStoreKeys []types.StoreKey
	if exposeAll(keysStr) {
		exposeStoreKeys = make([]types.StoreKey, 0, len(keys))
		for _, storeKey := range keys {
			exposeStoreKeys = append(exposeStoreKeys, storeKey)
		}
	} else {
		exposeStoreKeys = make([]types.StoreKey, 0, len(keysStr))
		for _, keyStr := range keysStr {
			if storeKey, ok := keys[keyStr]; ok {
				exposeStoreKeys = append(exposeStoreKeys, storeKey)
			}
		}
	}

	return exposeStoreKeys
}
