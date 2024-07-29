package iavl

// Config is the configuration for the IAVL tree.
type Config struct {
	CacheSize              int  `mapstructure:"cache-size" toml:"cache-size" comment:"CacheSize set the size of the iavl tree cache."`
	SkipFastStorageUpgrade bool `mapstructure:"skip-fast-storage-upgrade" toml:"skip-fast-storage-upgrade" comment:"If true, the tree will work like no fast storage and always not upgrade fast storage."`
}

// DefaultConfig returns the default configuration for the IAVL tree.
func DefaultConfig() *Config {
	return &Config{
		CacheSize:              1000,
		SkipFastStorageUpgrade: false,
	}
}
