package iavlv2

import (
	"github.com/cosmos/iavl/v2"
	"github.com/cosmos/iavl/v2/metrics"
)

// Config is the configuration for the IAVL v2 tree.
type Config struct {
	CheckpointInterval  int64         `mapstructure:"checkpoint-interval" toml:"checkpoint-interval" comment:"CheckpointInterval set the interval of the checkpoint."`
	CheckpointMemory    uint64        `mapstructure:"checkpoint-memory" toml:"checkpoint-memory" comment:"CheckpointMemory set the memory of the checkpoint."`
	StateStorage        bool          `mapstructure:"state-storage" toml:"state-storage" comment:"StateStorage set the state storage."`
	HeightFilter        int8          `mapstructure:"height-filter" toml:"height-filter" comment:"HeightFilter set the height filter."`
	EvictionDepth       int8          `mapstructure:"eviction-depth" toml:"eviction-depth" comment:"EvictionDepth set the eviction depth."`
	MetricsProxy        metrics.Proxy `mapstructure:"metrics-proxy" toml:"metrics-proxy" comment:"MetricsProxy set the metrics proxy."`
	PruneRatio          float64       `mapstructure:"prune-ratio" toml:"prune-ratio" comment:"PruneRatio set the prune ratio."`
	MinimumKeepVersions int64         `mapstructure:"minimum-keep-versions" toml:"minimum-keep-versions" comment:"MinimumKeepVersions set the minimum keep versions."`
}

// ToTreeOptions converts the configuration to IAVL v2 tree options.
func (c *Config) ToTreeOptions() iavl.TreeOptions {
	return iavl.TreeOptions{
		CheckpointInterval:  c.CheckpointInterval,
		CheckpointMemory:    c.CheckpointMemory,
		StateStorage:        c.StateStorage,
		HeightFilter:        c.HeightFilter,
		EvictionDepth:       c.EvictionDepth,
		MetricsProxy:        c.MetricsProxy,
		PruneRatio:          c.PruneRatio,
		MinimumKeepVersions: c.MinimumKeepVersions,
	}
}

// DefaultConfig returns the default configuration for the IAVL tree.
func DefaultConfig() Config {
	defaultOptions := iavl.DefaultTreeOptions()

	return Config{
		CheckpointInterval:  200,
		CheckpointMemory:    defaultOptions.CheckpointMemory,
		StateStorage:        defaultOptions.StateStorage,
		HeightFilter:        1,
		EvictionDepth:       22,
		MetricsProxy:        defaultOptions.MetricsProxy,
		PruneRatio:          1,
		MinimumKeepVersions: defaultOptions.MinimumKeepVersions,
	}
}
