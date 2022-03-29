package file

import (
	"fmt"
	"sync"

	"github.com/spf13/cast"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/plugin"
	"github.com/cosmos/cosmos-sdk/plugin/plugins/trace/service"
	serverTypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/store/types"
)

// Plugin name and version
const (
	// PLUGIN_NAME is the name for this streaming service plugin
	PLUGIN_NAME = "trace"

	// PLUGIN_VERSION is the version for this streaming service plugin
	PLUGIN_VERSION = "0.0.1"
)

// TOML configuration parameter keys
const (
	// KEYS_PARAM is a list of the StoreKeys we want to expose for this streaming service
	KEYS_PARAM = "keys"

	PRINT_DATA_TO_STDOUT_PARAM = "print_data_to_stdout"

	// HALT_APP_ON_DELIVERY_ERROR whether or not to halt the application when plugin fails to deliver message(s)
	HALT_APP_ON_DELIVERY_ERROR = "halt_app_on_delivery_error"
)

// Plugins is the exported symbol for loading this plugin
var Plugins = []plugin.Plugin{
	&streamingServicePlugin{},
}

type streamingServicePlugin struct {
	tss  *service.TraceStreamingService
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
func (ssp *streamingServicePlugin) Register(
	bApp *baseapp.BaseApp,
	marshaller codec.BinaryCodec,
	keys map[string]*types.KVStoreKey,
) error {
	// load all the params required for this plugin from the provided AppOptions
	tomlKeyPrefix := fmt.Sprintf("%s.%s.%s", plugin.PLUGINS_TOML_KEY, plugin.STREAMING_TOML_KEY, PLUGIN_NAME)
	printDataToStdout := cast.ToBool(ssp.opts.Get(fmt.Sprintf("%s.%s", tomlKeyPrefix, PRINT_DATA_TO_STDOUT_PARAM)))
	haltAppOnDeliveryError := cast.ToBool(ssp.opts.Get(fmt.Sprintf("%s.%s", tomlKeyPrefix, HALT_APP_ON_DELIVERY_ERROR)))


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

	var err error
	ssp.tss, err = service.NewTraceStreamingService(exposeStoreKeys, marshaller, printDataToStdout, haltAppOnDeliveryError)
	if err != nil {
		return err
	}
	// register the streaming service with the BaseApp
	bApp.SetStreamingService(ssp.tss)
	return nil
}

// Start satisfies the plugin.StateStreamingPlugin interface
func (ssp *streamingServicePlugin) Start(wg *sync.WaitGroup) error {
	return ssp.tss.Stream(wg)
}

// Close satisfies io.Closer
func (ssp *streamingServicePlugin) Close() error {
	return ssp.tss.Close()
}
