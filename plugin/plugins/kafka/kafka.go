package file

import (
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"strings"
	"sync"

	"github.com/spf13/cast"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/plugin"
	"github.com/cosmos/cosmos-sdk/plugin/plugins/kafka/service"
	serverTypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/store/types"
)

// Plugin name and version
const (
	// PLUGIN_NAME is the name for this streaming service plugin
	PLUGIN_NAME = "kafka"

	// PLUGIN_VERSION is the version for this streaming service plugin
	PLUGIN_VERSION = "0.0.1"
)

// TOML configuration parameter keys
const (
	// TOPIC_PREFIX_PARAM is the Kafka topic where events will be streamed to
	TOPIC_PREFIX_PARAM = "topic_prefix"

	// FLUSH_TIMEOUT_MS_PARAM is the timeout setting passed to the producer.Flush(timeoutMs)
	FLUSH_TIMEOUT_MS_PARAM = "flush_timeout_ms"

	// PRODUCER_CONFIG_PARAM is a map of the Kafka Producer configuration properties
	PRODUCER_CONFIG_PARAM = "producer"

	// KEYS_PARAM is a list of the StoreKeys we want to expose for this streaming service
	KEYS_PARAM = "keys"

	// ACK_MODE configures whether to operate in fire-and-forget or success/failure acknowledgement mode
	ACK_MODE = "ack"

	// DELIVERED_BLOCK_WAIT_LIMIT the amount of time to wait for acknowledgment of success/failure of
	// message delivery of the current block before considering the delivery of messages failed.
	DELIVERED_BLOCK_WAIT_LIMIT = "delivered_block_wait_limit"
)

// Plugins is the exported symbol for loading this plugin
var Plugins = []plugin.Plugin{
	&streamingServicePlugin{},
}

type streamingServicePlugin struct {
	kss  *service.KafkaStreamingService
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
	topicPrefix := cast.ToString(ssp.opts.Get(fmt.Sprintf("%s.%s", tomlKeyPrefix, TOPIC_PREFIX_PARAM)))
	flushTimeoutMs := cast.ToInt(ssp.opts.Get(fmt.Sprintf("%s.%s", tomlKeyPrefix, FLUSH_TIMEOUT_MS_PARAM)))
	ack := cast.ToBool(ssp.opts.Get(fmt.Sprintf("%s.%s", tomlKeyPrefix, ACK_MODE)))
	deliveredBlockWaitLimit := cast.ToDuration(ssp.opts.Get(fmt.Sprintf("%s.%s", tomlKeyPrefix, DELIVERED_BLOCK_WAIT_LIMIT)))
	producerConfig := cast.ToStringMap(ssp.opts.Get(fmt.Sprintf("%s.%s", tomlKeyPrefix, PRODUCER_CONFIG_PARAM)))
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

	// Validate minimum producer config properties
	producerConfigKey := fmt.Sprintf("%s.%s.%s.%s", tomlKeyPrefix, PRODUCER_CONFIG_PARAM)

	if len(producerConfig) == 0 {
		m := fmt.Sprintf("Failed to register plugin. Empty properties for '%s': " +
			"client will not be able to connect to Kafka cluster", producerConfigKey)
		return errors.New(m)
	} else {
		bootstrapServers := cast.ToString(producerConfig["bootstrap_servers"])
		if len(bootstrapServers) == 0 {
			m := fmt.Sprintf("Failed to register plugin. No \"%s.%s\" configured:" +
				" client will not be able to connect to Kafka cluster", producerConfigKey, "bootstrap_servers")
			return errors.New(m)
		}
		if strings.TrimSpace(bootstrapServers) == "" {
			m := fmt.Sprintf("Failed to register plugin. Empty \"%s.%s\" configured:" +
				" client will not be able to connect to Kafka cluster", producerConfigKey, "bootstrap_servers")
			return errors.New(m)
		}
	}

	// load producer config into a kafka.ConfigMap
	producerConfigMap := kafka.ConfigMap{}
	for key, element := range producerConfig {
		key = strings.ReplaceAll(key, "_", ".")
		if err := producerConfigMap.SetKey(key, element); err != nil {
			return err
		}
	}

	var err error
	ssp.kss, err = service.NewKafkaStreamingService(
		bApp.Logger(),
		producerConfigMap,
		topicPrefix,
		flushTimeoutMs,
		exposeStoreKeys,
		marshaller,
		ack,
		deliveredBlockWaitLimit,
	)
	if err != nil {
		return err
	}
	// register the streaming service with the BaseApp
	bApp.SetStreamingService(ssp.kss)
	return nil
}

// Start satisfies the plugin.StateStreamingPlugin interface
func (ssp *streamingServicePlugin) Start(wg *sync.WaitGroup) error {
	return ssp.kss.Stream(wg)
}

// Close satisfies io.Closer
func (ssp *streamingServicePlugin) Close() error {
	return ssp.kss.Close()
}
