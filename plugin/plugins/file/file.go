package file

import (
	"fmt"
	"sync"
	"time"

	"github.com/spf13/cast"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/plugin"
	"github.com/cosmos/cosmos-sdk/plugin/plugins/file/service"
	serverTypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/store/types"
)

// Plugin name and version
const (
	// PLUGIN_NAME is the name for this streaming service plugin
	PLUGIN_NAME = "file"

	// PLUGIN_VERSION is the version for this streaming service plugin
	PLUGIN_VERSION = "0.0.1"
)

// TOML configuration parameter keys
const (
	// PREFIX_PARAM is an optional prefix to prepend to the files we write
	PREFIX_PARAM = "prefix"

	// WRITE_DIR_PARAM is the directory we want to write files out to
	WRITE_DIR_PARAM = "write_dir"

	// KEYS_PARAM is a list of the StoreKeys we want to expose for this streaming service
	KEYS_PARAM = "keys"

	// HALT_APP_ON_DELIVERY_ERROR whether or not to halt the application when plugin fails to deliver message(s)
	HALT_APP_ON_DELIVERY_ERROR = "halt_app_on_delivery_error"
)

const minWaitDuration = time.Millisecond * 10

// Plugins is the exported symbol for loading this plugin
var Plugins = []plugin.Plugin{
	&streamingServicePlugin{},
}

type streamingServicePlugin struct {
	fss  *service.FileStreamingService
	opts serverTypes.AppOptions
}

var _ plugin.StateStreamingPlugin = (*streamingServicePlugin)(nil)

// Name satisfies the plugin.Plugin interface
func (ssp *streamingServicePlugin) Name() string {
	return PLUGIN_NAME
}

// Version satisfies the plugin.Plugin interface
func (ssp *streamingServicePlugin) Version() string {
	return PLUGIN_VERSION
}

// Init satisfies the plugin.Plugin interface
func (ssp *streamingServicePlugin) Init(env serverTypes.AppOptions) error {
	ssp.opts = env
	return nil
}

// Register satisfies the plugin.StateStreamingPlugin interface
func (ssp *streamingServicePlugin) Register(bApp *baseapp.BaseApp, marshaller codec.BinaryCodec, keys map[string]*types.KVStoreKey) error {
	// load all the params required for this plugin from the provided AppOptions
	tomlKeyPrefix := fmt.Sprintf("%s.%s.%s", plugin.PLUGINS_TOML_KEY, plugin.STREAMING_TOML_KEY, PLUGIN_NAME)
	filePrefix := cast.ToString(ssp.opts.Get(fmt.Sprintf("%s.%s", tomlKeyPrefix, PREFIX_PARAM)))
	fileDir := cast.ToString(ssp.opts.Get(fmt.Sprintf("%s.%s", tomlKeyPrefix, WRITE_DIR_PARAM)))
	// get the store keys allowed to be exposed for this streaming service
	exposeKeyStrings := cast.ToStringSlice(ssp.opts.Get(fmt.Sprintf("%s.%s", tomlKeyPrefix, KEYS_PARAM)))
	var exposeStoreKeys []types.StoreKey
	if len(exposeKeyStrings) > 0 {
		exposeStoreKeys = make([]types.StoreKey, 0, len(exposeKeyStrings))
		for _, keyStr := range exposeKeyStrings {
			if storeKey, ok := keys[keyStr]; ok {
				exposeStoreKeys = append(exposeStoreKeys, storeKey)
			}
		}
	} else { // if none are specified, we expose all the keys
		exposeStoreKeys = make([]types.StoreKey, 0, len(keys))
		for _, storeKey := range keys {
			exposeStoreKeys = append(exposeStoreKeys, storeKey)
		}
	}
	haltAppOnDeliveryError := cast.ToBool(ssp.opts.Get(fmt.Sprintf("%s.%s", tomlKeyPrefix, HALT_APP_ON_DELIVERY_ERROR)))
	var err error
	ssp.fss, err = service.NewFileStreamingService(fileDir, filePrefix, exposeStoreKeys, marshaller, haltAppOnDeliveryError)
	if err != nil {
		return err
	}
	// register the streaming service with the BaseApp
	bApp.SetStreamingService(ssp.fss)
	return nil
}

// Start satisfies the plugin.StateStreamingPlugin interface
func (ssp *streamingServicePlugin) Start(wg *sync.WaitGroup) error {
	return ssp.fss.Stream(wg)
}

// Close satisfies io.Closer
func (ssp *streamingServicePlugin) Close() error {
	return ssp.fss.Close()
}
