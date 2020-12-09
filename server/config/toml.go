package config

import (
	"bytes"
	"text/template"

	"github.com/spf13/viper"
	tmos "github.com/tendermint/tendermint/libs/os"
)

const defaultConfigTemplate = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

##### main base config options #####

# The minimum gas prices a validator is willing to accept for processing a
# transaction. A transaction's fees must meet the minimum of any denomination
# specified in this config (e.g. 0.25token1;0.0001token2).
minimum-gas-prices = "{{ .BaseConfig.MinGasPrices }}"

# default: the last 100 states are kept in addition to every 500th state; pruning at 10 block intervals
# nothing: all historic states will be saved, nothing will be deleted (i.e. archiving node)
# everything: all saved states will be deleted, storing only the current state; pruning at 10 block intervals
# custom: allow pruning options to be manually specified through 'pruning-keep-recent', 'pruning-keep-every', and 'pruning-interval'
pruning = "{{ .BaseConfig.Pruning }}"

# These are applied if and only if the pruning strategy is custom.
pruning-keep-recent = "{{ .BaseConfig.PruningKeepRecent }}"
pruning-keep-every = "{{ .BaseConfig.PruningKeepEvery }}"
pruning-interval = "{{ .BaseConfig.PruningInterval }}"

# HaltHeight contains a non-zero block height at which a node will gracefully
# halt and shutdown that can be used to assist upgrades and testing.
#
# Note: Commitment of state will be attempted on the corresponding block.
halt-height = {{ .BaseConfig.HaltHeight }}

# HaltTime contains a non-zero minimum block time (in Unix seconds) at which
# a node will gracefully halt and shutdown that can be used to assist upgrades
# and testing.
#
# Note: Commitment of state will be attempted on the corresponding block.
halt-time = {{ .BaseConfig.HaltTime }}

# InterBlockCache enables inter-block caching.
inter-block-cache = {{ .BaseConfig.InterBlockCache }}

##### backend configuration options #####
[backend]
enable_backend = "{{ .BackendConfig.EnableBackend }}"
enable_mkt_compute = "{{ .BackendConfig.EnableMktCompute }}"
log_sql = "{{ .BackendConfig.LogSQL }}"
clean_ups_kept_days = "{{ .BackendConfig.CleanUpsKeptDays }}"
clean_ups_time = "{{ .BackendConfig.CleanUpsTime }}"
[backend.orm_engine]
engine_type = "{{ .BackendConfig.OrmEngine.EngineType }}"
connect_str = "{{ js .BackendConfig.OrmEngine.ConnectStr }}"
[stream]
engine = "{{ .StreamConfig.Engine }}"
klines_query_connect = "{{ .StreamConfig.KlineQueryConnect }}"

worker_id = "{{ .StreamConfig.WorkerId }}"
redis_scheduler = "{{ .StreamConfig.RedisScheduler }}"
redis_lock = "{{ .StreamConfig.RedisLock }}"
local_lock_dir = "{{ js .StreamConfig.LocalLockDir }}"
cache_queue_capacity = "{{ .StreamConfig.CacheQueueCapacity }}"

market_topic = "{{ .StreamConfig.MarketTopic }}"
market_partition = "{{ .StreamConfig.MarketPartition }}"

market_service_enable = "{{ .StreamConfig.MarketServiceEnable }}"
market_nacos_urls = "{{ .StreamConfig.MarketNacosUrls }}"
market_nacos_namespace_id = "{{ .StreamConfig.MarketNacosNamespaceId }}"
market_nacos_clusters = "{{ .StreamConfig.MarketNacosClusters }}"
market_nacos_service_name = "{{ .StreamConfig.MarketNacosServiceName }}"
market_nacos_group_name = "{{ .StreamConfig.MarketNacosGroupName }}"

market_eureka_name = "{{ .StreamConfig.MarketEurekaName }}"
eureka_server_url = "{{ .StreamConfig.EurekaServerUrl }}"

rest_application_name = "{{ .StreamConfig.RestApplicationName }}"
rest_nacos_urls = "{{ .StreamConfig.RestNacosUrls }}"
rest_nacos_namespace_id = "{{ .StreamConfig.RestNacosNamespaceId }}"

pushservice_pulsar_public_topic = "{{ .StreamConfig.PushservicePulsarPublicTopic }}"
pushservice_pulsar_private_topic = "{{ .StreamConfig.PushservicePulsarPrivateTopic }}"
pushservice_pulsar_depth_topic = "{{ .StreamConfig.PushservicePulsarDepthTopic }}"
redis_require_pass = "{{ .StreamConfig.RedisRequirePass }}"
`

var configTemplate *template.Template

func init() {
	var err error
	tmpl := template.New("appConfigFileTemplate")
	if configTemplate, err = tmpl.Parse(defaultConfigTemplate); err != nil {
		panic(err)
	}
}

// ParseConfig retrieves the default environment configuration for the
// application.
func ParseConfig() (*Config, error) {
	conf := DefaultConfig()
	err := viper.Unmarshal(conf)
	return conf, err
}

// WriteConfigFile renders config using the template and writes it to
// configFilePath.
func WriteConfigFile(configFilePath string, config *Config) {
	var buffer bytes.Buffer

	if err := configTemplate.Execute(&buffer, config); err != nil {
		panic(err)
	}

	tmos.MustWriteFile(configFilePath, buffer.Bytes(), 0644)
}
