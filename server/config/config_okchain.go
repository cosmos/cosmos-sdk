package config

import (
	"os"
)

const (
	BackendOrmEngineTypeSqlite = "sqlite3"
	BackendOrmEngineTypeMysql  = "mysql"
)

var (
	DefaultBackendNodeHome     = os.ExpandEnv("$HOME/.okchaind")
	DefaultBackendNodeDataHome = DefaultBackendNodeHome + "/data"
)

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

	// if engine_type is sqlite3, it should be a local path, e.g.) /Users/lingting.fu/.okchaind/data/sqlite3/backend.db
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
	c.OrmEngine.ConnectStr = DefaultBackendNodeDataHome + string(os.PathSeparator) + c.OrmEngine.EngineType + string(os.PathSeparator) + "backend.sqlite3"

	return &c
}
