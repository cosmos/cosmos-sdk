package config

import (
	"os"
	"path/filepath"
)

const (
	BackendOrmEngineTypeSqlite = "sqlite3"
	BackendOrmEngineTypeMysql  = "mysql"
)

var defaultNodeHome = os.ExpandEnv("$HOME/.exchaind")

// SetNodeHome sets the root directory for all data.
func SetNodeHome(home string) {
	defaultNodeHome = home
}

// GetNodeHome returns the root directory for all data.
func GetNodeHome() string {
	return defaultNodeHome
}

type BackendConfig struct {
	EnableBackend    bool `json:"enable_backend" mapstructure:"enable_backend"`
	EnableMktCompute bool `json:"enable_mkt_compute" mapstructure:"enable_mkt_compute"`
	//HotKeptDays      int                  `json:"hot_kept_days" mapstructure:"hot_kept_days"`
	//UpdateFreq       int64                `json:"update_freq" mapstructure:"update_freq"`       // unit: second
	//BufferSize       int                  `json:"buffer_size" mapstructure:"buffer_size"`       //
	//SyncMode         string               `json:"sync_mode" mapstructure:"sync_mode"`           // mode: block or minutes
	LogSQL           bool                 `json:"log_sql" mapstructure:"log_sql"`               //
	CleanUpsKeptDays map[string]int       `json:"clean_ups_kept_days"`                          // 0 <= x <= 60
	CleanUpsTime     string               `json:"clean_ups_time" mapstructure:"clean_ups_time"` // e.g.) 00:00:00, CleanUp job will be fired at this time.
	OrmEngine        BackendOrmEngineInfo `json:"orm_engine" mapstructure:"orm_engine"`         //
}

type BackendOrmEngineInfo struct {
	// engine type should be sqlite3 or mysql
	EngineType string `json:"engine_type" mapstructure:"engine_type"`

	// if engine_type is sqlite3, it should be a local path, e.g.) /Users/lingting.fu/.exchaind/data/sqlite3/backend.db
	// if engine_type is mysql, it should be "[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]"
	ConnectStr string `json:"connect_str" mapstructure:"connect_str"`
}

func DefaultBackendConfig() *BackendConfig {
	c := BackendConfig{}

	c.EnableBackend = false
	c.EnableMktCompute = false
	//c.HotKeptDays = 3
	//c.UpdateFreq = 60
	//c.BufferSize = 4096
	c.LogSQL = false
	c.CleanUpsTime = "00:00:00"
	c.CleanUpsKeptDays = map[string]int{}
	c.CleanUpsKeptDays["kline_m1"] = 120
	c.CleanUpsKeptDays["kline_m3"] = 120
	c.CleanUpsKeptDays["kline_m5"] = 120

	c.OrmEngine.EngineType = BackendOrmEngineTypeSqlite
	c.OrmEngine.ConnectStr = filepath.Join(GetNodeHome(), "data", c.OrmEngine.EngineType, "backend.sqlite3")

	return &c
}

// StreamConfig - config for okchain stream module
type StreamConfig struct {
	Engine            string `json:"engine" mapstructure:"engine"`
	KlineQueryConnect string `json:"klines_query_connect" mapstructure:"klines_query_connect"`

	// distr-lock config
	WorkerId           string `json:"worker_id" mapstructure:"worker_id"`
	RedisScheduler     string `json:"redis_scheduler" mapstructure:"redis_scheduler"`
	RedisLock          string `json:"redis_lock" mapstructure:"redis_lock"`
	LocalLockDir       string `json:"local_lock_dir" mapstructure:"local_lock_dir"`
	CacheQueueCapacity int    `json:"cache_queue_capacity" mapstructure:"cache_queue_capacity"`

	// kafka/pulsar service config for transfering match results
	MarketTopic     string `json:"market_topic" mapstructure:"market_topic"`
	MarketPartition int    `json:"market_partition" mapstructure:"market_partition"`

	// market service of nacos config for getting market service url, used for registering token
	MarketServiceEnable    bool     `json:"market_service_enable" mapstructure:"market_service_enable"`
	MarketNacosUrls        string   `json:"market_nacos_urls" mapstructure:"market_nacos_urls"`
	MarketNacosNamespaceId string   `json:"market_nacos_namespace_id" mapstructure:"market_nacos_namespace_id"`
	MarketNacosClusters    []string `json:"market_nacos_clusters" mapstructure:"market_nacos_clusters"`
	MarketNacosServiceName string   `json:"market_nacos_service_name" mapstructure:"market_nacos_service_name"`
	MarketNacosGroupName   string   `json:"market_nacos_group_name" mapstructure:"market_nacos_group_name"`

	// market service of eurka config for getting market service url, used for registering token
	MarketEurekaName string `json:"market_eureka_name" mapstructure:"market_eureka_name"`
	EurekaServerUrl  string `json:"eureka_server_url" mapstructure:"eureka_server_url"`

	// restful service config for registering restful-node
	RestApplicationName  string `json:"rest_application_name" mapstructure:"rest_application_name"`
	RestNacosUrls        string `json:"rest_nacos_urls" mapstructure:"rest_nacos_urls"`
	RestNacosNamespaceId string `json:"rest_nacos_namespace_id" mapstructure:"rest_nacos_namespace_id"`

	// push service config
	PushservicePulsarPublicTopic  string `json:"pushservice_pulsar_public_topic" mapstructure:"pushservice_pulsar_public_topic"`
	PushservicePulsarPrivateTopic string `json:"pushservice_pulsar_private_topic" mapstructure:"pushservice_pulsar_private_topic"`
	PushservicePulsarDepthTopic   string `json:"pushservice_pulsar_depth_topic" mapstructure:"pushservice_pulsar_depth_topic"`
	RedisRequirePass              string `json:"redis_require_pass" mapstructure:"redis_require_pass"`
}

// DefaultStreamConfig returns default config for okchain stream module
func DefaultStreamConfig() *StreamConfig {
	return &StreamConfig{
		Engine: "",
	}
}
