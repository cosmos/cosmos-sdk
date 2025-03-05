// Package iavl implements the IAVL tree commitment store.
package iavl

// Config contains the configuration for the IAVL tree.
// It provides options to customize the behavior of the tree.
type Config struct {
	// CacheSize is the size of the cache in memory.
	// A larger cache size can improve performance at the cost of memory usage.
	CacheSize int `mapstructure:"cache-size"`

	// SkipFastStorageUpgrade determines whether to skip the fast storage upgrade.
	// If true, the tree will work like no fast storage and always not upgrade fast storage.
	SkipFastStorageUpgrade bool `mapstructure:"skip-fast-storage-upgrade"`

	// EnableHistoricalQueries enables querying historical data from the old tree.
	// When enabled, queries for versions before the migration height will use the old tree.
	EnableHistoricalQueries bool `mapstructure:"enable-historical-queries"`
}

// DefaultConfig returns the default configuration.
// It provides sensible defaults for all configuration options.
func DefaultConfig() *Config {
	return &Config{
		CacheSize:              1000000,
		SkipFastStorageUpgrade: false,
		EnableHistoricalQueries: false,
	}
}
