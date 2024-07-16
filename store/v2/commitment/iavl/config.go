package iavl

// Config is the configuration for the IAVL tree.
type Config struct {
	CacheSize              int  `mapstructure:"cache_size"`
	SkipFastStorageUpgrade bool `mapstructure:"skip_fast_storage_upgrade"`
}

// DefaultConfig returns the default configuration for the IAVL tree.
func DefaultConfig() *Config {
	return &Config{
		CacheSize:              1000,
		SkipFastStorageUpgrade: false,
	}
}
